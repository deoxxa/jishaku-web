package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

type getTorrentData struct {
	pageData
	Torrent torrent
}

var torrentTemplate = template.Must(template.New("template").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_torrent.html"))

func (a *app) getTorrent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := a.es.Get().Index(a.config.esIndex).Id(vars["id"]).Do()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !res.Found {
		http.Error(w, "not found", 404)
		return
	}

	var t torrent
	if err := json.Unmarshal(*res.Source, &t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	pageData := getTorrentData{
		pageData: pageData{
			Title: t.Name,
		},
		Torrent: t,
	}

	if err := torrentTemplate.ExecuteTemplate(w, "layout", pageData); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
