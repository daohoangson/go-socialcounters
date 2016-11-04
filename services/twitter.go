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
	var res result
	start := time.Now()
	oscUrl := "https://opensharecount.com/count.json?url=" + neturl.QueryEscape(url)
	utils.Verbosef(u, "Calling http.Client.Get(%s)", oscUrl)

	resp, err := u.HttpClient().Get(oscUrl)
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
