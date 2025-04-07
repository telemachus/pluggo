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

func cmdFrom(name, version string, args []string) *cmdEnv {
	cmd := &cmdEnv{name: name, version: version, exitVal: exitSuccess}

	og := opts.NewGroup(cmd.name)
	og.String(&cmd.confFile, "config", "")
	og.Bool(&cmd.helpWanted, "help")
	og.Bool(&cmd.helpWanted, "h")
	og.Bool(&cmd.quietWanted, "quiet")
	og.Bool(&cmd.versionWanted, "version")

	if err := og.Parse(args); err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.name, err)

		return cmd
	}

	// If the user calls for help or version, we're done.
	if cmd.helpWanted {
		fmt.Print(cmdUsage)

		return cmd
	}
	if cmd.versionWanted {
		fmt.Printf("%s %s\n", cmd.name, cmd.version)

		return cmd
	}

	// Do not continue if we cannot parse and validate arguments.
	extraArgs := og.Args()
	if err := validate(extraArgs); err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmd.name, err)

		return cmd
	}

	// Do not continue if we cannot get the user's home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(
			os.Stderr,
			"%s: cannot get user's home directory: %s\n",
			cmd.name,
			err,
		)

		return cmd
	}
	cmd.homeDir = homeDir
	// TODO: consider whether to make this configurable, probably in the
	// config file and maybe also on the command line.
	cmd.dataDir = filepath.Join(cmd.homeDir, dataDir)

	// Only set config file path if user does not specify their own.
	if cmd.confFile == "" {
		cmd.confFile = filepath.Join(cmd.homeDir, confFile)
	}

	return cmd
}

func (cmd *cmdEnv) noOp() bool {
	return cmd.exitVal != exitSuccess || cmd.helpWanted || cmd.versionWanted
}

func (cmd *cmdEnv) plugins() []PluginSpec {
	if cmd.noOp() {
		return nil
	}

	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(
			os.Stderr,
			"%s: failed to read config file: %s\n",
			cmd.name,
			err,
		)

		return nil
	}

	cfg := struct {
		Plugins []PluginSpec `json:"plugins"`
	}{
		Plugins: make([]PluginSpec, 0, 20),
	}

	if err := json.Unmarshal(conf, &cfg); err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(
			os.Stderr,
			"%s: failed to parse config file: %s\n",
			cmd.name,
			err,
		)

		return nil
	}

	// Every repository must specify a URL, a directory name, and a branch.
	return slices.DeleteFunc(cfg.Plugins, func(pSpec PluginSpec) bool {
		return pSpec.URL == "" || pSpec.Name == "" || pSpec.Branch == ""
	})
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
