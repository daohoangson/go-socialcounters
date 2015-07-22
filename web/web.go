package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/daohoangson/go-minify/css"
	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/utils"
)

func AllJs(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	url, countsJson, err := getCountsJson(u, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Logf("web.AllJs: getCountsJson error %v", err)
		return
	}

	jsData, err := ioutil.ReadFile("private/js/all.js")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Logf("web.AllJs: ReadFile error %v", err)
		return
	}
	js := MinifyJs(string(jsData))
	js = strings.Replace(js, "{url}", url, 1)
	js = strings.Replace(js, "{now}", fmt.Sprintf("%v", time.Now()), 1)

	// keep using css.MinifyFromFile because it does the data uri inline for us
	// TODO: drop this github.com/daohoangson/go-minify/css dependency
	css := css.MinifyFromFile("public/css/main.css")
	js = strings.Replace(js, "{css}", css, 1)

	js = strings.Replace(js, "{counts}", string(countsJson), 1)
	js = strings.Replace(js, "{target}", parseTargetAsJson(r), 1)

	writeJs(w, r, js)
}

func DataJson(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	_, countsJson, err := getCountsJson(u, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Logf("web.DataJson: getCountsJson error %v", err)
		return
	}

	writeJson(w, r, countsJson)
}

func JqueryPluginJs(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	jsData, err := ioutil.ReadFile("private/js/jquery.plugin.js")
	if (err != nil) {
		w.WriteHeader(http.StatusNotFound)
		u.Logf("web.JqueryPluginJs: ReadFile error %v", err)
		return
	}

	jsonUrl := fmt.Sprintf("//%s/js/data.json", r.Host)
	js := strings.Replace(string(jsData), "{jsonUrl}", jsonUrl, 1)

	writeJs(w, r, js)
}

func getCountsJson(u utils.Utils, r *http.Request) (string, string, error) {
	url := parseUrl(r)
	if !RulesAllowUrl(url) {
		return url, "{}", nil
	}

	if value, err := u.MemoryGet(url); err == nil {
		return url, string(value), nil
	}

	serviceResults := services.Batch(u.HttpClient(), u.ServiceFuncs(), url)
	dataMap := make(map[string]int64)
	for _, serviceResult := range serviceResults {
		dataMap[serviceResult.Service] = serviceResult.Count
	}
	
	dataByte, err := json.Marshal(dataMap)
	if err != nil {
		return url, "{}", err
	} else {
		u.MemorySet(url, dataByte, parseTtl(r))
	}

	return url, string(dataByte), nil
}

func parseTargetAsJson(r *http.Request) string {
	target := "'.socialcounters-container'";

	q := r.URL.Query()
	if targets, ok := q["target"]; ok {
		targetByte, err := json.Marshal(targets[0])
		if err == nil {
			target = string(targetByte);
		}
	}

	return target;
}

func parseTtl(r *http.Request) int64 {
	q := r.URL.Query()
	if ttls, ok := q["ttl"]; ok {
		ttl, err := strconv.ParseInt(ttls[0], 10, 64)
		if err == nil {
			return ttl
		}
	}

	return 300
}

func parseUrl(r *http.Request) string {
	q := r.URL.Query()
	if urls, ok := q["url"]; ok {
		return urls[0]
	}

	return ""
}

func writeJs(w http.ResponseWriter, r *http.Request, js string) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", parseTtl(r)))
	fmt.Fprintf(w, js)
}

func writeJson(w http.ResponseWriter, r *http.Request, json string) {
	q := r.URL.Query()
	var callback string
	if callbacks, ok := q["callback"]; ok {
		callback = callbacks[0]
	}

	if len(callback) > 0 {
		js := fmt.Sprintf("%s(%s);", callback, json);
		writeJs(w, r, js)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", parseTtl(r)))
		fmt.Fprintf(w, json)
	}
}