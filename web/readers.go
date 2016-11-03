package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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
