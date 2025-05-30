package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

// Exported functions and variables
var Tc = make(map[string]*template.Template) // Capitalize tc

// Embed all templates from the `templates` folder in the root.
var TemplateFS embed.FS

// Exported functions
func Render(w http.ResponseWriter, t string) {
	var tmpl *template.Template
	var err error

	_, inMap := Tc[t]
	if !inMap {
		log.Println("Parsing template and adding to cache")
		err = CreateTemplateCache(t)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		log.Println("Using cached template")
	}

	tmpl = Tc[t]

	var data struct {
		BrokerURL string
	}

	data.BrokerURL = os.Getenv("BROKER_URL")

	if err := tmpl.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func CreateTemplateCache(t string) error {
	templates := []string{
		fmt.Sprintf("templates/%s", t),
		"templates/header.partial.gohtml",
		"templates/footer.partial.gohtml",
		"templates/base.layout.gohtml",
	}

	tmpl, err := template.ParseFS(TemplateFS, templates...)
	if err != nil {
		return err
	}

	Tc[t] = tmpl

	return nil
}
