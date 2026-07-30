//go:build unix

package localupdate

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func replacementSupported() bool { return true }

func openNoFollowExecutable(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is a symbolic link", path)
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
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
	if current.Mode()&os.ModeSymlink != 0 || !current.Mode().IsRegular() {
		return fmt.Errorf("%s is no longer a regular non-symlink executable", path)
	}
	if !os.SameFile(opened, current) {
		return fmt.Errorf("%s no longer names the executable opened for commit", path)
	}
	return nil
}

func syncDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
