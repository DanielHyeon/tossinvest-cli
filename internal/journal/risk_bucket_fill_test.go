package journal

import (
	"context"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestRiskBucketAuthoritativePartialFillCommitsWithPositionAndCompletesActualEvidence(t *testing.T) {
	j, key, decisionID, reserved := riskBucketFillFixture(t, "partial", "risk-order-1")
	insertPosition(t, j, "risk-position", nil)
	if err := j.SetApplyHooks(ApplyHooks{Project: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
		_, err := tx.Exec(ctx, `UPDATE positions SET quantity=?,state='OPEN' WHERE id='risk-position'`, fill.CumulativeQuantity)
		return err
	}}); err != nil {
		t.Fatal(err)
	}
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-order-1", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	res, err := j.RecordFill(context.Background(), observation("risk-order-1", "4"))
	if err != nil || !res.Changed {
		t.Fatalf("fill=%+v err=%v", res, err)
	}
	state, err := j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "30", "20", false, true)
	var quantity string
	if err := j.db.QueryRow(`SELECT quantity FROM positions WHERE id='risk-position'`).Scan(&quantity); err != nil || quantity != "4" {
		t.Fatalf("position quantity=%q err=%v", quantity, err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_fills"); got != 1 {
		t.Fatalf("fills=%d", got)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_fill_allocations"); got != 5 {
		t.Fatalf("allocations=%d", got)
	}
	actual := riskBucketActual("12", "1", "0")
	result, err := j.completeRiskBucketFillActual(context.Background(), RiskBucketActualFillPlan{Owner: key, DecisionID: decisionID, OrderID: "risk-order-1", CumulativeFill: 4, Actual: actual, ObservedAt: riskFillNow})
	if err != nil || !result.ActualEvidenceCompleted {
		t.Fatalf("actual result=%+v err=%v", result, err)
	}
	state, err = j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "30", "48", false, false)
	duplicate, err := j.completeRiskBucketFillActual(context.Background(), RiskBucketActualFillPlan{Owner: key, DecisionID: decisionID, OrderID: "risk-order-1", CumulativeFill: 4, Actual: actual, ObservedAt: riskFillNow})
	if err != nil || !duplicate.Duplicate {
		t.Fatalf("duplicate=%+v err=%v", duplicate, err)
	}
}

func TestRiskBucketFillReplacementPredecessorLateFillCancelAndLateSuccessorAreConservative(t *testing.T) {
	j, key, decisionID, reserved := riskBucketFillFixture(t, "replacement", "risk-parent")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-parent", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), observation("risk-parent", "4")); err != nil {
		t.Fatal(err)
	}
	childReserved := riskReservedMap("30")
	recordConfirmedFillOrderScopeQuantity(t, j, "risk-child-intent", "risk-child-attempt", "risk-child", FillSnapshotScope{
		AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "BUY",
	}, "6")
	bindRiskOrderAttemptDecision(t, j, "risk-child", decisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-child", DecisionID: decisionID, PredecessorOrderID: "risk-parent", OrderQuantity: 6, ReservedMinor: childReserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: key, DecisionID: decisionID, OrderID: "risk-parent", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(2 * time.Second)}); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("replaced parent release error=%v", err)
	}
	stateBeforeChild, err := j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, stateBeforeChild, "30", "20", false, true)
	child := observation("risk-child", "3")
	child.Quantity = "6"
	if _, err := j.RecordFill(context.Background(), child); err != nil {
		t.Fatal(err)
	}
	lateParent := observation("risk-parent", "5")
	if _, err := j.RecordFill(context.Background(), lateParent); err != nil {
		t.Fatal(err)
	}
	state, err := j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "10", "40", false, true)
	firstRelease, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: key, DecisionID: decisionID, OrderID: "risk-child", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(2 * time.Second)})
	if err != nil || !firstRelease.Released {
		t.Fatalf("release=%+v err=%v", firstRelease, err)
	}
	retryRelease, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: key, DecisionID: decisionID, OrderID: "risk-child", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(2 * time.Second)})
	if err != nil || !retryRelease.AlreadyReleased {
		t.Fatalf("retry release=%+v err=%v", retryRelease, err)
	}
	eventsBeforeMismatch := countRiskBucketRows(t, j, "risk_bucket_events")
	if _, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: key, DecisionID: decisionID, OrderID: "risk-child", Reason: RiskBucketReleaseExpiry, ReleasedAt: riskFillNow.Add(3 * time.Second)}); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("mismatched release reason error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_events"); got != eventsBeforeMismatch {
		t.Fatalf("mismatched release wrote event: before=%d after=%d", eventsBeforeMismatch, got)
	}
	child.FilledQuantity = "4"
	if _, err := j.RecordFill(context.Background(), child); err != nil {
		t.Fatalf("late successor fill was dropped: %v", err)
	}
	state, err = j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "0", "45", true, true)
}

