package cli

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func makePlugins() []pluginSpec {
	return []pluginSpec{
		{URL: "https://github.com/foo/foo.git", Name: "foo.git", Branch: "foo"},
		{URL: "https://github.com/bar/bar.git", Name: "bar.git", Branch: "master"},
		{URL: "https://example.com/buzz/fizz.git", Name: "random.git", Branch: "main"},
	}
}

func fakeCmdEnv(confFile string) *cmdEnv {
	return &cmdEnv{
		name:     "test",
		confFile: confFile,
		homeDir:  "/tmp/test-home",
	}
}

func TestGetPluginsSuccess(t *testing.T) {
	t.Parallel()

	expected := makePlugins()
	confFile := "testdata/plugins.json"
	cmd := fakeCmdEnv(confFile)

	actual, err := cmd.plugins()
	if err != nil {
		t.Fatalf("test cannot finish since cmd.plugins() failed: %v", err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("cmd.plugins(%q) failure (-want +got)\n%s", confFile, diff)
	}
}

func TestGetPluginsFailure(t *testing.T) {
	t.Parallel()

	cmd := fakeCmdEnv("testdata/nope.json")
	_, err := cmd.plugins()

	if err == nil {
		t.Error("expected error")
	}
}

func TestPluginChecks(t *testing.T) {
	t.Parallel()

	confFile := "testdata/plugin-checks.json"
	cmd := fakeCmdEnv(confFile)

	plugins, err := cmd.plugins()
	if err != nil {
		t.Fatalf("test cannot finish since cmd.plugins() failed: %v", err)
	}

	// Only the last plugin should be valid (has all required fields).
	// The first has no name, second has no URL, third has no branch.
	if len(plugins) != 1 {
		t.Errorf("cmd.plugins(%q) expected len(plugins) = 1; actual: %d", confFile, len(plugins))
	}
}

func TestMissingDataDirError(t *testing.T) {
	t.Parallel()

	cmd := fakeCmdEnv("testdata/no-datadir.json")
	_, err := cmd.plugins()

	if err == nil {
		t.Error("expected error for missing dataDir")
	}
}
