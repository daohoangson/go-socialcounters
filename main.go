package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"
	"github.com/daohoangson/minify/css"
	"socialcounters/services"
)

func alljs(w http.ResponseWriter, r *http.Request) {
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
		js, err = alljsGenerate(c, url)
		if (err != nil) {
			w.WriteHeader(http.StatusInternalServerError)
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

func alljsGenerate(c appengine.Context, url string) (string, error) {
	jsData, err := ioutil.ReadFile("private/js/all.js")
	if (err != nil) {
		c.Debugf("Could not read all.js")
		return "", err
	}
	js := strings.Replace(string(jsData), "{url}", url, 1)

	css := css.MinifyFromFile("private/css/main.css")
	js = strings.Replace(js, "{css}", css, 1)

	client := urlfetch.Client(c)
	serviceResults := services.All(client, url)
	counts := make(map[string]float64)
	for _, serviceResult := range serviceResults {
		counts[serviceResult.Service] = serviceResult.Count

		if serviceResult.Error != nil {
			c.Debugf("%s error: %v", serviceResult.Service, serviceResult.Error)
		}
		if len(serviceResult.Response) > 0 {
			c.Debugf("%s response: %s", serviceResult.Service, serviceResult.Response)
		}
	}

	countsJson, err := json.Marshal(counts)
	if (err != nil) {
		c.Debugf("Could not marshal count map")
		return "", err
	}
	js = strings.Replace(js, "{counts}", string(countsJson), 1)

	return js, nil
}

func init() {
	http.HandleFunc("/js/all.js", alljs)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
}