func TestRiskBucketFillSidecarRollsBackWithAuthoritativeFillTransaction(t *testing.T) {
	j, key, decisionID, reserved := riskBucketFillFixture(t, "crash", "risk-crash")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-crash", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("synthetic projection crash")
	if err := j.SetApplyHooks(ApplyHooks{Project: func(context.Context, *ApplyTx, AppliedFill) error { return boom }}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), observation("risk-crash", "4")); !errors.Is(err, boom) {
		t.Fatalf("fill error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_fills"); got != 0 {
		t.Fatalf("risk fills=%d", got)
	}
	state, err := j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "50", "0", false, false)
	if _, err := j.LookupFill(context.Background(), "risk-crash"); !errors.Is(err, ErrFillNotFound) {
		t.Fatalf("authoritative snapshot survived crash: %v", err)
	}
}

func TestRiskBucketFillRestartReplayAndDriftLatchNeverDropAuthoritativeFill(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	key, decisionID, reserved := seedRiskBucketFillFixture(t, j, "restart", "risk-restart")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-restart", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), observation("risk-restart", "2")); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(riskFillNow), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if state, err := reopened.ReadRiskBucketState(context.Background(), key); err != nil || state.Digest == "" {
		t.Fatalf("restart state=%+v err=%v", state, err)
	}
	if _, err := reopened.db.Exec(`DROP TRIGGER risk_bucket_order_reservations_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := reopened.db.Exec(`UPDATE risk_bucket_order_reservations SET reserved_minor='49' WHERE order_key=(SELECT order_key FROM risk_bucket_orders WHERE order_id='risk-restart') AND reservation_id LIKE '%:sector'`); err != nil {
		t.Fatal(err)
	}
	obs := observation("risk-restart", "3")
	if result, err := reopened.RecordFill(context.Background(), obs); err != nil || !result.Changed {
		t.Fatalf("drift dropped fill result=%+v err=%v", result, err)
	}
	var latch int
	if err := reopened.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE latch='REPLAY_MISMATCH'`).Scan(&latch); err != nil || latch != 1 {
		t.Fatalf("latch=%d err=%v", latch, err)
	}
	var unaccounted int
	if err := reopened.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE event_type='FILL_UNACCOUNTED'`).Scan(&unaccounted); err != nil || unaccounted != 1 {
		t.Fatalf("unaccounted events=%d err=%v", unaccounted, err)
	}
	if _, err := reopened.ReadRiskBucketState(context.Background(), key); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("drift read err=%v", err)
	}
}

func TestRiskBucketOrphanReservationMappingLatchesWithoutDroppingAuthoritativeFill(t *testing.T) {
	j, _, decisionID, reserved := riskBucketFillFixture(t, "orphan", "risk-orphan")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-orphan", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DROP TRIGGER risk_bucket_order_reservations_no_delete`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DELETE FROM risk_bucket_order_reservations WHERE order_key=(SELECT order_key FROM risk_bucket_orders WHERE order_id='risk-orphan') AND reservation_id LIKE '%:sector'`); err != nil {
		t.Fatal(err)
	}
	result, err := j.RecordFill(context.Background(), observation("risk-orphan", "1"))
	if err != nil || !result.Changed {
		t.Fatalf("fill=%+v err=%v", result, err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_fills"); got != 0 {
		t.Fatalf("accounted corrupt fills=%d", got)
	}
	if _, err := j.LookupFill(context.Background(), "risk-orphan"); err != nil {
		t.Fatalf("authoritative fill missing: %v", err)
	}
	var latch int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE latch='REPLAY_MISMATCH'`).Scan(&latch); err != nil || latch != 1 {
		t.Fatalf("latch=%d err=%v", latch, err)
	}
}

