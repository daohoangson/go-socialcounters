package web

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/rs/cors"
	"github.com/daohoangson/go-socialcounters/utils"
)

var uf utils.UtilsFunc

func BuildHandler(utilsFunc utils.UtilsFunc, doGzip bool) http.Handler {
	uf = utilsFunc
	mux := http.NewServeMux()

	fs := http.FileServer(httpFs{http.Dir("public")})
	mux.Handle("/css/", fs)
	mux.Handle("/img/", fs)

	mux.HandleFunc("/", httpRedirect)
	mux.HandleFunc("/js/all.js", httpAllJs)
	mux.HandleFunc("/js/data.json", httpDataJson)
	mux.HandleFunc("/v2/js/data.json", httpDataJson2)
	mux.HandleFunc("/js/jquery.plugin.js", httpJqueryPluginJs)

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

func httpDataJson2(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	DataJson(u, w, r, false)
}

func httpJqueryPluginJs(w http.ResponseWriter, r *http.Request) {
	u := uf(w, r)
	JqueryPluginJs(u, w, r)
}
