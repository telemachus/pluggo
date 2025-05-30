package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func makePlugins() []PluginSpec {
	return []PluginSpec{
		{URL: "https://github.com/foo/foo.git", Name: "foo.git", Branch: "foo"},
		{URL: "https://github.com/bar/bar.git", Name: "bar.git", Branch: "master"},
		{URL: "https://example.com/buzz/fizz.git", Name: "random.git", Branch: "main"},
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

	actual := cmd.plugins()
	if cmd.errCount > 0 {
		t.Fatal("test cannot finish since cmd.plugins() failed")
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("cmd.plugins(%q) failure (-want +got)\n%s", confFile, diff)
	}
}

func TestGetPluginsFailure(t *testing.T) {
	cmd := fakeCmdEnv("testdata/nope.json")
	cmd.plugins()

	if cmd.errCount == 0 {
		t.Error("cmd.exitVal == 0; expected error")
	}
}

func TestRepoChecks(t *testing.T) {
	confFile := "testdata/plugin-checks.json"
	cmd := fakeCmdEnv(confFile)

	actual := cmd.plugins()
	if cmd.errCount > 0 {
		t.Fatal("test cannot finish since cmd.plugins() failed")
	}

	if len(actual) != 1 {
		t.Errorf(
			"cmd.plugins(%q) expected len(repos) = 1; actual: %d",
			confFile,
			len(actual),
		)
	}
}
