package services

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

import neturl "net/url"

func facebook(fbGraphRoot string, client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Facebook"
	
	fbGraphUrl := fmt.Sprintf("%s?ids=%s", fbGraphRoot, neturl.QueryEscape(url))
	resp, err := client.Get(fbGraphUrl)
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
	result.Response = json

	// use regex to avoid parsing the big json string (which is quite slow with the built-in json)
	r, err := regexp.Compile(`"share_count":([\d\.]+)`)
	if err != nil {
		result.Error = err
		return result
	}

	matches := r.FindStringSubmatch(json)
	if matches == nil {
		return result
	}

	count, err := strconv.ParseInt(matches[1], 10, 64)
	if err == nil {
		result.Count = count
	}

	return result
}

func FacebookCrossOrigin(client *http.Client, url string) ServiceResult {
	return facebook("http://crossorigin.me/http://graph.facebook.com/", client, url)
}

func FacebookDirect(client *http.Client, url string) ServiceResult {
	return facebook("https://graph.facebook.com/", client, url)
}
