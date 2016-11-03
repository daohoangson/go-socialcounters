package utils

import (
	"net/http"
	"time"
)

type Utils interface {
	HttpClient() *http.Client

	ConfigSet(key string, value string) error
	ConfigGet(key string) string

	MemorySet(key string, value string, ttl int64) error
	MemoryGet(key string) (string, error)

	HistorySave(service string, url string, count int64) error
	HistoryLoad(url string) ([]HistoryRecord, error)

	Schedule(task string, data interface{}) error

	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type UtilsFunc func(w http.ResponseWriter, r *http.Request) Utils

type HistoryRecord struct {
	Service string    `datastore:"service,noindex"`
	Url     string    `datastore:"url"`
	Count   int64     `datastore:"count,noindex"`
	Time    time.Time `datastore:"time"`
}
