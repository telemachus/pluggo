package git

import (
	"bytes"
	"errors"
	"os"
	"strings"
)

// ErrDetachedHead indicates that a repository is in detached HEAD state
var ErrDetachedHead = errors.New("repository is in detached HEAD state")

// BranchName returns the name of the current branch.
func BranchName(headFile string) (string, error) {
	headRef, err := branchRef(headFile)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(headRef, "refs/heads/"), nil
}

// BranchRef follows the head file to find a repo's current branch.
func branchRef(headFile string) (string, error) {
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "", err
	}

	if !bytes.HasPrefix(data, []byte("ref: ")) {
		return "", ErrDetachedHead
	}

	data = bytes.TrimPrefix(data, []byte("ref: "))
	data = trimLineEnd(data)

	return string(data), nil
}
