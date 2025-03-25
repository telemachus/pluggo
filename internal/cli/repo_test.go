package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func makeRepos() []Repo {
	return []Repo{
		{URL: "https://github.com/foo/foo.git", Name: "foo.git"},
		{URL: "https://github.com/bar/bar.git", Name: "bar.git"},
		{URL: "https://example.com/buzz/fizz.git", Name: "random.git"},
	}
}

func fakeCmdEnv(confFile string) *cmdEnv {
	return &cmdEnv{
		name:       "test",
		subCmdName: "testing",
		confFile:   confFile,
		exitVal:    exitSuccess,
	}
}

func TestGetReposSuccess(t *testing.T) {
	expected := makeRepos()
	confFile := "testdata/backups.json"
	cmd := fakeCmdEnv(confFile)
	actual := cmd.repos()

	if cmd.exitVal != exitSuccess {
		t.Fatal("test cannot finish since cmd.repos() failed")
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("cmd.repos(%q) failure (-want +got)\n%s", confFile, diff)
	}
}

func TestGetReposFailure(t *testing.T) {
	cmd := fakeCmdEnv("testdata/nope.json")
	cmd.repos()

	if cmd.exitVal != exitFailure {
		t.Error("cmd.exitVal expected exitFailure; actual exitSuccess")
	}
}

func TestRepoChecks(t *testing.T) {
	confFile := "testdata/repo-checks.json"
	cmd := fakeCmdEnv(confFile)
	actual := cmd.repos()

	if cmd.exitVal != exitSuccess {
		t.Fatal("test cannot finish since cmd.repos() failed")
	}

	if len(actual) != 0 {
		t.Errorf("cmd.repos(%q) expected len(repos) = 0; actual: %d", confFile, len(actual))
	}
}
