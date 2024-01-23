package dbkp

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Restore(path string, recipe Recipe, password []byte, pr ProgressReport) error {
	backupPath, err := filepath.Abs(filepath.Join(path, "dbkp"))
	if err != nil {
		return err
	}

	if password != nil {
		return restoreEncrypt(backupPath, recipe, password, pr)
	}

	return restorePlain(backupPath, recipe, pr)
}

func restorePlain(backupFolder string, recipe Recipe, pr ProgressReport) error {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stepsLen := len(recipe.Files) + len(recipe.Commands)

	for i, file := range recipe.Files {
		path := file.Path
		if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homePath, path[2:])
		}

		if pr != nil {
			pr(i, stepsLen, file.Name)
		}
		backupPath := filepath.Join(backupFolder, file.Name)

		_, err := os.Lstat(backupPath)
		if err != nil {
			return err
		}

		if err := os.RemoveAll(path); err != nil {
			return err
		}

		if err := copyFileOrFolder(backupPath, path, file); err != nil {
			return err
		}
	}

	shellPath, err := exec.LookPath("sh")
	if err != nil {
		return err
	}

	for i, command := range recipe.Commands {
		if pr != nil {
			pr(i+len(recipe.Files), stepsLen, command.Name)
		}

		backupPath := filepath.Join(backupFolder, command.Name)
		data, err := readFile(backupPath)
		if err != nil {
			return err
		}

		var stderr bytes.Buffer
		stdin := bytes.NewBuffer(data)

		if err := executeCommandInShell(shellPath, command.Restore, stdin, nil, &stderr); err != nil {
			return errors.Join(err, errors.New(fmt.Sprintf("Command failed with error\n: %s", stderr.String())))
		}
	}

	return nil
}

func restoreEncrypt(backupFile string, recipe Recipe, password []byte, pr ProgressReport) error {
	tar, err := loadTarball(backupFile, password, recipe)
	if err != nil {
		return err
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stepsLen := len(recipe.Files) + len(recipe.Commands)

	for i, file := range recipe.Files {
		path := file.Path
		if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homePath, path[2:])
		}

		if pr != nil {
			pr(i+1, stepsLen, file.Name)
		}

		subtarbuffer, err := tar.readFile(file.Name)
		subtar := Tarball{Buffer: subtarbuffer}
		if err != nil {
			return err
		}

		if err := os.RemoveAll(path); err != nil {
			return err
		}

		subtar.unpackInto(file.Name, path)
	}

	shellPath, err := exec.LookPath("sh")
	if err != nil {
		return err
	}

	for i, command := range recipe.Commands {
		if pr != nil {
			pr(i+len(recipe.Files)+1, stepsLen, command.Name)
		}

		var stderr bytes.Buffer
		stdin, err := tar.readFile(command.Name)
		if err != nil {
			return err
		}

		if err := executeCommandInShell(shellPath, command.Restore, &stdin, nil, &stderr); err != nil {
			return errors.Join(err, errors.New(fmt.Sprintf("Command failed with error\n: %s", stderr.String())))
		}
	}

	return nil
}
