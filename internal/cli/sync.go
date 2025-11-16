package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/telemachus/pluggo/internal/git"
)

// PluginSpec represents a plugin specified in a user's configuration file.
type PluginSpec struct {
	URL    string
	Name   string
	Branch string
	Opt    bool
	Pinned bool
}

type syncResults []result

func (cmd *cmdEnv) sync(pSpecs []PluginSpec, reporter *consoleReporter) {
	if cmd.noOp() || !cmd.ensurePluginDirs() {
		return
	}

	statesByName := cmd.makeStateMap()
	specsByName := makeSpecMap(pSpecs)

	unwanted := findUnwanted(statesByName, specsByName)
	cmd.results = make(syncResults, 0, len(pSpecs)+len(unwanted))

	reporter.start(cmd.name + ": processing " + strconv.Itoa(cap(cmd.results)) + " plugins...")
	cmd.removeAll(unwanted)
	cmd.reconcileLocal(statesByName, pSpecs)
}

func (cmd *cmdEnv) ensurePluginDirs() bool {
	for _, wantedDir := range []string{cmd.startDir, cmd.optDir} {
		if err := os.MkdirAll(wantedDir, os.ModePerm); err != nil {
			reason := fmt.Sprintf("cannot create directory %q: ", wantedDir)
			cmd.reportError(reason + err.Error())

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
	ch := make(chan result, len(pSpecs))
	for _, spec := range pSpecs {
		// cmd.reconcile is safe even when statesByName[spec.Name] is nil.
		go cmd.reconcile(statesByName[spec.Name], spec, ch)
	}

	for range len(pSpecs) {
		res := <-ch
		cmd.results = append(cmd.results, res)
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

		res := result{plugin: pluginName}
		res.opResult.set(opRemoved)
		cmd.results = append(cmd.results, res)
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
	res := result{plugin: pSpec.Name}

	if err := cmd.install(pSpec); err != nil {
		cmd.incrementWarn()
		res.opResult.set(opError)
		ch <- res
		return
	}

	res.opResult.set(opInstalled)
	ch <- res
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
	res := result{plugin: pSpec.Name}
	if err := cmd.reinstall(pState.Directory, pSpec); err != nil {
		cmd.incrementWarn()
		res.opResult.set(opError)
		ch <- res
		return
	}

	res.opResult.set(opReinstalled)
	res.reason = reason
	ch <- res
}

func (cmd *cmdEnv) manageUpdate(pState *PluginState, pSpec PluginSpec, ch chan<- result) {
	res := cmd.update(pState, pSpec)

	if res.opResult.has(opError) {
		cmd.incrementWarn()
		ch <- res

		return
	}

	if res.opResult.has(opPinned) {
		ch <- res

		return
	}

	newHash, err := git.HeadDigest(pState.Directory)
	if err != nil {
		cmd.incrementWarn()
		res.opResult.set(opError)
		ch <- res

		return
	}

	if !pState.Hash.Equals(newHash) {
		res.opResult.set(opUpdated)
	}

	ch <- res
}
