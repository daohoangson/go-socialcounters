package services

import (
	"io/ioutil"
	"net/http"

	"github.com/buger/jsonparser"
)

import neturl "net/url"

type twitterResponse struct {
	Count float64
	Url   string
}

func Twitter(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Twitter"
	result.Url = url

	resp, err := client.Get("https://opensharecount.com/count.json?url=" + neturl.QueryEscape(url))
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}
	result.Response = respBody

	if count, err := jsonparser.GetInt(respBody, "count"); err != nil {
		result.Error = err
		return result
	} else {
		result.Count = count
	}

	return result
}
