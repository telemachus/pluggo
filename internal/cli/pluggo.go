package cli

import (
	"fmt"
	"os"
)

const (
	cmdName     = "pluggo"
	cmdVersion  = "v0.0.1"
	confFile    = ".pluggo.json"
	lockFile    = ".pluggo.lock.json"
	dataDir     = ".local/share/nvim/site/pack/pluggo"
	exitSuccess = 0
	exitFailure = 1
)

// Pluggo runs the plugin manager and returns success or failure to the shell.
func Pluggo(args []string) int {
	cmd, err := cmdFrom(cmdName, cmdVersion, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return exitFailure
	}

	// Check for help or version flags
	if cmd.helpWanted {
		fmt.Print(cmdUsage)
		return exitSuccess
	}

	if cmd.versionWanted {
		fmt.Printf("%s %s\n", cmdName, cmdVersion)
		return exitSuccess
	}

	// Load configuration
	plugins, err := cmd.plugins()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return exitFailure
	}

	// Load lockfile
	lock, err := cmd.loadLockFile()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to load lockfile: %s\n", cmdName, err)
		return exitFailure
	}

	// Reconcile plugin state with configuration
	if err = cmd.reconcile(plugins, lock); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return exitFailure
	}

	// Save lockfile
	if err = cmd.saveLockFile(lock); err != nil {
		fmt.Fprintf(os.Stderr, "%s: failed to save lockfile: %s\n", cmdName, err)
		return exitFailure
	}

	return exitSuccess
}
