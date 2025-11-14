package web

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Templates returns the parsed HTML templates for server-side rendering.
func Templates() (*template.Template, error) {
	return template.New("base.gohtml").ParseFS(templateFS, "templates/*.gohtml")
}

// Static returns the embedded static filesystem with CSS/JS/WASM assets.
func Static() (fs.FS, error) {
	return fs.Sub(staticFS, "static")
}
