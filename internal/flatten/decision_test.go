package flatten_test

// decision_test.go covers the flatten saga's decisions
// (extend-execution-contract task 1.6).
//
// The requirement is two-sided and both sides matter: the emergency path must
// not be refused by the new verification (§0.3 — nothing may weaken or delay the
// exit), and it must not be exempt from it either (engine-safety: "수동 flatten은
// 청산·취소 결정을 ReductionIntent preimage와 함께 journal에 기록한 뒤 제출한다").
//
// What that means concretely is checked here: a decision row exists before the
// mutation, it is RISK_REDUCING with a ReductionIntent preimage, the attempt
// points at it, and no limit snapshot travels with it.

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// decisionRows reads every decision the saga wrote, oldest first.
func decisionRows(t *testing.T, path string) []journal.Decision {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatalf("opening the journal file: %v", err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(),
		`SELECT id, safety_class, preimage_kind, risk_preimage, risk_hash,
		        coalesce(client_order_id, ''), coalesce(limits_json, '')
		 FROM decisions ORDER BY rowid`)
	if err != nil {
		t.Fatalf("reading decisions: %v", err)
	}
	defer rows.Close()

	var out []journal.Decision
	for rows.Next() {
		var d journal.Decision
		if err := rows.Scan(&d.ID, &d.SafetyClass, &d.PreimageKind, &d.RiskPreimage,
			&d.RiskHash, &d.ClientOrderID, &d.LimitsJSON); err != nil {
			t.Fatal(err)
		}
		out = append(out, d)
	}
	return out
}

// TestFlattenCancelRecordsItsDecisionFirst: the cancel goes through, and the
// record of what authorised it is on disk with the attempt pointing at it.
func TestFlattenCancelRecordsItsDecisionFirst(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{
		{order("O-1", "005930", "BUY", "10", "70000", "KRW")},
	})

	report, err := h.saga(false).CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Cancelled != 1 {
		t.Fatalf("cancelled = %d, want 1 (%+v)", report.Cancelled, report.Outcomes)
	}

	decisions := decisionRows(t, h.path)
	if len(decisions) != 1 {
		t.Fatalf("decisions written = %d, want exactly one per cancel", len(decisions))
	}
	d := decisions[0]
	if d.SafetyClass != journal.SafetyClassRiskReducing {
		t.Errorf("safety class = %q, want RISK_REDUCING", d.SafetyClass)
	}
	if d.PreimageKind != journal.PreimageKindReductionIntent {
		t.Errorf("preimage kind = %q, want REDUCTION_INTENT", d.PreimageKind)
	}
	if d.LimitsJSON != "" {
		t.Errorf("a flatten decision carries no limit snapshot, got %q", d.LimitsJSON)
	}
	if d.ClientOrderID != "" {
		t.Errorf("a cancel decision carries no idempotency key, got %q", d.ClientOrderID)
	}
	if d.RiskHash != journal.HashPreimage(d.RiskPreimage) {
		t.Error("the stored hash does not cover the stored preimage")
	}
	preimage, err := journal.ParsePreimage(d.PreimageKind, d.RiskPreimage)
	if err != nil {
		t.Fatalf("the stored preimage is not canonical: %v", err)
	}
	reduction, ok := preimage.(journal.ReductionIntent)
	if !ok {
		t.Fatalf("preimage type = %T, want ReductionIntent", preimage)
	}
	if reduction.Symbol != "005930" || reduction.Side != "BUY" || reduction.Market != "kr" {
		t.Errorf("the preimage does not describe the cancelled order: %+v", reduction)
	}
	if reduction.MaxQuantity != "10" {
		t.Errorf("max quantity = %q, want the resting order's 10", reduction.MaxQuantity)
	}
	if !strings.Contains(reduction.Reason, "flatten") {
		t.Errorf("reason = %q, want it to say a flatten did this", reduction.Reason)
	}

	// The attempt records which decision entitled it.
	steps, err := h.journal.FlattenSteps(context.Background(), report.SagaID)
	if err != nil {
		t.Fatal(err)
	}
	if len(steps) != 1 || steps[0].AttemptID == "" {
		t.Fatalf("expected one step with an attempt: %+v", steps)
	}
	rec, err := h.journal.LookupAttempt(context.Background(), steps[0].AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.DecisionID != d.ID {
		t.Errorf("attempt decision_id = %q, want %q", rec.DecisionID, d.ID)
	}
	if rec.SafetyClass != journal.SafetyClassRiskReducing {
		t.Errorf("attempt safety class = %q", rec.SafetyClass)
	}
}

