package hackystats

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codegangsta/negroni"
)

type stats struct {
	b time.Time
	e time.Time
	m string
	u string
	s int
	l int
}

type hackyStats struct {
	m sync.RWMutex
	c []*stats
	p string
}

// New creates a new hackystats instance, serving the statistics output at a
// specific path
func New(path string) negroni.Handler {
	return &hackyStats{p: path}
}

func (h *hackyStats) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	nw := w.(negroni.ResponseWriter)

	if r.URL.Path == h.p {
		h.m.RLock()
		for i := len(h.c) - 1; i > 0; i-- {
			v := h.c[i]

			var d time.Duration
			if v.e == (time.Time{}) {
				d = time.Now().Sub(v.b)
			} else {
				d = v.e.Sub(v.b)
			}

			fmt.Fprintf(w, "%s %3d %8d %-12s %-6s %s\n", v.b.Format(time.RFC3339), v.s, v.l, d.String(), v.m, v.u)
		}
		h.m.RUnlock()

		return
	}

	e := &stats{time.Now(), time.Time{}, r.Method, r.URL.String(), 0, 0}

	h.m.Lock()
	h.c = append(h.c, e)
	h.m.Unlock()

	defer func() {
		e.e = time.Now()
		e.s = nw.Status()
		e.l = nw.Size()

		go func() {
			time.Sleep(time.Second * 60)

			h.m.Lock()
			for i, v := range h.c {
				if v == e {
					h.c = append(h.c[0:i], h.c[i+1:]...)
					break
				}
			}
			h.m.Unlock()
		}()
	}()

	next(w, r)
}
