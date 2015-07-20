package main

import (
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
	"socialcounters/web"
)

func gaeAllJs(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := r.URL.Query()
	var url string
	if urls, ok := q["url"]; ok {
		url = urls[0]
	}
	if len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		c.Debugf("No `url` specified for all.js")
		return
	}

	var js string
	ttl := 300
	if item, err := memcache.Get(c, url); err != nil {
		client := urlfetch.Client(c)
		js, err = web.AllJs(client, url)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
			c.Debugf("Could not prepare all.js %v", err)
			return
		}

		item := &memcache.Item{
			Key: url,
			Value: []byte(js),
			Expiration: time.Duration(ttl) * time.Second,
		}
		memcache.Add(c, item);
	} else {
		js = string(item.Value)
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", ttl))
	fmt.Fprintf(w, js)
}

func init() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", gaeAllJs)
}