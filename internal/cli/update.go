package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/telemachus/pluggo/internal/git"
)

func subCmdUpdate(cmd *cmdEnv) {
	if err := cmd.subCmdFrom(cmd.subCmdArgs); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
		return
	}
	rs := cmd.repos()
	cmd.update(rs)
}

func (cmd *cmdEnv) update(rs []Repo) {
	if cmd.noOp() {
		return
	}

	ch := make(chan result)
	for _, r := range rs {
		go cmd.updateOne(r, ch)
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

func (cmd *cmdEnv) updateOne(r Repo, ch chan<- result) {
	rDir := filepath.Join(cmd.dataDir, r.Name)
	fhBefore, err := git.NewFetchHead(filepath.Join(rDir, "FETCH_HEAD"))
	if err != nil {
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("%s %s: %s: %s", cmd.name, cmd.subCmdName, r.Name, err),
		}
		return
	}

	args := []string{"remote", "update"}
	gitCmd := exec.Command("git", args...)
	noGitPrompt := "GIT_TERMINAL_PROMPT=0"
	env := append(os.Environ(), noGitPrompt)
	gitCmd.Env = env
	gitCmd.Dir = rDir

	err = gitCmd.Run()
	if err != nil {
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("%s %s: %s: %s", cmd.name, cmd.subCmdName, r.Name, err),
		}
		return
	}

	fhAfter, err := git.NewFetchHead(filepath.Join(rDir, "FETCH_HEAD"))
	if err != nil {
		ch <- result{
			isErr: true,
			msg:   fmt.Sprintf("%s %s: %s: %s", cmd.name, cmd.subCmdName, r.Name, err),
		}
		return
	}

	if fhBefore.Equals(fhAfter) {
		ch <- result{
			isErr: false,
			msg:   fmt.Sprintf("%s: already up-to-date", r.Name),
		}
		return
	}

	ch <- result{
		isErr: false,
		msg:   fmt.Sprintf("%s: updated", r.Name),
	}
}
