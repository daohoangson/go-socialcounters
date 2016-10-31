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
	js := string(jsData)
	js = strings.Replace(js, "{url}", string(urlByte), 1)
	js = strings.Replace(js, "{now}", fmt.Sprintf("%v", time.Now()), 1)

	// keep using css.MinifyFromFile because it does the data uri inline for us
	// TODO: drop this github.com/daohoangson/go-minify/css dependency
	css := css.MinifyFromFile("public/css/main.css")
	css = MinifyCss(css)
	js = strings.Replace(js, "{css}", css, 1)

	js = strings.Replace(js, "{facebooksvg}", readSvgAsJson("private/img/facebook.svg"), 1)
	js = strings.Replace(js, "{twittersvg}", readSvgAsJson("private/img/twitter.svg"), 1)
	js = strings.Replace(js, "{googlesvg}", readSvgAsJson("private/img/google.svg"), 1)

	js = strings.Replace(js, "{ads}", getAdsAsJson(u), 1)
	js = strings.Replace(js, "{counts}", string(countsJson), 1)
	js = strings.Replace(js, "{shorten}", parseShortenAsBool(r), 1)
	js = strings.Replace(js, "{target}", parseTargetAsJson(r), 1)

	writeJs(u, w, r, js)
}

func DataJson(u utils.Utils, w http.ResponseWriter, r *http.Request, oneUrl bool) {
	_, countsJson, err := getCountsJson(u, r, oneUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.DataJson: getCountsJson error %v", err)
		return
	}

	writeJson(u, w, r, countsJson)
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
	js = strings.Replace(js, "{ads}", getAdsAsJson(u), 1)

	writeJs(u, w, r, js)
}

func getAdsAsJson(u utils.Utils) string {
	json, err := json.Marshal(u.ConfigGet("ADS"))
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
	ttl := parseTtl(u, r)
	dataMap := make(map[string]map[string]int64)
	requests := []services.ServiceRequest{}

	for _, requestedUrl := range requestedUrls {
		if !RulesAllowUrl(u, requestedUrl) {
			u.Errorf("Url not allowed %s", requestedUrl)
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
				var request services.ServiceRequest
				request.Service = requestedService
				request.Url = requestedUrl

				requests = append(requests, request)
			}
		}
	}

	if len(requests) > 0 {
		serviceResults := services.Batch(u, requests)

		for _, serviceResult := range serviceResults {
			serviceResultTtl := ttl
			dataMap[serviceResult.Url][serviceResult.Service] = serviceResult.Count

			if serviceResult.Error != nil {
				u.Errorf("Error for %s on %s: %s", serviceResult.Url, serviceResult.Service, serviceResult.Error)
			} else {
				if serviceResult.Count == 0 {
					serviceResultTtlRestricted := false

					if ttlRestrictedEnv := u.ConfigGet("TTL_COUNT_EQUALS_ZERO"); ttlRestrictedEnv != "" {
						if ttlRestricted, err := strconv.ParseInt(ttlRestrictedEnv, 10, 64); err == nil {
							serviceResultTtl = ttlRestricted
							serviceResultTtlRestricted = true
							u.Infof("Restricted TTL for %s on %s: %d", serviceResult.Url, serviceResult.Service, serviceResultTtl)
						}
					}

					if !serviceResultTtlRestricted {
						u.Debugf("%s(%s).Count == 0 without TTL restriction", serviceResult.Service, serviceResult.Url)
					}
				}
			}

			u.MemorySet(getCacheKeyForResult(serviceResult), []byte(fmt.Sprintf("%d", serviceResult.Count)), serviceResultTtl)
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

func parseTtl(u utils.Utils, r *http.Request) int64 {
	q := r.URL.Query()
	if ttls, ok := q["ttl"]; ok {
		if ttl, err := strconv.ParseInt(ttls[0], 10, 64); err == nil {
			return ttl
		}
	}

	if env := u.ConfigGet("TTL_DEFAULT"); env != "" {
		if ttl, err := strconv.ParseInt(env, 10, 64); err == nil {
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

	if r.Method == "POST" {
		r.ParseForm()
		if formUrls, ok := r.PostForm["url"]; ok {
			for _, formUrl := range formUrls {
				if len(formUrl) > 0 {
					urls = append(urls, formUrl)
				}
			}
		}
	}

	return urls
}

func parseServices(r *http.Request) []string {
	q := r.URL.Query()
	if services, ok := q["service"]; ok {
		return services
	}

	return []string{"Facebook", "Twitter", "Google"}
}

func getCacheKey(service string, url string) string {
	return fmt.Sprintf("%s/%s", service, url)
}

func getCacheKeyForResult(result services.ServiceResult) string {
	return getCacheKey(result.Service, result.Url)
}

func readSvgAsJson(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "''"
	}

	json, err := json.Marshal(MinifySvg(string(data)))
	if err != nil {
		return "''"
	}

	return string(json)
}
