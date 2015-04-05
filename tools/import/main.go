package main

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type MongoDate struct {
	Date int64 `json:"$date"`
}

func (m MongoDate) Time() time.Time {
	return time.Unix(m.Date/1000, m.Date%1000)
}

type entryFile struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

type entry struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	Comment   string `json:"comment"`
	Createdby struct {
		Client  string `json:"client"`
		Version string `json:"version"`
	} `json:"createdBy"`
	CreationDate MongoDate   `json:"creationDate"`
	Size         int         `json:"size"`
	Trackers     []string    `json:"trackers"`
	Files        []entryFile `json:"files"`
	Hash         string      `json:"hash"`
	Firstseen    MongoDate   `json:"firstSeen"`
	Locations    []string    `json:"locations"`
}

func formatFile(f entryFile) string {
	return "(" + strconv.Quote(f.Name) + "," + strconv.FormatInt(int64(f.Size), 10) + ")"
}

func formatFiles(f []entryFile) []string {
	r := make([]string, len(f))
	for i, v := range f {
		r[i] = formatFile(v)
	}

	return r
}

func formatList(s []string) string {
	r := make([]string, len(s))
	for i, v := range s {
		r[i] = strconv.Quote(v)
	}

	return "{" + strings.Join(r, ",") + "}"
}

var replacer = strings.NewReplacer(
	"\\", "\\\\",
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t",
	"\x00", "",
)

func q(s string) string {
	return replacer.Replace(s)
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	d := json.NewDecoder(f)

	for {
		var e entry
		if err := d.Decode(&e); err != nil {
			panic(err)
		}

		a := []string{
			q(e.ID),
			q(e.Name),
			q(e.Comment),
			strconv.FormatInt(int64(e.Size), 10),
			q(e.Firstseen.Time().Format(time.RFC3339)),
			q(formatList(formatFiles(e.Files))),
			q(formatList(e.Trackers)),
			q(formatList(e.Locations)),
		}

		os.Stdout.Write([]byte(strings.Join(a, "\t") + "\n"))
	}
}
