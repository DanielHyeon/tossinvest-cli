package execgw_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Persisted-decision tests (extend-execution-contract tasks 1.4 and 1.5).
//
// Two requirements are under test here. "결정 영속과 신뢰 경계": the gateway
// verifies against the row an issuer wrote, re-read immediately before the
// broker call, and never against anything the submitter supplied. And "결정의
// Safety Class와 형태 일치": the class is checked against what the mutation
// actually does, so it cannot be worn as a badge to skip the limits.

// --- tampering with the record ----------------------------------------------

// tamperDecision rewrites one column of a persisted decision behind the
// journal's back. It is the only way to build a decision the issuer API refuses
// to write, which is exactly what "the record is the authority" has to survive.
func tamperDecision(t *testing.T, j *journal.Journal, id, column string, value any) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+j.Path())
	if err != nil {
		t.Fatalf("opening the journal file: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(),
		"UPDATE decisions SET "+column+" = ? WHERE id = ?", value, id); err != nil {
		t.Fatalf("tampering with %s: %v", column, err)
	}
}

func deleteDecision(t *testing.T, j *journal.Journal, id string) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+j.Path())
	if err != nil {
		t.Fatalf("opening the journal file: %v", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), "DELETE FROM decisions WHERE id = ?", id); err != nil {
		t.Fatalf("deleting the decision: %v", err)
	}
}

func attemptHistory(t *testing.T, j *journal.Journal, attemptID string) []string {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+j.Path())
	if err != nil {
		t.Fatalf("opening the journal file: %v", err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(),
		"SELECT from_state, to_state FROM attempt_transitions WHERE attempt_id = ? ORDER BY id", attemptID)
	if err != nil {
		t.Fatalf("reading the transition history: %v", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var from, to string
		if err := rows.Scan(&from, &to); err != nil {
			t.Fatal(err)
		}
		out = append(out, from+"->"+to)
	}
	return out
}

// TestDecisionReferenceCarriesNoRiskData is the trust boundary as a compile-time
// fact: there is no field on the value a submitter passes that the verification
// could be talked into believing. If this test has to change, the boundary is
// being opened (engine-safety: "제출 호출자가 공급한 위험 데이터로 재검증해서는
// 안 된다").
func TestDecisionReferenceCarriesNoRiskData(t *testing.T) {
	typ := reflect.TypeOf(execgw.GuardianDecision{})
	want := map[string]bool{"ID": true, "Generation": true}
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		if !want[name] {
			t.Errorf("GuardianDecision.%s: a submission may reference a decision and nothing else — "+
				"risk data on this struct would make the verification circular", name)
		}
		delete(want, name)
	}
	for name := range want {
		t.Errorf("GuardianDecision.%s is missing", name)
	}
}

// TestSwappedStopPriceIsRefused is the spec's "손절 데이터 바꿔치기": the risk
// data the decision was justified by is rewritten after it was issued, and the
// hash over the canonical text no longer covers it.
func TestSwappedStopPriceIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	original, err := j.LookupDecision(context.Background(), req.Decision.ID)
	if err != nil {
		t.Fatal(err)
	}
	swapped := strings.Replace(original.RiskPreimage, `"stop_price":"63000"`, `"stop_price":"1"`, 1)
	if swapped == original.RiskPreimage {
		t.Fatalf("the fixture no longer contains the stop price: %s", original.RiskPreimage)
	}
	tamperDecision(t, j, req.Decision.ID, "risk_preimage", swapped)

	out, err := gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) {
		t.Fatalf("want a RejectedError, got %v", err)
	}
	if rejected.Reason != execgw.ReasonGuardianDecisionTampered {
		t.Errorf("reason: got %q, want %q", rejected.Reason, execgw.ReasonGuardianDecisionTampered)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("state: got %s, want NOT_DISPATCHED", out.State)
	}
}

