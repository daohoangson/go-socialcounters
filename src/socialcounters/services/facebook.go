package services

import (
	"encoding/json"
	"io"
	"net/http"
)

import neturl "net/url"

func Facebook(client *http.Client, url string) ServiceResult {
	var result ServiceResult
	result.Service = "Facebook"

	resp, err := client.Get("https://graph.facebook.com?ids=" + neturl.QueryEscape(url))
	if err != nil {
		result.Error = err
		return result
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	for {
		var f interface{}
		if err := dec.Decode(&f); err != nil {
			if err != io.EOF {
				result.Error = err
			}
			break
		}
		
		m := f.(map[string]interface{})
		if f2, ok := m[url]; ok {
			m2 := f2.(map[string]interface{})
			for k, v := range m2 {
				if k == "shares" {
					switch vv := v.(type) {
					    case float64:
					    	result.Count = vv
				    }
				}
			}
		}
	}

	return result
}