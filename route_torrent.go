package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/deoxxa/libtorrent"
	"github.com/gorilla/mux"
)

type getTorrentData struct {
	pageData
	Torrent torrent
}

var torrentTemplate = template.Must(template.New("template").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_torrent.html"))

func (a *app) getTorrent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	res, err := a.es.Get().Index(a.config.esIndex).Type("torrent").Id(vars["id"]).Do()
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

func (a *app) consumeTorrentStream(r io.Reader) (*torrent, error) {
	meta, err := libtorrent.ParseMetainfo(r)
	if err != nil {
		return nil, err
	}

	t := torrent{
		Hash:      fmt.Sprintf("%x", meta.InfoHash),
		FirstSeen: time.Now(),
		Name:      meta.Name,
		Files:     meta.Files,
		Trackers:  meta.AnnounceList,
	}

	for _, f := range t.Files {
		t.Size += f.Length
	}

	if _, err := a.es.Index().Index(a.config.esIndex).Type("torrent").Id(t.Hash).BodyJson(t).Do(); err != nil {
		return nil, err
	}

	return &t, nil
}

func (a *app) postTorrent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1024 * 1024 * 32); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if r.Form.Get("url") == "" {
		http.Error(w, "url parameter required", 406)
		return
	}

	res, err := http.Get(r.Form.Get("url"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	t, err := a.consumeTorrentStream(res.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	u, err := a.router.GetRoute("torrent_get").URL("id", t.Hash)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Add("location", u.String())
	w.WriteHeader(201)
}

func (a *app) postTorrentFile(w http.ResponseWriter, r *http.Request) {
	t, err := a.consumeTorrentStream(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	u, err := a.router.GetRoute("torrent_get").URL("id", t.Hash)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Add("location", u.String())
	w.WriteHeader(201)
}