// TestForeignIdempotencyKeyIsRefused: a key that is not derivable from the
// decision it sits on is refused. The key is what a replay addresses the order
// by, so one nobody can re-derive is an order nobody can recover.
func TestForeignIdempotencyKeyIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	tamperDecision(t, j, req.Decision.ID, "client_order_id",
		journal.DeriveClientOrderID("some-other-decision", 0))

	_, err := gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianKeyMismatch {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianKeyMismatch, err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
}

// hookAccount runs a callback when the preflight reads the balance, which is the
// window between the gateway loading the decision and the dispatch closure
// re-reading it.
type hookAccount struct {
	buyingPower float64
	hook        func()
}

func (a *hookAccount) BuyingPower(context.Context, string) (float64, error) {
	if a.hook != nil {
		a.hook()
	}
	return a.buyingPower, nil
}

func (a *hookAccount) HoldingQuantity(context.Context, string) (float64, error) { return 0, nil }

func gatewayWithPreflight(t *testing.T, broker trading.Broker, account execgw.AccountReader) (
	*execgw.Gateway, *journal.Journal, *clock.Fake,
) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
		Preflight: &execgw.Preflight{Account: account},
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw, j, clk
}

// TestExpiryIsRecheckedOnAFreshRowBeforeTheBrokerCall is the spec's "브로커 호출
// 직전 재검증", made observable: the decision is still valid when the gateway
// first checks it and expires before the dispatch. The refusal has to come from
// the second check, and the transition history proves it did — the attempt
// reached DISPATCH_STARTED first.
func TestExpiryIsRecheckedOnAFreshRowBeforeTheBrokerCall(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	account := &hookAccount{buyingPower: 10_000_000}
	gw, j, clk := gatewayWithPreflight(t, broker, account)
	req := placeRequest(t, j, clk)

	account.hook = func() {
		// Expire the persisted decision after the gateway has read it once.
		tamperDecision(t, j, req.Decision.ID, "expires_at",
			fixedNow.Add(-time.Second).UTC().Format(time.RFC3339))
	}

	out, err := gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianExpired {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianExpired, err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0 — the expiry must be caught before the call", places)
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("state: got %s, want NOT_DISPATCHED", out.State)
	}
	history := strings.Join(attemptHistory(t, j, out.AttemptID), ",")
	if !strings.Contains(history, "RECORDED->DISPATCH_STARTED") {
		t.Fatalf("history %q: the refusal must come from the re-check inside the dispatch, "+
			"which only happens after DISPATCH_STARTED", history)
	}
}

// TestRevokedDecisionIsRefusedAtTheLastMoment is the same window with the row
// removed rather than expired: the verification reads the journal, so a decision
// withdrawn after the attempt was recorded cannot still be submitted from a copy
// held in memory.
func TestRevokedDecisionIsRefusedAtTheLastMoment(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	account := &hookAccount{buyingPower: 10_000_000}
	gw, j, clk := gatewayWithPreflight(t, broker, account)
	req := placeRequest(t, j, clk)

	var armed bool
	account.hook = func() {
		if armed {
			return
		}
		armed = true
		// The attempt row references the decision, so the reference has to go
		// first for the delete to be possible at all.
		db, err := sql.Open("sqlite", "file:"+j.Path())
		if err != nil {
			t.Errorf("opening the journal file: %v", err)
			return
		}
		defer db.Close()
		if _, err := db.Exec("UPDATE mutation_attempts SET decision_id = NULL"); err != nil {
			t.Errorf("detaching the attempt: %v", err)
			return
		}
		deleteDecision(t, j, req.Decision.ID)
	}

	out, err := gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianMissing {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianMissing, err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("state: got %s, want NOT_DISPATCHED", out.State)
	}
}

// --- class ⇔ shape (task 1.5) ------------------------------------------------

