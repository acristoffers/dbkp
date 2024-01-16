package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "dbkp",
	Short: "A simple dotfiles backup solution",
	Long: `dbkp allows you to backup and restore dotfiles with ease.

    Create the configuration file to specify what files and folders you want to
    keep track of, then run "dbkp backup" to backup and "dbkp restore" to
    restore into a subfolder where the configuration file is. That simple. Pair
    it with git for version control.

    For example to backup fish and you bin folder into Dropbox:
      mkdir ~/Documents/Dropbox/dotfiles
      cd ~/Documents/Dropbox/dotfiles
      dbkp init
      dbkp add ~/.config/fish
      dbkp add ~/bin
      dbkp backup
    `,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
}
