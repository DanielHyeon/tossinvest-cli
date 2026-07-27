// Package enginelock is the engine runtime's single-instance mechanism, which is
// two separate things with two separate jobs (openspec change add-engine-runtime,
// engine-safety "엔진 런타임 수명주기").
//
//	flock    the exclusion. An advisory-but-kernel-enforced exclusive lock on a
//	         file in the journal directory, taken as the FIRST action of a start
//	         and released when the process ends however it ends. Two engines
//	         cannot hold it, including two that started in the same millisecond.
//	marker   the advisory signal. A file whose modification time says "an engine
//	         was alive within the last five minutes", refreshed every minute,
//	         removed on a clean stop. The console reads it to draw the engine's
//	         status and the autostart script reads it before spawning anything.
//
// # Why both, and why they are not the same thing
//
// The marker cannot be the exclusion. Its freshness is a timestamp, so two
// processes starting together both read "no engine" and both start — which is
// exactly the case the autostart script and the console's [엔진 시작] button
// create between them, and exactly the case a journal designed for a single
// writer must not be in. Review round 2 raised this and the answer was to stop
// asking an advisory file to be a lock: internal/runlock's package comment
// explains at length why *it* is advisory, and every one of those reasons still
// holds for the marker's job. None of them holds for the journal's.
//
// The flock cannot be the signal. It is invisible: a reader can only learn
// whether a lock is held by trying to take it, which for a console rendering a
// dashboard would mean fighting the engine for its own lock several times a
// minute. So the question "is an engine running" is answered by a timestamp, and
// the question "may I start one" is answered by the kernel.
//
// # Why the lock file is in the journal directory
//
// Because what is being protected is the journal. The engine's durable record is
// a single-writer SQLite database with a migration step, and the whole point of
// taking the lock before anything else — including opening the journal — is that
// the open and the migration happen inside the exclusion. A binary upgrade
// followed by autostart and a console button racing to start the new engine was
// precisely that path (review round 3).
//
// # Marker freshness numbers
//
// One minute refresh, five minutes stale — internal/runlock's, reused rather than
// re-chosen, and Fresh is literally runlock.Fresh. Four missed refreshes before a
// live engine is mistaken for a dead one, and a crashed engine's marker is
// believed for at most five minutes. The consequence of either mistake is a
// misdrawn status line or a refused autostart, never a second writer: the flock
// is what closes that.
package enginelock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

// LockFileName is the exclusion. It sits in the journal directory because the
// journal is what it protects.
const LockFileName = "engine.lock"

// MarkerFileName is the advisory "an engine is alive" signal, beside the lock.
const MarkerFileName = "engine-run.json"

// StaleAfter is how long a marker stays believable without being touched. It is
// runlock's, named here so a reader of this package does not have to follow the
// reference to learn the number.
const StaleAfter = runlock.StaleAfter

// RefreshEvery is how often a held marker is rewritten.
const RefreshEvery = runlock.RefreshEvery

// ErrAlreadyRunning means another process holds the journal directory's lock.
//
// It is a refusal and never a wait: an engine that queued behind another engine
// would come up minutes later with a stale idea of the account, and the operator
// who pressed the button twice would get two engines eventually rather than one
// now.
var ErrAlreadyRunning = errors.New(
	"engine: another engine instance already holds this journal directory")

// ErrLockUnsupported means this platform has no advisory file lock this package
// knows how to take.
//
// It is an honest refusal rather than a silent downgrade, which is the precedent
// soakproc_unix.go/soakproc_other.go set for the same question: a process
// detachment that does not work is a worse survey, while an exclusion that does
// not exclude is two writers on one journal.
var ErrLockUnsupported = errors.New(
	"engine: this platform has no file lock, so a second instance could not be excluded — " +
		"the engine runtime is supported on Unix only")

// lockHandle is the platform's held lock. The two implementations are
// flock_unix.go and flock_other.go.
type lockHandle interface{ release() }

// Lock is a held journal-directory lock.
type Lock struct {
	path string

	mu       sync.Mutex
	released bool
	handle   lockHandle
}

// Path reports where the lock file is.
func (l *Lock) Path() string { return l.path }

// Release drops the lock. It is safe to call more than once, and the kernel drops
// it anyway when the process ends — which is the property that makes a crashed
// engine cost nothing.
func (l *Lock) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return
	}
	l.released = true
	l.handle.release()
}

// Acquire takes the exclusive lock on dir's lock file.
//
// It creates the directory if it is missing, because on a first run the journal
// directory does not exist yet and the lock is taken before the journal opens.
func Acquire(dir string) (*Lock, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("enginelock: no journal directory was named")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("enginelock: creating %s: %w", dir, err)
	}
	path := filepath.Join(dir, LockFileName)
	handle, err := acquireLock(path)
	if err != nil {
		return nil, err
	}
	return &Lock{path: path, handle: handle}, nil
}

