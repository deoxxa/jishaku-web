package main

import (
	"html/template"
	"net/url"
	"time"
)

type torrent struct {
	Hash      string
	Name      string
	Size      uint64
	FirstSeen time.Time
	Files     []struct {
		Length uint64
		Path   string
	}
	Trackers  []string
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
