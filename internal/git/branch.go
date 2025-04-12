package git

import (
	"bytes"
	"io/fs"
	"os"
	"strings"
)

// Repository represents a Git repository.
type Repository struct {
	filesystem fs.FS
	rootPath   string
}

// NewRepo returns a new Repository that points to an os filesystem.
func NewRepo(rootPath string) *Repository {
	dirFS := os.DirFS(rootPath)

	return &Repository{
		filesystem: dirFS,
		rootPath:   rootPath,
	}
}

// NewRepoWithFS returns a Repository that points to a given filesystem.
func NewRepoWithFS(fileSys fs.FS) *Repository {
	return &Repository{
		filesystem: fileSys,
		rootPath:   "",
	}
}

// BranchRef follows the head file to find a repo's current branch.
func (repo *Repository) BranchRef() (string, error) {
	headPath := ".git/HEAD"
	data, err := fs.ReadFile(repo.filesystem, headPath)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("ref: "))
	data = trimLineEnds(data)

	return string(data), nil
}

// BranchName returns the name of the current branch.
func (repo *Repository) BranchName() (string, error) {
	headRef, err := repo.BranchRef()
	if err != nil {
		return "", err
	}

	// If headRef starts with refs/heads/, we know the current branch.
	if strings.HasPrefix(headRef, "refs/heads/") {
		return strings.TrimPrefix(headRef, "refs/heads/"), nil
	}

	// Otherwise, we are in a detached HEAD state.
	// TODO: this should return an error rather than an empty branch name.
	return "", nil
}
