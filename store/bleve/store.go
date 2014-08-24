package bleve

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/user"
	"path"

	"fknsrs.biz/p/jishaku-web"
	"github.com/boltdb/bolt"
	"github.com/couchbaselabs/bleve"
	"github.com/couchbaselabs/bleve/analysis"
	"github.com/couchbaselabs/bleve/analysis/token_filters/lower_case_filter"
	"github.com/couchbaselabs/bleve/analysis/token_filters/ngram_filter"
	"github.com/couchbaselabs/bleve/analysis/tokenizers/single_token"
	"github.com/couchbaselabs/bleve/registry"
)

type Config struct {
	Location string
}

type Store struct {
	location string
	docs     *bolt.DB
	index    bleve.Index
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

	docs, err := bolt.Open(path.Join(location, "docs"), 0600, nil)
	if err != nil {
		return nil, err
	}

	err = docs.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("docs"))
		return err
	})
	if err != nil {
		return nil, err
	}

	registry.RegisterAnalyzer("custom", func(config map[string]interface{}, cache *registry.Cache) (*analysis.Analyzer, error) {
		keywordTokenizer, err := cache.TokenizerNamed(single_token.Name)
		if err != nil {
			return nil, err
		}

		toLowerFilter, err := cache.TokenFilterNamed(lower_case_filter.Name)
		if err != nil {
			return nil, err
		}

		ngramFilter, err := cache.TokenFilterNamed(ngram_filter.Name)
		if err != nil {
			return nil, err
		}

		rv := analysis.Analyzer{
			Tokenizer: keywordTokenizer,
			TokenFilters: []analysis.TokenFilter{
				toLowerFilter,
				ngramFilter,
			},
		}

		return &rv, nil
	})

	index, err := bleve.Open(path.Join(location, "index"))
	if err != nil {
		if err.Error() != "cannot open index, path does not exist" {
			return nil, err
		}

		m := bleve.NewIndexMapping()
		m.SetDefaultAnalyzer("custom")

		index, err = bleve.New(path.Join(location, "index"), m)
		if err != nil {
			log.Printf("%#v", err)
			return nil, err
		}
	}

	s := &Store{
		location: location,
		docs:     docs,
		index:    index,
	}

	return s, nil
}

func (s *Store) Add(torrent *web.Torrent) error {
	data, err := json.Marshal(torrent)
	if err != nil {
		return err
	}

	err = s.docs.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte("docs")).Put([]byte(torrent.Hash), data)
	})
	if err != nil {
		return err
	}

	if err := s.index.Index(torrent.Hash, torrent); err != nil {
		return err
	}

	return nil
}

func (s *Store) Get(id string) (*web.Torrent, error) {
	var data []byte

	err := s.docs.View(func(tx *bolt.Tx) error {
		data = tx.Bucket([]byte("docs")).Get([]byte(id))

		return nil
	})
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
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

		err := s.docs.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("docs"))

			for i, d := range r.Hits {
				data := b.Get([]byte(d.ID))

				if data == nil {
					return errors.New("couldn't get document for search result")
				}

				var t web.Torrent
				if err := json.Unmarshal(data, &t); err != nil {
					return err
				}

				l[i] = &t
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return l, nil
}
