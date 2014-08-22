package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/olivere/elastic"
	"github.com/thraxil/paginate"
)

type searchData struct {
	pageData
	CurrentQuery string
	Pages        *paginate.Paginator
	Page         paginate.Page
	Torrents     []torrent
}

type searchItems struct {
	app   *app
	query string
	total int
}

func newSearchItems(app *app, query string) *searchItems {
	c := &searchItems{
		app:   app,
		query: query,
	}

	var q elastic.Query
	if c.query != "" {
		q = elastic.NewQueryStringQuery(c.query)
	} else {
		q = elastic.NewMatchAllQuery()
	}

	if r, err := c.app.es.Count(c.app.config.esIndex).Type("torrent").Query(q).Do(); err != nil {
		c.total = 0
	} else {
		c.total = int(r)
	}

	return c
}

func (c *searchItems) TotalItems() int {
	return c.total
}

func (c *searchItems) ItemRange(offset, count int) []interface{} {
	var q elastic.Query
	if c.query != "" {
		q = elastic.NewQueryStringQuery(c.query)
	} else {
		q = elastic.NewMatchAllQuery()
	}

	var l []interface{}

	if r, err := c.app.es.Search(c.app.config.esIndex).Type("torrent").Query(q).From(offset).Size(count).Do(); err == nil {
		l = make([]interface{}, len(r.Hits.Hits))

		for i, d := range r.Hits.Hits {
			var t torrent
			json.Unmarshal(*d.Source, &t)
			l[i] = t
		}
	}

	return l
}

var searchTemplate = template.Must(template.New("template").Funcs(templateFunctions).ParseFiles("templates/layout.html", "templates/page_search.html"))

func (a *app) getSearch(w http.ResponseWriter, r *http.Request) {
	items := newSearchItems(a, r.URL.Query().Get("q"))
	paginator := paginate.NewPaginator(items, 25)
	page := paginator.GetPage(r)

	torrents := make([]torrent, 0)

	for _, e := range page.Items() {
		if t, ok := e.(torrent); ok {
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
		http.Error(w, "uh oh", 500)
	}
}
