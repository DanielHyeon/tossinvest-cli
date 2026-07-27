package riskcalc_test

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// confirmed_floor_test.go pins the formula (task 4.2, design D6):
//
//	확정 하한 = max(0, min(신선 보유, 신선 매도가능) − 로컬 미체결 SELL)
//
// Three of these tests are the conservative rules rather than the arithmetic:
// a local quantity may never raise the bound, an absent or stale snapshot is
// zero, and the sellable quantity only lowers.

var floorNow = time.Date(2026, 3, 30, 0, 30, 0, 0, time.UTC)

func snap(quantity string, age time.Duration) *riskcalc.QuantitySnapshot {
	return &riskcalc.QuantitySnapshot{Quantity: quantity, AsOf: floorNow.Add(-age)}
}

func floor(t *testing.T, in riskcalc.ConfirmedFloorInputs) riskcalc.ConfirmedFloor {
	t.Helper()
	if in.Now.IsZero() {
		in.Now = floorNow
	}
	out, err := riskcalc.ConfirmedFloorQuantity(in)
	if err != nil {
		t.Fatalf("ConfirmedFloorQuantity: %v", err)
	}
	return out
}

func TestTheFloorIsTheFormula(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                  string
		holdings, sellable    string
		localSells, wantFloor string
		wantBound             string
	}{
		{"holdings bind", "10", "40", "0", "10", riskcalc.FloorBoundHoldings},
		{"sellable binds", "40", "10", "0", "10", riskcalc.FloorBoundSellable},
		{"local sells consume part", "10", "10", "4", "6", riskcalc.FloorBoundLocalSells},
		{"local sells consume all", "10", "10", "10", "0", riskcalc.FloorBoundLocalSells},
		{"local sells exceed the account", "10", "10", "25", "0", riskcalc.FloorBoundLocalSells},
		{"fractional", "0.5", "0.5", "0.25", "0.25", riskcalc.FloorBoundLocalSells},
		{"nothing held", "0", "0", "0", "0", riskcalc.FloorBoundHoldings},
	}
	for _, c := range cases {
		got := floor(t, riskcalc.ConfirmedFloorInputs{
			Symbol:                "005930",
			Holdings:              snap(c.holdings, time.Second),
			Sellable:              snap(c.sellable, time.Second),
			LocalOpenSellQuantity: c.localSells,
		})
		if got.Quantity != c.wantFloor {
			t.Errorf("%s: floor = %q, want %q (%s)", c.name, got.Quantity, c.wantFloor, got.Detail)
		}
		if got.Bound != c.wantBound {
			t.Errorf("%s: bound = %q, want %q", c.name, got.Bound, c.wantBound)
		}
	}
}

// TestAnAbsentSnapshotIsZero is the spec's "스냅샷 부재 → 0 (자동 경로의 로컬
// 매도 없음)". It is not an error, because a caller with a separate "could not
// compute" path is a caller that will eventually do something else in it.
func TestAnAbsentSnapshotIsZero(t *testing.T) {
	t.Parallel()

	noHoldings := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol: "005930", Holdings: nil, Sellable: snap("10", time.Second),
		LocalOpenSellQuantity: "0",
	})
	if !noHoldings.Zero() || noHoldings.Bound != riskcalc.FloorBoundNoSnapshot {
		t.Fatalf("an unread holding must floor at 0: %+v", noHoldings)
	}

	noSellable := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol: "005930", Holdings: snap("10", time.Second), Sellable: nil,
		LocalOpenSellQuantity: "0",
	})
	if !noSellable.Zero() || noSellable.Bound != riskcalc.FloorBoundNoSnapshot {
		t.Fatalf("an unread sellable quantity must floor at 0: %+v", noSellable)
	}
}

func TestAStaleSnapshotIsZero(t *testing.T) {
	t.Parallel()

	// AccountSnapshotStaleness is 10s.
	stale := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol:                "005930",
		Holdings:              snap("10", 11*time.Second),
		Sellable:              snap("10", time.Second),
		LocalOpenSellQuantity: "0",
	})
	if !stale.Zero() || stale.Bound != riskcalc.FloorBoundStaleSnapshot {
		t.Fatalf("an 11s-old holding must floor at 0: %+v", stale)
	}

	// The boundary is inclusive: exactly at the bound is still fresh, and one
	// second past it is not. A test that only checked "old" would pass against
	// an implementation that treated every snapshot as stale.
	atBound := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol:                "005930",
		Holdings:              snap("10", riskcalc.AccountSnapshotStaleness),
		Sellable:              snap("10", riskcalc.AccountSnapshotStaleness),
		LocalOpenSellQuantity: "0",
	})
	if atBound.Quantity != "10" {
		t.Fatalf("a snapshot exactly at the bound is fresh: %+v", atBound)
	}

	// A narrowed bound makes the same snapshot stale; a widened one does not
	// make a stale snapshot fresh (§0.9's one-way ratchet).
	narrowed := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol:                "005930",
		Holdings:              snap("10", 5*time.Second),
		Sellable:              snap("10", time.Second),
		Bounds:                riskcalc.Staleness{AccountSnapshot: 2 * time.Second},
		LocalOpenSellQuantity: "0",
	})
	if !narrowed.Zero() {
		t.Fatalf("a caller-narrowed bound must apply: %+v", narrowed)
	}
	widened := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol:                "005930",
		Holdings:              snap("10", 30*time.Second),
		Sellable:              snap("10", time.Second),
		Bounds:                riskcalc.Staleness{AccountSnapshot: time.Hour},
		LocalOpenSellQuantity: "0",
	})
	if !widened.Zero() {
		t.Fatalf("a caller must not be able to widen the bound: %+v", widened)
	}
}

