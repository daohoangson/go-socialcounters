package services

import (
	"github.com/daohoangson/go-socialcounters/utils"
)

var FacebookService = "Facebook"
var TwitterService = "Twitter"
var GoogleService = "Google"

type MapServiceCount map[string]int64
type MapUrlServiceCount map[string]MapServiceCount
type MapUrlResult map[string]result

type request struct {
	Service  string
	Worker   worker
	Urls     []string
	Ttl      int64
	Response []byte
	Error    error
	Results  MapUrlResult
}

type result struct {
	Count    int64
	Error    error
	Response []byte
}

type worker func(utils.Utils, *request)
type workerLegacy func(utils.Utils, string) result
