// Package git represents and manipulates git commands and objects.
package git

import (
	"bytes"
	"unicode"
)

// Digest represents a SHA-1 digest as a []byte.
type Digest []byte

// HeadDigest returns the SHA digest of a repository's current branch. It
// returns an error if the repository is in detached HEAD state or if the
// digest cannot be determined.
func HeadDigest(repoDir string) (Digest, error) {
	return HeadDigestWithReader(repoDir, defaultFileReader)
}

// HeadDigestWithReader is like HeadDigest but accepts a custom file reader for
// testing or other specialized situations.
func HeadDigestWithReader(repoDir string, fr FileReader) (Digest, error) {
	info, err := GetBranchInfoWithReader(repoDir, fr)
	if err != nil {
		return nil, err
	}

	return info.Hash, nil
}

// Equals checks whether one Digest is identical to another.
func (d Digest) Equals(other Digest) bool {
	return bytes.Equal(d, other)
}

// String returns a string representation of a Digest.
func (d Digest) String() string {
	return string(d)
}

func digestFrom(branchRef string, fr FileReader) (Digest, error) {
	data, err := fr.ReadFile(branchRef)
	if err != nil {
		return nil, err
	}

	return Digest(trimLineEnd(data)), nil
}

func trimLineEnd(data []byte) []byte {
	return bytes.TrimRightFunc(data, unicode.IsSpace)
}
