package execgw_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestA165EngineCancelLatchBlocksOnlyNewEntryUntilJournalledRelease(t *testing.T) {
	broker := &fakeBroker{err: &official.APIError{Code: 409, Body: `{"code":"already-processing"}`}}
	clk := newA165Clock()
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", CancelRefusal: execgw.EngineCancelRefusal(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a165ConfirmedOwner(t, j, "owner-intent", "owner-attempt", "O-a165")
	cancel := execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	}
	out, err := gw.Cancel(context.Background(), cancel)
	if err == nil || out.State != journal.StateFailedConfirmed || out.Reason != execgw.ReasonBrokerRejected {
		t.Fatalf("R3 outcome=%+v err=%v", out, err)
	}
	if _, cancels, _ := broker.totals(); cancels != 1 {
		t.Fatalf("cancel calls=%d, want 1", cancels)
	}

	_, err = gw.Place(context.Background(), placeRequest(t, j, clk))
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonSymbolInFlight {
		t.Fatalf("entry err=%v, want symbol latch", err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Fatalf("blocked entry reached broker: places=%d", places)
	}

	a165ConfirmedCancel(t, j, "release-intent", "release-attempt", "O-a165")
	_, err = gw.Place(context.Background(), placeRequest(t, j, clk))
	if places, _, _ := broker.totals(); places != 1 {
		t.Fatalf("released entry calls=%d, want 1", places)
	}
}

func TestA165LatchDoesNotBlockACancel(t *testing.T) {
	broker := &fakeBroker{err: &official.APIError{Code: 409, Body: `{"code":"already-processing"}`}}
	clk := newA165Clock()
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", CancelRefusal: execgw.EngineCancelRefusal(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a165ConfirmedOwner(t, j, "owner-intent", "owner-attempt", "O-a165")
	request := execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	}
	if out, err := gw.Cancel(context.Background(), request); err == nil || out.State != journal.StateFailedConfirmed {
		t.Fatalf("first cancel outcome=%+v err=%v", out, err)
	}
	request.Decision = exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2)
	if out, err := gw.Cancel(context.Background(), request); err == nil || out.State != journal.StateFailedConfirmed {
		t.Fatalf("latch must not block follow-up cancel: outcome=%+v err=%v", out, err)
	}
	if _, cancels, _ := broker.totals(); cancels != 2 {
		t.Fatalf("cancel calls=%d, want 2", cancels)
	}
}

func TestA165LatchDoesNotBlockAReduceOnlySell(t *testing.T) {
	broker := &fakeBroker{err: &official.APIError{Code: 409, Body: `{"code":"already-processing"}`}}
	clk := newA165Clock()
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", CancelRefusal: execgw.EngineCancelRefusal(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a165ConfirmedOwner(t, j, "owner-intent", "owner-attempt", "O-a165")
	if _, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	}); err == nil {
		t.Fatal("R3 cancel unexpectedly succeeded")
	}

	broker.err = nil
	broker.result = domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "S-a165"}
	intent := placeIntent()
	intent.Side = "sell"
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   intent,
		Decision: exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL", 2),
	})
	if err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("reduce-only sell outcome=%+v err=%v", out, err)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Fatalf("reduce-only sell must reach broker: places=%d", places)
	}
}

func TestA165TerminalEvidenceReleasesTheCancelLatch(t *testing.T) {
	broker := &fakeBroker{err: &official.APIError{Code: 409, Body: `{"code":"already-processing"}`}}
	clk := newA165Clock()
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", CancelRefusal: execgw.EngineCancelRefusal(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	a165ConfirmedOwner(t, j, "owner-intent", "owner-attempt", "O-a165")
	_, err = gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	})
	if err == nil {
		t.Fatal("R3 cancel unexpectedly succeeded")
	}
	if _, err := j.RecordFill(context.Background(), journal.FillObservation{
		OrderID: "O-a165", AccountRef: "acct-7", Symbol: "005930", Market: "kr", TradingDay: "2026-03-30", Side: "BUY",
		State: "CANCELLED", Terminal: true, Quantity: "2", FilledQuantity: "0", ObservedAt: fixedNow.Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	broker.err = nil
	broker.result = domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "P-a165"}
	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("terminal release outcome=%+v err=%v", out, err)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Fatalf("released entry calls=%d, want 1", places)
	}
}

