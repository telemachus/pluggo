package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/telemachus/pluggo/internal/git"
)

// updateResult represents the update of a single plugin.
type updateResult struct {
	err        error
	hashBefore git.Digest
	hashAfter  git.Digest
	moved      bool
	pinned     bool
	toOpt      bool
}

func (cmd *cmdEnv) install(pSpec PluginSpec, dir string) error {
	//nolint:gosec // pSpec.Branch and pSpec.URL come from user's own config file
	gitCmd := exec.Command("git", "clone", "--filter=blob:none", "-b", pSpec.Branch, pSpec.URL, dir)
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func (cmd *cmdEnv) reinstall(dir string, pSpec PluginSpec) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	return cmd.install(pSpec, pSpec.fullPath(cmd.dataDir))
}

func (cmd *cmdEnv) update(pState *PluginState, pSpec PluginSpec) updateResult {
	// Move between start/ and opt/ subdirectories as needed; return early
	// if the move fails.
	upRes, err := cmd.ensureSubDir(pState, pSpec)
	if err != nil {
		return upRes
	}

	// Return early if we only needed to move.
	if pSpec.Pin {
		return upRes
	}

	//nolint:gosec // pState.Directory comes from user's own config file
	updateCmd := exec.Command("git", "-C", pState.Directory, "pull", "--recurse-submodules")
	if err := updateCmd.Run(); err != nil {
		upRes.err = fmt.Errorf("git pull failed: %w", err)
	}

	return upRes
}

func (cmd *cmdEnv) ensureSubDir(pState *PluginState, pSpec PluginSpec) (updateResult, error) {
	upRes := updateResult{pinned: pSpec.Pin}
	pSpecPath := pSpec.fullPath(cmd.dataDir)

	// Return early if the directory is already in the right place.
	if pSpecPath == pState.Directory {
		return upRes, nil
	}

	upRes.moved = true
	upRes.toOpt = pSpec.Opt

	if err := os.Rename(pState.Directory, pSpecPath); err != nil {
		upRes.err = err
		return upRes, err
	}

	pState.Directory = pSpecPath

	return upRes, nil
}

func formatUpdateMsg(pluginName string, upRes updateResult) string {
	var msg strings.Builder
	msg.Grow(len(pluginName) + 50)

	msg.WriteString(pluginName)
	msg.WriteString(":")

	switch {
	case upRes.moved && upRes.toOpt:
		msg.WriteString(" moved from start/ to opt/")
	case upRes.moved:
		msg.WriteString(" moved from opt/ to start/")
	}

	if upRes.pinned {
		if upRes.moved {
			msg.WriteString(" and pinned (no update attempted)")
		} else {
			msg.WriteString(" pinned (no update attempted)")
		}
		return msg.String()
	}

	if !upRes.hashBefore.Equals(upRes.hashAfter) {
		if upRes.moved {
			msg.WriteString(" and updated")
		} else {
			msg.WriteString(" updated")
		}
		msg.WriteString(fmt.Sprintf(" from %.7s to %.7s", upRes.hashBefore, upRes.hashAfter))
	} else if !upRes.moved && !upRes.pinned {
		msg.WriteString(" already up to date")
	}

	return msg.String()
}
