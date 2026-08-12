package journal

// outbox_test.go covers the alert outbox at the storage level (task 4.3). The
// delivery policy that sits on top of it is tested in internal/obs.

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func outboxJournal(t *testing.T) (*Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC))
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

func TestEnqueueAlertIsIdempotentOnTheEventKey(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	first, err := j.EnqueueAlert(ctx, Alert{EventKey: "unknown|AAPL", Type: "broker.state_unknown",
		Severity: "critical", Body: "first"})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	second, err := j.EnqueueAlert(ctx, Alert{EventKey: "unknown|AAPL", Type: "broker.state_unknown",
		Severity: "critical", Body: "second"})
	if err != nil {
		t.Fatalf("EnqueueAlert (repeat): %v", err)
	}
	if first != second {
		t.Errorf("ids = %d and %d; the same condition must be one row", first, second)
	}

	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("rows = %d, want 1", len(pending))
	}
	// The first observation's text is kept: it is the one that describes what was
	// seen when the condition arose.
	if pending[0].Body != "first" {
		t.Errorf("body = %q, want the first observation", pending[0].Body)
	}
}

func TestEnqueueAlertRequiresAKeyAndAType(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	if _, err := j.EnqueueAlert(ctx, Alert{Type: "x"}); err == nil {
		t.Error("an alert with no event key cannot be deduplicated and must be refused")
	}
	if _, err := j.EnqueueAlert(ctx, Alert{EventKey: "k"}); err == nil {
		t.Error("an alert with no type must be refused")
	}
}

