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

// PluginSpec represents the state of a plugin as declared by a user's
// configuration file.
type PluginSpec struct {
	URL    string
	Name   string
	Branch string
	Opt    bool
	Pin    bool
}

// PluginState represents the state of a plugin as determined from the
// filesystem.
type PluginState struct {
	Name      string
	Directory string
	URL       string
	Branch    string
	Hash      string
}

func (cmd *cmdEnv) discoverPlugins() map[string]*PluginState {
	if cmd.noOp() {
		return nil
	}

	statesByName := make(map[string]*PluginState, 15)

	for _, dir := range []string{"start", "opt"} {
		baseDir := filepath.Join(cmd.dataDir, dir)

		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(baseDir)
		if err != nil {
			// TODO: don't set exitVal for this. Use an error field
			// that doesn't trigger noOp.
			cmd.exitVal = exitFailure
			fmt.Fprintf(
				os.Stderr,
				"%s: failed to read plugin directory %s: %s\n",
				cmd.name,
				baseDir,
				err,
			)

			continue
		}

		// Remove all entries that are not directories or not git repos.
		entries = slices.DeleteFunc(entries, func(entry os.DirEntry) bool {
			return !entry.IsDir() || !isGitRepo(entry, baseDir)
		})

		for _, entry := range entries {
			pluginDir := filepath.Join(baseDir, entry.Name())

			url, err := git.URL(pluginDir)
			if err != nil {
				// TODO: don't set exitVal for this. Use an error field
				// that doesn't trigger noOp.
				cmd.exitVal = exitFailure
				fmt.Fprintf(
					os.Stderr,
					"%s: failed to get URL for plugin %s: %s\n",
					cmd.name,
					entry.Name(),
					err,
				)

				continue
			}

			branch, err := git.BranchName(filepath.Join(pluginDir, ".git", "HEAD"))
			if err != nil {
				// TODO: bail out if we cannot get a branch?
				branch = ""
			}

			hash, err := git.HeadDigestString(pluginDir)
			if err != nil {
				// TODO: don't set exitVal for this. Use an error field
				// that doesn't trigger noOp.
				cmd.exitVal = exitFailure
				fmt.Fprintf(
					os.Stderr,
					"%s: failed to get hash for plugin %s: %s\n",
					cmd.name,
					entry.Name(),
					err,
				)

				continue
			}

			statesByName[entry.Name()] = &PluginState{
				Name:      entry.Name(),
				Directory: pluginDir,
				URL:       url,
				Branch:    branch,
				Hash:      hash,
			}
		}
	}

	return statesByName
}

func (cmd *cmdEnv) reconcile(pSpecs []PluginSpec) {
	if cmd.noOp() {
		return
	}

	statesByName := cmd.discoverPlugins()
	if statesByName == nil && cmd.exitVal == exitFailure {
		return
	}

	specsByName := make(map[string]PluginSpec, len(pSpecs))
	for _, p := range pSpecs {
		specsByName[p.Name] = p
	}

	// Abort if plugin directory does not exist and cannot be created.
	if err := os.MkdirAll(cmd.dataDir, os.ModePerm); err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(
			os.Stderr,
			"%s: failed to create plugin directory: %s\n",
			cmd.name,
			err,
		)

		return
	}

	ch := make(chan result)
	for _, spec := range pSpecs {
		go cmd.process(spec, statesByName[spec.Name], ch)
	}

	// Collect results
	var errs []string
	for range pSpecs {
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

	// Remove plugins that are in the filesystem but not the config.
	for stateName, state := range statesByName {
		if _, exists := specsByName[stateName]; !exists {
			if err := os.RemoveAll(state.Directory); err != nil {
				errs = append(
					errs,
					fmt.Sprintf("failed to remove %s: %s", stateName, err),
				)

				continue
			}

			if !cmd.quietWanted {
				fmt.Printf("%s: removed (not in configuration)\n", stateName)
			}
		}
	}

	if len(errs) > 0 {
		fmt.Fprintf(
			os.Stderr,
			"%s: failed to reconcile plugins:\n%s",
			cmd.name,
			strings.Join(errs, "\n"),
		)

		return
	}
}

type updateResult struct {
	moved      bool
	pin        bool
	toOpt      bool
	hashBefore string
	hashAfter  string
	err        error
}

func (cmd *cmdEnv) process(pSpec PluginSpec, pState *PluginState, ch chan<- result) {
	pluginDir := pSpec.fullPath(cmd.dataDir)

	// Case 1: a plugin in pSpec but not pState needs installation.
	if pState == nil {
		if err := cmd.install(pSpec, pluginDir); err != nil {
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

		return
	}

	// Case 2: a plugin with a configuration change needs reinstallation.
	var reason string
	switch {
	case pState.URL != pSpec.URL:
		reason = "repo URL changed"
	case pState.Branch != pSpec.Branch:
		reason = fmt.Sprintf(
			"switching from branch %s to %s",
			pState.Branch,
			pSpec.Branch,
		)
	}

	if reason != "" {
		if err := cmd.reinstall(pSpec, pState.Directory); err != nil {
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

		return
	}

	// Case 3: if we reach this point, the plugin should be checked for an update.
	hashBefore := pState.Hash

	upRes := cmd.update(pSpec, pState)
	if upRes.err != nil {
		ch <- result{
			isErr: true,
			msg: fmt.Sprintf(
				"failed to update %s: %s",
				pSpec.Name,
				upRes.err,
			),
		}

		return
	}

	hashAfter, err := git.HeadDigestString(pState.Directory)
	if err != nil {
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("failed to get updated hash for %s: %s", pSpec.Name, err),
		}

		return
	}

	upRes.hashBefore = hashBefore[:7]
	upRes.hashAfter = hashAfter[:7]
	msg := formatUpdateMsg(pSpec.Name, upRes)

	ch <- result{
		isErr: false,
		msg:   msg,
	}
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

	if upRes.hashBefore != upRes.hashAfter {
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

func (cmd *cmdEnv) reinstall(pSpec PluginSpec, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to remove existing directory: %w", err)
	}

	return cmd.install(pSpec, pSpec.fullPath(cmd.dataDir))
}

func (cmd *cmdEnv) update(pSpec PluginSpec, pState *PluginState) updateResult {
	upRes := updateResult{pin: pSpec.Pin}
	pluginDir := pSpec.fullPath(cmd.dataDir)

	// Move the plugin if necessary.
	if pluginDir != pState.Directory {
		upRes.moved = true
		upRes.toOpt = pSpec.Opt

		if err := os.MkdirAll(filepath.Dir(pluginDir), os.ModePerm); err != nil {
			upRes.err = fmt.Errorf(
				"failed to create parent directory for move: %w",
				err,
			)

			return upRes
		}

		if err := os.Rename(pState.Directory, pluginDir); err != nil {
			upRes.err = fmt.Errorf(
				"failed to move plugin directory: %w",
				err,
			)

			return upRes
		}

		pState.Directory = pluginDir
	}

	// If the plugin is pinned, then we should not update.
	if pSpec.Pin {
		return upRes
	}

	updateCmd := exec.Command(
		"git",
		"-C",
		pState.Directory,
		"pull",
		"--recurse-submodules",
	)
	if err := updateCmd.Run(); err != nil {
		upRes.err = fmt.Errorf("git pull failed: %s", err)
	}

	return upRes
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

func isGitRepo(entry os.DirEntry, baseDir string) bool {
	gitDir := filepath.Join(baseDir, entry.Name(), ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false
	}

	return true
}
