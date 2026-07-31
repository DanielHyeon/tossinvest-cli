//go:build unix

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const desiredLockMaxWait = 2 * time.Second

type desiredLock struct{ file *os.File }

func (l desiredLock) release() {
	if l.file == nil {
		return
	}
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func acquireDesiredLock(ctx context.Context, path string) (desiredLock, error) {
	if err := ctx.Err(); err != nil {
		return desiredLock{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return desiredLock{}, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return desiredLock{}, fmt.Errorf("scheduler: opening desired-state lock %s: %w", path, err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, desiredLockMaxWait)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			if err := waitCtx.Err(); err != nil {
				_ = file.Close()
				return desiredLock{}, err
			}
			return desiredLock{file: file}, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			_ = file.Close()
			return desiredLock{}, fmt.Errorf("scheduler: locking desired state %s: %w", path, err)
		}
		select {
		case <-waitCtx.Done():
			_ = file.Close()
			return desiredLock{}, waitCtx.Err()
		case <-ticker.C:
		}
	}
}
