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
	databaseName   = kingpin.Flag("database_name", "Use this database.").Short('d').Default("jishaku").String()
	databaseSocket = kingpin.Flag("database_socket", "Connect to postgres using this socket.").Short('s').String()
)

func main() {
	kingpin.Parse()

	initialiseApp()

	dbConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     *databaseSocket,
			Database: *databaseName,
		},
		MaxConnections: 4,
	}

	db, err := pgx.NewConnPool(dbConfig)
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
