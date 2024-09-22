package controller

import (
	"embed"
	"html/template"
	"io"

	"github.com/BenFaruna/url-shortener/internal/model"
)

var (
	//go:embed "templates/*"
	indexTemplate embed.FS
)

type Renderer struct {
	templ *template.Template
}

func NewRenderer() (*Renderer, error) {
	templ, err := template.ParseFS(indexTemplate, "templates/*.gohtml")

	if err != nil {
		return nil, err
	}
	return &Renderer{templ: templ}, nil
}

func (r *Renderer) RenderData(w io.Writer, filename string, data []model.URLInfo) error {
	if err := r.templ.ExecuteTemplate(w, filename, data); err != nil {
		return err
	}
	return nil
}

func (r *Renderer) Render(w io.Writer, filename string, data interface{}) error {
	if err := r.templ.ExecuteTemplate(w, filename, data); err != nil {
		return err
	}
	return nil
}
