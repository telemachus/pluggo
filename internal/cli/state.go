package cli

import (
	"context"
	"os"
	"path/filepath"
	"slices"
)

// makeStateMap scans the plugin directories and returns a map of installed plugins.
func (cmd *cmdEnv) makeStateMap(ctx context.Context) map[string]*pluginState {
	statesByName := make(map[string]*pluginState, 20)

	for _, baseDir := range []string{cmd.startDir, cmd.optDir} {
		states := cmd.scanPackDir(ctx, baseDir)
		for pluginName, state := range states {
			if _, exists := statesByName[pluginName]; exists {
				cmd.warnf("%s: duplicate plugin %q found in both start/ and opt/ directories", cmd.name, pluginName)
				continue
			}

			statesByName[pluginName] = state
		}
	}

	return statesByName
}

// scanPackDir scans a directory for git repositories representing plugins.
func (cmd *cmdEnv) scanPackDir(ctx context.Context, baseDir string) map[string]*pluginState {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		cmd.warnf("%s: skipping %q: cannot read directory: %s", cmd.name, baseDir, err)
		return nil
	}

	// Filter out non-git repositories.
	entries = slices.DeleteFunc(entries, func(entry os.DirEntry) bool {
		return !isRepo(filepath.Join(baseDir, entry.Name()))
	})

	states := make(map[string]*pluginState, len(entries))

	type result struct {
		state *pluginState
		name  string
	}
	results := make(chan result, len(entries))

	for _, entry := range entries {
		go func() {
			pluginName := entry.Name()
			state := cmd.createState(ctx, baseDir, pluginName)
			results <- result{name: pluginName, state: state}
		}()
	}

	for range entries {
		r := <-results
		if r.state != nil {
			states[r.name] = r.state
		}
	}

	return states
}

// createState determines the state of a single plugin.
func (cmd *cmdEnv) createState(ctx context.Context, baseDir, pluginName string) *pluginState {
	pluginDir := filepath.Join(baseDir, pluginName)

	url, err := repoURL(ctx, pluginDir)
	if err != nil {
		cmd.warnf("%s: skipping %q: cannot determine repo URL: %s", cmd.name, pluginName, err)
		return nil
	}

	info, err := getBranchInfo(ctx, pluginDir)
	if err != nil {
		cmd.warnf("%s: skipping %q: cannot determine repo state: %s", cmd.name, pluginName, err)
		return nil
	}

	return &pluginState{
		name:      pluginName,
		directory: pluginDir,
		url:       url,
		branch:    info.branch,
		hash:      info.hash,
	}
}
