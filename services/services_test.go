package services

import (
	"net/http"
	"testing"
)

func testOk(t *testing.T, f ServiceFunc, url string, expectedService string, expectedCount int64) {
	client := new(http.Client)
	serviceResult := f(client, url)

	if serviceResult.Error != nil {
		t.Fatalf("Unexpected `Error` (%s), Response=%s", serviceResult.Error, serviceResult.Response)
	}

	if serviceResult.Service != expectedService {
		t.Fatalf("Incorrect `Service` (%s)", serviceResult.Service)
	}

	if serviceResult.Count < expectedCount {
		t.Fatalf("Count is too small (%d, should be > %d)", serviceResult.Count, expectedCount)
	}

	t.Logf("%s(%s): Count=%d, Response=%s", expectedService,
		url, serviceResult.Count, serviceResult.Response)
}

func TestFacebook(t *testing.T) {
	testOk(t, Facebook, "https://facebook.com", "Facebook", int64(100000000))
}

func TestGoogle(t *testing.T) {
	testOk(t, Google, "https://google.com", "Google", int64(10000000))
}

func TestTwitter(t *testing.T) {
	testOk(t, Twitter, "https://twitter.com", "Twitter", int64(97000))
}