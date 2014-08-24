package web

import (
	"html/template"
	"net/http"
	"path"
)

var helpTemplate = template.Must(template.New("template").Funcs(templateFunctions).ParseFiles(path.Join(root, "templates/layout.html"), path.Join(root, "templates/page_help.html")))

func (a *app) getHelp(w http.ResponseWriter, r *http.Request) {
	pageData := pageData{
		Title: "Help",
	}

	if err := helpTemplate.ExecuteTemplate(w, "layout", pageData); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
