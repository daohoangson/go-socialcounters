package services

import (
	"github.com/daohoangson/go-socialcounters/utils"
)

const countNoValue = int64(-1)
const countInitValue = int64(0)
const serviceFacebook = "Facebook"
const serviceTwitter = "Twitter"

type mapServiceRequest map[string]request
type mapServiceCount map[string]int64
type mapURLServiceCount map[string]mapServiceCount
type mapURLResult map[string]result
type sliceCache []cache

type request struct {
	Service  string
	Worker   worker
	Urls     []string
	Response []byte
	Error    error
	Results  mapURLResult
}

type result struct {
	Count    int64
	Error    error
	Response []byte
}

type cache struct {
	Service string
	URL     string
	Count   int64
	TTLLeft int64
}

type worker func(utils.Utils, *request)
type workerLegacy func(utils.Utils, string) result
