//go:generate rice embed-go
package main

import (
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/alecthomas/kingpin"
	"github.com/codegangsta/negroni"
	"github.com/jackc/pgx"
	"github.com/meatballhat/negroni-logrus"
)

var (
	addr           = kingpin.Flag("addr", "Listen on this address.").Short('a').Default(":3000").String()
	databaseUrl    = kingpin.Flag("database_url", "Connect to this database.").Short('d').Default("postgres://localhost/jishaku").String()
	databaseSocket = kingpin.Flag("database_socket", "Override the database connection config to use this socket.").Short('s').String()
)

func main() {
	kingpin.Parse()

	initialiseApp()

	dbConfig, err := pgx.ParseURI(*databaseUrl)
	if err != nil {
		panic(err)
	}

	if *databaseSocket != "" {
		dbConfig.Host = *databaseSocket
	}

	db, err := pgx.Connect(dbConfig)
	if err != nil {
		panic(err)
	}

	a, err := newApp(db)
	if err != nil {
		panic(err)
	}

	s := negroni.New()
	s.Use(negroni.NewRecovery())
	s.Use(negronilogrus.NewMiddleware())
	s.Use(negroni.NewStatic(rice.MustFindBox("public").HTTPBox()))
	s.UseHandler(a)

	if err := http.ListenAndServe(*addr, s); err != nil {
		panic(err)
	}
}
