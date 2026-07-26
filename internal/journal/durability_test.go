package journal

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

const testNow = "2026-03-30T00:30:00Z"

func testIntent() Intent {
	return Intent{
		ID:          "intent-1",
		Market:      "us",
		TradingDay:  "2026-03-30",
		AccountRef:  "acct-1",
		Symbol:      "AAPL",
		Side:        "BUY",
		OrderType:   "LIMIT",
		TimeInForce: "DAY",
		Quantity:    "10",
		Price:       "200.5",
		Currency:    "USD",
		Source:      "engine/test",
		Fingerprint: "fp-1",
	}
}

func testRequest() PrepareRequest {
	return PrepareRequest{
		Intent:    testIntent(),
		Kind:      KindPlace,
		AttemptID: "attempt-1",
	}
}

// TestPrepareCommitsIntentAndRecordedAttempt is the durability entry point: the
// immutable intent and its RECORDED attempt are written in one transaction, and
// Prepare returns only after that transaction committed.
func TestPrepareCommitsIntentAndRecordedAttempt(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if a.ID() != "attempt-1" || a.IntentID() != "intent-1" {
		t.Fatalf("handle identity = (%s, %s)", a.ID(), a.IntentID())
	}
	if a.State() != StateRecorded {
		t.Fatalf("state = %s, want RECORDED", a.State())
	}
	if a.AttemptNo() != 1 {
		t.Fatalf("attempt_no = %d, want 1", a.AttemptNo())
	}

	got, err := j.LookupIntent(ctx, "intent-1")
	if err != nil {
		t.Fatalf("LookupIntent: %v", err)
	}
	want := testIntent()
	if got != want {
		t.Fatalf("intent round-trip mismatch:\n got %+v\nwant %+v", got, want)
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.State != StateRecorded {
		t.Errorf("state = %s, want RECORDED", rec.State)
	}
	if rec.Kind != KindPlace {
		t.Errorf("kind = %s, want PLACE", rec.Kind)
	}
	if rec.RecordedAt != testNow {
		t.Errorf("recorded_at = %q, want the injected clock instant %q", rec.RecordedAt, testNow)
	}
	if rec.DispatchStartedAt != "" || rec.SettledAt != "" {
		t.Errorf("timestamps set too early: %+v", rec)
	}

	// The append-only history starts with the RECORDED entry.
	history := attemptHistory(t, j, "attempt-1")
	if len(history) != 1 || history[0] != "->RECORDED" {
		t.Fatalf("history = %v, want [->RECORDED]", history)
	}
}

// TestPrepareStoresEmptyPriceAsNull keeps market orders distinguishable from a
// zero-priced limit order.
func TestPrepareStoresEmptyPriceAsNull(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	req := testRequest()
	req.Intent.OrderType = "MARKET"
	req.Intent.Price = ""
	if _, err := j.Prepare(ctx, req); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	var price sql.NullString
	if err := j.db.QueryRowContext(ctx, "SELECT price FROM intents WHERE id = 'intent-1'").Scan(&price); err != nil {
		t.Fatal(err)
	}
	if price.Valid {
		t.Fatalf("price = %q, want NULL for a market order", price.String)
	}
	got, err := j.LookupIntent(ctx, "intent-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Price != "" {
		t.Fatalf("round-tripped price = %q, want empty", got.Price)
	}
}

// TestPrepareSecondAttemptOnSameIntent covers CANCEL/AMEND following a PLACE: the
// intent stays a single immutable row while attempts accumulate.
func TestPrepareSecondAttemptOnSameIntent(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.Prepare(ctx, testRequest()); err != nil {
		t.Fatalf("first Prepare: %v", err)
	}
	second := testRequest()
	second.AttemptID = "attempt-2"
	second.Kind = KindCancel
	second.TargetOrderID = "order-1"

	a2, err := j.Prepare(ctx, second)
	if err != nil {
		t.Fatalf("second Prepare: %v", err)
	}
	if a2.AttemptNo() != 2 {
		t.Fatalf("attempt_no = %d, want 2", a2.AttemptNo())
	}

	var intents, attempts int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM intents").Scan(&intents); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM mutation_attempts").Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if intents != 1 || attempts != 2 {
		t.Fatalf("rows = (%d intents, %d attempts), want (1, 2)", intents, attempts)
	}
}

