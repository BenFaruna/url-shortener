package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
)

// HomeHandler accept requests to the home route and provide responses are redirection for short routes
func HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, "Hello World")
			return
		default:
			shortID := strings.TrimPrefix(r.URL.Path, "/")
			url, ok := db.Get(shortID)
			if !ok {
				errorHandler(w, r, 404)
				return
			}

			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}
	}
}

func ShortenHandler(shortStringFunc func() string) http.Handler {
	return Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shorten" {
			errorHandler(w, r, 404)
			return
		}
		var data Body
		json.NewDecoder(r.Body).Decode(&data)

		shortenedURL := shortStringFunc()

		shortenedURL, err := db.Add(data.URL, shortenedURL)

		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, err.Error())
			return
		}

		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&StatusMessage{
			Message: "url shortened",
			Data:    r.URL.Hostname() + shortenedURL,
		})
	}))
}

func GetFullAddressHandler() http.Handler {

	return Get(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortAddress := strings.TrimPrefix(r.URL.Path, "/address/")

		url, ok := db[shortAddress]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "address does not exist")
			return
		}

		json.NewEncoder(w).Encode(StatusMessage{
			Data:    url,
			Message: "address found",
		})
	}))
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if status == http.StatusNotFound {
		fmt.Fprintf(w, "route %q does not exists", r.URL.Path)
	}
}

func GenerateShortString() string {
	output := ""

	for i := 0; i < 6; i++ {
		n := rand.Intn(51)
		output += string(characters[n])
	}

	return output
}
