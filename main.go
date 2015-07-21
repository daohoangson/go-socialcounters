package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/web"
)

var serviceFuncs = []services.ServiceFunc{
	services.Facebook2,
	services.Twitter,
	services.Google,
}

func getCountsJson(r *http.Request) string {
	countsJson, err := web.CountsJson(r, &http.Client{}, serviceFuncs)
	if err != nil {
		log.Printf("web.CountsJson error %v", err)
	}

	return countsJson
}

func allJs(w http.ResponseWriter, r *http.Request) {
	countsJson := getCountsJson(r)
	js, err := web.AllJs(r, countsJson)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not get web.AllJs %v", err)
		return
	}

	web.JsWrite(w, r, js)
}

func dataJson(w http.ResponseWriter, r *http.Request) {
	countsJson := getCountsJson(r)

	web.JsonWrite(w, r, countsJson)
}

func main() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", allJs)
	http.HandleFunc("/js/data.json", dataJson)
	http.HandleFunc("/js/jquery.plugin.js", web.JQueryPluginJs)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	fmt.Printf("Listening on %s...\n", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}