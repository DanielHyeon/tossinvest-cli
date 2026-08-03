//go:build unix

package ratebudget

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

type unixLock struct{ file *os.File }

func (l unixLock) release() {
	if l.file == nil {
		return
	}
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func tryLock(path string) (lockHandle, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("ratebudget: opening %s: %w", path, err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrBusy
		}
		return nil, fmt.Errorf("ratebudget: locking %s: %w", path, err)
	}
	return unixLock{file: file}, nil
}