func TestRiskBucketAmbiguousSidecarLatchesAllApplicableReservationsWithoutDroppingFill(t *testing.T) {
	j, key, decisionID, reserved := riskBucketFillFixture(t, "ambiguous-sidecar", "risk-ambiguous")
	insertPosition(t, j, "risk-ambiguous-position", nil)
	if err := j.SetApplyHooks(ApplyHooks{Project: func(ctx context.Context, tx *ApplyTx, fill AppliedFill) error {
		_, err := tx.Exec(ctx, `UPDATE positions SET quantity=?,state='OPEN' WHERE id='risk-ambiguous-position'`, fill.CumulativeQuantity)
		return err
	}}); err != nil {
		t.Fatal(err)
	}
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-ambiguous", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	secondDecision := decisionID + ":corrupt-duplicate"
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_final_decisions(decision_id,transaction_id,account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence,created_at) SELECT ?,transaction_id||':corrupt',account_ref,market,symbol,q_candidate,q_existing_guardian,q_final,existing_reservation_id,request_digest,request_preimage,snapshot_set_digest,owner_prospective_generation,owner_lane_id,owner_campaign_id,owner_sequence+1,created_at FROM risk_bucket_final_decisions WHERE decision_id=?`, secondDecision, decisionID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_reservations(reservation_id,decision_id,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,risk_overage_latched,unknown_actual_latched,created_at,updated_at) SELECT reservation_id||':corrupt',?,existing_reservation_id,account_ref,market,symbol,owner_prospective_generation,bucket_dimension,bucket_value,policy_version,snapshot_id,reserved_minor,held_minor,filled_minor,overage_minor,state,risk_overage_latched,unknown_actual_latched,created_at,updated_at FROM risk_bucket_reservations WHERE decision_id=?`, secondDecision, decisionID); err != nil {
		t.Fatal(err)
	}
	secondOrderKey := riskBucketOrderKey(secondDecision, "risk-ambiguous")
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_orders(order_key,order_id,decision_id,order_quantity,cumulative_fill,quote_currency,base_currency,reservation_policy_digest,request_digest,state,created_at,updated_at) SELECT ?,'risk-ambiguous',?,order_quantity,0,quote_currency,base_currency,reservation_policy_digest,'corrupt-duplicate','ACTIVE',created_at,updated_at FROM risk_bucket_orders WHERE decision_id=?`, secondOrderKey, secondDecision, decisionID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO risk_bucket_order_reservations(order_key,reservation_id,reserved_minor) SELECT ?,reservation_id,reserved_minor FROM risk_bucket_reservations WHERE decision_id=?`, secondOrderKey, secondDecision); err != nil {
		t.Fatal(err)
	}

	result, err := j.RecordFill(context.Background(), observation("risk-ambiguous", "1"))
	if err != nil || !result.Changed {
		t.Fatalf("ambiguous sidecar dropped fill result=%+v err=%v", result, err)
	}
	if _, err := j.LookupFill(context.Background(), "risk-ambiguous"); err != nil {
		t.Fatalf("authoritative fill missing: %v", err)
	}
	var positionQuantity string
	if err := j.db.QueryRow(`SELECT quantity FROM positions WHERE id='risk-ambiguous-position'`).Scan(&positionQuantity); err != nil || positionQuantity != "1" {
		t.Fatalf("authoritative position quantity=%q err=%v", positionQuantity, err)
	}
	var ownerLatched, reservationsLatched, unaccounted int
	if err := j.db.QueryRow(`SELECT unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&ownerLatched); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? AND unknown_actual_latched=1`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&reservationsLatched); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE event_type='FILL_UNACCOUNTED'`).Scan(&unaccounted); err != nil {
		t.Fatal(err)
	}
	if ownerLatched != 1 || reservationsLatched != 10 || unaccounted != 1 {
		t.Fatalf("latches owner=%d reservations=%d unaccounted=%d", ownerLatched, reservationsLatched, unaccounted)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_fills"); got != 0 {
		t.Fatalf("ambiguous fill was falsely allocated: %d", got)
	}
}

func TestRiskBucketOwnershipAmbiguityLatchesEveryOwnerDecisionWithoutDroppingFill(t *testing.T) {
	j, key, decisionID, reserved := riskBucketFillFixture(t, "ownership-ambiguous", "risk-owner-ambiguous")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-owner-ambiguous", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	commitRiskBucketScaleIn(t, j, key, "ownership-ambiguous-second", "200", "50")
	recordConfirmedFillOrder(t, j, "risk-owner-conflict-intent", "risk-owner-conflict-attempt", "risk-owner-ambiguous")

	result, err := j.RecordFill(context.Background(), observation("risk-owner-ambiguous", "1"))
	if err != nil || !result.Changed {
		t.Fatalf("ambiguous ownership dropped fill result=%+v err=%v", result, err)
	}
	if _, err := j.LookupFill(context.Background(), "risk-owner-ambiguous"); err != nil {
		t.Fatalf("authoritative fill missing: %v", err)
	}
	var ownerLatched, reservationsLatched, unaccounted int
	if err := j.db.QueryRow(`SELECT unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&ownerLatched); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=? AND unknown_actual_latched=1`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&reservationsLatched); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE event_type='FILL_UNACCOUNTED'`).Scan(&unaccounted); err != nil {
		t.Fatal(err)
	}
	if ownerLatched != 1 || reservationsLatched != 10 || unaccounted != 1 {
		t.Fatalf("latches owner=%d reservations=%d unaccounted=%d", ownerLatched, reservationsLatched, unaccounted)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_fills"); got != 0 {
		t.Fatalf("ambiguous ownership was falsely allocated: %d", got)
	}
}

