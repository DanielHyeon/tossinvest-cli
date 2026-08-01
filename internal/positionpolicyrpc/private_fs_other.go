//go:build !unix

package positionpolicyrpc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func rejectSymlinkTraversal(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return err
	}
	if filepath.Clean(abs) != filepath.Clean(resolved) {
		return fmt.Errorf("position policy control: symlink traversal is forbidden: %s", path)
	}
	return nil
}

func validateOwnerAndLinks(os.FileInfo, bool) error {
	return errors.New("position policy control: filesystem ownership verification is unsupported on this platform")
}

func openDescriptorNoFollow(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("position policy control: descriptor is a symlink")
	}
	return os.Open(path)
}