// TestExitOverTheOrderLimitIsNotRefused is the spec's "주문 한도를 초과하는
// 청산": a liquidation bigger than any per-order limit goes through, because a
// RISK_REDUCING decision carries no limit snapshot and is not measured against
// one. Under the old rule this held for a cancel and failed for a sell, which is
// exactly the case an emergency exit is made of.
func TestExitOverTheOrderLimitIsNotRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)

	sell := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 100_000, Price: 70000, CurrencyMode: "KRW",
	}
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent: sell,
		Decision: exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL",
			sell.Quantity),
	})
	if err != nil {
		t.Fatalf("an exit must not be limit-checked: %v", err)
	}
	if out.State != journal.StateConfirmed {
		t.Fatalf("state: got %s, want CONFIRMED (%s)", out.State, out.Detail)
	}
}

// TestExitAboveItsCeilingIsRefused is the other side of the same rule: the
// decision's ceiling still binds, because selling more than was authorised is
// how an exit becomes a short position.
func TestExitAboveItsCeilingIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)

	sell := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 10, Price: 70000, CurrencyMode: "KRW",
	}
	_, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   sell,
		Decision: exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL", 4),
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianIntentMismatch {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianIntentMismatch, err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
}

// TestSmallerExitThanAuthorisedIsAllowed: an exit sized from a fresher sellable
// quantity is more conservative, and §0.3 does not permit refusing it.
func TestSmallerExitThanAuthorisedIsAllowed(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)

	sell := orderintent.PlaceIntent{
		Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 3, Price: 70000, CurrencyMode: "KRW",
	}
	if _, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   sell,
		Decision: exitDecision(t, j, clk, journal.KindPlace, "kr", "005930", "SELL", 10),
	}); err != nil {
		t.Fatalf("a smaller exit than authorised must go through: %v", err)
	}
}

// TestCancelWithAnEntryDecisionIsRefused is the mirror of the forged-class case:
// the class has to match the shape in both directions, or "EXPOSURE_RAISING"
// would become a way to attach a limit snapshot to something that ignores it.
func TestCancelWithAnEntryDecisionIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-9"}}
	gw, j, clk := newGateway(t, broker)

	intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	_, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent: intent,
		Order:  execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: raisingDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY",
			2, 70000, testLimits()),
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianClassMismatch {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianClassMismatch, err)
	}
	if _, cancels, _ := broker.totals(); cancels != 0 {
		t.Errorf("broker cancel calls: got %d, want 0", cancels)
	}
}

// TestEntryWithoutALimitSnapshotIsRefused is the spec's "한도 없는 진입 결정".
// The issuer API refuses to write one, so the row has to be emptied behind its
// back — and the gateway still refuses.
func TestEntryWithoutALimitSnapshotIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	tamperDecision(t, j, req.Decision.ID, "limits_json", nil)

	_, err := gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianLimitsUnset {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianLimitsUnset, err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
}

// TestExitCarryingLimitsIsRefused: a RISK_REDUCING decision with a limit
// snapshot is a row no writer in this build produces, and it is the shape that
// would put a refusable limit on an exit (§0.3).
func TestExitCarryingLimitsIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-9"}}
	gw, j, clk := newGateway(t, broker)

	intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	decision := exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2)
	tamperDecision(t, j, decision.ID, "limits_json", `{"max_quantity":1}`)

	_, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   intent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: decision,
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianDecisionTampered {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianDecisionTampered, err)
	}
}

// --- the recorded binding ----------------------------------------------------

// TestAttemptRecordsTheDecisionBindingAndTheWireBody: everything the replay
// procedure will need is on disk before the broker is called.
func TestAttemptRecordsTheDecisionBindingAndTheWireBody(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	out, err := gw.Place(context.Background(), req)
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.DecisionID != req.Decision.ID {
		t.Errorf("decision_id: got %q, want %q", rec.DecisionID, req.Decision.ID)
	}
	if rec.SafetyClass != journal.SafetyClassExposureRaising {
		t.Errorf("safety_class: got %q", rec.SafetyClass)
	}
	if rec.Generation != 0 {
		t.Errorf("generation: got %d, want 0 — this change issues no other", rec.Generation)
	}
	if rec.AccountRef != "acct-7" {
		t.Errorf("account_ref: got %q", rec.AccountRef)
	}
	want := journal.DeriveClientOrderID(req.Decision.ID, 0)
	if rec.ClientOrderID != want {
		t.Errorf("client_order_id: got %q, want %q", rec.ClientOrderID, want)
	}
	if rec.SerializerVersion != execgw.WireSerializerVersion {
		t.Errorf("serializer_version: got %q, want %q", rec.SerializerVersion, execgw.WireSerializerVersion)
	}
	if !strings.Contains(rec.WireBody, `"clientOrderId":"`+want+`"`) {
		t.Errorf("wire body %q does not carry the idempotency key", rec.WireBody)
	}
}

