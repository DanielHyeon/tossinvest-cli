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
//	         status, the autostart script reads it before spawning anything, and
//	         since a102 the console's boot survey reads it to decide whether the
//	         account is already reconciled.
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
// believed for at most five minutes.
//
// # What a wrong marker costs, after a102
//
// It used to be a misdrawn status line or a refused autostart, never a second
// writer — the flock closes that, and it still does. ReadyAt changed the other
// half. The console's boot survey now decides *when to spend the account's rate
// budget* from this file: a marker that says ready sends a read-only survey at
// the broker while the engine's restart recovery may still be reading, and that
// contention is what killed an engine one second into its life on 2026-08-13.
//
// So the failure directions are no longer symmetric, and the reader is written
// for that. Absence, staleness, an unparseable body, a missing ProcInstance and
// an unreadable process table all mean "not ready", which costs a bounded wait
// and then starts the survey anyway. Only a fresh marker whose PID *and* process
// instance both match a live engine is read as protection standing.
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
// It is JSON because things parse it — the console renders some of these fields,
// the boot survey decides on others, and a human reads the note; internal/runlock's
// marker is prose because nothing parses it. The reader is deliberately lenient —
// an unparseable marker is a running engine with an unknown build, not an error —
// so a format change can never turn into a dashboard that refuses to draw. That
// leniency has a cost after a102 and it is paid in the safe direction: a body the
// reader cannot parse loses ReadyAt too, which reads as "not ready" and delays a
// survey rather than racing a recovery.
type Marker struct {
	PID         int            `json:"pid"`
	StartedAt   time.Time      `json:"started_at"`
	RefreshedAt time.Time      `json:"refreshed_at"`
	Binary      binstamp.Stamp `json:"binary,omitempty"`
	Note        string         `json:"note,omitempty"`
	// ReadyAt is when this engine finished its restart recovery, and it is a
	// pointer because absence is the answer "not ready yet" (openspec change
	// a102).
	//
	// It is not StartedAt and it is not the file's existence: both of those
	// happen *before* recovery runs — the marker is written at boot step 6 and
	// the recovery is step 7 — so an autostart survey that starts on either one
	// spends the same account's rate budget the recovery is spending. The whole
	// point of a separate field is that only a finished recovery may write it.
	//
	// 없는 필드는 nil이다. a102 이전 빌드가 쓴 마커도 그대로 읽히고 "준비 안 됨"이 된다.
	ReadyAt *time.Time `json:"ready_at,omitempty"`
	// ProcInstance identifies *which run* of PID wrote this marker. It is an
	// opaque string this package neither builds nor interprets — the caller
	// supplies it through [Held.Identify], because deciding what identifies a
	// process instance is operating-system knowledge and this package has none
	// (the same line MarkerFileName's reader draws).
	//
	// It exists because PID is not an instance. A container recreate keeps the
	// journal volume, so a predecessor's marker stays fresh for StaleAfter with
	// its ready_at and its PID; the recreated PID namespace hands out PIDs
	// near-deterministically, so the replacement engine can be given the very
	// same number. Between the moment that replacement becomes visible and the
	// moment it writes its own marker, "the marker's PID is alive" is true about
	// the wrong process — and what it certifies is that the account is
	// protected (a102, gstack review P1).
	//
	// Absent means "cannot tell", which reads as not ready.
	ProcInstance string `json:"proc_instance,omitempty"`
}

// markerNote is written into every marker so somebody who finds the file with
// `cat` knows what it is and what deleting it does.
const markerNote = "advisory only. The exclusion is the flock on engine.lock; this file exists so the " +
	"console can draw the engine's status and the autostart script can check before spawning. " +
	"A stale one (older than 5m) is ignored. " +
	"ready_at is written only after the engine finishes its restart recovery, and the console's " +
	"boot survey waits for it before spending the account's rate budget: clear it or edit it and " +
	"the survey stops waiting (an invented one) or waits out its whole cap and starts anyway " +
	"(a removed one). proc_instance says which run of this pid wrote the file; delete the file " +
	"rather than hand-editing either field."

