package dbkp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func copyFileOrFolder(src string, dst string, file File) error {
	fileinfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if fileinfo.IsDir() {
		if len(file.Only) != 0 || len(file.Exclude) != 0 {
			if err := os.MkdirAll(dst, os.ModeDir|os.ModePerm); err != nil {
				return err
			}

			entries, err := fs.ReadDir(os.DirFS(src), ".")
			if err != nil {
				return err
			}

			only := len(file.Only) != 0
			exclude := len(file.Exclude) != 0

			for _, entry := range entries {
				test_only := only && Contains(file.Only, entry.Name())
				test_exclude := exclude && !Contains(file.Exclude, entry.Name())
				if test_only || test_exclude {
					srcpath := filepath.Join(src, entry.Name())
					dstpath := filepath.Join(dst, entry.Name())

					fileinfo, err := os.Stat(srcpath)
					if err != nil {
						return err
					}

					if fileinfo.IsDir() {
						if err := CopyDir(srcpath, dstpath); err != nil {
							return err
						}
					} else if fileinfo.Mode().IsRegular() {
						if err := CopyFile(srcpath, dstpath); err != nil {
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
							if err := CopyDir(realpath, dstpath); err != nil {
								return err
							}
						} else if fileinfo.Mode().IsRegular() {
							if err := CopyFile(realpath, dstpath); err != nil {
								return err
							}
						}
					}
				}
			}
		} else {
			if err := CopyDir(src, dst); err != nil {
				return err
			}
		}
	} else if fileinfo.Mode().IsRegular() {
		if err := CopyFile(src, dst); err != nil {
			return err
		}
	} else if fileinfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		realpath, err := filepath.EvalSymlinks(src)
		if err != nil {
			return nil
		}

		fileinfo, err := os.Stat(realpath)
		if err != nil {
			return nil
		}

		if fileinfo.IsDir() {
			if err := CopyDir(realpath, dst); err != nil {
				return err
			}
		} else if fileinfo.Mode().IsRegular() {
			if err := CopyFile(realpath, dst); err != nil {
				return err
			}
		}
	}

	return nil
}

func CopyFile(src string, dst string) error {
	if os.Link(src, dst) == nil {
		return nil
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := out.Sync(); err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.Chmod(dst, si.Mode()); err != nil {
		return err
	}

	return nil
}

func CopyDir(src string, dst string) error {
	fsys := os.DirFS(src)
	return fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		srcpath := filepath.Join(src, p)
		dstpath := filepath.Join(dst, p)

		if err != nil {
			return err
		}

		fileinfo, err := os.Lstat(srcpath)
		if err != nil {
			return err
		}

		if fileinfo.IsDir() {
			if err := os.MkdirAll(dstpath, os.ModeDir|os.ModePerm); err != nil {
				return err
			}
		} else if fileinfo.Mode().IsRegular() {
			if err := CopyFile(srcpath, dstpath); err != nil {
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
				if err := CopyDir(realpath, dstpath); err != nil {
					return err
				}
			} else if fileinfo.Mode().IsRegular() {
				if err := CopyFile(realpath, dstpath); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fileinfo.Size())
	if _, err := file.Read(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

func AskForPassword() ([]byte, error) {
	fmt.Print("Password: ")
	defer fmt.Println("")
	return term.ReadPassword(int(syscall.Stdin))
}

func DeriveKeyFromPassword(password []byte, salt string) ([]byte, string) {
	saltbytes, err := hex.DecodeString(salt)
	if err != nil || len(saltbytes) != 32 {
		saltbytes = make([]byte, 32)
		rand.Read(saltbytes)
	}
	salt = hex.EncodeToString(saltbytes)
	return pbkdf2.Key(password, saltbytes, 1000, 32, sha256.New), salt
}

func Encrypt(key []byte, data []byte) ([]byte, string, error) {
	saltbytes := make([]byte, 12)
	rand.Read(saltbytes)
	salt := hex.EncodeToString(saltbytes)

	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, "", err
	}

	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, "", err
	}

	ciphertext := aesgcm.Seal(nil, saltbytes, data, nil)
	return ciphertext, salt, nil
}

func Decrypt(key []byte, salt string, ciphertext []byte) ([]byte, error) {
	saltbytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, err
	} else if len(saltbytes) != 12 {
		return nil, errors.New("cannot decrypt: salt has wrong size")
	}

	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return nil, err
	}

	return aesgcm.Open(nil, saltbytes, ciphertext, nil)
}
