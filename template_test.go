package main_test

import (
	"bytes"
	"testing"

	urlshortener "github.com/BenFaruna/url-shortener"
	approval "github.com/approvals/go-approval-tests"
)

var (
	entries = []urlshortener.URLInfo{
		{"https://google.com", "wxYabC"},
		{"https://faceboox.com", "sFXZul"},
		{"https://go.dev", "YaCChm"},
	}
)

func TestIndexRender(t *testing.T) {
	buf := bytes.Buffer{}

	indexRenderer, err := urlshortener.NewIndexRenderer()
	handleError(t, err)

	indexRenderer.Render(&buf, entries)

	approval.VerifyString(t, buf.String())
}
