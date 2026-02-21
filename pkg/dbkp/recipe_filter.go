package dbkp

import (
	"fmt"
)

func filterRecipeByNames(recipe Recipe, names []string) (Recipe, error) {
	if len(names) == 0 {
		return recipe, nil
	}

	known := map[string]struct{}{}
	for _, file := range recipe.Files {
		known[file.Name] = struct{}{}
	}

	for _, command := range recipe.Commands {
		known[command.Name] = struct{}{}
	}

	for _, name := range names {
		if _, ok := known[name]; !ok {
			return Recipe{}, fmt.Errorf("unknown entry name: %s", name)
		}
	}

	selected := recipe
	selected.Files = nil
	selected.Commands = nil

	selectedNames := map[string]struct{}{}
	for _, name := range names {
		selectedNames[name] = struct{}{}
	}

	for _, file := range recipe.Files {
		if _, ok := selectedNames[file.Name]; ok {
			selected.Files = append(selected.Files, file)
		}
	}

	for _, command := range recipe.Commands {
		if _, ok := selectedNames[command.Name]; ok {
			selected.Commands = append(selected.Commands, command)
		}
	}

	return selected, nil
}
