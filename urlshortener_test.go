package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	urlshortener "github.com/BenFaruna/url-shortener"
)

func TestURLShortenerEndpoint(t *testing.T) {
	t.Run("test / returns the correct string", func(t *testing.T) {
		handler := urlshortener.HomeHandler()

		want := "Hello World"

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		got := response.Body.String()

		if got != want {
			t.Errorf("want %q, got %q", want, got)
		}
	})

	t.Run("test /api/vi/shorten returns the right response", func(t *testing.T) {
		url := "https://pkg.go.dev/net/http/httptest#NewRequest"
		address, err := shortenAddress(url)
		if err != nil {
			t.Fatal(err)
		}

		got := len(address)

		if got != 6 {
			t.Errorf("expected url length %d, got %d", 6, got)
		}
		if strings.ContainsAny(address, "-_/%!@#$^&*()=+1234567890") {
			t.Errorf("%s contains numbers or special charcters", address)
		}
	})

	t.Run("test /api/v1/shorten fails on get request", func(t *testing.T) {
		handler := urlshortener.ShortenHandler()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/shorten", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != 405 {
			t.Errorf("Expect failed request, got %d", response.Code)
		}
	})

	t.Run("test /api/v1/address/:string returns the full address", func(t *testing.T) {
		var output urlshortener.StatusMessage

		url := "https://pkg.go.dev/net/http/httptest#NewRequest"
		address, err := shortenAddress(url)
		if err != nil {
			t.Fatal(err)
		}

		handler := urlshortener.GetFullAddressHandler()
		request := httptest.NewRequest(http.MethodGet, "/address/"+address, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		json.NewDecoder(response.Body).Decode(&output)

		if output.Data != url {
			t.Errorf("expected %q, got %q", url, output.Data)
		}
	})

	t.Run("test /:string redirects to the correct URL", func(t *testing.T) {
		server := httptest.NewServer(urlshortener.HomeHandler())
		defer server.Close()
		url := server.URL
		address, err := shortenAddress(url)
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequest(http.MethodGet, "/"+address, nil)
		response := httptest.NewRecorder()
		urlshortener.HomeHandler().ServeHTTP(response, request)

		want := response.Result().StatusCode
		loc, err := response.Result().Location()

		if err != nil {
			t.Fatal(err)
		}

		if want != 301 {
			t.Errorf("expected 301, got %d", want)
		}

		if url != loc.String() {
			t.Errorf("expected %q, got %q", url, loc)
		}
	})
}

func shortenAddress(url string) (string, error) {
	buf := &bytes.Buffer{}
	data, err := json.Marshal(urlshortener.Body{Url: url})
	if err != nil {
		return "", err
	}
	buf.WriteString(string(data))

	var output urlshortener.StatusMessage

	// requirements:
	// url to shorten are sent as part of the body of the request
	// shortened urls are six characters
	// shortened urls are alphabets without special characters
	handler := urlshortener.ShortenHandler()
	request := httptest.NewRequest(http.MethodPost, "/shorten", buf)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	json.Unmarshal(response.Body.Bytes(), &output)
	return output.Data, nil
}
