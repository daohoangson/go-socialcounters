package services

import (
	"github.com/daohoangson/go-socialcounters/utils"
)

func Batch(u utils.Utils, requests []ServiceRequest) []ServiceResult {
	results := []ServiceResult{}
	if len(requests) < 1 {
		return results
	}

	ch := make(chan ServiceResult, len(requests))
	facebookUrls := []string{}

	for _, request := range requests {
		switch request.Service {
		case "Facebook":
			facebookUrls = append(facebookUrls, request.Url)
		case "Twitter":
			go func(f ServiceFunc, url string) {
				ch <- f(u, url)
			}(Twitter, request.Url)
		case "Google":
			go func(f ServiceFunc, url string) {
				ch <- f(u, url)
			}(Google, request.Url)
		default:
			ch <- buildDummyServiceResuilt(request.Service, request.Url)
		}
	}

	if len(facebookUrls) > 0 {
		go func(f ServiceMultiFunc, urls []string) {
			results := f(u, urls)
			for _, url := range urls {
				if r, ok := results.Results[url]; ok {
					ch <- r
				} else {
					ch <- buildDummyServiceResuilt("Facebook", url)
				}
			}
		}(FacebookMulti, facebookUrls)
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

func buildDummyServiceResuilt(service string, url string) ServiceResult {
	var r ServiceResult
	r.Service = service
	r.Url = url

	return r
}

type ServiceRequest struct {
	Service string
	Url     string
}

type ServiceResult struct {
	Service  string
	Url      string
	Count    int64
	Error    error
	Response []byte
}

type ServiceResults struct {
	Results  map[string]ServiceResult
	Error    error
	Response []byte
}

type ServiceFunc func(utils.Utils, string) ServiceResult
type ServiceMultiFunc func(utils.Utils, []string) ServiceResults
