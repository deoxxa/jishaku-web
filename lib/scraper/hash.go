package scraper

import (
	"encoding/hex"
	"fmt"
)

type Hash [20]byte

func (h Hash) String() string {
	return string(h[:])
}

func HashFromSlice(b []byte) (Hash, error) {
	var h Hash

	if len(b) != 20 {
		return h, fmt.Errorf("input slice should be 20 bytes; was %d", len(b))
	}

	copy(h[:], b)

	return h, nil
}

func HashFromString(s string) (Hash, error) {
	if b, err := hex.DecodeString(s); err != nil {
		return Hash{}, err
	} else {
		return HashFromSlice(b)
	}
}
