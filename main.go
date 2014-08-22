package main

import (
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.Lshortfile)

	s := &http.Server{
		Addr: ":3000",
		Handler: newApp(appConfig{
			domain:  "localhost",
			esHosts: []string{"http://127.0.0.1:9200"},
			esIndex: "jishaku",
		}),
	}

	log.Fatal(s.ListenAndServe())
}
