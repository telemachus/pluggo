package cli

import (
	"fmt"
	"os"
)

func subCmdSync(cmd *cmdEnv) {
	if err := cmd.subCmdFrom(cmd.subCmdArgs); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s: %s\n", cmd.name, cmd.subCmdName, err)
		return
	}
	rs := cmd.repos()
	cmd.update(rs)
	cmd.install(rs)
}
