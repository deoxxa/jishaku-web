package bleve

import (
	"encoding/json"
	"errors"
	"os"
	"os/user"
	"path"

	"fknsrs.biz/p/jishaku-web"
	"github.com/couchbaselabs/bleve"
	"github.com/couchbaselabs/bleve/analysis/token_filters/ngram_filter"
	"github.com/jmhodges/levigo"
)

type Config struct {
	Location string
}

type Store struct {
	docs  *levigo.DB
	index bleve.Index
	ro    *levigo.ReadOptions
	wo    *levigo.WriteOptions
}

func NewStore(c interface{}) (web.Store, error) {
	config, ok := c.(Config)
	if !ok {
		return nil, errors.New("invalid config type")
	}

	location := config.Location
	if location[0:1] == "~" {
		currentUser, err := user.Current()
		if err != nil {
			return nil, err
		}

		location = path.Join(currentUser.HomeDir, location[1:])
	}

	if err := os.MkdirAll(location, 0755); err != nil {
		return nil, err
	}

	opts := levigo.NewOptions()
	opts.SetCache(levigo.NewLRUCache(3 << 30))
	opts.SetCreateIfMissing(true)
	docs, err := levigo.Open(path.Join(location, "docs"), opts)
	if err != nil {
		return nil, err
	}

	index, err := bleve.Open(path.Join(location, "index"))
	if err != nil {
		if err.Error() != "cannot open index, path does not exist" {
			return nil, err
		}

		m := bleve.NewIndexMapping()
		m.DefaultAnalyzer = "simple"
		a := m.AnalyzerNamed("simple")
		a.TokenFilters = append(a.TokenFilters, ngram_filter.NewNgramFilter(2, 30))

		index, err = bleve.New(path.Join(location, "index"), m)
		if err != nil {
			return nil, err
		}
	}

	s := &Store{
		docs:  docs,
		index: index,
		ro:    levigo.NewReadOptions(),
		wo:    levigo.NewWriteOptions(),
	}

	return s, nil
}

func (s *Store) Add(torrent *web.Torrent) error {
	data, err := json.Marshal(torrent)
	if err != nil {
		return err
	}

	if err := s.docs.Put(s.wo, []byte(torrent.Hash), data); err != nil {
		return err
	}

	if err := s.index.Index(torrent.Hash, torrent); err != nil {
		return err
	}

	return nil
}

func (s *Store) Get(id string) (*web.Torrent, error) {
	data, err := s.docs.Get(s.ro, []byte(id))
	if err != nil {
		return nil, err
	}

	var t web.Torrent
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *Store) Count(query string) (int, error) {
	var q bleve.Query
	if query != "" {
		q = bleve.NewSyntaxQuery(query)
	} else {
		q = bleve.NewMatchAllQuery()
	}

	search := bleve.NewSearchRequest(q)
	if res, err := s.index.Search(search); err != nil {
		return -1, err
	} else {
		return int(res.Total), nil
	}
}

func (s *Store) Search(query string, offset int, count int) ([]*web.Torrent, error) {
	var q bleve.Query
	if query != "" {
		q = bleve.NewSyntaxQuery(query)
	} else {
		q = bleve.NewMatchAllQuery()
	}

	var l []*web.Torrent

	search := bleve.NewSearchRequest(q)
	search.From = offset
	search.Size = count
	if r, err := s.index.Search(search); err != nil {
		return nil, err
	} else {
		l = make([]*web.Torrent, len(r.Hits))

		for i, d := range r.Hits {
			data, err := s.docs.Get(s.ro, []byte(d.ID))
			if err != nil {
				return nil, err
			}

			if data == nil {
				return nil, errors.New("couldn't get document for search result")
			}

			var t web.Torrent
			if err := json.Unmarshal(data, &t); err != nil {
				return nil, err
			}

			l[i] = &t
		}
	}

	return l, nil
}