// TestFlattenIssuesOneDecisionPerCancel: decisions are per mutation, not per
// saga. A shared decision would mean a single nonce for several orders, and the
// second cancel would be refused as a reuse.
func TestFlattenIssuesOneDecisionPerCancel(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{
		{
			order("O-1", "005930", "BUY", "10", "70000", "KRW"),
			order("O-2", "000660", "SELL", "5", "120000", "KRW"),
		},
	})

	report, err := h.saga(false).CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Cancelled != 2 {
		t.Fatalf("cancelled = %d, want 2 (%+v)", report.Cancelled, report.Outcomes)
	}
	if got := len(decisionRows(t, h.path)); got != 2 {
		t.Fatalf("decisions written = %d, want one per cancel", got)
	}
}

// TestLiquidationDecisionCarriesAKeyAndNoLimits: the sell is an order creation,
// so its decision derives an idempotency key and the key reaches the intent that
// is submitted. It is still an exit, so it carries no limit snapshot and is not
// measured against one — under the old rule the exemption was keyed on the
// mutation verb and a reduce-only sell got none.
func TestLiquidationDecisionCarriesAKeyAndNoLimits(t *testing.T) {
	account := &accountFake{
		positions: []domain.Position{position("005930", "kr", 10)},
		sellable:  map[string]float64{"005930": 6},
		lower:     map[string]float64{"005930": 63000},
	}
	h := newLiqHarness(t, [][]json.RawMessage{{}}, account)
	ctx := context.Background()
	saga := h.saga(false)
	h.sell.after = account.setFlat

	if _, err := saga.CancelAll(ctx); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if _, err := saga.Liquidate(ctx); err != nil {
		t.Fatalf("Liquidate: %v", err)
	}

	decisions := decisionRows(t, h.path)
	if len(decisions) != 1 {
		t.Fatalf("decisions = %d, want one for the sell", len(decisions))
	}
	d := decisions[0]
	if d.SafetyClass != journal.SafetyClassRiskReducing {
		t.Errorf("safety class = %q, want RISK_REDUCING for a reduce-only sell", d.SafetyClass)
	}
	if d.LimitsJSON != "" {
		t.Errorf("an exit carries no limit snapshot, got %q", d.LimitsJSON)
	}
	want := journal.DeriveClientOrderID(d.ID, 0)
	if d.ClientOrderID != want {
		t.Errorf("client_order_id = %q, want the derived key %q", d.ClientOrderID, want)
	}

	orders := h.sell.orders()
	if len(orders) != 1 {
		t.Fatalf("orders = %+v, want one", orders)
	}
	if orders[0].ClientOrderID != want {
		t.Errorf("the submitted intent carried key %q, want %q — the replay contract "+
			"assumes the key was sent", orders[0].ClientOrderID, want)
	}
}

// TestFlattenDryRunWritesNoDecision: a dry run submits nothing, so it decides
// nothing. A decision row with no mutation behind it would be a spent nonce and
// a claimed key for an order that was never placed.
func TestFlattenDryRunWritesNoDecision(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{
		{order("O-1", "005930", "BUY", "10", "70000", "KRW")},
	})

	if _, err := h.saga(true).CancelAll(context.Background()); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if got := len(decisionRows(t, h.path)); got != 0 {
		t.Fatalf("a dry run wrote %d decisions, want 0", got)
	}
	if got := h.broker.cancelled(); len(got) != 0 {
		t.Fatalf("a dry run cancelled %v", got)
	}
}
