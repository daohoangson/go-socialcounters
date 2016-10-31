package services

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/daohoangson/go-socialcounters/utils"
)

var workers = map[string]worker{
	FacebookService: facebookWorker,
	TwitterService:  twitterWorker,
	GoogleService:   googleWorker,
}

func Batch(u utils.Utils, data *MapUrlServiceCount, ttl int64) {
	requests := make(map[string]request)
	var wg sync.WaitGroup

	for url, services := range *data {
		for service, count := range services {
			if count > 0 {
				continue
			}

			worker, ok := workers[service]
			if !ok {
				u.Errorf("Unrecognized service %s", service)
				continue
			}

			if value, err := u.MemoryGet(getCacheKey(service, url)); err == nil {
				if count, err := strconv.ParseInt(string(value), 10, 64); err == nil {
					(*data)[url][service] = count
					continue
				}
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

	for _, req := range requests {
		wg.Add(1)
		go func(u utils.Utils, req request) {
			defer wg.Done()

			req.Worker(u, &req)
			for url, res := range req.Results {
				(*data)[url][req.Service] = res.Count

				if res.Error != nil {
					u.Errorf("Error for %s on %s: %s", url, req.Service, res.Error)
				}

				cacheKey := getCacheKey(req.Service, url)
				cacheValue := []byte(fmt.Sprintf("%d", res.Count))
				cacheTtl := getCacheTtl(u, req, url, res)
				u.MemorySet(cacheKey, cacheValue, cacheTtl)
			}
		}(u, req)
	}

	wg.Wait()
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

	if ttlRestrictedConfig := u.ConfigGet("TTL_COUNT_EQUALS_ZERO"); ttlRestrictedConfig != "" {
		if ttlRestricted, err := strconv.ParseInt(ttlRestrictedConfig, 10, 64); err == nil {
			ttl = ttlRestricted
			resultTtlRestricted = true
			u.Infof("Restricted TTL for %s on %s: %d", url, req.Service, ttl)
		}
	}

	if !resultTtlRestricted {
		u.Debugf("%s(%s).Count == 0 without TTL restriction", req.Service, url)
	}

	return ttl
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
