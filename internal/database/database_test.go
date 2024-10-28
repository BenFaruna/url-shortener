package database_test

import (
	"fmt"
	"github.com/BenFaruna/url-shortener/internal/database"
	"os"
	"slices"
	"testing"

	"github.com/BenFaruna/url-shortener/internal/controller"
)

func TestShortUrlAdd(t *testing.T) {

	db := &database.ShortUrls{}

	t.Run("empty string returns an error", func(t *testing.T) {
		_, err := db.Add("", "")
		if err == nil || err.Error() != "cannot add empty string to db" {
			t.Errorf("Error message not triggered")
		}
	})

	for _, test := range Cases {
		t.Run(fmt.Sprintf("adding %q to db", test.URL), func(t *testing.T) {
			got, err := db.Add(test.URL, test.ShortAddress)
			if err != nil {
				t.Fatal(err)
			}

			if got != test.ShortAddress {
				t.Errorf("expected %q, got %q", test.ShortAddress, got)
			}

			got, _ = db.Get(test.ShortAddress)
			if test.URL != got {
				t.Errorf("expected %q, got %q", test.URL, got)
			}
		})
	}

	for _, test := range Cases {
		t.Run(fmt.Sprintf("adding duplicate URL returns previous short string - %q", test.URL), func(t *testing.T) {
			got, err := db.Add(test.URL, controller.GenerateShortString())
			HandleError(t, err)

			if got != test.ShortAddress {
				t.Errorf("expected %q, got %q", test.ShortAddress, got)
			}
		})
	}

	removeDBFile(t)
}

func TestShortUrlGet(t *testing.T) {
	db := &database.ShortUrls{}

	t.Run("db.Get returns correct result", func(t *testing.T) {
		//for _, test := range Cases {
		//
		//	_, err := db.Add(test.URL, test.ShortAddress)
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//}

		for _, test := range Cases {
			got, ok := db.Get(test.ShortAddress)

			if !ok {
				t.Fatalf("%q not found", test.ShortAddress)
			}

			if got != test.URL {
				t.Errorf("expected %q, got %q", test.URL, got)
			}
		}
	})

	removeDBFile(t)
}

func TestShortUrlGetAll(t *testing.T) {
	db := &database.ShortUrls{}

	for _, entry := range Cases {
		_, _ = db.Add(entry.URL, entry.ShortAddress)
	}

	entries := db.GetAll()

	for _, url := range Cases {
		if !slices.Contains(entries, url) {
			t.Errorf("%v not in %v", url, entries)
		}

	}
	removeDBFile(t)
}

func TestShortUrlSearchURL(t *testing.T) {
	db := &database.ShortUrls{}
	for _, entry := range Cases[:2] {
		_, _ = db.Add(entry.URL, entry.ShortAddress)
	}

	for _, entry := range Cases[:2] {
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

	for _, entry := range []database.URLInfo{
		{URL: "https://gmail.com", ShortAddress: controller.GenerateShortString()},
		{URL: "https://moba.com", ShortAddress: controller.GenerateShortString()},
	} {
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
	removeDBFile(t)
}

func removeDBFile(t testing.TB) {
	t.Cleanup(func() {
		os.RemoveAll("app.db")
	})
}

//func HandleError(t testing.TB, err error) {
//	t.Helper()
//	if err != nil {
//		t.Fatal(err)
//	}
//}
