// +build appengine

package main

import (
	"net/http"

	"github.com/daohoangson/go-socialcounters/services"
	"github.com/daohoangson/go-socialcounters/utils"
	"github.com/daohoangson/go-socialcounters/web"
)

func utilsFunc(w http.ResponseWriter, r *http.Request) utils.Utils {
	return utils.GaeNew(r)
}

func init() {
	services.Init()

	handler := web.BuildHandler(utilsFunc, false)
	http.Handle("/", handler)
}
