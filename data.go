package main

import (
	"errors"
)

type Body struct {
	URL string `json:"url"`
}

type StatusMessage struct {
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type ShortenedURLS map[string]string

var ErrorEmptyString = errors.New("cannot add empty string to db")
var ErrorDuplicateShortString = errors.New("duplicate short string entry")

var characters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func (s ShortenedURLS) Add(url, shortLink string) (string, error) {
	if url == "" || shortLink == "" {
		return "", ErrorEmptyString
	}

	if s.IsExists(shortLink) {
		return "", ErrorDuplicateShortString
	}

	existingLink, exists := s.SearchURL(url)
	if exists {
		return existingLink, nil
	}

	s[shortLink] = url
	return shortLink, nil
}

func (s ShortenedURLS) Get(shortURL string) (string, bool) {
	url, ok := s[shortURL]
	return url, ok
}

func (s ShortenedURLS) IsExists(shortURL string) bool {
	_, ok := s.Get(shortURL)
	return ok
}

func (s ShortenedURLS) SearchURL(url string) (string, bool) {
	for k, v := range s {
		if v == url {
			return k, true
		}
	}
	return "", false
}
