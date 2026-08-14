package reconcile_test

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a110MissingOrder is deliberately complete by default.  Each test that blanks
// one component is about whether the promotion evidence can prove sameness, not
// whether the ordinary missing-order block remains fail-closed.
func a110MissingOrder(account, market, day, symbol, side, id string) reconcile.LocalOrder {
	return reconcile.LocalOrder{
		AccountRef: account,
		Market:     market,
		TradingDay: day,
		Symbol:     symbol,
		Side:       side,
		OrderID:    id,
	}
}

func a110MissingDiff(orders ...reconcile.LocalOrder) reconcile.Diff {
	return reconcile.Diff{AccountRef: "acct-7", MissingOrders: orders}
}

func a110MissingObserve(t *testing.T, tracker *reconcile.Tracker, diff reconcile.Diff) reconcile.Outcome {
	t.Helper()
	out, err := tracker.Observe(context.Background(), diff)
	if err != nil {
		t.Fatalf("Observe(%+v): %v", diff, err)
	}
	return out
}

func a110MissingTracker(clk clock.Clock, gate *execgw.EntryGate, store reconcile.ReconcileStore) *reconcile.Tracker {
	return &reconcile.Tracker{
		Clock:       clk,
		Gate:        gate,
		Journal:     store,
		MinInterval: 30 * time.Second,
		MaxFailures: reconcile.DefaultMaxFailures,
		AccountRef:  "acct-7",
	}
}

func a110MissingAdvance(clk *clock.Fake) { clk.Advance(30 * time.Second) }

// TestA110CanonicalMissingOrderIdentityControlsPromotion says that the opaque
// order id is not enough.  A genuine repeated full identity still earns the
// existing account-wide operator block, but changing any canonical scope member
// begins a new streak instead of inheriting evidence from a different order.
func TestA110CanonicalMissingOrderIdentityControlsPromotion(t *testing.T) {
	t.Run("one complete identity reaches threshold", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110MissingTracker(clk, nil, nil)
		for _, order := range []reconcile.LocalOrder{
			a110MissingOrder(" acct-7 ", " US ", " 2026-03-30 ", " aapl ", " buy ", "opaque-7"),
			a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", "opaque-7"),
			a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", "opaque-7"),
		} {
			out := a110MissingObserve(t, tracker, a110MissingDiff(order))
			a110MissingAdvance(clk)
			if out.Failures == 0 {
				t.Fatalf("complete missing-order identity earned no streak: %+v", out)
			}
		}
		if !tracker.Permanent() {
			t.Fatal("one canonical missing order repeated at threshold did not promote")
		}
	})

	base := a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", "opaque-7")
	variations := []struct {
		name         string
		mutate       func(*reconcile.LocalOrder)
		foreignScope bool
	}{
		{name: "account", foreignScope: true, mutate: func(o *reconcile.LocalOrder) { o.AccountRef = "acct-8" }},
		{name: "market", mutate: func(o *reconcile.LocalOrder) { o.Market = "kr" }},
		{name: "trading day", mutate: func(o *reconcile.LocalOrder) { o.TradingDay = "2026-03-31" }},
		{name: "symbol", mutate: func(o *reconcile.LocalOrder) { o.Symbol = "MSFT" }},
		{name: "side", mutate: func(o *reconcile.LocalOrder) { o.Side = "SELL" }},
	}
	for _, tt := range variations {
		t.Run("reused opaque id with different "+tt.name+" cannot share", func(t *testing.T) {
			clk := clock.NewFake(asOf)
			tracker := a110MissingTracker(clk, nil, nil)
			for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
				a110MissingObserve(t, tracker, a110MissingDiff(base))
				a110MissingAdvance(clk)
			}
			changed := base
			tt.mutate(&changed)
			out := a110MissingObserve(t, tracker, a110MissingDiff(changed))
			if out.Permanent || tracker.Permanent() {
				t.Fatalf("changed %s inherited the opaque id's streak: %+v", tt.name, out)
			}
			if tt.foreignScope {
				if out.Failures != 0 {
					t.Fatalf("foreign account earned this tracker's streak = %d, want 0", out.Failures)
				}
				if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil ||
					rejected.Reason != execgw.ReasonReconcileMismatch {
					t.Fatalf("foreign account lost its ordinary fail-closed block: %v", rejected)
				}
			} else if out.Failures != 1 {
				t.Fatalf("changed %s streak = %d, want 1", tt.name, out.Failures)
			}
		})
	}
}

