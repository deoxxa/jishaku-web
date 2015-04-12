package scraper

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type (
	UnsupportedProtocolError error
	InternalError            error
	TrackerError             error
)

var (
	ErrUnimplemented = errors.New("unimplemented")
)

type Scrape struct {
	Incomplete int
	Complete   int
	Downloaded int
}

type Backend interface {
	Scrape(hashes []Hash) (map[Hash]Scrape, error)
	BatchSize() int
	String() string
}

type scrapeRequest struct {
	t string
	h Hash
	c chan Scrape
	e chan error
}

type backendJob struct {
	b Backend
	q []scrapeRequest
}

type canceler chan bool

func (c canceler) cancel() { c <- true }

type Scraper struct {
	reqs  chan scrapeRequest
	impls map[string]func(u *url.URL) (Backend, error)
	cache map[string]Backend
	queue map[Backend][]scrapeRequest
	sched map[Backend]canceler
}

func New() Scraper {
	return Scraper{
		reqs: make(chan scrapeRequest),
		impls: map[string]func(u *url.URL) (Backend, error){
			"http":  newHTTPTracker,
			"https": newHTTPTracker,
		},
		cache: make(map[string]Backend),
		queue: make(map[Backend][]scrapeRequest),
		sched: make(map[Backend]canceler),
	}
}

func (s *Scraper) Run() {
	for {
		// fmt.Printf("Run() loop\n")

		if r, ok := <-s.reqs; ok {
			s.processRequest(r)
		} else {
			return
		}
	}
}

func (s *Scraper) Stop() {
	close(s.reqs)
}

func (s *Scraper) processRequest(r scrapeRequest) {
	// fmt.Printf("processRequest()\n")

	if _, ok := s.cache[r.t]; !ok {
		u, err := url.Parse(r.t)
		if err != nil {
			r.e <- err
			return
		}

		f, ok := s.impls[strings.ToLower(u.Scheme)]
		if !ok {
			r.e <- UnsupportedProtocolError(fmt.Errorf("unsupported protocol: %s", u.Scheme))
			return
		}

		b, err := f(u)
		if err != nil {
			r.e <- err
			return
		}

		s.cache[r.t] = b
	}

	b := s.cache[r.t]

	s.queue[b] = append(s.queue[b], r)

	if len(s.queue[b]) < b.BatchSize() {
		s.later(b)
	} else {
		s.now(b)
	}
}

func (s *Scraper) later(b Backend) {
	if c, ok := s.sched[b]; ok {
		c.cancel()
	}

	s.sched[b] = make(canceler, 1)

	go func(c canceler) {
		select {
		case <-c:
		case <-time.After(time.Second * 1):
			// fmt.Printf("timed out: %q (%d)\n", b, len(s.queue[b]))

			s.now(b)
		}
	}(s.sched[b])
}

func (s *Scraper) now(b Backend) {
	q := s.queue[b]
	s.queue[b] = nil
	go s.processJob(backendJob{b, q})
}

func (s *Scraper) processJob(j backendJob) {
	// fmt.Printf("processJob(%q, %d)\n", j.b, len(j.q))

	go func() {
		h := make([]Hash, len(j.q))

		for i, v := range j.q {
			h[i] = v.h
		}

		r, err := j.b.Scrape(h)
		if err != nil {
			for _, v := range j.q {
				v.e <- err
			}
		}

		for _, v := range j.q {
			if e, ok := r[v.h]; ok {
				v.c <- e
			} else {
				v.e <- fmt.Errorf("result not found in response")
			}
		}
	}()
}

func (s *Scraper) Scrape(tracker string, hash Hash) (*Scrape, error) {
	// fmt.Printf("Scrape(%q, %q)\n", tracker, hash.String())

	req := scrapeRequest{
		t: tracker,
		h: hash,
		c: make(chan Scrape, 1),
		e: make(chan error, 1),
	}

	s.reqs <- req

	select {
	case r := <-req.c:
		return &r, nil
	case err := <-req.e:
		return nil, err
	}

	return nil, InternalError(fmt.Errorf("shouldn't get here"))
}
