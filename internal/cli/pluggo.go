// Package cli creates and runs a command line interface.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

const (
	cmdName    = "pluggo"
	cmdVersion = "v0.10.0"
	confFile   = ".pluggo.json"
)

// Pluggo runs the plugin manager and returns success or failure to the shell.
func Pluggo(args []string) int {
	cmd, err := cmdFrom(cmdName, cmdVersion, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return 1
	}

	if cmd.helpWanted || cmd.versionWanted {
		return 0
	}

	// Listen for SIGINT (Ctrl+C).
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	plugins, err := cmd.plugins()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return 1
	}

	if err := cmd.process(ctx, plugins); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return 1
	}

	if cmd.warnings.Load() > 0 {
		return 1
	}

	return 0
}

func (cmd *cmdEnv) process(ctx context.Context, pSpecs []pluginSpec) error {
	rep := newReporter("    ", cmd.quietWanted)

	if err := cmd.sync(ctx, pSpecs, rep); err != nil {
		return err
	}

	rep.finish(cmd.results)

	return nil
}