// TestCancelRecordsNoKeyOrWireBody: the cancel endpoint takes no clientOrderId
// (openapi), so there is no key to claim and nothing a replay could resend.
func TestCancelRecordsNoKeyOrWireBody(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "O-9"}}
	gw, j, clk := newGateway(t, broker)

	intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	out, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   intent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2),
	})
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.ClientOrderID != "" || rec.WireBody != "" || rec.SerializerVersion != "" {
		t.Fatalf("a cancel carries no key and no replayable body: %+v", rec)
	}
	if rec.SafetyClass != journal.SafetyClassRiskReducing {
		t.Errorf("safety_class: got %q, want RISK_REDUCING", rec.SafetyClass)
	}
}

// TestStoredWireBodyIsWhatTheBrokerReceived is the drift guard between this
// package's renderer and internal/official's builder. The stored bytes are what
// a replay resends; if they ever stop being what the client actually sends, the
// replay sends a different body under the same idempotency key — which the
// broker answers with "422 idempotency-key-conflict" (openapi), a response that
// says nothing about the original order.
func TestStoredWireBodyIsWhatTheBrokerReceived(t *testing.T) {
	var captured string
	gw, j, clk, _ := officialGateway(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/orders" {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("reading the captured body: %v", err)
			}
			captured = string(body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"orderId":"O-1"}}`))
	})

	req := placeRequest(t, j, clk)
	out, err := gw.Place(context.Background(), req)
	if err != nil {
		t.Fatalf("Place: %v", err)
	}
	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatal(err)
	}
	if captured == "" {
		t.Fatal("no request body was captured")
	}
	if rec.WireBody != captured {
		t.Fatalf("the stored body is not the one that was sent:\nstored %s\n  sent %s",
			rec.WireBody, captured)
	}
	// And the key really went to the broker: the replay contract assumes it did.
	var sent map[string]any
	if err := json.Unmarshal([]byte(captured), &sent); err != nil {
		t.Fatal(err)
	}
	if sent["clientOrderId"] != journal.DeriveClientOrderID(req.Decision.ID, 0) {
		t.Fatalf("clientOrderId on the wire = %v, want the derived key", sent["clientOrderId"])
	}
}

// TestResubmittingASpentDecisionIsRefusedAcrossProcesses: the key is claimed in
// the journal, so a second gateway — a restarted process — refuses the reuse
// even though its in-memory nonce store has never seen it.
func TestResubmittingASpentDecisionIsRefusedAcrossProcesses(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)
	if _, err := gw.Place(context.Background(), req); err != nil {
		t.Fatalf("first Place: %v", err)
	}

	restarted, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	_, err = restarted.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianNonceReused {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianNonceReused, err)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls: got %d, want 1", places)
	}
}

// --- the durable one-shot nonce (task 1.7) -----------------------------------

// TestSpentDecisionIsRefusedAfterAJournalRestart is the spec's "재시작 후 결정
// 재사용 시도" end to end: the process is gone, the decision snapshot is still on
// disk and unexpired, and the gateway refuses it because the consumption record
// is in the journal rather than in a map that died with the process.
//
// The mutation is a cancel on purpose: it carries no idempotency key, so the
// refusal can only come from the nonce.
func TestSpentDecisionIsRefusedAfterAJournalRestart(t *testing.T) {
	clk := clock.NewFake(fixedNow)
	path := filepath.Join(t.TempDir(), "journal.db")
	open := func() *journal.Journal {
		t.Helper()
		j, err := journal.Open(context.Background(), journal.Options{
			Path: path, Clock: clk,
			FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
		})
		if err != nil {
			t.Fatalf("journal.Open: %v", err)
		}
		return j
	}
	gatewayOver := func(j *journal.Journal, broker trading.Broker) *execgw.Gateway {
		t.Helper()
		gw, err := execgw.New(execgw.Options{
			Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
			AccountRef: "acct-7", Source: "test",
		})
		if err != nil {
			t.Fatalf("execgw.New: %v", err)
		}
		return gw
	}

	broker := &fakeBroker{result: domain.MutationResult{Kind: "cancel", Status: "accepted", OrderID: "O-9"}}
	first := open()
	decision := exitDecision(t, first, clk, journal.KindCancel, "kr", "005930", "BUY", 2)
	intent := orderintent.CancelIntent{OrderID: "O-1", Symbol: "005930"}
	req := execgw.CancelRequest{
		Intent:   intent,
		Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
		Decision: decision,
	}
	if _, err := gatewayOver(first, broker).Cancel(context.Background(), req); err != nil {
		t.Fatalf("first Cancel: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	restarted := open()
	t.Cleanup(func() { _ = restarted.Close() })
	// A different symbol, so the refusal cannot be the symbol latch: the same
	// decision, re-presented, must be refused for having been spent.
	req.Intent = orderintent.CancelIntent{OrderID: "O-2", Symbol: "000660"}
	req.Order.Market = "kr"
	_, err := gatewayOver(restarted, broker).Cancel(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianNonceReused {
		t.Fatalf("want %q after a restart, got %v", execgw.ReasonGuardianNonceReused, err)
	}
	if _, cancels, _ := broker.totals(); cancels != 1 {
		t.Errorf("broker cancel calls: got %d, want 1", cancels)
	}
}

// TestPartialLimitSnapshotIsRefused: four bounds out of five is not
// four-fifths of an authorisation. The issuer API cannot produce one, so the row
// is edited behind its back — and the entry is still refused.
func TestPartialLimitSnapshotIsRefused(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	partial := testLimits()
	partial.MaxDailyLossRatio = execgw.Limit{}
	encoded, err := execgw.EncodeLimits(partial)
	if err != nil {
		t.Fatal(err)
	}
	tamperDecision(t, j, req.Decision.ID, "limits_json", encoded)

	_, err = gw.Place(context.Background(), req)
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianLimitsUnset {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianLimitsUnset, err)
	}
	if !strings.Contains(rejected.Detail, "daily loss ratio") {
		t.Errorf("the refusal must name the missing bound, got %q", rejected.Detail)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
}

// TestLimitsValidate is the fail-closed table: every field is required, every
// value has to be a positive finite number, the ratio is a ratio, and the money
// bounds need a currency to be money in.
func TestLimitsValidate(t *testing.T) {
	if err := testLimits().Validate(); err != nil {
		t.Fatalf("the complete snapshot must validate: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*execgw.Limits)
		want   string
	}{
		{"quantity unset", func(l *execgw.Limits) { l.MaxQuantity = execgw.Limit{} }, "order quantity"},
		{"notional unset", func(l *execgw.Limits) { l.MaxNotional = execgw.Limit{} }, "order notional"},
		{"total exposure unset", func(l *execgw.Limits) { l.MaxTotalExposure = execgw.Limit{} },
			"total open exposure"},
		{"daily loss amount unset", func(l *execgw.Limits) { l.MaxDailyLossAmount = execgw.Limit{} },
			"daily loss amount"},
		{"daily loss ratio unset", func(l *execgw.Limits) { l.MaxDailyLossRatio = execgw.Limit{} },
			"daily loss ratio"},
		{"zero", func(l *execgw.Limits) { l.MaxQuantity = execgw.Bound(0) }, "authorises nothing"},
		{"negative", func(l *execgw.Limits) { l.MaxNotional = execgw.Bound(-1) }, "authorises nothing"},
		{"NaN", func(l *execgw.Limits) { l.MaxTotalExposure = execgw.Bound(math.NaN()) }, "finite"},
		{"Inf", func(l *execgw.Limits) { l.MaxDailyLossAmount = execgw.Bound(math.Inf(1)) }, "finite"},
		{"ratio above one", func(l *execgw.Limits) { l.MaxDailyLossRatio = execgw.Bound(1.5) },
			"bounds nothing"},
		{"no currency", func(l *execgw.Limits) { l.Currency = " " }, "currency"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			limits := testLimits()
			tc.mutate(&limits)
			err := limits.Validate()
			if err == nil {
				t.Fatal("want a refusal")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

// TestLimitCurrencyMustMatchTheOrder: the money bounds are expressed in one
// currency, so measuring an order in another against them is a comparison that
// was never valid — not a conversion to perform.
func TestLimitCurrencyMustMatchTheOrder(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{OrderID: "O-1"}}
	gw, j, clk := newGateway(t, broker)

	limits := testLimits()
	limits.Currency = "USD"
	intent := placeIntent() // KRW
	_, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   intent,
		Decision: entryDecision(t, j, clk, intent, limits),
	})
	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianLimitExceeded {
		t.Fatalf("want %q, got %v", execgw.ReasonGuardianLimitExceeded, err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
}

// TestNonceIsSpentBeforeTheBrokerCall: the consumption is committed with the
// dispatch, so an attempt that reached the broker has a consumption record
// whatever the broker then answered.
func TestNonceIsSpentBeforeTheBrokerCall(t *testing.T) {
	broker := &fakeBroker{err: errors.New("connection reset after the request was written")}
	gw, j, clk := newGateway(t, broker)
	req := placeRequest(t, j, clk)

	if _, err := gw.Place(context.Background(), req); err == nil {
		t.Fatal("the fixture broker must fail")
	}
	dec, err := j.LookupDecision(context.Background(), req.Decision.ID)
	if err != nil {
		t.Fatal(err)
	}
	spent, err := j.NonceSpent(context.Background(), dec.Nonce)
	if err != nil {
		t.Fatal(err)
	}
	if !spent {
		t.Fatal("a decision whose mutation reached the broker must be spent")
	}
}

// TestPlaceWireBodyShapes pins the renderer itself against the openapi variants.
func TestPlaceWireBodyShapes(t *testing.T) {
	cases := []struct {
		name   string
		intent orderintent.PlaceIntent
		want   string
	}{
		{
			name: "kr limit buy",
			intent: orderintent.PlaceIntent{
				Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
				Quantity: 2, Price: 70000, CurrencyMode: "KRW", ClientOrderID: "tos-key",
			},
			want: `{"symbol":"005930","side":"BUY","orderType":"LIMIT","quantity":"2",` +
				`"price":"70000","timeInForce":"DAY","clientOrderId":"tos-key",` +
				`"confirmHighValueOrder":false}`,
		},
		{
			name: "us market sell",
			intent: orderintent.PlaceIntent{
				Symbol: "AAPL", Market: "us", Side: "sell", OrderType: "market",
				Quantity: 5, CurrencyMode: "USD", ClientOrderID: "tos-key",
			},
			want: `{"symbol":"AAPL","side":"SELL","orderType":"MARKET","quantity":"5",` +
				`"clientOrderId":"tos-key","confirmHighValueOrder":false}`,
		},
		{
			name: "fractional buy",
			intent: orderintent.PlaceIntent{
				Symbol: "TSLA", Market: "us", Side: "buy", OrderType: "market",
				Amount: 100.5, Fractional: true, CurrencyMode: "USD", ClientOrderID: "tos-key",
			},
			want: `{"symbol":"TSLA","side":"BUY","orderType":"MARKET","orderAmount":"100.5",` +
				`"clientOrderId":"tos-key","confirmHighValueOrder":false}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := execgw.PlaceWireBody(tc.intent)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("wire body:\n got %s\nwant %s", got, tc.want)
			}
		})
	}
}
