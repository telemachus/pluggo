package cli

import (
	"os"
	"path/filepath"
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
		homeDir:  filepath.Join(os.TempDir(), "test-home"), // Absolute on all platforms.
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

	// Only the fourth plugin should be valid (has all required fields).
	// The first has no name, second has no URL, third has no branch, fifth
	// repeats the fourth's name, and sixth has an unsafe name.
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

func TestUnknownKeyError(t *testing.T) {
	t.Parallel()

	cmd := fakeCmdEnv("testdata/unknown-key.json")
	_, err := cmd.plugins()

	if err == nil {
		t.Error("expected error for unknown config key")
	}
}

func TestDuplicateKeyError(t *testing.T) {
	t.Parallel()

	cmd := fakeCmdEnv("testdata/duplicate-key.json")
	_, err := cmd.plugins()

	if err == nil {
		t.Error("expected error for duplicate config key")
	}
}

func TestRelativeDataDirError(t *testing.T) {
	t.Parallel()

	cmd := fakeCmdEnv("testdata/relative-datadir.json")
	_, err := cmd.plugins()

	if err == nil {
		t.Error("expected error for relative dataDir")
	}
}
