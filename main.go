package main

import (
	"net/http"
)

var db ShortenedURLS = make(ShortenedURLS)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", HomeHandler())
	mux.Handle("/api/v1/", APIMux())
	// mux.Handle("/static/", StaticMux())
	styles := http.FileServer(http.Dir("./static/css/"))
	mux.Handle("/styles/", http.StripPrefix("/styles/", styles))

	script := http.FileServer(http.Dir("./static/js/"))
	mux.Handle("/scripts/", http.StripPrefix("/scripts/", script))

	if err := http.ListenAndServe(":8000", IncomingRequest(mux)); err != nil {
		panic(err)
	}
}

func APIMux() http.Handler {
	shortenerMux := http.NewServeMux()

	shortenerMux.Handle("/shorten", ShortenHandler(GenerateShortString))
	shortenerMux.Handle("/address/", GetFullAddressHandler())

	return http.StripPrefix("/api/v1", shortenerMux)
}
