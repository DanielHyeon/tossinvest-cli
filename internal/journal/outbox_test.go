package journal

// outbox_test.go covers the alert outbox at the storage level (task 4.3). The
// delivery policy that sits on top of it is tested in internal/obs.

import (
	"context"
	"errors"
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

	id, err := j.EnqueueAlert(ctx, Alert{EventKey: "k", Type: "order.in_doubt", Severity: "critical"})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	for i := 0; i < 3; i++ {
		clk.Advance(time.Second)
		if err := j.MarkAlertAttemptFailed(ctx, id, "transport is down"); err != nil {
			t.Fatalf("MarkAlertAttemptFailed: %v", err)
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

	delivered, _ := j.EnqueueAlert(ctx, Alert{EventKey: "a", Type: "order.in_doubt"})
	acked, _ := j.EnqueueAlert(ctx, Alert{EventKey: "b", Type: "order.in_doubt"})

	if err := j.MarkAlertDelivered(ctx, delivered); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
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

func TestMarkingANonPendingAlertIsRefused(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	id, _ := j.EnqueueAlert(ctx, Alert{EventKey: "a", Type: "order.in_doubt"})
	if err := j.MarkAlertDelivered(ctx, id); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}

	if err := j.MarkAlertDelivered(ctx, id); !errors.Is(err, ErrAlertNotFound) {
		t.Errorf("double delivery = %v, want ErrAlertNotFound", err)
	}
	if err := j.MarkAlertAttemptFailed(ctx, 9999, "x"); !errors.Is(err, ErrAlertNotFound) {
		t.Errorf("unknown id = %v, want ErrAlertNotFound", err)
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

	// Simulate a v2 database by dropping everything later versions added and
	// rolling the recorded version back.
	first := open()
	for _, table := range []string{"flatten_steps", "flatten_sagas", "alert_outbox"} {
		if _, err := first.db.ExecContext(ctx, "DROP TABLE "+table); err != nil {
			t.Fatalf("dropping %s: %v", table, err)
		}
	}
	if _, err := first.db.ExecContext(ctx, "PRAGMA user_version = 2"); err != nil {
		t.Fatalf("rolling back the schema version: %v", err)
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
