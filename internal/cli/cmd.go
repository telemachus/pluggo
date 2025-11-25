package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/telemachus/opts"
)

type cmdEnv struct {
	homeDir       string
	name          string
	version       string
	confFile      string
	dataDir       string
	startDir      string
	optDir        string
	results       []result
	warnings      atomic.Uint64
	debugWanted   bool
	helpWanted    bool
	quietWanted   bool
	versionWanted bool
}

func cmdFrom(name, version string, args []string) (*cmdEnv, error) {
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

	if err := og.Parse(args); err != nil {
		return nil, fmt.Errorf("argument parsing error: %w", err)
	}

	// Return early (and without error) for help or version.
	if cmd.helpWanted {
		fmt.Print(cmdUsage)
		return cmd, nil
	}
	if cmd.versionWanted {
		fmt.Printf("%s %s\n", cmd.name, cmd.version)
		return cmd, nil
	}

	// We must know the user's HOME for future operations.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine HOME: %w", err)
	}
	cmd.homeDir = homeDir

	// Set default config if user hasn't specified their own.
	if cmd.confFile == "" {
		cmd.confFile = filepath.Join(cmd.homeDir, confFile)
	}

	return cmd, nil
}

func (cmd *cmdEnv) plugins() ([]pluginSpec, error) {
	cfg, err := cmd.loadConfig()
	if err != nil {
		return nil, err
	}

	if err := cmd.setupDirs(cfg.DataDir); err != nil {
		return nil, err
	}

	return cmd.filterPlugins(cfg.Plugins), nil
}

type config struct {
	Plugins []pluginSpec `json:"plugins"`
	DataDir []string     `json:"dataDir"`
}

func (cmd *cmdEnv) loadConfig() (config, error) {
	var cfg config

	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		return cfg, fmt.Errorf("cannot read config %q: %w", cmd.confFile, err)
	}

	if err := json.Unmarshal(conf, &cfg); err != nil {
		return cfg, fmt.Errorf("cannot parse config %q: %w", cmd.confFile, err)
	}

	return cfg, nil
}

func (cmd *cmdEnv) setupDirs(dataDirParts []string) error {
	// Substitute HOME in dataDir if present.
	if len(dataDirParts) >= 1 && dataDirParts[0] == "HOME" {
		dataDirParts[0] = cmd.homeDir
	}

	cmd.dataDir = filepath.Join(dataDirParts...)
	if cmd.dataDir == "" {
		return errors.New("dataDir is required in configuration")
	}

	cmd.startDir = filepath.Join(cmd.dataDir, "start")
	cmd.optDir = filepath.Join(cmd.dataDir, "opt")

	return nil
}

// filterPlugins drops any plugins that lack a name, URL, or branch.
func (cmd *cmdEnv) filterPlugins(plugins []pluginSpec) []pluginSpec {
	i := 0
	for _, pSpec := range plugins {
		if pSpec.Name == "" {
			if pSpec.URL != "" {
				fmt.Fprintf(os.Stderr, "%s: skipping plugin with URL %q: missing name\n", cmd.name, pSpec.URL)
			} else {
				fmt.Fprintf(os.Stderr, "%s: skipping plugin: missing name and URL\n", cmd.name)
			}

			continue
		}

		if pSpec.URL == "" {
			fmt.Fprintf(os.Stderr, "%s: skipping plugin %q: missing URL\n", cmd.name, pSpec.Name)

			continue
		}

		if pSpec.Branch == "" {
			fmt.Fprintf(os.Stderr, "%s: skipping plugin %q: missing branch\n", cmd.name, pSpec.Name)

			continue
		}

		plugins[i] = pSpec
		i++
	}

	return plugins[:i]
}

// pluginPath returns the full path where a plugin should be installed.
func (cmd *cmdEnv) pluginPath(pSpec pluginSpec) string {
	if pSpec.Opt {
		return filepath.Join(cmd.optDir, pSpec.Name)
	}

	return filepath.Join(cmd.startDir, pSpec.Name)
}

// warnf counts non-fatal failures and, in debug mode, displays them too.
func (cmd *cmdEnv) warnf(format string, args ...any) {
	cmd.warnings.Add(1)
	if cmd.debugWanted {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
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

Source code: https://github.com/telemachus/pluggo
`
