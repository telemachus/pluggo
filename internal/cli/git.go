package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Git command operations

func repoURL(ctx context.Context, repoDir string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "ls-remote", "--get-url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func isRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}

	// Most git repos have a .git directory; submodules have a .git file.
	return info.IsDir() || info.Mode().IsRegular()
}

func clone(ctx context.Context, url, branch, destDir string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone", "--filter=blob:none", "-b", branch, url, destDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func pull(ctx context.Context, repoDir string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "pull", "--recurse-submodules")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

// Git metadata operations

var errDetachedHead = errors.New("repository is in detached HEAD state")

// digest represents a git commit SHA-1 hash.
type digest []byte

func (d digest) equals(other digest) bool {
	return bytes.Equal(d, other)
}

func (d digest) String() string {
	return string(d)
}

type branchInfo struct {
	branch string
	hash   digest
}

// getBranchInfo returns the branch name and SHA digest of a git repository.
func getBranchInfo(ctx context.Context, repoDir string) (branchInfo, error) {
	// Try the filesystem first since it's far faster and usually works.
	info, err := getBranchInfoViaFilesystem(repoDir)
	if err == nil {
		return info, nil
	}

	// Use git itself for edge cases. This is slower but should (always?) work.
	return getBranchInfoViaGit(ctx, repoDir)
}

func getBranchInfoViaFilesystem(repoDir string) (branchInfo, error) {
	var info branchInfo

	branchRef, err := getBranchRef(repoDir)
	if err != nil {
		return info, err
	}

	info.branch = strings.TrimPrefix(branchRef, "refs/heads/")

	branchRefFile := filepath.Join(repoDir, ".git", branchRef)
	hash, err := readDigest(branchRefFile)
	if err != nil {
		return info, err
	}
	info.hash = hash

	return info, nil
}

func getBranchInfoViaGit(ctx context.Context, repoDir string) (branchInfo, error) {
	var info branchInfo

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get both branch name and hash in one call.
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return info, fmt.Errorf("failed to get branch info: %w", err)
	}

	lines := bytes.Split(bytes.TrimSpace(output), []byte("\n"))
	if len(lines) != 2 {
		return info, errors.New("unexpected git output format")
	}

	branch := string(lines[0])
	if branch == "HEAD" {
		return info, errDetachedHead
	}
	info.branch = branch
	info.hash = digest(lines[1])

	return info, nil
}

// getBranchRef reads the branch reference from .git/HEAD.
func getBranchRef(repoDir string) (string, error) {
	headFile := filepath.Join(repoDir, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "", err
	}

	if !bytes.HasPrefix(data, []byte("ref: ")) {
		return "", errDetachedHead
	}

	data = bytes.TrimPrefix(data, []byte("ref: "))
	data = trimLineEnd(data)

	return string(data), nil
}

func readDigest(filename string) (digest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return digest(trimLineEnd(data)), nil
}

func trimLineEnd(data []byte) []byte {
	data = bytes.TrimSuffix(data, []byte("\r\n"))
	data = bytes.TrimSuffix(data, []byte("\n"))

	return data
}
