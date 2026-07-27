package execgw_test

// reservation_gate_test.go is task 5.1: the gateway's HELD-reservation check on
// EXPOSURE_RAISING submissions (engine-safety "결정 영속과 신뢰 경계" MODIFIED —
// "예약 없는 진입 결정 제출 → Gateway가 거부한다").
//
// The check and the atomic issuance are two halves of one statement. The
// issuance makes "a submittable entry decision with no hold" unreachable from
// the issuer; this makes it refused wherever else it might come from. The tests
// below drive both halves: a decision issued without a hold is refused, and a
// decision the real Guardian issued goes through.
//
// The §0.3 direction is asserted too, because a check on the entry path that
// leaked into the exit path would be the worst kind of regression: an exit that
// cannot be submitted. RISK_REDUCING decisions take no reservation by contract
// (an exit lowers the aggregate it would reserve against), and the flatten saga
// — whose liquidation is exactly a ReductionIntent place — is the production
// caller that shape belongs to.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// TestAnEntryDecisionWithNoHeldReservationIsRefused.
//
// Everything else about the decision is right: it is persisted, unexpired,
// unspent, carries a complete limit snapshot and matches the order exactly. The
// one thing missing is the headroom, and the headroom is what the aggregate
// limits are made of.
func TestAnEntryDecisionWithNoHeldReservationIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	intent := placeIntent()
	// The plain issuer, deliberately: it records the decision and nothing else.
	decision, err := issuerFor(j, clk).IssueEntry(ctx, execgw.EntryRequest{
		Kind: journal.KindPlace, Market: intent.Market, Symbol: intent.Symbol, Side: intent.Side,
		Quantity: intent.Quantity, EntryPrice: intent.Price, StopPrice: intent.Price * 0.9,
		PolicyVersion: "test/v1", Limits: testLimits(),
	})
	if err != nil {
		t.Fatalf("issuing the entry decision: %v", err)
	}

	out, err := gw.Place(ctx, execgw.PlaceRequest{Intent: intent, Decision: decision})

	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("err = %v, want a gateway refusal", err)
	}
	if rejected.Reason != execgw.ReasonGuardianReservationMissing {
		t.Errorf("reason = %q, want %q", rejected.Reason, execgw.ReasonGuardianReservationMissing)
	}
	if !strings.Contains(rejected.Detail, "reservation") {
		t.Errorf("detail %q does not say what is missing", rejected.Detail)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls = %d; nothing may be sent for an unpaid authorisation", places)
	}
	// The refusal is journalled against the attempt, so "why did the engine not
	// trade" is answerable from the journal alone.
	if out.State != journal.StateNotDispatched {
		t.Errorf("state = %s, want NOT_DISPATCHED", out.State)
	}
	rec, err := j.LookupAttempt(ctx, out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != journal.StateNotDispatched ||
		rec.ReasonCode != string(execgw.ReasonGuardianReservationMissing) {
		t.Errorf("journalled attempt = %s/%s, want NOT_DISPATCHED with the reservation reason",
			rec.State, rec.ReasonCode)
	}
}

// TestAReleasedReservationNoLongerAuthorisesTheEntry: the check is about a HELD
// hold and not about a row having existed once. A released reservation has
// returned its headroom to the account, so the entry it used to pay for is no
// longer paid for.
func TestAReleasedReservationNoLongerAuthorisesTheEntry(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	intent := placeIntent()
	decision := entryDecision(t, j, clk, intent, testLimits())

	auditor := &modeAuditor{}
	if _, err := j.OperatorReleaseReservation(ctx, journal.OperatorReleaseRequest{
		ReservationID: "hold-" + decision.ID,
		Operator:      "operator",
		Reason:        "the order was cancelled at the broker",
		Evidence:      "broker order list is empty",
		Auditor:       auditor,
	}); err != nil {
		t.Fatalf("releasing the hold: %v", err)
	}

	_, err := gw.Place(ctx, execgw.PlaceRequest{Intent: intent, Decision: decision})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianReservationMissing {
		t.Fatalf("err = %v, want %q", err, execgw.ReasonGuardianReservationMissing)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls = %d", places)
	}
}

// TestAnExitNeedsNoReservation is the §0.3 direction and the flatten regression
// in one: a reduce-only sell carries a ReductionIntent decision with no limit
// snapshot and no hold, and it goes through. That is the exact shape
// internal/flatten's liquidation submits.
func TestAnExitNeedsNoReservation(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-2"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	sell := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 2, Price: 71000, CurrencyMode: "KRW",
	}
	decision := exitDecision(t, j, clk, journal.KindPlace, sell.Market, sell.Symbol, sell.Side, sell.Quantity)

	if held, err := j.HeldReservations(ctx, "acct-7"); err != nil || len(held) != 0 {
		t.Fatalf("precondition: held = %+v err = %v, want no holds at all", held, err)
	}
	out, err := gw.Place(ctx, execgw.PlaceRequest{Intent: sell, Decision: decision})
	if err != nil {
		t.Fatalf("an exit must not be refused for holding no reservation: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("state = %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls = %d, want 1", places)
	}
}

// TestACancelNeedsNoReservation: the same rule reached by the other verb, which
// is what the flatten saga's cancel phase does.
func TestACancelNeedsNoReservation(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "O-3"}}
	gw, j, clk := newGateway(t, broker)
	ctx := context.Background()

	decision := exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2)
	if _, err := gw.Cancel(ctx, execgw.CancelRequest{
		Intent:   orderintent.CancelIntent{OrderID: "O-3", Symbol: "005930"},
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: decision,
	}); err != nil {
		t.Fatalf("a cancel must not be refused for holding no reservation: %v", err)
	}
	if _, cancels, _ := broker.totals(); cancels != 1 {
		t.Errorf("broker cancel calls = %d, want 1", cancels)
	}
}

// TestTheGuardiansOwnIssuanceSubmits closes the loop the two halves make: the
// issuer that takes the hold atomically produces a decision the gateway's check
// accepts, and the order reaches the broker.
//
// Without this the pair could be consistent and useless — a check nothing can
// satisfy is a gate that is off, spelled differently.
func TestTheGuardiansOwnIssuanceSubmits(t *testing.T) {
	rig := newModeRig(t, filepath.Join(t.TempDir(), "journal.db"),
		domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-9"})
	ctx := context.Background()

	guardian := mustGuardian(t, execgw.RiskGuardianOptions{
		Journal: rig.j, Clock: rig.clk, AccountRef: "acct-7", Policy: guardianPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "add-core-domain/5.1",
	})
	issued, err := guardian.IssueEntry(ctx, execgw.EntryIssuance{
		Intent:  guardianIntent(),
		Account: guardianAccount(),
		Collect: func(ctx context.Context, _ int) (execgw.ExposureSnapshot, error) {
			v, err := rig.j.ReservationVersion(ctx, "acct-7")
			if err != nil {
				return execgw.ExposureSnapshot{}, err
			}
			return execgw.ExposureSnapshot{
				AsOf: rig.clk.Now(), Version: v,
				OpenExposure: riskcalc.Money{Amount: "0", Currency: "KRW"},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("IssueEntry: %v", err)
	}

	out, err := rig.gw.Place(ctx, execgw.PlaceRequest{
		Intent: guardianPlaceIntent(), Decision: issued.Decision,
	})
	if err != nil {
		t.Fatalf("the Guardian's own decision must submit: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Errorf("state = %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
	if places, _, _ := rig.broker.totals(); places != 1 {
		t.Errorf("broker place calls = %d, want 1", places)
	}
}
