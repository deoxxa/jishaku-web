package main

import (
	"log"
	"net/http"

	"fknsrs.biz/jishaku/web"
	"fknsrs.biz/jishaku/web/store/elasticsearch"
)

func main() {
	log.SetFlags(log.Lshortfile)

	app, err := web.NewApp(web.AppConfig{
		Domain: "localhost",
		Store: web.StoreFactory{
			Constructor: elasticsearch.NewStore,
			Config: elasticsearch.Config{
				Hosts: []string{
					"http://127.0.0.1:9200/",
				},
				Index: "jishaku",
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	s := &http.Server{
		Addr:    ":3000",
		Handler: app,
	}

	log.Fatal(s.ListenAndServe())
}
