package cli

import (
	"fmt"
	"os"
)

const (
	cmdName     = "pluggo"
	cmdVersion  = "v0.0.1"
	confFile    = ".pluggo.json"
	dataDir     = ".local/share/nvim/site/pack/pluggo"
	exitSuccess = 0
	exitFailure = 1
)

// Pluggo runs a subcommand and returns success or failure to the shell.
func Pluggo(args []string) int {
	cmd, err := cmdFrom(cmdName, cmdVersion, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdName, err)
		return exitFailure
	}

	switch cmd.subCmdName {
	case "update", "up":
		subCmdUpdate(cmd)
	case "install":
		subCmdInstall(cmd)
	case "sync":
		subCmdSync(cmd)
	default:
		fmt.Fprintf(os.Stderr, "%s: unrecognized subcommand %q\n", cmd.name, cmd.subCmdName)
		cmd.exitVal = exitFailure
	}

	return cmd.exitVal
}
