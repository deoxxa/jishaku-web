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
			var c int
			if v.e == (time.Time{}) {
				d = time.Now().Sub(v.b)
				c = -1
			} else {
				d = v.e.Sub(v.b)
				c = nw.Status()
			}

			fmt.Fprintf(w, "%s %3d %8d %-12s %-6s %s\n", v.b.Format(time.RFC3339), c, nw.Size(), d.String(), v.m, v.u)
		}
		h.m.RUnlock()

		return
	}

	e := &stats{time.Now(), time.Time{}, r.Method, r.URL.String()}

	h.m.Lock()
	h.c = append(h.c, e)
	h.m.Unlock()

	defer func() {
		e.e = time.Now()

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
