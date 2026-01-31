package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/acristoffers/dbkp/pkg/dbkp"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var listCmd = &cobra.Command{
	Use:   "list [/path/to/dbkp.toml]",
	Short: "Lists entries in dbkp.toml.",
	Long:  "Lists entries in dbkp.toml.",
	Run: func(cmd *cobra.Command, args []string) {
		machine, err := cmd.Flags().GetBool("machine")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not parse options: %s\n", err)
			os.Exit(1)
		}

		var recipePath string

		if len(args) == 1 {
			recipePath, err = filepath.Abs(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}
		} else if len(args) == 0 {
			recipePath, err = filepath.Abs(filepath.Join(".", "dbkp.toml"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Usage: dbkp list [/path/to/dbkp.toml]")
			os.Exit(1)
		}

		recipe, err := dbkp.LoadRecipe(recipePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred: %s\n", err)
			os.Exit(1)
		}

		if len(recipe.EncryptionSalt) > 0 && len(recipe.EncryptionSalt[0]) > 0 {
			fmt.Println("Encryption enabled")
		} else {
			fmt.Println("Encryption disabled")
		}

		if machine {
			for _, file := range recipe.Files {
				fmt.Println(formatFileMachine(file))
			}

			for _, command := range recipe.Commands {
				fmt.Println(formatCommandMachine(command))
			}
			return
		}

		renderer := lipgloss.NewRenderer(os.Stdout)
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			renderer.SetColorProfile(termenv.Ascii)
		}
		lipgloss.SetDefaultRenderer(renderer)

		filesTable := renderFilesTable(renderer, recipe.Files)
		if filesTable != "" {
			fmt.Println(filesTable)
		}

		commandsTable := renderCommandsTable(renderer, recipe.Commands)
		if commandsTable != "" {
			if filesTable != "" {
				fmt.Println()
			}
			fmt.Println(commandsTable)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("machine", "m", false, "Machine-readable output using tab separators")
}

func formatFileMachine(file dbkp.File) string {
	fields := []string{file.Name, file.Path}

	if len(file.Only) > 0 {
		fields = append(fields, fmt.Sprintf("Only: %s", strings.Join(file.Only, " ")))
	} else if len(file.Exclude) > 0 {
		fields = append(fields, fmt.Sprintf("Excluding: %s", strings.Join(file.Exclude, " ")))
	}

	if len(file.Symlinks) > 0 {
		entries := make([]string, 0, len(file.Symlinks))
		for _, pair := range file.Symlinks {
			entries = append(entries, fmt.Sprintf("%s -> %s", pair[0], pair[1]))
		}
		fields = append(fields, fmt.Sprintf("Symlinks: %s", strings.Join(entries, ", ")))
	}

	return strings.Join(fields, "\t")
}

func formatCommandMachine(command dbkp.Command) string {
	return strings.Join([]string{command.Name, command.Backup, command.Restore}, "\t")
}

func renderFilesTable(renderer *lipgloss.Renderer, files []dbkp.File) string {
	if len(files) == 0 {
		return ""
	}

	rows := make([][]string, 0, len(files))

	for _, file := range files {
		only := ""
		if len(file.Only) > 0 {
			only = strings.Join(file.Only, " ")
		}

		exclude := ""
		if len(file.Exclude) > 0 {
			exclude = strings.Join(file.Exclude, " ")
		}

		symlinks := ""
		if len(file.Symlinks) > 0 {
			entries := make([]string, 0, len(file.Symlinks))
			for _, pair := range file.Symlinks {
				entries = append(entries, fmt.Sprintf("%s -> %s", pair[0], pair[1]))
			}
			symlinks = strings.Join(entries, ", ")
		}

		rows = append(rows, []string{file.Name, file.Path, only, exclude, symlinks})
	}

	return renderTable(renderer, "Files", []string{"Name", "Path", "Only", "Exclude", "Symlinks"}, rows)
}

func renderCommandsTable(renderer *lipgloss.Renderer, commands []dbkp.Command) string {
	if len(commands) == 0 {
		return ""
	}

	rows := make([][]string, 0, len(commands))

	for _, command := range commands {
		rows = append(rows, []string{command.Name, command.Backup, command.Restore})
	}

	return renderTable(renderer, "Commands", []string{"Name", "Backup", "Restore"}, rows)
}

func renderTable(renderer *lipgloss.Renderer, title string, headers []string, rows [][]string) string {
	headerStyle := renderer.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "252"})

	titleStyle := renderer.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "247"})

	borderStyle := renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "240"})

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, _ int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return renderer.NewStyle()
		})

	return strings.Join([]string{titleStyle.Render(title), t.Render()}, "\n")
}
