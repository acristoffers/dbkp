package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Args:  cobra.MinimumNArgs(1),
	Short: "Removes and entry from the backup recipe",
	Long:  `Removes NAMEs from dbkp.toml.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		suggestions := []string{}

		path, err := filepath.Abs("./dbkp.toml")
		if err != nil {
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}

		recipe, err := dbkp.LoadRecipe(path)
		if err != nil {
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}

		for _, file := range recipe.Files {
			if strings.HasPrefix(file.Name, toComplete) && !slices.Contains(args, file.Name) {
				suggestions = append(suggestions, file.Name)
			}
		}
		for _, command := range recipe.Commands {
			if strings.HasPrefix(command.Name, toComplete) && !slices.Contains(args, command.Name) {
				suggestions = append(suggestions, command.Name)
			}
		}

		return suggestions, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		path, err := filepath.Abs("./dbkp.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get recipe path: %s\n", err)
			os.Exit(1)
		}

		recipe, err := dbkp.LoadRecipe(path)
		if err != nil {
			fmt.Printf("Cannot open file %s: %s\n", "./dbkp.toml", err)
			os.Exit(1)
		}

		keepFiles := []dbkp.File{}
	fileLoop:
		for _, file := range recipe.Files {
			for _, name := range args {
				if file.Name == name {
					continue fileLoop
				}
			}
			keepFiles = append(keepFiles, file)
		}
		recipe.Files = keepFiles

		keepCmd := []dbkp.Command{}
	commandLoop:
		for _, command := range recipe.Commands {
			for _, name := range args {
				if command.Name == name {
					continue commandLoop
				}
			}
			keepCmd = append(keepCmd, command)
		}
		recipe.Commands = keepCmd

		if err := recipe.WriteRecipe(path); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot open file %s: %s\n", path, err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
