package cli

import (
	"context"
	"fmt"
)

// reinstall removes and re-clones a plugin repository.
func (cmd *cmdEnv) reinstall(ctx context.Context, pState *pluginState, pSpec pluginSpec) error {
	relPath, err := cmd.relativePluginPath(pState.directory)
	if err != nil {
		return err
	}

	if err := cmd.dataRoot.RemoveAll(relPath); err != nil {
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

	srcRelPath, err := cmd.relativePluginPath(pState.directory)
	if err != nil {
		return "", err
	}

	dstRelPath, err := cmd.relativePluginPath(targetPath)
	if err != nil {
		return "", err
	}

	if err := cmd.dataRoot.Rename(srcRelPath, dstRelPath); err != nil {
		return "", err
	}

	// Update state with new location.
	pState.directory = targetPath

	if pSpec.Opt {
		return "opt", nil
	}

	return "start", nil
}

func (cmd *cmdEnv) update(ctx context.Context, pState *pluginState) error {
	return pull(ctx, pState.directory)
}

// hasConfigChanged checks whether a plugin should be reinstalled.
func (cmd *cmdEnv) hasConfigChanged(pState *pluginState, pSpec pluginSpec) (bool, string) {
	switch {
	case pState.url != pSpec.URL:
		return true, "plugin URL changed"
	case pState.branch != pSpec.Branch:
		return true, fmt.Sprintf("switching from branch %s to %s", pState.branch, pSpec.Branch)
	default:
		return false, ""
	}
}
