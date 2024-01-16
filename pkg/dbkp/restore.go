package dbkp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Restore(path string) error {
	configPath, err := filepath.Abs(filepath.Join(path, "dbkp.toml"))
	if err != nil {
		return err
	}

	config, err := LoadRecipe(configPath)
	if err != nil {
		return err
	}

	backupFolder, err := filepath.Abs(filepath.Join(path, "dbkp"))
	if err != nil {
		return err
	}

	if len(config.EncryptionSalt) != 0 && len(config.EncryptionSalt[0]) != 0 {
		return RestoreEncrypt(backupFolder, config)
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	for _, file := range config.Files {
		path := file.Path
		if strings.HasPrefix(path, "~") {
			path = strings.Replace(file.Path, "~", homePath, 1)
		}

		fmt.Printf("Restoring %s\n", path)
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

	return nil
}

func RestoreEncrypt(backupFile string, config Recipe) error {
	password, err := AskForPassword()
	if err != nil {
		return err
	}

	tar, err := LoadTarball(backupFile, password, config)
	if err != nil {
		return err
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	for _, file := range config.Files {
		path := file.Path
		if strings.HasPrefix(path, "~") {
			path = strings.Replace(file.Path, "~", homePath, 1)
		}

		fmt.Printf("Restoring %s\n", path)

		subtarbuffer, err := tar.ReadFile(file.Name)
		subtar := Tarball{Buffer: subtarbuffer}
		if err != nil {
			return err
		}

		if err := os.RemoveAll(path); err != nil {
			return err
		}

		subtar.UnpackInto(file.Name, path)
	}

	return nil
}
