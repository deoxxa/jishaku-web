package main

import (
	"html/template"
	"net/http"
)

var submitTemplate = template.Must(template.New("template").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_submit.html"))

func (a *app) getSubmit(w http.ResponseWriter, r *http.Request) {
	pageData := pageData{
		Title: "Submit Torrent",
	}

	if err := submitTemplate.ExecuteTemplate(w, "layout", pageData); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
