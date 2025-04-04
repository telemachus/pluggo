// Package cli creates and runs a command line interface.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/telemachus/pluggo/internal/opts"
)

type cmdEnv struct {
	name          string
	version       string
	confFile      string
	homeDir       string
	dataDir       string
	helpWanted    bool
	quietWanted   bool
	versionWanted bool
}

func cmdFrom(name, version string, args []string) (*cmdEnv, error) {
	cmd := &cmdEnv{name: name, version: version}

	og := opts.NewGroup(cmd.name)
	og.String(&cmd.confFile, "config", "")
	og.Bool(&cmd.helpWanted, "help")
	og.Bool(&cmd.helpWanted, "h")
	og.Bool(&cmd.quietWanted, "quiet")
	og.Bool(&cmd.versionWanted, "version")

	if err := og.Parse(args); err != nil {
		return nil, err
	}

	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := struct {
		Plugins []Plugin `json:"plugins"`
	}{
		Plugins: make([]Plugin, 0, 20),
	}

	if err := json.Unmarshal(conf, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Every repository must have a URL and a directory name.
	return slices.DeleteFunc(cfg.Plugins, func(plugin Plugin) bool {
		return plugin.URL == "" || plugin.Name == ""
	}), nil
}