// TestPrepareRejectsMutatedIntent enforces intent immutability, which the Phase 2
// ledger import depends on: the same id must never describe two different orders.
func TestPrepareRejectsMutatedIntent(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.Prepare(ctx, testRequest()); err != nil {
		t.Fatalf("first Prepare: %v", err)
	}

	mutated := testRequest()
	mutated.AttemptID = "attempt-2"
	mutated.Intent.Quantity = "999"

	a, err := j.Prepare(ctx, mutated)
	if !errors.Is(err, ErrIntentMutated) {
		t.Fatalf("want ErrIntentMutated, got (%v, %v)", a, err)
	}
	if a != nil {
		t.Fatal("no attempt handle may be returned when the write is refused")
	}
	// Nothing partial: the original quantity survives and no second attempt exists.
	got, err := j.LookupIntent(ctx, "intent-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Quantity != "10" {
		t.Fatalf("quantity = %q, want the original 10", got.Quantity)
	}
	if _, err := j.LookupAttempt(ctx, "attempt-2"); !errors.Is(err, ErrAttemptNotFound) {
		t.Fatalf("refused attempt must not exist, got %v", err)
	}
}

// TestPrepareValidation keeps a malformed intent out of the journal instead of
// discovering the missing field at the broker.
func TestPrepareValidation(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	cases := []struct {
		name   string
		mutate func(*PrepareRequest)
	}{
		{"empty intent id", func(r *PrepareRequest) { r.Intent.ID = "" }},
		{"empty attempt id", func(r *PrepareRequest) { r.AttemptID = "" }},
		{"empty symbol", func(r *PrepareRequest) { r.Intent.Symbol = "" }},
		{"empty side", func(r *PrepareRequest) { r.Intent.Side = "" }},
		{"empty quantity", func(r *PrepareRequest) { r.Intent.Quantity = "" }},
		{"empty currency", func(r *PrepareRequest) { r.Intent.Currency = "" }},
		{"empty market", func(r *PrepareRequest) { r.Intent.Market = "" }},
		{"empty trading day", func(r *PrepareRequest) { r.Intent.TradingDay = "" }},
		{"empty account ref", func(r *PrepareRequest) { r.Intent.AccountRef = "" }},
		{"empty source", func(r *PrepareRequest) { r.Intent.Source = "" }},
		{"empty fingerprint", func(r *PrepareRequest) { r.Intent.Fingerprint = "" }},
		{"unknown kind", func(r *PrepareRequest) { r.Kind = MutationKind("FROB") }},
		{"cancel without target order", func(r *PrepareRequest) { r.Kind = KindCancel }},
		{"amend without target order", func(r *PrepareRequest) { r.Kind = KindAmend }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := testRequest()
			tc.mutate(&req)
			a, err := j.Prepare(ctx, req)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("want ErrInvalidRequest, got (%v, %v)", a, err)
			}
			if a != nil {
				t.Fatal("no handle on a refused request")
			}
		})
	}

	var count int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM intents").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("refused requests wrote %d intent rows, want 0", count)
	}
}

// --- the decision binding (task 1.3) ----------------------------------------

// insertTestDecision writes the decisions row an attempt can point at, through
// the issuer API so the fixture is a decision the gateway would also accept.
func insertTestDecision(t *testing.T, j *Journal, id, accountRef, class string, generation int) {
	t.Helper()
	issued, err := time.Parse(time.RFC3339, testNow)
	if err != nil {
		t.Fatal(err)
	}
	req := DecisionRequest{
		ID: id, AccountRef: accountRef, Generation: generation, SafetyClass: class,
		Kind: KindPlace, Nonce: "nonce-" + id,
		IssuedAt: issued, ExpiresAt: issued.Add(time.Minute),
	}
	if class == SafetyClassExposureRaising {
		req.Preimage = RiskIntent{
			AccountRef: accountRef, Market: "us", Symbol: "AAPL", Side: "BUY",
			Quantity: "10", EntryPrice: "200.5", StopPrice: "190", PolicyVersion: "test/v1",
		}
		req.LimitsJSON = `{"max_quantity":"100"}`
	} else {
		req.Preimage = ReductionIntent{
			AccountRef: accountRef, Market: "us", Symbol: "AAPL", Side: "SELL",
			MaxQuantity: "10", Reason: "test",
		}
		// A cancel or a sell decision carries no key; the tests that need one ask
		// for a place decision.
		req.Kind = KindCancel
	}
	if _, err := j.RecordDecision(context.Background(), req); err != nil {
		t.Fatalf("RecordDecision(%s): %v", id, err)
	}
}

