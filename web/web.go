package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/daohoangson/go-minify/css"
	"github.com/daohoangson/go-socialcounters/services"
)

func GetUrl(r *http.Request) (string, error) {
	q := r.URL.Query()
	var url string
	if urls, ok := q["url"]; ok {
		url = urls[0]
	}
	if len(url) == 0 {
		return "", errors.New("No `url` specified for data.json")
	}

	return url, nil
}

func AllJs(r *http.Request, countsJson string) (string, error) {
	url, err := GetUrl(r);
	if err != nil {
		return "", err
	}

	jsData, err := ioutil.ReadFile("private/js/all.js")
	if err != nil {
		return "", err
	}
	js := MinifyJs(string(jsData))
	js = strings.Replace(js, "{url}", url, 1)
	js = strings.Replace(js, "{now}", fmt.Sprintf("%v", time.Now()), 1)

	// keep using css.MinifyFromFile because it does the data uri inline for us
	// TODO: drop this github.com/daohoangson/go-minify/css dependency
	css := css.MinifyFromFile("public/css/main.css")
	js = strings.Replace(js, "{css}", css, 1)

	js = strings.Replace(js, "{counts}", string(countsJson), 1)
	js = strings.Replace(js, "{target}", TargetJson(r), 1)

	return js, nil
}

func JsTtl(r *http.Request) uint64 {
	q := r.URL.Query()
	if ttls, ok := q["ttl"]; ok {
		ttl, err := strconv.ParseUint(ttls[0], 10, 64)
		if err == nil {
			return ttl
		}
	}

	return 300
}

func JsWrite(w http.ResponseWriter, r *http.Request, js string) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", JsTtl(r)))
	fmt.Fprintf(w, js)
}

func JQueryPluginJs(w http.ResponseWriter, r *http.Request) {
	jsData, err := ioutil.ReadFile("private/js/jquery.plugin.js")
	if (err != nil) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonUrl := fmt.Sprintf("//%s/js/data.json", r.Host)
	js := strings.Replace(string(jsData), "{jsonUrl}", jsonUrl, 1)

	JsWrite(w, r, js)
}

func CountsJson(r *http.Request, client *http.Client, serviceFuncs []services.ServiceFunc) (string, error) {
	url, err := GetUrl(r);
	if err != nil {
		return "{}", err
	}

	if !RulesAllowUrl(url) {
		return "{}", nil
	}

	serviceResults := services.Batch(client, serviceFuncs, url)
	dataMap := make(map[string]int64)
	var serviceError error
	for _, serviceResult := range serviceResults {
		dataMap[serviceResult.Service] = serviceResult.Count

		if serviceResult.Error != nil {
			serviceError = serviceResult.Error
		}
	}
	
	dataByte, err := json.Marshal(dataMap)
	if err != nil {
		return "{}", err
	}

	return string(dataByte), serviceError
}

func JsonWrite(w http.ResponseWriter, r *http.Request, json string) {
	q := r.URL.Query()
	var callback string
	if callbacks, ok := q["callback"]; ok {
		callback = callbacks[0]
	}

	if len(callback) > 0 {
		js := fmt.Sprintf("%s(%s);", callback, json);
		JsWrite(w, r, js)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", JsTtl(r)))
		fmt.Fprintf(w, json)
	}
}

func TargetJson(r *http.Request) string {
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

func InitFileServer() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
}