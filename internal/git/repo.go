package git

import (
	"fmt"
	"os/exec"
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

// IsWorkTree reports whether a directory is a git work tree.
func IsWorkTree(candidate string) bool {
	cmd := exec.Command("git", "-C", candidate, "rev-parse", "--is-inside-work-tree")

	output, err := cmd.Output()
	if err != nil {
		return false
	}

	isWorkTree := strings.TrimSpace(string(output))

	return isWorkTree == "true"
}