func TestRiskReducingFillNeverRequiresRiskBucketAccounting(t *testing.T) {
	j := openTestJournal(t)
	recordConfirmedFillOrderScope(t, j, "sell-intent", "sell-attempt", "sell-order", FillSnapshotScope{AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30", Symbol: "AAPL", Side: "SELL"})
	obs := observation("sell-order", "1")
	obs.Side = "SELL"
	if result, err := j.RecordFill(context.Background(), obs); err != nil || !result.Changed {
		t.Fatalf("sell fill=%+v err=%v", result, err)
	}
}

func TestRiskBucketUnsafeEvidenceAndReleaseMethodsAreNotExported(t *testing.T) {
	typeOfJournal := reflect.TypeOf(&Journal{})
	for _, name := range []string{"CompleteRiskBucketFillActual", "ReleaseRiskBucketOrder"} {
		if _, exists := typeOfJournal.MethodByName(name); exists {
			t.Fatalf("unsafe caller-authorized method %s is exported", name)
		}
	}
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, call := range []string{".completeRiskBucketFillActual(", ".releaseRiskBucketOrder("} {
			if strings.Contains(string(raw), call) {
				t.Fatalf("production caller exists in %s: %s", path, call)
			}
		}
	}
}

func TestRiskBucketOrderRemainingUsesExact256BitStringArithmetic(t *testing.T) {
	j, _, decisionID, reserved := riskBucketFillFixture(t, "big-minor", "risk-big")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-big", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), observation("risk-big", "4")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DROP TRIGGER risk_bucket_order_reservations_no_update`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DROP TRIGGER risk_bucket_fill_allocations_identity_guard`); err != nil {
		t.Fatal(err)
	}
	reservedHuge := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(100)).String()
	transferHuge := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(40)).String()
	if _, err := j.db.Exec(`UPDATE risk_bucket_order_reservations SET reserved_minor=? WHERE order_key=(SELECT order_key FROM risk_bucket_orders WHERE order_id='risk-big')`, reservedHuge); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_fill_allocations SET transfer_minor=? WHERE fill_id IN (SELECT fill_id FROM risk_bucket_fills WHERE order_id='risk-big')`, transferHuge); err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var orderKey string
	if err := tx.QueryRow(`SELECT order_key FROM risk_bucket_orders WHERE order_id='risk-big'`).Scan(&orderKey); err != nil {
		t.Fatal(err)
	}
	remaining, err := riskBucketOrderRemaining(context.Background(), tx, orderKey)
	if err != nil {
		t.Fatal(err)
	}
	for key, amount := range remaining {
		if amount != "60" {
			t.Fatalf("%+v remaining=%s", key, amount)
		}
	}
}

func TestRiskBucketActualAndReleaseRequireExactOwnerDecisionScopeForCollidingOrderID(t *testing.T) {
	j, usKey, usDecision, usReserved := riskBucketFillFixture(t, "scope-us", "shared-risk-order")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "shared-risk-order", DecisionID: usDecision, OrderQuantity: 10, ReservedMinor: usReserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-scope-kr", "acct-1")
	krPlan := riskBucketAdmissionFixture(t, "scope-kr", "acct-1", "lane-kr", "campaign-kr", "prospective-scope-kr", "100", "0")
	krPlan.ExistingReservationID = "existing-scope-kr"
	for i, bucket := range krPlan.Admission.Buckets {
		key := bucket.Key
		key.PolicyVersion = "policy-kr-v1"
		rebindRiskBucket(t, &krPlan, i, key)
	}
	krReceipt, err := j.CommitRiskBucketAdmission(context.Background(), krPlan)
	if err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "risk-scope-kr-intent", "risk-scope-kr-attempt", "shared-risk-order", FillSnapshotScope{AccountRef: "acct-1", Market: "kr", TradingDay: "2026-03-30", Symbol: "005930", Side: "BUY"})
	bindRiskOrderAttemptDecision(t, j, "shared-risk-order", krReceipt.DecisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "shared-risk-order", DecisionID: krReceipt.DecisionID, OrderQuantity: 10, ReservedMinor: riskReservedFromPlan(krPlan, "50"), ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	release, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: usKey, DecisionID: usDecision, OrderID: "shared-risk-order", Reason: RiskBucketReleaseExpiry, ReleasedAt: riskFillNow.Add(time.Second)})
	if err != nil || !release.Released {
		t.Fatalf("release=%+v err=%v", release, err)
	}
	var usState, krState string
	if err := j.db.QueryRow(`SELECT state FROM risk_bucket_orders WHERE decision_id=? AND order_id='shared-risk-order'`, usDecision).Scan(&usState); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT state FROM risk_bucket_orders WHERE decision_id=? AND order_id='shared-risk-order'`, krReceipt.DecisionID).Scan(&krState); err != nil {
		t.Fatal(err)
	}
	if usState != "RELEASED" || krState != "ACTIVE" {
		t.Fatalf("US/KR state=%s/%s", usState, krState)
	}
	wrong := RiskBucketActualFillPlan{Owner: krPlan.Owner.Key, DecisionID: usDecision, OrderID: "shared-risk-order", CumulativeFill: 1, Actual: riskBucketActual("5", "1", "0"), ObservedAt: riskFillNow}
	if _, err := j.completeRiskBucketFillActual(context.Background(), wrong); !errors.Is(err, ErrRiskBucketStateUnknown) {
		t.Fatalf("cross-scope actual error=%v", err)
	}
}

