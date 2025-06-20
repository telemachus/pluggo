package cli

const (
	cmdName     = "pluggo"
	cmdVersion  = "v0.6.0"
	confFile    = ".pluggo.json"
	exitSuccess = 0
	exitFailure = 1
)

// Pluggo runs the plugin manager and returns success or failure to the shell.
func Pluggo(args []string) int {
	cmd := cmdFrom(cmdName, cmdVersion, args)

	plugins := cmd.plugins()
	cmd.sync(plugins)

	return cmd.resolveExitValue()
}
