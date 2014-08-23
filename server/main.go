package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"fknsrs.biz/jishaku/web"
	"fknsrs.biz/jishaku/web/store/bleve"
	"fknsrs.biz/jishaku/web/store/elasticsearch"
)

var (
	flagStorage         = flag.String("storage", "local", "search engine type (local or elasticsearch)")
	flagStorageLocation = flag.String("storage_location", "~/.jishaku-web/data", "path to database (local storage)")
	flagStorageHost     = flag.String("storage_host", "http://127.0.0.1:9200/", "host(s) (elasticsearch)")
	flagStorageIndex    = flag.String("storage_index", "jishaku", "index name (elasticsearch)")
	flagListen          = flag.String("listen", ":3000", "[host]:port to listen on")
)

func main() {
	log.SetFlags(log.Lshortfile)

	flag.Parse()

	storeFactory := web.StoreFactory{}

	switch *flagStorage {
	case "elasticsearch":
		log.Printf("storage is elasticsearch, using index %s on %s", *flagStorageIndex, *flagStorageHost)

		storeFactory.Constructor = elasticsearch.NewStore
		storeFactory.Config = elasticsearch.Config{
			Hosts: strings.Split(*flagStorageHost, ","),
			Index: *flagStorageIndex,
		}
	case "local":
		log.Printf("storage is local at %s", *flagStorageLocation)

		storeFactory.Constructor = bleve.NewStore
		storeFactory.Config = bleve.Config{
			Location: *flagStorageLocation,
		}
	default:
		log.Fatal("unknown storage %s", *flagStorage)
	}

	app, err := web.NewApp(web.AppConfig{
		Store: storeFactory,
	})

	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		Addr:    *flagListen,
		Handler: app,
	}

	log.Printf("listening on %s", *flagListen)

	log.Fatal(s.ListenAndServe())
}