func TestRiskBucketOwnerAggregateAccountsTwoDecisionOrdersAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openTestJournalAt(t, path)
	key, firstDecision, firstReserved := seedRiskBucketFillFixture(t, j, "multi-decision", "risk-multi-one")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-multi-one", DecisionID: firstDecision, OrderQuantity: 10, ReservedMinor: firstReserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	secondReceipt, secondReserved := commitRiskBucketScaleIn(t, j, key, "multi-second", "200", "50")
	recordConfirmedFillOrder(t, j, "risk-multi-two-intent", "risk-multi-two-attempt", "risk-multi-two")
	bindRiskOrderAttemptDecision(t, j, "risk-multi-two", secondReceipt.DecisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-multi-two", DecisionID: secondReceipt.DecisionID, OrderQuantity: 10, ReservedMinor: secondReserved, CreatedAt: riskFillNow.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), observation("risk-multi-one", "4")); err != nil {
		t.Fatal(err)
	}
	secondFill := observation("risk-multi-two", "2")
	if _, err := j.RecordFill(context.Background(), secondFill); err != nil {
		t.Fatal(err)
	}
	state, err := j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	if state.QFinal != 20 {
		t.Fatalf("aggregate q_final=%d", state.QFinal)
	}
	assertRiskUsage(t, state, "70", "30", false, true)
	for _, check := range []struct{ orderID, decisionID string }{{"risk-multi-one", firstDecision}, {"risk-multi-two", secondReceipt.DecisionID}} {
		var own, cross int
		if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_fill_allocations a JOIN risk_bucket_fills f ON f.fill_id=a.fill_id JOIN risk_bucket_reservations r ON r.reservation_id=a.reservation_id WHERE f.order_id=? AND r.decision_id=?`, check.orderID, check.decisionID).Scan(&own); err != nil {
			t.Fatal(err)
		}
		if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_fill_allocations a JOIN risk_bucket_fills f ON f.fill_id=a.fill_id JOIN risk_bucket_reservations r ON r.reservation_id=a.reservation_id WHERE f.order_id=? AND r.decision_id<>?`, check.orderID, check.decisionID).Scan(&cross); err != nil {
			t.Fatal(err)
		}
		if own != 5 || cross != 0 {
			t.Fatalf("%s allocation own/cross=%d/%d", check.orderID, own, cross)
		}
	}
	for _, check := range []struct{ decisionID, held, filled string }{
		{firstDecision, "30", "20"},
		{secondReceipt.DecisionID, "40", "10"},
	} {
		var matches int
		if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND held_minor=? AND filled_minor=?`, check.decisionID, check.held, check.filled).Scan(&matches); err != nil {
			t.Fatal(err)
		}
		if matches != 5 {
			t.Fatalf("decision %s exact held/filled rows=%d", check.decisionID, matches)
		}
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(context.Background(), Options{Path: path, Clock: clock.NewFake(riskFillNow), FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if state, err := reopened.ReadRiskBucketState(context.Background(), key); err != nil || state.QFinal != 20 {
		t.Fatalf("restart aggregate state=%+v err=%v", state, err)
	}
	for _, actual := range []RiskBucketActualFillPlan{
		{Owner: key, DecisionID: firstDecision, OrderID: "risk-multi-one", CumulativeFill: 4, Actual: riskBucketActual("5", "1", "0"), ObservedAt: riskFillNow.Add(2 * time.Second)},
		{Owner: key, DecisionID: secondReceipt.DecisionID, OrderID: "risk-multi-two", CumulativeFill: 2, Actual: riskBucketActual("5", "1", "0"), ObservedAt: riskFillNow.Add(3 * time.Second)},
	} {
		if result, err := reopened.completeRiskBucketFillActual(context.Background(), actual); err != nil || !result.ActualEvidenceCompleted {
			t.Fatalf("actual completion=%+v err=%v", result, err)
		}
	}
	state, err = reopened.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "70", "30", false, false)
}

func TestRiskBucketLateFillCannotConsumeSiblingDecisionHeld(t *testing.T) {
	j, key, firstDecision, firstReserved := riskBucketFillFixture(t, "multi-late", "risk-multi-late-one")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-multi-late-one", DecisionID: firstDecision, OrderQuantity: 10, ReservedMinor: firstReserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	secondReceipt, secondReserved := commitRiskBucketScaleIn(t, j, key, "multi-late-second", "200", "50")
	recordConfirmedFillOrder(t, j, "risk-multi-late-two-intent", "risk-multi-late-two-attempt", "risk-multi-late-two")
	bindRiskOrderAttemptDecision(t, j, "risk-multi-late-two", secondReceipt.DecisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-multi-late-two", DecisionID: secondReceipt.DecisionID, OrderQuantity: 10, ReservedMinor: secondReserved, CreatedAt: riskFillNow.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	if result, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: key, DecisionID: firstDecision, OrderID: "risk-multi-late-one", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(2 * time.Second)}); err != nil || !result.Released {
		t.Fatalf("first release=%+v err=%v", result, err)
	}
	if result, err := j.RecordFill(context.Background(), observation("risk-multi-late-one", "1")); err != nil || !result.Changed {
		t.Fatalf("late fill=%+v err=%v", result, err)
	}
	state, err := j.ReadRiskBucketState(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	assertRiskUsage(t, state, "50", "5", true, true)
	for _, check := range []struct{ decisionID, held string }{{firstDecision, "0"}, {secondReceipt.DecisionID, "50"}} {
		var matches int
		if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_reservations WHERE decision_id=? AND held_minor=?`, check.decisionID, check.held).Scan(&matches); err != nil {
			t.Fatal(err)
		}
		if matches != 5 {
			t.Fatalf("decision %s held %s rows=%d", check.decisionID, check.held, matches)
		}
	}
}

