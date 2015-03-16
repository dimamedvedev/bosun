package collectors

import (
	"encoding/json"
	"net/http"
	//	"regexp"
	//	"strings"
	//	"time"

	"bosun.org/metadata"
	"bosun.org/opentsdb"
	//	"bosun.org/util"
)

func init() {
	collectors = append(collectors, &IntervalCollector{F: c_riak, Enable: enableRiak})
}

const (
	riakURL string = "http://localhost:8098/stats"
)

func enableRiak() bool {
	return enableURL(riakURL)()
}

func c_riak() (opentsdb.MultiDataPoint, error) {
	var md opentsdb.MultiDataPoint
	res, err := http.Get(riakURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var r map[string]interface{}
	j := json.NewDecoder(res.Body)
	if err := j.Decode(&r); err != nil {
		return nil, err
	}
	for k, v := range r {
		if _, ok := v.(float64); ok {
			Add(&md, "riak."+k, v, nil, metadata.Unknown, metadata.None, "")
		}
	}
	return md, nil
}
