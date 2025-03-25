// Package cli creates and runs a command line interface.
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/telemachus/pluggo/internal/opts"
)

type cmdEnv struct {
	name          string
	version       string
	subCmdName    string
	confFile      string
	homeDir       string
	dataDir       string
	subCmdArgs    []string
	exitVal       int
	helpWanted    bool
	quietWanted   bool
	versionWanted bool
}

func cmdFrom(name, version string, args []string) (*cmdEnv, error) {
	cmd := &cmdEnv{name: name, version: version, exitVal: exitSuccess}

	og := opts.NewGroup(cmd.name)
	og.String(&cmd.confFile, "config", "")
	og.Bool(&cmd.helpWanted, "help")
	og.Bool(&cmd.helpWanted, "h")
	og.Bool(&cmd.quietWanted, "quiet")
	og.Bool(&cmd.versionWanted, "version")

	if err := og.Parse(args); err != nil {
		return nil, err
	}

	// Quick and dirty, but why be fancy in these cases?
	if cmd.helpWanted {
		fmt.Print(cmdUsage)
		os.Exit(cmd.exitVal)
	}
	if cmd.versionWanted {
		fmt.Printf("%s %s\n", cmd.name, cmd.version)
		os.Exit(cmd.exitVal)
	}

	// Do not continue if we cannot parse and validate arguments or get the
	// user's home directory.
	extraArgs := og.Args()
	if err := validate(extraArgs); err != nil {
		return nil, err
	}
	cmd.subCmdName = extraArgs[0]
	cmd.subCmdArgs = extraArgs[1:]
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cmd.homeDir = homeDir

	if cmd.confFile == "" {
		cmd.confFile = filepath.Join(cmd.homeDir, confFile)
	}
	cmd.dataDir = filepath.Join(cmd.homeDir, dataDir)

	return cmd, nil
}

func (cmd *cmdEnv) subCmdFrom(args []string) error {
	// We need to preserve quietWanted in case the user set it earlier.
	pluggoQuietWanted := cmd.quietWanted

	og := opts.NewGroup(cmd.name + " " + cmd.subCmdName)
	og.String(&cmd.confFile, "config", cmd.confFile)
	og.Bool(&cmd.helpWanted, "help")
	og.Bool(&cmd.helpWanted, "h")
	og.Bool(&cmd.quietWanted, "quiet")
	og.Bool(&cmd.versionWanted, "version")

	if err := og.Parse(args); err != nil {
		cmd.exitVal = exitFailure
		return err
	}
	// If the user passes --quiet as an argument to pluggo, then we
	// should pass that value to pluggo <subcommand>.
	if pluggoQuietWanted {
		cmd.quietWanted = pluggoQuietWanted
	}

	// Quick and dirty, but why be fancy in these cases?
	if cmd.helpWanted {
		// TODO: make this print correct usage for subCmdName.
		cmd.subCmdUsage(cmd.subCmdName)
		os.Exit(cmd.exitVal)
	}
	if cmd.versionWanted {
		fmt.Printf("%s %s %s\n", cmd.name, cmd.subCmdName, cmd.version)
		os.Exit(cmd.exitVal)
	}

	// There should be no extra arguments.
	extraArgs := og.Args()
	if len(extraArgs) != 0 {
		cmd.exitVal = exitFailure
		var s string
		if len(extraArgs) > 1 {
			s = "s"
		}
		return fmt.Errorf("unrecognized argument%s: %+v", s, extraArgs)
	}

	return nil
}

func (cmd *cmdEnv) subCmdUsage(subCmdName string) {
	switch subCmdName {
	case "update", "up":
		fmt.Print(updateUsage)
	case "install":
		fmt.Print(installUsage)
	case "sync":
		fmt.Print(syncUsage)
	default:
		fmt.Fprintf(os.Stderr, "%s %s: unrecognized subcommand %q\n", cmd.name, cmd.subCmdName, subCmdName)
	}
}

func (cmd *cmdEnv) noOp() bool {
	return cmd.exitVal != exitSuccess
}

func (cmd *cmdEnv) prettyPath(s string) string {
	return strings.Replace(s, cmd.homeDir, "~", 1)
}

func (cmd *cmdEnv) repos() []Repo {
	if cmd.noOp() {
		return nil
	}

	conf, err := os.ReadFile(cmd.confFile)
	if err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
		return nil
	}

	cfg := struct {
		Repos []Repo `json:"repos"`
	}{
		Repos: make([]Repo, 0, 20),
	}
	err = json.Unmarshal(conf, &cfg)
	if err != nil {
		cmd.exitVal = exitFailure
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
		return nil
	}

	// Every repository must have a URL and a directory name.
	return slices.DeleteFunc(cfg.Repos, func(r Repo) bool {
		return r.URL == "" || r.Name == ""
	})
}

func validate(extra []string) error {
	if len(extra) < 1 {
		return errors.New("a subcommand is required")
	}

	// The only recognized subcommands are clone, up(date), and sync.
	recognized := map[string]struct{}{
		"clone":   {},
		"install": {},
		"sync":    {},
	}
	if _, ok := recognized[extra[0]]; !ok {
		return fmt.Errorf("unrecognized subcommand: %q", extra[0])
	}

	return nil
}
