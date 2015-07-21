package services

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

import neturl "net/url"

func Facebook1(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Facebook"

	query := `SELECT total_count FROM link_stat WHERE url="` + url + `"`

	// we have to go through crossorigin.me because for some reason Facebook returns bogus data
	// especially when request are made within GAE. I have tested with user agent and some other
	// GAE special request headers but haven't found the real culprit, yet...
	resp, err := client.Get("http://crossorigin.me/http://graph.facebook.com/fql?q=" + neturl.QueryEscape(query))
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
	r, err := regexp.Compile(`"total_count":([\d\.]+)`)
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

func Facebook2(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Facebook"

	query := `SELECT total_count FROM link_stat WHERE url="` + url + `"`
	resp, err := client.Get("https://graph.facebook.com/fql?q=" + neturl.QueryEscape(query))
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
	r, err := regexp.Compile(`"total_count":([\d\.]+)`)
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