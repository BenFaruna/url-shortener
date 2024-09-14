package controller

import (
	"embed"
	"html/template"
	"io"

	"github.com/BenFaruna/url-shortener/internal/model"
)

var (
	IndexTemplate embed.FS
)

type IndexRenderer struct {
	templ *template.Template
}

func NewIndexRenderer() (*IndexRenderer, error) {
	templ, err := template.ParseFS(IndexTemplate, "templates/*.gohtml")

	if err != nil {
		return nil, err
	}
	return &IndexRenderer{templ: templ}, nil
}

func (r *IndexRenderer) Render(w io.Writer, data []model.URLInfo) error {
	if err := r.templ.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		return err
	}
	return nil
}