// TestA110OpaqueOrderIDIsBytePreservingPromotionEvidence distinguishes the
// order id from the five normalized scope fields.  The opaque id is supplied by
// the broker and must remain byte-preserving: surrounding whitespace is not a
// formatting variant the tracker may silently erase, while a nonblank id that
// itself contains those bytes is still valid repeatable evidence.
func TestA110OpaqueOrderIDIsBytePreservingPromotionEvidence(t *testing.T) {
	t.Run("changed opaque id bytes reset the streak", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110MissingTracker(clk, nil, nil)
		base := a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", "opaque-7")
		for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
			a110MissingObserve(t, tracker, a110MissingDiff(base))
			a110MissingAdvance(clk)
		}
		changed := base
		changed.OrderID = " opaque-7 "
		out := a110MissingObserve(t, tracker, a110MissingDiff(changed))
		if out.Permanent || tracker.Permanent() || out.Failures != 1 {
			t.Fatalf("opaque id bytes were normalized into the old streak: %+v", out)
		}
	})

	t.Run("exact whitespace-containing opaque id remains valid repeatable evidence", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110MissingTracker(clk, nil, nil)
		order := a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", " opaque-7 ")
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			a110MissingObserve(t, tracker, a110MissingDiff(order))
			a110MissingAdvance(clk)
		}
		if !tracker.Permanent() || tracker.Failures() != reconcile.DefaultMaxFailures {
			t.Fatalf("exact nonblank opaque id bytes were rejected instead of promoted: permanent=%v failures=%d",
				tracker.Permanent(), tracker.Failures())
		}
	})
}

// TestA110ForeignMissingOrderAccountCannotEarnThisTrackerPermanent requires
// the complete missing-order identity to describe the account this tracker
// protects.  A foreign-account order remains fail-closed as ordinary evidence,
// but cannot earn an operator-only outage on acct-7.
func TestA110ForeignMissingOrderAccountCannotEarnThisTrackerPermanent(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := noStaleGate(clk)
	tracker := a110MissingTracker(clk, gate, nil)
	foreign := a110MissingOrder("acct-8", "us", "2026-03-30", "AAPL", "BUY", "opaque-foreign")
	for i := 0; i < reconcile.DefaultMaxFailures; i++ {
		out := a110MissingObserve(t, tracker, a110MissingDiff(foreign))
		if !out.Blocked {
			t.Fatalf("foreign-account missing order stopped being ordinary fail-closed evidence: %+v", out)
		}
		a110MissingAdvance(clk)
	}
	if tracker.Permanent() || tracker.Failures() != 0 {
		t.Fatalf("foreign account earned acct-7 permanent evidence: permanent=%v failures=%d",
			tracker.Permanent(), tracker.Failures())
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("foreign-account item lost its ordinary acct-7 fail-closed gate: %v", rejected)
	}
}

// TestA110MissingOrderPromotionRequiresTheTrackersAuthoritativeDiffAccount
// pins both halves of account authority.  A complete LocalOrder identity for
// acct-7 is not promotion evidence when the comparison itself is unscoped or
// belongs to acct-8.  The finding still blocks AAPL ordinarily; only the
// operator-only account escalation is withheld.
func TestA110MissingOrderPromotionRequiresTheTrackersAuthoritativeDiffAccount(t *testing.T) {
	order := a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", "opaque-7")
	for _, diffAccount := range []string{"", "acct-8"} {
		name := "blank diff account"
		if diffAccount != "" {
			name = "foreign diff account"
		}
		t.Run(name+" stays ordinary without promotion evidence", func(t *testing.T) {
			ctx := context.Background()
			clk := clock.NewFake(asOf)
			store := openJournal(t)
			gate := noStaleGate(clk)
			tracker := a110MissingTracker(clk, gate, store)
			diff := reconcile.Diff{AccountRef: diffAccount, MissingOrders: []reconcile.LocalOrder{order}}
			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				out := a110MissingObserve(t, tracker, diff)
				if !out.Blocked {
					t.Errorf("%s stopped being ordinary fail-closed evidence: %+v", name, out)
				}
				a110MissingAdvance(clk)
			}
			if tracker.Permanent() || tracker.Failures() != 0 {
				t.Errorf("%s earned acct-7 promotion evidence: permanent=%v failures=%d",
					name, tracker.Permanent(), tracker.Failures())
			}
			if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
				rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Errorf("%s lost its ordinary AAPL gate: %v", name, rejected)
			}
			states, err := store.ActiveReconcileStates(ctx)
			if err != nil {
				t.Fatalf("ActiveReconcileStates: %v", err)
			}
			for _, state := range states {
				if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
					t.Errorf("%s created a durable account permanent row: %+v", name, state)
				}
			}
		})
	}

	t.Run("current authoritative diff account still promotes", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		store := openJournal(t)
		tracker := a110MissingTracker(clk, nil, store)
		diff := reconcile.Diff{AccountRef: "acct-7", MissingOrders: []reconcile.LocalOrder{order}}
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			a110MissingObserve(t, tracker, diff)
			a110MissingAdvance(clk)
		}
		if !tracker.Permanent() || tracker.Failures() != reconcile.DefaultMaxFailures {
			t.Fatalf("authoritative acct-7 diff failed to promote: permanent=%v failures=%d",
				tracker.Permanent(), tracker.Failures())
		}
	})
}

