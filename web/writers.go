package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/daohoangson/go-socialcounters/utils"
)

func writeCacheControlHeaders(u utils.Utils, w http.ResponseWriter, r *http.Request) {
	ttl := parseTtl(u, r)
	w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", ttl))

	expires := time.Now().Add(time.Duration(ttl) * time.Second)
	w.Header().Set("Expires", expires.Format(time.RFC1123))
}

func writeJs(u utils.Utils, w http.ResponseWriter, r *http.Request, js string) {
	writeCacheControlHeaders(u, w, r)
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, MinifyJs(js))
}

func writeJson(u utils.Utils, w http.ResponseWriter, r *http.Request, json string) {
	q := r.URL.Query()
	var callback string
	if callbacks, ok := q["callback"]; ok {
		callback = callbacks[0]
	}

	if len(callback) > 0 {
		js := fmt.Sprintf("%s(%s);", callback, json)
		writeJs(u, w, r, js)
	} else {
		writeCacheControlHeaders(u, w, r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, MinifyJson(json))
	}
}