package main

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cheggaaa/pb"
	"github.com/jackc/pgx"

	"fknsrs.biz/p/jishaku-web/lib/scraper"
)

var (
	SELECT_QUERY = `select "info_hash", "trackers" from "torrents" where "torrents"."last_scrape" is null or "torrents"."last_scrape" < (now() - interval '1 day') order by "torrents"."last_scrape" asc nulls first, "torrents"."first_seen" desc limit 1000`
	INSERT_QUERY = `insert into "scrapes" ("info_hash", "tracker", "time", "success", "downloaded", "complete", "incomplete") values ($1, $2, now(), $3, $4, $5, $6)`
	UPDATE_QUERY = `update "torrents" set "last_scrape" = now() where "info_hash" = $1`
)

func scraperCommandFunction(databaseDSN string, debug bool) {
	connConfig, err := pgx.ParseDSN(databaseDSN)
	if err != nil {
		panic(err)
	}

	dbConfig := pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 4,
	}

	if debug {
		dbConfig.Logger = (*wrappedLogger)(logrus.StandardLogger())
	}

	db, err := pgx.NewConnPool(dbConfig)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	s := scraper.New()
	go s.Run()

	for {
		rows, err := db.Query(SELECT_QUERY)
		if err != nil {
			panic(err)
		}

		var wg1 sync.WaitGroup

		bar := pb.New(0)

		var c int
		for rows.Next() {
			c++

			var infoHash string
			var trackers []string
			if err := rows.Scan(&infoHash, &trackers); err != nil {
				panic(err)
				// continue
			}

			wg1.Add(1)

			go func(infoHash string, trackers []string) {
				defer wg1.Done()

				var wg2 sync.WaitGroup

				h, err := scraper.HashFromString(infoHash)
				if err != nil {
					panic(err)
				}

				for _, t := range trackers {
					bar.Total += 1
					wg2.Add(1)

					go func(t string) {
						defer wg2.Done()
						defer bar.Increment()

						l := logrus.WithFields(logrus.Fields{
							"info_hash": infoHash,
							"tracker":   t,
						})

						res, err := s.Scrape(t, h)
						if err != nil {
							if _, ok := err.(scraper.UnsupportedProtocolError); !ok {
								l.Error(err.Error())
							}

							if _, err := db.Exec(INSERT_QUERY, infoHash, t, false, 0, 0, 0); err != nil {
								panic(err)
							}

							return
						}

						if _, err := db.Exec(INSERT_QUERY, infoHash, t, true, res.Downloaded, res.Complete, res.Incomplete); err != nil {
							panic(err)
						}
					}(t)
				}

				wg2.Wait()

				if _, err := db.Exec(UPDATE_QUERY, infoHash); err != nil {
					panic(err)
				}
			}(infoHash, trackers)
		}

		logrus.WithField("count", c).Info("waiting")

		bar.Start()

		wg1.Wait()

		bar.Finish()

		if c == 0 {
			time.Sleep(time.Minute)
		}
	}
}
