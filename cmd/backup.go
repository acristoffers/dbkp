package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup [/path/to/dbkp.toml]",
	Short: "Executes the backup in dbkp.toml.",
	Long:  `Executes the backup in dbkp.toml.`,
	Run: func(cmd *cobra.Command, args []string) {
		encrypt, err := cmd.Flags().GetBool("encrypt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		path := ""
		recipePath := ""

		if len(args) == 1 {
			recipePath, err = filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}

			path = filepath.Dir(recipePath)
		} else {
			path, err = filepath.Abs(".")
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}

			recipePath, err = filepath.Abs(filepath.Join(path, "dbkp.toml"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}
		}

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
			if err := dbkp.Backup(path, recipe, password, channel); err != nil {
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
