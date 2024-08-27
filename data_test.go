package main_test

import (
	"fmt"
	"testing"

	urlshortener "github.com/BenFaruna/url-shortener"
)

var cases = []struct{ url, shortUrl string }{
	{url: "https://google.com", shortUrl: urlshortener.GenerateShortString()},
	{url: "https://facebook.com", shortUrl: urlshortener.GenerateShortString()},
	{url: "https://x.com", shortUrl: urlshortener.GenerateShortString()},
	{url: "https://reddit.com", shortUrl: urlshortener.GenerateShortString()},
}

func TestDBAdd(t *testing.T) {

	db := make(urlshortener.ShortenedURLS)

	t.Run("empty string returns an error", func(t *testing.T) {
		_, err := db.Add("", "")
		if err == nil || err.Error() != "cannot add empty string to db" {
			t.Errorf("Error message not triggered")
		}
	})

	for _, test := range cases {
		t.Run(fmt.Sprintf("adding %q to db", test.url), func(t *testing.T) {
			got, err := db.Add(test.url, test.shortUrl)
			if err != nil {
				t.Fatal(err)
			}

			if got != test.shortUrl {
				t.Errorf("expected %q, got %q", test.shortUrl, got)
			}

			got = db[test.shortUrl]
			if test.url != db[test.shortUrl] {
				t.Errorf("expected %q, got %q", test.url, got)
			}
		})
	}

	for _, test := range cases {
		t.Run(fmt.Sprintf("adding duplicate url returns previous short string - %q", test.url), func(t *testing.T) {
			got, err := db.Add(test.url, urlshortener.GenerateShortString())
			handleError(t, err)

			if got != test.shortUrl {
				t.Errorf("expected %q, got %q", test.shortUrl, got)
			}
		})

	}
}

func TestDBGet(t *testing.T) {
	db := make(urlshortener.ShortenedURLS)

	t.Run("db.Get returns correct result", func(t *testing.T) {
		for _, test := range cases {

			_, err := db.Add(test.url, test.shortUrl)
			if err != nil {
				t.Fatal(err)
			}
		}

		for _, test := range cases {
			got, ok := db.Get(test.shortUrl)

			if !ok {
				t.Fatalf("%q not found", test.shortUrl)
			}

			if got != test.url {
				t.Errorf("expected %q, got %q", test.url, got)
			}
		}
	})
}

func TestDBSearchURL(t *testing.T) {
	db := make(urlshortener.ShortenedURLS)
	for _, entry := range cases[:2] {
		_, err := db.Add(entry.url, entry.shortUrl)
		handleError(t, err)
	}

	for _, entry := range cases[:2] {
		t.Run(fmt.Sprintf("search for existing %q", entry.url), func(t *testing.T) {
			shortUrl, exists := db.SearchURL(entry.url)

			if !exists {
				t.Error("expected true, got false")
			}

			if shortUrl != entry.shortUrl {
				t.Errorf("expected %q, got %q", entry.shortUrl, shortUrl)
			}
		})
	}

	for _, entry := range cases[2:] {
		t.Run(fmt.Sprintf("search for non existing %q", entry.url), func(t *testing.T) {
			shortUrl, exists := db.SearchURL(entry.url)

			if exists {
				t.Error("expected false, got true")
			}

			if shortUrl != "" {
				t.Errorf("expected %q, got %q", entry.shortUrl, shortUrl)
			}
		})
	}
}
