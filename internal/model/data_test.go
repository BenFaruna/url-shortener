package model_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/BenFaruna/url-shortener/internal/controller"
	"github.com/BenFaruna/url-shortener/internal/model"
)

var cases = []model.URLInfo{
	{URL: "https://google.com", ShortAddress: controller.GenerateShortString()},
	{URL: "https://facebook.com", ShortAddress: controller.GenerateShortString()},
	{URL: "https://x.com", ShortAddress: controller.GenerateShortString()},
	{URL: "https://reddit.com", ShortAddress: controller.GenerateShortString()},
}

func TestDBAdd(t *testing.T) {

	db := make(model.ShortenedURLS)

	t.Run("empty string returns an error", func(t *testing.T) {
		_, err := db.Add("", "")
		if err == nil || err.Error() != "cannot add empty string to db" {
			t.Errorf("Error message not triggered")
		}
	})

	for _, test := range cases {
		t.Run(fmt.Sprintf("adding %q to db", test.URL), func(t *testing.T) {
			got, err := db.Add(test.URL, test.ShortAddress)
			if err != nil {
				t.Fatal(err)
			}

			if got != test.ShortAddress {
				t.Errorf("expected %q, got %q", test.ShortAddress, got)
			}

			got = db[test.ShortAddress]
			if test.URL != db[test.ShortAddress] {
				t.Errorf("expected %q, got %q", test.URL, got)
			}
		})
	}

	for _, test := range cases {
		t.Run(fmt.Sprintf("adding duplicate URL returns previous short string - %q", test.URL), func(t *testing.T) {
			got, err := db.Add(test.URL, controller.GenerateShortString())
			handleError(t, err)

			if got != test.ShortAddress {
				t.Errorf("expected %q, got %q", test.ShortAddress, got)
			}
		})

	}
}

func TestDBGet(t *testing.T) {
	db := make(model.ShortenedURLS)

	t.Run("db.Get returns correct result", func(t *testing.T) {
		for _, test := range cases {

			_, err := db.Add(test.URL, test.ShortAddress)
			if err != nil {
				t.Fatal(err)
			}
		}

		for _, test := range cases {
			got, ok := db.Get(test.ShortAddress)

			if !ok {
				t.Fatalf("%q not found", test.ShortAddress)
			}

			if got != test.URL {
				t.Errorf("expected %q, got %q", test.URL, got)
			}
		}
	})
}

func TestDBGetAll(t *testing.T) {
	db := make(model.ShortenedURLS)

	for _, entry := range cases {
		t.Log(entry)
		_, err := db.Add(entry.URL, entry.ShortAddress)
		handleError(t, err)
	}

	entries := db.GetAll()

	for _, url := range cases {
		if !slices.Contains(entries, url) {
			t.Errorf("%v not in %v", url, entries)
		}

	}
}

func TestDBSearchURL(t *testing.T) {
	db := make(model.ShortenedURLS)
	for _, entry := range cases[:2] {
		_, err := db.Add(entry.URL, entry.ShortAddress)
		handleError(t, err)
	}

	for _, entry := range cases[:2] {
		t.Run(fmt.Sprintf("search for existing %q", entry.URL), func(t *testing.T) {
			ShortAddress, exists := db.SearchURL(entry.URL)

			if !exists {
				t.Error("expected true, got false")
			}

			if ShortAddress != entry.ShortAddress {
				t.Errorf("expected %q, got %q", entry.ShortAddress, ShortAddress)
			}
		})
	}

	for _, entry := range cases[2:] {
		t.Run(fmt.Sprintf("search for non existing %q", entry.URL), func(t *testing.T) {
			ShortAddress, exists := db.SearchURL(entry.URL)

			if exists {
				t.Error("expected false, got true")
			}

			if ShortAddress != "" {
				t.Errorf("expected %q, got %q", entry.ShortAddress, ShortAddress)
			}
		})
	}
}

func handleError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
