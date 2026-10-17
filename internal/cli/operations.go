package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// reinstall removes and re-clones a plugin repository.
func (cmd *cmdEnv) reinstall(ctx context.Context, dir string, pSpec pluginSpec) error {
	// Verify dir is directly within the expected plugin directories.
	if parent := filepath.Dir(dir); parent != cmd.startDir && parent != cmd.optDir {
		return fmt.Errorf("refusing to remove directory outside plugin paths: %s", dir)
	}

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	return clone(ctx, pSpec.URL, pSpec.Branch, cmd.pluginPath(pSpec))
}

// move relocates a plugin, returning where the plugin was moved and any error.
func (cmd *cmdEnv) move(pState *pluginState, pSpec pluginSpec) (string, error) {
	targetPath := cmd.pluginPath(pSpec)

	// Return early if no move is needed.
	if targetPath == pState.directory {
		return "", nil
	}

	if err := os.Rename(pState.directory, targetPath); err != nil {
		return "", err
	}

	// Update state with new location.
	pState.directory = targetPath

	if pSpec.Opt {
		return "opt", nil
	}

	return "start", nil
}

// hasConfigChanged checks whether a plugin should be reinstalled.
func hasConfigChanged(pState *pluginState, pSpec pluginSpec) (bool, string) {
	switch {
	case pState.url != pSpec.URL:
		return true, "plugin URL changed"
	case pState.branch != pSpec.Branch:
		return true, fmt.Sprintf("switching from branch %s to %s", pState.branch, pSpec.Branch)
	default:
		return false, ""
	}
}
