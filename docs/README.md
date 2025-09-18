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
go install github.com/telemachus/pluggo@latest
```

You can also [download the source code][pluggo] and install it locally by
running `make install`.

## Configuration and Use

Pluggo needs a configuration file with the following structure:

```json
{
    "dataDir": [
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
            "pinned": true
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

### Re `"dataDir"`

+ The `"dataDir"` array will be combined with an OS-specific path separator.
  E.g., `${HOME}/.local/share/nvim/site/pack/pluggo` on Linux and macOS for the
  JSON above.
+ Plugins are installed in start/ or opt/ subdirectories of that directory. See
  below for the `"opt"` item in `"plugins"`.
+ As a convenience, if the first item in the `"dataDirs"` array is `"HOME"`,
  pluggo will replace this value with the user's home directory.

### Re `"plugins"`

+ The `"plugins"` array should contain anonymous objects that specify plugins.
+ Each of the objects in that array must have a key and value for `"branch"`,
  `"name"`, and `"url"`. The `"url"` value must be a full URL for a git
  download. There is no special treatment of GitHub repos. (In other words,
  pluggo will not automagically translate the URL `"name/plugin"` as
  `https://github.com/name/plugin`.)
+ Each plugin object may specify a boolean value for `"pinned"` and `"opt"`.
+ If `"pinned"` is true, the plugin will not be updated.
+ If `"opt"` is true, the plugin will be installed in an `opt` subdirectory of
  `"dataDir"`. If `"opt"` is not specified or false, plugins will be installed
  in a `start` subdirectory.

Pluggo does not have subcommands. When the user runs pluggo, the tool will bring
the state of local plugins into sync with the configuration file. Pluggo will
remove plugins that are installed locally but are not in the configuration file.
Any plugin that does not have `"pinned": true` in its configuration will be
updated. As needed, plugins will be moved between the start/ and opt/
subdirectories depending on the configuration file and their local state.

## Tips

By default, pluggo will look for a configuration file at `${HOME}/.pluggo.json`.
If you want to use pluggo only for Vim or Neovim, you should go ahead and use
that file for your configuration. However, if you want to use pluggo for both
Vim and Neovim, you can create different configuration files and then run pluggo
with the `-config` flag.

```shell
pluggo -config="${HOME}/.config/nvim/nvim-pluggo.json"
pluggo -config="${HOME}/.vim/vim-pluggo.json"
```

Pluggo does not update help tags. If you want to update those for Neovim, the
following works well.

```shell
pluggo -config=${HOME}/.config/nvim/nvim-pluggo.json
nvim --headless -c 'helptags ALL | quit'
```

To update help tags in Vim, the following works though it's probably overkill.

```vim
" Run `:helptags ALL` every time that Vim starts.
autocmd VimEnter * :helptags ALL
```
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
