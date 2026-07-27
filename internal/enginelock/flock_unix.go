//go:build unix

package enginelock

// flock_unix.go is the exclusion itself: LOCK_EX|LOCK_NB on an open file
// descriptor in the journal directory.
//
// # Why flock and not a pid file
//
// A pid file is a claim; flock is a fact the kernel maintains. The failure the
// engine has to survive is a crash — SIGKILL, an OOM, a power cut — and a pid file
// left behind by one of those wedges every subsequent start until somebody deletes
// it by hand. A flock is dropped when the process ends, however it ends, because
// the descriptor is closed by the kernel. There is nothing to clean up and nothing
// to check for staleness.
//
// # Non-blocking, always
//
// LOCK_NB, so a second start refuses immediately instead of queueing. An engine
// that waited would come up whenever the first one happened to stop, with an
// operator long gone and a picture of the account from another era.

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
	// Unlock before close is belt and braces: closing the descriptor drops the
	// lock on every Unix, and an explicit unlock makes that visible to a reader
	// who does not know it.
	_ = syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	_ = l.file.Close()
}

func acquireLock(path string) (lockHandle, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("enginelock: opening %s: %w", path, err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, fmt.Errorf("%w (%s)", ErrAlreadyRunning, path)
		}
		return nil, fmt.Errorf("enginelock: locking %s: %w", path, err)
	}
	return unixLock{file: file}, nil
}
