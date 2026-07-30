package engine_test

// exit_e2e_test.go is task 7.6: the whole exit path, end to end, over HTTP.
//
//	entry fill → projection → exit state → observation → ratchet raise
//	→ 40 % partial proposal → partial fill → taken_ratio moves in the apply tx
//	→ breach → full liquidation → CLOSED
//
// Everything except the broker is real: the journal with both apply hooks bound,
// the position projection, internal/exitpolicy, the Guardian's reduce-only
// issuance, internal/execgw's sealed submission sequence, and the official
// client talking to an httptest server. The broker is the httptest server, and
// the fills are recorded through journal.RecordFill — which is the entry point
// the fill detector calls, so the apply transaction under test is the production
// one.
//
// The automation gate stays off (§0.2, and no test may place a live order).
// What that costs is only the master switch: the observer is constructed
// directly here rather than through Context.ExitObserver, which refuses without
// a verified gate. Every check between an intent and the broker still runs.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// --- the broker ----------------------------------------------------------------

// e2eBroker is an httptest broker that answers the four calls this path makes:
// the handshake, the price read, the order create, and the read-back the gateway
// confirms a created order with.
type e2eBroker struct {
	mu sync.Mutex
	// last is the price /api/v1/prices answers with, per symbol.
	last map[string]string
	// orders records the create bodies, in order.
	orders []map[string]any
	// symbols records the order id → symbol mapping the read-back answers from.
	symbols map[string]string
	next    int
	// priceCalls counts the price reads, for the rate-budget assertion.
	priceCalls int
}

