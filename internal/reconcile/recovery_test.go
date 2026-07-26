package reconcile_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Restart-recovery tests (harden-execution-base task 3.5).
//
// The requirement is reconciliation's "재시작 복구": journal resolution → account
// reads → state reconstruction, and no new order until that is done. The tests
// below run the real journal across a simulated crash, and drive a real
// execgw.Gateway to prove the refusal is a property of the wiring rather than of
// a comment.

// --- fakes for the resolver -------------------------------------------------

// resolverOrders is an execgw.OrderPager and OrderReader that reports an empty
// broker, which is what "the order never made it" looks like.
type resolverOrders struct {
	pages  map[string]execgw.OrderPage
	single map[string]json.RawMessage
}

func (o resolverOrders) OrdersPageRaw(_ context.Context, q execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	if o.pages == nil {
		return execgw.OrderPage{}, nil
	}
	return o.pages[q.Status+"|"+cursor], nil
}

func (o resolverOrders) OrderRaw(_ context.Context, id string) (json.RawMessage, error) {
	if raw, ok := o.single[id]; ok {
		return raw, nil
	}
	return nil, errors.New("no such order")
}

// resolverAccount corroborates absence: nothing moved.
type resolverAccount struct {
	buyingPower float64
	holding     float64
}

func (a resolverAccount) BuyingPower(context.Context, string) (float64, error) {
	return a.buyingPower, nil
}

func (a resolverAccount) HoldingQuantity(context.Context, string) (float64, error) {
	return a.holding, nil
}

// --- helpers ----------------------------------------------------------------

// openJournalAt opens a journal with the position projection bound (task 6.3).
//
// The FSProber is fixed because this repository lives on ntfs; the guard has its
// own tests in internal/journal. The apply hook is bound because since task 6.3
// the reconciliation's local belief is the projection, so a journal that records
// fills without projecting them believes it holds nothing.
func openJournalAt(t *testing.T, path string) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     path,
		Clock:    clock.NewFake(asOf),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition}); err != nil {
		t.Fatalf("SetApplyHooks: %v", err)
	}
	return j
}

func recoveryCollector(holdings []domain.Position, orders []json.RawMessage) *reconcile.Collector {
	rec := &recorder{}
	return &reconcile.Collector{
		Orders:     &fakeOrders{rec: rec, pages: map[string]execgw.OrderPage{"": {Orders: orders}}},
		Positions:  &fakeHoldings{rec: rec, positions: holdings},
		Balance:    &fakeBalance{rec: rec, cash: map[string]float64{"USD": 1000}},
		Currencies: []string{"USD"},
		AccountRef: "acct-7",
		Clock:      clock.System(),
	}
}

func newResolver(j *journal.Journal, gate *execgw.EntryGate) *execgw.Resolver {
	return &execgw.Resolver{
		Journal: j,
		Orders:  resolverOrders{},
		Order:   resolverOrders{},
		Account: resolverAccount{buyingPower: 1000, holding: 0},
		// A real clock with millisecond intervals: the resolution procedure's own
		// stabilisation is covered by internal/execgw's tests, and this one is
		// about the sequence around it.
		Clock: clock.System(),
		Gate:  gate,
		Config: execgw.ResolveConfig{
			StableObservations: 2,
			MinObservation:     time.Millisecond,
			PollInterval:       time.Millisecond,
			MaxDuration:        5 * time.Second,
		},
	}
}

func recoveryOptions(j *journal.Journal, gate *execgw.EntryGate, c *reconcile.Collector) reconcile.Options {
	return reconcile.Options{
		Journal:    j,
		Resolver:   newResolver(j, gate),
		Collector:  c,
		Gate:       gate,
		Clock:      clock.System(),
		AccountRef: "acct-7",
		Stabilise: reconcile.Stabilisation{
			Interval: time.Millisecond, Required: 2, MaxAttempts: 5,
		},
	}
}

// crashMidDispatch records an intent, marks it dispatched, then drops the journal
// handle without settling — which is exactly what a kill -9 leaves behind.
func crashMidDispatch(t *testing.T, path string) {
	t.Helper()
	ctx := context.Background()
	j := openJournalAt(t, path)
	attempt, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-crash", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10", Price: "200",
			Currency: "USD", Source: "engine", Fingerprint: "fp-crash",
			Notes: execgw.EncodeBaseline(&execgw.Baseline{
				BuyingPower: 1000, Holding: 0, Currency: "USD",
			}),
		},
		Kind: journal.KindPlace, AttemptID: "attempt-crash",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
}

