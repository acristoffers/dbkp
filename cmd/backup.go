package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup [dbkp.toml] [name ...]",
	Short: "Executes the backup in dbkp.toml.",
	Long:  `Executes the backup in dbkp.toml.`,
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
		encrypt, err := cmd.Flags().GetBool("encrypt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

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
		if encrypt || len(recipe.EncryptionSalt) > 0 && len(recipe.EncryptionSalt[0]) > 0 {
			password1, err := dbkp.AskForPassword()
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}

			fmt.Println("Type again, for confirmation.")

			password2, err := dbkp.AskForPassword()
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}

			if !bytes.Equal(password1, password2) {
				fmt.Fprintln(os.Stderr, "Passwords do not match")
				os.Exit(1)
			}

			password = password1
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
			if err := dbkp.BackupSelected(path, recipe, password, channel, names); err != nil {
				bar.Clear()
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}
		}()

		for c := range channel {
			bar.ChangeMax64(int64(c.Total))
			bar.Describe(fmt.Sprintf("Backing up %s", c.Name))
			bar.Set64(int64(c.Count))
		}

		bar.Clear()
	},
}

func init() {
	RootCmd.AddCommand(backupCmd)
	backupCmd.Flags().BoolP("encrypt", "e", false, "Enables encryption for this backup, if it is not enabled already")
}
