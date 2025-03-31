// Package git represents and manipulates git commands and objects.
package git

import (
	"bytes"
	"os"
	"path/filepath"
)

// Digest represents a SHA-1 digest as a []byte.
type Digest []byte

// HeadDigest follows .git/HEAD to get the SHA digest of a repository's current branch.
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

func branchRef(headFile string) (string, error) {
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("ref: "))
	data = bytes.TrimSuffix(data, []byte{'\n'})

	return string(data), nil
}

func digestFrom(branchRef string) (Digest, error) {
	data, err := os.ReadFile(branchRef)
	if err != nil {
		return Digest(nil), err
	}

	return Digest(data), nil
}
