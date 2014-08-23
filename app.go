package web

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"bitbucket.org/pkg/inflect"
	"code.google.com/p/go-uuid/uuid"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
)

var templateFunctions = template.FuncMap{
	"ago":  humanize.Time,
	"size": humanize.Bytes,
	"iso":  func(t time.Time) string { return t.Format(time.RFC3339) },
	"host": func(s string) string {
		if u, err := url.Parse(s); err != nil {
			return ""
		} else {
			return u.Host
		}
	},
	"plural": func(n int, s string) string {
		if n == 1 {
			return s
		} else {
			return inflect.Pluralize(s)
		}
	},
}

type pageData struct {
	Title string
}

type AppConfig struct {
	Domain string
	Store  StoreFactory
}

type app struct {
	config AppConfig
	store  Store
	router *mux.Router
}

type appRoute struct {
	*app
	fn func(w http.ResponseWriter, r *http.Request)
}

func (a *appRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.fn(w, r)
}

func NewApp(c AppConfig) (*app, error) {
	store, err := c.Store.build()
	if err != nil {
		return nil, err
	}

	a := &app{
		config: c,
		store:  store,
		router: mux.NewRouter(),
	}

	a.router.NotFoundHandler = http.FileServer(http.Dir("./public"))

	a.router.NewRoute().Name("search_get").Methods("GET").Path("/").Handler(&appRoute{
		app: a,
		fn:  a.getSearch,
	})

	a.router.NewRoute().Name("torrent_get").Methods("GET").Path("/torrent/{id:[0-9a-f]{40}}").Handler(&appRoute{
		app: a,
		fn:  a.getTorrent,
	})

	a.router.NewRoute().Name("submit_get").Methods("GET").Path("/submit").Handler(&appRoute{
		app: a,
		fn:  a.getSubmit,
	})

	a.router.NewRoute().Name("torrent_post").Methods("POST").Path("/torrent").Handler(&appRoute{
		app: a,
		fn:  a.postTorrent,
	})

	a.router.NewRoute().Name("torrent_post_file").Methods("POST").Path("/torrent").Headers("content-type", "application/x-bittorrent").Handler(&appRoute{
		app: a,
		fn:  a.postTorrentFile,
	})

	a.router.NewRoute().Name("help_get").Methods("GET").Path("/help").Handler(&appRoute{
		app: a,
		fn:  a.getHelp,
	})

	return a, nil
}

type wrappedWriter struct {
	http.ResponseWriter
	status int
}

func newWrappedWriter(w http.ResponseWriter) *wrappedWriter {
	return &wrappedWriter{
		ResponseWriter: w,
		status:         200,
	}
}

func (w *wrappedWriter) WriteHeader(status int) {
	w.status = status

	w.ResponseWriter.WriteHeader(status)
}

func (h *app) ServeHTTP(_w http.ResponseWriter, r *http.Request) {
	w := newWrappedWriter(_w)

	id := uuid.New()
	t := time.Now()

	log.Printf("request time=%#v id=%s method=%s path=%#v", time.Now().Format(time.RFC3339), id, r.Method, r.URL.String())
	defer func() {
		log.Printf("response time=%#v id=%s status=%d duration=%s", time.Now().Format(time.RFC3339), id, w.status, time.Since(t))
	}()

	h.router.ServeHTTP(w, r)
}
