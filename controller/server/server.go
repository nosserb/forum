package server

import (
	"html/template"
	"log"
)

func ParseTemplates(pattern string) *template.Template {
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		log.Fatalf("Erreur parsing templates: %v", err)
	}
	return tmpl
}
