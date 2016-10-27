package services

import (
	"net/http"
)

func Batch(client *http.Client, requests []ServiceRequest) []ServiceResult {
	results := []ServiceResult{}
	if len(requests) < 1 {
		return results
	}

	ch := make(chan ServiceResult, len(requests))

	for _, request := range requests {
		go func(serviceFunc ServiceFunc, url string) {
			ch <- serviceFunc(client, url)
		}(request.Func, request.Url)
	}

	for {
		select {
		case r := <-ch:
			results = append(results, r)
			if len(results) == len(requests) {
				return results
			}
		}
	}

	return results
}

type ServiceRequest struct {
	Func ServiceFunc
	Url  string
}

type ServiceResult struct {
	Service  string
	Url      string
	Count    int64
	Error    error
	Response string
}

type ServiceFunc func(*http.Client, string) ServiceResult
