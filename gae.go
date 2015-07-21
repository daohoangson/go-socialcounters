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

func getCountsJson(c appengine.Context, r *http.Request) (string, error) {
	ttl := web.JsTtl(r)
	var data string

	url, err := web.GetUrl(r)
	if err != nil {
		return "", err
	}

	if item, err := memcache.Get(c, url); err != nil {
		client := urlfetch.Client(c)
		data, err = web.CountsJson(r, client, serviceFuncs)
		if (err != nil) {
			c.Debugf("web.CountsJson error %v", err)
		} else {
			item := &memcache.Item{
				Key: url,
				Value: []byte(data),
				Expiration: time.Duration(ttl) * time.Second,
			}
			memcache.Add(c, item);
		}
	} else {
		data = string(item.Value)
	}

	return data, nil
}

func allJs(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	countsJson, err := getCountsJson(c, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.Debugf("Could not getCountsJson %v", err)
		return
	}

	js, err := web.AllJs(r, countsJson)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.Debugf("Could not get web.AllJs %v", err)
		return
	}

	web.JsWrite(w, r, js)
}

func dataJson(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	countsJson, err := getCountsJson(c, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.Debugf("Could not getCountsJson %v", err)
		return
	}

	web.JsonWrite(w, r, countsJson)
}

func init() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", allJs)
	http.HandleFunc("/js/data.json", dataJson)
	http.HandleFunc("/js/jquery.plugin.js", web.JQueryPluginJs)
}