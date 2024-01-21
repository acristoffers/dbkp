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
func Backup(path string, recipe Recipe, password []byte, pr ProgressReport) error {
	if password != nil {
		recipe.EncryptionSalt = [2]string{"a", ""}
		return backupEncrypted(path, recipe, password, pr)
	}

	return backupPlain(path, recipe, pr)
}

// Executes a plain file backup (without encryption). pr is called before
// attempting to execute the backup of file/folder/command, if it is non-nil.
func backupPlain(path string, recipe Recipe, pr ProgressReport) error {
	backupFolder, err := filepath.Abs(filepath.Join(path, "dbkp-tmp"))
	if err != nil {
		return err
	}

	if err := os.RemoveAll(backupFolder); err != nil {
		return err
	}

	if err := os.MkdirAll(backupFolder, os.ModeDir|os.ModePerm); err != nil {
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

		backupPath := filepath.Join(backupFolder, file.Name)
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
			pr(i+len(recipe.Files)+1, stepsLen, command.Name)
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		backupPath := filepath.Join(backupFolder, command.Name)

		if err := executeCommandInShell(shellPath, command.Backup, nil, &stdout, &stderr); err != nil {
			return errors.Join(err, errors.New(fmt.Sprintf("Command failed with error\n: %s", stderr.String())))
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

	if err := os.RemoveAll(filepath.Join(path, "dbkp")); err != nil {
		return err
	}

	if err := os.Rename(backupFolder, filepath.Join(path, "dbkp")); err != nil {
		return err
	}

	return nil
}

// Executes an encrypted backup of recipe. A password is expected to be given
// (i.e.: non-nil/non-empty).
func backupEncrypted(path string, recipe Recipe, password []byte, pr ProgressReport) error {
	backupFile, err := filepath.Abs(filepath.Join(path, "dbkp"))
	if err != nil {
		return err
	}

	if err := os.RemoveAll(backupFile); err != nil {
		return err
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tarball := Tarball{}
	tarball.makeWrite()

	stepsLen := len(recipe.Files) + len(recipe.Commands)

	for i, file := range recipe.Files {
		path := file.Path
		if strings.HasPrefix(path, "~/") {
			path = filepath.Join(homePath, path[2:])
		}

		if pr != nil {
			pr(i+1, stepsLen, file.Name)
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

	for i, command := range recipe.Commands {
		if pr != nil {
			pr(i+len(recipe.Files)+1, stepsLen, command.Name)
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		if err := executeCommandInShell(shellPath, command.Backup, nil, &stdout, &stderr); err != nil {
			return errors.Join(err, errors.New(fmt.Sprintf("Command failed with error\n: %s", stderr.String())))
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
