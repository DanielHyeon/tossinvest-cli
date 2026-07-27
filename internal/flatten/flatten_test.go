package flatten_test

// flatten_test.go covers the cancel-all saga (task 4.4).
//
// The cases that carry the requirement are: every outstanding order is found
// (including page two), a crash mid-way resumes without re-cancelling what was
// already cancelled, an ambiguous cancel holds its symbol back rather than being
// assumed away, and a dry run mutates nothing.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/flatten"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

var flattenNow = time.Date(2026, 7, 26, 11, 0, 0, 0, time.UTC)

// --- fakes ------------------------------------------------------------------

// pagedOrders serves the open-order list in pages, so the "order on page two"
// case is exercised rather than assumed.
type pagedOrders struct {
	mu    sync.Mutex
	pages [][]json.RawMessage
	calls int
}

func (p *pagedOrders) OrdersPageRaw(_ context.Context, _ execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++

	index := 0
	if cursor != "" {
		fmt.Sscanf(cursor, "page-%d", &index)
	}
	if index >= len(p.pages) {
		return execgw.OrderPage{}, nil
	}
	page := execgw.OrderPage{Orders: p.pages[index]}
	if index+1 < len(p.pages) {
		page.NextCursor = fmt.Sprintf("page-%d", index+1)
		page.HasNext = true
	}
	return page, nil
}

// cancelBroker records every cancel and can be told to fail or hang on specific
// order ids.
type cancelBroker struct {
	mu       sync.Mutex
	cancels  []string
	places   int
	refuse   map[string]error
	ambiguou map[string]bool
}

func newCancelBroker() *cancelBroker {
	return &cancelBroker{refuse: map[string]error{}, ambiguou: map[string]bool{}}
}

func (b *cancelBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (domain.MutationResult, error) {
	b.mu.Lock()
	b.places++
	b.mu.Unlock()
	return domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "P-1"}, nil
}

func (b *cancelBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (b *cancelBroker) CancelPendingOrder(_ context.Context, intent orderintent.CancelIntent) (domain.MutationResult, error) {
	b.mu.Lock()
	b.cancels = append(b.cancels, intent.OrderID)
	err := b.refuse[intent.OrderID]
	ambiguous := b.ambiguou[intent.OrderID]
	b.mu.Unlock()

	if ambiguous {
		// 429 is the canonical ambiguous outcome: a rate limiter can answer after
		// the request reached the matching engine, so the cancel may or may not
		// have taken effect (retry-matrix §1, execgw.statusOf).
		return domain.MutationResult{}, official.ErrRateLimited
	}
	if err != nil {
		return domain.MutationResult{}, err
	}
	return domain.MutationResult{Kind: "cancel", Status: "accepted",
		OrderID: intent.OrderID + "-c", OriginalOrderID: intent.OrderID}, nil
}

func (b *cancelBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (domain.MutationResult, error) {
	return domain.MutationResult{}, errors.New("not used")
}

func (b *cancelBroker) cancelled() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]string(nil), b.cancels...)
}

// --- harness ----------------------------------------------------------------

type harness struct {
	journal *journal.Journal
	gateway *execgw.Gateway
	gate    *execgw.EntryGate
	orders  *pagedOrders
	broker  *cancelBroker
	clock   *clock.Fake
	path    string
}

func openJournalAt(t *testing.T, path string, clk clock.Clock) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		Clock:    clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	return j
}

