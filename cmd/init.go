package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a dbkp project in the current directory",
	Long:  `Creates an empty recipe in the current folder.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := filepath.Abs("./dbkp.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get recipe path: %s\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "File already exists, not overriding\n")
			os.Exit(1)
		}

		ex, err := cmd.Flags().GetBool("example")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		encrypt, err := cmd.Flags().GetBool("encrypt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		if ex {
			if err := dbkp.WriteExampleRecipe(path); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot open file %s: %s\n", path, err)
				os.Exit(1)
			}
		} else {
			recipe := dbkp.Recipe{}
			if encrypt {
				iv := make([]byte, 12)
				salt := make([]byte, 32)
				rand.Read(iv)
				rand.Read(salt)
				recipe.EncryptionSalt = [2]string{hex.EncodeToString(salt), hex.EncodeToString(iv)}
			}
			if err := recipe.WriteRecipe(path); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot open file %s: %s\n", path, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().Bool("example", false, "Initializes with an example recipe instead")
	initCmd.Flags().Bool("encrypt", false, "Enables encryption for this backup")
}
