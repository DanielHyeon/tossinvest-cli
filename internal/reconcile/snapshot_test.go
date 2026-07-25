package reconcile_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// Snapshot-contract tests (harden-execution-base task 3.4).
//
// What is being pinned is the reconciliation spec's contract: the fixed read
// order, the as-of stamp, "부분 실패한 스냅샷은 폐기한다", and the stabilisation
// rule of two agreeing snapshots at least MinInterval apart.

var asOf = time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC)

// recorder logs every read in the order it happened, which is the only way to
// test an ordering requirement.
type recorder struct {
	mu    sync.Mutex
	calls []string
}

func (r *recorder) log(what string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, what)
}

func (r *recorder) order() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.calls...)
}

type fakeOrders struct {
	rec    *recorder
	pages  map[string]execgw.OrderPage
	err    error
	visits int
}

func (f *fakeOrders) OrdersPageRaw(_ context.Context, _ execgw.OrderQuery, cursor string) (execgw.OrderPage, error) {
	f.rec.log("orders")
	f.visits++
	if f.err != nil {
		return execgw.OrderPage{}, f.err
	}
	return f.pages[cursor], nil
}

type fakeHoldings struct {
	rec       *recorder
	positions []domain.Position
	err       error
}

func (f *fakeHoldings) Positions(context.Context) ([]domain.Position, error) {
	f.rec.log("holdings")
	if f.err != nil {
		return nil, f.err
	}
	return f.positions, nil
}

type fakeBalance struct {
	rec  *recorder
	cash map[string]float64
	err  error
}

func (f *fakeBalance) BuyingPower(_ context.Context, currency string) (float64, error) {
	f.rec.log("balance/" + currency)
	if f.err != nil {
		return 0, f.err
	}
	return f.cash[currency], nil
}

func orderJSON(id, symbol, side, qty, filled, price string) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(
		`{"orderId":%q,"symbol":%q,"side":%q,"status":"OPEN","quantity":%q,"price":%q,`+
			`"execution":{"filledQuantity":%q}}`, id, symbol, side, qty, price, filled))
}

func newCollector(t *testing.T) (*reconcile.Collector, *recorder, *fakeOrders, *fakeHoldings, *fakeBalance) {
	t.Helper()
	rec := &recorder{}
	orders := &fakeOrders{rec: rec, pages: map[string]execgw.OrderPage{"": {}}}
	holdings := &fakeHoldings{rec: rec}
	balance := &fakeBalance{rec: rec, cash: map[string]float64{"USD": 1000, "KRW": 500000}}
	c := &reconcile.Collector{
		Orders:     orders,
		Positions:  holdings,
		Balance:    balance,
		Currencies: []string{"USD"},
		AccountRef: "acct-7",
		Clock:      clock.NewFake(asOf),
	}
	return c, rec, orders, holdings, balance
}

// TestSnapshotReadsInTheContractedOrder pins (open orders → holdings → balances).
// The order decides which way a mid-snapshot fill errs, and the contract picks
// the direction that raises a false alarm rather than a false all-clear.
func TestSnapshotReadsInTheContractedOrder(t *testing.T) {
	c, rec, _, _, _ := newCollector(t)
	if _, err := c.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	got := strings.Join(rec.order(), ",")
	if got != "orders,holdings,balance/USD" {
		t.Fatalf("read order = %q, want orders,holdings,balance/USD", got)
	}
}

// TestSnapshotWalksTheOpenListToTheLastPage: a partial list would make every
// order past page one look like somebody else's.
func TestSnapshotWalksTheOpenListToTheLastPage(t *testing.T) {
	c, _, orders, _, _ := newCollector(t)
	orders.pages = map[string]execgw.OrderPage{
		"": {
			Orders:     []json.RawMessage{orderJSON("o-1", "AAPL", "BUY", "10", "0", "200")},
			NextCursor: "p2", HasNext: true,
		},
		"p2": {Orders: []json.RawMessage{orderJSON("o-2", "MSFT", "BUY", "5", "0", "400")}},
	}

	snap, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(snap.OpenOrders) != 2 {
		t.Fatalf("open orders = %d, want both pages", len(snap.OpenOrders))
	}
	if orders.visits != 2 {
		t.Fatalf("page fetches = %d, want 2", orders.visits)
	}
}

