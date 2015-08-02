package collectors

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"bosun.org/_third_party/github.com/dimamedvedev/pstree"

	"bosun.org/opentsdb"
)

const (
	runitServicesDir = "/etc/service"
)

func RunitServices(whiteList, blackList string) error {
	collectors = append(collectors,
		&IntervalCollector{
			F: func() (opentsdb.MultiDataPoint, error) {
				return runitServices(whiteList, blackList)
			},
			name: fmt.Sprintf("runit-%s-%s", whiteList, blackList),
		},
	)
	return nil
}

func runitServices(whiteList, blackList string) (opentsdb.MultiDataPoint, error) {
	whiteRe, err := regexp.Compile(whiteList)
	if err != nil {
		return nil, err
	}
	if blackList == "" {
		blackList = "^$"
	}
	blackRe, err := regexp.Compile(blackList)
	if err != nil {
		return nil, err
	}
	psTree, err := pstree.New()
	if err != nil {
		return nil, err
	}

	wps, err := runitWatchedProc(psTree, "/", "/etc/service", whiteRe, blackRe)
	if err != nil {
		return nil, err
	}
	var md opentsdb.MultiDataPoint
	for _, wp := range wps {
		if e := linuxProcMonitor(wp, &md); e != nil {
			return nil, e
		}
	}
	return md, nil
}

// ReadPid tries to find and read pidfile into an int
func ReadPid(prefix string, filepaths ...string) (int, error) {
	for _, filepath := range filepaths {
		content, err := ioutil.ReadFile(filepath)
		if err != nil {
			continue
		}
		trimmed := strings.Trim(string(content[:]), "\n")
		_, err = os.Stat(prefix + "proc/" + trimmed)
		if err != nil {
			continue
		}
		pid, err := strconv.ParseInt(trimmed, 10, 32)
		if err != nil {
			continue
		}
		return int(pid), nil
	}
	return 0, fmt.Errorf("pid does not exist\n")
}

func runitWatchedProc(psTree *pstree.Tree, prefix string, svcDir string, whiteRe, blackRe *regexp.Regexp) ([]*WatchedProc, error) {
	matches, err := filepath.Glob(prefix + svcDir + "/*/supervise/pid")
	if err != nil {

		return nil, err
	}
	wp := []*WatchedProc{}
	for _, match := range matches {
		pid, err := ReadPid(prefix, match)
		if err != nil {
			fmt.Printf("Err: %v", err)
			continue
		}
		matchSplit := strings.Split(match, "/")
		// '/etc/service/name/supervise/pid'
		name := matchSplit[len(matchSplit)-3]
		if !whiteRe.MatchString(name) || blackRe.MatchString(name) {
			continue
		}
		wp = append(wp,
			&WatchedProc{
				Name:      name,
				Processes: psTree.SubTreeMapID(pid),
				ArgMatch:  regexp.MustCompile(""),
				idPool:    nil,
			},
		)
	}
	return wp, nil
}
