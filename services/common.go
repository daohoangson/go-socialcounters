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

var dataNeedRefresh *MapUrlServiceCount
var dataNeedRefreshCount = int64(0)

func DataSetup() MapUrlServiceCount {
	return make(MapUrlServiceCount)
}

func DataAdd(data *MapUrlServiceCount, service string, url string) {
	if _, ok := (*data)[url]; !ok {
		(*data)[url] = make(MapServiceCount)
	}

	(*data)[url][service] = COUNT_NO_VALUE
}

func FillData(u utils.Utils, data *MapUrlServiceCount) {
	fillDataFromCache(u, data, true)
	requests := buildRequests(u, data, true)
	executeRequests(u, &requests, data)
	scheduleRefreshIfNeeded(u)

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

func Refresh(u utils.Utils, data *MapUrlServiceCount) {
	utils.Verbosef(u, "service.Refresh(%s)", data)

	requests := buildRequests(u, data, false)
	executeRequests(u, &requests, data)
}

func fillDataFromCache(u utils.Utils, data *MapUrlServiceCount, handleNoValueOnly bool) {
	for url, services := range *data {
		for service, count := range services {
			if handleNoValueOnly && count != COUNT_NO_VALUE {
				continue
			}

			if count, ttlLeft, err := getCache(u, getCacheKey(service, url)); err == nil {
				(*data)[url][service] = count
				trackDataNeedRefresh(u, service, url, count, ttlLeft)
				continue
			}
		}
	}
}

func buildRequests(u utils.Utils, data *MapUrlServiceCount, handleNoValueOnly bool) MapServiceRequest {
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

			// temporary mark the cached count as fresh to avoid other process
			// also trying to refresh it, we will take care of it later
			temporaryCount := count
			if temporaryCount == COUNT_NO_VALUE {
				temporaryCount = COUNT_INITIAL_VALUE
			}
			setCache(u, getCacheKey(service, url), temporaryCount)

			if req, ok := requests[service]; ok {
				req.Urls = append(req.Urls, url)
				requests[service] = req
			} else {
				var newReq request
				newReq.Service = service
				newReq.Worker = worker
				newReq.Urls = []string{url}
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
				setCache(u, cacheKey, res.Count)

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

func trackDataNeedRefresh(u utils.Utils, service string, url string, cacheCount int64, cacheTtlLeft int64) {
	if cacheTtlLeft > utils.ConfigGetIntWithDefault(u, "REFRESH_TTL_LEFT_THRESHOLD", 10) {
		return
	}

	if dataNeedRefresh == nil {
		newData := make(MapUrlServiceCount)
		dataNeedRefresh = &newData
	}
	DataAdd(dataNeedRefresh, service, url)
	(*dataNeedRefresh)[url][service] = cacheCount

	// intentionally do not count unique url because if one url got flagged
	// multiple times, it should be refreshed anyway
	dataNeedRefreshCount++

	// temporary mark the cached count as fresh to avoid other process
	// also trying to refresh it, we will take care of it later
	setCache(u, getCacheKey(service, url), cacheCount)
}

func scheduleRefreshIfNeeded(u utils.Utils) {
	if dataNeedRefresh == nil {
		return
	}

	if dataNeedRefreshCount < utils.ConfigGetIntWithDefault(u, "REFRESH_BATCH_SIZE", 20) {
		return
	}

	u.Schedule("refresh", dataNeedRefresh)
	dataNeedRefresh = nil
	dataNeedRefreshCount = 0
}

func getCacheKey(service string, url string) string {
	return fmt.Sprintf("%s/%s", service, url)
}

func setCache(u utils.Utils, key string, count int64) error {
	ttl := utils.ConfigGetTtlDefault(u)
	if count < 1 {
		ttl = utils.ConfigGetIntWithDefault(u, "TTL_RESTRICTED", 60)
	}
	value := fmt.Sprintf("%d;%d", count, time.Now().Unix() + ttl)
	ttlMemory := utils.ConfigGetIntWithDefault(u, "TTL_MEMORY", 86400)

	utils.Verbosef(u, "u.MemorySet(%s, %s, %d)", key, value, ttlMemory)

	return u.MemorySet(key, value, ttlMemory)
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

	utils.Verbosef(u, "u.MemoryGet(%s) = (%d, %d, %v)", key, count, ttlLeft, e)

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
