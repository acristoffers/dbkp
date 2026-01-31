# dbkp — dotfiles backup and restore

`dbkp` is a small CLI that snapshots your dotfiles into a folder you control and restores them
later, optionally encrypted. It keeps a simple `dbkp.toml` recipe, can back up files and folders,
and can also capture command output and replay it during restore. You can put the backup folder in
git, sync it with a drive, or use any storage you prefer.

## What it is

- A declarative recipe (`dbkp.toml`) for dotfiles.
- A backup/restore runner for files, folders, and command output.
- A lightweight tool that doesn’t impose how you store or version the backup.

## What it is not

- Not a dotfiles manager or symlink farm like GNU Stow.
- Not a system-level deployment tool like Nix.
- Not a VCS or storage backend by itself.

If you need package management, system configuration, or large-scale machine provisioning, you
probably want Nix. If you only need symlink management, you probably want GNU Stow. dbkp focuses on
backing up and restoring dotfiles and related command state.

I personally pair it with a git repository to add the version control layer. You can see my
[dotfiles](https://github.com/acristoffers/dotfiles) for an example.

## Install

### Nix (flakes)

```bash
nix install github:acristoffers/dbkp
```

### Go (build from source)

```bash
go install github.com/acristoffers/dbkp@latest
```

### Prebuilt binaries

Download a binary from the project’s GitHub Releases page and place it on your
`PATH`.

## Usage

### Initialize a backup folder

```bash
mkdir -p ~/dotfiles-backup
cd ~/dotfiles-backup
dbkp init
```

Enable encryption when creating the recipe:

```bash
dbkp init --encrypt
```

### Add files and folders

```bash
dbkp add ~/.config/fish
dbkp add ~/bin
```

Exclude entries (Go regex, matched against relative paths):

```bash
dbkp add ~/bin --exclude 'cache$',tmp
```

Only include specific entries inside the added path:

```bash
dbkp add ~/.config --only fish,alacritty
```

Add symlink mappings (source inside backup → target path):

```bash
dbkp add ~/.config/neovim --symlinks .,~/.neovim,init.vim,~/.vimrc
# The following symlinks will be created by the restore command:
# ~/.neovim pointing to ~/.config/neovim
# ~/.vimrc pointing to ~/.config/neovim/init.vim
```

### Add commands

Save the output of a command during backup and feed it to another command during restore:

```bash
dbkp add --command brew.leaves --backup "brew leaves" --restore "xargs brew install"
```

You can pipe and use `xargs` as needed:

```bash
dbkp add --command flatpak --backup "flatpak list --columns=ref --app | tail -n +1" --restore "xargs flatpak install -y --noninteractive --or-update"
```

The output of the backup command is saved to a file (i.e. `brew.leaves` or `flatpak`) and the file's
contents are piped into the restore command, so the last example is the same as:

```bash
# backup
flatpak list --columns=ref --app | tail -n +1 > flatpak
# restore
cat flatpak | xargs flatpak install -y --noninteractive --or-update
```

### Run backup and restore

```bash
dbkp backup
dbkp restore
```

Force encryption on an existing, unencrypted recipe:

```bash
dbkp backup --encrypt
```

### Remove entries

The argument is the `Name` inside `dbkp.toml` you want to remove:

```bash
dbkp remove fish
dbkp remove brew.leaves
```

### List entries

```bash
dbkp list
dbkp list --machine
```

## License

Mozilla Public License 2.0. See `LICENSE`.
