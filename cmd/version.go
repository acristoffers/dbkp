package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows version and exits",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version 2.1.1")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
