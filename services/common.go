package services

import (
	"net/http"
)

func Batch(client *http.Client, serviceFuncs []ServiceFunc, url string) []ServiceResult {
	ch := make(chan ServiceResult, 3)
	results := []ServiceResult{}

	for _, serviceFunc := range serviceFuncs {
		go func(serviceFunc ServiceFunc, url string) {
			ch <- serviceFunc(client, url)
		}(serviceFunc, url)
	}

	for {
		select {
			case r := <-ch:
				results = append(results, r)
				if len(results) == len(serviceFuncs) {
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

type ServiceFunc func(*http.Client, string) ServiceResult