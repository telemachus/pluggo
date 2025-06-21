package git

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
)

// ErrDetachedHead indicates that a repository is in detached HEAD state
var ErrDetachedHead = errors.New("repository is in detached HEAD state")

// BranchInfo contains the branch name and SHA digest of a git repository.
type BranchInfo struct {
	Branch string
	Hash   Digest
}

type branch struct {
	name string
	ref  string
}

// GetBranchInfo returns the branch name and SHA digest of a git repository.
// It returns an error if the branch name or SHA digest cannot be determined.
func GetBranchInfo(repoDir string) (BranchInfo, error) {
	return GetBranchInfoWithReader(repoDir, defaultFileReader)
}

// GetBranchInfoWithReader is like GetBranchInfo but accepts a custom file
// reader for testing or other specialized situations.
func GetBranchInfoWithReader(repoDir string, fr FileReader) (BranchInfo, error) {
	var info BranchInfo

	br, err := getBranch(repoDir, fr)
	if err != nil {
		return info, err
	}
	info.Branch = br.name

	branchRefFile := filepath.Join(repoDir, ".git", br.ref)
	hash, err := digestFrom(branchRefFile, fr)
	if err != nil {
		return info, err
	}
	info.Hash = hash

	return info, nil
}

// BranchName returns the branch name for a git repository. It returns an error
// if the branch name cannot be determined.
func BranchName(repoDir string) (string, error) {
	return BranchNameWithReader(repoDir, defaultFileReader)
}

// BranchNameWithReader is like BranchName but accepts a custom file reader for
// testing or other specialized situations.
func BranchNameWithReader(repoDir string, fr FileReader) (string, error) {
	br, err := getBranch(repoDir, fr)
	if err != nil {
		return "", err
	}

	return br.name, nil
}

func getBranch(repoDir string, fr FileReader) (branch, error) {
	headFile := filepath.Join(repoDir, ".git", "HEAD")

	branchRefPath, err := branchRef(headFile, fr)
	if err != nil {
		return branch{}, err
	}

	name := strings.TrimPrefix(branchRefPath, "refs/heads/")

	return branch{name: name, ref: branchRefPath}, nil
}

func branchRef(headFile string, fr FileReader) (string, error) {
	data, err := fr.ReadFile(headFile)
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
