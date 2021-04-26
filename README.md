# dbkp - dotfiles backup

dbkp simply backups and restores dotfiles using `rsync -La`. You can use it with
any version control or backup strategy you want.

## Instalation

Install with pip:

`pip3 install dbkp`

## Usage

Put the file dbkp.json on a folder and configure it as you want. Then either run
`dbkp backup` from the folder or pass the path to the configuration file as
`dbkp backup /path/to/dbkp.json`. A folder named dotfiles in the same folder as
the configuration file will be created and all dotfiles will be synced inside
it. Use `dbkp restore` in the same way to rsync the files from the dotfiles
folder into their places.

## Configuration example

The configuration is a list of files to backup, or objects that adds some
features to the specific dotfile.

A string will backup the folder/file. This will backup the file `~/.vimrc`:

```json
["~/.vimrc"]
```

It is the same as

```json
[
  {
    "path": "~/.vimrc",
    "alias": "vimrc"
  }
]
```

`alias` is the name the file/folder will have inside the dotfiles folder. By
default it is the name of the file/folder without a leading dot.

It is also possible to exclude subfiles/subfolders from a folder. The complete
folder will be synced and then the files/folders will be deleted. You can
specify both `only` and `exclude`, but `exclude` will be ignored in this case.
`only` will remove all files but the ones listed and `exclude` will only exclude
the ones listed.

```json
[
  {
    "path": "~/.config/fish",
    "only": ["functions"],
    "exclude": ["completions"]
  }
]
```

The `links` options allows to create symlinks after restoring. It is a list of
either strings or lists of 2 string elements. If the element is a string, then a
symlink will be created pointing to `path`. If it a list of 2 strings, the
second is the symlink and will point to `path`/`first element`. In the example,
`~/.vimrc` will point to `~/.config/nvim/init.nvim` and `~/neovim` will point to
`~/.config/nvim`.

```json
[
  {
    "path": "~/.config/nvim",
    "links": ["~/neovim", ["init.vim", "~/.vimrc"]]
  }
]
```

It is also possible to run commands to do the backup/restore.

```json
[
  {
    "backup": "brew leaves",
    "restore": "xargs brew install",
    "alias": "brew.leaves"
  }
]
```

This is the same as backing up with

```sh
brew leaves > brew.leaves
```

and then restoring with

```sh
cat brew.leaves | xargs brew install
```

The current working directory is changed to the folder containing the
configuration file before executing anything, so if you want to specify files in
the command line, remember that: you need to quote file paths if they contain
spaces and your `alias` file is in `dotfiles/:alias"`.

## LICENSE

The MIT License (MIT)

Copyright (c) 2021 Álan Crístoffer e Sousa

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