func TestRiskBucketScaleInAdmissionWhileOrderActiveStillRejectsBucketDrift(t *testing.T) {
	j, key, decisionID, reserved := riskBucketFillFixture(t, "active-scale", "risk-active-scale")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-active-scale", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	seedExistingRiskReservation(t, j, "existing-active-scale-second", "acct-1")
	second := riskBucketAdmissionFixture(t, "active-scale-second", "acct-1", "lane-short", "campaign-1", key.ProspectiveGeneration, "200", "50")
	second.ExistingReservationID = "existing-active-scale-second"
	second.Owner.Key = key
	second.Admission.Policy.QuoteCurrency = "USD"
	second.Admission.Policy.AccountCurrency = "KRW"
	rebindRiskBucket(t, &second, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &second, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &second, 2, riskbucket.BucketKey{Dimension: riskbucket.DimensionStrategy, Value: "strategy-drift", PolicyVersion: "policy-v1"})
	if _, err := j.CommitRiskBucketAdmission(context.Background(), second); !errors.Is(err, ErrRiskBucketSnapshotMismatch) {
		t.Fatalf("active scale-in drift error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_final_decisions"); got != 1 {
		t.Fatalf("decisions=%d", got)
	}
}

func TestRiskBucketOrderRegistrationRequiresExactConfirmedOfficialOrderAuthority(t *testing.T) {
	j, _, decisionID, reserved := riskBucketFillFixture(t, "authority", "risk-authority")
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET decision_id=NULL WHERE broker_order_id='risk-authority'`); err != nil {
		t.Fatal(err)
	}
	err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-authority", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "reservation-policy-v1", QuoteCurrency: "KRW", BaseCurrency: "KRW", CreatedAt: riskFillNow})
	if !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("authority error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_orders"); got != 0 {
		t.Fatalf("orders=%d", got)
	}
}

func TestRiskBucketOrderRegistrationRejectsQuantityDivergentFromConfirmedIntent(t *testing.T) {
	j, _, decisionID, reserved := riskBucketFillFixture(t, "authority-quantity", "risk-authority-quantity")
	if _, err := j.db.Exec(`UPDATE intents SET quantity='9' WHERE id=(SELECT intent_id FROM mutation_attempts WHERE broker_order_id='risk-authority-quantity')`); err != nil {
		t.Fatal(err)
	}
	err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
		OrderID: "risk-authority-quantity", DecisionID: decisionID, OrderQuantity: 10,
		ReservedMinor: reserved, CreatedAt: riskFillNow,
	})
	if !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("authority quantity error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_orders"); got != 0 {
		t.Fatalf("orders=%d", got)
	}
}

func TestRiskBucketOrderRegistrationRejectsBrokerOrderIDCollisionAcrossOwnerDecisions(t *testing.T) {
	j, key, firstDecision, firstReserved := riskBucketFillFixture(t, "owner-order-collision", "risk-owner-order-collision")
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-owner-order-collision", DecisionID: firstDecision, OrderQuantity: 10, ReservedMinor: firstReserved, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	secondReceipt, secondReserved := commitRiskBucketScaleIn(t, j, key, "owner-order-collision-second", "200", "50")
	recordConfirmedFillOrder(t, j, "risk-owner-order-collision-second-intent", "risk-owner-order-collision-second-attempt", "risk-owner-order-collision")
	var legacyDecisionID string
	if err := j.db.QueryRow(`SELECT legacy.decision_id FROM risk_bucket_final_decisions d JOIN risk_reservations legacy ON legacy.id=d.existing_reservation_id WHERE d.decision_id=?`, secondReceipt.DecisionID).Scan(&legacyDecisionID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET decision_id=? WHERE id='risk-owner-order-collision-second-attempt'`, legacyDecisionID); err != nil {
		t.Fatal(err)
	}
	err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-owner-order-collision", DecisionID: secondReceipt.DecisionID, OrderQuantity: 10, ReservedMinor: secondReserved, CreatedAt: riskFillNow.Add(time.Second)})
	if !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("owner broker order collision error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_orders"); got != 1 {
		t.Fatalf("orders=%d", got)
	}
}

func TestRiskBucketOrderRegistrationDerivesCanonicalPolicyCurrencyAndBindsDivergentRetry(t *testing.T) {
	j, _, decisionID, reserved := riskBucketFillFixture(t, "canonical-order", "risk-canonical")
	plan := RiskBucketOrderPlan{OrderID: "risk-canonical", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, ReservationPolicyDigest: "caller-controlled", QuoteCurrency: "KRW", BaseCurrency: "USD", CreatedAt: riskFillNow}
	if err := j.RegisterRiskBucketOrder(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	authority, err := loadRiskBucketOrderAuthority(context.Background(), tx, decisionID)
	tx.Rollback()
	if err != nil {
		t.Fatal(err)
	}
	var storedPolicy, storedQuote, storedBase string
	if err := j.db.QueryRow(`SELECT reservation_policy_digest,quote_currency,base_currency FROM risk_bucket_orders WHERE decision_id=?`, decisionID).Scan(&storedPolicy, &storedQuote, &storedBase); err != nil {
		t.Fatal(err)
	}
	if storedPolicy != authority.digest || storedQuote != "USD" || storedBase != "KRW" {
		t.Fatalf("stored authority policy=%q quote/base=%s/%s expected=%q USD/KRW", storedPolicy, storedQuote, storedBase, authority.digest)
	}
	divergent := plan
	divergent.ReservedMinor = riskReservedMap("49")
	if err := j.RegisterRiskBucketOrder(context.Background(), divergent); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("divergent reservation retry error=%v", err)
	}
	divergent = plan
	divergent.OrderQuantity = 11
	if err := j.RegisterRiskBucketOrder(context.Background(), divergent); !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("divergent quantity retry error=%v", err)
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_orders"); got != 1 {
		t.Fatalf("orders after divergent retries=%d", got)
	}
}

