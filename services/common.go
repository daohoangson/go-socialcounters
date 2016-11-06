package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daohoangson/go-socialcounters/utils"
)

const DELAY_HANDLER_NAME_REFRESH = "services.Refresh";

var workers = map[string]worker{
	SERVICE_FACEBOOK: facebookWorker,
	SERVICE_TWITTER:  twitterWorker,
	SERVICE_GOOGLE:   googleWorker,
}

var dataNeedRefresh *MapUrlServiceCount
var dataNeedRefreshCount = int64(0)

func Init() {
	utils.DelayHandlers[DELAY_HANDLER_NAME_REFRESH] = refreshHandler
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
	requests := buildRequests(u, data, false)
	executeRequests(u, &requests, data)
}

func fillDataFromCache(u utils.Utils, data *MapUrlServiceCount, handleNoValueOnly bool) {
	caches := make(sliceCache, 0)

	for url, services := range *data {
		for service, count := range services {
			if handleNoValueOnly && count != COUNT_NO_VALUE {
				continue
			}

			caches = append(caches, cache{Service: service, Url: url, Count: COUNT_NO_VALUE})
		}
	}

	if err := getCaches(u, &caches); err != nil {
		u.Errorf("services.getCaches(%d) error %v", len(caches), err)
	} else {
		for _, cache := range caches {
			(*data)[cache.Url][cache.Service] = cache.Count
		}

		checkCachesForRefresh(u, &caches)
	}
}

func buildRequests(u utils.Utils, data *MapUrlServiceCount, handleNoValueOnly bool) MapServiceRequest {
	utils.Verbosef(u, "services.buildRequests(%s, %s)", data, handleNoValueOnly)

	requests := make(MapServiceRequest)
	caches := make(sliceCache, 0)

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
			caches = append(caches, cache{Service: service, Url: url, Count: temporaryCount})

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

	setCaches(u, &caches)

	return requests
}

func executeRequests(u utils.Utils, requests *MapServiceRequest, data *MapUrlServiceCount) {
	utils.Verbosef(u, "services.executeRequests(%s)", requests)
	if len(*requests) < 1 {
		return
	}

	var wg sync.WaitGroup
	var cacheWg sync.WaitGroup
	cacheCh := make(chan cache, 1)
	caches := make(sliceCache, 0)
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
					if res.Count > COUNT_INITIAL_VALUE && res.Count > oldCount {
						cacheWg.Add(1)
						cacheCh <- cache{Service: req.Service, Url: url, Count: res.Count}

						historyWg.Add(1)
						historyCh <- utils.HistoryRecord{Service: req.Service, Url: url, Count: res.Count, Time: historyTime}
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

	go func() {
		for history := range historyCh {
			histories = append(histories, history)
			historyWg.Done()
		}
	}()

	wg.Wait()
	cacheWg.Wait()
	historyWg.Wait()

	setCaches(u, &caches)
	if err := u.HistorySave(&histories); err != nil {
		u.Errorf("u.HistorySave error %v", err)
	}
}

func checkCachesForRefresh(u utils.Utils, caches *sliceCache) {
	newCaches := make(sliceCache, 0)
	ttlLeftThreshold := utils.ConfigGetIntWithDefault(u, "REFRESH_TTL_LEFT_THRESHOLD", 10)

	for _, c := range *caches {
		if c.TtlLeft > ttlLeftThreshold {
			continue
		}

		if c.Count < COUNT_INITIAL_VALUE {
			continue
		}

		if dataNeedRefresh == nil {
			newData := make(MapUrlServiceCount)
			dataNeedRefresh = &newData
		}
		DataAdd(dataNeedRefresh, c.Service, c.Url)
		(*dataNeedRefresh)[c.Url][c.Service] = c.Count

		// intentionally do not count unique url because if one url got flagged
		// multiple times, it should be refreshed anyway
		dataNeedRefreshCount++

		// temporary mark the cached count as fresh to avoid other process
		// also trying to refresh it, we will take care of it later
		newCaches = append(newCaches, cache{Service: c.Service, Url: c.Url, Count: c.Count})
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

	u.Delay(DELAY_HANDLER_NAME_REFRESH, *dataNeedRefresh)
	dataNeedRefresh = nil
	dataNeedRefreshCount = 0
}

func refreshHandler(u utils.Utils, args ...interface{}) error {
	data, ok := args[0].(MapUrlServiceCount)
	if !ok {
		return errors.New(fmt.Sprintf("services.refreshHandler: data could not be extracted from %v", args))
	}
	Refresh(u, &data)

	return nil
}

func getCacheKey(c cache) string {
	return fmt.Sprintf("%s/%s", c.Service, c.Url)
}

func setCaches(u utils.Utils, caches *sliceCache) {
	if caches == nil || len(*caches) < 1 {
		return
	}

	items := make([]utils.MemoryItem, len(*caches))
	ttlMemory := utils.ConfigGetIntWithDefault(u, "TTL_MEMORY", 86400)
	ttlDefault := utils.ConfigGetTtlDefault(u)
	ttlRestricted := utils.ConfigGetIntWithDefault(u, "TTL_RESTRICTED", 60)

	for index, cache := range *caches {
		valueTtl := ttlDefault
		if cache.Count < 1 {
			valueTtl = ttlRestricted
		}

		items[index] = utils.MemoryItem{
			Key:   getCacheKey(cache),
			Value: fmt.Sprintf("%d;%d", cache.Count, time.Now().Unix()+valueTtl),
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
				(*caches)[index].TtlLeft = exp - time.Now().Unix()
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
