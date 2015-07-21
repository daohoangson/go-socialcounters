package services

import (
	"encoding/json"
	"io"
	"net/http"
)

import neturl "net/url"

type twitterResponse struct {
	Count float64
	Url string
}

func Twitter(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Twitter"

	resp, err := client.Get("https://cdn.api.twitter.com/1/urls/count.json?url=" + neturl.QueryEscape(url))
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		var tr twitterResponse
		if err := dec.Decode(&tr); err != nil {
			if err != io.EOF {
				result.Error = err
			}
			break
		}

		result.Count = int64(tr.Count)
	}

	return result
}