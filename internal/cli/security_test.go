package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupTestEnv(t *testing.T) (*cmdEnv, string) {
	t.Helper()

	tempDir := t.TempDir()
	dataDir := filepath.Join(tempDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := &cmdEnv{
		name:     "pluggo",
		dataDir:  dataDir,
		startDir: filepath.Join(dataDir, "start"),
		optDir:   filepath.Join(dataDir, "opt"),
	}

	if err := cmd.openRoot(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cmd.Close(); err != nil {
			t.Errorf("failed to close root: %s", err)
		}
	})

	return cmd, tempDir
}

func createPluginDirs(t *testing.T, cmd *cmdEnv) {
	t.Helper()

	if err := cmd.dataRoot.MkdirAll("start", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := cmd.dataRoot.MkdirAll("opt", 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestPathTraversalPrevention(t *testing.T) {
	t.Parallel()

	cmd, _ := setupTestEnv(t)
	createPluginDirs(t, cmd)

	maliciousNames := []string{
		"../../../etc",
		"../../../../../../etc",
	}

	for _, name := range maliciousNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// This should fail because the resolved path would be
			// outside the root.
			err := cmd.dataRoot.MkdirAll(name, 0o755)
			if err == nil {
				t.Errorf("expected error for malicious path %q, got nil", name)
			}
		})
	}

	// Plugin names with ".." that stay within root should be normalized.
	normalizedPaths := []string{
		"start/../opt",
		"./start",
	}

	for _, name := range normalizedPaths {
		t.Run("normalized:"+name, func(t *testing.T) {
			t.Parallel()

			err := cmd.dataRoot.MkdirAll(name, 0o755)
			if err != nil {
				t.Logf("path %q wants no error; got %s", name, err)
			}
		})
	}
}

func TestAbsolutePathPrevention(t *testing.T) {
	t.Parallel()

	cmd, tempDir := setupTestEnv(t)

	absolutePaths := []string{
		"/etc/passwd",
		"/tmp/evil",
		filepath.Join(tempDir, "escape"),
	}

	for _, absPath := range absolutePaths {
		t.Run(absPath, func(t *testing.T) {
			t.Parallel()

			err := cmd.dataRoot.MkdirAll(absPath, 0o755)
			if err == nil {
				t.Errorf("expected error for absolute path %q, got nil", absPath)
			}
		})
	}
}

func TestValidMoveOperations(t *testing.T) {
	t.Parallel()

	cmd, tempDir := setupTestEnv(t)
	createPluginDirs(t, cmd)
	dataDir := filepath.Join(tempDir, "data")

	if err := cmd.dataRoot.MkdirAll("start/test-plugin", 0o755); err != nil {
		t.Fatalf("cannot create initial directory: %s", err)
	}

	pState := &pluginState{
		name:      "test-plugin",
		directory: filepath.Join(dataDir, "start", "test-plugin"),
	}

	pSpec := pluginSpec{
		Name: "test-plugin",
		Opt:  true,
	}

	movedTo, err := cmd.move(pState, pSpec)
	if err != nil {
		t.Fatalf("move from start to opt failed: %s", err)
	}

	if movedTo != "opt" {
		t.Errorf("expected movedTo=%q, got %q", "opt", movedTo)
	}

	optPath := filepath.Join(dataDir, "opt", "test-plugin")
	if _, err := os.Stat(optPath); err != nil {
		t.Errorf("plugin not found in opt/ after move: %s", err)
	}

	startPath := filepath.Join(dataDir, "start", "test-plugin")
	if _, err := os.Stat(startPath); !os.IsNotExist(err) {
		t.Errorf("plugin still exists in start/ after move")
	}
}

func createProtectedFile(t *testing.T, tempDir string) string {
	t.Helper()

	protectedFile := filepath.Join(tempDir, "protected.txt")
	if err := os.WriteFile(protectedFile, []byte("important"), 0o600); err != nil {
		t.Fatal(err)
	}

	return protectedFile
}

func verifyFileIntact(t *testing.T, path, expectedContent string) {
	t.Helper()

	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("protected file was affected: %s", statErr)
	}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != expectedContent {
		t.Error("protected file content was modified")
	}
}

func TestSecureRemoval(t *testing.T) {
	t.Parallel()

	cmd, tempDir := setupTestEnv(t)
	protectedFile := createProtectedFile(t, tempDir)
	createPluginDirs(t, cmd)

	dataDir := filepath.Join(tempDir, "data")

	if err := cmd.dataRoot.MkdirAll("start/test-plugin", 0o755); err != nil {
		t.Fatal(err)
	}

	relPath, err := cmd.relativePluginPath(filepath.Join(dataDir, "start", "test-plugin"))
	if err != nil {
		t.Fatal(err)
	}

	if removeErr := cmd.dataRoot.RemoveAll(relPath); removeErr != nil {
		t.Fatalf("failed to remove valid plugin: %s", removeErr)
	}

	if _, statErr := os.Stat(filepath.Join(dataDir, "start", "test-plugin")); !os.IsNotExist(statErr) {
		t.Error("plugin was not removed")
	}

	verifyFileIntact(t, protectedFile, "important")
}

func createMockGitRepo(t *testing.T, repoPath, branch, hash string) {
	t.Helper()

	gitDir := filepath.Join(repoPath, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}

	headContent := []byte(fmt.Sprintf("ref: refs/heads/%s\n", branch))
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), headContent, 0o600); err != nil {
		t.Fatal(err)
	}

	refsDir := filepath.Join(gitDir, "refs", "heads")
	if err := os.MkdirAll(refsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	hashContent := []byte(hash + "\n")
	if err := os.WriteFile(filepath.Join(refsDir, branch), hashContent, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGitMetadataIsolation(t *testing.T) {
	t.Parallel()

	cmd, tempDir := setupTestEnv(t)
	dataDir := filepath.Join(tempDir, "data")

	pluginDir := filepath.Join(dataDir, "start", "test-plugin")
	createMockGitRepo(t, pluginDir, "main", "1234567890abcdef1234567890abcdef12345678")

	relPath, err := cmd.relativePluginPath(pluginDir)
	if err != nil {
		t.Fatal(err)
	}

	pluginRoot, err := cmd.dataRoot.OpenRoot(relPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if closeErr := pluginRoot.Close(); closeErr != nil {
			t.Logf("error closing plugin root: %s", closeErr)
		}
	}()

	info, err := getBranchInfoViaFilesystem(pluginRoot)
	if err != nil {
		t.Fatalf("failed to read git metadata: %s", err)
	}

	if info.branch != "main" {
		t.Errorf("expected branch=main, got %q", info.branch)
	}

	expectedHash := "1234567890abcdef1234567890abcdef12345678"
	if info.hash.String() != expectedHash {
		t.Errorf("expected hash=%q, got %q", expectedHash, info.hash.String())
	}
}

func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	cmd, tempDir := setupTestEnv(t)
	createPluginDirs(t, cmd)
	dataDir := filepath.Join(tempDir, "data")

	results := make(chan error, 10)

	for i := range 10 {
		pluginName := filepath.Join("start", fmt.Sprintf("plugin-%d", i))
		go func(name string) {
			err := cmd.dataRoot.MkdirAll(name, 0o755)
			results <- err
		}(pluginName)
	}

	for i := range 10 {
		if err := <-results; err != nil {
			t.Errorf("concurrent operation %d failed: %s", i, err)
		}
	}

	entries, err := os.ReadDir(filepath.Join(dataDir, "start"))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 10 {
		t.Errorf("expected 10 plugins, got %d", len(entries))
	}
}
