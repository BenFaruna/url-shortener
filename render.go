package main

import (
	"embed"
	"html/template"
	"io"
)

var (
	//go:embed "templates/*"
	indexTemplate embed.FS
)

type IndexRenderer struct {
	templ *template.Template
}

func NewIndexRenderer() (*IndexRenderer, error) {
	templ, err := template.ParseFS(indexTemplate, "templates/*.gohtml")

	if err != nil {
		return nil, err
	}
	return &IndexRenderer{templ: templ}, nil
}

func (r *IndexRenderer) Render(w io.Writer, data []URLInfo) error {
	if err := r.templ.ExecuteTemplate(w, "index.gohtml", data); err != nil {
		return err
	}
	return nil
}
