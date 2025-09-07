// Package cli creates and runs a command line interface.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/telemachus/opts"
)

type cmdEnv struct {
	name          string
	version       string
	confFile      string
	homeDir       string
	dataDir       string
	startDir      string
	optDir        string
	results       syncResults
	errCount      int
	warnCount     int
	mu            sync.Mutex
	debugWanted   bool
	helpWanted    bool
	quietWanted   bool
	versionWanted bool
}

func cmdFrom(name, version string, args []string) *cmdEnv {
	cmd := &cmdEnv{
		name:    name,
		version: version,
	}

	og := opts.NewGroup(cmd.name)
	og.String(&cmd.confFile, "config", "")
	og.Bool(&cmd.debugWanted, "debug")
	og.Bool(&cmd.helpWanted, "help")
	og.Bool(&cmd.helpWanted, "h")
	og.Bool(&cmd.quietWanted, "quiet")
	og.Bool(&cmd.versionWanted, "version")
	og.Bool(&cmd.versionWanted, "V")

	// Return if parsing fails or there are leftover arguments.
	if err := og.Parse(args); err != nil {
		cmd.reportError(err.Error())
		fmt.Fprint(os.Stderr, cmdUsage)

		return cmd
	}

	// Return if the user calls for help or version.
	if cmd.helpWanted {
		fmt.Print(cmdUsage)

		return cmd
	}
	if cmd.versionWanted {
		fmt.Printf("%s %s\n", cmd.name, cmd.version)

		return cmd
	}

	// Return if we cannot get the user's home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		cmd.reportError("cannot determine HOME: " + err.Error())

		return cmd
	}
	cmd.homeDir = homeDir

	// Only set config file path if user does not specify their own.
	if cmd.confFile == "" {
		cmd.confFile = filepath.Join(cmd.homeDir, confFile)
	}

	return cmd
}

func (cmd *cmdEnv) noOp() bool {
	return cmd.errCount > 0 || cmd.helpWanted || cmd.versionWanted
}

func (cmd *cmdEnv) plugins() []PluginSpec {
	if cmd.noOp() {
		return nil
	}

	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		reason := fmt.Sprintf("cannot read config %q: ", cmd.confFile)
		cmd.reportError(reason + err.Error())

		return nil
	}

	cfg := struct {
		Plugins []PluginSpec `json:"plugins"`
		DataDir []string     `json:"dataDir"`
	}{
		Plugins: make([]PluginSpec, 0, 20),
		DataDir: make([]string, 0, 10),
	}

	if err := json.Unmarshal(conf, &cfg); err != nil {
		reason := fmt.Sprintf("cannot parse config %q: ", cmd.confFile)
		cmd.reportError(reason + err.Error())

		return nil
	}

	if len(cfg.DataDir) >= 1 && cfg.DataDir[0] == "HOME" {
		cfg.DataDir[0] = cmd.homeDir
	}
	cmd.dataDir = filepath.Join(cfg.DataDir...)
	if cmd.dataDir == "" {
		cmd.reportError("dataDir is required in configuration file")

		return nil
	}
	cmd.startDir = filepath.Join(cmd.dataDir, "start")
	cmd.optDir = filepath.Join(cmd.dataDir, "opt")

	// Every plugin must specify a URL, a name, and a branch.
	return slices.DeleteFunc(cfg.Plugins, func(pSpec PluginSpec) bool {
		return pSpec.URL == "" || pSpec.Name == "" || pSpec.Branch == ""
	})
}

func (cmd *cmdEnv) pluginPath(pSpec PluginSpec) string {
	if pSpec.Opt {
		return filepath.Join(cmd.optDir, pSpec.Name)
	}

	return filepath.Join(cmd.startDir, pSpec.Name)
}

func (cmd *cmdEnv) reportWarning(action, reason string, err error) {
	cmd.warnCount++
	if cmd.debugWanted && err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s: %s: %s\n", cmd.name, action, reason, err)
	} else {
		fmt.Fprintf(os.Stderr, "%s: %s: %s\n", cmd.name, action, reason)
	}
}

func (cmd *cmdEnv) reportError(msg string) {
	cmd.errCount++
	fmt.Fprintln(os.Stderr, cmd.name+": cannot continue: "+msg)
}

func (cmd *cmdEnv) incrementWarn() {
	cmd.mu.Lock()
	cmd.warnCount++
	cmd.mu.Unlock()
}

func (cmd *cmdEnv) resolveExitValue() int {
	if cmd.errCount+cmd.warnCount > 0 {
		return exitFailure
	}

	return exitSuccess
}

var cmdUsage = `usage: pluggo [options]

Manage Vim or Neovim plugins

Options:
      --config=FILE	Use FILE as config file (default ~/.pluggo.json)
      --quiet		Print only error messages
      --debug		Print additional low-level error messages


General:
  -h, --help		Print this help and exit
  -V, --version		Print version and exit

For more information or to file a bug report, visit https://github.com/telemachus/pluggo
`
