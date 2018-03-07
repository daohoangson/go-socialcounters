package services

import (
	"github.com/daohoangson/go-socialcounters/utils"
)

const SERVICE_FACEBOOK = "Facebook"
const SERVICE_TWITTER = "Twitter"

const COUNT_NO_VALUE = int64(-1)
const COUNT_INITIAL_VALUE = int64(0)

type MapServiceRequest map[string]request
type MapServiceCount map[string]int64
type MapUrlServiceCount map[string]MapServiceCount
type MapUrlResult map[string]result
type sliceCache []cache

type request struct {
	Service  string
	Worker   worker
	Urls     []string
	Response []byte
	Error    error
	Results  MapUrlResult
}

type result struct {
	Count    int64
	Error    error
	Response []byte
}

type cache struct {
	Service string
	Url     string
	Count   int64
	TtlLeft int64
}

type worker func(utils.Utils, *request)
type workerLegacy func(utils.Utils, string) result
