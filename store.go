package web

type StoreFactory struct {
	Constructor func(config interface{}) (Store, error)
	Config      interface{}
}

func (f *StoreFactory) build() (Store, error) {
	return f.Constructor(f.Config)
}

type Store interface {
	Add(torrent *Torrent) error
	Get(id string) (*Torrent, error)
	Count(query string) (int, error)
	Search(query string, offset int, count int) ([]*Torrent, error)
}
