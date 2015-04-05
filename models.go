package main

import (
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/zeebo/bencode"
)

func init() {
	pgx.DefaultTypeFormats["torrent_file"] = pgx.BinaryFormatCode
	pgx.DefaultTypeFormats["_torrent_file"] = pgx.BinaryFormatCode
}

type Torrent struct {
	Comment      string
	CreationDate int64              `bencode:"creation date"`
	RawInfo      bencode.RawMessage `bencode:"info"`
	Info         struct {
		Pieces      []byte
		PieceLength int64 `bencode:"piece length"`
		Name        string
		Length      int64
		Files       []struct {
			Length int64
			Path   string
		}
	} `bencode:"-"`
	Announce     string
	AnnounceList [][]string `bencode:"announce-list"`
}

type Entry struct {
	InfoHash  string          `json:"infoHash"`
	Name      string          `json:"name"`
	Size      int64           `json:"size"`
	FirstSeen time.Time       `json:"firstSeen"`
	Files     TorrentFileList `json:"files"`
	Trackers  []string        `json:"trackers"`
	Locations []string        `json:"locations"`
}

func (e Entry) MagnetURI() template.URL {
	q := url.Values{
		"dn": {e.Name},
		"tr": e.Trackers,
	}

	return template.URL("magnet:?xt=urn:btih:" + e.InfoHash + "&" + q.Encode())
}

type TorrentFile struct {
	Path   string `json:"path"`
	Length int64  `json:"length"`
}

func (t *TorrentFile) Scan(r *pgx.ValueReader) error {
	_ = r.ReadInt32() // length of payload; unused

	n := r.ReadInt32()
	if n != 2 {
		return pgx.SerializationError(fmt.Sprintf("TorrentFile.Scan expected to have two elements"))
	}

	t1 := r.ReadOid()
	if t1 != pgx.TextOid {
		return pgx.SerializationError(fmt.Sprintf("TorrentFile.Scan expected first element to be text"))
	}
	t.Path = r.ReadString(r.ReadInt32())

	t2 := r.ReadOid()
	if t2 != pgx.Int8Oid {
		return pgx.SerializationError(fmt.Sprintf("TorrentFile.Scan expected second element to be int8"))
	}
	if l := r.ReadInt32(); l != 8 {
		return pgx.SerializationError(fmt.Sprintf("TorrentFile.Scan expected second element's length to be 8 (probably?)"))
	}
	t.Length = r.ReadInt64()

	return nil
}

func (t TorrentFile) String() string {
	return fmt.Sprintf(`(%q, %d)`, t.Path, t.Length)
}

type TorrentFileList []TorrentFile

func (t TorrentFileList) String() string {
	s := make([]string, len(t))
	for i, e := range t {
		s[i] = strconv.Quote(e.String())
	}

	return "{" + strings.Join(s, ",") + "}"
}

func (t *TorrentFileList) Scan(r *pgx.ValueReader) error {
	if r.Type().DataTypeName != "_torrent_file" {
		return pgx.SerializationError(fmt.Sprintf("TorrentFileList.Scan cannot decode %s (OID %d)", r.Type().DataTypeName, r.Type().DataType))
	}

	ndim := r.ReadInt32()
	if ndim != 1 {
		return pgx.SerializationError(fmt.Sprintf("TorrentFileList.Scan expected to get 1 dimension, instead got %d", ndim))
	}

	hasNull := r.ReadInt32() == 1
	if hasNull {
		return pgx.SerializationError(fmt.Sprintf("TorrentFileList.Scan didn't expect to have null values"))
	}

	_ = r.ReadOid() // element type; unused

	dims, lbound := make([]int32, ndim), make([]int32, ndim)
	for i := 0; i < int(ndim); i++ {
		dims[i] = r.ReadInt32()
		lbound[i] = r.ReadInt32()
	}

	if dims[0] != lbound[0] {
		return pgx.SerializationError(fmt.Sprintf("TorrentFileList.Scan expected the count and lower bound to be equal"))
	}

	for i := 0; i < int(dims[0]); i++ {
		var f TorrentFile
		if err := f.Scan(r); err != nil {
			return err
		} else {
			*t = append(*t, f)
		}
	}

	return nil
}
