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
var gaeConfigDatastoreKind = "Config"

type GaeConfig struct {
	Value   string    `datastore:"value,noindex"`
	Modifed time.Time `datastore:"modified,noindex"`
}

var gaeHistoryRecordKind = "HistoryRecord"

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

	k := datastore.NewKey(u.context, gaeConfigDatastoreKind, key, 0, nil)
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
	k := datastore.NewKey(u.context, gaeConfigDatastoreKind, key, 0, nil)
	datastore.Get(u.context, k, &e)

	u.Infof("Loaded via datastore config[%s] = %q, modified = %s", key, e.Value, e.Modifed)
	gaeConfigs[key] = e.Value

	return e.Value
}

func (u GAE) MemorySet(key string, value string, ttl int64) error {
	item := &memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: time.Duration(ttl) * time.Second,
	}

	return memcache.Set(u.context, item)
}

func (u GAE) MemoryGet(key string) (string, error) {
	item, err := memcache.Get(u.context, key)
	if err != nil {
		return "", err
	}

	return string(item.Value), nil
}

func (u GAE) HistorySave(service string, url string, count int64) error {
	var r HistoryRecord
	r.Service = service
	r.Url = url
	r.Count = count
	r.Time = time.Now()

	k := datastore.NewIncompleteKey(u.context, gaeHistoryRecordKind, nil)
	if _, err := datastore.Put(u.context, k, &r); err != nil {
		return err
	}

	return nil
}

func (u GAE) HistoryLoad(url string) ([]HistoryRecord, error) {
	records := []HistoryRecord{}

	q := datastore.NewQuery(gaeHistoryRecordKind).
		Filter("url =", url).
		Order("time")

	for t := q.Run(u.context);; {
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

func (u GAE) Schedule(task string, data interface{}, delay int64) error {
	json, err := json.Marshal(data)
	if err != nil {
		return err
	}

	t := taskqueue.Task{
		Delay: time.Duration(delay) * time.Second,
		Path: "/tasks/" + task,
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
