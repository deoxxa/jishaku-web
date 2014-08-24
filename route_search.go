package web

import (
	"html/template"
	"net/http"
	"path"

	"github.com/thraxil/paginate"
)

type searchData struct {
	pageData
	CurrentQuery string
	Pages        *paginate.Paginator
	Page         paginate.Page
	Torrents     []*Torrent
}

type searchItems struct {
	app   *app
	query string
	total int
}

func newSearchItems(app *app, query string) (*searchItems, error) {
	c := &searchItems{
		app:   app,
		query: query,
	}

	r, err := app.store.Count(query)
	if err != nil {
		return nil, err
	}

	c.total = r

	return c, nil
}

func (c *searchItems) TotalItems() int {
	return c.total
}

func (c *searchItems) ItemRange(offset, count int) []interface{} {
	var l []interface{}

	torrents, err := c.app.store.Search(c.query, offset, count)
	if err != nil {
		panic(err)
	}

	l = make([]interface{}, len(torrents))

	for i, torrent := range torrents {
		l[i] = torrent
	}

	return l
}

var searchTemplate = template.Must(template.New("template").Funcs(templateFunctions).ParseFiles(path.Join(root, "templates/layout.html"), path.Join(root, "templates/page_search.html")))

func (a *app) getSearch(w http.ResponseWriter, r *http.Request) {
	items, err := newSearchItems(a, r.URL.Query().Get("q"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	paginator := paginate.NewPaginator(items, 25)
	page := paginator.GetPage(r)

	torrents := make([]*Torrent, 0)

	for _, e := range page.Items() {
		if t, ok := e.(*Torrent); !ok {
			http.Error(w, "error getting torrents", 500)
			return
		} else {
			torrents = append(torrents, t)
		}
	}

	pageData := searchData{
		pageData: pageData{
			Title: "Search Results",
		},
		CurrentQuery: r.URL.Query().Get("q"),
		Pages:        paginator,
		Page:         page,
		Torrents:     torrents,
	}

	if err := searchTemplate.ExecuteTemplate(w, "layout", pageData); err != nil {
		http.Error(w, err.Error(), 500)
	}
}
