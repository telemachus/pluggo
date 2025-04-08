package cli

const (
	cmdName     = "pluggo"
	cmdVersion  = "v0.1.0"
	confFile    = ".pluggo.json"
	lockFile    = ".pluggo.lock.json"
	exitSuccess = 0
	exitFailure = 1
)

// Pluggo runs the plugin manager and returns success or failure to the shell.
func Pluggo(args []string) int {
	cmd := cmdFrom(cmdName, cmdVersion, args)

	plugins := cmd.plugins()
	cmd.reconcile(plugins)

	// TODO: return cmd.exitVal + len(cmd.errs)? This would allow me to
	// track smaller errors without fully bailing out. That is, cmd.exitVal
	// causes cmd.noOp() to return true, but len(cmd.errs) does not. This
	// matters because if I set cmd.exitVal to exitFailure because of minor
	// errors, then (e.g.) the lockfile will not be saved.
	return cmd.exitVal
}
