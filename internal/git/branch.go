package git

import (
	"bytes"
	"os"
	"strings"
)

// BranchRef follows the head file to find a repo's current branch.
func BranchRef(headFile string) (string, error) {
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("ref: "))
	data = trimLineEnds(data)

	return string(data), nil
}

// BranchName returns the name of the current branch.
func BranchName(headFile string) (string, error) {
	headRef, err := BranchRef(headFile)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(headRef, "refs/heads/"), nil
}
