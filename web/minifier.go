package web

import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/js"
)

func MinifyJs(js string) string {
	m := getMinifier()
	minified, err := m.String("application/javascript", js)

	if err == nil {
		return minified
	} else {
		return js
	}
}

var minifier *minify.M

func getMinifier() *minify.M {
	if minifier == nil {
		minifier = minify.New()
		minifier.AddFunc("application/javascript", js.Minify)
	}

	return minifier
}
