package services

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/daohoangson/go-socialcounters/utils"
)

var u = utils.OtherNew(nil)

func assert(t *testing.T, req request, expectedUrl string, expectedCount int64) {
	expectedResultFound := false

	for url, res := range req.Results {
		if url != expectedUrl {
			continue
		}
		expectedResultFound = true

		if res.Error != nil {
			t.Fatalf("Unexpected `Error` (%s), Response=%s", res.Error, res.Response)
		}

		if res.Count < expectedCount {
			t.Fatalf("Count is too small (%d, should be > %d)", res.Count, expectedCount)
		}

		t.Logf("%s(%s): Count=%d, Response=%s", req.Service, url, res.Count, res.Response)
	}

	if !expectedResultFound {
		t.Fatalf("Expected result for url %s could not be found", expectedUrl)
	}
}

func testOne(t *testing.T, f worker, url string, expectedCount int64) {
	testMulti(t, f, []string{url}, []int64{expectedCount})
}

func testMulti(t *testing.T, f worker, urls []string, expectedCounts []int64) {
	if len(expectedCounts) != len(urls) {
		t.Fatalf("Not enough expected counts (%d), there are %d urls", len(expectedCounts), len(urls))
	}

	var req request
	req.Service = runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	req.Urls = urls
	req.Results = make(MapUrlResult)
	f(u, &req)

	if len(req.Results) != len(urls) {
		t.Fatalf("Not enough results (%d), requested for %d urls", len(req.Results), len(urls))
	}

	for index, url := range urls {
		assert(t, req, url, expectedCounts[index])
	}
}

func TestFacebook(t *testing.T) {
	urls := []string{"https://facebook.com", "https://developers.facebook.com"}
	expectedCounts := []int64{int64(100000000), int64(200000)}
	testMulti(t, facebookWorker, urls, expectedCounts)
}

func TestTwitter(t *testing.T) {
	testOne(t, twitterWorker, "https://twitter.com", int64(97000))
}

func TestFillData(t *testing.T) {
	facebookUrl1 := "https://facebook.com"
	facebookUrl2 := "https://developers.facebook.com"
	twitterUrl1 := "https://twitter.com"
	twitterUrl2 := "http://opensharecount.com"

	mapping := map[string][]string{
		SERVICE_FACEBOOK: []string{facebookUrl1, facebookUrl2},
		SERVICE_TWITTER:  []string{twitterUrl1, twitterUrl2},
	}

	data := DataSetup()
	for service, urls := range mapping {
		for _, url := range urls {
			DataAdd(&data, service, url)
		}
	}

	FillData(u, &data)

	for service, urls := range mapping {
		for _, url := range urls {
			if count, ok := data[url][service]; !ok {
				t.Fatalf("Count(%s, %s) could not be found", service, url)
			} else {
				if count < 1 {
					t.Fatalf("Count(%s, %s) == %d", service, url, count)
				}
			}
		}
	}
}
