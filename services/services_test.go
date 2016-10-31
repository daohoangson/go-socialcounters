package services

import (
	"strings"
	"testing"

	"github.com/daohoangson/go-socialcounters/utils"
)

var u = utils.OtherNew(nil)

func assert(t *testing.T, serviceResult ServiceResult, url string, expectedService string, expectedCount int64) {
	if serviceResult.Error != nil {
		t.Fatalf("Unexpected `Error` (%s), Response=%s", serviceResult.Error, serviceResult.Response)
	}

	if serviceResult.Url != url {
		t.Fatalf("Incorrect `Url` %s, requested %s", serviceResult.Url, url)
	}

	if serviceResult.Service != expectedService {
		t.Fatalf("Incorrect `Service` (%s)", serviceResult.Service)
	}

	if serviceResult.Count < expectedCount {
		t.Fatalf("Count is too small (%d, should be > %d)", serviceResult.Count, expectedCount)
	}
}

func testOne(t *testing.T, f ServiceFunc, url string, expectedService string, expectedCount int64) {
	serviceResult := f(u, url)
	assert(t, serviceResult, url, expectedService, expectedCount)

	t.Logf("%s(%s): Count=%d, Response=%s", serviceResult.Service,
		serviceResult.Url, serviceResult.Count, serviceResult.Response)
}

func testMulti(t *testing.T, f ServiceMultiFunc, urls []string, expectedService string, expectedCounts []int64) {
	if len(expectedCounts) != len(urls) {
		t.Fatalf("Not enough expected counts (%d), there are %d urls", len(expectedCounts), len(urls))
	}

	serviceResults := f(u, urls)

	if len(serviceResults.Results) != len(urls) {
		t.Fatalf("Not enough results (%d), requested for %d urls", len(serviceResults.Results), len(urls))
	}

	for index, url := range urls {
		serviceResult := serviceResults.Results[url]
		expectedCount := expectedCounts[index]

		assert(t, serviceResult, url, expectedService, expectedCount)

		t.Logf("%s(%s): Count=%d", serviceResult.Service, serviceResult.Url, serviceResult.Count)
	}

	t.Logf("%s(%s): Response=%s", expectedService,
		strings.Join(urls, ", "), serviceResults.Response)
}

func TestFacebook(t *testing.T) {
	urls := []string{"https://facebook.com", "https://developers.facebook.com"}
	expectedCounts := []int64{int64(100000000), int64(200000)}
	testMulti(t, FacebookMulti, urls, "Facebook", expectedCounts)
}

func TestGoogle(t *testing.T) {
	testOne(t, Google, "https://google.com", "Google", int64(10000000))
}

func TestTwitter(t *testing.T) {
	testOne(t, Twitter, "https://twitter.com", "Twitter", int64(97000))
}

func TestBatch(t *testing.T) {
	requests := []ServiceRequest{ServiceRequest{Service: "Facebook", Url: "https://facebook.com"},
		ServiceRequest{Service: "Google", Url: "https://google.com"},
		ServiceRequest{Service: "Twitter", Url: "https://twitter.com"},
		ServiceRequest{Service: "Facebook", Url: "https://developers.facebook.com"}}

	serviceResults := Batch(u, requests)

	for _, serviceResult := range serviceResults {
		if serviceResult.Count < 1 {
			t.Fatalf("Count(%s, %s) == %d", serviceResult.Service, serviceResult.Url, serviceResult.Count)
		}

		t.Logf("%s(%s): Count=%d, Response=%s", serviceResult.Service,
			serviceResult.Url, serviceResult.Count, serviceResult.Response)
	}
}