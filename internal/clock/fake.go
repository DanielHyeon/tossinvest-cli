package clock

import (
	"context"
	"sync"
	"time"
)

// Fake is a deterministic Clock for tests. Time only moves when a test moves it,
// so a stabilisation loop or a staleness threshold can be driven to its exact
// boundary instead of being approximated with wall-clock sleeps.
//
// Fake is safe for concurrent use: the execution path runs its polling loops in
// goroutines and those tests run under -race.
type Fake struct {
	mu       sync.Mutex
	now      time.Time
	sleepers []*fakeSleeper
}

type fakeSleeper struct {
	deadline time.Time
	wake     chan struct{}
}

// NewFake returns a Fake positioned at now (normalised to UTC).
func NewFake(now time.Time) *Fake { return &Fake{now: now.UTC()} }

// Now returns the fake's current instant in UTC.
func (f *Fake) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

// Since returns the fake's elapsed time since t.
func (f *Fake) Since(t time.Time) time.Duration { return f.Now().Sub(t) }

// Sleep blocks until the fake's time reaches the deadline or ctx is done. A
// non-positive duration returns immediately.
//
// A sleep that is never reached by an Advance blocks forever by design: a test
// that forgets to advance should hang visibly rather than pass on wall time.
func (f *Fake) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	f.mu.Lock()
	s := &fakeSleeper{deadline: f.now.Add(d), wake: make(chan struct{})}
	f.sleepers = append(f.sleepers, s)
	f.mu.Unlock()

	select {
	case <-ctx.Done():
		f.drop(s)
		return ctx.Err()
	case <-s.wake:
		return nil
	}
}

// Advance moves the fake forward by d and wakes every sleeper whose deadline has
// been reached.
func (f *Fake) Advance(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	woken := f.collectDue()
	f.mu.Unlock()
	wakeAll(woken)
}

// Set jumps the fake to an absolute instant (normalised to UTC) and wakes every
// sleeper whose deadline has been reached. Moving backwards is allowed so tests
// can exercise clock-skew handling.
func (f *Fake) Set(t time.Time) {
	f.mu.Lock()
	f.now = t.UTC()
	woken := f.collectDue()
	f.mu.Unlock()
	wakeAll(woken)
}

// Sleepers reports how many goroutines are currently blocked in Sleep.
func (f *Fake) Sleepers() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sleepers)
}

// WaitForSleepers blocks until at least n goroutines are parked in Sleep, or
// within elapses. It exists so a test can advance time only after the code under
// test has actually reached its sleep, which removes the classic
// "advance-too-early" flake.
func (f *Fake) WaitForSleepers(n int, within time.Duration) bool {
	deadline := time.Now().Add(within)
	for {
		if f.Sleepers() >= n {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(time.Millisecond)
	}
}

// collectDue removes and returns every sleeper whose deadline is at or before
// the current fake time. Callers must hold f.mu.
func (f *Fake) collectDue() []*fakeSleeper {
	var due []*fakeSleeper
	kept := f.sleepers[:0]
	for _, s := range f.sleepers {
		if !s.deadline.After(f.now) {
			due = append(due, s)
			continue
		}
		kept = append(kept, s)
	}
	f.sleepers = kept
	return due
}

// drop removes a sleeper that gave up (context cancellation).
func (f *Fake) drop(target *fakeSleeper) {
	f.mu.Lock()
	defer f.mu.Unlock()
	kept := f.sleepers[:0]
	for _, s := range f.sleepers {
		if s != target {
			kept = append(kept, s)
		}
	}
	f.sleepers = kept
}

func wakeAll(sleepers []*fakeSleeper) {
	for _, s := range sleepers {
		close(s.wake)
	}
}
