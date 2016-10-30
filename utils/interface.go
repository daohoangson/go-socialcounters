package utils

import (
	"net/http"
)

type Utils interface {
	HttpClient() *http.Client

	MemorySet(key string, value []byte, ttl int64) error
	MemoryGet(key string) ([]byte, error)

	DbSet(key string, hash map[string]string) error
	DbGet(key string) (map[string]string, error)

	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type UtilsFunc func(w http.ResponseWriter, r *http.Request) Utils
