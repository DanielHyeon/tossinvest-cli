//go:build unix

package config

// adoption_flock_unix.go serializes config saves (change
// console-adoption-controls, review round-1 P2-4). Same mechanism as
// internal/enginelock's, for the same reason: flock is a fact the kernel drops
// when the process ends however it ends, so a crashed saver wedges nobody.
// Blocking, unlike the engine's: a save holds the lock for milliseconds and the
// caller is an interactive human who would rather wait than get a refusal.

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type configLock struct{ file *os.File }

func (l configLock) release() {
	if l.file == nil {
		return
	}
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func acquireConfigLock(path string) (configLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return configLock{}, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return configLock{}, fmt.Errorf("config: opening lock %s: %w", path, err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		_ = file.Close()
		return configLock{}, fmt.Errorf("config: locking %s: %w", path, err)
	}
	return configLock{file: file}, nil
}