// TestPrepareRecordsTheDecisionBinding is the RECORDED half of the extended
// contract: which decision entitled this attempt, which key it may use and which
// bytes it may send are on disk *before* anything is dispatched (design D1/D2).
func TestPrepareRecordsTheDecisionBinding(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertTestDecision(t, j, "dec-1", "acct-1", SafetyClassExposureRaising, 0)

	req := testRequest()
	req.AccountRef = "acct-1"
	req.DecisionID = "dec-1"
	req.SafetyClass = SafetyClassExposureRaising
	req.Generation = 0
	req.ClientOrderID = DeriveClientOrderID("dec-1", 0)
	req.WireBody = `{"symbol":"AAPL","side":"BUY"}`
	req.SerializerVersion = "execgw/place/v1"

	if _, err := j.Prepare(ctx, req); err != nil {
		t.Fatalf("Prepare: %v", err)
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.DecisionID != "dec-1" {
		t.Errorf("decision_id = %q, want dec-1", rec.DecisionID)
	}
	if rec.SafetyClass != SafetyClassExposureRaising {
		t.Errorf("safety_class = %q", rec.SafetyClass)
	}
	if rec.Generation != 0 {
		t.Errorf("generation = %d, want 0", rec.Generation)
	}
	if rec.AccountRef != "acct-1" {
		t.Errorf("account_ref = %q, want acct-1", rec.AccountRef)
	}
	if rec.ClientOrderID != DeriveClientOrderID("dec-1", 0) {
		t.Errorf("client_order_id = %q", rec.ClientOrderID)
	}
	if rec.WireBody != req.WireBody {
		t.Errorf("wire_body = %q, want the bytes as supplied", rec.WireBody)
	}
	if rec.SerializerVersion != "execgw/place/v1" {
		t.Errorf("serializer_version = %q", rec.SerializerVersion)
	}
	if rec.ReplayCount != 0 || rec.LastReplayAt != "" {
		t.Errorf("replay bookkeeping starts empty, got (%d, %q)", rec.ReplayCount, rec.LastReplayAt)
	}
}

// TestPrepareRecordsAccountRefWithoutADecision keeps every attempt indexable by
// the partial UNIQUE on the idempotency key: the column is populated from the
// intent even when nothing supplies it (issues.md 0.1 → 1.3).
func TestPrepareRecordsAccountRefWithoutADecision(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.Prepare(ctx, testRequest()); err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.AccountRef != "acct-1" {
		t.Fatalf("account_ref = %q, want the intent's acct-1", rec.AccountRef)
	}
	if rec.DecisionID != "" || rec.SafetyClass != "" || rec.ClientOrderID != "" {
		t.Fatalf("an undecided attempt must carry no binding: %+v", rec)
	}
}

// TestPrepareRejectsAccountRefMismatch is the Manager's ruling on issues.md
// 0.1: the value must equal the intent's, and a disagreement is refused rather
// than reconciled. A decision issued for one account must not be able to
// authorise an attempt on another.
func TestPrepareRejectsAccountRefMismatch(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertTestDecision(t, j, "dec-1", "acct-OTHER", SafetyClassRiskReducing, 0)

	req := testRequest()
	req.AccountRef = "acct-OTHER"
	req.DecisionID = "dec-1"
	req.SafetyClass = SafetyClassRiskReducing

	a, err := j.Prepare(ctx, req)
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("want ErrInvalidRequest, got (%v, %v)", a, err)
	}
	if a != nil {
		t.Fatal("no handle may be returned for a refused binding")
	}
	var attempts int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM mutation_attempts").Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 0 {
		t.Fatalf("refused binding wrote %d attempt rows, want 0", attempts)
	}
}