func newHarness(t *testing.T, pages [][]json.RawMessage) *harness {
	t.Helper()
	clk := clock.NewFake(flattenNow)
	path := filepath.Join(t.TempDir(), "journal.db")
	j := openJournalAt(t, path, clk)
	t.Cleanup(func() { _ = j.Close() })

	broker := newCancelBroker()
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	gw, err := execgw.New(execgw.Options{
		Journal: j,
		Trading: trading.NewService(config.Trading{
			Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		}, broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "flatten-test",
		Entry:      gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return &harness{
		journal: j, gateway: gw, gate: gate,
		orders: &pagedOrders{pages: pages}, broker: broker, clock: clk, path: path,
	}
}

func (h *harness) saga(dryRun bool) *flatten.Saga {
	return &flatten.Saga{
		Journal:    h.journal,
		Gateway:    h.gateway,
		Gate:       h.gate,
		Orders:     h.orders,
		Clock:      h.clock,
		AccountRef: "acct-7",
		Operator:   "test-operator",
		Reason:     "test",
		DryRun:     dryRun,
	}
}

func order(id, symbol, side, quantity, price, currency string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(
		`{"orderId":%q,"symbol":%q,"side":%q,"status":"OPEN","quantity":%q,"price":%q,"currency":%q,`+
			`"execution":{"filledQuantity":"0"}}`,
		id, symbol, side, quantity, price, currency))
}

// --- tests ------------------------------------------------------------------

// TestCancelAllWalksEveryPage: an order on page two that we never saw is an order
// this flatten leaves live.
func TestCancelAllWalksEveryPage(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{
		{order("O-1", "005930", "BUY", "10", "70000", "KRW")},
		{order("O-2", "000660", "SELL", "5", "120000", "KRW")},
		{order("O-3", "AAPL", "BUY", "3", "180", "USD")},
	})

	report, err := h.saga(false).CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Found != 3 {
		t.Errorf("found = %d, want 3 across three pages", report.Found)
	}
	if report.Cancelled != 3 {
		t.Errorf("cancelled = %d, want 3 (%+v)", report.Cancelled, report.Outcomes)
	}
	if !report.Settled() {
		t.Errorf("report is not settled: %+v", report)
	}

	got := h.broker.cancelled()
	if len(got) != 3 {
		t.Fatalf("broker cancels = %v, want three", got)
	}
	for _, want := range []string{"O-1", "O-2", "O-3"} {
		if !containsString(got, want) {
			t.Errorf("%s was never cancelled; got %v", want, got)
		}
	}
}

// TestCancelAllBlocksEntriesFirst: the block is raised before anything slower
// happens, and it does not clear itself.
func TestCancelAllBlocksEntriesFirst(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{{order("O-1", "005930", "BUY", "10", "70000", "KRW")}})

	if rejected := h.gate.CheckEntry(); rejected != nil {
		t.Fatalf("precondition: the gate must start open, got %v", rejected)
	}
	if _, err := h.saga(false).CancelAll(context.Background()); err != nil {
		t.Fatalf("CancelAll: %v", err)
	}

	rejected := h.gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonFlattenInProgress {
		t.Fatalf("gate = %v, want a flatten_in_progress latch", rejected)
	}

	// And it stays: starting to trade again after an emergency exit is a separate
	// human decision.
	h.clock.Advance(time.Hour)
	if h.gate.CheckEntry() == nil {
		t.Error("the flatten latch must not clear itself")
	}
}

// TestDryRunSubmitsNothing is the spec's "--dry-run" scenario: the target list is
// produced and no mutation occurs.
func TestDryRunSubmitsNothing(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{
		{order("O-1", "005930", "BUY", "10", "70000", "KRW"),
			order("O-2", "AAPL", "SELL", "2", "180", "USD")},
	})
	ctx := context.Background()

	report, err := h.saga(true).CancelAll(ctx)
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Found != 2 || len(report.Outcomes) != 2 {
		t.Fatalf("report = %+v, want two targets listed", report)
	}
	if !report.DryRun {
		t.Error("the report must say it was a dry run")
	}
	if got := h.broker.cancelled(); len(got) != 0 {
		t.Fatalf("a dry run submitted %d cancel(s): %v", len(got), got)
	}

	// No attempt reached the journal either: "mutation 0건" is about the account,
	// and the journal is the record of what was attempted against it.
	pending, err := h.journal.PendingAttempts(ctx)
	if err != nil {
		t.Fatalf("PendingAttempts: %v", err)
	}
	if len(pending) != 0 {
		t.Errorf("a dry run left %d attempt(s) in the journal", len(pending))
	}

	// And the saga did not advance: the next real run must see the same work.
	saga, err := h.journal.ActiveFlatten(ctx)
	if err != nil {
		t.Fatalf("ActiveFlatten: %v", err)
	}
	if saga.Phase != journal.FlattenPhaseBlocking {
		t.Errorf("phase = %q, want the saga still at the start", saga.Phase)
	}
}