// --- tests ------------------------------------------------------------------

// TestGateIsBlockedBeforeRecoveryEvenRuns: the dangerous state is the default.
// A Recovery that is constructed and never run leaves the engine unable to open
// a position.
func TestGateIsBlockedBeforeRecoveryEvenRuns(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("precondition: the gate should start open, got %v", rejected)
	}

	if _, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil))); err != nil {
		t.Fatal(err)
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("constructing a recovery must block new entries, got %v", rejected)
	}
}

// TestRecoveryReleasesTheLatchOnlyWhenItCompletes.
func TestRecoveryReleasesTheLatchOnlyWhenItCompletes(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	r, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil)))
	if err != nil {
		t.Fatal(err)
	}
	if r.Complete() {
		t.Fatal("a recovery reports complete before it has run")
	}

	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.Complete || !r.Complete() {
		t.Fatalf("report = %+v, want complete", report)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a clean recovery must release the latch, got %v", rejected)
	}
	if report.SnapshotsTaken < 2 {
		t.Fatalf("snapshots taken = %d, want the stabilisation to need at least 2",
			report.SnapshotsTaken)
	}
}

// TestCrashMidDispatchBecomesInDoubtAndIsResolved is the crash/restart case: the
// previous process died between "sent" and "answered", and the new one has to
// establish which before it trades.
func TestCrashMidDispatchBecomesInDoubtAndIsResolved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	crashMidDispatch(t, path)

	// A new process opens the same journal.
	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	r, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil)))
	if err != nil {
		t.Fatal(err)
	}

	report, err := r.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.MovedToInDoubt) != 1 || report.MovedToInDoubt[0] != "attempt-crash" {
		t.Fatalf("MovedToInDoubt = %v, want the mid-dispatch attempt", report.MovedToInDoubt)
	}
	if len(report.Resolutions) != 1 {
		t.Fatalf("resolutions = %+v, want 1", report.Resolutions)
	}
	// The broker shows nothing and the account never moved, so absence is proven.
	if report.Resolutions[0].State != journal.StateFailedConfirmed {
		t.Fatalf("resolution state = %s, want FAILED_CONFIRMED (proven absent)",
			report.Resolutions[0].State)
	}
	if len(report.Unresolved) != 0 {
		t.Fatalf("unresolved = %v, want none", report.Unresolved)
	}

	stored, err := j.LookupAttempt(context.Background(), "attempt-crash")
	if err != nil {
		t.Fatal(err)
	}
	if stored.State != journal.StateFailedConfirmed {
		t.Fatalf("the journal must record the resolution durably, got %s", stored.State)
	}
}

// TestRecordedButNeverSentIsClosedSafely: the payoff of writing the intent before
// dispatching. Nothing left the process, so nothing is ambiguous.
func TestRecordedButNeverSentIsClosedSafely(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()

	first := openJournalAt(t, path)
	if _, err := first.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{
			ID: "intent-1", Market: "us", TradingDay: "2026-03-30", AccountRef: "acct-7",
			Symbol: "AAPL", Side: "BUY", OrderType: "LIMIT", Quantity: "10", Price: "200",
			Currency: "USD", Source: "engine", Fingerprint: "fp-1",
		},
		Kind: journal.KindPlace, AttemptID: "attempt-1",
	}); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	j := openJournalAt(t, path)
	defer j.Close()
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	r, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil)))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.NotDispatched) != 1 {
		t.Fatalf("NotDispatched = %v, want the unsent attempt", report.NotDispatched)
	}
	if len(report.Resolutions) != 0 {
		t.Fatalf("an unsent attempt needs no resolution, got %+v", report.Resolutions)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("a provably unsent attempt must not keep the engine shut: %v", rejected)
	}
}

// TestRecoveryFailsClosedWhenTheAccountWillNotSettle: an account that keeps
// changing is not a reason to trade on the last reading.
func TestRecoveryFailsClosedWhenTheAccountWillNotSettle(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})

	rec := &recorder{}
	moving := &movingHoldings{rec: rec}
	c := &reconcile.Collector{
		Orders:     &fakeOrders{rec: rec, pages: map[string]execgw.OrderPage{"": {}}},
		Positions:  moving,
		Balance:    &fakeBalance{rec: rec, cash: map[string]float64{"USD": 1000}},
		Currencies: []string{"USD"},
		AccountRef: "acct-7",
		Clock:      clock.System(),
	}

	r, err := reconcile.New(recoveryOptions(j, gate, c))
	if err != nil {
		t.Fatal(err)
	}
	report, err := r.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("err = %v, want ErrRecoveryIncomplete", err)
	}
	if report.Complete || r.Complete() {
		t.Fatal("an unsettled account must not be reported as recovered")
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("the latch must stay shut, got %v", rejected)
	}
}

