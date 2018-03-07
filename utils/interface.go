package utils

import (
	"net/http"
	"time"
)

type Utils interface {
	HttpGet(url string) ([]byte, error)

	Delay(handlerName string, args ...interface{}) error

	ConfigSet(key string, value string) error
	ConfigGet(key string) string

	MemorySet(items *[]MemoryItem) error
	MemoryGet(items *[]MemoryItem) error

	HistorySave(records *[]HistoryRecord) error
	HistoryLoad(url string) ([]HistoryRecord, error)

	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type UtilsFunc func(w http.ResponseWriter, r *http.Request) Utils

type DelayHandler func(u Utils, args ...interface{}) error

type MemoryItem struct {
	Key   string
	Value string
	Ttl   int64
}

type HistoryRecord struct {
	Service string    `datastore:"service,noindex"`
	Url     string    `datastore:"url"`
	Count   int64     `datastore:"count,noindex"`
	Time    time.Time `datastore:"time"`
}
