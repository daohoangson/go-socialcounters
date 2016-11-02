package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/daohoangson/go-socialcounters/utils"
)

func parseShortenAsBool(r *http.Request) string {
	q := r.URL.Query()
	if _, ok := q["shorten"]; ok {
		return "true"
	}

	return "false"
}

func parseTargetAsJson(r *http.Request) string {
	target := "'.socialcounters-container'"

	q := r.URL.Query()
	if targets, ok := q["target"]; ok {
		targetByte, err := json.Marshal(targets[0])
		if err == nil {
			target = string(targetByte)
		}
	}

	return target
}

func parseTtl(u utils.Utils, r *http.Request) int64 {
	q := r.URL.Query()
	if ttls, ok := q["ttl"]; ok {
		if ttl, err := strconv.ParseInt(ttls[0], 10, 64); err == nil {
			return ttl
		}
	}

	if ttl, err := utils.ConfigGetInt(u, "TTL_DEFAULT"); err == nil {
		return ttl
	}

	return 300
}

func parseUrls(r *http.Request) []string {
	urls := []string{}

	q := r.URL.Query()
	if queryUrls, ok := q["url"]; ok {
		for _, queryUrl := range queryUrls {
			if len(queryUrl) > 0 {
				urls = append(urls, queryUrl)
			}
		}
	}

	if r.Method == "POST" {
		r.ParseForm()
		if formUrls, ok := r.PostForm["url"]; ok {
			for _, formUrl := range formUrls {
				if len(formUrl) > 0 {
					urls = append(urls, formUrl)
				}
			}
		}
	}

	return urls
}

func parseServices(r *http.Request) []string {
	q := r.URL.Query()
	if services, ok := q["service"]; ok {
		return services
	}

	return []string{"Facebook", "Twitter", "Google"}
}

func readSvgAsJson(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "''"
	}

	json, err := json.Marshal(MinifySvg(string(data)))
	if err != nil {
		return "''"
	}

	return string(json)
}