func TestRiskBucketReplacementRequiresOneExactActivePredecessor(t *testing.T) {
	j, _, decisionID, reserved := riskBucketFillFixture(t, "single-parent", "risk-parent-once")
	parent := RiskBucketOrderPlan{OrderID: "risk-parent-once", DecisionID: decisionID, OrderQuantity: 10, ReservedMinor: reserved, CreatedAt: riskFillNow}
	if err := j.RegisterRiskBucketOrder(context.Background(), parent); err != nil {
		t.Fatal(err)
	}
	for _, orderID := range []string{"risk-child-one", "risk-child-two"} {
		recordConfirmedFillOrder(t, j, orderID+"-intent", orderID+"-attempt", orderID)
		bindRiskOrderAttemptDecision(t, j, orderID, decisionID)
		plan := RiskBucketOrderPlan{OrderID: orderID, DecisionID: decisionID, PredecessorOrderID: "risk-parent-once", OrderQuantity: 10, ReservedMinor: reserved, CreatedAt: riskFillNow.Add(time.Second)}
		err := j.RegisterRiskBucketOrder(context.Background(), plan)
		if orderID == "risk-child-one" && err != nil {
			t.Fatal(err)
		}
		if orderID == "risk-child-two" && !errors.Is(err, ErrRiskBucketReplayMismatch) {
			t.Fatalf("reused predecessor error=%v", err)
		}
	}
	if got := countRiskBucketRows(t, j, "risk_bucket_orders"); got != 2 {
		t.Fatalf("orders after predecessor reuse=%d", got)
	}

	j2, key2, decision2, reserved2 := riskBucketFillFixture(t, "terminal-parent", "risk-terminal-parent")
	if err := j2.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-terminal-parent", DecisionID: decision2, OrderQuantity: 10, ReservedMinor: reserved2, CreatedAt: riskFillNow}); err != nil {
		t.Fatal(err)
	}
	if _, err := j2.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{Owner: key2, DecisionID: decision2, OrderID: "risk-terminal-parent", Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(time.Second)}); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j2, "terminal-child-intent", "terminal-child-attempt", "risk-terminal-child")
	bindRiskOrderAttemptDecision(t, j2, "risk-terminal-child", decision2)
	err := j2.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{OrderID: "risk-terminal-child", DecisionID: decision2, PredecessorOrderID: "risk-terminal-parent", OrderQuantity: 10, ReservedMinor: reserved2, CreatedAt: riskFillNow.Add(2 * time.Second)})
	if !errors.Is(err, ErrRiskBucketReplayMismatch) {
		t.Fatalf("terminal predecessor error=%v", err)
	}
	if got := countRiskBucketRows(t, j2, "risk_bucket_orders"); got != 1 {
		t.Fatalf("orders after terminal predecessor=%d", got)
	}
}

var riskFillNow = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

func riskBucketFillFixture(t *testing.T, suffix, orderID string) (*Journal, riskbucket.OwnerKey, string, map[riskbucket.BucketKey]string) {
	t.Helper()
	j := openTestJournal(t)
	key, decision, reserved := seedRiskBucketFillFixture(t, j, suffix, orderID)
	return j, key, decision, reserved
}
func seedRiskBucketFillFixture(t *testing.T, j *Journal, suffix, orderID string) (riskbucket.OwnerKey, string, map[riskbucket.BucketKey]string) {
	t.Helper()
	seedExistingRiskReservation(t, j, "existing-fill-"+suffix, "acct-1")
	plan := riskBucketAdmissionFixture(t, "fill-"+suffix, "acct-1", "lane-short", "campaign-1", "prospective-fill-"+suffix, "100", "0")
	plan.ExistingReservationID = "existing-fill-" + suffix
	plan.Owner.Key.Market = riskbucket.MarketUS
	plan.Owner.Key.Symbol = "AAPL"
	plan.Admission.Policy.QuoteCurrency = "USD"
	plan.Admission.Policy.AccountCurrency = "KRW"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(riskbucket.MarketUS), PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	receipt, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrder(t, j, "risk-intent-"+suffix, "risk-attempt-"+suffix, orderID)
	bindRiskOrderAttemptDecision(t, j, orderID, receipt.DecisionID)
	return plan.Owner.Key, receipt.DecisionID, riskReservedMap("50")
}

