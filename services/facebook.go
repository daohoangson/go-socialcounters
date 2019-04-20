package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"

	neturl "net/url"
)

const fbGraphURLTemplate = "https://graph.facebook.com/v3.2/?ids=%s&fields=engagement&access_token=%s|%s"

func facebookWorker(u utils.Utils, req *request) {
	start := time.Now()
	urls := strings.Join(req.Urls, ",")
	url, err := prepareFbGraphURL(u, urls)
	if err != nil {
		req.Error = err
		return
	}
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
			res.Count = countInitValue
			if engagement, _, _, err := jsonparser.Get(urlResp, "engagement"); err == nil {
				if reactionCount, err := jsonparser.GetInt(engagement, "reaction_count"); err == nil {
					res.Count += reactionCount
				}
				if commentCount, err := jsonparser.GetInt(engagement, "comment_count"); err == nil {
					res.Count += commentCount
				}
				if shareCount, err := jsonparser.GetInt(engagement, "share_count"); err == nil {
					res.Count += shareCount
				}
			}
		}

		req.Results[url] = res
	}

	return
}

func prepareFbGraphURL(u utils.Utils, ids string) (string, error) {
	appID := u.ConfigGet("FACEBOOK_APP_ID")
	if appID == "" {
		return "", errors.New("Env var FACEBOOK_APP_ID must be configured")
	}

	appSecret := u.ConfigGet("FACEBOOK_APP_SECRET")
	if appSecret == "" {
		return "", errors.New("Env var FACEBOOK_APP_SECRET must be configured")
	}

	return fmt.Sprintf(fbGraphURLTemplate, neturl.QueryEscape(ids), appID, appSecret), nil
}
