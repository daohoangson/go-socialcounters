package utils

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

type gaeConfigMap map[string]string

var gaeConfigCached gaeConfigMap = nil

func gaeConfigGet(key string) string {
	if value, ok := gaeConfigCached[key]; ok {
		return value
	} else {
		return ""
	}
}

const GAE_DATASTORE_KIND_CONFIG = "Config"
const GAE_DATASTORE_KIND_HISTORY_RECORD = "HistoryRecord"
const GAE_DELAY_KEY = "go-socialcounters"

type GaeConfig struct {
	Value   string    `datastore:"value,noindex"`
	Modifed time.Time `datastore:"modified,noindex"`
}

type GAE struct {
	context context.Context
}

func GaeNew(r *http.Request) Utils {
	utils := new(GAE)
	utils.context = appengine.NewContext(r)

	return utils
}

func (u GAE) HttpGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := urlfetch.Client(u.context)
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (u GAE) ConfigSet(key string, value string) error {
	configSecret := os.Getenv("CONFIG_SECRET")
	if len(configSecret) < 1 {
		return errors.New("Env var CONFIG_SECRET must be configured to use ConfigSet")
	}

	var e GaeConfig
	e.Value = value
	e.Modifed = time.Now()

	k := datastore.NewKey(u.context, GAE_DATASTORE_KIND_CONFIG, key, 0, nil)
	if _, err := datastore.Put(u.context, k, &e); err != nil {
		return err
	}

	gaeConfigCached = nil
	u.Infof("Saved config[%s] = %q", key, value)
	return nil
}

func (u GAE) ConfigGet(key string) string {
	if gaeConfigCached != nil {
		return gaeConfigGet(key)
	}

	configSecret := os.Getenv("CONFIG_SECRET")
	if len(configSecret) < 1 {
		env := os.Getenv(key)
		if (key != configKeyVerbose) {
			Verbosef(u, "Loaded via env config[%s] = %q", key, env)
		}
		return env
	}

	gaeConfigCached = make(gaeConfigMap)

	q := datastore.NewQuery(GAE_DATASTORE_KIND_CONFIG)
	for i := q.Run(u.context); ; {
		var c GaeConfig
		gaeKey, err := i.Next(&c)
		if err == datastore.Done {
			break
		}
		if err == nil {
			u.Infof("Requested %s, loaded via datastore config[%s] = %q, modified = %s", key, gaeKey.StringID(), c.Value, c.Modifed)
			gaeConfigCached[gaeKey.StringID()] = c.Value
		} else {
			u.Errorf("ConfigGet datastore.Query.Run() err=%v", err)
			break
		}
	}

	return gaeConfigGet(key)
}

var gaeDelayFunc = delay.Func(GAE_DELAY_KEY, func(c context.Context, delayFuncArgs ...interface{}) error {
	handlerName, ok := delayFuncArgs[0].(string)
	if !ok {
		log.Errorf(c, "GAE.Delay: handler name could not be extracted from %v", delayFuncArgs)
		return nil
	}

	handler, ok := DelayHandlers[handlerName]
	if !ok {
		log.Errorf(c, "GAE.Delay: handler %s could not be found", handlerName)
		return nil
	}

	u := new(GAE)
	u.context = c
	args := delayFuncArgs[1:]
	Verbosef(u, "GAE.Delay: executing %s(%v)", handlerName, &args)

	return handler(u, args...)
})

func (u GAE) Delay(handlerName string, args ...interface{}) error {
	delayFuncArgs := append([]interface{}{handlerName}, args...)
	gaeDelayFunc.Call(u.context, delayFuncArgs...)
	Verbosef(u, "GAE.Delay: delaying %s(%v)", handlerName, &args)

	return nil
}

func (u GAE) MemorySet(items *[]MemoryItem) error {
	if items == nil || len(*items) < 1 {
		return nil
	}

	gaeItems := make([]*memcache.Item, len(*items))
	for index, item := range *items {
		gaeItems[index] = &memcache.Item{
			Key:        item.Key,
			Value:      []byte(item.Value),
			Expiration: time.Duration(item.Ttl) * time.Second,
		}

		Verbosef(u, "GAE.MemorySet item[%d] = %v", index, item)
	}

	return memcache.SetMulti(u.context, gaeItems)
}

func (u GAE) MemoryGet(items *[]MemoryItem) error {
	if items == nil || len(*items) < 1 {
		return nil
	}

	keys := make([]string, len(*items))
	for index, item := range *items {
		keys[index] = item.Key
	}

	gaeItems, err := memcache.GetMulti(u.context, keys)
	if err != nil {
		return err
	}

	for index, item := range *items {
		if gaeItem, ok := gaeItems[item.Key]; ok {
			(*items)[index].Value = string(gaeItem.Value)
		}
	}

	return nil
}

func (u GAE) HistorySave(records *[]HistoryRecord) error {
	if records == nil || len(*records) < 1 {
		return nil
	}

	keys := make([]*datastore.Key, len(*records))
	src := make([]*HistoryRecord, len(*records))
	for index, _ := range *records {
		keys[index] = datastore.NewIncompleteKey(u.context, GAE_DATASTORE_KIND_HISTORY_RECORD, nil)
		src[index] = &(*records)[index]

		Verbosef(u, "GAE.HistorySave src[%d] = %v", index, src[index])
	}

	if _, err := datastore.PutMulti(u.context, keys, src); err != nil {
		return err
	}

	return nil
}

func (u GAE) HistoryLoad(url string) ([]HistoryRecord, error) {
	records := []HistoryRecord{}

	q := datastore.NewQuery(GAE_DATASTORE_KIND_HISTORY_RECORD).
		Filter("url =", url).
		Order("time")

	for t := q.Run(u.context); ; {
		var r HistoryRecord
		_, err := t.Next(&r)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return records, err
		}

		records = append(records, r)
	}

	return records, nil
}

func (u GAE) Errorf(format string, args ...interface{}) {
	log.Errorf(u.context, format, args...)
}

func (u GAE) Infof(format string, args ...interface{}) {
	log.Infof(u.context, format, args...)
}

func (u GAE) Debugf(format string, args ...interface{}) {
	log.Debugf(u.context, format, args...)
}
