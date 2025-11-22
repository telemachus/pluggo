package cli

import (
	"context"
	"fmt"
	"os"
)

// sync brings the local plugin state into agreement with the config file.
func (cmd *cmdEnv) sync(ctx context.Context, pSpecs []pluginSpec, rep *reporter) error {
	if err := cmd.ensurePluginDirs(); err != nil {
		return err
	}

	rep.start(cmd.name + ": processing plugins...")

	statesByName := cmd.makeStateMap(ctx)
	specsByName := makeSpecMap(pSpecs)

	unwanted := findUnwanted(statesByName, specsByName)
	cmd.results = make([]result, 0, len(pSpecs)+len(unwanted))

	cmd.removeAll(unwanted)
	cmd.reconcileLocal(ctx, statesByName, pSpecs)

	return nil
}

// ensurePluginDirs creates the start/ and opt/ directories if needed.
func (cmd *cmdEnv) ensurePluginDirs() error {
	for _, wantedDir := range []string{cmd.startDir, cmd.optDir} {
		if err := os.MkdirAll(wantedDir, 0o755); err != nil {
			return fmt.Errorf("cannot create directory %q: %w", wantedDir, err)
		}
	}

	return nil
}

// makeSpecMap converts a slice of plugin specs into a map by name.
func makeSpecMap(pSpecs []pluginSpec) map[string]pluginSpec {
	specsByName := make(map[string]pluginSpec, len(pSpecs))
	for _, p := range pSpecs {
		specsByName[p.Name] = p
	}

	return specsByName
}

// findUnwanted identifies plugins installed locally but not in the config.
func findUnwanted(statesByName map[string]*pluginState, specsByName map[string]pluginSpec) map[string]string {
	unwanted := make(map[string]string, len(statesByName))
	for pluginName, state := range statesByName {
		if _, exists := specsByName[pluginName]; !exists {
			unwanted[pluginName] = state.directory
		}
	}

	return unwanted
}

// removeAll removes unwanted plugins.
func (cmd *cmdEnv) removeAll(unwanted map[string]string) {
	for pluginName, pluginPath := range unwanted {
		if err := os.RemoveAll(pluginPath); err != nil {
			cmd.warnf("%s: skipping %q: failed to remove plugin: %s", cmd.name, pluginName, err)
			continue
		}

		cmd.results = append(cmd.results, result{
			plugin: pluginName,
			status: removed,
		})
	}
}

// reconcileLocal processes all plugins in parallel using goroutines.
func (cmd *cmdEnv) reconcileLocal(ctx context.Context, statesByName map[string]*pluginState, pSpecs []pluginSpec) {
	const maxWorkers = 15
	sem := make(chan struct{}, maxWorkers)
	ch := make(chan result, len(pSpecs))

	for _, spec := range pSpecs {
		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			cmd.reconcile(ctx, statesByName[spec.Name], spec, ch)
		}()
	}

	for range pSpecs {
		res := <-ch
		cmd.results = append(cmd.results, res)
	}
}

// reconcile determines what action to take for a single plugin.
// This is the main decision tree: if not installed, install; if config changed, reinstall; otherwise move (if needed) and update (unless pinned).
func (cmd *cmdEnv) reconcile(ctx context.Context, pState *pluginState, pSpec pluginSpec, ch chan<- result) {
	// Plugin not installed locally: clone it.
	if pState == nil {
		cmd.manageClone(ctx, pSpec, ch)
		return
	}

	// URL or branch have changed: reinstall.
	if changed, reason := cmd.hasConfigChanged(pState, pSpec); changed {
		cmd.manageReinstall(ctx, pState, pSpec, reason, ch)
		return
	}

	// URL and branch unchanged: move if needed, then update if not pinned.
	cmd.manageMoveAndUpdate(ctx, pState, pSpec, ch)
}

func (cmd *cmdEnv) manageClone(ctx context.Context, pSpec pluginSpec, ch chan<- result) {
	if err := clone(ctx, pSpec.URL, pSpec.Branch, cmd.pluginPath(pSpec)); err != nil {
		cmd.warnf("%s: clone %q failed: %s", cmd.name, pSpec.Name, err)
		ch <- result{
			plugin: pSpec.Name,
			err:    err,
		}

		return
	}

	ch <- result{
		plugin: pSpec.Name,
		status: installed,
	}
}

func (cmd *cmdEnv) manageReinstall(ctx context.Context, pState *pluginState, pSpec pluginSpec, reason string, ch chan<- result) {
	if err := cmd.reinstall(ctx, pState.directory, pSpec); err != nil {
		cmd.warnf("%s: reinstall %q failed: %s", cmd.name, pSpec.Name, err)
		ch <- result{
			plugin: pSpec.Name,
			err:    err,
		}

		return
	}

	ch <- result{
		plugin: pSpec.Name,
		status: reinstalled,
		reason: reason,
	}
}

func (cmd *cmdEnv) manageMoveAndUpdate(ctx context.Context, pState *pluginState, pSpec pluginSpec, ch chan<- result) {
	res := result{
		plugin: pSpec.Name,
		// Default status is unchanged.
		status: unchanged,
	}

	// First, move the plugin if requested.
	movedTo, err := cmd.move(pState, pSpec)
	if err != nil {
		cmd.warnf("%s: move %q failed: %s", cmd.name, pSpec.Name, err)
		res.err = err
		ch <- res

		return
	}

	if movedTo != "" {
		res.movedTo = movedTo
	}

	// Next, update the plugin if not pinned.
	if pSpec.Pinned {
		res.pinned = true
		ch <- res

		return
	}

	oldHash := pState.hash
	if updateErr := cmd.update(ctx, pState); updateErr != nil {
		cmd.warnf("%s: update %q failed: %s", cmd.name, pSpec.Name, updateErr)
		res.err = updateErr
		ch <- res

		return
	}

	// Determine whether the plugin was actually updated.
	info, err := getBranchInfo(ctx, pState.directory)
	if err != nil {
		cmd.warnf("%s: cannot determine new hash for %q: %s", cmd.name, pSpec.Name, err)
		res.err = err
		ch <- res

		return
	}

	if !oldHash.equals(info.hash) {
		res.status = updated
	}

	ch <- res
}
