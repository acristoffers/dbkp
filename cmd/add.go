package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [FILE|FOLDER...]",
	Short: "Adds files/folders or commands from the file system to the backup",
	Long: `Adds FILE|FOLDER or command to dbkp.toml.

    - dbkp add /path/to/file
      Creates an entry using the file or folder name, as
      [[Files]]
        Name = "file"
        Path = "/path/to/file"

    - dbkp add --command brew.leaves --backup "brew leaves" --restore "xargs brew install"
      Creates a command entry as
      [[Commands]]
        Name = "brew.leaves"
        Backup = "brew leaves"
        Restore = "xargs brew install"

    Symlinks in the --symlinks option create a symlink to the first element in
    the second element, that is, assuming $PATH as the Path entry,
      --symlinks .,~/.neovim,init.vim,~/.vimrc
    is the same as
      ln -s $PATH/. ~/.neovim
      ln -s $PATH/init.vim ~/.vimrc
    `,
	Run: func(cmd *cobra.Command, args []string) {
		only, err := cmd.Flags().GetStringSlice("only")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		exclude, err := cmd.Flags().GetStringSlice("exclude")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		symlinks, err := cmd.Flags().GetStringSlice("symlinks")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		command, err := cmd.Flags().GetString("command")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		backup, err := cmd.Flags().GetString("backup")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		restore, err := cmd.Flags().GetString("restore")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		if len(only) > 0 && len(args) > 1 {
			fmt.Fprintf(os.Stderr, "If --only is given, than only one path is allowed.\n")
			os.Exit(1)
		} else if len(exclude) > 0 && len(args) > 1 {
			fmt.Fprintf(os.Stderr, "If --exclude is given, than only one path is allowed.\n")
			os.Exit(1)
		} else if len(symlinks) > 0 && len(args) > 1 {
			fmt.Fprintf(os.Stderr, "If --symlinks is given, than only one path is allowed.\n")
			os.Exit(1)
		} else if len(exclude) > 0 && len(only) > 0 {
			fmt.Fprintf(os.Stderr, "--exclude and --only are mutually exclusive.\n")
			os.Exit(1)
		} else if len(command) == 0 && len(args) == 0 {
			fmt.Fprintf(os.Stderr, "The file path is required.\n")
			os.Exit(1)
		} else if len(command) != 0 && len(args) != 0 {
			fmt.Fprintf(os.Stderr, "No positional arguments are accepted when using --command.\n")
			os.Exit(1)
		} else if len(command) != 0 && (len(backup) == 0 || len(restore) == 0) {
			fmt.Fprintf(os.Stderr, "--backup and --restore are required when --command is passed.\n")
			os.Exit(1)
		}

		configPath, err := filepath.Abs("./dbkp.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get recipe path: %s\n", err)
			os.Exit(1)
		}

		config, err := dbkp.LoadRecipe(configPath)
		if err != nil {
			fmt.Printf("Cannot open file %s: %s\n", configPath, err)
			os.Exit(1)
		}

		homePath, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Cannot get home path: %s\n", err)
			os.Exit(1)
		}

		names := []string{}
		for _, file := range config.Files {
			names = append(names, file.Name)
		}
		for _, command := range config.Commands {
			names = append(names, command.Name)
		}

	path:
		for _, pathString := range args {
			path, err := filepath.Abs(pathString)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s: Error getting path: %s.\n", path, err)
				continue path
			}

			if _, err := os.Stat(path); err != nil {
				fmt.Fprintf(os.Stderr, "Skipping %s: File does not exist.\n", path)
				continue path
			}

			fileName := filepath.Base(path)
			for _, name := range names {
				if fileName == name {
					fmt.Fprintf(os.Stderr, "Skipping %s: Name already exists in the recipe.\n", name)
					continue path
				}
			}

			if strings.HasPrefix(path, homePath) {
				path = strings.Replace(path, homePath, "~", 1)
			}

			if strings.HasPrefix(fileName, ".") {
				fileName = strings.Replace(fileName, ".", "", 1)
			}

			file := dbkp.File{
				Name: fileName,
				Path: path,
			}

			if len(only) > 0 {
				file.Only = only
			}
			if len(exclude) > 0 {
				file.Exclude = exclude
			}
			if len(symlinks) > 0 {
				if len(symlinks)%2 != 0 {
					fmt.Fprintf(os.Stderr, "Symlinks requires an even number of arguments.\n")
					os.Exit(1)
				}

				var grouped [][2]string
				for i := 0; i < len(symlinks); i += 2 {
					pair := [2]string{symlinks[i], symlinks[i+1]}
					grouped = append(grouped, pair)
				}

				file.Symlinks = grouped
			}

			config.Files = append(config.Files, file)
			names = append(names, fileName)
		}

		if len(command) > 0 {
			config.Commands = append(config.Commands, dbkp.Command{
				Name:    command,
				Backup:  backup,
				Restore: restore,
			})
		}

		if err := config.WriteRecipe(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot open file %s: %s.\n", configPath, err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
	addCmd.Flags().StringSliceP("only", "o", []string{}, "Adds files/folders to the Only entry. Example: --only file1,file2,file3")
	addCmd.Flags().StringSliceP("exclude", "e", []string{}, "Adds files/folders to the Exclude entry. Example: --exclude file1,file2,file3")
	addCmd.Flags().StringSliceP("symlinks", "s", []string{}, "Adds symlinks. Example: --symlinks .,~/.neovim,init.vim,~/.vimrc")
	addCmd.Flags().StringP("command", "c", "", "Adds a command instead of a file. The name must be a valid file name: --command brew.leaves")
	addCmd.Flags().StringP("backup", "b", "", "The backup command. Its output will be saved to Command Name: --backup 'brew leaves'")
	addCmd.Flags().StringP("restore", "r", "", "The restore command. The Command Name file will be read and piped into this command's stdin: --backup 'xargs brew install'")
}
