package cli

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/telemachus/opts"
)

type cmdEnv struct {
	stdout        io.Writer
	stderr        io.Writer
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

func cmdFrom(name, version string, args []string, stdout, stderr io.Writer) (*cmdEnv, error) {
	cmd := &cmdEnv{
		stdout:  stdout,
		stderr:  stderr,
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

	if err := og.ParseAll(args); err != nil {
		return nil, fmt.Errorf("argument parsing error: %w", err)
	}

	if cmd.helpWanted {
		fmt.Fprint(cmd.stdout, cmdUsage)
		return cmd, nil
	}
	if cmd.versionWanted {
		fmt.Fprintf(cmd.stdout, "%s %s\n", cmd.name, cmd.version)
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
	DataDir string       `json:"dataDir"`
	Plugins []pluginSpec `json:"plugins"`
}

func (cmd *cmdEnv) loadConfig() (config, error) {
	var cfg config

	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		return cfg, fmt.Errorf("cannot read config %q: %w", cmd.confFile, err)
	}

	if err := json.Unmarshal(conf, &cfg, json.RejectUnknownMembers(true)); err != nil {
		return cfg, fmt.Errorf("cannot parse config %q: %w", cmd.confFile, err)
	}

	return cfg, nil
}

func (cmd *cmdEnv) setupDirs(dataDir string) error {
	cmd.dataDir = tildeSwap(cmd.homeDir, dataDir)
	if cmd.dataDir == "" {
		return errors.New("dataDir is required in configuration")
	}
	if !filepath.IsAbs(cmd.dataDir) {
		return fmt.Errorf("dataDir must be an absolute path: %q", cmd.dataDir)
	}

	cmd.startDir = filepath.Join(cmd.dataDir, "start")
	cmd.optDir = filepath.Join(cmd.dataDir, "opt")

	return nil
}

// tildeSwap expands a bare "~" or a leading "~/" to the home directory and
// returns any other path unchanged.
func tildeSwap(home, path string) string {
	if path == "~" {
		return home
	}
	if s, ok := strings.CutPrefix(path, "~/"); ok {
		return filepath.Join(home, s)
	}

	return path
}

// filterPlugins drops plugins that lack a name, URL, or branch. It also drops
// those that have an invalid name or that repeat another plugin's name.
func (cmd *cmdEnv) filterPlugins(plugins []pluginSpec) []pluginSpec {
	seen := make(map[string]bool, len(plugins))
	i := 0
	for _, pSpec := range plugins {
		if problem := specProblem(pSpec, seen); problem != "" {
			switch {
			case pSpec.Name != "":
				fmt.Fprintf(cmd.stderr, "%s: skipping plugin %q: %s\n", cmd.name, pSpec.Name, problem)
			case pSpec.URL != "":
				fmt.Fprintf(cmd.stderr, "%s: skipping plugin with URL %q: %s\n", cmd.name, pSpec.URL, problem)
			default:
				fmt.Fprintf(cmd.stderr, "%s: skipping plugin: %s\n", cmd.name, problem)
			}

			continue
		}

		seen[pSpec.Name] = true
		plugins[i] = pSpec
		i++
	}

	return plugins[:i]
}

// specProblem returns why a plugin spec is unusable, or "" if it is usable.
func specProblem(pSpec pluginSpec, seen map[string]bool) string {
	switch {
	case pSpec.Name == "" && pSpec.URL == "":
		return "missing name and URL"
	case pSpec.Name == "":
		return "missing name"
	case !isValidName(pSpec.Name):
		return "name must be a single directory name"
	case seen[pSpec.Name]:
		return "duplicate name"
	case pSpec.URL == "":
		return "missing URL"
	case pSpec.Branch == "":
		return "missing branch"
	default:
		return ""
	}
}

// isValidName reports whether name can safely be a plugin's directory name.
func isValidName(name string) bool {
	return name != "." && filepath.IsLocal(name) && filepath.Base(name) == name
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
		fmt.Fprintf(cmd.stderr, format+"\n", args...)
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
