// Pluggo manages Vim or Neovim plugins.
package main

import (
	"os"

	"github.com/telemachus/pluggo/internal/cli"
)

func main() {
	os.Exit(cli.Pluggo(os.Args[1:]))
}