// Held is a marker this process is publishing.
//
// It replaces the bare release function Hold used to return because the marker
// grew a second thing a caller can say about itself — [Held.Ready] — and a
// second return value would have left the two able to disagree about which
// marker they meant (openspec change a102).
//
// A Held whose first write failed is inert: Release removes nothing and Ready
// publishes nothing. That is the same contract the old no-op release carried,
// and it matters — the engine does not refuse to start over a marker it could
// not write (cmd/tossctl/engine.go step 6), so an inert handle is a state the
// production caller reaches and then keeps using.
type Held struct {
	path string

	// mu guards marker, because Ready and the refresh goroutine both touch it.
	// Before a102 nothing mutated the marker after Hold returned, so the
	// goroutine could simply capture it by value.
	mu     sync.Mutex
	marker Marker
	live   bool

	done chan struct{}
	once sync.Once
}

// Release drops the marker. It is safe to call more than once, and safe on an
// inert handle.
//
// Clearing live under the mutex is what stops a refresh already inside the
// ticker's case from writing the file back after the remove. A102 A2 reproduced
// that resurrection deterministically, and after this change what a resurrected
// marker says is not "a misdrawn status line" but "protection is standing" about
// a process that has stopped.
func (h *Held) Release() {
	if h == nil {
		return
	}
	h.mu.Lock()
	if !h.live {
		h.mu.Unlock()
		return
	}
	h.live = false
	h.mu.Unlock()

	h.once.Do(func() {
		close(h.done)
		if rerr := os.Remove(h.path); rerr != nil && !errors.Is(rerr, os.ErrNotExist) {
			// A marker that could not be removed goes stale on its own within
			// StaleAfter, which is the whole reason freshness is a timestamp
			// rather than existence.
			return
		}
	})
}

// Ready publishes "this engine has finished its restart recovery".
//
// It is idempotent and keeps the first instant: the signal means the moment
// protection came up, and a later call moving it would turn that into "the last
// time somebody said so". Calling it on an inert handle does nothing.
//
// The write happens inside the mutex. Guarding only the struct left the file
// unguarded, and A2 measured what that costs: ready_at erased on disk in 139 of
// 3000 concurrent runs, because two writers can order their struct reads and
// their file writes differently.
func (h *Held) Ready(now time.Time) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.live || h.marker.ReadyAt != nil {
		return
	}
	at := now.UTC()
	h.marker.ReadyAt = &at
	// A ready signal that could not be written costs the console a wait it did
	// not need, never a refused start — the same asymmetry the first write has.
	_ = write(h.path, h.marker, now)
}

// Identify records which run of this PID owns the marker.
//
// The token is opaque here and is built by the caller, which is the only side
// that may know what an operating system considers a process instance. It is a
// separate call rather than an argument to Hold so that Hold's contract — and
// its eight callers — stay as they are, and because the value is only ever
// consulted alongside ReadyAt, which no marker carries until the recovery
// finishes two boot steps later.
//
// Calling it on an inert handle does nothing, and an empty token is ignored: a
// build that cannot compute one leaves the field out, and a reader treats the
// absence as "cannot tell" rather than as agreement.
func (h *Held) Identify(token string) {
	if h == nil || strings.TrimSpace(token) == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.live || h.marker.ProcInstance == token {
		return
	}
	h.marker.ProcInstance = token
	_ = write(h.path, h.marker, h.marker.RefreshedAt)
}

