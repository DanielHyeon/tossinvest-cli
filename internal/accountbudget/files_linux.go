//go:build linux

package accountbudget

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

const maxStateBytes = 128 << 10

func validScope(s string) bool {
	if len(s) == 0 || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
func checkedDir(path string) (int, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == "/" {
		return -1, errors.New("invalid root")
	}
	fd, e := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if e != nil {
		return -1, e
	}
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if part == "" || part == "." || part == ".." {
			unix.Close(fd)
			return -1, errors.New("bad path")
		}
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		unix.Close(fd)
		if err != nil {
			return -1, err
		}
		fd = next
	}
	var st unix.Stat_t
	var fs unix.Statfs_t
	if e = unix.Fstat(fd, &st); e != nil {
		unix.Close(fd)
		return -1, e
	}
	if st.Uid != uint32(os.Geteuid()) || st.Mode&07777 != 0700 {
		unix.Close(fd)
		return -1, errors.New("unsafe root")
	}
	if e = unix.Fstatfs(fd, &fs); e != nil {
		unix.Close(fd)
		return -1, e
	}
	if fs.Type != unix.EXT4_SUPER_MAGIC && fs.Type != unix.TMPFS_MAGIC {
		unix.Close(fd)
		return -1, errors.New("unsupported filesystem")
	}
	return fd, nil
}
func checkedFile(dir int, name string, flags int) (*os.File, unix.Stat_t, error) {
	var st unix.Stat_t
	if name == "" || strings.Contains(name, "/") || name == "." || name == ".." {
		return nil, st, errors.New("invalid name")
	}
	fd, e := unix.Openat(dir, name, flags|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0600)
	if e != nil {
		return nil, st, e
	}
	if e = unix.Fstat(fd, &st); e != nil || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Mode&07777 != 0600 || st.Uid != uint32(os.Geteuid()) || st.Nlink != 1 {
		unix.Close(fd)
		return nil, st, errors.New("unsafe file")
	}
	return os.NewFile(uintptr(fd), name), st, nil
}
func readBounded(f *os.File, max int) ([]byte, error) {
	before, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if before.Size() > int64(max) {
		return nil, errors.New("oversized file")
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(max)+1))
	if e != nil {
		return nil, e
	}
	after, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if len(b) > max || int64(len(b)) != before.Size() || !os.SameFile(before, after) || before.Size() != after.Size() || before.ModTime() != after.ModTime() {
		return nil, errors.New("changed file")
	}
	return b, nil
}
func checkIdentity(root int, name string, expected unix.Stat_t) error {
	f, st, e := checkedFile(root, name, unix.O_RDONLY)
	if e != nil {
		return e
	}
	defer f.Close()
	if st.Dev != expected.Dev || st.Ino != expected.Ino {
		return fmt.Errorf("%s inode changed", name)
	}
	return nil
}
