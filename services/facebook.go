package services

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/buger/jsonparser"
)

import neturl "net/url"

func Facebook(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Facebook"
	result.Url = url
	result.Error = errors.New("Not implemented")

	return result
}

func FacebookMulti(client *http.Client, urls []string) ServiceResults {
	var results ServiceResults
	results.Results = make(map[string]ServiceResult)

	resp, err := client.Get(prepareFbGraphUrl(strings.Join(urls, ",")))
	if err != nil {
		results.Error = err
		return results
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		results.Error = err
		return results
	}
	results.Response = respBody

	for _, url := range urls {
		var result ServiceResult
		result.Service = "Facebook"
		result.Url = url
		if shareCount, err := jsonparser.GetInt(respBody, url, "share", "share_count"); err == nil {
			result.Count = shareCount
		} else {
			result.Error = err
		}

		results.Results[result.Url] = result
	}

	return results
}

func prepareFbGraphUrl(ids string) string {
	accessToken := ""
	if appId := os.Getenv("FACEBOOK_APP_ID"); appId != "" {
		if appSecret := os.Getenv("FACEBOOK_APP_SECRET"); appSecret != "" {
			accessToken = fmt.Sprintf("&access_token=%s|%s", appId, appSecret)
		}
	}

	return fmt.Sprintf("https://graph.facebook.com/?ids=%s&fields=share%s", neturl.QueryEscape(ids), accessToken)
}
