package web

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/daohoangson/go-socialcounters/utils"
	"github.com/rs/cors"
)

var uf utils.UtilsFunc

func BuildHandler(utilsFunc utils.UtilsFunc, doGzip bool) http.Handler {
	uf = utilsFunc
	mux := http.NewServeMux()

	fs := http.FileServer(httpFs{http.Dir("public")})
	mux.Handle("/css/", fs)
	mux.Handle("/html/", fs)
	mux.Handle("/img/", fs)
	mux.Handle("/favicon.ico", fs)

	mux.HandleFunc("/", httpRedirect)
	mux.HandleFunc("/js/all.js", httpAllJs)
	mux.HandleFunc("/js/data.json", httpDataJson)
	mux.HandleFunc("/js/jquery.plugin.js", httpJqueryPluginJs)
	mux.HandleFunc("/v2/js/data.json", httpDataJson2)
	mux.HandleFunc("/v2/js/history.json", httpHistoryJson)

	mux.HandleFunc("/config", httpConfig)

	handler := cors.Default().Handler(mux)
	if !doGzip {
		return handler
	}

	return gziphandler.GzipHandler(handler)
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
	DataJson(u, w, r, true)
}

func httpJqueryPluginJs(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	JqueryPluginJs(u, w, r)
}

func httpDataJson2(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	DataJson(u, w, r, false)
}

func httpHistoryJson(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	HistoryJson(u, w, r)
}

func httpConfig(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	if r.Method == "GET" {
		ConfigGet(u, w, r)
	} else {
		ConfigPost(u, w, r)
	}
}
