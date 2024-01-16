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

type Tarball struct {
	Buffer  bytes.Buffer
	Writter *tar.Writer
}

func (tarball *Tarball) MakeWrite() {
	tarball.Writter = tar.NewWriter(&tarball.Buffer)
}

func (tarball *Tarball) CloseWrite() error {
	if tarball.Writter == nil {
		return nil
	}

	if err := tarball.Writter.Close(); err != nil {
		return err
	}

	return nil
}

func (tarball Tarball) AddFileOrFolder(name string, path string, file File) error {
	fileinfo, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if fileinfo.IsDir() {
		if len(file.Only) != 0 || len(file.Exclude) != 0 {
			entries, err := fs.ReadDir(os.DirFS(path), ".")
			if err != nil {
				return err
			}

			only := len(file.Only) != 0
			exclude := len(file.Exclude) != 0

			for _, entry := range entries {
				test_only := only && Contains(file.Only, entry.Name())
				test_exclude := exclude && !Contains(file.Exclude, entry.Name())
				if test_only || test_exclude {
					srcpath := filepath.Join(path, entry.Name())
					dstpath := filepath.Join(name, entry.Name())

					fileinfo, err := os.Stat(srcpath)
					if err != nil {
						return err
					}

					if fileinfo.IsDir() {
						if err := tarball.AddFolder(dstpath, path); err != nil {
							return err
						}
					} else if fileinfo.Mode().IsRegular() {
						contents, err := readFile(srcpath)
						if err != nil {
							return nil
						}

						if err := tarball.AddFile(dstpath, contents); err != nil {
							return err
						}
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
							if err := tarball.AddFolder(dstpath, realpath); err != nil {
								return err
							}
						} else if fileinfo.Mode().IsRegular() {
							contents, err := readFile(path)
							if err != nil {
								return nil
							}

							if err := tarball.AddFile(name, contents); err != nil {
								return err
							}
						}
					}
				}
			}
		} else {
			if err := tarball.AddFolder(name, path); err != nil {
				return err
			}
		}
	} else if fileinfo.Mode().IsRegular() {
		contents, err := readFile(path)
		if err != nil {
			return nil
		}

		if err := tarball.AddFile(name, contents); err != nil {
			return err
		}
	} else if fileinfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		realpath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil
		}

		fileinfo, err := os.Stat(realpath)
		if err != nil {
			return nil
		}

		if fileinfo.IsDir() {
			if err := tarball.AddFolder(name, realpath); err != nil {
				return err
			}
		} else if fileinfo.Mode().IsRegular() {
			contents, err := readFile(path)
			if err != nil {
				return nil
			}

			if err := tarball.AddFile(name, contents); err != nil {
				return err
			}
		}
	}

	return nil
}

func (tarball Tarball) AddFile(name string, contents []byte) error {
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

func (tarball Tarball) AddFolder(name string, path string) error {
	fsys := os.DirFS(path)
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		srcpath := filepath.Join(path, p)
		dstpath := filepath.Join(name, p)

		if err != nil {
			return err
		}

		fileinfo, err := os.Lstat(srcpath)
		if err != nil {
			return err
		}

		if fileinfo.Mode().IsRegular() {
			contents, err := readFile(srcpath)
			if err != nil {
				return err
			}

			if err := tarball.AddFile(dstpath, contents); err != nil {
				return err
			}
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
				if err := tarball.AddFolder(dstpath, realpath); err != nil {
					return err
				}
			} else if fileinfo.Mode().IsRegular() {
				contents, err := readFile(realpath)
				if err != nil {
					return err
				}

				if err := tarball.AddFile(dstpath, contents); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (tarball Tarball) ReadFile(name string) (bytes.Buffer, error) {
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

func (tarball Tarball) UnpackInto(name string, path string) error {
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

func LoadTarball(path string, password []byte, config Recipe) (Tarball, error) {
	tarball := Tarball{}

	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return tarball, err
	}

	key, _ := DeriveKeyFromPassword(password, config.EncryptionSalt[0])
	data, err := Decrypt(key, config.EncryptionSalt[1], ciphertext)
	if err != nil {
		return tarball, err
	}

	tarball.Buffer.Write(data)

	return tarball, nil
}

func (tarball Tarball) WriteToFile(path string, password []byte, config Recipe) error {
	if err := tarball.Writter.Close(); err != nil {
		return err
	}

	key, keysalt := DeriveKeyFromPassword(password, "")
	ciphertext, iv, err := Encrypt(key, tarball.Buffer.Bytes())
	if err != nil {
		return err
	}

	tomlPath := filepath.Join(filepath.Dir(path), "dbkp.toml")
	config.EncryptionSalt = [2]string{keysalt, iv}
	if err := config.WriteRecipe(tomlPath); err != nil {
		return err
	}

	return os.WriteFile(path, ciphertext, 0600)
}
