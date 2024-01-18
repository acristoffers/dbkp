# dbkp - dotfiles backup

dbkp simply backups and restores dotfiles. You can use it with any version
control or backup strategy you want.

## Instalation

With nix flakes: `nix install github:acristoffers/dbkp`

With go: `go install github.com/acristoffers/dbkp@latest`

## Usage

# Create the backup folder and configuration file

Create a folder where you want to backup to. I put this folder into git for
version control. Initialise the backup with `dbkp init` or, if you want
encryption (GCM-AES-256) `dbkp init --encrypt`.

It will create a `dbkp.toml` (with some random data if you passed `--encrypt`).

# Backing up/restoring files

Now, add some files with:

```bash
dbkp add ~/bin
```

which adds the folder `~/bin` to the backup (but does not backup yet). However,
I have a folder inside `~/bin` that I don't want to backup, so I run this
instead:

```bash
dbkp add ~/bin -e tree-sitter-grammars
```

which is going to skip `~/bin/tree-sitter-grammars`. There is also an
`--only|-o` option that only picks the given files/folders. To pass more than
one, separate their names by commas.

To backup, run `dbkp backup`. If you want to encrypt a previously unencrypted
backup, pass the `-e` flag. If you want to stop encrypting files, edit
`dbkp.toml` and make sure that the line of `EncryptionSalt` reads:

```toml
EncryptionSalt = ["", ""]
```

To restore, run `dbkp restore`.

# Backing up/restoring with commands

dbkp also supports backup/restore through commands. It will execute the Backup
command and save its `stdout` during backup, and will read the saved content and
feed it to Restore's `stdin`. The commands will be executed by `sh -c`.

To add a backup command:

```bash
dbkp add --command gnome-settings --backup "dconf dump /" --restore "dconf load /"
```

If needed, you can use `sh` to pipe things and `xargs` to turn `stdin` into
arguments:

```bash
dbkp add --command flatpak --backup "flatpak list --columns=ref --app | tail -n +1" --restore "xargs flatpak install -y --noninteractive --or-update"
```
