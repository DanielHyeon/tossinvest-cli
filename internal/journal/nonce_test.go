package journal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// Durable one-shot nonce tests (extend-execution-contract task 1.7).
//
// The requirement is "결정 nonce의 durable 저장": a restart must not lose a
// consumption record, a persisted decision snapshot must not be usable twice,
// the record must be written in the same transaction as the dispatch, and it
// must outlive the longest decision TTL.

// cancelDecision records a decision with no idempotency key, so two attempts can
// be prepared against it — which is what makes the one-shot property testable
// without the key's UNIQUE index refusing the second one first.
func cancelDecision(t *testing.T, j *Journal, id string) Decision {
	t.Helper()
	issued := testIssued(t)
	dec, err := j.RecordDecision(context.Background(), DecisionRequest{
		ID: id, AccountRef: "acct-1", SafetyClass: SafetyClassRiskReducing, Kind: KindCancel,
		Preimage: ReductionIntent{
			AccountRef: "acct-1", Market: "us", Symbol: "AAPL", Side: "SELL",
			MaxQuantity: "10", Reason: "test",
		},
		Nonce: "nonce-" + id, IssuedAt: issued, ExpiresAt: issued.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}
	return dec
}

func boundAttempt(t *testing.T, j *Journal, dec Decision, intentID, attemptID string) *Attempt {
	t.Helper()
	req := testRequest()
	req.Intent.ID = intentID
	req.AttemptID = attemptID
	req.Kind = KindCancel
	req.TargetOrderID = "order-1"
	req.DecisionID = dec.ID
	req.SafetyClass = dec.SafetyClass
	a, err := j.Prepare(context.Background(), req)
	if err != nil {
		t.Fatalf("Prepare(%s): %v", attemptID, err)
	}
	return a
}

func spentNonceCount(t *testing.T, j *Journal) int {
	t.Helper()
	var n int
	if err := j.db.QueryRowContext(context.Background(),
		"SELECT count(*) FROM spent_nonces").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestNonceIsSpentInTheDispatchTransaction: the consumption record and the
// DISPATCH_STARTED transition are one write. Two writes would have a window in
// which a crash leaves a dispatched attempt whose decision is still unspent.
func TestNonceIsSpentInTheDispatchTransaction(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	dec := cancelDecision(t, j, "dec-1")
	a := boundAttempt(t, j, dec, "intent-1", "attempt-1")

	if spentNonceCount(t, j) != 0 {
		t.Fatal("nothing is spent before the dispatch is recorded")
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}

	var nonce, decisionID, consumedAt string
	if err := j.db.QueryRowContext(ctx,
		"SELECT nonce, decision_id, consumed_at FROM spent_nonces").
		Scan(&nonce, &decisionID, &consumedAt); err != nil {
		t.Fatalf("reading the consumption record: %v", err)
	}
	if nonce != dec.Nonce || decisionID != dec.ID {
		t.Errorf("consumption record = (%q, %q), want (%q, %q)", nonce, decisionID, dec.Nonce, dec.ID)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if consumedAt != rec.DispatchStartedAt {
		t.Errorf("consumed_at %q != dispatch_started_at %q — the two are one transaction",
			consumedAt, rec.DispatchStartedAt)
	}
}

// TestSecondDispatchOnOneDecisionIsRefused is the one-shot property, and the
// refusal has to leave nothing behind: the second attempt stays RECORDED, and
// there is still exactly one consumption record.
func TestSecondDispatchOnOneDecisionIsRefused(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	dec := cancelDecision(t, j, "dec-1")

	first := boundAttempt(t, j, dec, "intent-1", "attempt-1")
	if err := first.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("first MarkDispatchStarted: %v", err)
	}

	second := boundAttempt(t, j, dec, "intent-2", "attempt-2")
	err := second.MarkDispatchStarted(ctx)
	if !errors.Is(err, ErrNonceSpent) {
		t.Fatalf("want ErrNonceSpent, got %v", err)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-2")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateRecorded {
		t.Errorf("state = %s, want RECORDED — the refused transition must roll back", rec.State)
	}
	if rec.DispatchStartedAt != "" {
		t.Errorf("dispatch_started_at = %q, want empty", rec.DispatchStartedAt)
	}
	if got := len(attemptHistory(t, j, "attempt-2")); got != 1 {
		t.Errorf("history entries = %d, want only the RECORDED one", got)
	}
	if got := spentNonceCount(t, j); got != 1 {
		t.Errorf("consumption records = %d, want 1", got)
	}
}

// TestSpentNonceSurvivesARestart is the spec's "재시작 후 결정 재사용 시도": the
// process is gone, the decision snapshot is still on disk, and it is still spent.
func TestSpentNonceSurvivesARestart(t *testing.T) {
	path := tempJournalPath(t)
	ctx := context.Background()
	clk := clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC))

	j := openTestJournalAt(t, path)
	dec := cancelDecision(t, j, "dec-1")
	first := boundAttempt(t, j, dec, "intent-1", "attempt-1")
	if err := first.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(ctx, Options{
		Path: path, Clock: clk,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if err != nil {
		t.Fatalf("reopening: %v", err)
	}
	t.Cleanup(func() { _ = reopened.Close() })

	spent, err := reopened.NonceSpent(ctx, dec.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	if !spent {
		t.Fatal("a restart must not lose the consumption record")
	}
	second := boundAttempt(t, reopened, dec, "intent-2", "attempt-2")
	if err := second.MarkDispatchStarted(ctx); !errors.Is(err, ErrNonceSpent) {
		t.Fatalf("want ErrNonceSpent after the restart, got %v", err)
	}
}

// TestAttemptWithoutADecisionSpendsNothing: the paths that predate the decision
// contract have no nonce, and a dispatch that consumed nothing is not a refusal
// to make here — the gateway refuses those on its own terms.
func TestAttemptWithoutADecisionSpendsNothing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if got := spentNonceCount(t, j); got != 0 {
		t.Fatalf("consumption records = %d, want 0", got)
	}
}

// TestPruneRefusesToOutliveNothing is the retention invariant made executable:
// no decision may survive the record of its own consumption, so a retention
// shorter than the longest TTL on disk deletes nothing and says why.
func TestPruneRefusesToOutliveNothing(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	dec := cancelDecision(t, j, "dec-1") // TTL of one minute
	a := boundAttempt(t, j, dec, "intent-1", "attempt-1")
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}

	longest, err := j.MaxDecisionTTL(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if longest != time.Minute {
		t.Fatalf("longest TTL = %s, want 1m", longest)
	}

	now := testIssued(t).Add(time.Hour)
	removed, err := j.PruneSpentNonces(ctx, now, 30*time.Second)
	if !errors.Is(err, ErrRetentionTooShort) {
		t.Fatalf("want ErrRetentionTooShort, got %v", err)
	}
	if removed != 0 {
		t.Errorf("removed = %d, want 0", removed)
	}
	if got := spentNonceCount(t, j); got != 1 {
		t.Fatalf("consumption records = %d, want 1 — a refused prune deletes nothing", got)
	}
}

// TestPruneRemovesRecordsOlderThanTheRetention: with a retention that clears the
// invariant, old records go and recent ones stay.
func TestPruneRemovesRecordsOlderThanTheRetention(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	dec := cancelDecision(t, j, "dec-1")
	a := boundAttempt(t, j, dec, "intent-1", "attempt-1")
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}

	issued := testIssued(t)
	// Still inside the retention window: nothing goes.
	removed, err := j.PruneSpentNonces(ctx, issued.Add(30*time.Minute), time.Hour)
	if err != nil {
		t.Fatalf("PruneSpentNonces: %v", err)
	}
	if removed != 0 || spentNonceCount(t, j) != 1 {
		t.Fatalf("removed = %d, records = %d, want the record kept", removed, spentNonceCount(t, j))
	}

	// Past it: the record goes, and the decision it belonged to is long expired.
	removed, err = j.PruneSpentNonces(ctx, issued.Add(3*time.Hour), time.Hour)
	if err != nil {
		t.Fatalf("PruneSpentNonces: %v", err)
	}
	if removed != 1 || spentNonceCount(t, j) != 0 {
		t.Fatalf("removed = %d, records = %d, want the old record removed",
			removed, spentNonceCount(t, j))
	}
}
