package main

import (
	"fmt"
	"log"
	"net/http"

	"socialcounters/web"
)

func mainAllJs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var url string
	if urls, ok := q["url"]; ok {
		url = urls[0]
	}
	if len(url) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("No `url` specified for all.js")
		return
	}

	ttl := 300	
	client := &http.Client{}
	js, err := web.AllJs(client, url)
	if (err != nil) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not prepare all.js %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", fmt.Sprintf("public; max-age=%d", ttl))
	fmt.Fprintf(w, js)
}

func main() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", mainAllJs)

	log.Fatal(http.ListenAndServe(":8080", nil))
}