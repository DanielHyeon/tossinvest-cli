//go:build unix

package protectionreadiness

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func currentOwnerUID() (uint32, bool) { return uint32(os.Geteuid()), true }

func fileOwner(info os.FileInfo) (uint32, bool) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, false
	}
	return stat.Uid, true
}

func secureParentFD(path string, owner uint32) (int, error) {
	parent := filepath.Dir(path)
	if filepath.Clean(path) != path || !filepath.IsAbs(path) || filepath.Base(path) == "." || filepath.Base(path) == string(filepath.Separator) {
		return -1, errors.New("invalid protection readiness path")
	}
	// Walk from a pinned root descriptor so O_NOFOLLOW applies to every path
	// component, not only the final config directory. Each open is relative to
	// the already-open parent, eliminating the Lstat/open replacement window.
	fd, err := unix.Open(string(filepath.Separator), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY, 0)
	if err != nil {
		return -1, err
	}
	components := strings.Split(strings.TrimPrefix(parent, string(filepath.Separator)), string(filepath.Separator))
	for _, component := range components {
		if component == "" || component == "." || component == ".." {
			_ = unix.Close(fd)
			return -1, errors.New("invalid protection readiness parent component")
		}
		next, openErr := unix.Openat(fd, component, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		_ = unix.Close(fd)
		if openErr != nil {
			return -1, openErr
		}
		fd = next
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || uint32(stat.Uid) != owner || stat.Mode&unix.S_IFMT != unix.S_IFDIR || stat.Mode&0o777 != 0o700 {
		_ = unix.Close(fd)
		return -1, errors.New("protection readiness parent changed while opening")
	}
	return fd, nil
}

func readOwnedFile(path string, owner uint32, mode os.FileMode) (os.FileInfo, []byte, error) {
	parentFD, err := secureParentFD(path, owner)
	if err != nil {
		return nil, nil, err
	}
	defer unix.Close(parentFD)
	fd, err := unix.Openat(parentFD, filepath.Base(path), unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = unix.Close(fd)
		return nil, nil, errors.New("cannot wrap protection readiness descriptor")
	}
	defer file.Close()
	before, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}
	uid, ok := fileOwner(before)
	if !ok || uid != owner || !before.Mode().IsRegular() || before.Mode().Perm() != mode || before.Size() <= 0 || before.Size() > maximumProductionFile {
		return nil, nil, errors.New("unsafe protection readiness file")
	}
	data, err := io.ReadAll(io.LimitReader(file, maximumProductionFile+1))
	if err != nil || int64(len(data)) != before.Size() {
		return nil, nil, errors.New("protection readiness file changed while read")
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(before, after) || before.Size() != after.Size() || before.ModTime() != after.ModTime() || before.Mode() != after.Mode() {
		return nil, nil, errors.New("protection readiness file changed while read")
	}
	return after, data, nil
}

func acquireStateLock(configDir string, owner uint32) (func(), bool, func() error, error) {
	lockPath := filepath.Join(configDir, ".protection-readiness-state.lock")
	parentFD, err := secureParentFD(lockPath, owner)
	if err != nil {
		return nil, false, nil, err
	}
	defer unix.Close(parentFD)
	name := filepath.Base(lockPath)
	fd, err := unix.Openat(parentFD, name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0o600)
	if errors.Is(err, unix.EEXIST) {
		fd, err = unix.Openat(parentFD, name, unix.O_RDWR|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	}
	if err != nil {
		return nil, false, nil, err
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || uint32(stat.Uid) != owner || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Mode&0o777 != 0o600 {
		_ = unix.Close(fd)
		return nil, false, nil, errors.New("unsafe protection readiness state lock")
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		_ = unix.Close(fd)
		return nil, false, nil, err
	}
	// Refresh after acquiring flock: a previous holder may have persisted the
	// marker while this process was waiting.
	if err := unix.Fstat(fd, &stat); err != nil || uint32(stat.Uid) != owner || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Mode&0o777 != 0o600 {
		_ = unix.Flock(fd, unix.LOCK_UN)
		_ = unix.Close(fd)
		return nil, false, nil, errors.New("protection readiness state lock changed while waiting")
	}
	// File existence is not a bootstrap proof: another process can open the
	// just-created inode and win flock before its creator. Only this marker,
	// written after durable state exists, makes later state absence corruption.
	var marker [1]byte
	read, readErr := unix.Pread(fd, marker[:], 0)
	if readErr != nil || (stat.Size != 0 && (stat.Size != 1 || read != 1 || marker[0] != 1)) {
		_ = unix.Flock(fd, unix.LOCK_UN)
		_ = unix.Close(fd)
		return nil, false, nil, errors.New("invalid protection readiness bootstrap marker")
	}
	anchored := stat.Size == 1 && read == 1 && marker[0] == 1
	mark := func() error {
		if err := unix.Ftruncate(fd, 0); err != nil {
			return err
		}
		if written, err := unix.Pwrite(fd, []byte{1}, 0); err != nil || written != 1 {
			if err != nil {
				return err
			}
			return io.ErrShortWrite
		}
		return unix.Fsync(fd)
	}
	return func() {
		_ = unix.Flock(fd, unix.LOCK_UN)
		_ = unix.Close(fd)
	}, anchored, mark, nil
}
