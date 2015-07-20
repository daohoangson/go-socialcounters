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

func tryMemcache(w http.ResponseWriter, r *http.Request, dataName string,
	dataFunc func(*http.Request, *http.Client, []services.ServiceFunc) (string, error),
	writeFunc func(http.ResponseWriter, *http.Request, string)) {
	c := appengine.NewContext(r)
	ttl := web.JsTtl(r)
	memcacheKey := r.RequestURI
	var data string

	if item, err := memcache.Get(c, memcacheKey); err != nil {
		client := urlfetch.Client(c)
		data, err = dataFunc(r, client, serviceFuncs)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			c.Debugf("Could not prepare %s %v", dataName, err)
			return
		}

		item := &memcache.Item{
			Key: memcacheKey,
			Value: []byte(data),
			Expiration: time.Duration(ttl) * time.Second,
		}
		memcache.Add(c, item);
	} else {
		data = string(item.Value)
	}

	writeFunc(w, r, data)
}

func allJs(w http.ResponseWriter, r *http.Request) {
	tryMemcache(w, r, "all.js", web.AllJs, web.JsWrite)
}

func dataJson(w http.ResponseWriter, r *http.Request) {
	tryMemcache(w, r, "data.json", web.DataJson, web.JsonWrite)
}

func init() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", allJs)
	http.HandleFunc("/js/data.json", dataJson)
	http.HandleFunc("/js/jquery.plugin.js", web.JQueryPluginJs)
}