package soak_test

// pause_test.go covers the one place the survey yields (task 1.7 ③).
//
// The soak and the supervised live verification share an account, a credential
// and a rate limit. On 2026-07-26 they overlapped and the verification lost three
// steps to a 429 (measurements.md M4). The survey is the cheap one — there will be
// another cycle in fifteen minutes — so it waits.

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/soak"
)

// pauseSwitch is a PauseWhile a test can flip from another goroutine.
type pauseSwitch struct {
	mu     sync.Mutex
	paused bool
	asked  int
	reason string
}

func (p *pauseSwitch) fn() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.asked++
	return p.paused, p.reason
}

func (p *pauseSwitch) set(paused bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = paused
}

func (p *pauseSwitch) timesAsked() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.asked
}

// countingReads counts the one read every cycle begins with, so "the survey sent
// nothing while it was paused" is an assertion rather than a hope.
type countingReads struct {
	soak.Reads
	mu sync.Mutex
	n  int
}

func counting(r soak.Reads) *countingReads { return &countingReads{Reads: r} }

func (c *countingReads) Accounts(ctx context.Context) ([]string, error) {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
	return c.Reads.Accounts(ctx)
}

func (c *countingReads) accountsCalls() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

// progressBuffer collects the survey's operator lines under -race.
type progressBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (w *progressBuffer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.Write(p)
}

func (w *progressBuffer) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.b.String()
}

// TestACycleIsHeldBackWhileAVerificationIsRunning, and released when it ends.
func TestACycleIsHeldBackWhileAVerificationIsRunning(t *testing.T) {
	reads := counting(newStubReads())
	gate := &pauseSwitch{paused: true, reason: "a live verification is running"}
	var progress progressBuffer

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Cycles = 1
		o.PauseWhile = gate.fn
		o.Progress = &progress
	})

	done := make(chan error, 1)
	go func() { done <- r.Run(context.Background()) }()

	// The runner must park in the pause poll, not run the cycle.
	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the runner never parked; it ran the cycle straight through the pause")
	}
	if got := reads.accountsCalls(); got != 0 {
		t.Fatalf("the survey made %d account read(s) while paused; it must send nothing", got)
	}

	gate.set(false)
	fake.Advance(soak.PausePoll)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the survey never resumed after the pause cleared")
	}

	if got := reads.accountsCalls(); got == 0 {
		t.Error("the survey did not run its cycle after the pause cleared")
	}
	log := progress.String()
	if !strings.Contains(log, "paused") {
		t.Errorf("the survey paused without saying so:\n%s", log)
	}
	if !strings.Contains(log, "a live verification is running") {
		t.Errorf("the pause log does not say why:\n%s", log)
	}
	if !strings.Contains(log, "resuming") {
		t.Errorf("the survey resumed without saying so:\n%s", log)
	}
}

// TestAnUnpausedSurveyIsUntouched — the default path must not gain a poll, a
// sleep or a line of output.
func TestAnUnpausedSurveyIsUntouched(t *testing.T) {
	reads := counting(newStubReads())
	gate := &pauseSwitch{paused: false}
	var progress progressBuffer

	r, _ := newRunner(t, reads, func(o *soak.Options) {
		o.Cycles = 2
		o.PauseWhile = gate.fn
		o.Progress = &progress
	})
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := gate.timesAsked(); got != 2 {
		t.Errorf("the pause was consulted %d time(s), want once per cycle (2)", got)
	}
	if strings.Contains(progress.String(), "paused") {
		t.Errorf("an unpaused survey logged a pause:\n%s", progress.String())
	}
}

// TestNoPauseHookIsTheOldBehaviour. Every existing caller and every existing
// test passes no hook; the survey must behave exactly as it did.
func TestNoPauseHookIsTheOldBehaviour(t *testing.T) {
	reads := counting(newStubReads())
	r, _ := newRunner(t, reads, func(o *soak.Options) { o.Cycles = 1 })
	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := reads.accountsCalls(); got == 0 {
		t.Error("a survey with no pause hook did not run")
	}
}

// TestAPausedSurveyStillStopsOnCtrlC. A soak parked on a pause must not need the
// verification to finish before it can be killed.
func TestAPausedSurveyStillStopsOnCtrlC(t *testing.T) {
	reads := counting(newStubReads())
	gate := &pauseSwitch{paused: true, reason: "a live verification is running"}

	r, fake := newRunner(t, reads, func(o *soak.Options) {
		o.Cycles = 0
		o.PauseWhile = gate.fn
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()

	if !fake.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the runner never parked in the pause poll")
	}
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("Run returned nil after being cancelled while paused")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a paused survey ignored the cancellation")
	}
	if got := reads.accountsCalls(); got != 0 {
		t.Errorf("the survey made %d read(s) despite never leaving the pause", got)
	}
}
