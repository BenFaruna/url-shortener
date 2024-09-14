package main

import (
	"embed"
	"net/http"

	"github.com/BenFaruna/url-shortener/internal/controller"
)

var (
	//go:embed "templates/*"
	indexTemplate embed.FS
)

func main() {
	controller.IndexTemplate = indexTemplate
	mux := http.NewServeMux()

	mux.Handle("/", controller.HomeHandler())
	mux.Handle("/api/v1/", APIMux())

	styles := http.FileServer(http.Dir("./static/css/"))
	mux.Handle("/styles/", http.StripPrefix("/styles/", styles))

	script := http.FileServer(http.Dir("./static/js/"))
	mux.Handle("/scripts/", http.StripPrefix("/scripts/", script))

	if err := http.ListenAndServe(":8000", controller.IncomingRequest(mux)); err != nil {
		panic(err)
	}
}

func APIMux() http.Handler {
	shortenerMux := http.NewServeMux()

	shortenerMux.Handle("/shorten", controller.ShortenHandler(controller.GenerateShortString))
	shortenerMux.Handle("/address/", controller.GetFullAddressHandler())

	return http.StripPrefix("/api/v1", shortenerMux)
}