// TestA110IncompleteMissingOrderIdentityNeverPromotes keeps the direction of
// failure explicit: an incomplete order remains an entry block, but because the
// tracker cannot prove it is the same order as before it cannot manufacture an
// operator-only account outage from the repeats.
func TestA110IncompleteMissingOrderIdentityNeverPromotes(t *testing.T) {
	base := a110MissingOrder("acct-7", "us", "2026-03-30", "AAPL", "BUY", "opaque-7")
	blanks := []struct {
		name   string
		mutate func(*reconcile.LocalOrder)
	}{
		{name: "account", mutate: func(o *reconcile.LocalOrder) { o.AccountRef = " " }},
		{name: "market", mutate: func(o *reconcile.LocalOrder) { o.Market = " " }},
		{name: "trading day", mutate: func(o *reconcile.LocalOrder) { o.TradingDay = " " }},
		{name: "symbol", mutate: func(o *reconcile.LocalOrder) { o.Symbol = " " }},
		{name: "side", mutate: func(o *reconcile.LocalOrder) { o.Side = " " }},
		{name: "opaque order id", mutate: func(o *reconcile.LocalOrder) { o.OrderID = " " }},
	}
	for _, tt := range blanks {
		t.Run(tt.name+" stays ordinary but earns no evidence", func(t *testing.T) {
			clk := clock.NewFake(asOf)
			gate := noStaleGate(clk)
			tracker := a110MissingTracker(clk, gate, nil)
			order := base
			tt.mutate(&order)
			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				out := a110MissingObserve(t, tracker, a110MissingDiff(order))
				if !out.Blocked {
					t.Fatalf("blank %s stopped being an ordinary block: %+v", tt.name, out)
				}
				a110MissingAdvance(clk)
			}
			if tracker.Permanent() || tracker.Failures() != 0 {
				t.Fatalf("blank %s earned permanent evidence: permanent=%v failures=%d", tt.name,
					tracker.Permanent(), tracker.Failures())
			}
			if rejected := gate.CheckEntryFor("us", order.Symbol); rejected == nil ||
				rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Fatalf("blank %s ordinary gate = %v, want mismatch", tt.name, rejected)
			}
		})
	}

	t.Run("valid sibling still earns its own streak", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110MissingTracker(clk, nil, nil)
		incomplete := base
		incomplete.OrderID = " "
		valid := a110MissingOrder("acct-7", "kr", "2026-03-30", "005930", "SELL", "opaque-valid")
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			a110MissingObserve(t, tracker, a110MissingDiff(incomplete, valid))
			a110MissingAdvance(clk)
		}
		if !tracker.Permanent() {
			t.Fatal("valid sibling did not earn its independently repeated promotion")
		}
	})
}

