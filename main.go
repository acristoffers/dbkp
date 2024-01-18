/*
dbkp allows you to backup and restore dotfiles with ease.

Create the configuration file to specify what files and folders you want to keep
track of, then run "dbkp backup" to backup and "dbkp restore" to restore into a
subfolder where the configuration file is. That simple. Pair it with git for
version control.

For example to backup fish and you bin folder into Dropbox:

      mkdir ~/Documents/Dropbox/dotfiles
      cd ~/Documents/Dropbox/dotfiles
      dbkp init
      dbkp add ~/.config/fish
      dbkp add ~/bin
      dbkp backup

Usage:
  dbkp [command]

Available Commands:
  add         Adds files/folders or commands from the file system to the backup
  backup      Executes the backup in dbkp.toml.
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Creates a dbkp project in the current directory
  remove      Removes and entry from the backup recipe
  restore     Restores the backup in dbkp.toml.
  version     Shows version and exits

Flags:
  -h, --help   help for dbkp

Use "dbkp [command] --help" for more information about a command.
*/
package main

import "github.com/acristoffers/dbkp/cmd"

func main() {
	cmd.Execute()
}
