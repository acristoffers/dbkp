// Package dbkp implements dotfiles backup and restore.
//
// The dbkp package allows to easily backup and restore dotfiles to a folder. It
// will create a folder called dbkp containing all files and folders described
// in the Recipe struct. If encryption is used, instead of backing up to a
// folder, it will create a tarball with all files and encrypt the tarball using
// GCM-AES-256. Keys are derived from passwords using PBKDF2.
package dbkp

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Represents a File or Folder backup.
type File struct {
	Name     string      // Uniquely represents this File and is also the name of the file/folder inside the backup folder.
	Path     string      // The path to the file/folder to be backed up in the filesystem.
	Only     []string    // If Path is a folder, only backs up the items in Only, skipping all others.
	Exclude  []string    // If Path is a folder, excludes the items in Exclude from the backup.
	Symlinks [][2]string // After restoring, creates symlinks from /path/to/backup/Name/Symlinks[][0] into Symlinks[][1].
}

// Represents a pair of Backup and Restore commands.
// Backup and Restore are strings because they will both be executed as
// `sh -c 'CMD'`. The output of Backup is saved to a file
// /path/to/backup/Name, which is read into the input of Restore when
// restoring.
type Command struct {
	Name    string // Uniquely represents this File and is also the name of the file/folder inside the backup folder.
	Backup  string // The backup command to execute.
	Restore string // The restore command to execute.
}

// Identifies all elements of a backup, specifying what to backup/restore and
// whether the backup is encrypted. The pair of keys are regenerated every time
// a backup is done and the dbkp.toml file is created when the Tarball is
// written in the same folder as the Tarball itself.
type Recipe struct {
	EncryptionSalt [2]string // A pair of randon data. The first is for the key generator and the second for the encryption algorithm.
	Files          []File    // A list of File to backup/restore.
	Commands       []Command // A list of Command to backup/restore.
}

// Loads a recipe from a path. Just a convenience function to parse the TOML
// file.
func LoadRecipe(path string) (Recipe, error) {
	var recipe Recipe

	if _, err := os.Stat(path); err != nil {
		return recipe, err
	}

	_, err := toml.DecodeFile(path, &recipe)
	if err != nil {
		return recipe, err
	}

	return recipe, nil
}

// Write an recipe prefilled with some examples to path.
func WriteExampleRecipe(path string) error {
	files := []File{
		{
			Name: "doom",
			Path: "~/.config/doom",
		},
		{
			Path: "~/.config/nvim",
			Name: "neovim",
			Symlinks: [][2]string{
				{".", "~/.neovim"},
				{"init.vim", "~/.vimrc"},
			},
		},
		{
			Path:    "~/.config/fish",
			Name:    "fish",
			Exclude: []string{"fish_variables"},
		},
		{
			Path: "~/.ssh",
			Name: "ssh",
			Only: []string{"config", "key", "key.pub"},
		},
	}
	commands := []Command{
		{
			Name:    "brew",
			Backup:  "brew leaves",
			Restore: "xargs brew install",
		},
		{
			Name:    "flatpak",
			Backup:  "flatpak list --columns=ref --app | tail -n +1",
			Restore: "xargs flatpak install",
		},
	}
	recipe := Recipe{
		Files:    files,
		Commands: commands,
	}
	return recipe.WriteRecipe(path)
}

// Saves the recipe to a TOML file at path.
func (recipe Recipe) WriteRecipe(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(recipe)
}
