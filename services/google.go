package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/daohoangson/go-socialcounters/utils"
)

func Google(u utils.Utils, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Google"
	result.Url = url

	urlJson, err := json.Marshal(url)
	if err != nil {
		result.Error = err
		return result
	}

	// http://bradsknutson.com/blog/get-google-share-count-url/
	body := `[{"method":"pos.plusones.get","id":"p","params":{"nolog":true,"id":` + string(urlJson) + `,"source":"widget","userId":"@viewer","groupId":"@self"},"jsonrpc":"2.0","key":"p","apiVersion":"v1"}]`
	req, err := http.NewRequest("POST", "https://clients6.google.com/rpc", bytes.NewBufferString(body))
	if err != nil {
		result.Error = err
		return result
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := u.HttpClient().Do(req)
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}
	result.Response = respBody

	jsonparser.ArrayEach(respBody, func(element []byte, _ jsonparser.ValueType, _ int, err error) {
		if err != nil {
			result.Error = err
			return
		}

		if count, err := jsonparser.GetFloat(element, "result", "metadata", "globalCounts", "count"); err != nil {
			result.Error = err
			return
		} else {
			result.Count = int64(count)
		}
	})

	return result
}
