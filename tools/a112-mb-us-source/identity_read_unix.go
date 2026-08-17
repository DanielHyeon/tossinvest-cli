//go:build unix

package main

import (
	"errors"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// readRegularNoFollow pins the opened inode and rejects symlinks, devices,
// directories, and FIFOs before reading any identity input.
func readRegularNoFollow(path string) ([]byte, error) {
	return readRegularNoFollowWithLimit(path, maxIdentityInputBytes)
}

func readRegularNoFollowWithLimit(path string, limit int64) ([]byte, error) {
	if limit < 0 {
		return nil, errors.New("identity input limit is invalid")
	}
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	if file == nil {
		_ = unix.Close(fd)
		return nil, errors.New("identity input descriptor is invalid")
	}
	defer file.Close()
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Size < 0 {
		return nil, errors.New("identity input is not a regular no-follow file")
	}
	if stat.Size > limit {
		return nil, errors.New("identity input exceeds size limit")
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errors.New("identity input exceeds size limit")
	}
	var after unix.Stat_t
	if err := unix.Fstat(fd, &after); err != nil || after.Mode&unix.S_IFMT != unix.S_IFREG || after.Dev != stat.Dev || after.Ino != stat.Ino || after.Size != stat.Size || int64(len(data)) != after.Size {
		return nil, errors.New("identity input changed or truncated during read")
	}
	return data, nil
}
