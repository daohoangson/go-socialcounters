package web

import (
	"net/http"

	"github.com/daohoangson/go-socialcounters/utils"
)

var uf utils.UtilsFunc

func HttpInit(utilsFunc utils.UtilsFunc) {
	uf = utilsFunc

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.HandleFunc("/js/all.js", httpAllJs)
	http.HandleFunc("/js/data.json", httpDataJson)
	http.HandleFunc("/js/jquery.plugin.js", httpJqueryPluginJs)
}

func httpAllJs(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	AllJs(u, w, r)
}

func httpDataJson(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	DataJson(u, w, r)
}

func httpJqueryPluginJs(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	JqueryPluginJs(u, w, r)
}
