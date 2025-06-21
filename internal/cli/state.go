package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/telemachus/pluggo/internal/git"
)

// PluginState represents a plugin installed locally.
type PluginState struct {
	Name      string
	Directory string
	URL       string
	Branch    string
	Hash      git.Digest
}

func (cmd *cmdEnv) makeStateMap() map[string]*PluginState {
	// We cannot know how many plugins the user works with, but we can
	// assume that they work with *some* plugins. Twenty seems like
	// a reasonable initial allocation.
	statesByName := make(map[string]*PluginState, 20)

	for _, dir := range []string{"start", "opt"} {
		states := cmd.scanPackDir(dir)
		for pluginName, state := range states {
			statesByName[pluginName] = state
		}
	}

	return statesByName
}

func (cmd *cmdEnv) scanPackDir(dir string) map[string]*PluginState {
	baseDir := filepath.Join(cmd.dataDir, dir)
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		action := fmt.Sprintf("skipping %q", baseDir)
		cmd.reportWarning(action, "cannot read directory", err)
		return nil
	}

	// Ignore anything that is not a git repository.
	entries = slices.DeleteFunc(entries, func(entry os.DirEntry) bool {
		return !git.IsRepo(filepath.Join(baseDir, entry.Name()))
	})

	states := make(map[string]*PluginState, len(entries))

	for _, entry := range entries {
		pluginName := entry.Name()
		if state := cmd.createState(baseDir, pluginName); state != nil {
			states[pluginName] = state
		}
	}

	return states
}

func (cmd *cmdEnv) createState(baseDir, pluginName string) *PluginState {
	pluginDir := filepath.Join(baseDir, pluginName)

	url, err := cmd.getPluginURL(pluginDir, pluginName)
	if err != nil {
		return nil
	}

	info, err := git.GetBranchInfo(pluginDir)
	if err != nil {
		action := fmt.Sprintf("skipping %q", pluginName)
		cmd.reportWarning(action, "cannot determine repo state", err)
		return nil
	}

	return &PluginState{
		Name:      pluginName,
		Directory: pluginDir,
		URL:       url,
		Branch:    info.Branch,
		Hash:      info.Hash,
	}
}

func (cmd *cmdEnv) getPluginURL(pluginDir, pluginName string) (string, error) {
	url, err := git.URL(pluginDir)
	if err != nil {
		action := fmt.Sprintf("skipping %q", pluginName)
		cmd.reportWarning(action, "cannot determine repo URL", err)
		return "", err
	}
	return url, nil
}