func TestFailedAttemptsAccumulateAndTheRowStaysPending(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx,
		Alert{EventKey: "k", Type: "order.in_doubt", Severity: "critical"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	id := claim.ID
	// One claim, three attempts under it. A sender that recorded a failure and
	// lost the row between attempts would be handing its remaining retries to
	// somebody else's send, so the lease is deliberately kept across all three.
	for i := 0; i < 3; i++ {
		clk.Advance(time.Second)
		res, ferr := j.MarkAlertAttemptFailed(ctx, id, claim.Token, "transport is down")
		if ferr != nil {
			t.Fatalf("MarkAlertAttemptFailed: %v", ferr)
		}
		if res.Outcome != SettleApplied {
			t.Fatalf("outcome = %v on attempt %d, want applied — the lease is held across the budget",
				res.Outcome, i+1)
		}
	}

	alert, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if alert.State != AlertPending {
		t.Errorf("state = %q, want PENDING — a critical alert is preserved, not discarded", alert.State)
	}
	if alert.Attempts != 3 {
		t.Errorf("attempts = %d, want 3", alert.Attempts)
	}
	if alert.LastAttemptAt == nil {
		t.Error("the last attempt time must be recorded")
	}
	if alert.LastError != "transport is down" {
		t.Errorf("last error = %q", alert.LastError)
	}
}

func TestDeliveryAndAcknowledgementAreDistinctStates(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx,
		Alert{EventKey: "a", Type: "order.in_doubt"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	delivered := claim.ID
	acked, _ := j.EnqueueAlert(ctx, Alert{EventKey: "b", Type: "order.in_doubt"})

	if _, err := j.MarkAlertDelivered(ctx, delivered, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
	// Acknowledgement needs no lease and never did: the operator outranks the
	// machine that is midway through telling them.
	if err := j.AcknowledgeAlert(ctx, acked, "operator"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}

	count, err := j.UndeliveredCount(ctx)
	if err != nil {
		t.Fatalf("UndeliveredCount: %v", err)
	}
	if count != 0 {
		t.Errorf("undelivered = %d, want 0", count)
	}

	a, _ := j.LookupAlert(ctx, delivered)
	if a.State != AlertDelivered || a.DeliveredAt == nil {
		t.Errorf("delivered row = %+v", a)
	}
	b, _ := j.LookupAlert(ctx, acked)
	if b.State != AlertAcknowledged || b.AcknowledgedBy != "operator" {
		t.Errorf("acknowledged row = %+v", b)
	}
}

func TestAcknowledgeRequiresAnOperator(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	id, _ := j.EnqueueAlert(ctx, Alert{EventKey: "a", Type: "order.in_doubt"})

	if err := j.AcknowledgeAlert(ctx, id, "   "); err == nil {
		t.Fatal("acknowledging without an identity must fail — this is a human overriding the machine")
	}
}

// TestMarkingANonPendingAlertIsRefused: a settlement lands once, and what
// happens on the second one is now said rather than raised.
//
// It used to be one error for both cases. a099 splits them because the caller's
// behaviour splits: a row that is no longer PENDING was settled by somebody —
// routinely by an operator acknowledging mid-flight — and a sender that hears it
// simply stops. A row that is not there at all is the ledger having lost
// something a sender was holding, and that is a fault.
func TestMarkingANonPendingAlertIsRefused(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	claim, err := j.ClaimAlertForDelivery(ctx,
		Alert{EventKey: "a", Type: "order.in_doubt"}, claimRemind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if res, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	} else if res.Outcome != SettleApplied {
		t.Fatalf("outcome = %v on the first delivery, want applied", res.Outcome)
	}

	res, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token)
	if err != nil {
		t.Fatalf("double delivery returned an error: %v", err)
	}
	if res.Outcome != SettleAlreadySettled {
		t.Errorf("double delivery = %v, want already-settled", res.Outcome)
	}

	missing, err := j.MarkAlertAttemptFailed(ctx, 9999, claim.Token, "x")
	if err != nil {
		t.Fatalf("settling an unknown id returned an error: %v", err)
	}
	if missing.Outcome != SettleNotFound {
		t.Errorf("unknown id = %v, want not-found", missing.Outcome)
	}
}

// TestSettlingWithoutAClaimIsRefused: the token is the whole of the exclusion,
// so a settlement that presents none is refused before it reaches the row.
// Without this a caller could settle any row by id, which is what the ledger did
// before a099 and what let a stalled sender overwrite the send another one was
// making.
func TestSettlingWithoutAClaimIsRefused(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	id, err := j.EnqueueAlert(ctx, Alert{EventKey: "a", Type: "order.in_doubt"})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}

	if _, err := j.MarkAlertDelivered(ctx, id, ""); err == nil {
		t.Error("marking delivered with no token succeeded — the token is the exclusion")
	}
	if _, err := j.MarkAlertAttemptFailed(ctx, id, "", "x"); err == nil {
		t.Error("recording a failure with no token succeeded")
	}
	if _, err := j.ReleaseAlertClaim(ctx, id, ""); err == nil {
		t.Error("releasing with no token succeeded")
	}
	if got := alertState(t, j, id); got != AlertPending {
		t.Errorf("state = %q, want %q — a refused settlement must not touch the row", got, AlertPending)
	}
}

// TestClaimingWithoutANameIsRefused guards the other end of the same exclusion.
//
// A lease has to say who holds it: the holder's name is what an operator reads
// in engine.alert_claim_held when a send is suppressed, and what a steal names
// as the sender it displaced. An anonymous claim would take the row and leave
// nothing to name, so the refusal happens before the transaction opens — this
// journal's transactions take the write lock up front, and waiting for it only
// to reject the input would be a real wait.
func TestClaimingWithoutANameIsRefused(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	alert := Alert{EventKey: "a", Type: "order.in_doubt", Severity: "critical"}

	if _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, ""); err == nil {
		t.Error("an unnamed claimant took the lease — nothing would name the holder")
	}
	if _, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, "   "); err == nil {
		t.Error("a blank claimant took the lease — the name is trimmed before it is stored")
	}
	// And the refusal wrote nothing: no row was even recorded, because the guard
	// sits ahead of the recording half.
	pending, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("rows = %d, want 0 — a refused claim must not record the alert either", len(pending))
	}
}

// TestOutboxSurvivesTheV2Migration is the §0.6 test: a journal written by an
// older build migrates forward and gains every table the current one adds.
func TestOutboxSurvivesTheV2Migration(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC))
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	open := func() *Journal {
		j, err := Open(ctx, Options{Path: path, Clock: clk,
			FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		return j
	}

	// A real v2 database: the migration list stops at version 2, so the file on
	// disk is what the build of that era wrote. (It used to be simulated by
	// dropping the tables later versions added, which stopped working at v5:
	// that version also adds columns, and an ALTER cannot be undone by a DROP
	// TABLE. The assertions below are unchanged.)
	first, err := Open(ctx, Options{Path: path, Clock: clk,
		FSProber:          FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		migrationOverride: &migrationPlan{steps: migrationsThrough(2), target: 2}})
	if err != nil {
		t.Fatalf("Open at schema v2: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second := open()
	defer second.Close()

	version, err := second.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion: %v", err)
	}
	if version != SchemaVersion {
		t.Fatalf("version after migration = %d, want %d", version, SchemaVersion)
	}
	if _, err := second.EnqueueAlert(ctx, Alert{EventKey: "k", Type: "order.in_doubt"}); err != nil {
		t.Fatalf("the migrated database must accept alerts: %v", err)
	}
	if _, err := second.StartFlatten(ctx, FlattenSaga{ID: "s-1", AccountRef: "acct-7"}); err != nil {
		t.Fatalf("the migrated database must accept flatten sagas: %v", err)
	}
}
