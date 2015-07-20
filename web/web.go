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

func AllJs(r *http.Request, client *http.Client, serviceFuncs []services.ServiceFunc) (string, error) {
	q := r.URL.Query()
	var url string
	if urls, ok := q["url"]; ok {
		url = urls[0]
	}
	if len(url) == 0 {
		return "", errors.New("No `url` specified for all.js")
	}

	jsData, err := ioutil.ReadFile("private/js/all.js")
	if (err != nil) {
		return "", err
	}
	js := strings.Replace(string(jsData), "{url}", url, 1)
	js = strings.Replace(js, "{now}", fmt.Sprintf("%v", time.Now()), 1)

	css := css.MinifyFromFile("private/css/main.css")
	js = strings.Replace(js, "{css}", css, 1)

	serviceResults := services.Batch(client, serviceFuncs, url)
	counts := make(map[string]float64)
	var lastErr error
	for _, serviceResult := range serviceResults {
		counts[serviceResult.Service] = serviceResult.Count

		if serviceResult.Error != nil {
			lastErr = serviceResult.Error
		}
	}
	if len(counts) == 0 && lastErr != nil {
		return "", nil
	}

	countsJson, err := json.Marshal(counts)
	if err != nil {
		return "", err
	}
	js = strings.Replace(js, "{counts}", string(countsJson), 1)

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

func InitFileServer() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
}