//go:build linux || darwin

package soak

import (
	"os"
	"syscall"
)

// openRenewalStatus는 검증과 읽기를 하나의 descriptor에 묶는다. O_NONBLOCK는
// FIFO가 regular file 검사 전에 콘솔 요청을 기다리지 못하게 한다.
func openRenewalStatus(path string) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func renewalStatusOwnedByCurrentUser(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(stat.Uid) == os.Getuid()
}

func renewalStatusReaderSupported() bool { return true }
