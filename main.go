package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/utils"
	"github.com/daohoangson/go-socialcounters/web"
	"google.golang.org/appengine"
)

func utilsFuncGae(_ http.ResponseWriter, r *http.Request) utils.Utils {
	return utils.GaeNew(r)
}

func utilsFuncOther(_ http.ResponseWriter, r *http.Request) utils.Utils {
	return utils.OtherNew(r)
}

func main() {
	services.Init()

	if len(os.Getenv("GAE_SERVICE")) > 0 {
		http.Handle("/", web.BuildHandler(utilsFuncGae, true))
		appengine.Main()
		return
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	fmt.Printf("Listening on %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, web.BuildHandler(utilsFuncOther, true)))
}
