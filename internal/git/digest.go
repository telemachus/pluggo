// Package git represents and manipulates git commands and objects.
package git

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// Digest represents a SHA-1 digest as a []byte.
type Digest []byte

// HeadDigest follows .git/HEAD to get the SHA digest of a repository's current branch.
func HeadDigest(dir string) (Digest, error) {
	head := filepath.Join(dir, ".git", "HEAD")
	br, err := BranchRef(head)
	if err != nil {
		return nil, err
	}

	digest, err := digestFrom(filepath.Join(dir, ".git", br))
	if err != nil {
		return nil, err
	}

	return digest, nil
}

// HeadDigestString returns a hex encoded string version of a HeadDigest.
func HeadDigestString(dir string) (string, error) {
	digest, err := HeadDigest(dir)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", digest), nil
}

// Equals checks whether one Digest is identical to another.
func (d Digest) Equals(other Digest) bool {
	return bytes.Equal(d, other)
}

func digestFrom(branchRef string) (Digest, error) {
	data, err := os.ReadFile(branchRef)
	if err != nil {
		return Digest(nil), err
	}

	return Digest(data), nil
}

func trimLineEnds(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte("\n"), []byte(""))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte(""))

	return data
}
