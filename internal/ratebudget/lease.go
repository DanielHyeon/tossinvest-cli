// Package ratebudget serializes optional Open API reads with supervised live
// verification across processes that use the same TossOS profile.
package ratebudget

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const FileName = "openapi-rate-budget.lock"

var (
	ErrBusy        = errors.New("Open API rate budget is reserved by live verification or another metadata read")
	ErrUnsupported = errors.New("Open API rate-budget file locks are unsupported on this platform")
)

type lockHandle interface{ release() }

type Lease struct {
	path string

	mu       sync.Mutex
	released bool
	handle   lockHandle
}

func (l *Lease) Path() string { return l.path }

func (l *Lease) Release() {
	if l == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return
	}
	l.released = true
	l.handle.release()
}

func TryAcquire(path string) (*Lease, bool, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "." || strings.TrimSpace(path) == "" {
		return nil, false, fmt.Errorf("ratebudget: no lock path was named")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, false, fmt.Errorf("ratebudget: creating %s: %w", filepath.Dir(path), err)
	}
	handle, err := tryLock(path)
	if errors.Is(err, ErrBusy) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &Lease{path: path, handle: handle}, true, nil
}

func Acquire(ctx context.Context, path string) (*Lease, error) {
	for {
		lease, ok, err := TryAcquire(path)
		if err != nil {
			return nil, err
		}
		if ok {
			return lease, nil
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}
