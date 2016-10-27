package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/daohoangson/go-minify/css"
	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/utils"
)

func AllJs(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	url, countsJson, err := getCountsJson(u, r, true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.AllJs: getCountsJson error %v", err)
		return
	}

	urlByte, err := json.Marshal(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.AllJs: json.Marshal(url) error %v", err)
		return
	}

	jsData, err := ioutil.ReadFile("private/js/all.js")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.AllJs: ReadFile error %v", err)
		return
	}
	js := MinifyJs(string(jsData))
	js = strings.Replace(js, "{url}", string(urlByte), 1)
	js = strings.Replace(js, "{now}", fmt.Sprintf("%v", time.Now()), 1)

	// keep using css.MinifyFromFile because it does the data uri inline for us
	// TODO: drop this github.com/daohoangson/go-minify/css dependency
	css := css.MinifyFromFile("public/css/main.css")
	js = strings.Replace(js, "{css}", css, 1)

	js = strings.Replace(js, "{facebooksvg}", readFileAsJson("private/img/facebook.svg"), 1)
	js = strings.Replace(js, "{twittersvg}", readFileAsJson("private/img/twitter.svg"), 1)
	js = strings.Replace(js, "{googlesvg}", readFileAsJson("private/img/google.svg"), 1)

	js = strings.Replace(js, "{ads}", getAdsAsJson(), 1)
	js = strings.Replace(js, "{counts}", string(countsJson), 1)
	js = strings.Replace(js, "{shorten}", parseShortenAsBool(r), 1)
	js = strings.Replace(js, "{target}", parseTargetAsJson(r), 1)

	writeJs(w, r, js)
}

func DataJson(u utils.Utils, w http.ResponseWriter, r *http.Request, oneUrl bool) {
	_, countsJson, err := getCountsJson(u, r, oneUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.DataJson: getCountsJson error %v", err)
		return
	}

	writeJson(w, r, countsJson)
}

func JqueryPluginJs(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	jsData, err := ioutil.ReadFile("private/js/jquery.plugin.js")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		u.Errorf("web.JqueryPluginJs: ReadFile error %v", err)
		return
	}

	jsonUrl := fmt.Sprintf("//%s/js/data.json", r.Host)
	js := strings.Replace(string(jsData), "{jsonUrl}", jsonUrl, 1)
	js = strings.Replace(js, "{ads}", getAdsAsJson(), 1)

	writeJs(w, r, js)
}

func getAdsAsJson() string {
	json, err := json.Marshal(os.Getenv("ADS"))
	if err != nil {
		return "''"
	}

	return string(json)
}

func getCountsJson(u utils.Utils, r *http.Request, oneUrl bool) (string, string, error) {
	requestedUrls := parseUrls(r)
	if len(requestedUrls) < 1 {
		return "", "{}", nil
	}

	url := ""
	if oneUrl {
		url = requestedUrls[0]
		requestedUrls = []string{url}
	}

	requestedServices := parseServices(r)
	ttl := parseTtl(r)
	dataMap := make(map[string]map[string]int64)
	requests := []services.ServiceRequest{}

	for _, requestedUrl := range requestedUrls {
		if !RulesAllowUrl(u, requestedUrl) {
			continue
		}

		dataMap[requestedUrl] = make(map[string]int64)

		for _, requestedService := range requestedServices {
			if value, err := u.MemoryGet(getCacheKey(requestedService, requestedUrl)); err == nil {
				if count, err := strconv.ParseInt(string(value), 10, 64); err == nil {
					dataMap[requestedUrl][requestedService] = count
				}
			}

			if _, ok := dataMap[requestedUrl][requestedService]; !ok {
				if serviceFunc := u.ServiceFunc(requestedService); serviceFunc != nil {
					var request services.ServiceRequest
					request.Func = serviceFunc
					request.Url = requestedUrl

					requests = append(requests, request)
				}
			}
		}
	}

	if len(requests) > 0 {
		serviceResults := services.Batch(u.HttpClient(), requests)

		for _, serviceResult := range serviceResults {
			dataMap[serviceResult.Url][serviceResult.Service] = serviceResult.Count
			u.MemorySet(getCacheKeyForResult(serviceResult), []byte(fmt.Sprintf("%d", serviceResult.Count)), ttl)
		}
	}

	var dataByte []byte
	var dataErr error
	if oneUrl {
		dataByte, dataErr = json.Marshal(dataMap[url])
	} else {
		dataByte, dataErr = json.Marshal(dataMap)
	}
	if dataErr != nil {
		return url, "{}", dataErr
	}

	return url, string(dataByte), nil
}

func parseShortenAsBool(r *http.Request) string {
	q := r.URL.Query()
	if _, ok := q["shorten"]; ok {
		return "true"
	}

	return "false"
}

func parseTargetAsJson(r *http.Request) string {
	target := "'.socialcounters-container'"

	q := r.URL.Query()
	if targets, ok := q["target"]; ok {
		targetByte, err := json.Marshal(targets[0])
		if err == nil {
			target = string(targetByte)
		}
	}

	return target
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

func parseUrls(r *http.Request) []string {
	urls := []string{}

	q := r.URL.Query()
	if queryUrls, ok := q["url"]; ok {
		for _, queryUrl := range queryUrls {
			if len(queryUrl) > 0 {
				urls = append(urls, queryUrl)
			}
		}
	}

	return urls
}

func parseServices(r *http.Request) []string {
	q := r.URL.Query()
	if services, ok := q["services"]; ok {
		return services
	}

	return []string{"Facebook", "Twitter", "Google"}
}

func getCacheKey(service string, url string) string {
	return fmt.Sprintf("%s/%s", service, url)
}

func getCacheKeyForResult(result services.ServiceResult) string{
	return getCacheKey(result.Service, result.Url)
}

func readFileAsJson(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "''"
	}

	json, err := json.Marshal(string(data))
	if err != nil {
		return "''"
	}

	return string(json)
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
		js := fmt.Sprintf("%s(%s);", callback, json)
		writeJs(w, r, js)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", parseTtl(r)))
		fmt.Fprintf(w, json)
	}
}
