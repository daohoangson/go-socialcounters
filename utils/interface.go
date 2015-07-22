package utils

import (
	"net/http"

	"github.com/daohoangson/go-socialcounters/services"
)

type Utils interface {

	ServiceFuncs() []services.ServiceFunc

	HttpClient() *http.Client

	MemorySet(key string, value []byte, ttl int64) error
	MemoryGet(key string) ([]byte, error)

	DbSet(key string, hash map[string]string) error
	DbGet(key string) (map[string]string, error)

	Logf(format string, args ...interface{})

}

type UtilsFunc func(w http.ResponseWriter, r *http.Request) Utils