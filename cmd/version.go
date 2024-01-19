package cmd

import (
	"fmt"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version and exits",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version %s", dbkp.Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
