// Package git represents and manipulates git commands and objects.
package git

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"strings"
)

// Digest represents a SHA-1 digest as a []byte.
type Digest []byte

// HeadDigest follows .git/HEAD to get the SHA digest of a repository's current
// branch.
func (repo *Repository) HeadDigest() (Digest, error) {
	// TODO: add .git to the root path of the filesystem and use
	// a Branch function here.
	headPath := ".git/HEAD"
	headData, err := fs.ReadFile(repo.filesystem, headPath)
	if err != nil {
		return nil, err
	}

	headData = trimLineEnds(headData)
	headStr := string(headData)

	// Check if it's a reference or a direct hash (detached HEAD).
	// TODO: we should never get this far. If this is a detached HEAD,
	// we should have received an error above.
	if strings.HasPrefix(headStr, "ref: ") {
		// It's a reference, follow it
		br, err := repo.BranchRef()
		if err != nil {
			return nil, err
		}

		relPath := filepath.ToSlash(".git/" + br)
		digest, err := repo.digestFrom(relPath)
		if err != nil {
			return nil, err
		}

		return digest, nil
	}

	// It's a detached HEAD, the content is already the hash.
	// TODO: fix this. We don't want to return the digest of a detached HEAD.
	return Digest(headData), nil
}

// HeadDigestString returns the SHA digest as a string.
func (repo *Repository) HeadDigestString() (string, error) {
	digest, err := repo.HeadDigest()
	if err != nil {
		return "", err
	}

	return string(digest), nil
}

// Equals checks whether one Digest is identical to another.
func (d Digest) Equals(other Digest) bool {
	return bytes.Equal(d, other)
}

func (repo *Repository) digestFrom(branchRef string) (Digest, error) {
	data, err := fs.ReadFile(repo.filesystem, branchRef)
	if err != nil {
		return Digest(nil), err
	}

	data = trimLineEnds(data)

	return Digest(data), nil
}

func trimLineEnds(data []byte) []byte {
	data = bytes.ReplaceAll(data, []byte("\n"), []byte(""))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte(""))

	return data
}
