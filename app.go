package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/zeebo/bencode"
)

const (
	QUERY_LIST   = `select "info_hash", "name", "size", "first_seen" from "torrents" order by "first_seen" desc limit 50`
	QUERY_SEARCH = `select "info_hash", "name", "size", "first_seen" from "torrents" where "name" ilike any($1) order by "first_seen" desc limit 50`
	QUERY_VIEW   = `select "info_hash", "name", "size", "first_seen", "files", "trackers", "locations" from "torrents" where "info_hash" = $1`
	QUERY_INSERT = `insert into "torrents" ("info_hash", "name", "size", "first_seen", "files", "trackers", "locations") values ($1, $2, $3, $4, $5, $6, $7)`
)

var (
	template_search = template.Must(template.New("search").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_search.html"))
	template_view   = template.Must(template.New("view").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_view.html"))
	template_submit = template.Must(template.New("submit").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_submit.html"))
	template_help   = template.Must(template.New("help").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_help.html"))
)

type app struct {
	db *pgx.Conn
	r  *mux.Router
}

func newApp(db *pgx.Conn) (app, error) {
	r := mux.NewRouter()

	a := app{db, r}

	r.NewRoute().Methods("GET").Path("/").HandlerFunc(a.Search)
	r.NewRoute().Methods("GET").Path("/torrent/{id:[0-9a-f]{40}}").HandlerFunc(a.View)
	r.NewRoute().Methods("GET").Path("/submit").HandlerFunc(a.ShowSubmit)
	r.NewRoute().Methods("POST").Path("/torrent").HandlerFunc(a.Submit)
	r.NewRoute().Methods("GET").Path("/help").HandlerFunc(a.ShowHelp)

	return a, nil
}

func (a app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.r.ServeHTTP(w, r)
}

func (a app) Search(w http.ResponseWriter, r *http.Request) {
	var rows *pgx.Rows

	if q := r.URL.Query().Get("q"); q == "" {
		if r, err := a.db.Query(QUERY_LIST); err != nil {
			panic(err)
		} else {
			rows = r
		}
	} else {
		words := strings.Split(q, " ")
		for i, w := range words {
			words[i] = "%" + w + "%"
		}

		if r, err := a.db.Query(QUERY_SEARCH, words); err != nil {
			panic(err)
		} else {
			rows = r
		}
	}

	l := make([]Entry, 0, 100)
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.InfoHash, &e.Name, &e.Size, &e.FirstSeen); err != nil {
			panic(err)
		}

		l = append(l, e)
	}

	d := searchTemplateData{
		templateData: templateData{
			Title:        "Search",
			CurrentQuery: r.URL.Query().Get("q"),
		},
		Entries: l,
	}

	if err := template_search.ExecuteTemplate(w, "layout", d); err != nil {
		panic(err)
	}
}

func (a app) View(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var e Entry
	if err := a.db.QueryRow(QUERY_VIEW, vars["id"]).Scan(&e.InfoHash, &e.Name, &e.Size, &e.FirstSeen, &e.Files, &e.Trackers, &e.Locations); err != nil {
		if err == pgx.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	d := viewTemplateData{
		templateData: templateData{
			Title:        e.Name,
			CurrentQuery: r.URL.Query().Get("q"),
		},
		Entry: e,
	}

	if err := template_view.ExecuteTemplate(w, "layout", d); err != nil {
		panic(err)
	}
}

func (a app) ShowSubmit(w http.ResponseWriter, r *http.Request) {
	d := templateData{
		Title:        "Submit Entry",
		CurrentQuery: r.URL.Query().Get("q"),
	}

	if err := template_submit.ExecuteTemplate(w, "layout", d); err != nil {
		panic(err)
	}
}

func (a app) Submit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1024 * 128); err != nil {
		panic(err)
	}

	u := r.Form.Get("url")

	res, err := http.Get(u)
	if err != nil {
		panic(err)
	}

	var t Torrent
	if err := bencode.NewDecoder(res.Body).Decode(&t); err != nil {
		panic(err)
	}
	if err := bencode.NewDecoder(bytes.NewReader(t.RawInfo)).Decode(&t.Info); err != nil {
		panic(err)
	}

	h := sha1.New()
	if _, err := h.Write(t.RawInfo); err != nil {
		panic(err)
	}
	d := h.Sum(nil)

	e := Entry{
		InfoHash:  hex.EncodeToString(d),
		Name:      t.Info.Name,
		Locations: []string{u},
		FirstSeen: time.Now(),
	}

	if len(t.Info.Files) == 0 {
		e.Files = []TorrentFile{
			{
				Path:   t.Info.Name,
				Length: t.Info.Length,
			},
		}

		e.Size = t.Info.Length
	} else {
		for _, f := range t.Info.Files {
			e.Files = append(e.Files, TorrentFile{
				Path:   f.Path,
				Length: f.Length,
			})

			e.Size += f.Length
		}
	}

	if t.Announce != "" {
		e.Trackers = append(e.Trackers, t.Announce)
	}

	for _, p := range t.AnnounceList {
		for _, p := range p {
			e.Trackers = append(e.Trackers, p)
		}
	}

	if _, err := a.db.Exec(QUERY_INSERT, e.InfoHash, e.Name, e.Size, e.FirstSeen, e.Files.String(), e.Trackers, e.Locations); err != nil {
		if strings.Contains(err.Error(), `duplicate key value violates unique constraint "torrents_pkey"`) {
			w.Header().Set("location", "/torrent/"+e.InfoHash)
			w.WriteHeader(http.StatusFound)
			return
		}

		panic(err)
	}

	w.Header().Set("location", "/torrent/"+e.InfoHash)
	w.WriteHeader(http.StatusSeeOther)
}

func (a app) ShowHelp(w http.ResponseWriter, r *http.Request) {
	d := templateData{
		Title:        "Help",
		CurrentQuery: r.URL.Query().Get("q"),
	}

	if err := template_help.ExecuteTemplate(w, "layout", d); err != nil {
		panic(err)
	}
}
