//go:build unix

package positionpolicyrpc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
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

func validateOwnerAndLinks(info os.FileInfo, requireOneLink bool) error {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("position policy control: cannot verify filesystem ownership")
	}
	if stat.Uid != uint32(os.Geteuid()) {
		return errors.New("position policy control: filesystem object has the wrong owner")
	}
	if requireOneLink && stat.Nlink != 1 {
		return errors.New("position policy control: descriptor must have exactly one hard link")
	}
	return nil
}

func openDescriptorNoFollow(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}
