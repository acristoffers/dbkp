package dbkp

import (
	"os"

	"github.com/BurntSushi/toml"
)

type File struct {
	Name     string
	Path     string
	Only     []string
	Exclude  []string
	Symlinks [][2]string
}

type Command struct {
	Name    string
	Backup  string
	Restore string
}

type Recipe struct {
	EncryptionSalt [2]string
	Files          []File
	Commands       []Command
}

func LoadRecipe(path string) (Recipe, error) {
	var config Recipe

	if _, err := os.Stat(path); err != nil {
		return config, err
	}

	_, err := toml.DecodeFile(path, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

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
	config := Recipe{
		Files:    files,
		Commands: commands,
	}
	return config.WriteRecipe(path)
}

func (config Recipe) WriteRecipe(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(file)
	return encoder.Encode(config)
}
