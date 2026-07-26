// Package runlock is a one-file advisory marker: "a live verification is using
// this account's rate budget right now".
//
// # What problem it solves
//
// `tossctl soak run` walks the read endpoints on a timer for days. `tossctl
// verify run` and `tossctl console` walk a supervised live procedure against the
// same account, the same credentials and the same rate limit. On 2026-07-26 the
// two overlapped and the verification lost three steps to a broker 429 that had
// nothing to do with the account (measurements.md M4): a soak cycle had used the
// same endpoints minutes earlier.
//
// The verification is the scarce thing. It needs a person watching, it places
// real orders, and a lost step costs another one. A soak cycle is worth nothing
// in particular and there will be another in fifteen minutes. So the soak yields.
//
// # Why it is advisory, and what that buys
//
// This is not mutual exclusion and it must not be mistaken for it. Two processes
// that both want the lock both get it; the only thing the file does is let the
// soak notice and wait.
//
// That is deliberate, and the failure modes are why:
//
//	a crashed verify   leaves the file behind. A real lock would wedge the soak
//	                   until somebody deleted it by hand — days of a multi-day
//	                   survey lost to a file. Freshness is judged from the
//	                   modification time and a lock nobody has touched for
//	                   StaleAfter is ignored, so a crash costs one pause and
//	                   nothing else.
//	an unwritable dir  is not a reason to refuse a verification. Acquire's error
//	                   is for the caller to mention and carry on; nothing here is
//	                   a precondition for anything.
//	a stopped soak     is not this package's business. It never blocks the writer
//	                   and it never blocks the reader. There is no IPC, no daemon
//	                   and no protocol — one file, one timestamp.
//
// The file's contents are for a person reading it with `cat`; nothing parses
// them. Freshness is the modification time, because that is the one fact the
// filesystem maintains whether or not the writer got as far as flushing.
package runlock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileName is the marker's name. It sits in the data directory next to
// verifylive's evidence record, because the thing it is about is that record's
// run.
const FileName = "verify-run.lock"

// StaleAfter is how long a lock stays believable without being touched.
//
// Five minutes. The refresh interval is a minute, so four missed refreshes are
// needed before a live run is mistaken for a dead one — and the cost of that
// mistake is one soak cycle competing for the rate budget, not a lost
// measurement. In the other direction the cost of a long window is a crashed
// verification pausing the soak for that long, which is why it is minutes.
const StaleAfter = 5 * time.Minute

// RefreshEvery is how often a held lock is touched.
const RefreshEvery = time.Minute

// Lock is a held marker.
type Lock struct {
	path string

	mu       sync.Mutex
	released bool
}

// Acquire creates or refreshes the marker.
//
// It does not fail when the file already exists: this is advisory, and a second
// verification is a thing the console's own process boundary prevents, not
// something a lock file should be arbitrating.
func Acquire(path string, now time.Time) (*Lock, error) {
	l := &Lock{path: path}
	if err := l.write(now); err != nil {
		return nil, err
	}
	return l, nil
}

// Path reports where the marker is.
func (l *Lock) Path() string { return l.path }

// Refresh touches the marker so a reader keeps seeing it as live.
func (l *Lock) Refresh(now time.Time) error {
	l.mu.Lock()
	released := l.released
	l.mu.Unlock()
	if released {
		return nil
	}
	return l.write(now)
}

// Release removes the marker. It is safe to call more than once.
//
// A failure to remove is not returned as an error the caller has to handle: the
// file goes stale on its own, which is the whole reason freshness is a timestamp
// rather than existence.
func (l *Lock) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return
	}
	l.released = true
	if err := os.Remove(l.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}
}

func (l *Lock) write(now time.Time) error {
	dir := filepath.Dir(l.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("runlock: creating %s: %w", dir, err)
	}
	body := fmt.Sprintf("pid %d\nrefreshed %s\n\nAdvisory only. `tossctl soak run` pauses its cycle while this\n"+
		"file is less than %s old, so a live verification does not compete with the\n"+
		"survey for the same rate limit. Delete it if it is left behind; a stale one\n"+
		"is ignored anyway.\n", os.Getpid(), now.UTC().Format(time.RFC3339), StaleAfter)
	if err := os.WriteFile(l.path, []byte(body), 0o600); err != nil {
		return fmt.Errorf("runlock: writing %s: %w", l.path, err)
	}
	// The contents are for a person; the modification time is the fact. Setting it
	// explicitly keeps an injected clock authoritative in a test rather than
	// whatever the filesystem happened to stamp.
	if err := os.Chtimes(l.path, now, now); err != nil {
		return fmt.Errorf("runlock: timestamping %s: %w", l.path, err)
	}
	return nil
}

// Fresh reports whether a marker exists and has been touched inside staleAfter.
//
// A missing file, an unreadable one and an old one are all "no": every failure
// direction here has to be "carry on", because the reader is a survey that must
// not be stopped by this package's problems.
func Fresh(path string, now time.Time, staleAfter time.Duration) (bool, time.Time) {
	info, err := os.Stat(path)
	if err != nil {
		return false, time.Time{}
	}
	modified := info.ModTime()
	age := now.Sub(modified)
	if age < 0 {
		// A timestamp in the future is a clock that disagrees with itself, not a
		// dead run. Treating it as fresh errs towards yielding.
		return true, modified
	}
	return age <= staleAfter, modified
}

// Hold acquires the marker and keeps it fresh until ctx ends or release is
// called.
//
// The returned release is always non-nil, so a caller can defer it without
// checking the error first — which is the shape this needs, because the error is
// something to mention and ignore rather than something to abort on.
func Hold(ctx context.Context, path string) (release func(), err error) {
	lock, err := Acquire(path, time.Now())
	if err != nil {
		return func() {}, err
	}

	done := make(chan struct{})
	var once sync.Once
	stop := func() {
		once.Do(func() {
			close(done)
			lock.Release()
		})
	}

	go func() {
		ticker := time.NewTicker(RefreshEvery)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = lock.Refresh(time.Now())
			}
		}
	}()

	return stop, nil
}
