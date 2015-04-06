//go:generate rice embed-go
package main

import (
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	kp             = kingpin.New("jishaku", "Jishaku Toshokan")
	databaseSocket = kp.Flag("database_socket", "Connect to postgres using this socket.").Short('s').String()
	databaseName   = kp.Flag("database_name", "Use this database.").Short('d').Default("jishaku").String()
	webCommand     = kp.Command("web", "Run the web server.")
	addr           = webCommand.Flag("addr", "Listen on this address.").Short('a').Default(":3000").String()
)

func main() {
	switch kingpin.MustParse(kp.Parse(os.Args[1:])) {
	case webCommand.FullCommand():
		webCommandFunction(*databaseSocket, *databaseName, *addr)
	default:
		kp.Usage(os.Stderr)
		os.Exit(1)
	}
}