// TestPrepareValidatesTheDecisionBinding refuses shapes that could only mean a
// caller is claiming something the record does not support.
func TestPrepareValidatesTheDecisionBinding(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertTestDecision(t, j, "dec-1", "acct-1", SafetyClassRiskReducing, 0)

	cases := []struct {
		name   string
		mutate func(*PrepareRequest)
	}{
		{"safety class without a decision", func(r *PrepareRequest) {
			r.SafetyClass = SafetyClassExposureRaising
		}},
		{"idempotency key without a decision", func(r *PrepareRequest) {
			r.ClientOrderID = DeriveClientOrderID("dec-1", 0)
		}},
		{"generation without a decision", func(r *PrepareRequest) { r.Generation = 1 }},
		{"unknown safety class", func(r *PrepareRequest) {
			r.DecisionID = "dec-1"
			r.SafetyClass = "PROBABLY_FINE"
		}},
		{"negative generation", func(r *PrepareRequest) {
			r.DecisionID = "dec-1"
			r.SafetyClass = SafetyClassRiskReducing
			r.Generation = -1
		}},
		{"malformed idempotency key", func(r *PrepareRequest) {
			r.DecisionID = "dec-1"
			r.SafetyClass = SafetyClassRiskReducing
			r.ClientOrderID = "not a valid key"
		}},
		{"idempotency key on a cancel", func(r *PrepareRequest) {
			r.Kind = KindCancel
			r.TargetOrderID = "order-1"
			r.DecisionID = "dec-1"
			r.SafetyClass = SafetyClassRiskReducing
			r.ClientOrderID = DeriveClientOrderID("dec-1", 0)
		}},
		{"wire body without a serializer version", func(r *PrepareRequest) {
			r.WireBody = `{"symbol":"AAPL"}`
		}},
		{"serializer version without a wire body", func(r *PrepareRequest) {
			r.SerializerVersion = "execgw/place/v1"
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := testRequest()
			tc.mutate(&req)
			a, err := j.Prepare(ctx, req)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("want ErrInvalidRequest, got (%v, %v)", a, err)
			}
			if a != nil {
				t.Fatal("no handle on a refused request")
			}
		})
	}
}

// TestPrepareRefusesAnUnpersistedDecision: the foreign key is the structural
// half of "발급자가 Gateway 호출 전에 결정을 영속한다" — an attempt cannot name a
// decision that was never written.
func TestPrepareRefusesAnUnpersistedDecision(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	req := testRequest()
	req.DecisionID = "never-written"
	req.SafetyClass = SafetyClassRiskReducing

	if _, err := j.Prepare(ctx, req); err == nil {
		t.Fatal("an attempt naming a decision that does not exist must be refused")
	}
	var attempts int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM mutation_attempts").Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 0 {
		t.Fatalf("attempt rows = %d, want 0", attempts)
	}
}

// TestIdempotencyKeyIsClaimedOnce: one key, one attempt per account. Two
// attempts under the same key would each believe they own the broker order that
// key names.
func TestIdempotencyKeyIsClaimedOnce(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertTestDecision(t, j, "dec-1", "acct-1", SafetyClassExposureRaising, 0)
	key := DeriveClientOrderID("dec-1", 0)

	req := testRequest()
	req.DecisionID = "dec-1"
	req.SafetyClass = SafetyClassExposureRaising
	req.ClientOrderID = key
	if _, err := j.Prepare(ctx, req); err != nil {
		t.Fatalf("first Prepare: %v", err)
	}

	second := req
	second.AttemptID = "attempt-2"
	if _, err := j.Prepare(ctx, second); err == nil {
		t.Fatal("a second attempt must not claim the same idempotency key")
	}
}