func commitRiskBucketScaleIn(t *testing.T, j *Journal, key riskbucket.OwnerKey, suffix, limit, held string) (RiskBucketAdmissionReceipt, map[riskbucket.BucketKey]string) {
	t.Helper()
	existingID := "existing-fill-" + suffix
	seedExistingRiskReservation(t, j, existingID, key.AccountID)
	plan := riskBucketAdmissionFixture(t, "fill-"+suffix, key.AccountID, "lane-short", "campaign-1", key.ProspectiveGeneration, limit, held)
	plan.ExistingReservationID = existingID
	plan.Owner.Key = key
	plan.Admission.Policy.QuoteCurrency = "USD"
	plan.Admission.Policy.AccountCurrency = "KRW"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: string(key.Market), PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: key.Symbol, PolicyVersion: "policy-v1"})
	receipt, err := j.CommitRiskBucketAdmission(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	return receipt, riskReservedFromPlan(plan, "50")
}

func bindRiskOrderAttemptDecision(t *testing.T, j *Journal, orderID, riskDecisionID string) {
	t.Helper()
	result, err := j.db.Exec(`UPDATE mutation_attempts SET decision_id=(SELECT legacy.decision_id FROM risk_bucket_final_decisions d JOIN risk_reservations legacy ON legacy.id=d.existing_reservation_id WHERE d.decision_id=?) WHERE id IN (SELECT a.id FROM mutation_attempts a JOIN intents i ON i.id=a.intent_id JOIN risk_bucket_final_decisions d ON d.decision_id=? WHERE a.broker_order_id=? AND TRIM(i.account_ref)=d.account_ref AND UPPER(TRIM(i.market))=d.market AND UPPER(TRIM(i.symbol))=d.symbol)`, riskDecisionID, riskDecisionID, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if count, err := result.RowsAffected(); err != nil || count != 1 {
		t.Fatalf("bound attempts=%d err=%v", count, err)
	}
}

func riskReservedMap(amount string) map[riskbucket.BucketKey]string {
	return map[riskbucket.BucketKey]string{
		{Dimension: riskbucket.DimensionHorizon, Value: "SHORT", PolicyVersion: "policy-v1"}:           amount,
		{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"}:               amount,
		{Dimension: riskbucket.DimensionStrategy, Value: "strategy-alpha", PolicyVersion: "policy-v1"}: amount,
		{Dimension: riskbucket.DimensionSector, Value: "sector-tech", PolicyVersion: "policy-v1"}:      amount,
		{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"}:             amount,
	}
}
func riskReservedMapKR(amount string) map[riskbucket.BucketKey]string {
	return map[riskbucket.BucketKey]string{
		{Dimension: riskbucket.DimensionHorizon, Value: "SHORT", PolicyVersion: "policy-v1"}: amount, {Dimension: riskbucket.DimensionMarket, Value: "KR", PolicyVersion: "policy-v1"}: amount, {Dimension: riskbucket.DimensionStrategy, Value: "strategy-alpha", PolicyVersion: "policy-v1"}: amount, {Dimension: riskbucket.DimensionSector, Value: "sector-tech", PolicyVersion: "policy-v1"}: amount, {Dimension: riskbucket.DimensionSymbol, Value: "005930", PolicyVersion: "policy-v1"}: amount,
	}
}
func riskReservedFromPlan(plan RiskBucketAdmissionPlan, amount string) map[riskbucket.BucketKey]string {
	out := make(map[riskbucket.BucketKey]string, len(plan.Admission.Buckets))
	for _, bucket := range plan.Admission.Buckets {
		out[bucket.Key] = amount
	}
	return out
}
func riskBucketActual(price, fx, fee string) *riskbucket.ActualFillEvidence {
	return &riskbucket.ActualFillEvidence{QuoteCurrency: "USD", BaseCurrency: "KRW", PriceQuote: price, FXRateQuoteToBase: fx, AllocatedFeeBaseMinor: fee, Price: riskbucket.Evidence{Source: "official-fill", Version: "v1", Digest: "price-digest", Official: true, Frozen: true, ObservedAt: riskFillNow, FreshUntil: riskFillNow.Add(time.Minute)}, FX: riskbucket.Evidence{Source: "official-fx", Version: "v1", Digest: "fx-digest", Official: true, Frozen: true, ObservedAt: riskFillNow, FreshUntil: riskFillNow.Add(time.Minute)}, EvaluatedAt: riskFillNow}
}
func assertRiskUsage(t *testing.T, state RiskBucketState, held, filled string, overage, unknown bool) {
	t.Helper()
	if len(state.Usage) != 5 {
		t.Fatalf("usage=%+v", state.Usage)
	}
	for key, usage := range state.Usage {
		if usage.HeldMinor != held || usage.FilledMinor != filled || usage.Latches[riskbucket.LatchRiskOverage] != overage || usage.Latches[riskbucket.LatchUnknownActualRisk] != unknown {
			t.Fatalf("%+v usage=%+v want held=%s filled=%s overage=%v unknown=%v", key, usage, held, filled, overage, unknown)
		}
	}
	if state.OwnerLatches[riskbucket.LatchRiskOverage] != overage || state.OwnerLatches[riskbucket.LatchUnknownActualRisk] != unknown {
		t.Fatalf("owner latches=%+v want overage=%v unknown=%v", state.OwnerLatches, overage, unknown)
	}
}
