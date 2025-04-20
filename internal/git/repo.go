package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// URL returns the URL of a git repository as a string.
func URL(repo string) (string, error) {
	cmd := exec.Command("git", "-C", repo, "ls-remote", "--get-url")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsRepo reports whether a directory appears to be a git repository.
// Note: this is a faster but less careful check than IsWorkTree.
func IsRepo(candidate string) bool {
	info, err := os.Stat(filepath.Join(candidate, ".git"))
	if err != nil {
		return false
	}

	// Most git repos have a .git directory; submodules have a .git file.
	return info.IsDir() || info.Mode().IsRegular()
}

// IsWorkTree reports whether a directory is a git work tree.
// Note: this is a slower but more careful check than IsGitDir.
func IsWorkTree(candidate string) bool {
	cmd := exec.Command("git", "-C", candidate, "rev-parse", "--is-inside-work-tree")

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	isWorkTree := strings.TrimSpace(string(output))

	return isWorkTree == "true"
}