// TestDecisionBindingIsImmutableAcrossTransitions: the binding is written at
// RECORDED and no lifecycle transition may touch it. A replay reads these bytes
// long after the attempt moved on.
func TestDecisionBindingIsImmutableAcrossTransitions(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertTestDecision(t, j, "dec-1", "acct-1", SafetyClassExposureRaising, 0)

	req := testRequest()
	req.DecisionID = "dec-1"
	req.SafetyClass = SafetyClassExposureRaising
	req.ClientOrderID = DeriveClientOrderID("dec-1", 0)
	req.WireBody = `{"symbol":"AAPL","side":"BUY","quantity":"10"}`
	req.SerializerVersion = "execgw/place/v1"

	a, err := j.Prepare(ctx, req)
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	before, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.MarkAcked(ctx, "order-42"); err != nil {
		t.Fatal(err)
	}
	if err := a.Settle(ctx, StateConfirmed, "ok", "confirmed"); err != nil {
		t.Fatal(err)
	}
	after, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if after.DecisionID != before.DecisionID || after.SafetyClass != before.SafetyClass ||
		after.Generation != before.Generation || after.ClientOrderID != before.ClientOrderID ||
		after.WireBody != before.WireBody || after.SerializerVersion != before.SerializerVersion ||
		after.AccountRef != before.AccountRef {
		t.Fatalf("the decision binding changed across transitions:\nbefore %+v\n after %+v", before, after)
	}
}

// TestPrepareFailureBlocksSubmission is the ordering guarantee in the spec:
// "WHEN the journal transaction fails to commit THEN no broker submission
// happens". The handle returned by Prepare is the only capability that reaches
// the broker, so a failed write makes submission unspellable rather than merely
// discouraged. The disk-full condition is simulated with SQLite's page cap.
func TestPrepareFailureBlocksSubmission(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	submissions := 0
	submit := func(a *Attempt) { submissions++ }

	fillDatabase(t, j)

	req := testRequest()
	// A large note forces new page allocations, which the page cap refuses.
	req.Intent.Notes = strings.Repeat("x", 2<<20)

	a, err := j.Prepare(ctx, req)
	if err == nil {
		t.Fatal("Prepare must fail when the database cannot grow")
	}
	if a != nil {
		t.Fatal("no handle may be returned when the commit did not happen")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "full") {
		t.Logf("disk-full simulation surfaced as: %v", err)
	}
	if a != nil {
		submit(a)
	}
	if submissions != 0 {
		t.Fatalf("broker submissions = %d, want 0", submissions)
	}

	// Nothing half-written, and the journal is still usable once space is back.
	var count int
	if err := j.db.QueryRowContext(ctx, "SELECT count(*) FROM intents WHERE id = 'intent-1'").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("partial intent rows = %d, want 0", count)
	}
	if err := j.checkIntegrity(ctx); err != nil {
		t.Fatalf("integrity after a failed write: %v", err)
	}

	if _, err := j.db.ExecContext(ctx, "PRAGMA max_page_count = 1073741823"); err != nil {
		t.Fatal(err)
	}
	if _, err := j.Prepare(ctx, testRequest()); err != nil {
		t.Fatalf("Prepare after space was restored: %v", err)
	}
}

// fillDatabase caps the database at its current size, which makes any further page
// allocation return SQLITE_FULL — the closest a root-less test can get to a full
// disk.
func fillDatabase(t *testing.T, j *Journal) {
	t.Helper()
	ctx := context.Background()
	var pages int
	if err := j.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pages); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.ExecContext(ctx, "PRAGMA max_page_count = "+strconv.Itoa(pages)); err != nil {
		t.Fatal(err)
	}
}

// tempJournalPath returns a journal path inside a fresh temp directory.
func tempJournalPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "journal.db")
}

