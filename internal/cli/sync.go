package cli

import (
	"fmt"
	"os"
)

func subCmdSync(cmd *cmdEnv) {
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
	cmd.install(plugins)
}
