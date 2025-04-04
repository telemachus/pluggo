package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func makePlugins() []Plugin {
	return []Plugin{
		{URL: "https://github.com/foo/foo.git", Name: "foo.git"},
		{URL: "https://github.com/bar/bar.git", Name: "bar.git"},
		{URL: "https://example.com/buzz/fizz.git", Name: "random.git"},
	}
}

func fakeCmdEnv(confFile string) *cmdEnv {
	return &cmdEnv{
		name:     "test",
		confFile: confFile,
	}
}

func TestGetPluginsSuccess(t *testing.T) {
	expected := makePlugins()
	confFile := "testdata/plugins.json"
	cmd := fakeCmdEnv(confFile)

	actual, err := cmd.plugins()
	if err != nil {
		t.Fatal("test cannot finish since cmd.plugins() failed")
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("cmd.plugins(%q) failure (-want +got)\n%s", confFile, diff)
	}
}

func TestGetPluginsFailure(t *testing.T) {
	cmd := fakeCmdEnv("testdata/nope.json")
	_, err := cmd.plugins()

	if err == nil {
		t.Error("err == nil; expected non-nil error")
	}
}

func TestRepoChecks(t *testing.T) {
	confFile := "testdata/plugin-checks.json"
	cmd := fakeCmdEnv(confFile)

	actual, err := cmd.plugins()
	if err != nil {
		t.Fatal("test cannot finish since cmd.plugins() failed")
	}

	if len(actual) != 0 {
		t.Errorf(
			"cmd.plugins(%q) expected len(repos) = 0; actual: %d",
			confFile,
			len(actual),
		)
	}
}
