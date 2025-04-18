package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/telemachus/pluggo/internal/git"
)

// PluginSpec represents a plugin in a user's configuration file.
type PluginSpec struct {
	URL    string
	Name   string
	Branch string
	Opt    bool
	Pin    bool
}

// PluginState represents a plugin on disk.
type PluginState struct {
	Name      string
	Directory string
	URL       string
	Branch    string
	Hash      git.Digest
}

// updateResult organizes information about the update of one plugin.
type updateResult struct {
	err        error
	hashBefore git.Digest
	hashAfter  git.Digest
	moved      bool
	pin        bool
	toOpt      bool
}

func (cmd *cmdEnv) sync(pSpecs []PluginSpec) {
	if cmd.noOp() || !cmd.pluginDirExists() {
		return
	}

	statesByName := cmd.makeStateMap()
	specsByName := makeSpecMap(pSpecs)

	cmd.processAll(statesByName, pSpecs)

	unwanted := findUnwanted(statesByName, specsByName)
	cmd.removeAll(unwanted)
}

func (cmd *cmdEnv) pluginDirExists() bool {
	if err := os.MkdirAll(cmd.dataDir, os.ModePerm); err != nil {
		cmd.errCount++
		fmt.Fprintf(os.Stderr, "%s: failed to create plugin directory: %s\n", cmd.name, err)

		return false
	}

	return true
}

func (cmd *cmdEnv) makeStateMap() map[string]*PluginState {
	// We cannot know how many plugins the user works with, but we can
	// assume that they work with *some* plugins. Twenty seems like
	// a reasonable initial allocation.
	statesByName := make(map[string]*PluginState, 20)

	for _, dir := range []string{"start", "opt"} {
		states := cmd.scanPackDir(dir)
		for name, state := range states {
			statesByName[name] = state
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
		cmd.warnCount++
		fmt.Fprintf(os.Stderr, "%s: failed to read plugin directory %s: %s\n", cmd.name, baseDir, err)

		return nil
	}

	entries = slices.DeleteFunc(entries, func(entry os.DirEntry) bool {
		return !entry.IsDir() || !isGitRepo(baseDir, entry.Name())
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

func (cmd *cmdEnv) createState(baseDir, repoName string) *PluginState {
	pluginDir := filepath.Join(baseDir, repoName)

	url, err := git.URL(pluginDir)
	if err != nil {
		cmd.warnCount++
		fmt.Fprintf(os.Stderr, "%s: failed to get URL for plugin %s: %s\n", cmd.name, repoName, err)

		return nil
	}

	branch, err := git.BranchName(filepath.Join(pluginDir, ".git", "HEAD"))
	if err != nil {
		cmd.warnCount++
		fmt.Fprintf(os.Stderr, "%s: failed to get branch for plugin %s: %s\n", cmd.name, repoName, err)

		return nil
	}

	hash, err := git.HeadDigest(pluginDir)
	if err != nil {
		cmd.warnCount++
		fmt.Fprintf(os.Stderr, "%s: failed to get hash for plugin %s: %s\n", cmd.name, repoName, err)

		return nil
	}

	return &PluginState{
		Name:      repoName,
		Directory: pluginDir,
		URL:       url,
		Branch:    branch,
		Hash:      hash,
	}
}

func makeSpecMap(pSpecs []PluginSpec) map[string]PluginSpec {
	specsByName := make(map[string]PluginSpec, len(pSpecs))
	for _, p := range pSpecs {
		specsByName[p.Name] = p
	}

	return specsByName
}

func (cmd *cmdEnv) processAll(statesByName map[string]*PluginState, pSpecs []PluginSpec) {
	ch := make(chan result)
	for _, spec := range pSpecs {
		go cmd.reconcile(statesByName[spec.Name], spec, ch)
	}

	for range pSpecs {
		res := <-ch

		if !cmd.quietWanted {
			res.publish()
		} else if res.isErr {
			res.publishError()
		}
	}
}

func findUnwanted(statesByName map[string]*PluginState, specsByName map[string]PluginSpec) map[string]string {
	unwanted := make(map[string]string, len(statesByName))

	for pluginName, state := range statesByName {
		if _, exists := specsByName[pluginName]; !exists {
			unwanted[pluginName] = state.Directory
		}
	}

	return unwanted
}

func (cmd *cmdEnv) removeAll(unwanted map[string]string) {
	for pluginName, pluginPath := range unwanted {
		if err := os.RemoveAll(pluginPath); err != nil {
			cmd.warnCount++
			fmt.Fprintf(os.Stderr, "%s: failed to remove %s: %s\n", cmd.name, pluginName, err)

			continue
		}

		if !cmd.quietWanted {
			fmt.Printf("%s: removed (not in configuration)\n", pluginName)
		}
	}
}

func (cmd *cmdEnv) reconcile(pState *PluginState, pSpec PluginSpec, ch chan<- result) {
	if pState == nil {
		cmd.manageInstall(pSpec, ch)

		return
	}

	if changed, reason := cmd.hasConfigChanged(pState, pSpec); changed {
		cmd.manageReinstall(pState, pSpec, reason, ch)

		return
	}

	cmd.manageUpdate(pState, pSpec, ch)
}

func (cmd *cmdEnv) hasConfigChanged(pState *PluginState, pSpec PluginSpec) (bool, string) {
	switch {
	case pState.URL != pSpec.URL:
		return true, "repo URL changed"
	case pState.Branch != pSpec.Branch:
		return true, fmt.Sprintf(
			"switching from branch %s to %s",
			pState.Branch,
			pSpec.Branch,
		)
	default:
		return false, ""
	}
}

func (cmd *cmdEnv) manageInstall(pSpec PluginSpec, ch chan<- result) {
	pluginDir := pSpec.fullPath(cmd.dataDir)

	if err := cmd.install(pSpec, pluginDir); err != nil {
		cmd.incrementWarn()
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to install %s: %s", pSpec.Name, err),
		}

		return
	}

	ch <- result{
		isErr: false,
		msg:   fmt.Sprintf("%s: installed", pSpec.Name),
	}
}

func (cmd *cmdEnv) manageReinstall(pState *PluginState, pSpec PluginSpec, reason string, ch chan<- result) {
	if err := cmd.reinstall(pState.Directory, pSpec); err != nil {
		cmd.incrementWarn()
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to reinstall %s: %s", pSpec.Name, err),
		}

		return
	}

	ch <- result{
		isErr: false,
		msg:   fmt.Sprintf("%s: reinstalled (%s)", pSpec.Name, reason),
	}
}

func (cmd *cmdEnv) manageUpdate(pState *PluginState, pSpec PluginSpec, ch chan<- result) {
	upRes := cmd.update(pState, pSpec)
	if upRes.err != nil {
		cmd.incrementWarn()
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to update %s: %s", pSpec.Name, upRes.err),
		}

		return
	}

	hashAfter, err := git.HeadDigest(pState.Directory)
	if err != nil {
		cmd.incrementWarn()
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to get updated hash for %s: %s", pSpec.Name, err),
		}

		return
	}

	upRes.hashBefore = pState.Hash
	upRes.hashAfter = hashAfter

	ch <- result{
		isErr: false,
		msg:   formatUpdateMsg(pSpec.Name, upRes),
	}
}

