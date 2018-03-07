package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

import neturl "net/url"

func facebookWorker(u utils.Utils, req *request) {
	start := time.Now()
	urls := strings.Join(req.Urls, ",")
	url := prepareFbGraphURL(u, urls)
	utils.Verbosef(u, "Doing GET %s...", url)

	respBody, err := u.HttpGet(url)
	if err != nil {
		req.Error = err
		return
	}
	req.Response = respBody
	u.Debugf("facebookWorker(urls=%s) took %s", urls, time.Since(start))

	for _, url := range req.Urls {
		var res result

		if urlResp, _, _, err := jsonparser.Get(respBody, url); err != nil {
			res.Error = err
			res.Response = respBody
		} else {
			res.Response = urlResp
			if shareCount, err := jsonparser.GetInt(urlResp, "share", "share_count"); err != nil {
				// it's alright, for new urls Facebook does not return share.share_count at all
				res.Count = countInitValue
			} else {
				res.Count = shareCount
			}
		}

		req.Results[url] = res
	}

	return
}

func prepareFbGraphURL(u utils.Utils, ids string) string {
	scheme := "http"
	accessToken := ""
	if appID := u.ConfigGet("FACEBOOK_APP_ID"); appID != "" {
		if appSecret := u.ConfigGet("FACEBOOK_APP_SECRET"); appSecret != "" {
			scheme = "https"
			accessToken = fmt.Sprintf("&access_token=%s|%s", appID, appSecret)
		}
	}

	return fmt.Sprintf("%s://graph.facebook.com/?ids=%s&fields=share%s", scheme, neturl.QueryEscape(ids), accessToken)
}
