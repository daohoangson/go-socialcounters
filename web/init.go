package web

import (
	"net/http"
	"os"

	"github.com/daohoangson/go-socialcounters/utils"
)

var uf utils.UtilsFunc

func HttpInit(utilsFunc utils.UtilsFunc) {
	uf = utilsFunc

	fs := http.FileServer(httpFs{http.Dir("public")})
	http.Handle("/css/", fs)
	http.Handle("/img/", fs)

	http.HandleFunc("/", httpRedirect)
	http.HandleFunc("/js/all.js", httpAllJs)
	http.HandleFunc("/js/data.json", httpDataJson)
	http.HandleFunc("/js/jquery.plugin.js", httpJqueryPluginJs)
}

// start of https://groups.google.com/forum/#!msg/golang-nuts/bStLPdIVM6w/AXLz0hNqCrUJ
type httpFs struct {
	fs http.FileSystem
}

func (fs httpFs) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return httpFile{f}, nil
}

type httpFile struct {
	http.File
}

func (f httpFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

// end of https://groups.google.com/forum/#!msg/golang-nuts/bStLPdIVM6w/AXLz0hNqCrUJ

func httpRedirect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", "https://daohoangson.github.io/go-socialcounters/")
	w.WriteHeader(http.StatusMovedPermanently)
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