// TestAsOfIsTheOldestPartOfTheSnapshot: a freshness claim has to be true of the
// whole reading, so it is stamped from the first read, not the last.
func TestAsOfIsTheOldestPartOfTheSnapshot(t *testing.T) {
	rec := &recorder{}
	clk := clock.NewFake(asOf)
	orders := &fakeOrders{rec: rec, pages: map[string]execgw.OrderPage{"": {}}}
	holdings := &fakeHoldings{rec: rec}
	balance := &fakeBalance{rec: rec, cash: map[string]float64{"USD": 10}}
	c := &reconcile.Collector{
		Orders: orders, Positions: holdings, Balance: balance,
		Currencies: []string{"USD"}, AccountRef: "acct-7", Clock: clk,
	}

	snap, err := c.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !snap.AsOf.Equal(asOf) {
		t.Fatalf("AsOf = %s, want the instant the first read started (%s)", snap.AsOf, asOf)
	}
	if snap.Age(asOf.Add(30*time.Second)) != 30*time.Second {
		t.Fatalf("Age = %s, want 30s", snap.Age(asOf.Add(30*time.Second)))
	}
}

// TestPartialSnapshotIsDiscardedWhole is the "부분 실패한 스냅샷은 폐기한다"
// requirement expressed as a type: there is no value this function can return
// that represents most of an account.
func TestPartialSnapshotIsDiscardedWhole(t *testing.T) {
	boom := errors.New("broker said no")

	t.Run("orders fail", func(t *testing.T) {
		c, _, orders, _, _ := newCollector(t)
		orders.err = boom
		snap, err := c.Collect(context.Background())
		if !errors.Is(err, reconcile.ErrPartialSnapshot) {
			t.Fatalf("err = %v, want ErrPartialSnapshot", err)
		}
		if !snap.Empty() {
			t.Fatalf("a failed collect must return no snapshot, got %+v", snap)
		}
	})

	t.Run("holdings fail after orders succeeded", func(t *testing.T) {
		c, _, orders, holdings, _ := newCollector(t)
		orders.pages = map[string]execgw.OrderPage{
			"": {Orders: []json.RawMessage{orderJSON("o-1", "AAPL", "BUY", "10", "0", "200")}},
		}
		holdings.err = boom
		snap, err := c.Collect(context.Background())
		if !errors.Is(err, reconcile.ErrPartialSnapshot) {
			t.Fatalf("err = %v, want ErrPartialSnapshot", err)
		}
		if !snap.Empty() || len(snap.OpenOrders) != 0 {
			t.Fatalf("the orders that did arrive must be discarded too, got %+v", snap)
		}
	})

	t.Run("balance fails last", func(t *testing.T) {
		c, _, _, holdings, balance := newCollector(t)
		holdings.positions = []domain.Position{{Symbol: "AAPL", Quantity: 10}}
		balance.err = boom
		snap, err := c.Collect(context.Background())
		if !errors.Is(err, reconcile.ErrPartialSnapshot) {
			t.Fatalf("err = %v, want ErrPartialSnapshot", err)
		}
		if !snap.Empty() {
			t.Fatalf("a snapshot missing only its cash is still discarded, got %+v", snap)
		}
	})
}

// TestUnreadableOrderDiscardsTheSnapshot: a record we cannot parse is an unknown,
// and an unknown must not be silently absent from a comparison that decides
// whether the engine knows its own exposure.
func TestUnreadableOrderDiscardsTheSnapshot(t *testing.T) {
	c, _, orders, _, _ := newCollector(t)
	orders.pages = map[string]execgw.OrderPage{
		"": {Orders: []json.RawMessage{json.RawMessage(`{"symbol":"AAPL"}`)}},
	}
	if _, err := c.Collect(context.Background()); !errors.Is(err, reconcile.ErrPartialSnapshot) {
		t.Fatalf("an order with no id must discard the snapshot, got %v", err)
	}
}

// TestCollectorRefusesIncompleteWiring: a snapshot without a cash reading cannot
// answer whether an order was affordable, so it is not a snapshot.
func TestCollectorRefusesIncompleteWiring(t *testing.T) {
	c, _, _, _, _ := newCollector(t)
	c.Currencies = nil
	if _, err := c.Collect(context.Background()); err == nil {
		t.Fatal("a collector with no currencies must refuse")
	}
}

// TestDigestIgnoresWhenWeLooked: stabilisation asks whether the account stopped
// moving, and two readings a second apart always differ in their clocks.
func TestDigestIgnoresWhenWeLooked(t *testing.T) {
	base := reconcile.Snapshot{
		AsOf: asOf, CompletedAt: asOf.Add(time.Second), AccountRef: "acct-7",
		Holdings: []reconcile.Holding{{Symbol: "AAPL", Quantity: "10", AveragePrice: "200"}},
	}
	later := base
	later.AsOf = asOf.Add(time.Minute)
	later.CompletedAt = asOf.Add(time.Minute + time.Second)

	if base.Digest() != later.Digest() {
		t.Fatalf("the digest must not depend on when we looked:\n%q\n%q",
			base.Digest(), later.Digest())
	}
}

