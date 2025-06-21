// Pluggo uses git to manage plugins for vim and neovim.
package main

import (
	"os"

	"github.com/telemachus/pluggo/internal/cli"
)

func main() {
	os.Exit(cli.Pluggo(os.Args[1:]))
}
