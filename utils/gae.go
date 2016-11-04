// +build appengine

package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/taskqueue"
	"appengine/urlfetch"
)

var gaeConfigs = make(map[string]string)

const GAE_DATASTORE_KIND_CONFIG = "Config"
const GAE_DATASTORE_KIND_HISTORY_RECORD = "HistoryRecord"

type GaeConfig struct {
	Value   string    `datastore:"value,noindex"`
	Modifed time.Time `datastore:"modified,noindex"`
}

type GAE struct {
	context appengine.Context
}

func GaeNew(r *http.Request) Utils {
	utils := new(GAE)
	utils.context = appengine.NewContext(r)

	return utils
}

func (u GAE) HttpClient() *http.Client {
	return urlfetch.Client(u.context)
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

	gaeConfigs = make(map[string]string)
	u.Infof("Saved config[%s] = %q", key, value)
	return nil
}

func (u GAE) ConfigGet(key string) string {
	if value, ok := gaeConfigs[key]; ok {
		return value
	}

	configSecret := os.Getenv("CONFIG_SECRET")
	if len(configSecret) < 1 {
		env := os.Getenv(key)
		u.Infof("Loaded via env config[%s] = %q", key, env)
		gaeConfigs[key] = env

		return env
	}

	var e GaeConfig
	k := datastore.NewKey(u.context, GAE_DATASTORE_KIND_CONFIG, key, 0, nil)
	datastore.Get(u.context, k, &e)

	u.Infof("Loaded via datastore config[%s] = %q, modified = %s", key, e.Value, e.Modifed)
	gaeConfigs[key] = e.Value

	return e.Value
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

func (u GAE) Schedule(task string, data interface{}) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	t := taskqueue.Task{
		Path:    "/tasks/" + task,
		Payload: json,
	}
	if _, err := taskqueue.Add(u.context, &t, ""); err != nil {
		return err
	}

	return nil
}

func (u GAE) Errorf(format string, args ...interface{}) {
	u.context.Errorf(format, args...)
}

func (u GAE) Infof(format string, args ...interface{}) {
	u.context.Infof(format, args...)
}

func (u GAE) Debugf(format string, args ...interface{}) {
	u.context.Debugf(format, args...)
}