// resumeRun opens the journal at path and runs one saga against the given pages.
func resumeRun(t *testing.T, path string, clk *clock.Fake, pages [][]json.RawMessage,
	configure func(*cancelBroker), sagaID string,
) (flatten.CancelReport, *cancelBroker, *journal.Journal) {
	t.Helper()
	j := openJournalAt(t, path, clk)
	broker := newCancelBroker()
	if configure != nil {
		configure(broker)
	}
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	gw, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    trading.NewService(openPolicy(), broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Entry:      gate,
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	saga := &flatten.Saga{
		Journal: j, Gateway: gw, Gate: gate,
		Orders: &pagedOrders{pages: pages}, Clock: clk, AccountRef: "acct-7",
		NewID: func() string { return sagaID },
	}
	report, err := saga.CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	return report, broker, j
}

// TestResumeAfterCrashDoesNotRecancel is the crash-resume requirement.
//
// The first run cancels two orders and the process dies. The second run opens the
// same journal file, finds the unfinished saga, and must act only on what is left
// — re-cancelling a settled order would, with no idempotency key at the broker,
// be a cancel aimed at whatever order number has since been issued.
func TestResumeAfterCrashDoesNotRecancel(t *testing.T) {
	clk := clock.NewFake(flattenNow)
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	// --- run 1 --------------------------------------------------------------
	reportA, brokerA, first := resumeRun(t, path, clk, [][]json.RawMessage{{
		order("O-1", "005930", "BUY", "10", "70000", "KRW"),
		order("O-2", "000660", "SELL", "5", "120000", "KRW"),
	}}, nil, "saga-1")
	if reportA.Cancelled != 2 {
		t.Fatalf("first run = %+v, want both cancelled", reportA)
	}
	if len(brokerA.cancelled()) != 2 {
		t.Fatalf("first run broker calls = %v", brokerA.cancelled())
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// --- run 2: same file, new process. A third order has appeared since. ----
	reportB, brokerB, second := resumeRun(t, path, clk, [][]json.RawMessage{{
		order("O-3", "AAPL", "BUY", "1", "180", "USD"),
	}}, nil, "saga-should-not-be-used")
	defer second.Close()

	if !reportB.Resumed {
		t.Error("the second run must resume the first saga, not start a new one")
	}
	if reportB.SagaID != "saga-1" {
		t.Errorf("saga id = %q, want saga-1", reportB.SagaID)
	}

	cancelled := brokerB.cancelled()
	for _, settled := range []string{"O-1", "O-2"} {
		if containsString(cancelled, settled) {
			t.Errorf("the resumed run re-cancelled the settled order %s: %v", settled, cancelled)
		}
	}
	if !containsString(cancelled, "O-3") {
		t.Errorf("the resumed run did not cancel the outstanding order O-3: %v", cancelled)
	}

	// One saga, one step per order — the resume extended the plan rather than
	// duplicating it.
	steps, err := second.FlattenSteps(ctx, "saga-1")
	if err != nil {
		t.Fatalf("FlattenSteps: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("steps = %d, want 3 (one per order ever seen): %+v", len(steps), steps)
	}
}

// TestResumeNeverResubmitsAnAmbiguousCancel is the other half of resume safety.
//
// A cancel whose outcome is unknown may well have worked. Submitting a second one
// against the same order number is the single most dangerous thing this package
// could do — with no idempotency key, and with the broker issuing a new order
// number on every cancel, the second request can land on an order that has
// nothing to do with the first. So a resumed run looks again and never re-sends.
func TestResumeNeverResubmitsAnAmbiguousCancel(t *testing.T) {
	clk := clock.NewFake(flattenNow)
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	pages := [][]json.RawMessage{{order("O-2", "000660", "SELL", "5", "120000", "KRW")}}

	reportA, _, first := resumeRun(t, path, clk, pages,
		func(b *cancelBroker) { b.ambiguou["O-2"] = true }, "saga-1")
	if reportA.InDoubt != 1 {
		t.Fatalf("first run = %+v, want the cancel in doubt", reportA)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Run 2 sees the same order still open and a step recorded as IN_DOUBT.
	reportB, brokerB, second := resumeRun(t, path, clk, pages, nil, "unused")
	defer second.Close()

	if got := brokerB.cancelled(); len(got) != 0 {
		t.Fatalf("the resumed run re-submitted an ambiguous cancel: %v", got)
	}
	if reportB.Settled() {
		t.Error("an unresolved cancel must keep the saga unsettled")
	}

	held, err := second.UnsettledCancelSymbols(ctx, reportB.SagaID)
	if err != nil {
		t.Fatalf("UnsettledCancelSymbols: %v", err)
	}
	if _, blocked := held["000660"]; !blocked {
		t.Errorf("the symbol must stay held across the restart; held = %v", held)
	}
}

// TestUnresolvedCancelHoldsItsSymbol is engine-safety's oversell scenario: the
// symbol whose cancel is unresolved must be visible to the liquidation phase as
// held.
func TestUnresolvedCancelHoldsItsSymbol(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{{
		order("O-1", "005930", "BUY", "10", "70000", "KRW"),
		order("O-2", "AAPL", "SELL", "4", "180", "USD"),
	}})
	h.broker.ambiguou["O-2"] = true
	ctx := context.Background()

	report, err := h.saga(false).CancelAll(ctx)
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Settled() {
		t.Fatal("a saga with an unresolved cancel must not report itself settled")
	}
	if report.InDoubt != 1 {
		t.Errorf("in doubt = %d, want 1 (%+v)", report.InDoubt, report.Outcomes)
	}

	held, err := h.journal.UnsettledCancelSymbols(ctx, report.SagaID)
	if err != nil {
		t.Fatalf("UnsettledCancelSymbols: %v", err)
	}
	if _, blocked := held["AAPL"]; !blocked {
		t.Errorf("AAPL must be held out of liquidation; held = %v", held)
	}
	if _, blocked := held["005930"]; blocked {
		t.Errorf("005930 settled cleanly and must not be held; held = %v", held)
	}
}

// TestARefusedCancelIsNotAFailure: an order the broker will not cancel because it
// is already gone satisfies the saga's goal, and must not stop the remaining
// cancels (§0.3 — no early return with orders still live).
func TestARefusedCancelIsNotAFailure(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{{
		order("O-1", "005930", "BUY", "10", "70000", "KRW"),
		order("O-2", "000660", "SELL", "5", "120000", "KRW"),
		order("O-3", "AAPL", "BUY", "1", "180", "USD"),
	}})
	h.broker.refuse["O-2"] = definitiveRefusal()

	report, err := h.saga(false).CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Failed != 1 {
		t.Errorf("failed = %d, want 1 (%+v)", report.Failed, report.Outcomes)
	}
	if report.Cancelled != 2 {
		t.Errorf("cancelled = %d, want the other two (%+v)", report.Cancelled, report.Outcomes)
	}
	if !report.Settled() {
		t.Error("a definitively refused cancel is settled: the order is not live")
	}
	// The refusal must not have stopped the loop.
	if got := h.broker.cancelled(); len(got) != 3 {
		t.Errorf("broker saw %v; every order must have been tried", got)
	}
}

// TestNothingToCancelIsAValidFlatten.
func TestNothingToCancelIsAValidFlatten(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{{}})

	report, err := h.saga(false).CancelAll(context.Background())
	if err != nil {
		t.Fatalf("CancelAll: %v", err)
	}
	if report.Found != 0 || !report.Settled() {
		t.Errorf("report = %+v, want an empty settled run", report)
	}
	if report.Phase != journal.FlattenPhaseCancelled {
		t.Errorf("phase = %q, want CANCELLED", report.Phase)
	}
}

// TestUnreadableOrderStopsTheSaga: reporting "everything cancelled" while an
// order we could not parse is still live would be the worst possible outcome.
func TestUnreadableOrderStopsTheSaga(t *testing.T) {
	h := newHarness(t, [][]json.RawMessage{{
		json.RawMessage(`{"symbol":"AAPL","quantity":"1"}`), // no orderId
	}})

	if _, err := h.saga(false).CancelAll(context.Background()); err == nil {
		t.Fatal("an order that cannot be cancelled by id must stop the saga loudly")
	}
}

// TestSagaRefusesWithoutItsDependencies.
func TestSagaRefusesWithoutItsDependencies(t *testing.T) {
	h := newHarness(t, nil)
	cases := map[string]*flatten.Saga{
		"no journal": {Gateway: h.gateway, Orders: h.orders, AccountRef: "a"},
		"no orders":  {Journal: h.journal, Gateway: h.gateway, AccountRef: "a"},
		"no gateway": {Journal: h.journal, Orders: h.orders, AccountRef: "a"},
		"no account": {Journal: h.journal, Gateway: h.gateway, Orders: h.orders},
	}
	for name, saga := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := saga.CancelAll(context.Background()); !errors.Is(err, flatten.ErrNotConfigured) {
				t.Fatalf("err = %v, want ErrNotConfigured", err)
			}
		})
	}
}

// TestMarketInference: the payload has no market field, and getting it wrong
// journals the cancel against the wrong trading calendar.
func TestMarketInference(t *testing.T) {
	cases := []struct {
		symbol, currency, want string
	}{
		{"005930", "KRW", "kr"},
		{"AAPL", "USD", "us"},
		{"005930", "", "kr"},  // six digits
		{"AAPL", "", "us"},    // not six digits
		{"AAPL", "KRW", "kr"}, // currency wins: a KRW order is a KR order
		{"123456", "USD", "us"},
	}
	for _, tc := range cases {
		if got := flatten.MarketOf(tc.symbol, tc.currency); got != tc.want {
			t.Errorf("MarketOf(%q, %q) = %q, want %q", tc.symbol, tc.currency, got, tc.want)
		}
	}
}

// --- helpers ----------------------------------------------------------------

func openPolicy() config.Trading {
	return config.Trading{Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true}
}

// definitiveRefusal is a 400: the broker understood the cancel and refused it,
// which proves the mutation did not execute.
func definitiveRefusal() error {
	return &official.APIError{Code: http.StatusBadRequest, Body: `{"message":"the order is not cancellable"}`}
}

func containsString(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
