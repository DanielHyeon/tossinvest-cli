//go:build windows

package httpapi

import (
	"errors"
	"os"
	"path/filepath"
)

func validateSecurityStoreDirectory(path string) error {
	info, err := os.Lstat(filepath.Clean(path))
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("security database parent must be a real existing directory")
	}
	return nil
}

func validateSecurityStoreFile(info os.FileInfo) error {
	if info == nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("database must be a regular file, not a symlink")
	}
	return nil
}