// TestWriteTransactionsUseBeginImmediate proves the write lock is taken at BEGIN,
// not discovered at COMMIT. That is what makes "commit succeeded" a reliable
// precondition for submitting: a second writer is refused before it can do any
// work, and the first writer's commit cannot fail on a late lock conflict.
func TestWriteTransactionsUseBeginImmediate(t *testing.T) {
	path := tempJournalPath(t)
	first := openTestJournalWithBusy(t, path, 100*time.Millisecond)
	second := openTestJournalWithBusy(t, path, 100*time.Millisecond)
	ctx := context.Background()

	txA, err := first.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("first BeginTx: %v", err)
	}

	// With BEGIN DEFERRED this would succeed and only fail later; with
	// BEGIN IMMEDIATE the write lock is already held.
	txB, err := second.db.BeginTx(ctx, nil)
	if err == nil {
		txB.Rollback()
		txA.Rollback()
		t.Fatal("a second write transaction must not start while the first holds the write lock")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "lock") &&
		!strings.Contains(strings.ToLower(err.Error()), "busy") {
		t.Fatalf("want a lock/busy error, got %v", err)
	}

	if err := txA.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	txB2, err := second.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("second writer after the lock was released: %v", err)
	}
	if err := txB2.Rollback(); err != nil {
		t.Fatal(err)
	}
}

// TestAttemptTransitionsAppendHistory covers the durable transition primitive: the
// attempt row is updated and the transition is appended, in one transaction.
func TestAttemptTransitionsAppendHistory(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if a.State() != StateDispatchStarted {
		t.Fatalf("state = %s", a.State())
	}
	if err := a.MarkAcked(ctx, "order-42"); err != nil {
		t.Fatalf("MarkAcked: %v", err)
	}
	if err := a.Settle(ctx, StateConfirmed, "ok", "confirmed by fill poll"); err != nil {
		t.Fatalf("Settle: %v", err)
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateConfirmed {
		t.Errorf("state = %s, want CONFIRMED", rec.State)
	}
	if rec.BrokerOrderID != "order-42" {
		t.Errorf("broker_order_id = %q, want order-42", rec.BrokerOrderID)
	}
	if rec.DispatchStartedAt != testNow || rec.SettledAt != testNow {
		t.Errorf("timestamps = (%q, %q), want the injected clock instant",
			rec.DispatchStartedAt, rec.SettledAt)
	}
	if rec.ReasonCode != "ok" {
		t.Errorf("reason_code = %q", rec.ReasonCode)
	}

	want := []string{"->RECORDED", "RECORDED->DISPATCH_STARTED", "DISPATCH_STARTED->ACKED", "ACKED->CONFIRMED"}
	got := attemptHistory(t, j, "attempt-1")
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("history = %v, want %v", got, want)
	}
}

// TestTransitionFromWrongStateIsRejected keeps the record honest: an ACK can only
// follow a dispatch we actually recorded.
func TestTransitionFromWrongStateIsRejected(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	// RECORDED → ACKED is outside the lifecycle table (nothing was dispatched),
	// so the stricter of the two guards reports it.
	if err := a.MarkAcked(ctx, "order-42"); !errors.Is(err, ErrIllegalTransition) {
		t.Fatalf("MarkAcked from RECORDED: want ErrIllegalTransition, got %v", err)
	}

	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateRecorded {
		t.Fatalf("state changed to %s despite the refusal", rec.State)
	}
	if got := attemptHistory(t, j, "attempt-1"); len(got) != 1 {
		t.Fatalf("refused transition appended history: %v", got)
	}
}

// TestSettleRequiresTerminalState prevents Settle from being used to sneak an
// attempt back into a live state.
func TestSettleRequiresTerminalState(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []AttemptState{StateRecorded, StateDispatchStarted, StateAcked, StateInDoubt} {
		if err := a.Settle(ctx, state, "r", "d"); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("Settle(%s): want ErrInvalidRequest, got %v", state, err)
		}
	}
	for _, state := range []AttemptState{
		StateConfirmed, StateNotDispatched, StateFailedConfirmed, StateUnresolvedInDoubt,
	} {
		if !state.IsTerminal() {
			t.Errorf("%s must report IsTerminal", state)
		}
	}
	if StateRecorded.IsTerminal() || StateInDoubt.IsTerminal() {
		t.Error("live states must not report IsTerminal")
	}
}

