package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Executes the backup in dbkp.toml.",
	Long:  `Executes the backup in dbkp.toml.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := filepath.Abs(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		encrypt, err := cmd.Flags().GetBool("encrypt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		configPath, err := filepath.Abs(filepath.Join(path, "dbkp.toml"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		config, err := dbkp.LoadRecipe(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		password := []byte{}
		if encrypt || len(config.EncryptionSalt) > 0 && len(config.EncryptionSalt[0]) > 0 {
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
		report := func(count int, total int, name string) {
			bar.ChangeMax(total)
			bar.Describe(fmt.Sprintf("Backing up %s", name))
			bar.Set(count)
		}

		if err := dbkp.Backup(path, config, password, report); err != nil {
			bar.Clear()
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		bar.Clear()
	},
}

func init() {
	RootCmd.AddCommand(backupCmd)
	backupCmd.Flags().BoolP("encrypt", "e", false, "Enables encryption for this backup, if it is not enabled already")
}
