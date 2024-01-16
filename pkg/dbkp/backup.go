package dbkp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Backup(path string, encrypt bool) error {
	configPath, err := filepath.Abs(filepath.Join(path, "dbkp.toml"))
	if err != nil {
		return err
	}

	config, err := LoadRecipe(configPath)
	if err != nil {
		return err
	}

	if len(config.EncryptionSalt) != 0 && len(config.EncryptionSalt[0]) != 0 {
		return BackupEncrypted(path, config)
	} else if encrypt {
		config.EncryptionSalt = [2]string{"a", ""}
		return BackupEncrypted(path, config)
	}

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

	for _, file := range config.Files {
		path := file.Path
		if strings.HasPrefix(path, "~") {
			path = strings.Replace(file.Path, "~", homePath, 1)
		}

		fmt.Printf("Backing up %s\n", path)

		backupPath := filepath.Join(backupFolder, file.Name)
		if err := copyFileOrFolder(path, backupPath, file); err != nil {
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

func BackupEncrypted(path string, config Recipe) error {
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

	password, err := AskForPassword()
	if err != nil {
		return err
	}

	tarball := Tarball{}
	tarball.MakeWrite()

	for _, file := range config.Files {
		path := file.Path
		if strings.HasPrefix(path, "~") {
			path = strings.Replace(path, "~", homePath, 1)
		}

		fmt.Printf("Backing up %s\n", path)

		subtarball := Tarball{}
		subtarball.MakeWrite()

		if err := subtarball.AddFileOrFolder(file.Name, path, file); err != nil {
			return err
		}

		if err := subtarball.CloseWrite(); err != nil {
			return err
		}

		if err := tarball.AddFile(file.Name, subtarball.Buffer.Bytes()); err != nil {
			return err
		}
	}

	if err := tarball.WriteToFile(backupFile, password, config); err != nil {
		return err
	}

	return nil
}
