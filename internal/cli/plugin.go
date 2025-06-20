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

// PluginSpec represents a plugin specified in a user's configuration file.
type PluginSpec struct {
	URL    string
	Name   string
	Branch string
	Opt    bool
	Pin    bool
}

// PluginState represents a plugin installed locally.
type PluginState struct {
	Name      string
	Directory string
	URL       string
	Branch    string
	Hash      git.Digest
}

// updateResult represents the update of a single plugin.
type updateResult struct {
	err        error
	hashBefore git.Digest
	hashAfter  git.Digest
	moved      bool
	pinned     bool
	toOpt      bool
}

func (cmd *cmdEnv) sync(pSpecs []PluginSpec) {
	if cmd.noOp() || !cmd.ensurePluginDirs() {
		return
	}

	statesByName := cmd.makeStateMap()
	specsByName := makeSpecMap(pSpecs)

	unwanted := findUnwanted(statesByName, specsByName)
	cmd.removeAll(unwanted)

	cmd.reconcileLocal(statesByName, pSpecs)
}

func (cmd *cmdEnv) ensurePluginDirs() bool {
	for _, subDir := range []string{"start", "opt"} {
		wantedDir := filepath.Join(cmd.dataDir, subDir)
		if err := os.MkdirAll(wantedDir, os.ModePerm); err != nil {
			reason := fmt.Sprintf("cannot create directory %q", wantedDir)
			cmd.reportError(reason, err)

			return false
		}
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
		for pluginName, state := range states {
			statesByName[pluginName] = state
		}
	}

	return statesByName
}

func makeSpecMap(pSpecs []PluginSpec) map[string]PluginSpec {
	specsByName := make(map[string]PluginSpec, len(pSpecs))
	for _, p := range pSpecs {
		specsByName[p.Name] = p
	}

	return specsByName
}

func (cmd *cmdEnv) reconcileLocal(statesByName map[string]*PluginState, pSpecs []PluginSpec) {
	ch := make(chan result)
	for _, spec := range pSpecs {
		// cmd.reconcile is safe even when statesByName[spec.Name] is nil.
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
			action := fmt.Sprintf("skipping %q", pluginName)
			cmd.reportWarning(action, "failed to remove plugin", err)

			continue
		}

		if !cmd.quietWanted {
			fmt.Printf("%s: removed (not in configuration)\n", pluginName)
		}
	}
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

	branch, err := cmd.getPluginBranch(pluginDir, pluginName)
	if err != nil {
		return nil
	}

	hash, err := cmd.getPluginHash(pluginDir, pluginName)
	if err != nil {
		return nil
	}

	return &PluginState{
		Name:      pluginName,
		Directory: pluginDir,
		URL:       url,
		Branch:    branch,
		Hash:      hash,
	}
}

func (cmd *cmdEnv) getPluginURL(pluginDir, pluginName string) (string, error) {
	url, err := git.URL(pluginDir)
	if err != nil {
		action := fmt.Sprintf("skipping %q", pluginName)
		cmd.reportWarning(action, "cannot determine URL", err)

		return "", err
	}

	return url, nil
}

func (cmd *cmdEnv) getPluginBranch(pluginDir, pluginName string) (string, error) {
	// TODO: the caller should not need to specify ".git" or "HEAD".
	branch, err := git.BranchName(filepath.Join(pluginDir, ".git", "HEAD"))
	if err != nil {
		action := fmt.Sprintf("skipping %q", pluginName)
		cmd.reportWarning(action, "cannot determine branch", err)

		return "", err
	}

	return branch, nil
}

func (cmd *cmdEnv) getPluginHash(pluginDir, pluginName string) (git.Digest, error) {
	hash, err := git.HeadDigest(pluginDir)
	if err != nil {
		action := fmt.Sprintf("skipping %q", pluginName)
		cmd.reportWarning(action, "cannot determine SHA", err)

		return git.Digest{}, err
	}

	return hash, nil
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

func (cmd *cmdEnv) hasConfigChanged(pState *PluginState, pSpec PluginSpec) (bool, string) {
	switch {
	case pState.URL != pSpec.URL:
		return true, "plugin URL changed"
	case pState.Branch != pSpec.Branch:
		return true, fmt.Sprintf("switching from branch %s to %s", pState.Branch, pSpec.Branch)
	default:
		return false, ""
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

	if pSpec.Pin {
		return upRes
	}

	//nolint:gosec // pSpec.Directory comes from user's own config file
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

func (pSpec *PluginSpec) fullPath(dataDir string) string {
	switch pSpec.Opt {
	case true:
		return filepath.Join(dataDir, "opt", pSpec.Name)
	default:
		return filepath.Join(dataDir, "start", pSpec.Name)
	}
}
