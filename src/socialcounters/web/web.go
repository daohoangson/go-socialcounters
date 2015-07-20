package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/daohoangson/minify/css"
	"socialcounters/services"
)

func AllJs(client *http.Client, url string) (string, error) {
	jsData, err := ioutil.ReadFile("private/js/all.js")
	if (err != nil) {
		return "", err
	}
	js := strings.Replace(string(jsData), "{url}", url, 1)

	css := css.MinifyFromFile("private/css/main.css")
	js = strings.Replace(js, "{css}", css, 1)

	serviceResults := services.All(client, url)
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

func InitFileServer() {
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
}