func newE2EBroker(t *testing.T) (*e2eBroker, *httptest.Server) {
	t.Helper()
	b := &e2eBroker{last: map[string]string{}, symbols: map[string]string{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/oauth2/token":
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case r.URL.Path == "/api/v1/accounts":
			_, _ = w.Write([]byte(
				`{"result":[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		case r.URL.Path == "/api/v1/prices":
			b.servePrices(w, r)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodPost:
			b.serveCreate(t, w, r)
		case r.URL.Path == "/api/v1/orders" && r.Method == http.MethodGet:
			// The open-order list. Empty is a real answer and it is the one this
			// broker gives: the tracer needs the *freshness* of the read, not its
			// contents.
			_, _ = w.Write([]byte(`{"result":{"orders":[],"nextCursor":"","hasNext":false}}`))
		case r.URL.Path == "/api/v1/buying-power":
			_, _ = w.Write([]byte(`{"result":{"cashBuyingPower":"100000000","currency":"KRW"}}`))
		case r.URL.Path == "/api/v1/holdings":
			_, _ = w.Write([]byte(`{"result":{"items":[]}}`))
		case strings.HasPrefix(r.URL.Path, "/api/v1/orders/") && r.Method == http.MethodGet:
			b.serveReadBack(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return b, srv
}

func (b *e2eBroker) servePrices(w http.ResponseWriter, r *http.Request) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.priceCalls++
	items := make([]map[string]string, 0, 2)
	for _, symbol := range strings.Split(r.URL.Query().Get("symbols"), ",") {
		symbol = strings.TrimSpace(symbol)
		price, ok := b.last[symbol]
		if !ok {
			continue
		}
		items = append(items, map[string]string{
			"symbol": symbol, "lastPrice": price, "currency": "KRW",
		})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"result": items})
}

func (b *e2eBroker) serveCreate(t *testing.T, w http.ResponseWriter, r *http.Request) {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b.mu.Lock()
	b.next++
	id := fmt.Sprintf("BO-%d", b.next)
	b.orders = append(b.orders, body)
	symbol, _ := body["symbol"].(string)
	b.symbols[id] = symbol
	key, _ := body["clientOrderId"].(string)
	b.mu.Unlock()

	_ = json.NewEncoder(w).Encode(map[string]any{
		"result": map[string]any{"orderId": id, "clientOrderId": key},
	})
}

func (b *e2eBroker) serveReadBack(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
	b.mu.Lock()
	symbol, ok := b.symbols[id]
	b.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"result": map[string]any{"orderId": id, "symbol": symbol, "status": "OPEN"},
	})
}

func (b *e2eBroker) quote(symbol, price string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.last[symbol] = price
}

func (b *e2eBroker) placed() []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]map[string]any(nil), b.orders...)
}

func (b *e2eBroker) reads() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.priceCalls
}

// --- the engine ------------------------------------------------------------------

type e2eStack struct {
	t        *testing.T
	broker   *e2eBroker
	engine   *engine.Context
	observer *engine.ExitObserver
	clk      *clock.Fake
	dir      string
	srv      *httptest.Server
}

var e2eNow = time.Date(2026, 3, 30, 1, 0, 0, 0, time.UTC)

func newE2EStack(t *testing.T) *e2eStack {
	t.Helper()
	broker, srv := newE2EBroker(t)
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	return openE2EStack(t, broker, srv, dir)
}

// newEntryCapableE2EStack is newE2EStack for the one suite that opens a position
// rather than closing one.
//
// The exit stacks do not need it: every mutation they make is a sell or a cancel,
// and interlock clause 6 does not touch those (execgw/protection.go). The tracer
// buys, and a buy on a build with no broker-resident protective execution is
// exactly what the clause refuses — so the tracer's slice has to say out loud
// that it is running as the protective-order change will, not as this build does.
func newEntryCapableE2EStack(t *testing.T) *e2eStack {
	t.Helper()
	broker, srv := newE2EBroker(t)
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	return openE2EStackProtected(t, broker, srv, dir)
}

// openE2EStack builds (or re-builds, for the restart test) the engine over one
// config directory and one journal file.
func openE2EStack(t *testing.T, broker *e2eBroker, srv *httptest.Server, dir string) *e2eStack {
	t.Helper()
	return buildE2EStack(t, broker, srv, dir, false)
}

// openE2EStackProtected is openE2EStack with interlock clause 6 satisfied, so
// the gateway it builds admits a buy. See newEntryCapableE2EStack.
func openE2EStackProtected(t *testing.T, broker *e2eBroker, srv *httptest.Server, dir string) *e2eStack {
	t.Helper()
	return buildE2EStack(t, broker, srv, dir, true)
}

func buildE2EStack(t *testing.T, broker *e2eBroker, srv *httptest.Server,
	dir string, protectionReady bool,
) *e2eStack {
	t.Helper()
	clk := clock.NewFake(e2eNow)
	opts := engine.Options{
		ConfigDir: dir,
		Clock:     clk,
		OfficialOptions: []official.Option{
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		},
	}
	opts.SetJournalProberForTest(journal.FixedFSProber(journal.FSInfo{
		Name: "ext4", Magic: journal.MagicExt,
	}))
	if protectionReady {
		opts.SetProtectionReadyForTest()
	}
	eng, err := engine.New(opts)
	if err != nil {
		t.Fatalf("engine.New: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })

	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal: eng.Journal, Clock: clk, AccountRef: eng.AccountRef,
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(),
		PolicyVersion: "add-core-domain/7.6",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	observer, err := engine.NewExitObserver(engine.ExitObserverOptions{
		Journal: eng.Journal, Prices: eng.Official, Retrier: eng.Retrier,
		Issuer: guardian, Submit: eng.Gateway, Alerts: eng.Notifier,
		Costs: costs.DefaultModel(), AccountRef: eng.AccountRef, Clock: clk,
	})
	if err != nil {
		t.Fatalf("NewExitObserver: %v", err)
	}
	return &e2eStack{t: t, broker: broker, engine: eng, observer: observer, clk: clk, dir: dir, srv: srv}
}

// entry puts a filled buy in the ledger: the decision, the intent, the
// acknowledged attempt and the fill the projection is built from.
func (s *e2eStack) entry(symbol, quantity, limit, stop string) journal.Position {
	s.t.Helper()
	ctx := context.Background()
	limits, err := execgw.EncodeLimits(execgw.Limits{
		MaxQuantity: execgw.Bound(100), MaxNotional: execgw.Bound(1_000_000),
		MaxTotalExposure: execgw.Bound(10_000_000), MaxDailyLossAmount: execgw.Bound(100_000),
		MaxDailyLossRatio: execgw.Bound(0.01), Currency: "KRW",
	})
	if err != nil {
		s.t.Fatalf("EncodeLimits: %v", err)
	}
	decisionID := "d-e2e-entry"
	if _, err := s.engine.Journal.RecordDecision(ctx, journal.DecisionRequest{
		ID: decisionID, AccountRef: s.engine.AccountRef,
		SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{
			AccountRef: s.engine.AccountRef, Market: "kr", Symbol: symbol, Side: "BUY",
			Quantity: quantity, EntryPrice: limit, StopPrice: stop, TargetPrice: "90000",
			PolicyVersion: "add-core-domain/7.6",
		},
		LimitsJSON: limits, Nonce: "nonce-" + decisionID,
		IssuedAt: s.clk.Now(), ExpiresAt: s.clk.Now().Add(time.Hour),
	}); err != nil {
		s.t.Fatalf("RecordDecision: %v", err)
	}

	attempt, err := s.engine.Journal.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "i-e2e-entry", Market: "kr", TradingDay: "2026-03-30",
			AccountRef: s.engine.AccountRef, Symbol: symbol, Side: "BUY",
			OrderType: "LIMIT", TimeInForce: "DAY", Quantity: quantity, Price: limit,
			Currency: "KRW", Source: "engine", Fingerprint: "fp-e2e-entry",
		},
		Kind: journal.KindPlace, AttemptID: "a-e2e-entry",
		AccountRef: s.engine.AccountRef, DecisionID: decisionID,
		SafetyClass:   journal.SafetyClassExposureRaising,
		ClientOrderID: journal.DeriveClientOrderID(decisionID, 0),
	})
	if err != nil {
		s.t.Fatalf("Prepare(entry): %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		s.t.Fatalf("MarkDispatchStarted: %v", err)
	}
	if err := attempt.MarkAcked(ctx, "BO-entry"); err != nil {
		s.t.Fatalf("MarkAcked: %v", err)
	}
	if err := attempt.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""); err != nil {
		s.t.Fatalf("Settle: %v", err)
	}
	s.fill("BO-entry", symbol, quantity, quantity, limit, true)

	p, err := s.engine.Journal.CurrentPosition(ctx, s.engine.AccountRef, "kr", symbol)
	if err != nil {
		s.t.Fatalf("CurrentPosition: %v", err)
	}
	return p
}

// fill records one cumulative fill observation — the entry point the fill
// detector calls, so the apply transaction is the production one.
func (s *e2eStack) fill(orderID, symbol, ordered, filled, avg string, terminal bool) {
	s.t.Helper()
	state := "OPEN_PARTIALLY_FILLED"
	if terminal {
		state = "CLOSED_FILLED"
	}
	if _, err := s.engine.Journal.RecordFill(context.Background(), journal.FillObservation{
		OrderID: orderID, Symbol: symbol, Market: "kr", State: state, Terminal: terminal,
		Quantity: ordered, FilledQuantity: filled, AveragePrice: avg,
		ObservedAt: s.clk.Now().Format(time.RFC3339),
	}); err != nil {
		s.t.Fatalf("RecordFill(%s): %v", orderID, err)
	}
}

func (s *e2eStack) observe() engine.ExitCycle {
	s.t.Helper()
	cycle := s.observer.ObserveOnce(context.Background())
	if cycle.Err != nil {
		s.t.Fatalf("observation cycle: %v", cycle.Err)
	}
	return cycle
}

func (s *e2eStack) state(positionID string) journal.ExitState {
	s.t.Helper()
	st, err := s.engine.Journal.ExitState(context.Background(), positionID)
	if err != nil {
		s.t.Fatalf("ExitState: %v", err)
	}
	return st
}

func (s *e2eStack) position(positionID string) journal.Position {
	s.t.Helper()
	p, err := s.engine.Journal.LookupPosition(context.Background(), positionID)
	if err != nil {
		s.t.Fatalf("LookupPosition: %v", err)
	}
	return p
}

// lastBrokerOrder is the broker order id the most recent create produced.
func (s *e2eStack) lastBrokerOrder() string {
	s.t.Helper()
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()
	return fmt.Sprintf("BO-%d", s.broker.next)
}

// --- the walk ---------------------------------------------------------------------

// TestTheWholeExitPathEndToEnd is the sequence task 7.6 names, in one test,
// because the sequence *is* the requirement: each step's output is the next
// step's input, and asserting them separately would let the joins rot.
func TestTheWholeExitPathEndToEnd(t *testing.T) {
	s := newE2EStack(t)
	ctx := context.Background()

	// 1. The entry fills. The projection has it; the exit policy does not know
	//    about it yet, because the loop has not run.
	p := s.entry("005930", "10", "70000", "68000")
	if p.Quantity != "10" || p.State != journal.PositionOpen {
		t.Fatalf("projection = %+v, want 10 held and OPEN", p)
	}
	if _, err := s.engine.Journal.ExitState(ctx, p.ID); err == nil {
		t.Fatal("nothing should have opened an exit state before the loop ran")
	}

	// 2. The first observation opens the state at the entry decision's stop and
	//    judges from there. D5's first correction: protected from t0.
	s.broker.quote("005930", "70100")
	if cycle := s.observe(); cycle.Opened != 1 {
		t.Fatalf("opened = %d, want the exit state", cycle.Opened)
	}
	if got := s.state(p.ID); got.Baseline != "68000" || got.InitialRisk != "2000" {
		t.Fatalf("state = %+v, want the entry stop as the baseline", got)
	}

	// 3. +0.5R raises the baseline to −0.5R and proposes nothing.
	s.broker.quote("005930", "71000")
	s.observe()
	raised := s.state(p.ID)
	if raised.RatchetLevel != journal.RatchetHalfRisk || raised.Baseline != "69000" {
		t.Fatalf("state = %+v, want HALF_RISK at 69000", raised)
	}
	if len(s.broker.placed()) != 0 {
		t.Fatal("a baseline raise asks for no order")
	}

	// 4. +1.0R proposes the 40 % partial, and it reaches the broker as a LIMIT
	//    sell of 4 under the intent id the proposal armed.
	s.broker.quote("005930", "72000")
	if cycle := s.observe(); cycle.Proposed != 1 {
		t.Fatalf("proposed = %d at +1.0R, want the partial", cycle.Proposed)
	}
	placed := s.broker.placed()
	if len(placed) != 1 {
		t.Fatalf("broker saw %d orders, want the partial", len(placed))
	}
	if placed[0]["side"] != "SELL" || fmt.Sprint(placed[0]["quantity"]) != "4" {
		t.Fatalf("broker body = %+v, want a sell of 4", placed[0])
	}
	if placed[0]["orderType"] != "LIMIT" {
		t.Fatalf("broker body = %+v; automated orders are LIMIT only", placed[0])
	}
	armed := s.state(p.ID)
	if armed.PendingAction != string(exitpolicy.ActionRatchetPartial) {
		t.Fatalf("pending = %+v, want the armed partial", armed)
	}
	partialOrder := s.lastBrokerOrder()

	// 5. The partial fills. The apply hook moves taken_ratio_total and resolves
	//    the proposal *in the fill's own transaction* — that is the whole point
	//    of D7's atomic apply point.
	s.fill(partialOrder, "005930", "4", "4", "72000", true)
	afterFill := s.state(p.ID)
	if afterFill.TakenRatioTotal != "0.4" {
		t.Fatalf("taken ratio = %s, want 0.4 of the initial quantity", afterFill.TakenRatioTotal)
	}
	if afterFill.Pending() {
		t.Fatalf("the fill answered the proposal; it must not still be armed: %+v", afterFill)
	}
	if held := s.position(p.ID); held.Quantity != "6" {
		t.Fatalf("projection = %s, want 6 left", held.Quantity)
	}

	// 6. The partial does not come round again: once taken, the ratchet's 40 % is
	//    once per position.
	s.broker.quote("005930", "72500")
	s.observe()
	if len(s.broker.placed()) != 1 {
		t.Fatalf("broker saw %d orders; the partial is once per position", len(s.broker.placed()))
	}

	// 7. The price falls through the baseline. The whole remainder is liquidated.
	s.broker.quote("005930", "68900")
	if cycle := s.observe(); cycle.Proposed != 1 {
		t.Fatalf("proposed = %d on the breach, want the liquidation", cycle.Proposed)
	}
	placed = s.broker.placed()
	if len(placed) != 2 {
		t.Fatalf("broker saw %d orders, want the partial and the liquidation", len(placed))
	}
	if fmt.Sprint(placed[1]["quantity"]) != "6" {
		t.Fatalf("liquidation body = %+v, want the whole remainder", placed[1])
	}
	liquidation := s.lastBrokerOrder()

	// 8. It fills. The position closes, the exit state completes, and the loop's
	//    working set empties.
	s.fill(liquidation, "005930", "6", "6", "68900", true)
	closed := s.position(p.ID)
	if closed.State != journal.PositionClosed || closed.Quantity != "0" {
		t.Fatalf("position = %+v, want CLOSED at zero", closed)
	}
	done := s.state(p.ID)
	if !done.Completed {
		t.Fatalf("exit state = %+v, want completed", done)
	}
	if done.TakenRatioTotal != "1" {
		t.Errorf("taken ratio = %s, want the whole position", done.TakenRatioTotal)
	}

	open, err := s.engine.Journal.OpenExitStates(ctx, s.engine.AccountRef)
	if err != nil {
		t.Fatalf("OpenExitStates: %v", err)
	}
	if len(open) != 0 {
		t.Errorf("working set = %+v, want empty once the position closed", open)
	}
	if cycle := s.observe(); cycle.Judged != 0 {
		t.Errorf("judged = %d after the close; a completed policy judges nothing further", cycle.Judged)
	}

	// 9. The history reads as the walk. It is append-only and it is the join the
	//    provenance query follows.
	events, err := s.engine.Journal.ExitEvents(ctx, p.ID)
	if err != nil {
		t.Fatalf("ExitEvents: %v", err)
	}
	var actions []string
	for _, e := range events {
		if e.Action != "" {
			actions = append(actions, e.Action)
		}
	}
	want := []string{
		journal.ExitEventOpened,
		string(exitpolicy.ActionRatchetPartial),
		journal.ExitEventProposalFilled,
		string(exitpolicy.ActionBaselineBreach),
		journal.ExitEventProposalFilled,
		journal.ExitEventCompleted,
	}
	if strings.Join(actions, ",") != strings.Join(want, ",") {
		t.Errorf("history = %v\nwant     %v", actions, want)
	}

	// The rate budget: one price read per cycle that had something to observe,
	// whatever the portfolio holds — five judged cycles above, and none for the
	// sixth, whose working set was empty because the position had closed.
	if s.broker.reads() != 5 {
		t.Errorf("price reads = %d, want one per cycle that had something to observe", s.broker.reads())
	}
}

// TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder is the crash test:
// the process dies with a proposal armed and an order live, and the restart
// finds both rather than proposing the level again.
func TestACrashBetweenArmingAndFillingResumesWithoutASecondOrder(t *testing.T) {
	broker, srv := newE2EBroker(t)
	dir := isolate(t)
	writeEngineConfig(t, dir)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")

	s := openE2EStack(t, broker, srv, dir)
	p := s.entry("005930", "10", "70000", "68000")
	broker.quote("005930", "72000")
	s.observe() // opens the state, reaches +1.0R and arms the partial

	armed := s.state(p.ID)
	if !armed.Pending() || armed.PendingIntentID == "" {
		t.Fatalf("state = %+v, want an armed proposal carrying its intent", armed)
	}
	partialOrder := s.lastBrokerOrder()
	if len(broker.placed()) != 1 {
		t.Fatalf("broker saw %d orders, want the partial", len(broker.placed()))
	}

	// The process dies. Everything in memory is gone; the file is not.
	if err := s.engine.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	restarted := openE2EStack(t, broker, srv, dir)

	// The restored state is the armed one, and the intent id survived — which is
	// what lets the fill that is already on its way resolve it.
	restored := restarted.state(p.ID)
	if restored.PendingIntentID != armed.PendingIntentID {
		t.Fatalf("restored intent = %q, want %q", restored.PendingIntentID, armed.PendingIntentID)
	}
	attempted, err := restarted.engine.Journal.IntentAttempted(context.Background(), restored.PendingIntentID)
	if err != nil {
		t.Fatalf("IntentAttempted: %v", err)
	}
	if !attempted {
		t.Fatal("the order was submitted before the crash, so its intent must show an attempt")
	}

	// A fresh observation at the same price proposes nothing: the outstanding
	// proposal suppresses it, which is what "미재발의" and "중복발의 없음" mean
	// together.
	broker.quote("005930", "72500")
	restarted.observe()
	if len(broker.placed()) != 1 {
		t.Fatalf("broker saw %d orders after the restart, want no second submission", len(broker.placed()))
	}

	// And the fill still lands on the proposal it answered.
	restarted.fill(partialOrder, "005930", "4", "4", "72000", true)
	settled := restarted.state(p.ID)
	if settled.Pending() {
		t.Fatalf("state = %+v, want the proposal resolved by its own fill", settled)
	}
	if settled.TakenRatioTotal != "0.4" {
		t.Errorf("taken ratio = %s, want 0.4", settled.TakenRatioTotal)
	}
}

// --- analytics retention (task 8.1) -------------------------------------------

// TestRetentionIsAsynchronousAndCannotTightenTheMode is the isolation
// requirement in one test: the sweep runs on its own goroutine, its failures do
// not propagate, and it raises no operating-mode transition.
func TestRetentionIsAsynchronousAndCannotTightenTheMode(t *testing.T) {
	s := newE2EStack(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan error, 1)
	go func() { done <- s.engine.RunAnalyticsRetention(ctx, s.clk, time.Hour) }()

	if !s.clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("the retention loop never reached its interval sleep")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Error("the loop returns its context's error, not a pruning verdict")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunAnalyticsRetention did not return after its context was cancelled")
	}

	mode, err := s.engine.Journal.CurrentOperatingMode(context.Background(), s.engine.AccountRef)
	if err != nil {
		t.Fatalf("CurrentOperatingMode: %v", err)
	}
	if mode.Mode != journal.ModeNormal {
		t.Errorf("mode = %s; an analytics job may never tighten the operating mode", mode.Mode)
	}
}

// TestAFailedSweepDoesNotPropagate: the journal is closed underneath it, which
// is the harshest failure the sweep can meet, and the loop keeps going.
//
// The handle is put back after Close so the sweep meets a *closed* journal
// rather than an absent one — an absent one is a no-op and would prove nothing
// about the error branch.
func TestAFailedSweepDoesNotPropagate(t *testing.T) {
	s := newE2EStack(t)
	handle := s.engine.Journal
	if err := s.engine.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	s.engine.Journal = handle
	if _, err := handle.PruneTradeOutcomes(context.Background(), e2eNow); err == nil {
		t.Fatal("the fixture must present a failing prune, or the loop is not being tested")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.engine.RunAnalyticsRetention(ctx, s.clk, time.Hour) }()
	if !s.clk.WaitForSleepers(1, 2*time.Second) {
		t.Fatal("a failed sweep must still reach the next interval rather than returning")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunAnalyticsRetention did not return")
	}
}