// TestMarkInDoubtFromDispatchStarted is the ambiguous-transport path: the attempt
// stays blocking until the resolution procedure finishes.
func TestMarkInDoubtFromDispatchStarted(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	a, err := j.Prepare(ctx, testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if err := a.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.MarkInDoubt(ctx, "transport_timeout", "read timeout after request was written"); err != nil {
		t.Fatalf("MarkInDoubt: %v", err)
	}
	rec, err := j.LookupAttempt(ctx, "attempt-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.State != StateInDoubt {
		t.Fatalf("state = %s, want IN_DOUBT", rec.State)
	}
	if rec.SettledAt != "" {
		t.Fatalf("IN_DOUBT is not terminal, settled_at must stay empty (got %q)", rec.SettledAt)
	}
	if rec.ReasonCode != "transport_timeout" {
		t.Fatalf("reason_code = %q", rec.ReasonCode)
	}
}

// TestLookupMissingRecords documents the not-found contract used by the restart
// recovery pass.
func TestLookupMissingRecords(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()

	if _, err := j.LookupIntent(ctx, "nope"); !errors.Is(err, ErrIntentNotFound) {
		t.Errorf("LookupIntent: want ErrIntentNotFound, got %v", err)
	}
	if _, err := j.LookupAttempt(ctx, "nope"); !errors.Is(err, ErrAttemptNotFound) {
		t.Errorf("LookupAttempt: want ErrAttemptNotFound, got %v", err)
	}
}

// TestOpenRefusesCorruptDatabase is the spec's corruption scenario: startup is
// refused and the message tells the operator what to do. Silently continuing on a
// damaged journal would mean deciding about live orders from a record we know is
// wrong.
func TestOpenRefusesCorruptDatabase(t *testing.T) {
	path := tempJournalPath(t)
	j := openTestJournalAt(t, path)
	ctx := context.Background()

	// Enough rows that the file has real content pages to damage.
	for i := 0; i < 50; i++ {
		req := testRequest()
		req.Intent.ID = "intent-" + strconv.Itoa(i)
		req.AttemptID = "attempt-" + strconv.Itoa(i)
		req.Intent.Notes = strings.Repeat("payload", 100)
		if _, err := j.Prepare(ctx, req); err != nil {
			t.Fatalf("Prepare %d: %v", i, err)
		}
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}

	corruptPages(t, path, 2, 8)

	reopened, err := Open(ctx, Options{
		Path:     path,
		FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
	})
	if reopened != nil {
		reopened.Close()
	}
	if !errors.Is(err, ErrCorrupt) {
		t.Fatalf("want ErrCorrupt, got %v", err)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("refusal must name the journal file: %v", err)
	}
	for _, hint := range []string{"reconcile", EnvDataDir} {
		if !strings.Contains(err.Error(), hint) {
			t.Errorf("refusal must include recovery guidance (%q missing): %v", hint, err)
		}
	}
}

// corruptPages overwrites whole pages with garbage, starting at 1-based page
// number from (page 1 holds the header; damaging it would fail differently).
func corruptPages(t *testing.T, path string, from, count int) {
	t.Helper()
	const pageSize = 4096
	f, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	garbage := make([]byte, pageSize)
	for i := range garbage {
		garbage[i] = 0xA5
	}
	wrote := 0
	for p := from; p < from+count; p++ {
		off := int64(p-1) * pageSize
		if off+pageSize > info.Size() {
			break
		}
		if _, err := f.WriteAt(garbage, off); err != nil {
			t.Fatal(err)
		}
		wrote++
	}
	if wrote == 0 {
		t.Fatalf("database file is only %d bytes; nothing to corrupt", info.Size())
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}

// attemptHistory returns the append-only transition log as "from->to" strings.
func attemptHistory(t *testing.T, j *Journal, attemptID string) []string {
	t.Helper()
	rows, err := j.db.QueryContext(context.Background(),
		"SELECT from_state, to_state FROM attempt_transitions WHERE attempt_id = ? ORDER BY id", attemptID)
	if err != nil {
		t.Fatal(err)
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
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func openTestJournalWithBusy(t *testing.T, path string, busy time.Duration) *Journal {
	t.Helper()
	j, err := Open(context.Background(), Options{
		Path:        path,
		Clock:       clock.NewFake(time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)),
		FSProber:    FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		BusyTimeout: busy,
	})
	if err != nil {
		t.Fatalf("Open(%s): %v", path, err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}
