package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

func googleWorker(u utils.Utils, r *request) {
	legacyBatch(u, r, googleLegacy)
}

func googleLegacy(u utils.Utils, url string) result {
	start := time.Now()
	var res result

	urlJson, err := json.Marshal(url)
	if err != nil {
		res.Error = err
		return res
	}

	// http://bradsknutson.com/blog/get-google-share-count-url/
	body := `[{"method":"pos.plusones.get","id":"p","params":{"nolog":true,"id":` + string(urlJson) + `,"source":"widget","userId":"@viewer","groupId":"@self"},"jsonrpc":"2.0","key":"p","apiVersion":"v1"}]`
	req, err := http.NewRequest("POST", "https://clients6.google.com/rpc", bytes.NewBufferString(body))
	if err != nil {
		res.Error = err
		return res
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := u.HttpClient().Do(req)
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
	u.Debugf("googleLegacy(url=%s) took %s: %s", url, time.Since(start), respBody)

	jsonparser.ArrayEach(respBody, func(element []byte, _ jsonparser.ValueType, _ int, err error) {
		if err != nil {
			res.Error = err
			return
		}

		if count, err := jsonparser.GetFloat(element, "result", "metadata", "globalCounts", "count"); err != nil {
			res.Error = err
			return
		} else {
			res.Count = int64(count)
		}
	})

	return res
}
