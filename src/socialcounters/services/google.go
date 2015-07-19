package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

func Google(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Google"

	urlJson, err := json.Marshal(url)
	if (err != nil) {
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
	resp, err := client.Do(req)
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
	json := string(respBody)

	// use regex to avoid parsing the big json string (which is quite slow with the built-in json)
	r, err := regexp.Compile(`"count": ([\d\.]+)`)
	if err != nil {
		result.Error = err
		return result
	}

	matches := r.FindStringSubmatch(json)
	if matches == nil {
		return result
	}
	
	count, err := strconv.ParseFloat(matches[1], 64)
	if (err == nil) {
		result.Error = err
		result.Count = count
	}

	return result
}