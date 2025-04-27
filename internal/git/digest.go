// Package git represents and manipulates git commands and objects.
package git

import (
	"bytes"
	"os"
	"path/filepath"
	"unicode"
)

// Digest represents a SHA-1 digest as a []byte.
type Digest []byte

// HeadDigest follows .git/HEAD to get the SHA digest of a repository's current
// branch. HeadDigest returns an error if the repository is in detached HEAD
// state or if the digest cannot be determined.
func HeadDigest(dir string) (Digest, error) {
	head := filepath.Join(dir, ".git", "HEAD")
	br, err := branchRef(head)
	if err != nil {
		return nil, err
	}

	digest, err := digestFrom(filepath.Join(dir, ".git", br))
	if err != nil {
		return nil, err
	}

	return digest, nil
}

// Equals checks whether one Digest is identical to another.
func (d Digest) Equals(other Digest) bool {
	return bytes.Equal(d, other)
}

func (d Digest) String() string {
	return string(d)
}

func digestFrom(branchRef string) (Digest, error) {
	data, err := os.ReadFile(branchRef)
	if err != nil {
		return nil, err
	}

	return Digest(data), nil
}

func trimLineEnd(data []byte) []byte {
	return bytes.TrimRightFunc(data, unicode.IsSpace)
}
