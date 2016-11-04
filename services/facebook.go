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
	url := prepareFbGraphUrl(u, strings.Join(req.Urls, ","))
	utils.Verbosef(u, "Calling http.Client.Get(%s)", url)

	resp, err := u.HttpClient().Get(url)
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
	u.Debugf("facebookWorker(urls=%d) took %s", len(req.Urls), time.Since(start))

	for _, url := range req.Urls {
		var res result

		if respUrl, _, _, err := jsonparser.Get(respBody, url); err != nil {
			res.Error = err
			res.Response = respBody
		} else {
			res.Response = respUrl
			if shareCount, err := jsonparser.GetInt(respUrl, "share", "share_count"); err != nil {
				res.Error = err
			} else {
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

	return fmt.Sprintf("http://graph.facebook.com/?ids=%s&fields=share%s", neturl.QueryEscape(ids), accessToken)
}