// movingHoldings never returns the same holdings twice.
type movingHoldings struct {
	rec  *recorder
	seen int
}

func (m *movingHoldings) Positions(context.Context) ([]domain.Position, error) {
	m.rec.log("holdings")
	m.seen++
	return []domain.Position{{Symbol: "AAPL", Quantity: float64(m.seen)}}, nil
}

// TestRecoveryThatEndsInDisagreementBlocksEntriesForADifferentReason: finishing
// the sequence is not the same as everything being fine.
func TestRecoveryThatEndsInDisagreementBlocksEntriesForADifferentReason(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.db")
	ctx := context.Background()
	j := openJournalAt(t, path)
	defer j.Close()

	// The engine believes it holds 10 AAPL.
	confirmedOrder(t, j, "intent-1", "attempt-1", "o-1", "AAPL", "BUY")
	if _, err := j.RecordFill(ctx, journal.FillObservation{
		OrderID: "o-1", Symbol: "AAPL", Market: "us", State: "FILLED", Terminal: true,
		Quantity: "10", FilledQuantity: "10", ObservedAt: "2026-03-30T01:30:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	// The account says 4.
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})
	c := recoveryCollector([]domain.Position{{Symbol: "AAPL", Quantity: 4}}, nil)
	r, err := reconcile.New(recoveryOptions(j, gate, c))
	if err != nil {
		t.Fatal(err)
	}

	report, err := r.Run(ctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !report.Complete {
		t.Fatal("the sequence itself completed and should say so")
	}
	if !report.Diff.BlocksEntry() {
		t.Fatalf("a 10 vs 4 disagreement must block: %s", report.Diff.Summary())
	}
	rejected := gate.CheckEntry()
	if rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("want a reconciliation block after recovery, got %v", rejected)
	}
	if got := report.Diff.Quantities[0].Authority(); got != "4" {
		t.Fatalf("authority = %q, want the account's 4", got)
	}
}

// --- gateway integration ----------------------------------------------------

// stubBroker records whether anything reached the broker.
type stubBroker struct {
	places  int
	cancels int
}

func (b *stubBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (domain.MutationResult, error) {
	b.places++
	return domain.MutationResult{Status: "success", OrderID: "broker-1"}, nil
}

func (b *stubBroker) CancelPendingOrder(context.Context, orderintent.CancelIntent) (domain.MutationResult, error) {
	b.cancels++
	return domain.MutationResult{Status: "success", OrderID: "broker-1"}, nil
}

func (b *stubBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (domain.MutationResult, error) {
	return domain.MutationResult{Status: "success", OrderID: "broker-2"}, nil
}

func (b *stubBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

// TestGatewayRefusesNewOrdersUntilRecoveryCompletes is the spec's "복구 완료 전
// 주문 시도" scenario, driven through the real gateway: the request is refused,
// the reason says why, and nothing reaches the broker.
func TestGatewayRefusesNewOrdersUntilRecoveryCompletes(t *testing.T) {
	j := openJournal(t)
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	broker := &stubBroker{}

	gw, err := execgw.New(execgw.Options{
		Journal: j,
		Trading: trading.NewService(config.Trading{
			Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		}, broker),
		Clock:      clk,
		AccountRef: "acct-7",
		Source:     "test",
		Entry:      gate,
	})
	if err != nil {
		t.Fatal(err)
	}

	r, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil)))
	if err != nil {
		t.Fatal(err)
	}

	intent := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 200, CurrencyMode: "USD",
	}
	issuer := &execgw.Issuer{Journal: j, Clock: clk, AccountRef: "acct-7", TTL: time.Minute}
	entry := func() execgw.GuardianDecision {
		t.Helper()
		d, err := issuer.IssueEntry(context.Background(), execgw.EntryRequest{
			Market: "us", Symbol: "AAPL", Side: "buy", Quantity: 1, EntryPrice: 200,
			StopPrice: 180, PolicyVersion: "test/v1",
			Limits: execgw.Limits{
				MaxQuantity:        execgw.Bound(10),
				MaxNotional:        execgw.Bound(100000),
				MaxTotalExposure:   execgw.Bound(500000),
				MaxDailyLossAmount: execgw.Bound(20000),
				MaxDailyLossRatio:  execgw.Bound(0.02),
				Currency:           "USD",
			},
		})
		if err != nil {
			t.Fatalf("issuing the entry decision: %v", err)
		}
		return d
	}

	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent: intent, Decision: entry(),
	})
	if err == nil {
		t.Fatal("a place before recovery completes must be refused")
	}
	if out.Reason != execgw.ReasonRecoveryIncomplete {
		t.Fatalf("reason = %q, want %q", out.Reason, execgw.ReasonRecoveryIncomplete)
	}
	if !strings.Contains(out.Detail, "recovery") {
		t.Fatalf("detail = %q, want it to say why", out.Detail)
	}
	if broker.places != 0 {
		t.Fatalf("broker places = %d, want 0 — nothing may reach the broker", broker.places)
	}
	if out.State != journal.StateNotDispatched {
		t.Fatalf("attempt state = %s, want NOT_DISPATCHED", out.State)
	}

	// After recovery, the same request is allowed through.
	if _, err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// A fresh decision: the first one is spent, and a spent decision authorises
	// nothing however the recovery went.
	if _, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent: intent, Decision: entry(),
	}); err != nil {
		t.Fatalf("after recovery the place should succeed: %v", err)
	}
	if broker.places != 1 {
		t.Fatalf("broker places = %d, want 1", broker.places)
	}
}

