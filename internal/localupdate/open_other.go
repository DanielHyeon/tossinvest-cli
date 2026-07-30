//go:build !unix

package localupdate

import (
	"fmt"
	"os"
)

func replacementSupported() bool { return false }

func openNoFollowExecutable(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is a symbolic link", path)
	}
	return os.Open(path)
}

func sameOpenFile(open *os.File, path string) error {
	opened, err := open.Stat()
	if err != nil {
		return err
	}
	current, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if current.Mode()&os.ModeSymlink != 0 || !os.SameFile(opened, current) {
		return fmt.Errorf("%s changed", path)
	}
	return nil
}

func syncDirectory(string) error { return ErrUnsupported }
