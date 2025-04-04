package cli

import "github.com/MakeNowJust/heredoc"

var cmdUsage = heredoc.Docf(`
    usage: pluggo [options]

    Manages Vim/Neovim plugins by synchronizing them with your configuration file.
    
    Options
            --config=FILE    Use FILE as config file (default ~/.pluggo.json)
            --quiet          Print only error messages

        -h, --help           Print this help and exit
            --version        Print version and exit

    For more information or to file a bug report, visit https://github.com/telemachus/pluggo
    `, "`")