func TestA165PolicyDoesNotChangeNilOrNonCancelAmbiguousOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy *execgw.CancelRefusalPolicy
		invoke func(*execgw.Gateway, *journal.Journal, *clock.Fake) (execgw.Outcome, error)
	}{
		{
			name: "nil policy cancel", policy: nil,
			invoke: func(gw *execgw.Gateway, j *journal.Journal, clk *clock.Fake) (execgw.Outcome, error) {
				return gw.Cancel(context.Background(), execgw.CancelRequest{
					Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
					Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
					Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
				})
			},
		},
		{
			name: "engine policy place", policy: execgw.EngineCancelRefusal(),
			invoke: func(gw *execgw.Gateway, j *journal.Journal, clk *clock.Fake) (execgw.Outcome, error) {
				return gw.Place(context.Background(), placeRequest(t, j, clk))
			},
		},
		{
			name: "engine policy amend", policy: execgw.EngineCancelRefusal(),
			invoke: func(gw *execgw.Gateway, j *journal.Journal, clk *clock.Fake) (execgw.Outcome, error) {
				price := 70500.0
				return gw.Amend(context.Background(), execgw.AmendRequest{
					Intent: orderintent.AmendIntent{OrderID: "O-a165", Price: &price}, Symbol: "005930",
					Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
					Decision: raisingDecision(t, j, clk, journal.KindAmend, "kr", "005930", "BUY", 2, price, testLimits()),
				})
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broker := &fakeBroker{err: &official.APIError{Code: 409, Body: `{"code":"already-processing"}`}}
			clk := newA165Clock()
			j := openJournal(t, clk)
			gw, err := execgw.New(execgw.Options{
				Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
				AccountRef: "acct-7", Source: "test", CancelRefusal: tc.policy,
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			out, err := tc.invoke(gw, j, clk)
			if err == nil || out.State != journal.StateInDoubt || strings.Contains(out.Detail, "[a165 ") {
				t.Fatalf("outcome=%+v err=%v, want unchanged IN_DOUBT", out, err)
			}
		})
	}
}

func TestA165EnginePolicyLeavesOther409Ambiguous(t *testing.T) {
	broker := &fakeBroker{err: &official.APIError{Code: 409, Body: `{"code":"other"}`}}
	clk := newA165Clock()
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", CancelRefusal: execgw.EngineCancelRefusal(),
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	out, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-a165", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	})
	if err == nil || out.State != journal.StateInDoubt || strings.Contains(out.Detail, "[a165 ") {
		t.Fatalf("outcome=%+v err=%v, want unchanged IN_DOUBT", out, err)
	}
}

func newA165Clock() *clock.Fake {
	return clock.NewFake(fixedNow)
}

func a165ConfirmedOwner(t *testing.T, j *journal.Journal, intentID, attemptID, orderID string) {
	t.Helper()
	a165DriveConfirmed(t, j, intentID, attemptID, journal.KindPlace, "", orderID)
}

func a165ConfirmedCancel(t *testing.T, j *journal.Journal, intentID, attemptID, orderID string) {
	t.Helper()
	a165DriveConfirmed(t, j, intentID, attemptID, journal.KindCancel, orderID, "C-a165")
}

func a165DriveConfirmed(t *testing.T, j *journal.Journal, intentID, attemptID string, kind journal.MutationKind, target, brokerOrderID string) {
	t.Helper()
	intent := journal.Intent{ID: intentID, Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-7", Symbol: "005930", Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "2", Price: "70000", Currency: "KRW", Source: "test", Fingerprint: "fp-" + intentID}
	a, err := j.Prepare(context.Background(), journal.PrepareRequest{Intent: intent, Kind: kind, AttemptID: attemptID, TargetOrderID: target})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := a.MarkDispatchStarted(context.Background()); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := a.MarkAcked(context.Background(), brokerOrderID); err != nil {
		t.Fatalf("MarkAcked: %v", err)
	}
	if err := a.Settle(context.Background(), journal.StateConfirmed, journal.ReasonBrokerAcknowledged, "confirmed"); err != nil {
		t.Fatalf("Settle: %v", err)
	}
}
