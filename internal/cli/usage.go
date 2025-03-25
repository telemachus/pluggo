package cli

import "github.com/MakeNowJust/heredoc"

var cmdUsage = heredoc.Docf(`
	usage: pluggo [options] <subcommand> [options]

	Options
	        --config=FILE	Use FILE as config file (default ~/.pluggo.json)
	        --quiet		Print only error messages

	    -h, --help		Print this help and exit
	        --version	Print version and exit

	Subcommands
	    install		Install plugins using %[1]sgit clone --filter=blob:none%[1]s
	    update|up		Update plugins using %[1]sgit fetch%[1]s
	    sync		Run both update and install (in that order)

	For more information or to file a bug report, visit https://github.com/telemachus/pluggo
	`, "`")

var installUsage = heredoc.Docf(`
	usage: pluggo install [options]

	Install plugins using %[1]sgit clone --filter=blob:none%[1]s

	Options
	        --config=FILE	Use FILE as config file (default ~/.pluggo.json)
	        --quiet		Print only error messages

	    -h, --help		Print this help and exit
	        --version	Print version and exit

	For more information or to file a bug report, visit https://github.com/telemachus/pluggo
	`, "`")

var updateUsage = heredoc.Docf(`
	usage: pluggo update [options]

	Update Neovim plugins using %[1]sgit fetch%[1]s

	Options
	        --config=FILE	Use FILE as config file (default ~/.pluggo.json)
	        --quiet		Print only error messages

	    -h, --help		Print this help and exit
	        --version	Print version and exit

	For more information or to file a bug report, visit https://github.com/telemachus/pluggo
	`, "`")

var syncUsage = heredoc.Docf(`
	usage: pluggo sync [options]

	Install using %[1]sgit clone --filter=blob:none%[1]s and update them
	using %[1]sgit fetch%[1]s (in that order)

	Options
	        --config=FILE	Use FILE as config file (default ~/.pluggo.json)
	        --quiet		Print only error messages

	    -h, --help		Print this help and exit
	        --version	Print version and exit

	For more information or to file a bug report, visit https://github.com/telemachus/pluggo
	`, "`")
