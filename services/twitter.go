package services

import (
	"errors"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

import neturl "net/url"

func twitterWorker(u utils.Utils, r *request) {
	legacyBatch(u, r, twitterLegacy)
}

func twitterLegacy(u utils.Utils, url string) result {
	var res result
	start := time.Now()
	oscURL := "http://opensharecount.com/count.json?url=" + neturl.QueryEscape(url)
	utils.Verbosef(u, "Doing GET %s...", oscURL)

	respBody, err := u.HttpGet(oscURL)
	if err != nil {
		res.Error = err
		return res
	}
	res.Response = respBody
	u.Debugf("twitterLegacy(url=%s) took %s", url, time.Since(start))

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
