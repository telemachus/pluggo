package cli

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// makeStateMap scans the plugin directories and returns a map of installed plugins.
func (cmd *cmdEnv) makeStateMap(ctx context.Context) map[string]*pluginState {
	statesByName := make(map[string]*pluginState, 20)

	for _, dirInfo := range []struct {
		name string
		path string
	}{
		{"start", cmd.startDir},
		{"opt", cmd.optDir},
	} {
		dirRoot, err := cmd.dataRoot.OpenRoot(dirInfo.name)
		if err != nil {
			// Directory doesn't exist yet or can't be opened; skip it
			continue
		}

		states := cmd.scanPackDir(ctx, dirRoot, dirInfo.path)
		if closeErr := dirRoot.Close(); closeErr != nil {
			cmd.warnf("%s: error closing %s root: %s", cmd.name, dirInfo.name, closeErr)
		}

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
func (cmd *cmdEnv) scanPackDir(ctx context.Context, dirRoot *os.Root, baseDir string) map[string]*pluginState {
	entries, err := fs.ReadDir(dirRoot.FS(), ".")
	if err != nil {
		cmd.warnf("%s: skipping %q: cannot read directory: %s", cmd.name, baseDir, err)
		return nil
	}

	// Filter out non-git repositories.
	entries = slices.DeleteFunc(entries, func(entry fs.DirEntry) bool {
		return !isRepo(dirRoot, entry.Name())
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

	relPath, err := cmd.relativePluginPath(pluginDir)
	if err != nil {
		cmd.warnf("%s: skipping %q: %s", cmd.name, pluginName, err)
		return nil
	}

	pluginRoot, err := cmd.dataRoot.OpenRoot(relPath)
	if err != nil {
		cmd.warnf("%s: skipping %q: cannot open plugin root: %s", cmd.name, pluginName, err)
		return nil
	}
	defer func() {
		if closeErr := pluginRoot.Close(); closeErr != nil {
			cmd.warnf("%s: error closing plugin root for %q: %s", cmd.name, pluginName, closeErr)
		}
	}()

	info, err := getBranchInfo(ctx, pluginRoot, pluginDir)
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
