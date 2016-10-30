// +build appengine

package utils

import (
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
)

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

func (u GAE) MemorySet(key string, value []byte, ttl int64) error {
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: time.Duration(ttl) * time.Second,
	}

	return memcache.Add(u.context, item)
}

func (u GAE) MemoryGet(key string) ([]byte, error) {
	item, err := memcache.Get(u.context, key)
	if err != nil {
		return nil, err
	}

	return item.Value, nil
}

func (u GAE) DbSet(key string, hash map[string]string) error {
	return errors.New("Not implemented")
}

func (u GAE) DbGet(key string) (map[string]string, error) {
	return nil, errors.New("Not implemented")
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
