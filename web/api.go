package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func HistoryJson(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	requestedUrls := parseUrls(r)
	url := requestedUrls[0]

	records, err := u.HistoryLoad(url)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.HistoryJson: HistoryLoad error %v", err)
		return
	}

	data := make(map[string]HistorySlot)
	slotSize := int64(300) // each slot lasts 5 minutes
	for _, record := range records {
		slotInt64 := record.Time.Unix() / slotSize * slotSize
		slotString := fmt.Sprintf("%d", slotInt64)

		if slot, ok := data[slotString]; ok {
			if _, ok := slot.Counts[record.Service]; !ok {
				slot.Counts[record.Service] = record.Count
				slot.Total += record.Count
				data[slotString] = slot
			}
		} else {
			data[slotString] = HistorySlot{
				Time:   time.Unix(slotInt64, 0).Format(time.RFC1123),
				Counts: map[string]int64{record.Service: record.Count},
				Total:  record.Count,
			}
		}
	}

	historyJson, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		u.Errorf("web.HistoryJson: json.Marshal error %v", err)
		return
	}

	writeJson(u, w, r, string(historyJson))
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
	requestedServices := parseServices(r)
	data := services.DataSetup()

	if len(requestedUrls) < 1 {
		return "", "{}", nil
	}

	url := ""
	if oneUrl {
		url = requestedUrls[0]
		requestedUrls = []string{url}
	}

	for _, requestedUrl := range requestedUrls {
		if !RulesAllowUrl(u, requestedUrl) {
			u.Errorf("Url not allowed %s", requestedUrl)
			continue
		}

		for _, requestedService := range requestedServices {
			services.DataAdd(&data, requestedService, requestedUrl)
		}
	}

	services.FillData(u, &data)

	var dataByte []byte
	var dataErr error
	if oneUrl {
		dataByte, dataErr = json.Marshal(data[url])
	} else {
		dataByte, dataErr = json.Marshal(data)
	}
	if dataErr != nil {
		return url, "{}", dataErr
	}

	return url, string(dataByte), nil
}
