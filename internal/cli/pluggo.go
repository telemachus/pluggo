package cli

const (
	cmdName     = "pluggo"
	cmdVersion  = "v0.8.0"
	confFile    = ".pluggo.json"
	exitSuccess = 0
	exitFailure = 1
)

// Pluggo runs the plugin manager and returns success or failure to the shell.
func Pluggo(args []string) int {
	cmd := cmdFrom(cmdName, cmdVersion, args)

	plugins := cmd.plugins()
	cmd.process(plugins)

	return cmd.resolveExitValue()
}

func (cmd *cmdEnv) process(pSpecs []PluginSpec) {
	if cmd.noOp() || !cmd.ensurePluginDirs() {
		return
	}

	reporter := newConsoleReporter("    ", cmd.quietWanted)

	cmd.sync(pSpecs, reporter)

	reporter.finish(cmd.results)
}
