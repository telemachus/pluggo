package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/telemachus/pluggo/internal/git"
)

// LockFile stores Plugins.
type LockFile struct {
	Plugins map[string]*PluginState `json:"plugins"`
}

// loadLockFile loads the lockfile or creates an empty one if it doesn't exist
func (cmd *cmdEnv) loadLockFile() (*LockFile, error) {
	lockPath := filepath.Join(cmd.homeDir, lockFile)
	data, err := os.ReadFile(lockPath)

	// If file doesn't exist, return an empty lockfile
	if os.IsNotExist(err) {
		return &LockFile{
			Plugins: make(map[string]*PluginState),
		}, nil
	}

	if err != nil {
		return nil, err
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// saveLockFile writes the lockfile to disk
func (cmd *cmdEnv) saveLockFile(lock *LockFile) error {
	lockPath := filepath.Join(cmd.homeDir, lockFile)
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(lockPath, data, 0o644)
}

// reconcile aligns the plugin state with the user's configuration
func (cmd *cmdEnv) reconcile(plugins []Plugin, lock *LockFile) error {
	// Create a map of configured plugins for easier lookup
	configPlugins := make(map[string]Plugin)
	for _, p := range plugins {
		configPlugins[p.Name] = p
	}

	// Create a map of lockfile plugins for easier lookup
	lockPlugins := make(map[string]*PluginState)
	for name, state := range lock.Plugins {
		lockPlugins[name] = state
	}

	// Ensure plugin directory exists
	if err := os.MkdirAll(cmd.dataDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Process each configured plugin
	ch := make(chan result)
	for _, plugin := range plugins {
		go cmd.process(plugin, lockPlugins[plugin.Name], lock, ch)
	}

	// Collect results
	var errs []string
	for range plugins {
		res := <-ch
		if res.isErr {
			errs = append(errs, res.msg)
		}

		if !cmd.quietWanted {
			res.publish()
		} else if res.isErr {
			res.publishError()
		}
	}

	// Remove plugins that are in lockfile but not in config
	for name, state := range lock.Plugins {
		if _, exists := configPlugins[name]; !exists {
			if err := os.RemoveAll(state.Directory); err != nil {
				errs = append(errs, fmt.Sprintf("failed to remove %s: %s", name, err))
				continue
			}

			delete(lock.Plugins, name)

			if !cmd.quietWanted {
				fmt.Printf("%s: removed (not in configuration)\n", name)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to reconcile plugins:\n%s", strings.Join(errs, "\n"))
	}

	// Generate helptags
	if err := cmd.updateHelptags(); err != nil {
		return fmt.Errorf("failed to update helptags: %w", err)
	}

	return nil
}

// process handles a single plugin
func (cmd *cmdEnv) process(plugin Plugin, state *PluginState, lock *LockFile, ch chan<- result) {
	pluginDir := plugin.fullPath(cmd.dataDir)

	// Case 1: Plugin not in lockfile - needs installation
	if state == nil {
		if err := cmd.install(plugin, pluginDir, lock); err != nil {
			ch <- result{
				isErr: true,
				msg:   fmt.Sprintf("failed to install %s: %s", plugin.Name, err),
			}
			return
		}

		ch <- result{
			isErr: false,
			msg:   fmt.Sprintf("%s: installed", plugin.Name),
		}
		return
	}

	// Case 2: Plugin configuration changed (URL or branch)
	// Fixed branch handling: Now checks for any branch mismatch, including when
	// a branch is removed (plugin.Branch is empty but state.Branch is not)
	if state.URL != plugin.URL || state.Branch != plugin.Branch {
		reason := "configuration changed"
		if state.Branch != plugin.Branch {
			if plugin.Branch == "" {
				reason = "branch removed, switching to default branch"
			} else if state.Branch == "" {
				reason = fmt.Sprintf("switching to branch %s", plugin.Branch)
			} else {
				reason = fmt.Sprintf("switching from branch %s to %s", state.Branch, plugin.Branch)
			}
		}

		if err := cmd.reinstall(plugin, state.Directory, lock); err != nil {
			ch <- result{
				isErr: true,
				msg:   fmt.Sprintf("failed to reinstall %s: %s", plugin.Name, err),
			}
			return
		}

		ch <- result{
			isErr: false,
			msg:   fmt.Sprintf("%s: reinstalled (%s)", plugin.Name, reason),
		}
		return
	}

	// Case 3: Plugin exists but needs update
	// First, get the hash before update for comparison
	digestBefore, err := git.HeadDigest(state.Directory)
	if err != nil {
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to get hash for %s: %s", plugin.Name, err),
		}
		return
	}
	hashBefore := fmt.Sprintf("%x", digestBefore)

	// Update the plugin
	if err := cmd.update(plugin, state, lock); err != nil {
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to update %s: %s", plugin.Name, err),
		}
		return
	}

	// Get the hash after update
	updatedState := lock.Plugins[plugin.Name]

	// Compare hashes to determine if anything changed
	if hashBefore != updatedState.Hash {
		ch <- result{
			isErr: false,
			msg: fmt.Sprintf("%s: updated from %s to %s",
				plugin.Name,
				hashBefore[:7],
				updatedState.Hash[:7]),
		}
	} else {
		ch <- result{
			isErr: false,
			msg:   fmt.Sprintf("%s: already up to date", plugin.Name),
		}
	}
}

// install installs a new plugin
func (cmd *cmdEnv) install(plugin Plugin, dir string, lock *LockFile) error {
	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(dir), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Install plugin
	args := plugin.installArgs(dir)
	gitCmd := exec.Command("git", args...)
	if output, err := gitCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone failed: %s", output)
	}

	// Get hash of installed plugin
	digest, err := git.HeadDigest(dir)
	if err != nil {
		return fmt.Errorf("failed to get hash: %w", err)
	}

	// Update lockfile
	lock.Plugins[plugin.Name] = &PluginState{
		Name:      plugin.Name,
		Directory: dir,
		URL:       plugin.URL,
		Branch:    plugin.Branch,
		Hash:      fmt.Sprintf("%x", digest),
	}

	return nil
}

// reinstall reinstalls a plugin that had its configuration changed
func (cmd *cmdEnv) reinstall(plugin Plugin, dir string, lock *LockFile) error {
	// Remove existing directory
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	// Install with new configuration
	return cmd.install(plugin, plugin.fullPath(cmd.dataDir), lock)
}

// update updates an existing plugin
func (cmd *cmdEnv) update(plugin Plugin, state *PluginState, lock *LockFile) error {
	// Update plugin
	updateCmd := exec.Command("git", "-C", state.Directory, "pull", "--recurse-submodules")
	if output, err := updateCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %s", output)
	}

	// Get new hash after update
	digestAfter, err := git.HeadDigest(state.Directory)
	if err != nil {
		return fmt.Errorf("failed to get hash: %w", err)
	}

	hashAfter := fmt.Sprintf("%x", digestAfter)

	// Update lockfile with new hash
	lock.Plugins[plugin.Name] = &PluginState{
		Name:      plugin.Name,
		Directory: state.Directory,
		URL:       plugin.URL,
		Branch:    plugin.Branch,
		Hash:      hashAfter,
	}

	return nil
}

// updateHelptags updates Neovim helptags
func (cmd *cmdEnv) updateHelptags() error {
	nvimCmd := exec.Command("nvim", "--headless", "-c", "helptags ALL | quit")
	output, err := nvimCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nvim helptags failed: %s", output)
	}
	return nil
}
