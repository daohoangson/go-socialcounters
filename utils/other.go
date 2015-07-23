// +build !appengine

package utils

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/bmizerany/mc"
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
	conn := getMcConn(u)
	if conn == nil {
		return nil
	}

	return conn.Set(key, string(value), 0, 0, int(ttl))
}

func (u Other) MemoryGet(key string) ([]byte, error) {
	conn := getMcConn(u)
	if conn == nil {
		return nil, errors.New("No memcache connection")
	}

	value, _, _, err := conn.Get(key)
	return []byte(value), err
}

func (u Other) DbSet(key string, hash map[string]string) error {
	return errors.New("Not implemented")
}

func (u Other) DbGet(key string) (map[string]string, error) {
	return nil, errors.New("Not implemented")
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
