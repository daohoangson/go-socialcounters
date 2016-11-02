package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daohoangson/go-socialcounters/utils"
)

var workers = map[string]worker{
	SERVICE_FACEBOOK: facebookWorker,
	SERVICE_TWITTER:  twitterWorker,
	SERVICE_GOOGLE:   googleWorker,
}

func DataSetup() MapUrlServiceCount {
	return make(MapUrlServiceCount)
}

func DataAdd(data *MapUrlServiceCount, service string, url string) {
	if _, ok := (*data)[url]; !ok {
		(*data)[url] = make(MapServiceCount)
	}

	(*data)[url][service] = COUNT_NO_VALUE
}

func Batch(u utils.Utils, data *MapUrlServiceCount, ttl int64) {
	dataNeedRefresh := fillDataFromCache(u, data, true)
	requests := buildRequests(u, data, ttl, true)
	executeRequests(u, &requests, data)

	if len(dataNeedRefresh) > 0 {
		u.Schedule(fmt.Sprintf("refresh?ttl=%d", ttl), dataNeedRefresh, 1)
	}

	// make sure returned value has a minimum value of 0
	// non-positive count result doesn't make sense...
	for url, services := range *data {
		for service, count := range services {
			if count < 0 {
				(*data)[url][service] = 0
			}
		}
	}
}

func Refresh(u utils.Utils, data *MapUrlServiceCount, ttl int64) {
	dataNeedRefresh := fillDataFromCache(u, data, false)
	requests := buildRequests(u, &dataNeedRefresh, ttl, false)

	executeRequests(u, &requests, data)
}

func fillDataFromCache(u utils.Utils, data *MapUrlServiceCount, handleNoValueOnly bool) MapUrlServiceCount {
	dataNeedRefresh := make(MapUrlServiceCount)

	for url, services := range *data {
		for service, count := range services {
			if handleNoValueOnly && count != COUNT_NO_VALUE {
				continue
			}

			if count, ttlLeft, err := getCache(u, getCacheKey(service, url)); err == nil {
				(*data)[url][service] = count
				trackDataNeedRefresh(u, &dataNeedRefresh, service, url, count, ttlLeft)
				continue
			}
		}
	}

	return dataNeedRefresh
}

func buildRequests(u utils.Utils, data *MapUrlServiceCount, ttl int64, handleNoValueOnly bool) MapServiceRequest {
	requests := make(MapServiceRequest)

	for url, services := range *data {
		for service, count := range services {
			if handleNoValueOnly && count != COUNT_NO_VALUE {
				continue
			}

			worker, ok := workers[service]
			if !ok {
				u.Errorf("services.buildRequests: Unrecognized service %s", service)
				continue
			}

			if req, ok := requests[service]; ok {
				req.Urls = append(req.Urls, url)
				requests[service] = req
			} else {
				var newReq request
				newReq.Service = service
				newReq.Worker = worker
				newReq.Urls = []string{url}
				newReq.Ttl = ttl
				newReq.Results = make(MapUrlResult)

				requests[service] = newReq
			}
		}
	}

	return requests
}

func executeRequests(u utils.Utils, requests *MapServiceRequest, data *MapUrlServiceCount) {
	var wg sync.WaitGroup
	for _, req := range *requests {
		wg.Add(1)
		go func(req request) {
			defer wg.Done()

			req.Worker(u, &req)
			for url, res := range req.Results {
				cacheKey := getCacheKey(req.Service, url)
				cacheTtl := getCacheTtl(u, req, url, res)
				setCache(u, cacheKey, res.Count, cacheTtl)

				oldCount, _ := (*data)[url][req.Service]
				(*data)[url][req.Service] = res.Count

				if res.Error != nil {
					u.Errorf("services.executeRequests: %s on %s: %v", url, req.Service, res.Error)
				} else {
					if res.Count > oldCount {
						u.HistorySave(req.Service, url, res.Count)
					}
				}
			}
		}(req)
	}

	wg.Wait()
}

func trackDataNeedRefresh(u utils.Utils, data *MapUrlServiceCount, service string, url string, cacheCount int64, cacheTtlLeft int64) {
	ttlNeedsRefresh, err := utils.ConfigGetInt(u, "TTL_NEEDS_REFRESH")
	if err != nil || ttlNeedsRefresh < 1 || cacheTtlLeft < 1 || cacheTtlLeft > ttlNeedsRefresh {
		return
	}

	DataAdd(data, service, url)
	(*data)[url][service] = cacheCount
}

func getCacheKey(service string, url string) string {
	return fmt.Sprintf("%s/%s", service, url)
}

func getCacheTtl(u utils.Utils, req request, url string, res result) int64 {
	ttl := req.Ttl

	if res.Error != nil || res.Count > 0 {
		return ttl
	}

	resultTtlRestricted := false

	if ttlRestricted, err := utils.ConfigGetInt(u, "TTL_COUNT_EQUALS_ZERO"); err == nil {
		ttl = ttlRestricted
		resultTtlRestricted = true
		u.Infof("Restricted TTL for %s on %s: %d", url, req.Service, ttl)
	}

	if !resultTtlRestricted {
		u.Debugf("%s(%s).Count == 0 without TTL restriction", req.Service, url)
	}

	return ttl
}

func setCache(u utils.Utils, key string, count int64, ttl int64) error {
	value := fmt.Sprintf("%d;%d", count, time.Now().Unix() + ttl)
	return u.MemorySet(key, value, ttl)
}

func getCache(u utils.Utils, key string) (int64, int64, error) {
	count := int64(0)
	ttlLeft := int64(0)
	var e error

	if value, err := u.MemoryGet(key); err != nil {
		e = err
	} else {
		parts := strings.Split(value, ";")

		if parsed, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
			e = err
		} else {
			count = parsed
		}

		if e == nil && len(parts) >= 2 {
			// value with expire timestamp
			if exp, expErr := strconv.ParseInt(parts[1], 10, 64); expErr == nil {
				ttlLeft = exp - time.Now().Unix()
			}
		}
	}

	return count, ttlLeft, e
}

func legacyBatch(u utils.Utils, req *request, f workerLegacy) {
	var wg sync.WaitGroup

	for _, url := range req.Urls {
		wg.Add(1)
		go func(u utils.Utils, req *request, f workerLegacy, url string) {
			defer wg.Done()
			req.Results[url] = f(u, url)
		}(u, req, f, url)
	}

	wg.Wait()
}
