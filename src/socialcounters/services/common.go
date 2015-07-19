package services

import (
	"net/http"
)

func All(client *http.Client, url string) []ServiceResult {
	ch := make(chan ServiceResult, 3)
	results := []ServiceResult{}
	services := []func(*http.Client, string) ServiceResult{
		Facebook,
		Twitter,
		Google,
	}

	for _, service := range services {
		go func(service func(*http.Client, string) ServiceResult, url string) {
			ch <- service(client, url)
		}(service, url)
	}

	for {
		select {
			case r := <-ch:
				results = append(results, r)
				if len(results) == len(services) {
					return results
				}
		}
	}

	return results
}

type ServiceResult struct {
	Service string
	Count float64
	Error error
	Response string
}