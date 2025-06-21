package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

func (cmd *cmdEnv) hasConfigChanged(pState *PluginState, pSpec PluginSpec) (changed bool, reason string) {
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

func (pSpec *PluginSpec) fullPath(dataDir string) string {
	switch pSpec.Opt {
	case true:
		return filepath.Join(dataDir, "opt", pSpec.Name)
	default:
		return filepath.Join(dataDir, "start", pSpec.Name)
	}
}
