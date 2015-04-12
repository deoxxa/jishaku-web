package scraper

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zeebo/bencode"
)

type httpTrackerScrapeResponse struct {
	Files map[string]struct {
		Incomplete int
		Complete   int
		Downloaded int
	}
	FailureReason string `bencode:"failure reason"`
	Flags         struct {
		MinRequestInterval int `bencode:"min_request_interval"`
	}
}

type httpTracker struct {
	u *url.URL
	c *http.Client
}

func newHTTPTracker(u *url.URL) (Backend, error) {
	t := httpTracker{
		u: u,
		c: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   time.Second * 5,
		},
	}

	return &t, nil
}

func (h *httpTracker) BatchSize() int {
	return 50
}

func (h *httpTracker) String() string {
	return h.u.String()
}

func (h *httpTracker) Scrape(hashes []Hash) (map[Hash]Scrape, error) {
	u := *h.u

	if p := strings.LastIndex(u.Path, "/"); p != -1 {
		if strings.HasPrefix(u.Path[p:], "/announce") {
			u.Path = u.Path[0:p] + "/scrape" + u.Path[p+9:]
		} else {
			return nil, InternalError(fmt.Errorf("can't make scrape url from %q", u.String()))
		}
	}

	q := u.Query()
	q.Del("info_hash")
	for _, v := range hashes {
		q.Add("info_hash", v.String())
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, InternalError(err)
	}

	res, err := h.c.Do(req)
	if err != nil {
		return nil, InternalError(err)
	}
	defer res.Body.Close()

	var v httpTrackerScrapeResponse
	if err := bencode.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, InternalError(err)
	}

	if v.FailureReason != "" {
		return nil, TrackerError(errors.New(v.FailureReason))
	}

	r := make(map[Hash]Scrape)

	for _, h := range hashes {
		e, ok := v.Files[h.String()]
		if !ok {
			continue
		}

		r[h] = Scrape{
			Incomplete: e.Incomplete,
			Complete:   e.Complete,
			Downloaded: e.Downloaded,
		}
	}

	return r, nil
}