// TestALocalQuantityCanNeverRaiseTheFloor is the SHALL NOT: 로컬 파생 수량을
// 하한을 올리는 근거로 사용해서는 안 된다. The local ledger does not know about
// orders a human placed in the app, so it can be larger than the account's.
func TestALocalQuantityCanNeverRaiseTheFloor(t *testing.T) {
	t.Parallel()

	base := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol: "005930", Holdings: snap("10", time.Second), Sellable: snap("10", time.Second),
		LocalOpenSellQuantity: "0",
	})
	if base.Quantity != "10" {
		t.Fatalf("baseline floor = %q, want 10", base.Quantity)
	}

	// Every non-zero local quantity lowers it, and a negative one — the only
	// shape that could raise it — is refused outright rather than clamped.
	for _, local := range []string{"1", "9.9999", "10", "1000"} {
		got := floor(t, riskcalc.ConfirmedFloorInputs{
			Symbol: "005930", Holdings: snap("10", time.Second), Sellable: snap("10", time.Second),
			LocalOpenSellQuantity: local,
		})
		cmp, err := riskcalc.CompareDecimal(got.Quantity, base.Quantity)
		if err != nil {
			t.Fatalf("CompareDecimal: %v", err)
		}
		if cmp > 0 {
			t.Errorf("local open sells of %s raised the floor to %s", local, got.Quantity)
		}
	}
	_, err := riskcalc.ConfirmedFloorQuantity(riskcalc.ConfirmedFloorInputs{
		Now: floorNow, Symbol: "005930",
		Holdings: snap("10", time.Second), Sellable: snap("10", time.Second),
		LocalOpenSellQuantity: "-5",
	})
	if !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("a negative local quantity would raise the bound and must be refused, got %v", err)
	}
}

// TestAnUnknownLocalQuantityIsRefused: "" is unknown, not zero. Treating it as
// zero would authorise selling quantity a resting order has already committed.
func TestAnUnknownLocalQuantityIsRefused(t *testing.T) {
	t.Parallel()

	_, err := riskcalc.ConfirmedFloorQuantity(riskcalc.ConfirmedFloorInputs{
		Now: floorNow, Symbol: "005930",
		Holdings: snap("10", time.Second), Sellable: snap("10", time.Second),
		LocalOpenSellQuantity: "",
	})
	if !errors.Is(err, riskcalc.ErrUnknownInput) {
		t.Fatalf("an empty local quantity is unknown, not zero: %v", err)
	}
}

// TestTheSellableQuantityOnlyLowers: its semantics are [미측정 — 2b 2.8], so a
// value above the holding must change nothing.
func TestTheSellableQuantityOnlyLowers(t *testing.T) {
	t.Parallel()

	for _, sellable := range []string{"10", "11", "1000000"} {
		got := floor(t, riskcalc.ConfirmedFloorInputs{
			Symbol: "005930", Holdings: snap("10", time.Second), Sellable: snap(sellable, time.Second),
			LocalOpenSellQuantity: "0",
		})
		if got.Quantity != "10" {
			t.Errorf("sellable %s changed the floor to %s; it may only lower", sellable, got.Quantity)
		}
	}
	lower := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol: "005930", Holdings: snap("10", time.Second), Sellable: snap("3", time.Second),
		LocalOpenSellQuantity: "0",
	})
	if lower.Quantity != "3" || lower.Bound != riskcalc.FloorBoundSellable {
		t.Fatalf("a smaller sellable quantity must bind: %+v", lower)
	}
}

func TestMalformedQuantitiesAreRefused(t *testing.T) {
	t.Parallel()

	for _, in := range []riskcalc.ConfirmedFloorInputs{
		{Now: floorNow, Holdings: snap("-1", time.Second), Sellable: snap("1", time.Second), LocalOpenSellQuantity: "0"},
		{Now: floorNow, Holdings: snap("1", time.Second), Sellable: snap("abc", time.Second), LocalOpenSellQuantity: "0"},
		{Holdings: snap("1", time.Second), Sellable: snap("1", time.Second), LocalOpenSellQuantity: "0"},
	} {
		if _, err := riskcalc.ConfirmedFloorQuantity(in); err == nil {
			t.Errorf("input %+v must be refused", in)
		}
	}
}

// TestASnapshotWithNoAsOfIsNotFresh: a quantity that arrives with no timestamp
// cannot be shown to be recent, and "we have a number" is not freshness.
func TestASnapshotWithNoAsOfIsNotFresh(t *testing.T) {
	t.Parallel()

	got := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol:                "005930",
		Holdings:              &riskcalc.QuantitySnapshot{Quantity: "10"},
		Sellable:              snap("10", time.Second),
		LocalOpenSellQuantity: "0",
	})
	if !got.Zero() {
		t.Fatalf("a snapshot with no as-of must floor at 0: %+v", got)
	}
}

// TestAFutureSnapshotIsNotFresh: a negative age means the clocks disagree, and
// a clock nobody can trust is not a freshness proof.
func TestAFutureSnapshotIsNotFresh(t *testing.T) {
	t.Parallel()

	got := floor(t, riskcalc.ConfirmedFloorInputs{
		Symbol:                "005930",
		Holdings:              snap("10", -time.Minute),
		Sellable:              snap("10", time.Second),
		LocalOpenSellQuantity: "0",
	})
	if !got.Zero() {
		t.Fatalf("a snapshot dated in the future must floor at 0: %+v", got)
	}
}
