package dbkp

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Represents a tarball file and holds its data. Should only be used to read or
// write, but not both at the same time. Since all operations is done
// in-memmory, this structure is not ideal for large files, but since
// configuration files are mostly only a few KB, at most a few MB, it is not a
// problem.
type Tarball struct {
	Buffer  bytes.Buffer
	Writter *tar.Writer
}

// This function makes the Tarball writeable. No read is allowed after this.
func (tarball *Tarball) makeWrite() {
	tarball.Writter = tar.NewWriter(&tarball.Buffer)
}

// Closes the writter, finishing the tarball structure. The only safe thing to
// do after calling this function is to save the file to disk.
func (tarball *Tarball) closeWrite() error {
	if tarball.Writter == nil {
		return nil
	}

	if err := tarball.Writter.Close(); err != nil {
		return err
	}

	return nil
}

// Add the file/folder present in path to a file/folder named name in the
// tarball, respecting the restrictions in file.
func (tarball Tarball) addFileOrFolder(name string, path string, file File) error {
	fileinfo, err := os.Lstat(path)
	if err != nil {
		return err
	}

	filter, err := newPathFilter(file)
	if err != nil {
		return err
	}

	if fileinfo.IsDir() {
		return tarball.addFolderWithFilter(name, path, "", filter)
	} else if fileinfo.Mode().IsRegular() {
		contents, err := readFile(path)
		if err != nil {
			return nil
		}

		return tarball.addFile(name, contents)
	} else if fileinfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		realpath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil
		}

		fileinfo, err = os.Stat(realpath)
		if err != nil {
			return nil
		}

		if fileinfo.IsDir() {
			return tarball.addFolderWithFilter(name, realpath, "", filter)
		} else if fileinfo.Mode().IsRegular() {
			contents, err := readFile(path)
			if err != nil {
				return nil
			}

			return tarball.addFile(name, contents)
		}
	}

	return nil
}

// Adds contents as a file to the tarball as name.
func (tarball Tarball) addFile(name string, contents []byte) error {
	tw := tarball.Writter

	hdr := &tar.Header{
		Name: name,
		Mode: 0600,
		Size: int64(len(contents)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	if _, err := tw.Write(contents); err != nil {
		return err
	}

	return nil
}

func (tarball Tarball) addFolderWithFilter(name string, path string, prefix string, filter pathFilter) error {
	fsys := os.DirFS(path)
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if p == "." {
			return nil
		}

		srcpath := filepath.Join(path, p)
		dstpath := filepath.Join(name, p)

		fileinfo, err := os.Lstat(srcpath)
		if err != nil {
			return err
		}

		rel := p
		if prefix != "" {
			rel = filepath.Join(prefix, p)
		}

		skip, skipDir := filter.shouldSkip(rel, fileinfo.IsDir())
		if skip {
			if skipDir {
				return fs.SkipDir
			}
			return nil
		}

		if fileinfo.IsDir() {
			return nil
		} else if fileinfo.Mode().IsRegular() {
			contents, err := readFile(srcpath)
			if err != nil {
				return err
			}

			return tarball.addFile(dstpath, contents)
		} else if fileinfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			realpath, err := filepath.EvalSymlinks(srcpath)
			if err != nil {
				return nil
			}

			fileinfo, err := os.Stat(realpath)
			if err != nil {
				return nil
			}

			if fileinfo.IsDir() {
				return tarball.addFolderWithFilter(dstpath, realpath, rel, filter)
			} else if fileinfo.Mode().IsRegular() {
				contents, err := readFile(realpath)
				if err != nil {
					return err
				}

				return tarball.addFile(dstpath, contents)
			}
		}

		return nil
	})
}

// Reads a file from the tarball, returning its contents in a bytes.Buffer.
func (tarball Tarball) readFile(name string) (bytes.Buffer, error) {
	tr := tar.NewReader(&tarball.Buffer)
	var buffer bytes.Buffer

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		} else if err != nil {
			return buffer, err
		}

		if hdr.Name == name {
			if _, err := io.Copy(&buffer, tr); err != nil {
				return buffer, err
			}

			return buffer, nil
		}
	}

	return buffer, nil
}

// Saves all the contents of a tarball into path. name is removed from the
// beginning of the path (name is usually File.Name, which is was used to add
// the file/folder to the tarball in the first place).
func (tarball Tarball) unpackInto(name string, path string) error {
	tr := tar.NewReader(&tarball.Buffer)

	for {
		var buffer bytes.Buffer

		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		} else if err != nil {
			return err
		}

		if _, err := io.Copy(&buffer, tr); err != nil {
			return err
		}

		dstpath := filepath.Join(path, strings.Replace(hdr.Name, name, "", 1))
		dstdir := filepath.Dir(dstpath)
		if err := os.MkdirAll(dstdir, os.ModeDir|os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(dstpath, buffer.Bytes(), 0600); err != nil {
			return err
		}
	}

	return nil
}

// Decrypts and reads the tarball into memory.
func loadTarball(path string, password []byte, recipe Recipe) (Tarball, error) {
	tarball := Tarball{}

	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return tarball, err
	}

	key, _ := DeriveKeyFromPassword(password, recipe.EncryptionSalt[0])
	data, err := Decrypt(key, recipe.EncryptionSalt[1], ciphertext)
	if err != nil {
		return tarball, err
	}

	tarball.Buffer.Write(data)

	return tarball, nil
}

// Writes the tarball contents to file, encrypted.
func (tarball Tarball) writeToFile(path string, password []byte, recipe Recipe) error {
	if err := tarball.Writter.Close(); err != nil {
		return err
	}

	key, keysalt := DeriveKeyFromPassword(password, "")
	ciphertext, iv, err := Encrypt(key, tarball.Buffer.Bytes())
	if err != nil {
		return err
	}

	tomlPath := filepath.Join(filepath.Dir(path), "dbkp.toml")
	recipe.EncryptionSalt = [2]string{keysalt, iv}
	if err := recipe.WriteRecipe(tomlPath); err != nil {
		return err
	}

	return os.WriteFile(path, ciphertext, 0600)
}

// Copy existing tarball entries into dst, skipping any entry that belongs to
// the excluded names.
func (tarball Tarball) copyEntriesExcluding(dst *Tarball, excluded map[string]struct{}) error {
	tr := tar.NewReader(&tarball.Buffer)
	tw := dst.Writter

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if isExcludedEntry(hdr.Name, excluded) {
			continue
		}

		var buffer bytes.Buffer
		if _, err := io.Copy(&buffer, tr); err != nil {
			return err
		}

		newHdr := &tar.Header{
			Name: hdr.Name,
			Mode: hdr.Mode,
			Size: int64(buffer.Len()),
		}
		if err := tw.WriteHeader(newHdr); err != nil {
			return err
		}
		if _, err := tw.Write(buffer.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func isExcludedEntry(name string, excluded map[string]struct{}) bool {
	for key := range excluded {
		if name == key || strings.HasPrefix(name, key+"/") {
			return true
		}
	}
	return false
}
