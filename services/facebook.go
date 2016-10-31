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

func facebookWorker(u utils.Utils, req *request) {
	start := time.Now()
	resp, err := u.HttpClient().Get(prepareFbGraphUrl(u, strings.Join(req.Urls, ",")))
	if err != nil {
		req.Error = err
		return
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		req.Error = err
		return
	}
	req.Response = respBody
	u.Debugf("facebookWorker(urls=%s) took %s: %s", strings.Join(req.Urls, ", "), time.Since(start), respBody)

	for _, url := range req.Urls {
		var res result

		if respUrl, _, _, err := jsonparser.Get(respBody, url); err != nil {
			res.Error = err
		} else {
			res.Response = respUrl
			if shareCount, err := jsonparser.GetInt(respUrl, "share", "share_count"); err == nil {
				res.Count = shareCount
			}
		}

		req.Results[url] = res
	}

	return
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
