package controller_test

import (
	"bytes"
	"github.com/BenFaruna/url-shortener/internal/database"
	"testing"

	"github.com/BenFaruna/url-shortener/internal/controller"
	approval "github.com/approvals/go-approval-tests"
)

var (
	entries = []database.URLInfo{
		{URL: "https://google.com", ShortAddress: "wxYabC"},
		{URL: "https://faceboox.com", ShortAddress: "sFXZul"},
		{URL: "https://go.dev", ShortAddress: "YaCChm"},
	}
)

func TestIndexRender(t *testing.T) {
	buf := bytes.Buffer{}

	indexRenderer, err := controller.NewRenderer()
	handleError(t, err)

	indexRenderer.RenderData(&buf, "index.gohtml", entries)

	approval.VerifyString(t, buf.String())
}