// refresh rewrites the marker at at, from what the holder currently knows.
//
// It is the ticker goroutine's body, named so a test can drive it: RefreshEvery
// is one minute and is a spec number, so the alternative was to make the period
// injectable and let a test move a number the spec fixed.
//
// Reading the marker from guarded state rather than from a captured copy is the
// whole reason this exists: a refresh built from the value captured at Hold time
// would rewrite the file without ready_at, and a ready engine would become
// unready one minute later. The live check is inside the same lock Release takes,
// so a refresh cannot bring back a marker that has just been removed.
func (h *Held) refresh(at time.Time) error {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.live {
		return nil
	}
	return write(h.path, h.marker, at)
}

// Hold writes the marker and keeps it fresh until ctx ends or the returned
// handle is released.
//
// The handle is always non-nil so a caller can defer Release without checking
// the error first: a marker that could not be written is a status line the
// console cannot draw, which is worth mentioning and is not worth refusing a
// start over. The exclusion has already been taken by then.
func Hold(ctx context.Context, path string, now time.Time) (*Held, error) {
	m := Marker{PID: os.Getpid(), StartedAt: now.UTC(), Note: markerNote}
	m.Binary, _ = binstamp.Self()

	h := &Held{path: path, marker: m, done: make(chan struct{})}
	if werr := write(path, m, now); werr != nil {
		return h, werr
	}
	h.mu.Lock()
	h.live = true
	h.mu.Unlock()

	go func() {
		ticker := time.NewTicker(RefreshEvery)
		defer ticker.Stop()
		for {
			select {
			case <-h.done:
				return
			case <-ctx.Done():
				return
			case at := <-ticker.C:
				_ = h.refresh(at)
			}
		}
	}()

	return h, nil
}

// write replaces the marker atomically.
//
// It used to be os.WriteFile, which truncates and then fills. Every reader that
// looked inside that window got a short or empty file, and the reader is a
// dashboard and an autostart script: A2 measured 3617 torn reads in 12259
// attempts against the old shape. A rename over the same directory is atomic, so
// a reader sees either the previous marker or the next one and never a half of
// either — which is exactly the property the lenient reader could not supply for
// itself, because "unparseable" degrades to "running, build unknown" and after
// a102 that includes losing the ready signal.
//
// The temporary file is created in the marker's own directory so the rename
// never crosses a filesystem, and it is removed on every failure path so a
// journal directory does not collect debris from a full disk.
func write(path string, m Marker, at time.Time) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("enginelock: creating %s: %w", dir, err)
	}
	m.RefreshedAt = at.UTC()
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("enginelock: rendering the marker: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".engine-run-*.json")
	if err != nil {
		return fmt.Errorf("enginelock: staging %s: %w", path, err)
	}
	staged := tmp.Name()
	if _, err := tmp.Write(append(body, '\n')); err != nil {
		tmp.Close()
		os.Remove(staged)
		return fmt.Errorf("enginelock: writing %s: %w", staged, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginelock: closing %s: %w", staged, err)
	}
	if err := os.Chmod(staged, 0o600); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginelock: securing %s: %w", staged, err)
	}
	// The contents are for a reader; the modification time is the fact. Setting it
	// explicitly keeps an injected clock authoritative in a test rather than
	// whatever the filesystem happened to stamp — and doing it before the rename
	// means the marker is never visible with the wrong timestamp.
	if err := os.Chtimes(staged, at, at); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginelock: timestamping %s: %w", staged, err)
	}
	if err := os.Rename(staged, path); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginelock: publishing %s: %w", path, err)
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

// Ready reports a live engine that has finished its restart recovery.
//
// Both halves are required and neither is redundant. Running without ReadyAt is
// the boot window this signal exists to name — the engine holds the lock and the
// marker but the account is not yet reconciled. ReadyAt without Running is a
// crashed engine's last word, believed for at most StaleAfter, and reading it as
// "ready" would tell the survey a dead engine is protecting the account.
//
// It is one method rather than two conditions at the call sites so that
// "the engine is ready" has a single definition (openspec change a102).
func (s Status) Ready() bool { return s.Running && s.Marker.ReadyAt != nil }

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