// TestA110IncidentChangingSymbolsReleaseOrdinaryBlocksWithoutPermanent is a
// sanitized, durable version of 2026-08-07.  The first two disagreements earn
// adjustment credits, three different symbols appear before the recheck, and
// the recheck releases only the credited ordinary rows.  A false account-wide
// permanent row must never survive that sequence.
func TestA110IncidentChangingSymbolsReleaseOrdinaryBlocksWithoutPermanent(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	store := openJournal(t)
	gate := noStaleGate(clk)
	tracker := a110MissingTracker(clk, gate, store)

	first := mismatchDiffAt(clk, "AAPL", "10", "4")
	a110MissingObserve(t, tracker, first)
	tracker.AdjustmentApplied(first.AsOf, "AAPL")
	a110MissingAdvance(clk)

	second := mismatchDiffAt(clk, "MSFT", "7", "2")
	a110MissingObserve(t, tracker, second)
	tracker.AdjustmentApplied(second.AsOf, "MSFT")
	a110MissingAdvance(clk)

	third := mismatchDiffAt(clk, "TSLA", "3", "0")
	out := a110MissingObserve(t, tracker, third)
	if out.Permanent || tracker.Permanent() {
		t.Fatalf("three changing symbols became a false permanent block: %+v", out)
	}
	a110MissingAdvance(clk)

	out = a110MissingObserve(t, tracker, cleanDiffAt(clk))
	if tracker.Permanent() {
		t.Fatalf("a false account permanent survived the incident recheck: %+v", tracker.Blocks())
	}
	cleared := map[string]bool{}
	for _, block := range out.Cleared {
		cleared[block.Symbol] = true
	}
	if !cleared["AAPL"] || !cleared["MSFT"] || len(cleared) != 2 {
		t.Fatalf("recheck cleared = %+v, want exactly the two adjustment-backed rows", out.Cleared)
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("ordinary symbol rows must not leave a permanent account gate: %v", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "TSLA"); rejected == nil ||
		rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("uncredited TSLA row must remain ordinary fail-closed: %v", rejected)
	}
	states, err := store.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	for _, state := range states {
		if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
			t.Fatalf("incident left a durable account permanent row: %+v", state)
		}
	}
}

// TestA110TransientStreakIsNeverReconstructedFromDurableOrdinaryRows records
// the process-local boundary.  The journal restores ordinary safety blocks, not
// a proof that an exact dispute occurred twice before this process began.
func TestA110TransientStreakIsNeverReconstructedFromDurableOrdinaryRows(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	store := openJournal(t)
	first := a110MissingTracker(clk, noStaleGate(clk), store)
	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110MissingObserve(t, first, mismatchDiff("AAPL", "10", "4"))
		a110MissingAdvance(clk)
	}

	for _, load := range []struct {
		name string
		fn   func(*reconcile.Tracker) error
	}{
		{name: "restore", fn: func(t *reconcile.Tracker) error { return t.Restore(ctx) }},
		{name: "refresh", fn: func(t *reconcile.Tracker) error { return t.Refresh(ctx) }},
	} {
		t.Run(load.name+" does not manufacture the third observation", func(t *testing.T) {
			tracker := a110MissingTracker(clk, noStaleGate(clk), store)
			if err := load.fn(tracker); err != nil {
				t.Fatalf("%s: %v", load.name, err)
			}
			if tracker.Failures() != 0 {
				t.Fatalf("%s reconstructed a transient streak before an observation: %d", load.name, tracker.Failures())
			}
			out := a110MissingObserve(t, tracker, mismatchDiff("AAPL", "10", "4"))
			if out.Permanent || tracker.Permanent() || out.Failures != 1 {
				t.Fatalf("%s turned a durable ordinary row into a third streak: %+v", load.name, out)
			}
		})
	}

	t.Run("a durable permanent still restores", func(t *testing.T) {
		permanentStore := openJournal(t)
		permanentClock := clock.NewFake(asOf)
		writer := a110MissingTracker(permanentClock, noStaleGate(permanentClock), permanentStore)
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			a110MissingObserve(t, writer, mismatchDiff("AAPL", "10", "4"))
			a110MissingAdvance(permanentClock)
		}
		if !writer.Permanent() {
			t.Fatal("precondition: one exact dispute must create a durable permanent block")
		}
		restarted := a110MissingTracker(permanentClock, noStaleGate(permanentClock), permanentStore)
		if err := restarted.Restore(ctx); err != nil {
			t.Fatalf("Restore permanent: %v", err)
		}
		if !restarted.Permanent() {
			t.Fatal("durable permanent was lost on Restore")
		}
	})
}
