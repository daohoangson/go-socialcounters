package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"appengine"
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

	jsData, err := ioutil.ReadFile("private/js/all.js")
	if (err != nil) {
		w.WriteHeader(http.StatusNotFound)
		c.Debugf("Could not read all.js")
		return
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
	}

	countsJson, err := json.Marshal(counts)
	if (err != nil) {
		c.Debugf("Could not marshal count map")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	js = strings.Replace(js, "{counts}", string(countsJson), 1)

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public; max-age=300")
	fmt.Fprintf(w, js)
}

func init() {
	http.HandleFunc("/js/all.js", alljs)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
}