package journal

import (
	"context"
	"testing"
)

func TestA165ConfirmedCancelAfterLatchRequiresRecordedAfterTheLatch(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	owner := a165Prepare(t, j, "owner-intent", "owner-attempt", KindPlace, "")
	a165Confirm(t, owner, "O-a165")

	before := a165Prepare(t, j, "before-intent", "before-attempt", KindCancel, "O-a165")
	a165Confirm(t, before, "")
	latch := a165Prepare(t, j, "latch-intent", "latch-attempt", KindCancel, "O-a165")
	if err := latch.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("latch dispatch started: %v", err)
	}
	if err := latch.Settle(ctx, StateFailedConfirmed, "broker_rejected",
		"broker refused [a165 sends=1 branch=cancel_already_processing stop=definitive outcome=definitive]"); err != nil {
		t.Fatalf("latch settle: %v", err)
	}
	if _, found, err := j.ConfirmedCancelAfterLatch(ctx, latch.ID()); err != nil || found {
		t.Fatalf("pre-latch cancel: found=%v err=%v", found, err)
	}

	after := a165Prepare(t, j, "after-intent", "after-attempt", KindCancel, "O-a165")
	a165Confirm(t, after, "")
	got, found, err := j.ConfirmedCancelAfterLatch(ctx, latch.ID())
	if err != nil || !found || got.ID != after.ID() {
		t.Fatalf("post-latch cancel: got=%+v found=%v err=%v", got, found, err)
	}
}

func a165Prepare(t *testing.T, j *Journal, intentID, attemptID string, kind MutationKind, target string) *Attempt {
	t.Helper()
	intent := testIntent()
	intent.ID = intentID
	intent.Fingerprint = "fp-" + intentID
	a, err := j.Prepare(context.Background(), PrepareRequest{
		Intent: intent, Kind: kind, AttemptID: attemptID, TargetOrderID: target,
	})
	if err != nil {
		t.Fatalf("Prepare(%s): %v", attemptID, err)
	}
	return a
}

func a165Confirm(t *testing.T, a *Attempt, orderID string) {
	t.Helper()
	ctx := context.Background()
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted(%s): %v", a.ID(), err)
	}
	if a.Kind() == KindPlace {
		if err := a.MarkAcked(ctx, orderID); err != nil {
			t.Fatalf("MarkAcked(%s): %v", a.ID(), err)
		}
	}
	if err := a.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "confirmed"); err != nil {
		t.Fatalf("Settle(%s): %v", a.ID(), err)
	}
}
