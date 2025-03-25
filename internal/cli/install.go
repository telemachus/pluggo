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

	rs := cmd.repos()
	cmd.install(rs)
}

func (cmd *cmdEnv) install(rs []Repo) {
	if cmd.noOp() {
		return
	}

	// err := os.MkdirAll(cmd.dataDir, os.ModePerm)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
	// 	cmd.exitVal = exitFailure
	// 	return
	// }

	ch := make(chan result)
	for _, r := range rs {
		go cmd.installOne(r, ch)
	}
	for range rs {
		res := <-ch
		switch cmd.quietWanted {
		case true:
			res.publishError()
		default:
			res.publish()
		}
	}
}

func (cmd *cmdEnv) installOne(r Repo, ch chan<- result) {
	// Normally, it is a bad idea to check whether a directory exists
	// before trying an operation. However, this case is an exception.
	// git clone will retrun an error if a repo with that name in that
	// location already exists. But for the purpose of this command, there
	// is no error. If a directory with the repo's name exists, I simply
	// want to skip that repo.
	repoFullPath := r.fullPath(cmd.dataDir)
	if _, err := os.Stat(repoFullPath); err == nil {
		ch <- result{
			isErr: false,
			msg:   fmt.Sprintf("%s: %s is already installed", cmd.name, cmd.prettyPath(repoFullPath)),
		}
		return
	}

	args := r.installArgs(repoFullPath)
	gitCmd := exec.Command("git", args...)
	// noGitPrompt := "GIT_TERMINAL_PROMPT=0"
	// env := append(os.Environ(), noGitPrompt)
	// gitCmd.Env = env

	err := gitCmd.Run()
	if err != nil {
		cmd.exitVal = exitFailure
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("%s %s: %s: %s", cmd.name, cmd.subCmdName, r.Name, err),
		}
		return
	}

	ch <- result{
		isErr: false,
		msg:   fmt.Sprintf("%s: %s installed", "pluggo", r.Name),
	}
}
