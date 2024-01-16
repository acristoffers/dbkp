package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restores the backup in dbkp.toml.",
	Long:  `Restores the backup in dbkp.toml.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := filepath.Abs(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		if err := dbkp.Restore(path); err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(restoreCmd)
}
