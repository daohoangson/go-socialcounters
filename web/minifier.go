package web

import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
)

func MinifyCss(css string) string {
	m := getMinifier()
	minified, err := m.String("text/css", css)

	if err == nil {
		return minified
	} else {
		return css
	}
}

func MinifyJs(js string) string {
	m := getMinifier()
	minified, err := m.String("application/javascript", js)

	if err == nil {
		return minified
	} else {
		return js
	}
}

func MinifyJson(json string) string {
	m := getMinifier()
	minified, err := m.String("application/json", json)

	if err == nil {
		return minified
	} else {
		return json
	}
}

func MinifySvg(svg string) string {
	m := getMinifier()
	minified, err := m.String("image/svg+xml", svg)

	if err == nil {
		return minified
	} else {
		return svg
	}
}

var minifier *minify.M

func getMinifier() *minify.M {
	if minifier == nil {
		minifier = minify.New()
		minifier.AddFunc("text/css", css.Minify)
		minifier.AddFunc("application/javascript", js.Minify)
		minifier.AddFunc("application/json", json.Minify)
		minifier.AddFunc("image/svg+xml", svg.Minify)
	}

	return minifier
}
