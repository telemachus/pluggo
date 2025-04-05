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

	"github.com/telemachus/pluggo/internal/opts"
)

type cmdEnv struct {
	name          string
	version       string
	confFile      string
	homeDir       string
	dataDir       string
	exitVal       int
	helpWanted    bool
	quietWanted   bool
	versionWanted bool
}

// TODO: write cmd.noOp method.

// TODO: return only *cmdEnv, but make sure that cmd.exitVal is properly set.
func cmdFrom(name, version string, args []string) (*cmdEnv, error) {
	cmd := &cmdEnv{name: name, version: version, exitVal: exitSuccess}

	og := opts.NewGroup(cmd.name)
	og.String(&cmd.confFile, "config", "")
	og.Bool(&cmd.helpWanted, "help")
	og.Bool(&cmd.helpWanted, "h")
	og.Bool(&cmd.quietWanted, "quiet")
	og.Bool(&cmd.versionWanted, "version")

	if err := og.Parse(args); err != nil {
		cmd.exitVal = exitFailure

		return cmd, err
	}

	// Quick and dirty, but why be fancy in these cases?
	if cmd.helpWanted {
		fmt.Print(cmdUsage)
		// TODO: refactor this to return cmd, nil. Exit from caller.
		os.Exit(cmd.exitVal)
	}
	if cmd.versionWanted {
		fmt.Printf("%s %s\n", cmd.name, cmd.version)
		// TODO: refactor this to return cmd, nil. Exit from caller.
		os.Exit(cmd.exitVal)
	}

	// Do not continue if we cannot parse and validate arguments or get the
	// user's home directory.
	extraArgs := og.Args()
	if err := validate(extraArgs); err != nil {
		cmd.exitVal = exitFailure

		return cmd, err
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		cmd.exitVal = exitFailure

		return cmd, err
	}
	cmd.homeDir = homeDir

	// Set default config file path if not specified
	if cmd.confFile == "" {
		cmd.confFile = filepath.Join(cmd.homeDir, confFile)
	}
	cmd.dataDir = filepath.Join(cmd.homeDir, dataDir)

	return cmd, nil
}

func (cmd *cmdEnv) plugins() ([]Plugin, error) {
	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		// TODO: set cmd.exitVal, print error, and return.
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := struct {
		Plugins []Plugin `json:"plugins"`
	}{
		Plugins: make([]Plugin, 0, 20),
	}

	if err := json.Unmarshal(conf, &cfg); err != nil {
		// TODO: set cmd.exitVal, print error, and return.
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Every repository must have a URL and a directory name.
	return slices.DeleteFunc(cfg.Plugins, func(plugin Plugin) bool {
		return plugin.URL == "" || plugin.Name == ""
	}), nil
}

func validate(extra []string) error {
	extraCount := len(extra)
	var s rune
	if extraCount > 0 {
		if extraCount > 1 {
			s = 's'
		}

		return fmt.Errorf(
			"unrecognized argument%c: %s",
			s,
			quotedSlice(extra),
		)
	}

	return nil
}

func quotedSlice(items []string) string {
	quotedSlice := make([]string, len(items))
	for i, str := range items {
		quotedSlice[i] = strconv.Quote(str)
	}

	return strings.Join(quotedSlice, " ")
}

var cmdUsage = `usage: pluggo [options]

Manage Vim or Neovim plugins

Options
    --config=FILE    Use FILE as config file (default ~/.pluggo.json)
    --quiet          Print only error messages

-h, --help           Print this help and exit
    --version        Print version and exit

For more information or to file a bug report, visit https://github.com/telemachus/pluggo
`
