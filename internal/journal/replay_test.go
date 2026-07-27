package journal

// Replay bookkeeping tests (extend-execution-contract task 2.1/2.2/2.4).
//
// The properties under test are the ones the cap is worth having for: the count
// is durable, it is taken before the send, it cannot be raced past the cap, and
// the one answer that proves nothing (`409 request-in-progress`) gives it back
// without giving back the wait.

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

var replayStart = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

// openReplayJournal opens a journal whose clock the test can advance, which is
// what makes last_replay_at observable.
func openReplayJournal(t *testing.T) (*Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(replayStart)
	j, err := Open(context.Background(), Options{
		Path:     filepath.Join(t.TempDir(), "journal.db"),
		Clock:    clk,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j, clk
}

// inDoubtAttempt drives one attempt to IN_DOUBT, which is the only state a
// replay is defined for.
func inDoubtAttempt(t *testing.T, j *Journal, intentID, attemptID string) *Attempt {
	t.Helper()
	ctx := context.Background()
	intent := testIntent()
	intent.ID = intentID
	a, err := j.Prepare(ctx, PrepareRequest{Intent: intent, Kind: KindPlace, AttemptID: attemptID})
	if err != nil {
		t.Fatalf("Prepare(%s): %v", attemptID, err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(%s): %v", attemptID, err)
	}
	if err := a.MarkInDoubt(ctx, "test", "the outcome is unknown"); err != nil {
		t.Fatalf("MarkInDoubt(%s): %v", attemptID, err)
	}
	return a
}

func TestMarkReplayStartedCountsAndStampsTheAttempt(t *testing.T) {
	j, clk := openReplayJournal(t)
	ctx := context.Background()
	inDoubtAttempt(t, j, "intent-1", "attempt-1")

	clk.Advance(90 * time.Second)
	state, err := j.MarkReplayStarted(ctx, "attempt-1", 2)
	if err != nil {
		t.Fatalf("MarkReplayStarted: %v", err)
	}
	if state.Count != 1 {
		t.Errorf("count: got %d, want 1", state.Count)
	}
	want := RFC3339(replayStart.Add(90 * time.Second))
	if state.LastAt != want {
		t.Errorf("last replay at: got %q, want %q", state.LastAt, want)
	}

	// Durable, not just returned: the cap has to survive the restart that a lost
	// response so often comes with.
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.ReplayCount != 1 || rec.LastReplayAt != want {
		t.Errorf("stored bookkeeping: count=%d lastAt=%q", rec.ReplayCount, rec.LastReplayAt)
	}
}

func TestMarkReplayStartedRefusesPastTheCap(t *testing.T) {
	j, _ := openReplayJournal(t)
	ctx := context.Background()
	inDoubtAttempt(t, j, "intent-1", "attempt-1")

	for i := 1; i <= 2; i++ {
		if _, err := j.MarkReplayStarted(ctx, "attempt-1", 2); err != nil {
			t.Fatalf("replay %d: %v", i, err)
		}
	}
	state, err := j.MarkReplayStarted(ctx, "attempt-1", 2)
	if !errors.Is(err, ErrReplayCapReached) {
		t.Fatalf("third replay: got %v, want ErrReplayCapReached", err)
	}
	if state.Count != 2 {
		t.Errorf("count after the refusal: got %d, want it left at 2", state.Count)
	}
}

func TestMarkReplayStartedRefusesAnAttemptThatIsNotInDoubt(t *testing.T) {
	j, _ := openReplayJournal(t)
	ctx := context.Background()
	intent := testIntent()
	if _, err := j.Prepare(ctx, PrepareRequest{Intent: intent, Kind: KindPlace, AttemptID: "attempt-1"}); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	if _, err := j.MarkReplayStarted(ctx, "attempt-1", 2); !errors.Is(err, ErrReplayNotInDoubt) {
		t.Fatalf("RECORDED attempt: got %v, want ErrReplayNotInDoubt", err)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.ReplayCount != 0 || rec.LastReplayAt != "" {
		t.Errorf("a refused replay must leave no trace: count=%d lastAt=%q", rec.ReplayCount, rec.LastReplayAt)
	}
}

func TestMarkReplayStartedRefusesAMissingAttempt(t *testing.T) {
	j, _ := openReplayJournal(t)
	if _, err := j.MarkReplayStarted(context.Background(), "nope", 2); !errors.Is(err, ErrAttemptNotFound) {
		t.Fatalf("got %v, want ErrAttemptNotFound", err)
	}
}

// TestRefundReplayReturnsTheCountButKeepsTheWait is the `409 request-in-progress`
// rule: the cap is not consumed, but the minimum interval still is — the answer
// to "still processing" is to wait, not to hammer.
func TestRefundReplayReturnsTheCountButKeepsTheWait(t *testing.T) {
	j, clk := openReplayJournal(t)
	ctx := context.Background()
	inDoubtAttempt(t, j, "intent-1", "attempt-1")

	clk.Advance(30 * time.Second)
	if _, err := j.MarkReplayStarted(ctx, "attempt-1", 2); err != nil {
		t.Fatalf("MarkReplayStarted: %v", err)
	}
	stampedAt := RFC3339(replayStart.Add(30 * time.Second))

	clk.Advance(time.Second)
	state, err := j.RefundReplay(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("RefundReplay: %v", err)
	}
	if state.Count != 0 {
		t.Errorf("count after the refund: got %d, want 0", state.Count)
	}
	if state.LastAt != stampedAt {
		t.Errorf("last replay at: got %q, want it untouched at %q", state.LastAt, stampedAt)
	}
}

func TestRefundReplayNeverGoesNegative(t *testing.T) {
	j, _ := openReplayJournal(t)
	ctx := context.Background()
	inDoubtAttempt(t, j, "intent-1", "attempt-1")

	state, err := j.RefundReplay(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("RefundReplay: %v", err)
	}
	if state.Count != 0 {
		t.Errorf("count: got %d, want 0", state.Count)
	}
}

// TestMutationsDispatchedSinceSeesOnlyRealOtherTraffic pins the contamination
// query the absence judgement depends on: another attempt on the same symbol
// counts, the attempt under resolution does not, a different symbol does not,
// and neither does an attempt that carries a dispatch stamp but provably sent
// nothing.
func TestMutationsDispatchedSinceSeesOnlyRealOtherTraffic(t *testing.T) {
	j, clk := openReplayJournal(t)
	ctx := context.Background()

	subject := inDoubtAttempt(t, j, "intent-1", "attempt-1")
	_ = subject
	windowStart := RFC3339(clk.Now())

	clk.Advance(10 * time.Second)
	inDoubtAttempt(t, j, "intent-2", "attempt-2") // same symbol, inside the window

	// A different symbol is not contamination.
	other := testIntent()
	other.ID = "intent-3"
	other.Symbol = "MSFT"
	otherAttempt, err := j.Prepare(ctx, PrepareRequest{Intent: other, Kind: KindPlace, AttemptID: "attempt-3"})
	if err != nil {
		t.Fatalf("Prepare(attempt-3): %v", err)
	}
	if err := otherAttempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(attempt-3): %v", err)
	}

	// Dispatch-stamped but refused before a byte left: not contamination.
	unsent := testIntent()
	unsent.ID = "intent-4"
	unsentAttempt, err := j.Prepare(ctx, PrepareRequest{Intent: unsent, Kind: KindPlace, AttemptID: "attempt-4"})
	if err != nil {
		t.Fatalf("Prepare(attempt-4): %v", err)
	}
	if err := unsentAttempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(attempt-4): %v", err)
	}
	if err := unsentAttempt.Settle(ctx, StateNotDispatched, "test", "refused before sending"); err != nil {
		t.Fatalf("Settle(attempt-4): %v", err)
	}

	found, err := j.MutationsDispatchedSince(ctx, "us", "AAPL", windowStart, "attempt-1")
	if err != nil {
		t.Fatalf("MutationsDispatchedSince: %v", err)
	}
	if len(found) != 1 || found[0].ID != "attempt-2" {
		ids := make([]string, 0, len(found))
		for _, rec := range found {
			ids = append(ids, rec.ID)
		}
		t.Fatalf("contaminating attempts: got %v, want [attempt-2]", ids)
	}
}
