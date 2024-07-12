package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [/path/to/dbkp.toml]",
	Short: "Restores the backup in dbkp.toml.",
	Long:  `Restores the backup in dbkp.toml.`,
	Run: func(cmd *cobra.Command, args []string) {
		var path string
		var recipePath string
		var err error

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
			if err := dbkp.Restore(path, recipe, password, channel); err != nil {
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
