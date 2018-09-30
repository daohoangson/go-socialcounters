package services

import (
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daohoangson/go-socialcounters/utils"
)

const delayHandlerNameRefresh = "services.Refresh"

var workers = map[string]worker{
	serviceFacebook: facebookWorker,
	serviceTwitter:  twitterWorker,
}

var dataNeedRefresh *mapURLServiceCount
var dataNeedRefreshCount = int64(0)

// Init sets up global environment
func Init() {
	utils.DelayHandlers[delayHandlerNameRefresh] = refreshHandler

	// for GAE
	gob.Register(mapURLServiceCount{})
}

// DataSetup prepares data before processing
func DataSetup() mapURLServiceCount {
	return make(mapURLServiceCount)
}

// DataAdd includes the specified service / url combination for processing
func DataAdd(data *mapURLServiceCount, service string, url string) {
	if _, ok := (*data)[url]; !ok {
		(*data)[url] = make(mapServiceCount)
	}

	(*data)[url][service] = countNoValue
}

// FillData uses values from caches and databases to return values as fast as possible,
// if no values are usable, it will go and fetch from remote services
func FillData(u utils.Utils, data *mapURLServiceCount) {
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

// Refresh executes fetch from remote services
func Refresh(u utils.Utils, data *mapURLServiceCount) {
	requests := buildRequests(u, data, false)
	executeRequests(u, &requests, data)
}

func fillDataFromCache(u utils.Utils, data *mapURLServiceCount, handleNoValueOnly bool) {
	caches := make(sliceCache, 0)

	for url, services := range *data {
		for service, count := range services {
			if handleNoValueOnly && count != countNoValue {
				continue
			}

			caches = append(caches, cache{Service: service, URL: url, Count: countNoValue})
		}
	}

	if err := getCaches(u, &caches); err != nil {
		u.Errorf("services.getCaches(%d) error %v", len(caches), err)
	} else {
		for _, cache := range caches {
			(*data)[cache.URL][cache.Service] = cache.Count
		}

		checkCachesForRefresh(u, &caches)
	}
}

func buildRequests(u utils.Utils, data *mapURLServiceCount, handleNoValueOnly bool) mapServiceRequest {
	utils.Verbosef(u, "services.buildRequests(%s, %s)", data, handleNoValueOnly)

	requests := make(mapServiceRequest)
	caches := make(sliceCache, 0)

	for url, services := range *data {
		for service, count := range services {
			if handleNoValueOnly && count != countNoValue {
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
			if temporaryCount == countNoValue {
				temporaryCount = countInitValue
			}
			caches = append(caches, cache{Service: service, URL: url, Count: temporaryCount})

			if req, ok := requests[service]; ok {
				req.Urls = append(req.Urls, url)
				requests[service] = req
			} else {
				var newReq request
				newReq.Service = service
				newReq.Worker = worker
				newReq.Urls = []string{url}
				newReq.Results = make(mapURLResult)

				requests[service] = newReq
			}
		}
	}

	setCaches(u, &caches)

	return requests
}

func executeRequests(u utils.Utils, requests *mapServiceRequest, data *mapURLServiceCount) {
	if len(*requests) < 1 {
		return
	}
	utils.Verbosef(u, "services.executeRequests(%s)", requests)

	var wg sync.WaitGroup
	var cacheWg sync.WaitGroup
	cacheCh := make(chan cache, 1)
	caches := make(sliceCache, 0)

	historySave := utils.ConfigGetIntWithDefault(u, "HISTORY_SAVE", 0) > 0
	var historyWg sync.WaitGroup
	historyTime := time.Now()
	historyCh := make(chan utils.HistoryRecord, 1)
	histories := make([]utils.HistoryRecord, 0)

	wg.Add(len(*requests))
	for _, req := range *requests {
		go func(req request) {
			defer wg.Done()

			req.Worker(u, &req)
			if req.Error != nil {
				u.Errorf("services.%s: %v", req.Service, req.Error)
			}

			for url, res := range req.Results {
				oldCount, _ := (*data)[url][req.Service]
				(*data)[url][req.Service] = res.Count

				if res.Error != nil {
					u.Errorf("services.%s: %s error %v response %s", req.Service, url, res.Error, res.Response)
				} else {
					if res.Count > countInitValue && res.Count > oldCount {
						cacheWg.Add(1)
						cacheCh <- cache{Service: req.Service, URL: url, Count: res.Count}

						if historySave {
							historyWg.Add(1)
							historyCh <- utils.HistoryRecord{Service: req.Service, Url: url, Count: res.Count, Time: historyTime}
						}
					}
				}
			}
		}(req)
	}

	go func() {
		for cache := range cacheCh {
			caches = append(caches, cache)
			cacheWg.Done()
		}
	}()

	wg.Wait()

	cacheWg.Wait()
	setCaches(u, &caches)

	if historySave {
		go func() {
			for history := range historyCh {
				histories = append(histories, history)
				historyWg.Done()
			}
		}()

		historyWg.Wait()

		if err := u.HistorySave(&histories); err != nil {
			u.Errorf("u.HistorySave error %v", err)
		}
	} else {
		utils.Verbosef(u, "Skipped saving history")
	}
}

func checkCachesForRefresh(u utils.Utils, caches *sliceCache) {
	newCaches := make(sliceCache, 0)
	ttlLeftThreshold := utils.ConfigGetIntWithDefault(u, "REFRESH_TTL_LEFT_THRESHOLD", 10)

	for _, c := range *caches {
		if c.TTLLeft > ttlLeftThreshold {
			continue
		}

		if c.Count < countInitValue {
			continue
		}

		if dataNeedRefresh == nil {
			newData := make(mapURLServiceCount)
			dataNeedRefresh = &newData
		}
		DataAdd(dataNeedRefresh, c.Service, c.URL)
		(*dataNeedRefresh)[c.URL][c.Service] = c.Count

		// intentionally do not count unique url because if one url got flagged
		// multiple times, it should be refreshed anyway
		dataNeedRefreshCount++

		// temporary mark the cached count as fresh to avoid other process
		// also trying to refresh it, we will take care of it later
		newCaches = append(newCaches, cache{Service: c.Service, URL: c.URL, Count: c.Count})
	}

	setCaches(u, &newCaches)
}

func scheduleRefreshIfNeeded(u utils.Utils) {
	if dataNeedRefresh == nil {
		return
	}

	if dataNeedRefreshCount < utils.ConfigGetIntWithDefault(u, "REFRESH_BATCH_SIZE", 20) {
		return
	}

	u.Delay(delayHandlerNameRefresh, *dataNeedRefresh)
	dataNeedRefresh = nil
	dataNeedRefreshCount = 0
}

func refreshHandler(u utils.Utils, args ...interface{}) error {
	data, ok := args[0].(mapURLServiceCount)
	if !ok {
		return fmt.Errorf("services.refreshHandler: data could not be extracted from %v", args)
	}
	Refresh(u, &data)

	return nil
}

func getCacheKey(c cache) string {
	return fmt.Sprintf("%s/%s", c.Service, c.URL)
}

func setCaches(u utils.Utils, caches *sliceCache) {
	if caches == nil || len(*caches) < 1 {
		return
	}

	items := make([]utils.MemoryItem, len(*caches))
	ttlMemory := utils.ConfigGetIntWithDefault(u, "TTL_MEMORY", 86400)
	ttlDefault := utils.ConfigGetTTLDefault(u)
	ttlRestricted := utils.ConfigGetIntWithDefault(u, "TTL_RESTRICTED", 60)

	for index, cache := range *caches {
		valueTTL := ttlDefault
		if cache.Count < 1 {
			valueTTL = ttlRestricted
		}

		items[index] = utils.MemoryItem{
			Key:   getCacheKey(cache),
			Value: fmt.Sprintf("%d;%d", cache.Count, time.Now().Unix()+valueTTL),
			Ttl:   ttlMemory,
		}
	}

	if err := u.MemorySet(&items); err != nil {
		u.Errorf("u.MemorySet(%d) error %v", len(items), err)
	}
}

func getCaches(u utils.Utils, caches *sliceCache) error {
	if caches == nil || len(*caches) < 1 {
		return nil
	}

	items := make([]utils.MemoryItem, len(*caches))

	for index, cache := range *caches {
		items[index] = utils.MemoryItem{Key: getCacheKey(cache)}
	}

	if err := u.MemoryGet(&items); err != nil {
		u.Errorf("u.MemoryGet(%d) error %v", len(items), err)
		return err
	}

	for index, item := range items {
		parts := strings.Split(item.Value, ";")

		if count, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			(*caches)[index].Count = count
		}

		if len(parts) >= 2 {
			// value with expire timestamp
			if exp, expErr := strconv.ParseInt(parts[1], 10, 64); expErr == nil {
				(*caches)[index].TTLLeft = exp - time.Now().Unix()
			}
		}

		utils.Verbosef(u, "services.getCaches caches[%d] = %v", index, &(*caches)[index])
	}

	return nil
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
