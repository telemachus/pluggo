package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/telemachus/pluggo/internal/git"
)

func subCmdUpdate(cmd *cmdEnv) {
	if err := cmd.subCmdFrom(cmd.subCmdArgs); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"%s %s: %s\n",
			cmd.name,
			cmd.subCmdName,
			err,
		)

		return
	}

	plugins := cmd.plugins()
	cmd.update(plugins)
}

func (cmd *cmdEnv) update(plugins []Plugin) {
	if cmd.noOp() {
		return
	}

	ch := make(chan result)
	for _, plugin := range plugins {
		go cmd.updateOne(plugin, ch)
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

func (cmd *cmdEnv) updateOne(plugin Plugin, ch chan<- result) {
	pluginDir := plugin.fullPath(cmd.dataDir)

	digestBefore, err := git.HeadDigest(pluginDir)
	if err != nil {
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

	args := []string{"-C", pluginDir, "pull", "--recurse-submodules"}
	gitCmd := exec.Command("git", args...)

	err = gitCmd.Run()
	if err != nil {
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

	digestAfter, err := git.HeadDigest(pluginDir)
	if err != nil {
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

	if digestBefore.Equals(digestAfter) {
		ch <- result{
			isErr: false,
			msg:   fmt.Sprintf("%s: already up-to-date", plugin.Name),
		}

		return
	}

	ch <- result{
		isErr: false,
		msg:   fmt.Sprintf("%s: updated", plugin.Name),
	}
}