// TestDigestIsOrderIndependent keeps a broker that returns the same holdings in a
// different order from looking like a change.
func TestDigestIsOrderIndependent(t *testing.T) {
	a := reconcile.Snapshot{AsOf: asOf, Holdings: []reconcile.Holding{
		{Symbol: "AAPL", Quantity: "10"}, {Symbol: "MSFT", Quantity: "5"},
	}}
	b := reconcile.Snapshot{AsOf: asOf, Holdings: []reconcile.Holding{
		{Symbol: "MSFT", Quantity: "5"}, {Symbol: "AAPL", Quantity: "10"},
	}}
	if a.Digest() != b.Digest() {
		t.Fatalf("holdings order must not change the digest:\n%q\n%q", a.Digest(), b.Digest())
	}
}

// TestStabilisationNeedsTwoAgreeingSnapshotsSpacedApart is the spec's rule:
// "최소 간격(기본 2초)을 둔 연속 2회 동일 스냅샷".
func TestStabilisationNeedsTwoAgreeingSnapshotsSpacedApart(t *testing.T) {
	s := &reconcile.Stabiliser{}
	snapAt := func(at time.Time, qty string) reconcile.Snapshot {
		return reconcile.Snapshot{
			AsOf: at, CompletedAt: at.Add(100 * time.Millisecond), AccountRef: "acct-7",
			Holdings: []reconcile.Holding{{Symbol: "AAPL", Quantity: qty}},
		}
	}

	first := s.Offer(snapAt(asOf, "10"))
	if !first.Counted || first.Stable {
		t.Fatalf("first snapshot = %+v, want counted but not yet stable", first)
	}

	// Too soon: not evidence, and not allowed to break anything either.
	tooSoon := s.Offer(snapAt(asOf.Add(500*time.Millisecond), "10"))
	if tooSoon.Counted {
		t.Fatalf("a snapshot inside the interval must not count: %+v", tooSoon)
	}
	if tooSoon.Stable {
		t.Fatal("an uncounted snapshot must not declare stability")
	}
	if !strings.Contains(tooSoon.Why, "stabilisation interval") {
		t.Fatalf("Why = %q, want an explanation naming the interval", tooSoon.Why)
	}

	// Far enough apart and identical: stable.
	stable := s.Offer(snapAt(asOf.Add(2*time.Second), "10"))
	if !stable.Counted || !stable.Stable || stable.Streak != 2 {
		t.Fatalf("second spaced snapshot = %+v, want stable with a streak of 2", stable)
	}
}

// TestAChangedSnapshotResetsTheStreak: an account that is still moving has not
// stabilised, however many times we look.
func TestAChangedSnapshotResetsTheStreak(t *testing.T) {
	s := &reconcile.Stabiliser{}
	snapAt := func(at time.Time, qty string) reconcile.Snapshot {
		return reconcile.Snapshot{
			AsOf: at, CompletedAt: at, AccountRef: "acct-7",
			Holdings: []reconcile.Holding{{Symbol: "AAPL", Quantity: qty}},
		}
	}

	s.Offer(snapAt(asOf, "10"))
	changed := s.Offer(snapAt(asOf.Add(3*time.Second), "12"))
	if changed.Stable || changed.Streak != 1 {
		t.Fatalf("a changed snapshot = %+v, want the streak reset to 1", changed)
	}
	again := s.Offer(snapAt(asOf.Add(6*time.Second), "12"))
	if !again.Stable {
		t.Fatalf("two agreeing snapshots after the change = %+v, want stable", again)
	}
}

// TestDiscardedSnapshotsAreNotEvidence: a failed collect must not be able to
// corroborate anything.
func TestDiscardedSnapshotsAreNotEvidence(t *testing.T) {
	s := &reconcile.Stabiliser{}
	res := s.Offer(reconcile.Snapshot{})
	if res.Counted || res.Stable {
		t.Fatalf("the zero snapshot = %+v, want ignored", res)
	}
}

// TestResetClearsCorroboration stops a reconciliation inheriting the evidence it
// already acted on.
func TestResetClearsCorroboration(t *testing.T) {
	s := &reconcile.Stabiliser{}
	snap := reconcile.Snapshot{AsOf: asOf, CompletedAt: asOf, AccountRef: "acct-7"}
	s.Offer(snap)
	later := snap
	later.AsOf = asOf.Add(3 * time.Second)
	if !s.Offer(later).Stable {
		t.Fatal("expected stability before the reset")
	}

	s.Reset()
	if _, ok := s.Last(); ok {
		t.Fatal("Reset must forget the last snapshot")
	}
	third := snap
	third.AsOf = asOf.Add(6 * time.Second)
	if s.Offer(third).Stable {
		t.Fatal("after a reset the first snapshot must not be stable on its own")
	}
}
