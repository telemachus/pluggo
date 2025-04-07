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
