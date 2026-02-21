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

// Executes the backup of the recipe into path/dbkp. If a password is given,
// make it an encrypted backup.
func Backup(path string, recipe Recipe, password []byte, pr chan<- ProgressReport) error {
	return BackupSelected(path, recipe, password, pr, nil)
}

// Executes the backup of the selected names into path/dbkp. If names is empty,
// it behaves like a full backup.
func BackupSelected(path string, recipe Recipe, password []byte, pr chan<- ProgressReport, names []string) error {
	selectedRecipe, err := filterRecipeByNames(recipe, names)
	if err != nil {
		return err
	}

	partial := len(names) > 0
	if password != nil {
		return backupEncrypted(path, recipe, selectedRecipe, password, pr, partial)
	}

	return backupPlain(path, selectedRecipe, pr, partial)
}

// Executes a plain file backup (without encryption). pr is called before
// attempting to execute the backup of file/folder/command, if it is non-nil.
func backupPlain(path string, recipe Recipe, pr chan<- ProgressReport, partial bool) error {
	defer close(pr)

	backupFolder := ""
	backupTmp := ""
	var err error
	if partial {
		backupFolder, err = filepath.Abs(filepath.Join(path, "dbkp"))
		if err != nil {
			return err
		}

		if err := os.MkdirAll(backupFolder, os.ModeDir|os.ModePerm); err != nil {
			return err
		}
	} else {
		backupTmp, err = filepath.Abs(filepath.Join(path, "dbkp-tmp"))
		if err != nil {
			return err
		}

		if err := os.RemoveAll(backupTmp); err != nil {
			return err
		}

		if err := os.MkdirAll(backupTmp, os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		backupFolder = backupTmp
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	stepsLen := uint64(len(recipe.Files) + len(recipe.Commands))

	for i, file := range recipe.Files {
		path := file.Path
		if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homePath, path[2:])
		}

		if pr != nil {
			pr <- ProgressReport{uint64(i), stepsLen, file.Name}
		}

		backupPath := filepath.Join(backupFolder, file.Name)
		if partial {
			if err := os.RemoveAll(backupPath); err != nil {
				return err
			}
		}
		if err := copyFileOrFolder(path, backupPath, file); err != nil {
			return err
		}
	}

	shellPath, err := exec.LookPath("sh")
	if err != nil {
		return err
	}

	for i, command := range recipe.Commands {
		if pr != nil {
			pr <- ProgressReport{uint64(i + len(recipe.Files)), stepsLen, command.Name}
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		backupPath := filepath.Join(backupFolder, command.Name)

		if err := executeCommandInShell(shellPath, command.Backup, nil, &stdout, &stderr); err != nil {
			return errors.Join(err, fmt.Errorf("Command failed with error\n: %s", stderr.String()))
		}

		f, err := os.Create(backupPath)
		if err != nil {
			return err
		}
		defer func() {
			if e := f.Close(); e != nil {
				err = e
			}
		}()

		if _, err := f.Write(stdout.Bytes()); err != nil {
			return err
		}
	}

	if !partial {
		if err := os.RemoveAll(filepath.Join(path, "dbkp")); err != nil {
			return err
		}

		if err := os.Rename(backupFolder, filepath.Join(path, "dbkp")); err != nil {
			return err
		}
	}

	return nil
}

// Executes an encrypted backup of recipe. A password is expected to be given
// (i.e.: non-nil/non-empty).
func backupEncrypted(path string, recipe Recipe, selected Recipe, password []byte, pr chan<- ProgressReport, partial bool) error {
	defer close(pr)

	backupFile, err := filepath.Abs(filepath.Join(path, "dbkp"))
	if err != nil {
		return err
	}

	var existing Tarball
	if partial {
		if _, err := os.Stat(backupFile); err == nil && len(recipe.EncryptionSalt) > 0 && len(recipe.EncryptionSalt[0]) > 0 {
			existing, err = loadTarball(backupFile, password, recipe)
			if err != nil {
				return err
			}
		}
	} else {
		if err := os.RemoveAll(backupFile); err != nil {
			return err
		}
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tarball := Tarball{}
	tarball.makeWrite()

	selectedNames := map[string]struct{}{}
	for _, file := range selected.Files {
		selectedNames[file.Name] = struct{}{}
	}

	for _, command := range selected.Commands {
		selectedNames[command.Name] = struct{}{}
	}

	if partial && existing.Buffer.Len() > 0 {
		if err := existing.copyEntriesExcluding(&tarball, selectedNames); err != nil {
			return err
		}
	}

	stepsLen := uint64(len(selected.Files) + len(selected.Commands))

	for i, file := range selected.Files {
		path := file.Path
		if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homePath, path[2:])
		}

		if pr != nil {
			pr <- ProgressReport{uint64(i + 1), stepsLen, file.Name}
		}

		subtarball := Tarball{}
		subtarball.makeWrite()

		if err := subtarball.addFileOrFolder(file.Name, path, file); err != nil {
			return err
		}

		if err := subtarball.closeWrite(); err != nil {
			return err
		}

		if err := tarball.addFile(file.Name, subtarball.Buffer.Bytes()); err != nil {
			return err
		}
	}

	shellPath, err := exec.LookPath("sh")
	if err != nil {
		return err
	}

	for i, command := range selected.Commands {
		if pr != nil {
			pr <- ProgressReport{uint64(i + len(selected.Files) + 1), stepsLen, command.Name}
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		if err := executeCommandInShell(shellPath, command.Backup, nil, &stdout, &stderr); err != nil {
			return errors.Join(err, fmt.Errorf("Command failed with error\n: %s", stderr.String()))
		}

		if err := tarball.addFile(command.Name, stdout.Bytes()); err != nil {
			return err
		}
	}

	if err := tarball.writeToFile(backupFile, password, recipe); err != nil {
		return err
	}

	return nil
}
