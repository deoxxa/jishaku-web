package main

import (
	"net/http"

	"fknsrs.biz/p/negroni-hackystats"
	"github.com/GeertJohan/go.rice"
	"github.com/codegangsta/negroni"
	"github.com/jackc/pgx"
	"github.com/meatballhat/negroni-logrus"
)

func webCommandFunction(databaseSocket, databaseName, addr string) {
	initialiseTemplates()

	dbConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     databaseSocket,
			Database: databaseName,
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
	s.Use(hackystats.New("/_stats"))
	s.Use(negroni.NewRecovery())
	s.Use(negronilogrus.NewMiddleware())
	s.Use(negroni.NewStatic(rice.MustFindBox("public").HTTPBox()))
	s.UseHandler(a)

	if err := http.ListenAndServe(addr, s); err != nil {
		panic(err)
	}
}
