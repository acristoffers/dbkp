package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [dbkp.toml] [name ...]",
	Short: "Restores the backup in dbkp.toml.",
	Long:  `Restores the backup in dbkp.toml.`,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		suggestions := []string{}

		recipePath, names, err := resolveRecipePathAndNames(args)
		if err != nil {
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}

		recipe, err := dbkp.LoadRecipe(recipePath)
		if err != nil {
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		}

		for _, file := range recipe.Files {
			if strings.HasPrefix(file.Name, toComplete) && !slices.Contains(names, file.Name) {
				suggestions = append(suggestions, file.Name)
			}
		}

		for _, command := range recipe.Commands {
			if strings.HasPrefix(command.Name, toComplete) && !slices.Contains(names, command.Name) {
				suggestions = append(suggestions, command.Name)
			}
		}

		return suggestions, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		recipePath, names, err := resolveRecipePathAndNames(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		path := filepath.Dir(recipePath)

		recipe, err := dbkp.LoadRecipe(recipePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		password := []byte{}
		if len(recipe.EncryptionSalt) > 0 && len(recipe.EncryptionSalt[0]) > 0 {
			password, err = dbkp.AskForPassword()
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}
		} else {
			password = nil
		}

		bar := progressbar.NewOptions(100,
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionThrottle(0),
			progressbar.OptionShowCount(),
			progressbar.OptionFullWidth(),
			progressbar.OptionSetRenderBlankState(true))

		channel := make(chan dbkp.ProgressReport)

		go func() {
			if err := dbkp.RestoreSelected(path, recipe, password, channel, names); err != nil {
				bar.Clear()
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}
		}()

		for c := range channel {
			bar.ChangeMax64(int64(c.Total))
			bar.Describe(fmt.Sprintf("Restoring %s", c.Name))
			bar.Set64(int64(c.Count))
		}

		bar.Clear()
	},
}

func init() {
	RootCmd.AddCommand(restoreCmd)
}
