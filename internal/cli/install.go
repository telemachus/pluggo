package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func subCmdInstall(cmd *cmdEnv) {
	if err := cmd.subCmdFrom(cmd.subCmdArgs); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
		return
	}

	plugins := cmd.plugins()
	cmd.install(plugins)
}

func (cmd *cmdEnv) install(plugins []Plugin) {
	if cmd.noOp() {
		return
	}

	// Strictly speaking, this is unnecessary since git itself will create
	// directories as needed. However, if permissions prevent creating
	// a parent directory, then the tool will uselessly exec git and try to
	// clone plugin after plugin. This way, I bail out early.
	err := os.MkdirAll(cmd.dataDir, os.ModePerm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
		cmd.exitVal = exitFailure
		return
	}

	ch := make(chan result)
	for _, plugin := range plugins {
		go cmd.installOne(plugin, ch)
	}
	for range plugins {
		res := <-ch
		switch cmd.quietWanted {
		case true:
			res.publishError()
		default:
			res.publish()
		}
	}
}

func (cmd *cmdEnv) installOne(plugin Plugin, ch chan<- result) {
	// Normally, it is a bad idea to check whether a directory exists
	// before trying an operation. However, this case is an exception.
	// git clone will return an error if a repo with that name in that
	// location already exists. But for the purpose of this command, there
	// is no error. If a directory with the repo's name exists, I simply
	// want to skip that repo.
	pluginFullPath := plugin.fullPath(cmd.dataDir)
	if _, err := os.Stat(pluginFullPath); err == nil {
		ch <- result{
			isErr: false,
			msg: fmt.Sprintf(
				"%s: %s already present in %s",
				cmd.name,
				plugin.Name,
				cmd.prettyPath(cmd.dataDir),
			),
		}
		return
	}

	args := plugin.installArgs(pluginFullPath)
	gitCmd := exec.Command("git", args...)

	err := gitCmd.Run()
	if err != nil {
		cmd.exitVal = exitFailure
		ch <- result{
			isErr: true,
			msg: fmt.Sprintf(
				"%s %s: %s: %s",
				cmd.name,
				cmd.subCmdName,
				plugin.Name,
				err,
			),
		}
		return
	}

	ch <- result{
		isErr: false,
		msg:   fmt.Sprintf("%s: %s installed", "pluggo", plugin.Name),
	}
}
