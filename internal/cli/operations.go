package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (cmd *cmdEnv) install(pSpec PluginSpec) error {
	dir := cmd.pluginPath(pSpec)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	//nolint:gosec // pSpec.Branch and pSpec.URL come from user's own config file
	gitCmd := exec.CommandContext(ctx, "git", "clone", "--filter=blob:none", "-b", pSpec.Branch, pSpec.URL, dir)
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func (cmd *cmdEnv) reinstall(dir string, pSpec PluginSpec) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	return cmd.install(pSpec)
}

func (cmd *cmdEnv) update(pState *PluginState, pSpec PluginSpec) result {
	// Move between start/ and opt/ subdirectories as needed; return early
	// if the move fails.
	res, err := cmd.ensureSubDir(pState, pSpec)
	if err != nil {
		return res
	}

	// Return early if we only needed to move.
	if pSpec.Pinned {
		return res
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	//nolint:gosec // pState.Directory comes from user's own config file
	updateCmd := exec.CommandContext(ctx, "git", "-C", pState.Directory, "pull", "--recurse-submodules")
	if err := updateCmd.Run(); err != nil {
		res.opResult.set(opError)
	}

	return res
}

func (cmd *cmdEnv) ensureSubDir(pState *PluginState, pSpec PluginSpec) (result, error) {
	res := result{plugin: pSpec.Name}

	if pSpec.Pinned {
		res.opResult.set(opPinned)
	}

	pSpecPath := cmd.pluginPath(pSpec)

	// Return early if the directory is already in the right place.
	if pSpecPath == pState.Directory {
		return res, nil
	}

	if err := os.Rename(pState.Directory, pSpecPath); err != nil {
		res.opResult.set(opError)

		return res, err
	}

	res.opResult.set(opMoved)
	res.toOpt = pSpec.Opt

	pState.Directory = pSpecPath

	return res, nil
}
