// +build !appengine

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/daohoangson/go-socialcounters/utils"
	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/web"
)

func utilsFunc(w http.ResponseWriter, r *http.Request) utils.Utils {
	return utils.OtherNew(r)
}

func main() {
	services.Init()
	handler := web.BuildHandler(utilsFunc, true)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	fmt.Printf("Listening on %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
