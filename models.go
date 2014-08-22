package main

import (
	"html/template"
	"net/url"
	"time"
)

type torrent struct {
	Hash      string
	FirstSeen time.Time
	Name      string
	Comment   string
	CreatedBy struct {
		Client  string
		Version string
	}
	CreationDate time.Time
	Size         uint64
	Trackers     []string
	Files        []struct {
		Name string
		Size uint64
	}
	Locations []string
}

func (t *torrent) MagnetURI() (template.URL, error) {
	q := url.Values{
		"xt": {"urn:btih:" + t.Hash},
		"dn": {t.Name},
		"tr": t.Trackers,
	}

	return template.URL("magnet:" + q.Encode()), nil
}
