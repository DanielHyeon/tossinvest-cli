package clock_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// TestSystemClockNowIsUTC pins the contract every journal timestamp depends on:
// the system clock reports UTC so a machine-local timezone can never leak into a
// persisted record.
func TestSystemClockNowIsUTC(t *testing.T) {
	c := clock.System()
	now := c.Now()
	if now.Location() != time.UTC {
		t.Fatalf("System().Now() location: want UTC, got %v", now.Location())
	}
	if d := time.Since(now); d < -time.Second || d > time.Minute {
		t.Fatalf("System().Now() is not close to time.Now(): delta %v", d)
	}
}

func TestSystemLeaseAnchorRetainsMonotonicReading(t *testing.T) {
	c := clock.System()
	anchor := clock.LeaseAnchor(c)
	// Round(0) is Go's documented way to strip a time.Time's monotonic
	// reading. Operator == also compares that opaque reading, unlike Equal.
	if anchor == anchor.Round(0) {
		t.Fatal("system lease anchor lost its monotonic reading")
	}
	if elapsed := clock.LeaseElapsed(c, anchor); elapsed < 0 {
		t.Fatalf("system lease elapsed immediately went backwards: %v", elapsed)
	}
}

// TestSystemClockSleepRespectsContext proves a cancelled context aborts a sleep
// instead of holding an execution-path goroutine for the full duration.
func TestSystemClockSleepRespectsContext(t *testing.T) {
	c := clock.System()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	err := c.Sleep(ctx, time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Sleep error: want context.Canceled, got %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Sleep ignored cancellation: took %v", elapsed)
	}
}

// TestSystemClockSleepShortDuration verifies a real sleep actually advances.
func TestSystemClockSleepShortDuration(t *testing.T) {
	c := clock.System()
	start := c.Now()
	if err := c.Sleep(context.Background(), 5*time.Millisecond); err != nil {
		t.Fatalf("Sleep: %v", err)
	}
	if c.Since(start) <= 0 {
		t.Fatalf("Since(start) = %v, want > 0", c.Since(start))
	}
}

// TestFakeClockNowAdvance covers the deterministic driver used by every
// stabilisation / staleness test in the execution path.
func TestFakeClockNowAdvance(t *testing.T) {
	base := time.Date(2026, 3, 29, 0, 30, 0, 0, time.UTC)
	f := clock.NewFake(base)

	if !f.Now().Equal(base) {
		t.Fatalf("Now: want %v, got %v", base, f.Now())
	}
	f.Advance(90 * time.Second)
	if want := base.Add(90 * time.Second); !f.Now().Equal(want) {
		t.Fatalf("Now after Advance: want %v, got %v", want, f.Now())
	}
	if got := f.Since(base); got != 90*time.Second {
		t.Fatalf("Since: want 90s, got %v", got)
	}
	// Set jumps to an absolute instant and normalises to UTC.
	f.Set(time.Date(2026, 3, 30, 9, 0, 0, 0, time.FixedZone("KST", 9*3600)))
	if f.Now().Location() != time.UTC {
		t.Fatalf("Set must normalise to UTC, got %v", f.Now().Location())
	}
	if want := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC); !f.Now().Equal(want) {
		t.Fatalf("Now after Set: want %v, got %v", want, f.Now())
	}
}

// TestFakeClockSleepWakesOnAdvance is the reason the fake exists: a sleep only
// returns when test-controlled time passes its deadline, never on wall time.
func TestFakeClockSleepWakesOnAdvance(t *testing.T) {
	f := clock.NewFake(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))

	done := make(chan error, 1)
	go func() { done <- f.Sleep(context.Background(), 10*time.Second) }()

	if !f.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("sleeper never registered")
	}
	// Not enough to reach the deadline: the sleeper must stay blocked.
	f.Advance(9 * time.Second)
	select {
	case err := <-done:
		t.Fatalf("Sleep returned early (err=%v) at %v", err, f.Now())
	case <-time.After(50 * time.Millisecond):
	}

	f.Advance(time.Second)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Sleep: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Sleep did not wake after the deadline was reached")
	}
	if f.Sleepers() != 0 {
		t.Fatalf("Sleepers after wake: want 0, got %d", f.Sleepers())
	}
}

// TestFakeClockSleepCancel proves a blocked fake sleep is cancellable, so a
// stalled stabilisation loop can be shut down in tests and in production code
// that shares the interface.
func TestFakeClockSleepCancel(t *testing.T) {
	f := clock.NewFake(time.Unix(0, 0).UTC())
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- f.Sleep(ctx, time.Hour) }()
	if !f.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("sleeper never registered")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("want context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel did not unblock the fake sleep")
	}
	if f.Sleepers() != 0 {
		t.Fatalf("cancelled sleeper was not removed: %d left", f.Sleepers())
	}
}

// TestFakeClockNonPositiveSleep documents that a zero/negative sleep is a no-op
// rather than a permanent block.
func TestFakeClockNonPositiveSleep(t *testing.T) {
	f := clock.NewFake(time.Unix(0, 0).UTC())
	if err := f.Sleep(context.Background(), 0); err != nil {
		t.Fatalf("Sleep(0): %v", err)
	}
	if err := f.Sleep(context.Background(), -time.Second); err != nil {
		t.Fatalf("Sleep(-1s): %v", err)
	}
}

// TestFakeClockConcurrentAccess is the -race guard: many goroutines read the
// clock and sleep while another advances it.
func TestFakeClockConcurrentAccess(t *testing.T) {
	f := clock.NewFake(time.Unix(0, 0).UTC())
	var wg sync.WaitGroup

	// Sleepers register first so the advancer below is guaranteed to cross their
	// deadlines; a sleep whose deadline is never reached blocks by design.
	const sleepers = 4
	for i := 0; i < sleepers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = f.Sleep(context.Background(), time.Hour)
		}()
	}
	if !f.WaitForSleepers(sleepers, 5*time.Second) {
		t.Fatalf("only %d/%d sleepers registered", f.Sleepers(), sleepers)
	}

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = f.Now()
				_ = f.Since(time.Unix(0, 0).UTC())
				_ = f.Sleepers()
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 200; j++ {
			f.Advance(time.Second)
		}
		// Past every sleeper's deadline: 200s < 1h, so finish the job.
		f.Advance(time.Hour)
	}()
	wg.Wait()
	if f.Sleepers() != 0 {
		t.Fatalf("Sleepers after all deadlines passed: want 0, got %d", f.Sleepers())
	}
}

// TestFakeClockSatisfiesInterface keeps the fake substitutable for the real one.
func TestFakeClockSatisfiesInterface(t *testing.T) {
	var c clock.Clock = clock.NewFake(time.Unix(0, 0).UTC())
	if c.Now().IsZero() {
		t.Fatal("fake clock returned the zero time")
	}
}
