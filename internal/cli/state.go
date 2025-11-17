package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"

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

	for _, baseDir := range []string{cmd.startDir, cmd.optDir} {
		states := cmd.scanPackDir(baseDir)
		for pluginName, state := range states {
			statesByName[pluginName] = state
		}
	}

	return statesByName
}

func (cmd *cmdEnv) scanPackDir(baseDir string) map[string]*PluginState {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		cmd.warnf("%s: skipping %s: cannot read directory: %s", cmd.name, strconv.Quote(baseDir), err)
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
		cmd.warnf("%s: skipping %s: cannot determine repo state: %s", cmd.name, strconv.Quote(pluginName), err)
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
		cmd.warnf("%s: skipping %s: cannot determine repo URL: %s", cmd.name, strconv.Quote(pluginName), err)
		return "", err
	}
	return url, nil
}
