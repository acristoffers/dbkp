package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

func resolveRecipePathAndNames(args []string) (string, []string, error) {
	if len(args) == 0 {
		path, err := filepath.Abs("./dbkp.toml")
		return path, nil, err
	}

	if looksLikeTomlPath(args[0]) {
		path, err := filepath.Abs(args[0])
		if err != nil {
			return "", nil, err
		}
		return path, args[1:], nil
	}

	path, err := filepath.Abs("./dbkp.toml")
	return path, args, err
}

func looksLikeTomlPath(value string) bool {
	if strings.HasSuffix(value, ".toml") {
		return true
	}

	if _, err := os.Stat(value); err == nil {
		return true
	}

	return false
}
