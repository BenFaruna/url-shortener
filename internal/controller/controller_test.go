package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BenFaruna/url-shortener/internal/controller"
	"github.com/BenFaruna/url-shortener/internal/model"
)

func TestURLShortenerEndpoint(t *testing.T) {
	t.Run("/api/vi/shorten returns the right response", func(t *testing.T) {
		url := "https://pkg.go.dev/net/http/httptest#NewRequest"
		address, err := shortenAddress(url, controller.GenerateShortString)
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

	t.Run("/api/v1/shorten fails on get request", func(t *testing.T) {
		handler := controller.ShortenHandler(controller.GenerateShortString)
		request := httptest.NewRequest(http.MethodGet, "/api/v1/shorten", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != 405 {
			t.Errorf("Expect failed request, got %d", response.Code)
		}
	})

	t.Run("/shorten with duplicate short address returns a 403 error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		url := "https://github.com/BenFaruna"

		server := httptest.NewServer(controller.ShortenHandler(generateSameString))
		defer server.Close()

		// first short string entry
		json.NewEncoder(buf).Encode(model.Body{URL: url})
		http.Post(server.URL+"/shorten", "application/json", buf)

		// duplicate short string request
		json.NewEncoder(buf).Encode(model.Body{URL: url})
		response, err := http.Post(server.URL+"/shorten", "application/json", buf)
		handleError(t, err)

		statusCode := response.StatusCode
		if statusCode != http.StatusForbidden {
			t.Errorf("expected status code 403, got %d", statusCode)
		}

		_, err = buf.ReadFrom(response.Body)
		handleError(t, err)

		if buf.String() != model.ErrorDuplicateShortString.Error() {
			t.Errorf("Expected error message %q, got %q", model.ErrorEmptyString, buf.String())
		}
	})

	t.Run("/api/v1/address/:string returns the full address", func(t *testing.T) {
		var output model.StatusMessage

		url := "https://pkg.go.dev/net/http/httptest#NewRequest"
		address, err := shortenAddress(url, controller.GenerateShortString)
		handleError(t, err)

		handler := controller.GetFullAddressHandler()
		request := httptest.NewRequest(http.MethodGet, "/address/"+address, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		json.NewDecoder(response.Body).Decode(&output)

		if output.Data != url {
			t.Errorf("expected %q, got %q", url, output.Data)
		}
	})

	t.Run("/:string redirects to the correct URL", func(t *testing.T) {
		server := httptest.NewServer(controller.HomeHandler())
		defer server.Close()
		url := server.URL
		address, err := shortenAddress(url, controller.GenerateShortString)
		handleError(t, err)

		request := httptest.NewRequest(http.MethodGet, "/"+address, nil)
		response := httptest.NewRecorder()
		controller.HomeHandler().ServeHTTP(response, request)

		want := response.Result().StatusCode
		loc, err := response.Result().Location()
		handleError(t, err)

		if want != 301 {
			t.Errorf("expected 301, got %d", want)
		}

		if url != loc.String() {
			t.Errorf("expected %q, got %q", url, loc)
		}
	})
}

func shortenAddress(url string, shortStringGenerator func() string) (string, error) {
	buf := &bytes.Buffer{}
	data, err := json.Marshal(model.Body{URL: url})
	if err != nil {
		return "", err
	}
	buf.WriteString(string(data))

	var output model.StatusMessage

	// requirements:
	// url to shorten are sent as part of the body of the request
	// shortened urls are six characters
	// shortened urls are alphabets without special characters
	handler := controller.ShortenHandler(shortStringGenerator)
	request := httptest.NewRequest(http.MethodPost, "/shorten", buf)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	json.Unmarshal(response.Body.Bytes(), &output)
	return output.Data, nil
}

func generateSameString() string {
	return "AbCxYz"
}

func handleError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
