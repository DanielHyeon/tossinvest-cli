package engine_test

// a064 task 2: the exit loop reports its failed cycles.
//
// ExitCycle.Err's own declaration says "It is reported and not returned by Run".
// Before this change nothing reported it: Run discarded the whole cycle
// (`_ = o.ObserveOnce(ctx)`) and the exit loop is registered in the runtime with
// no Health, so the supervisor could not count cycle failures either. Every
// failure this loop had — a judgement it could not record, a quarantine it just
// created, a working set it could not build — was written to a struct field that
// nothing read.

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// cycleLog captures the loop's structured lines.
type cycleLog struct{ buf *bytes.Buffer }

func newCycleLog() (*cycleLog, *obs.Logger) {
	buf := &bytes.Buffer{}
	return &cycleLog{buf: buf}, obs.NewLogger(obs.LogOptions{
		Writer: buf, JSON: true, Level: slog.LevelDebug,
	})
}

func (c *cycleLog) lines(event obs.EventType) []map[string]any {
	var out []map[string]any
	for _, raw := range strings.Split(strings.TrimSpace(c.buf.String()), "\n") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		line := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			continue
		}
		if line["event"] == string(event) {
			out = append(out, line)
		}
	}
	return out
}

// runOneCycle drives Run for exactly one cycle and then cancels it, so the test
// exercises the production entry point rather than ObserveOnce.
func runOneCycle(t *testing.T, h *exitHarness) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- h.observer.Run(ctx) }()
	// The fake clock's Sleep is what the loop parks on after its first cycle, so
	// a registered sleeper means exactly one cycle has completed.
	if !h.clk.WaitForSleepers(1, 3*time.Second) {
		t.Fatal("the loop never reached its first sleep")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
}

func TestRunReportsAFailedCycle(t *testing.T) {
	capture, logger := newCycleLog()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Log = logger })
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()

	// The two records of the entry price disagree, so the judgement transaction
	// quarantines the generation and returns the sentinel — a cycle failure.
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET entry_price='69000' WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 70900)

	runOneCycle(t, h)

	lines := capture.lines(obs.EventExitCycleFailed)
	if len(lines) != 1 {
		t.Fatalf("cycle failure lines = %d, want one:\n%s", len(lines), capture.buf.String())
	}
	if !strings.Contains(strings.ToLower(capture.buf.String()), "quarantin") {
		t.Fatalf("the reported line does not carry the cause:\n%s", capture.buf.String())
	}
}

func TestRunSaysNothingAboutASuccessfulCycle(t *testing.T) {
	capture, logger := newCycleLog()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Log = logger })
	openLadderState(t, h, "005930")
	h.quote("005930", 70500)

	runOneCycle(t, h)

	if lines := capture.lines(obs.EventExitCycleFailed); len(lines) != 0 {
		t.Fatalf("a healthy cycle produced %d failure line(s):\n%s", len(lines), capture.buf.String())
	}
}

func TestAFailedCycleAloneRaisesNoAlert(t *testing.T) {
	// Design D1: cycle failures are logged, not alerted. A transient ledger error
	// must not be able to enqueue a critical alert, fail to deliver it, latch the
	// entry gate and escalate a live account to ENTRY_BLOCKED.
	_, logger := newCycleLog()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Log = logger })
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET entry_price='69000' WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 70900)

	before := len(h.alerts.events)
	runOneCycle(t, h)

	// The quarantine announcement is its own named condition and is expected. What
	// must not appear is an alert whose subject is "the cycle failed".
	for _, e := range h.alerts.events[before:] {
		if e.Type != obs.EventExitSnapshotQuarantined {
			t.Fatalf("a failed cycle raised %s", e.Type)
		}
	}
	if got := h.mode(); got == journal.ModeEntryBlocked {
		t.Fatal("a failed cycle escalated the operating mode on its own")
	}
}

func TestRunKeepsGoingAfterAFailedCycle(t *testing.T) {
	// The landed contract: Run returns the context's error and never a
	// judgement's. A loop that exited on a failed cycle would remove the
	// protection it exists to provide.
	_, logger := newCycleLog()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Log = logger })
	p := openLadderState(t, h, "005930")
	h.quote("005930", 70500)
	h.observe()
	db := openLedger(t, h)
	if _, err := db.Exec(`UPDATE exit_states SET entry_price='69000' WHERE position_id=?`, p.ID); err != nil {
		t.Fatal(err)
	}
	h.quote("005930", 70900)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- h.observer.Run(ctx) }()
	if !h.clk.WaitForSleepers(1, 3*time.Second) {
		t.Fatal("the loop never reached its first sleep")
	}
	// It parked on the interval instead of returning: the failed cycle did not
	// end the loop.
	cancel()
	select {
	case err := <-done:
		if !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("Run returned %v, want the context's error", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Run did not return after cancellation")
	}
}

func TestRunReturnsTheContextError(t *testing.T) {
	// B2: the loop checks the context before every cycle, so an already-cancelled
	// context produces no observation at all. This is the graceful stop —
	// engine-safety: 컨텍스트 취소에 의한 반환은 정상 종료다.
	_, logger := newCycleLog()
	h := newExitHarness(t, func(o *engine.ExitObserverOptions) { o.Log = logger })
	openLadderState(t, h, "005930")
	h.quote("005930", 70500)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	before := len(h.alerts.events)
	err := h.observer.Run(ctx)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("Run returned %v, want the context's error", err)
	}
	if len(h.alerts.events) != before {
		t.Fatal("a cancelled Run still observed a cycle")
	}
}
