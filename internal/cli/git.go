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

// runGit runs git with args and returns its standard output. Git fails rather
// than prompting for credentials, and it gets a chance to clean up when ctx is
// done. On failure, the error includes the last line of git's error output,
// which usually explains the problem.
func runGit(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = 5 * time.Second

	output, err := cmd.Output()
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		if line := lastLine(exitErr.Stderr); line != "" {
			return output, fmt.Errorf("%s (%w)", line, err)
		}
	}

	return output, err
}

func lastLine(data []byte) string {
	data = bytes.TrimSpace(data)
	if _, after, found := bytes.CutLast(data, []byte("\n")); found {
		return string(bytes.TrimSpace(after))
	}

	return string(data)
}

func repoURL(ctx context.Context, repoDir string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Read the raw URL; ls-remote --get-url would apply insteadOf rewrites.
	output, err := runGit(ctx, "-C", repoDir, "config", "--get", "remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func isRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))

	return err == nil && info.IsDir()
}

func clone(ctx context.Context, url, branch, destDir string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if _, err := runGit(ctx, "clone", "--filter=blob:none", "-b", branch, url, destDir); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func pull(ctx context.Context, repoDir string) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if _, err := runGit(ctx, "-C", repoDir, "pull"); err != nil {
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

	// Get both hash and branch name in one call. Order matters: --abbrev-ref
	// applies to every argument after it, so the hash must come first.
	output, err := runGit(ctx, "-C", repoDir, "rev-parse", "HEAD", "--abbrev-ref", "HEAD")
	if err != nil {
		return info, fmt.Errorf("failed to get branch info: %w", err)
	}

	hash, branch, found := bytes.Cut(bytes.TrimSpace(output), []byte("\n"))
	if !found {
		return info, errors.New("unexpected git output format")
	}

	if string(branch) == "HEAD" {
		return info, errDetachedHead
	}
	info.branch = string(branch)
	info.hash = digest(hash)

	return info, nil
}

// getBranchRef reads the branch reference from .git/HEAD.
func getBranchRef(repoDir string) (string, error) {
	headFile := filepath.Join(repoDir, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "", err
	}

	ref, found := bytes.CutPrefix(data, []byte("ref: "))
	if !found {
		return "", errDetachedHead
	}

	return string(bytes.TrimSpace(ref)), nil
}

func readDigest(filename string) (digest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return digest(bytes.TrimSpace(data)), nil
}