func (cmd *cmdEnv) install(pSpec PluginSpec, dir string) error {
	if err := os.MkdirAll(filepath.Dir(dir), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	args := pSpec.installArgs(dir)
	gitCmd := exec.Command("git", args...)
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %s", err)
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
	upRes := updateResult{pin: pSpec.Pin}
	pluginDir := pSpec.fullPath(cmd.dataDir)

	if pluginDir != pState.Directory {
		upRes.moved = true
		upRes.toOpt = pSpec.Opt

		if err := os.MkdirAll(filepath.Dir(pluginDir), os.ModePerm); err != nil {
			upRes.err = fmt.Errorf("failed to create parent directory for move: %w", err)
			return upRes
		}

		if err := os.Rename(pState.Directory, pluginDir); err != nil {
			upRes.err = fmt.Errorf("failed to move plugin directory: %w", err)
			return upRes
		}

		pState.Directory = pluginDir
	}

	if pSpec.Pin {
		return upRes
	}

	updateCmd := exec.Command("git", "-C", pState.Directory, "pull", "--recurse-submodules")
	if err := updateCmd.Run(); err != nil {
		upRes.err = fmt.Errorf("git pull failed: %s", err)
	}

	return upRes
}

func formatUpdateMsg(pluginName string, upRes updateResult) string {
	var msg strings.Builder
	msg.WriteString(pluginName)
	msg.WriteString(":")

	switch {
	case upRes.moved && upRes.toOpt:
		msg.WriteString(" moved from start/ to opt/")
	case upRes.moved:
		msg.WriteString(" moved from opt/ to start/")
	}

	if upRes.pin {
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
		msg.WriteString(fmt.Sprintf(" from %s to %s", upRes.hashBefore, upRes.hashAfter))
	} else {
		if !upRes.moved && !upRes.pin {
			msg.WriteString(" already up to date")
		}
	}

	return msg.String()
}

func (pSpec *PluginSpec) fullPath(dataDir string) string {
	switch pSpec.Opt {
	case true:
		return filepath.Join(dataDir, "opt", pSpec.Name)
	default:
		return filepath.Join(dataDir, "start", pSpec.Name)
	}
}

func (pSpec *PluginSpec) installArgs(fullPath string) []string {
	args := make([]string, 0, 6)

	args = append(args, "clone", "--filter=blob:none")
	if pSpec.Branch != "" {
		args = append(args, "-b", pSpec.Branch)
	}
	args = append(args, pSpec.URL, fullPath)

	return args
}

func isGitRepo(baseDir, repoName string) bool {
	gitDir := filepath.Join(baseDir, repoName, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}

	return true
}
