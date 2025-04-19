# pluggo

**WARNING**: this is beta software. Read on before trying it.

## What is pluggo?

Pluggo is a command line tool to manage plugins for Vim and Neovim.

## Why is pluggo?

I wanted a tool to install and update plugins for Vim and Neovim, but I did not
want the tool itself to be a Vim or Neovim plugin. So I wrote this in Go.

## What are pluggo's goals?

+ Pluggo should work for both Vim and Neovim.
+ Pluggo should be straightforward to use.
+ Pluggo should rely on the package conventions that Vim and Neovim share.
+ Pluggo should not try to do too much.

## Requirements and Installation

Pluggo requires a working installation of Git in the user's `PATH`. (This may
change if I switch from `exec` calls to a pure-Go library for Git operations.)

To install, run the following command.

```shell
go install https://pkg.go.dev/github.com/telemachus/pluggo/cmd/pluggo@latest
```

You can also [download the source code][pluggo] and install it locally by
running `make install`.

## Configuration and Use

Pluggo needs a configuration file with the following structure:

```json
{
    "dataDirs": [
        "HOME",
        ".local",
        "share",
        "nvim",
        "site",
        "pack",
        "pluggo"
    ],
    "plugins": [
        {
            "branch": "master",
            "name": "nvim-snippy",
            "url": "https://github.com/dcampos/nvim-snippy",
            "pin": true
        },
        {
            "branch": "master",
            "name": "vim-startuptime",
            "url": "https://github.com/dstein64/vim-startuptime",
            "opt": true
        }
    ]
}
```

Both `"dataDirs"` and `"plugins"` are required. The `"dataDirs"` array will be
combined into a top-level directory. (As a convenience, if the first item in the
`"dataDirs"` array is `"HOME"`, pluggo will replace this value with the user's
home directory.) Plugins are installed in start/ or opt/ subdirectories of that
directory. The `"plugins"` array should contain anonymous objects that specify
plugins. Each of the objects in that array must have a key and value for
`"branch"`, `"name"`, and `"url"`. In addition, each plugin object may specify
a boolean value for `"pin"` and `"opt"`.

Pluggo does not have subcommands. When the user runs pluggo, the tool will bring
the state of local plugins into sync with the configuration file. Pluggo will
remove plugins that are installed locally but are not in the configuration file.
Any plugin that does not have `"pin": true` in its configuration will be
updated. As needed, plugins will be moved between the start/ and opt/
subdirectories depending on the configuration file and their local state.

## Suggestions, Requests, and Problems

I intend to keep pluggo minimal, and I have designed it to fit my needs.
Nevertheless, if you have a suggestion or a request, please [file an
issue][issue]. If something doesn't work as expected or this README isn't clear,
please [file an issue][issue].

[pluggo]: https://github.com/telemachus/pluggo
[issue]: https://github.com/telemachus/pluggo/issues

(c) 2025 Peter Aronoff. BSD 3-Clause license; see [LICENSE.txt][license] for
details.

[license]: /LICENSE.txt
