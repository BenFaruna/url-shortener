package controller_test

import (
	"bytes"
	"testing"

	"github.com/BenFaruna/url-shortener/internal/controller"
	"github.com/BenFaruna/url-shortener/internal/model"
	approval "github.com/approvals/go-approval-tests"
)

var (
	entries = []model.URLInfo{
		{URL: "https://google.com", ShortAddress: "wxYabC"},
		{URL: "https://faceboox.com", ShortAddress: "sFXZul"},
		{URL: "https://go.dev", ShortAddress: "YaCChm"},
	}
)

func TestIndexRender(t *testing.T) {
	buf := bytes.Buffer{}

	indexRenderer, err := controller.NewIndexRenderer()
	handleError(t, err)

	indexRenderer.Render(&buf, entries)

	approval.VerifyString(t, buf.String())
}
