package main

type Body struct {
	Url string `json:"url"`
}

type StatusMessage struct {
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

type ShortenedURLS map[string]string

var characters string = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
