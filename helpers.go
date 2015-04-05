package main

import (
	"html/template"
	"net/url"
	"time"

	"bitbucket.org/pkg/inflect"
	"github.com/dustin/go-humanize"
)

var templateFunctions = template.FuncMap{
	"ago":  humanize.Time,
	"size": func(v int64) string { return humanize.Bytes(uint64(v)) },
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

type templateData struct {
	Title        string
	CurrentQuery string
}

type searchTemplateData struct {
	templateData
	Entries []Entry
}

type viewTemplateData struct {
	templateData
	Entry Entry
}
