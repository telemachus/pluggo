// Package cli creates and runs a command line interface.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/telemachus/opts"
)

type cmdEnv struct {
	log           *logger
	homeDir       string
	name          string
	version       string
	confFile      string
	dataDir       string
	startDir      string
	optDir        string
	results       syncResults
	stats         stats
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
	// Create logger early for use in setup.
	cmd.log = newLogger(false, false)

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
		cmd.errorf("%s: argument parsing error: %s", cmd.name, err)

		return cmd
	}
	extraArgs := og.Args()
	if err := validate(extraArgs); err != nil {
		cmd.errorf("%s: %s", cmd.name, err)

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

	// Update logger with actual settings.
	cmd.log = newLogger(cmd.debugWanted, cmd.quietWanted)

	// Return if we cannot get the user's home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		cmd.errorf("%s: cannot determine HOME: %s", cmd.name, err)

		return cmd
	}
	cmd.homeDir = homeDir

	// Set config file path only if user does not specify their own.
	if cmd.confFile == "" {
		cmd.confFile = filepath.Join(cmd.homeDir, confFile)
	}

	return cmd
}

func (cmd *cmdEnv) noOp() bool {
	_, errs := cmd.stats.snapshot()
	return errs > 0 || cmd.helpWanted || cmd.versionWanted
}

func (cmd *cmdEnv) plugins() []PluginSpec {
	if cmd.noOp() {
		return nil
	}

	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		cmd.errorf("%s: cannot read config %q: %s", cmd.name, cmd.confFile, err)

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
		cmd.errorf("%s: cannot parse config %q: %s", cmd.name, cmd.confFile, err)

		return nil
	}

	if len(cfg.DataDir) >= 1 && cfg.DataDir[0] == "HOME" {
		cfg.DataDir[0] = cmd.homeDir
	}
	cmd.dataDir = filepath.Join(cfg.DataDir...)
	if cmd.dataDir == "" {
		cmd.errorf("%s: dataDir is required in configuration", cmd.name)

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

func (cmd *cmdEnv) errorf(format string, args ...any) {
	cmd.stats.incError()
	cmd.log.errorf(format, args...)
}

func (cmd *cmdEnv) warnf(format string, args ...any) {
	cmd.stats.incWarn()
	cmd.log.warnf(format, args...)
}

func (cmd *cmdEnv) debugf(format string, args ...any) {
	cmd.log.debugf(format, args...)
}

func validate(extra []string) error {
	extraCount := len(extra)
	var s rune
	if extraCount > 0 {
		if extraCount > 1 {
			s = 's'
		}

		return fmt.Errorf("unrecognized argument%c: %s", s, quotedSlice(extra))
	}

	return nil
}

func quotedSlice(items []string) string {
	if len(items) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(strconv.Quote(items[0]))
	for _, str := range items[1:] {
		b.WriteString(" ")
		b.WriteString(strconv.Quote(str))
	}

	return b.String()
}

func (cmd *cmdEnv) resolveExitValue() int {
	warns, errs := cmd.stats.snapshot()
	if warns+errs > 0 {
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
