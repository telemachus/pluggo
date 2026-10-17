package cli

import (
	"os"
	"os/exec"
	"testing"
)

func TestGitMetadata(t *testing.T) {
	t.Parallel()

	const url = "https://example.com/user/plugin.git"
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"remote", "add", "origin", url},
		{"config", "url.git@example.com:.insteadOf", "https://example.com/"},
		{"commit", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL="+os.DevNull,
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	gotURL, err := repoURL(t.Context(), dir)
	if err != nil {
		t.Fatalf("repoURL failed: %v", err)
	}
	if gotURL != url {
		t.Errorf("repoURL = %q; want %q", gotURL, url)
	}

	// The git and filesystem paths must agree.
	viaGit, err := getBranchInfoViaGit(t.Context(), dir)
	if err != nil {
		t.Fatalf("getBranchInfoViaGit failed: %v", err)
	}
	viaFS, err := getBranchInfoViaFilesystem(dir)
	if err != nil {
		t.Fatalf("getBranchInfoViaFilesystem failed: %v", err)
	}
	if viaGit.branch != viaFS.branch || !viaGit.hash.equals(viaFS.hash) {
		t.Errorf("via git = {%s %s}; via filesystem = {%s %s}",
			viaGit.branch, viaGit.hash, viaFS.branch, viaFS.hash)
	}
}
