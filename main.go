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

func allJs(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	js, err := web.AllJs(r, client, serviceFuncs)
	if (err != nil) {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not prepare all.js %v", err)
		return
	}

	web.JsWrite(w, js)
}

func main() {
	web.InitFileServer()
	http.HandleFunc("/js/all.js", allJs)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	fmt.Printf("Listening on %s...\n", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}