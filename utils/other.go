// +build !appengine

package utils

import (
	"errors"
	"log"
	"net/http"

	"github.com/daohoangson/go-socialcounters/services"
)

type Other struct {
}

func OtherNew(r *http.Request) Utils {
	utils := new(Other)

	return utils
}

var serviceFuncs = []services.ServiceFunc{
	services.Facebook2,
	services.Twitter,
	services.Google,
}
func (u Other) ServiceFuncs() []services.ServiceFunc {
	return serviceFuncs
}

func (u Other) HttpClient() *http.Client {
	return &http.Client{}
}

func (u Other) MemorySet(key string, value []byte, ttl int64) error {
	return errors.New("Not implemented")
}

func (u Other) MemoryGet(key string) ([]byte, error) {
	return nil, errors.New("Not implemented")
}

func (u Other) DbSet(key string, hash map[string]string) error {
	return errors.New("Not implemented")
}

func (u Other) DbGet(key string) (map[string]string, error) {
	return nil, errors.New("Not implemented")
}

func (u Other) Logf(format string, args ...interface{}) {
	log.Printf(format, args...)
}