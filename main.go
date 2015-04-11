//go:generate rice embed-go
package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	kp             = kingpin.New("jishaku", "Jishaku Toshokan")
	databaseDSN    = kp.Flag("database_dsn", "Connect to postgres using this information.").Default("dbname=jishaku").String()
	debug          = kp.Flag("debug", "Enable debug logging.").Bool()
	webCommand     = kp.Command("web", "Run the web server.")
	addr           = webCommand.Flag("addr", "Listen on this address.").Default(":3000").String()
	scraperCommand = kp.Command("scraper", "Run the scraper daemon.")
)

func main() {
	switch kingpin.MustParse(kp.Parse(os.Args[1:])) {
	case webCommand.FullCommand():
		webCommandFunction(*databaseDSN, *debug, *addr)
	case scraperCommand.FullCommand():
		scraperCommandFunction(*databaseDSN, *debug)
	default:
		kp.Usage(os.Stderr)
		os.Exit(1)
	}
}