// --- the advisory marker ----------------------------------------------------------

// Marker is what a running engine publishes about itself.
//
// It is JSON because the console renders three of these fields and a human reads
// the fourth; internal/runlock's marker is prose because nothing parses it. The
// reader is deliberately lenient — an unparseable marker is a running engine with
// an unknown build, not an error — so a format change can never turn into a
// dashboard that refuses to draw.
type Marker struct {
	PID         int            `json:"pid"`
	StartedAt   time.Time      `json:"started_at"`
	RefreshedAt time.Time      `json:"refreshed_at"`
	Binary      binstamp.Stamp `json:"binary,omitempty"`
	Note        string         `json:"note,omitempty"`
}

// markerNote is written into every marker so somebody who finds the file with
// `cat` knows what it is and what deleting it does.
const markerNote = "advisory only. The exclusion is the flock on engine.lock; this file exists so the " +
	"console can draw the engine's status and the autostart script can check before spawning. " +
	"A stale one (older than 5m) is ignored."

// Hold writes the marker and keeps it fresh until ctx ends or the returned
// release is called.
//
// The release is always non-nil so a caller can defer it without checking the
// error first: a marker that could not be written is a status line the console
// cannot draw, which is worth mentioning and is not worth refusing a start over.
// The exclusion has already been taken by then.
func Hold(ctx context.Context, path string, now time.Time) (release func(), err error) {
	m := Marker{PID: os.Getpid(), StartedAt: now.UTC(), Note: markerNote}
	m.Binary, _ = binstamp.Self()

	if werr := write(path, m, now); werr != nil {
		return func() {}, werr
	}

	done := make(chan struct{})
	var once sync.Once
	stop := func() {
		once.Do(func() {
			close(done)
			if rerr := os.Remove(path); rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
				// A marker that could not be removed goes stale on its own within
				// StaleAfter, which is the whole reason freshness is a timestamp
				// rather than existence.
				return
			}
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
			case at := <-ticker.C:
				_ = write(path, m, at)
			}
		}
	}()

	return stop, nil
}

func write(path string, m Marker, at time.Time) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("enginelock: creating %s: %w", filepath.Dir(path), err)
	}
	m.RefreshedAt = at.UTC()
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("enginelock: rendering the marker: %w", err)
	}
	if err := os.WriteFile(path, append(body, '\n'), 0o600); err != nil {
		return fmt.Errorf("enginelock: writing %s: %w", path, err)
	}
	// The contents are for a reader; the modification time is the fact. Setting it
	// explicitly keeps an injected clock authoritative in a test rather than
	// whatever the filesystem happened to stamp.
	if err := os.Chtimes(path, at, at); err != nil {
		return fmt.Errorf("enginelock: timestamping %s: %w", path, err)
	}
	return nil
}

// Status is what a reader learns about the engine.
type Status struct {
	// Path is where the marker was looked for.
	Path string
	// Running reports a marker touched inside StaleAfter.
	Running bool
	// RefreshedAt is the marker's modification time, zero when there is none.
	RefreshedAt time.Time
	// Marker is the parsed body. Its zero value is what an unparseable or absent
	// marker produces, and Binary.Known() is false there — which reads as "we
	// could not look", never as "the build moved on".
	Marker Marker
}

// Stale reports that the running engine was started from a different executable
// than the one installed now.
//
// It answers "is the process still the installed build?" from the engine's own
// statement about itself, which is the arrangement the dashboard already uses for
// the soak: no process table to scrape, no pid to guess at. An unknown stamp on
// either side is not a difference (binstamp.Stamp.Same), so a marker written by
// an older build reads as unknown rather than as stale.
func (s Status) Stale(installed binstamp.Stamp) bool {
	return s.Running && s.Marker.Binary.Known() && !s.Marker.Binary.Same(installed)
}

// Read reports what the marker at path says.
//
// Every failure direction is "no engine": a missing file, an unreadable one and an
// old one all answer the same, because the reader is a dashboard and an autostart
// script and neither may be stopped by this package's problems.
func Read(path string, now time.Time) Status {
	s := Status{Path: path}
	if strings.TrimSpace(path) == "" {
		return s
	}
	fresh, modified := runlock.Fresh(path, now, StaleAfter)
	s.Running, s.RefreshedAt = fresh, modified
	if !fresh {
		return s
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	var m Marker
	if err := json.Unmarshal(body, &m); err != nil {
		// A running engine whose marker we cannot parse is still a running engine.
		return s
	}
	s.Marker = m
	return s
}

// MarkerPath is where the marker lives for a given journal directory.
func MarkerPath(dir string) string { return filepath.Join(dir, MarkerFileName) }

// LockPath is where the exclusion lives for a given journal directory.
func LockPath(dir string) string { return filepath.Join(dir, LockFileName) }
