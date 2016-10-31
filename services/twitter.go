package services

import (
	"errors"
	"io/ioutil"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

import neturl "net/url"

func twitterWorker(u utils.Utils, r *request) {
	legacyBatch(u, r, twitterLegacy)
}

func twitterLegacy(u utils.Utils, url string) result {
	start := time.Now()
	var res result

	resp, err := u.HttpClient().Get("https://opensharecount.com/count.json?url=" + neturl.QueryEscape(url))
	if err != nil {
		res.Error = err
		return res
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res.Error = err
		return res
	}
	res.Response = respBody
	u.Debugf("twitterLegacy(url=%s) took %s: %s", url, time.Since(start), respBody)

	if count, err := jsonparser.GetInt(respBody, "count"); err != nil {
		res.Error = err
		return res
	} else {
		if count == 0 {
			if oscError, err := jsonparser.GetString(respBody, "error"); err != nil {
				res.Error = errors.New(oscError)
				return res
			}
		}

		res.Count = count
	}

	return res
}
