package services

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

import neturl "net/url"

func FacebookMulti(u utils.Utils, urls []string) ServiceResults {
	start := time.Now()
	var results ServiceResults
	results.Results = make(map[string]ServiceResult)

	resp, err := u.HttpClient().Get(prepareFbGraphUrl(u, strings.Join(urls, ",")))
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
	u.Debugf("FacebookMulti(urls=%s) took %s: %s", strings.Join(urls, ", "), time.Since(start), respBody)

	for _, url := range urls {
		var result ServiceResult
		result.Service = "Facebook"
		result.Url = url

		if respUrl, _, _, err := jsonparser.Get(respBody, url); err != nil {
			result.Error = err
		} else {
			result.Response = respUrl
			if shareCount, err := jsonparser.GetInt(respUrl, "share", "share_count"); err == nil {
				result.Count = shareCount
			}
		}

		results.Results[result.Url] = result
	}

	return results
}

func prepareFbGraphUrl(u utils.Utils, ids string) string {
	accessToken := ""
	if appId := u.ConfigGet("FACEBOOK_APP_ID"); appId != "" {
		if appSecret := u.ConfigGet("FACEBOOK_APP_SECRET"); appSecret != "" {
			accessToken = fmt.Sprintf("&access_token=%s|%s", appId, appSecret)
		}
	}

	return fmt.Sprintf("https://graph.facebook.com/?ids=%s&fields=share%s", neturl.QueryEscape(ids), accessToken)
}
