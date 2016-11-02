// +build !appengine

package utils

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/bmizerany/mc"
)

type Other struct {
}

func OtherNew(r *http.Request) Utils {
	utils := new(Other)

	return utils
}

func (u Other) HttpClient() *http.Client {
	return &http.Client{}
}

func (u Other) ConfigSet(key string, value string) error {
	return errors.New("Not implemented")
}

func (u Other) ConfigGet(key string) string {
	return os.Getenv(key)
}

func (u Other) MemorySet(key string, value string, ttl int64) error {
	conn := getMcConn(u)
	if conn == nil {
		return nil
	}

	return conn.Set(key, value, 0, 0, int(ttl))
}

func (u Other) MemoryGet(key string) (string, error) {
	conn := getMcConn(u)
	if conn == nil {
		return "", errors.New("No memcache connection")
	}

	value, _, _, err := conn.Get(key)
	return value, err
}

func (u Other) HistorySave(service string, url string, count int64) error {
	return errors.New("Not implemented")
}

func (u Other) HistoryLoad(url string) ([]HistoryRecord, error) {
	return nil, errors.New("Not implemented")
}

func (u Other) Schedule(task string, data interface{}, delay int64) error {
	return errors.New("Not implemented")
}

func (u Other) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (u Other) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (u Other) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

var mcConn *mc.Conn
var mcPrepared = false

func getMcConn(u Other) *mc.Conn {
	if !mcPrepared {
		if addr := os.Getenv("MEMCACHIER_SERVERS"); addr != "" {
			if m, err := mc.Dial("tcp", addr); err == nil {
				username := os.Getenv("MEMCACHIER_USERNAME")
				password := os.Getenv("MEMCACHIER_PASSWORD")

				if username != "" && password != "" {
					// only try to authenticate if both username and password are set
					err = m.Auth(os.Getenv("MEMCACHIER_USERNAME"), os.Getenv("MEMCACHIER_PASSWORD"))
					if err == nil {
						u.Infof("Other.getMcConn: mc.Auth ok")
						mcConn = m
					} else {
						u.Errorf("Other.getMcConn: mc.Auth error %v", err)
					}
				} else {
					// most of the case, the server does not require authentication
					u.Infof("Other.getMcConn: mc.Dial ok")
					mcConn = m
				}
			} else {
				u.Errorf("Other.getMcConn: mc.Dial error %v", err)
			}
		}

		mcPrepared = true
	}

	return mcConn
}
