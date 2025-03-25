// Package git represents and manipulates git commands and objects.
package git

import (
	"bytes"
	"os"
)

// FetchHead represents a git FETCH_HEAD as a []byte.
type FetchHead []byte

// NewFetchHead reads a git FETCH_HEAD and returns a FetchHead and possible error.
func NewFetchHead(f string) (FetchHead, error) {
	fh, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}

	return FetchHead(fh), nil
}

// Equals checks whether one FetchHead is identical to another.
func (fh FetchHead) Equals(fhOther FetchHead) bool {
	return bytes.Equal(fh, fhOther)
}
