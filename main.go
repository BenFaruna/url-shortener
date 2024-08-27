package main

import (
	"net/http"
)

var db ShortenedURLS = make(ShortenedURLS)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", HomeHandler())
	mux.Handle("/api/v1/", APIMux())

	if err := http.ListenAndServe(":8000", IncomingRequest(mux)); err != nil {
		panic(err)
	}
}

func APIMux() http.Handler {
	shortenerMux := http.NewServeMux()

	shortenerMux.Handle("/shorten", ShortenHandler(GenerateShortString))
	shortenerMux.Handle("/address/", GetFullAddressHandler())

	return http.StripPrefix("/api/v1", shortenerMux)
	// return shortenerMux
}