// TestExitsStayOpenWhileRecoveryIsIncomplete is §0.3: an operator who starts the
// engine into a mess must still be able to liquidate through it.
func TestExitsStayOpenWhileRecoveryIsIncomplete(t *testing.T) {
	j := openJournal(t)
	clk := clock.NewFake(asOf)
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	broker := &stubBroker{}

	gw, err := execgw.New(execgw.Options{
		Journal: j,
		Trading: trading.NewService(config.Trading{
			Place: true, Sell: true, Cancel: true, Amend: true, AllowLiveOrderActions: true,
		}, broker),
		Clock: clk, AccountRef: "acct-7", Source: "test", Entry: gate,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := reconcile.New(recoveryOptions(j, gate, recoveryCollector(nil, nil))); err != nil {
		t.Fatal(err)
	}
	if rejected := gate.CheckEntry(); rejected == nil {
		t.Fatal("precondition: entries should be blocked")
	}

	cancelIntent := orderintent.CancelIntent{OrderID: "broker-1", Symbol: "AAPL"}
	issuer := &execgw.Issuer{Journal: j, Clock: clk, AccountRef: "acct-7", TTL: time.Minute}
	decision, err := issuer.IssueReduction(context.Background(), execgw.ReductionRequest{
		Kind: journal.KindCancel, Market: "us", Symbol: "AAPL", Side: "BUY",
		MaxQuantity: 1, Reason: "exit during recovery",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gw.Cancel(context.Background(), execgw.CancelRequest{
		Intent:   cancelIntent,
		Order:    execgw.OrderRef{Market: "us", Side: "BUY", Quantity: 1, Price: 200, Currency: "USD"},
		Decision: decision,
	}); err != nil {
		t.Fatalf("a cancel must not be gated by an incomplete recovery: %v", err)
	}
	if broker.cancels != 1 {
		t.Fatalf("broker cancels = %d, want 1 — the exit stays open", broker.cancels)
	}
}

// TestRecoveryRefusesIncompleteWiring keeps a recovery that cannot block anything
// from being mistaken for one that can.
func TestRecoveryRefusesIncompleteWiring(t *testing.T) {
	j := openJournal(t)
	gate := execgw.NewEntryGate(clock.NewFake(asOf), map[execgw.RequiredQuery]time.Duration{})

	opts := recoveryOptions(j, gate, recoveryCollector(nil, nil))
	opts.Gate = nil
	if _, err := reconcile.New(opts); err == nil {
		t.Fatal("a recovery with no gate must be refused")
	}

	opts = recoveryOptions(j, gate, recoveryCollector(nil, nil))
	opts.Resolver = nil
	if _, err := reconcile.New(opts); err == nil {
		t.Fatal("a recovery with no resolver must be refused")
	}
}
