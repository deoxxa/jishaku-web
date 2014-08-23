package elasticsearch

import (
	"encoding/json"
	"errors"
	"net/http"

	"fknsrs.biz/p/jishaku-web"
	"github.com/olivere/elastic"
)

type Config struct {
	Hosts []string
	Index string
}

type Store struct {
	client *elastic.Client
	index  string
}

func NewStore(c interface{}) (web.Store, error) {
	config, ok := c.(Config)
	if !ok {
		return nil, errors.New("invalid config type")
	}

	client, err := elastic.NewClient(&http.Client{}, config.Hosts...)
	if err != nil {
		return nil, err
	}

	s := &Store{
		client: client,
		index:  config.Index,
	}

	return s, nil
}

func (s *Store) Add(torrent *web.Torrent) error {
	if _, err := s.client.Index().Index(s.index).Type("torrent").Id(torrent.Hash).BodyJson(torrent).Do(); err != nil {
		return err
	} else {
		return nil
	}
}

func (s *Store) Get(id string) (*web.Torrent, error) {
	res, err := s.client.Get().Index(s.index).Type("torrent").Id(id).Do()
	if err != nil {
		return nil, err
	}

	if !res.Found {
		return nil, nil
	}

	var t web.Torrent
	if err := json.Unmarshal(*res.Source, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *Store) Count(query string) (int, error) {
	var q elastic.Query
	if query != "" {
		q = elastic.NewQueryStringQuery(query)
	} else {
		q = elastic.NewMatchAllQuery()
	}

	if r, err := s.client.Count(s.index).Type("torrent").Query(q).Do(); err != nil {
		return -1, err
	} else {
		return int(r), nil
	}
}

func (s *Store) Search(query string, offset int, count int) ([]*web.Torrent, error) {
	var q elastic.Query
	if query != "" {
		q = elastic.NewQueryStringQuery(query)
	} else {
		q = elastic.NewMatchAllQuery()
	}

	var l []*web.Torrent

	if r, err := s.client.Search(s.index).Type("torrent").Query(q).From(offset).Size(count).Do(); err != nil {
		return nil, err
	} else {
		l = make([]*web.Torrent, len(r.Hits.Hits))

		for i, d := range r.Hits.Hits {
			var t web.Torrent
			if err := json.Unmarshal(*d.Source, &t); err != nil {
				return nil, err
			}
			l[i] = &t
		}
	}

	return l, nil
}
