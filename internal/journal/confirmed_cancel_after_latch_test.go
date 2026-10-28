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

func TestA165ConfirmedCancelAfterLatchUsesTheLatchScopeAndOneOwner(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name        string
		beforeLatch bool
		ownerCount  int
		afterDay    string
		afterScope  a165Scope
		afterTarget string
		wantFound   bool
	}{
		{
			name: "same target after latch releases across trading days", ownerCount: 1,
			afterDay: "2026-03-31", wantFound: true,
		},
		{
			name: "recorded before latch does not release even if confirmed after", beforeLatch: true,
			ownerCount: 1, wantFound: false,
		},
		{
			name: "different account does not release", ownerCount: 1,
			afterScope: a165Scope{account: "acct-other"}, wantFound: false,
		},
		{
			name: "different market does not release", ownerCount: 1,
			afterScope: a165Scope{market: "kr"}, wantFound: false,
		},
		{
			name: "different symbol does not release", ownerCount: 1,
			afterScope: a165Scope{symbol: "005930"}, wantFound: false,
		},
		{
			name: "different target does not release", ownerCount: 1,
			afterTarget: "O-other", wantFound: false,
		},
		{
			name: "reused order number has two owners", ownerCount: 2, wantFound: false,
		},
		{
			name: "no confirmed owner does not release", ownerCount: 0, wantFound: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			for n := 0; n < tc.ownerCount; n++ {
				owner := a165PrepareScoped(t, j, "owner-intent-"+string(rune('a'+n)), "owner-attempt-"+string(rune('a'+n)), KindPlace, "", a165Scope{})
				a165Confirm(t, owner, "O-a165")
			}

			var before *Attempt
			if tc.beforeLatch {
				before = a165PrepareScoped(t, j, "before-intent", "before-attempt", KindCancel, "O-a165", tc.afterScope)
			}
			latch := a165PrepareScoped(t, j, "latch-intent", "latch-attempt", KindCancel, "O-a165", a165Scope{})
			if err := latch.MarkDispatchStarted(ctx); err != nil {
				t.Fatalf("latch dispatch started: %v", err)
			}
			if err := latch.Settle(ctx, StateFailedConfirmed, "broker_rejected",
				"broker refused [a165 sends=1 branch=cancel_already_processing stop=definitive outcome=definitive]"); err != nil {
				t.Fatalf("latch settle: %v", err)
			}

			if before != nil {
				a165Confirm(t, before, "C-before")
			} else {
				afterScope := tc.afterScope
				afterScope.day = tc.afterDay
				afterTarget := tc.afterTarget
				if afterTarget == "" {
					afterTarget = "O-a165"
				}
				after := a165PrepareScoped(t, j, "after-intent", "after-attempt", KindCancel, afterTarget, afterScope)
				a165Confirm(t, after, "C-after")
			}

			got, found, err := j.ConfirmedCancelAfterLatch(ctx, latch.ID())
			if err != nil || found != tc.wantFound {
				t.Fatalf("ConfirmedCancelAfterLatch: got=%+v found=%v want=%v err=%v", got, found, tc.wantFound, err)
			}
		})
	}
}

type a165Scope struct {
	account string
	market  string
	symbol  string
	day     string
}

func a165Prepare(t *testing.T, j *Journal, intentID, attemptID string, kind MutationKind, target string) *Attempt {
	return a165PrepareScoped(t, j, intentID, attemptID, kind, target, a165Scope{})
}

func a165PrepareScoped(t *testing.T, j *Journal, intentID, attemptID string, kind MutationKind, target string, scope a165Scope) *Attempt {
	t.Helper()
	intent := testIntent()
	intent.ID = intentID
	intent.Fingerprint = "fp-" + intentID
	if scope.account != "" {
		intent.AccountRef = scope.account
	}
	if scope.market != "" {
		intent.Market = scope.market
	}
	if scope.symbol != "" {
		intent.Symbol = scope.symbol
	}
	if scope.day != "" {
		intent.TradingDay = scope.day
	}
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
	if orderID == "" {
		orderID = "ack-" + a.ID()
	}
	if err := a.MarkAcked(ctx, orderID); err != nil {
		t.Fatalf("MarkAcked(%s): %v", a.ID(), err)
	}
	if err := a.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "confirmed"); err != nil {
		t.Fatalf("Settle(%s): %v", a.ID(), err)
	}
}
