package services

import (
	"io/ioutil"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

import neturl "net/url"

type twitterResponse struct {
	Count float64
	Url   string
}

func Twitter(u utils.Utils, url string) ServiceResult {
	start := time.Now()
	var result ServiceResult
	result.Service = "Twitter"
	result.Url = url

	resp, err := u.HttpClient().Get("https://opensharecount.com/count.json?url=" + neturl.QueryEscape(url))
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
	u.Debugf("Twitter(url=%s) took %s: %s", url, time.Since(start), respBody)

	if count, err := jsonparser.GetInt(respBody, "count"); err != nil {
		result.Error = err
		return result
	} else {
		result.Count = count
	}

	return result
}
