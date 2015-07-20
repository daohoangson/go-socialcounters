// +build appengine

package main

import (
	"net/http"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/web"
)

var serviceFuncs = []services.ServiceFunc{
	services.Facebook1,
	services.Twitter,
	services.Google,
}

func allJs(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	ttl := web.JsTtl(r)
	memcacheKey := r.RequestURI
	var js string

	if item, err := memcache.Get(c, memcacheKey); err != nil {
		client := urlfetch.Client(c)
		js, err = web.AllJs(r, client, serviceFuncs)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			c.Debugf("Could not prepare all.js %v", err)
			return
		}

		item := &memcache.Item{
			Key: memcacheKey,
			Value: []byte(js),
			Expiration: time.Duration(ttl) * time.Second,
		}
		memcache.Add(c, item);
	} else {
		js = string(item.Value)
	}

	web.JsWrite(w, r, js)
}

func init() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", allJs)
}