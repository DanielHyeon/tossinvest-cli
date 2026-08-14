package reconcile_test

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// a110_promotion_identity_test.go is deliberately a black-box contract for the
// promotion boundary.  Ordinary blocks are fail-closed on every mismatch; only
// the evidence that earns an account-wide operator block is identity-sensitive.

func a110QuantityDiff(items ...reconcile.QuantityMismatch) reconcile.Diff {
	return reconcile.Diff{AccountRef: "acct-7", Quantities: items}
}

func a110Quantity(symbol, local, broker string) reconcile.QuantityMismatch {
	return reconcile.QuantityMismatch{Symbol: symbol, Local: local, Broker: broker}
}

func a110Observe(t *testing.T, tracker *reconcile.Tracker, diff reconcile.Diff) reconcile.Outcome {
	t.Helper()
	out, err := tracker.Observe(context.Background(), diff)
	if err != nil {
		t.Fatalf("Observe(%+v): %v", diff, err)
	}
	return out
}

func a110ObserveError(t *testing.T, tracker *reconcile.Tracker, diff reconcile.Diff) reconcile.Outcome {
	t.Helper()
	out, err := tracker.Observe(context.Background(), diff)
	if err == nil {
		t.Fatalf("Observe(%+v) succeeded, want durable-write failure", diff)
	}
	return out
}

func a110Tracker(clk clock.Clock, gate *execgw.EntryGate, store reconcile.ReconcileStore) *reconcile.Tracker {
	return &reconcile.Tracker{
		Clock:       clk,
		Gate:        gate,
		Journal:     store,
		MinInterval: 30 * time.Second,
		MaxFailures: reconcile.DefaultMaxFailures,
		AccountRef:  "acct-7",
	}
}

func a110Advance(clk *clock.Fake) { clk.Advance(30 * time.Second) }

func a110HasAccountPermanent(t *testing.T, store reconcile.ReconcileStore) bool {
	t.Helper()
	states, err := store.ActiveReconcileStates(context.Background())
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	for _, state := range states {
		if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
			return true
		}
	}
	return false
}

// a110AssertNoTransientAccountPermanent uses only the public projection.  The
// pending bit is intentionally private; the observable fail-safe contract is
// stronger: after continuity breaks, no account-wide permanent block may remain
// at all, in memory, in the gate, or in the durable store.
func a110AssertNoTransientAccountPermanent(t *testing.T, tracker *reconcile.Tracker, gate *execgw.EntryGate) {
	t.Helper()
	if tracker.Permanent() {
		t.Fatalf("stale pending permanent still marked tracker permanent: %+v", tracker.Blocks())
	}
	for _, block := range tracker.Blocks() {
		if block.Scope == reconcile.ScopeAccount && block.Permanent {
			t.Fatalf("stale pending account permanent survived continuity break: %+v", block)
		}
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("stale account-wide permanent left the account gate closed: %v", rejected)
	}
}

// TestA110ChangingQuantityDisputesNeverPoolIntoPermanent proves the incident
// shape: three independently blocking symbols remain ordinary blocks instead of
// becoming a false account-wide "look again did not work" verdict.
func TestA110ChangingQuantityDisputesNeverPoolIntoPermanent(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, nil)

	for _, mismatch := range []reconcile.QuantityMismatch{
		a110Quantity("AAPL", "10", "4"),
		a110Quantity("MSFT", "7", "2"),
		a110Quantity("TSLA", "3", "0"),
	} {
		out := a110Observe(t, tracker, a110QuantityDiff(mismatch))
		if !out.Blocked {
			t.Fatalf("%s was not fail-closed as an ordinary mismatch: %+v", mismatch.Symbol, out)
		}
		if rejected := gate.CheckEntryFor("us", mismatch.Symbol); rejected == nil ||
			rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Fatalf("ordinary block for %s = %v, want reconcile mismatch", mismatch.Symbol, rejected)
		}
		a110Advance(clk)
	}

	if tracker.Permanent() {
		t.Fatalf("different disputes pooled into permanent block: %+v", tracker.Blocks())
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Fatalf("changing symbol mismatches must not close account-wide gate: %v", rejected)
	}
}

func TestA110QuantityPromotionRejectsBlankOrForeignDiffAccount(t *testing.T) {
	for _, diffAccount := range []string{"", "acct-8"} {
		name := diffAccount
		if name == "" {
			name = "blank"
		}
		t.Run(name, func(t *testing.T) {
			clk := clock.NewFake(asOf)
			gate := noStaleGate(clk)
			tracker := a110Tracker(clk, gate, nil) // tracker authority is acct-7
			diff := reconcile.Diff{
				AccountRef: diffAccount,
				Quantities: []reconcile.QuantityMismatch{a110Quantity("AAPL", "10", "4")},
			}
			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				out := a110Observe(t, tracker, diff)
				if !out.Blocked {
					t.Fatalf("%q diff account stopped being ordinary fail-closed: %+v", diffAccount, out)
				}
				a110Advance(clk)
			}
			if tracker.Permanent() || tracker.Failures() != 0 {
				t.Fatalf("%q diff account earned acct-7 promotion: permanent=%v failures=%d", diffAccount, tracker.Permanent(), tracker.Failures())
			}
			if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Fatalf("%q diff account lost ordinary AAPL block: %v", diffAccount, rejected)
			}
		})
	}
}

func TestA110QuantityTupleIdentityIsExactAndCanonical(t *testing.T) {
	t.Run("changed tuple starts a new streak", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110Tracker(clk, nil, nil)
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
		out := a110Observe(t, tracker, mismatchDiff("AAPL", "11", "4"))
		if out.Permanent || tracker.Permanent() {
			t.Fatalf("changed quantity tuple inherited prior streak: %+v", out)
		}
		if out.Failures != 1 {
			t.Fatalf("changed tuple streak = %d, want 1", out.Failures)
		}
	})

	t.Run("equivalent decimal spellings continue one tuple", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110Tracker(clk, nil, nil)
		a110Observe(t, tracker, mismatchDiff(" aapl ", "0010.00", "4.0"))
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4.000"))
		a110Advance(clk)
		out := a110Observe(t, tracker, mismatchDiff("AAPL", "+10.0000", "04"))
		if !out.Permanent || out.Failures != reconcile.DefaultMaxFailures {
			t.Fatalf("canonical decimal spellings failed to continue the exact dispute: %+v", out)
		}
	})

	t.Run("negative and positive zero spellings continue one tuple", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110Tracker(clk, nil, nil)
		a110Observe(t, tracker, mismatchDiff("AAPL", "-0", "-0.000"))
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("AAPL", "0.0", "0"))
		a110Advance(clk)
		out := a110Observe(t, tracker, mismatchDiff("AAPL", "-0.0000", "+0.000"))
		if !out.Permanent || out.Failures != reconcile.DefaultMaxFailures {
			t.Fatalf("zero spellings failed to continue one exact tuple: %+v", out)
		}
	})

	t.Run("large decimal strings never collide through float64", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110Tracker(clk, nil, nil)
		a110Observe(t, tracker, mismatchDiff("AAPL", "9007199254740992", "0"))
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("AAPL", "9007199254740992.0", "0.0"))
		a110Advance(clk)
		out := a110Observe(t, tracker, mismatchDiff("AAPL", "9007199254740993", "0"))
		if out.Permanent || tracker.Permanent() {
			t.Fatalf("distinct 2^53-adjacent decimal tuples collided: %+v", out)
		}
		if out.Failures != 1 {
			t.Fatalf("new large-decimal tuple streak = %d, want 1", out.Failures)
		}
	})
}

func TestA110ComparerPreservesExactLargeQuantityPromotionIdentity(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := a110Tracker(clk, nil, nil)
	snap := snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "1"}}, nil)

	// These integers collide in float64.  The same first value is deliberately
	// repeated after the second: with exact Diff strings neither identity is
	// consecutive, while a comparer that rounds both to 2^53 manufactures a
	// three-observation promotion streak before Tracker ever sees the evidence.
	for _, exactLocal := range []string{
		"9007199254740992",
		"9007199254740993",
		"9007199254740992",
	} {
		diff := (reconcile.Comparer{}).Compare(snap, reconcile.LocalState{
			AccountRef: "acct-7",
			Positions:  map[string]string{"AAPL": exactLocal},
		})
		if len(diff.Quantities) != 1 {
			t.Fatalf("Comparer(%s) quantities = %+v, want one real mismatch", exactLocal, diff.Quantities)
		}
		if got := diff.Quantities[0].Local; got != exactLocal {
			t.Errorf("Comparer collapsed exact local %s to %s before promotion", exactLocal, got)
		}
		a110Observe(t, tracker, diff)
		a110Advance(clk)
	}
	if tracker.Permanent() || tracker.Failures() != 1 {
		t.Fatalf("Comparer-rounded identities falsely accumulated: permanent=%v failures=%d blocks=%+v", tracker.Permanent(), tracker.Failures(), tracker.Blocks())
	}
}

func TestA110ComparerDoesNotTreatDistinctLargeIntegersAsEqual(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := a110Tracker(clk, nil, nil)
	snap := snapshotWith([]reconcile.Holding{{
		Symbol: "AAPL", Quantity: "9007199254740992",
	}}, nil)
	diff := (reconcile.Comparer{}).Compare(snap, reconcile.LocalState{
		AccountRef: "acct-7",
		Positions:  map[string]string{"AAPL": "9007199254740993"},
	})

	if len(diff.Quantities) != 1 {
		t.Fatalf("distinct exact large integers compared equal: diff=%+v", diff)
	}
	if got := diff.Quantities[0]; got.Local != "9007199254740993" || got.Broker != "9007199254740992" {
		t.Fatalf("large-integer mismatch lost exact evidence: %+v", got)
	}
	out := a110Observe(t, tracker, diff)
	if !out.Blocked || out.Failures != 1 || out.Permanent {
		t.Fatalf("first exact large-integer mismatch = %+v, want ordinary block and streak one", out)
	}
}

func TestA110ComparerMaxFloatULPDoesNotOverflowToEquality(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, nil)
	maxFloatFixed := strconv.FormatFloat(math.MaxFloat64, 'f', -1, 64)
	diff := (reconcile.Comparer{}).Compare(
		snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "1"}}, nil),
		reconcile.LocalState{
			AccountRef: "acct-7",
			Positions:  map[string]string{"AAPL": maxFloatFixed},
		},
	)

	if len(diff.Quantities) != 1 || diff.Matched != 0 || !diff.BlocksEntry() {
		t.Fatalf("MaxFloat64 nextafter overflow hid an exact mismatch: diff=%+v", diff)
	}
	out := a110Observe(t, tracker, diff)
	if !out.Blocked || out.Failures != 1 || out.Permanent {
		t.Fatalf("MaxFloat64 mismatch = %+v, want ordinary block and one exact observation", out)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("MaxFloat64 mismatch gate = %v, want ordinary reconcile mismatch", rejected)
	}
}

func TestA110ComparerDoesNotTreatDistinctExactIntegersOneULPApartAsArtifact(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, nil)
	diff := (reconcile.Comparer{}).Compare(
		snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "9007199254740994"}}, nil),
		reconcile.LocalState{
			AccountRef: "acct-7",
			Positions:  map[string]string{"AAPL": "9007199254740992"},
		},
	)

	if len(diff.Quantities) != 1 || diff.Matched != 0 || !diff.BlocksEntry() {
		t.Fatalf("distinct exact integers one float64 ULP apart were hidden as an artifact: %+v", diff)
	}
	out := a110Observe(t, tracker, diff)
	if !out.Blocked || out.Failures != 1 || out.Permanent {
		t.Fatalf("one-ULP exact-integer mismatch = %+v, want ordinary block and one exact observation", out)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("one-ULP exact-integer gate = %v, want ordinary reconcile mismatch", rejected)
	}
}

func TestA110ToleranceZeroRejectsNonArtifactDecimalDifferences(t *testing.T) {
	for _, tc := range []struct {
		name   string
		local  string
		broker string
	}{
		{name: "relative epsilon is not business tolerance", local: "1000000", broker: "1000000.0005"},
		{name: "near zero epsilon is not business tolerance", local: "0", broker: "0.0000000005"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			diff := (reconcile.Comparer{}).Compare(
				snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: tc.broker}}, nil),
				reconcile.LocalState{AccountRef: "acct-7", Positions: map[string]string{"AAPL": tc.local}},
			)
			if len(diff.Quantities) != 1 || !diff.BlocksEntry() {
				t.Fatalf("distinct exact quantities were hidden by legacy epsilon: local=%s broker=%s diff=%+v", tc.local, tc.broker, diff)
			}
		})
	}
}

func TestA110ComparerRejectsIdenticalInvalidQuantityStrings(t *testing.T) {
	for _, invalid := range []string{"NaN", "+Inf", "not-a-decimal"} {
		t.Run(invalid, func(t *testing.T) {
			clk := clock.NewFake(asOf)
			gate := noStaleGate(clk)
			tracker := a110Tracker(clk, gate, nil)

			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				snap := snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: invalid}}, nil)
				diff := (reconcile.Comparer{}).Compare(snap, reconcile.LocalState{
					AccountRef: "acct-7",
					Positions:  map[string]string{"AAPL": invalid},
				})
				if len(diff.Quantities) != 1 {
					t.Errorf("Comparer treated identical invalid %q as a valid match: %+v", invalid, diff)
				}
				out, err := tracker.Observe(context.Background(), diff)
				if err != nil {
					t.Fatalf("Observe(%q): %v", invalid, err)
				}
				if !out.Blocked {
					t.Errorf("identical invalid %q did not stay ordinary fail-closed: %+v", invalid, out)
				}
				a110Advance(clk)
			}

			if tracker.Failures() != 0 || tracker.Permanent() {
				t.Fatalf("invalid %q earned promotion evidence: failures=%d permanent=%v", invalid, tracker.Failures(), tracker.Permanent())
			}
			if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Errorf("Tracker.EntryAllowed(%q) = %v, want ordinary mismatch", invalid, rejected)
			}
			if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Errorf("entry gate(%q) = %v, want ordinary mismatch", invalid, rejected)
			}
		})
	}
}

func TestA110UnreadableExternalBrokerQuantityFailsClosed(t *testing.T) {
	for _, invalid := range []string{"", "NaN", "+Inf", "not-a-decimal"} {
		name := invalid
		if name == "" {
			name = "blank"
		}
		t.Run(name, func(t *testing.T) {
			clk := clock.NewFake(asOf)
			gate := noStaleGate(clk)
			tracker := a110Tracker(clk, gate, nil)
			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				snap := snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: invalid}}, nil)
				diff := (reconcile.Comparer{}).Compare(snap, reconcile.LocalState{
					AccountRef: "acct-7", Positions: map[string]string{},
				})
				if len(diff.Quantities) != 1 {
					t.Errorf("unreadable external broker quantity %q was nonblocking: %+v", invalid, diff)
				}
				if len(diff.ExternalPos) != 0 {
					t.Errorf("unreadable broker quantity %q was classified as owner exposure: %+v", invalid, diff.ExternalPos)
				}
				out, err := tracker.Observe(context.Background(), diff)
				if err != nil {
					t.Fatalf("Observe(%q): %v", invalid, err)
				}
				if !out.Blocked {
					t.Errorf("unreadable broker quantity %q did not remain fail-closed: %+v", invalid, out)
				}
				a110Advance(clk)
			}
			if tracker.Failures() != 0 || tracker.Permanent() {
				t.Fatalf("unreadable broker quantity %q earned promotion: failures=%d permanent=%v", invalid, tracker.Failures(), tracker.Permanent())
			}
			if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Errorf("Tracker.EntryAllowed(%q) = %v, want ordinary mismatch", invalid, rejected)
			}
			if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Errorf("entry gate(%q) = %v, want ordinary mismatch", invalid, rejected)
			}
		})
	}

	t.Run("valid positive external holding remains nonblocking", func(t *testing.T) {
		diff := (reconcile.Comparer{}).Compare(
			snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "4"}}, nil),
			reconcile.LocalState{AccountRef: "acct-7", Positions: map[string]string{}},
		)
		if diff.BlocksEntry() || len(diff.Quantities) != 0 || len(diff.ExternalPos) != 1 {
			t.Fatalf("valid positive owner exposure lost external-position contract: %+v", diff)
		}
	})
}

func TestA110NegativeBrokerOnlyHoldingFailsClosedWithoutPromotion(t *testing.T) {
	clk := clock.NewFake(asOf)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, nil)

	for i := 0; i < reconcile.DefaultMaxFailures; i++ {
		diff := (reconcile.Comparer{}).Compare(
			snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: "-1"}}, nil),
			reconcile.LocalState{AccountRef: "acct-7", Positions: map[string]string{}},
		)
		if len(diff.Quantities) != 1 || len(diff.ExternalPos) != 0 || !diff.BlocksEntry() {
			t.Errorf("negative broker-only quantity was classified as external/nonblocking: %+v", diff)
		}
		out := a110Observe(t, tracker, diff)
		if !out.Blocked {
			t.Errorf("negative broker-only quantity did not fail closed: %+v", out)
		}
		a110Advance(clk)
	}

	if tracker.Failures() != 0 || tracker.Permanent() {
		t.Fatalf("impossible negative broker-only projection earned promotion: failures=%d permanent=%v blocks=%+v", tracker.Failures(), tracker.Permanent(), tracker.Blocks())
	}
	if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("negative broker-only Tracker.EntryAllowed = %v, want ordinary mismatch", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("negative broker-only entry gate = %v, want ordinary mismatch", rejected)
	}
}

func TestA110BlankRawQuantityPathsFailClosedWithoutPromotion(t *testing.T) {
	for _, tc := range []struct {
		name   string
		local  string
		broker string
	}{
		{name: "blank on both sides", local: "", broker: ""},
		{name: "blank local against positive broker", local: "", broker: "4"},
		{name: "positive local against blank broker", local: "10", broker: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clk := clock.NewFake(asOf)
			gate := noStaleGate(clk)
			tracker := a110Tracker(clk, gate, nil)
			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				diff := (reconcile.Comparer{}).Compare(
					snapshotWith([]reconcile.Holding{{Symbol: "AAPL", Quantity: tc.broker}}, nil),
					reconcile.LocalState{AccountRef: "acct-7", Positions: map[string]string{"AAPL": tc.local}},
				)
				if len(diff.Quantities) != 1 || !diff.BlocksEntry() {
					t.Errorf("blank raw quantity was not an ordinary mismatch: local=%q broker=%q diff=%+v", tc.local, tc.broker, diff)
				}
				out, err := tracker.Observe(context.Background(), diff)
				if err != nil {
					t.Fatalf("Observe(local=%q broker=%q): %v", tc.local, tc.broker, err)
				}
				if !out.Blocked {
					t.Errorf("blank raw quantity did not remain fail-closed: %+v", out)
				}
				a110Advance(clk)
			}
			if tracker.Failures() != 0 || tracker.Permanent() {
				t.Fatalf("blank raw quantity earned promotion: failures=%d permanent=%v", tracker.Failures(), tracker.Permanent())
			}
			if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Errorf("blank raw quantity gate = %v, want ordinary mismatch", rejected)
			}
		})
	}
}

func TestA110CollectorPreservesUnreadableRawHoldingQuantity(t *testing.T) {
	t.Run("blank raw holding stays evidence and fails closed", func(t *testing.T) {
		collector, _, _, holdings, _ := newCollector(t)
		collector.Positions = &a110RawHoldingsReader{
			fakeHoldings: holdings,
			raw:          []reconcile.RawHolding{{Symbol: "AAPL", Quantity: ""}},
		}
		snap, err := collector.Collect(context.Background())
		if err != nil {
			t.Fatalf("Collect: %v", err)
		}
		if len(snap.Holdings) != 1 || snap.Holdings[0].Quantity != "" {
			t.Errorf("Collector destroyed blank raw quantity evidence: %+v", snap.Holdings)
		}

		clk := clock.NewFake(asOf)
		gate := noStaleGate(clk)
		tracker := a110Tracker(clk, gate, nil)
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			diff := (reconcile.Comparer{}).Compare(snap, reconcile.LocalState{
				AccountRef: "acct-7", Positions: map[string]string{},
			})
			if len(diff.Quantities) != 1 || !diff.BlocksEntry() {
				t.Errorf("collected blank holding was not an ordinary mismatch: %+v", diff)
			}
			out, observeErr := tracker.Observe(context.Background(), diff)
			if observeErr != nil {
				t.Fatalf("Observe: %v", observeErr)
			}
			if !out.Blocked {
				t.Errorf("collected blank holding did not remain fail-closed: %+v", out)
			}
			a110Advance(clk)
		}
		if tracker.Failures() != 0 || tracker.Permanent() {
			t.Fatalf("collected blank holding earned promotion: failures=%d permanent=%v", tracker.Failures(), tracker.Permanent())
		}
		if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Errorf("collected blank holding gate = %v, want ordinary mismatch", rejected)
		}
	})

	t.Run("positive raw holding remains canonical external exposure", func(t *testing.T) {
		collector, _, _, holdings, _ := newCollector(t)
		collector.Positions = &a110RawHoldingsReader{
			fakeHoldings: holdings,
			raw:          []reconcile.RawHolding{{Symbol: " aapl ", Quantity: "04.00"}},
		}
		snap, err := collector.Collect(context.Background())
		if err != nil {
			t.Fatalf("Collect: %v", err)
		}
		if len(snap.Holdings) != 1 || snap.Holdings[0].Symbol != "AAPL" || snap.Holdings[0].Quantity != "4" {
			t.Fatalf("positive raw holding canonicalization = %+v, want AAPL/4", snap.Holdings)
		}
		diff := (reconcile.Comparer{}).Compare(snap, reconcile.LocalState{
			AccountRef: "acct-7", Positions: map[string]string{},
		})
		if diff.BlocksEntry() || len(diff.ExternalPos) != 1 {
			t.Fatalf("positive collected holding lost nonblocking external contract: %+v", diff)
		}
	})
}

func TestA110SnapshotDigestAndStabiliserDistinguishBlankHoldingFromExactZero(t *testing.T) {
	ctx := context.Background()
	collector, _, _, holdings, _ := newCollector(t)
	raw := &a110RawHoldingsReader{
		fakeHoldings: holdings,
		raw:          []reconcile.RawHolding{{Symbol: "AAPL", Quantity: ""}},
	}
	collector.Positions = raw
	blank, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect blank holding: %v", err)
	}
	collector.Clock = clock.NewFake(asOf.Add(3 * time.Second))
	raw.raw = []reconcile.RawHolding{{Symbol: "AAPL", Quantity: "0"}}
	zero, err := collector.Collect(ctx)
	if err != nil {
		t.Fatalf("Collect zero holding: %v", err)
	}
	if len(blank.Holdings) != 1 || blank.Holdings[0].Quantity != "" ||
		len(zero.Holdings) != 1 || zero.Holdings[0].Quantity != "0" {
		t.Fatalf("collector did not preserve blank/zero evidence: blank=%+v zero=%+v", blank.Holdings, zero.Holdings)
	}
	if blank.Digest() == zero.Digest() {
		t.Errorf("snapshot digest aliased unreadable blank quantity with exact zero: %q", blank.Digest())
	}

	at := func(snap reconcile.Snapshot, when time.Time) reconcile.Snapshot {
		snap.AsOf = when
		snap.CompletedAt = when.Add(100 * time.Millisecond)
		return snap
	}
	for _, tc := range []struct {
		name          string
		first, second reconcile.Snapshot
	}{
		{name: "blank then zero", first: blank, second: zero},
		{name: "zero then blank", first: zero, second: blank},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stabiliser := &reconcile.Stabiliser{MinInterval: 2 * time.Second, Required: 2}
			first := stabiliser.Offer(at(tc.first, asOf))
			second := stabiliser.Offer(at(tc.second, asOf.Add(3*time.Second)))
			if !first.Counted || first.Stable || first.Streak != 1 {
				t.Fatalf("first mixed observation = %+v, want counted streak one", first)
			}
			if !second.Counted || second.Stable || second.Streak != 1 {
				t.Errorf("mixed unreadable/zero snapshots falsely corroborated: %+v", second)
			}
		})
	}

	for _, tc := range []struct {
		name string
		snap reconcile.Snapshot
	}{
		{name: "blank corroborates blank", snap: blank},
		{name: "zero corroborates zero", snap: zero},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stabiliser := &reconcile.Stabiliser{MinInterval: 2 * time.Second, Required: 2}
			stabiliser.Offer(at(tc.snap, asOf))
			second := stabiliser.Offer(at(tc.snap, asOf.Add(3*time.Second)))
			if !second.Counted || !second.Stable || second.Streak != 2 {
				t.Errorf("identical raw quantity evidence did not corroborate: %+v", second)
			}
		})
	}
}

func TestA110DuplicateIdentityCountsOnceAndCleanClearsEveryStreak(t *testing.T) {
	clk := clock.NewFake(asOf)
	tracker := a110Tracker(clk, nil, nil)
	duplicate := a110Quantity("AAPL", "10", "4")

	for observation := 1; observation <= 2; observation++ {
		out := a110Observe(t, tracker, a110QuantityDiff(duplicate, duplicate))
		if out.Failures != observation || out.Permanent {
			t.Fatalf("duplicate observation %d = %+v, want one streak increment", observation, out)
		}
		a110Advance(clk)
	}
	if out := a110Observe(t, tracker, reconcile.Diff{AccountRef: "acct-7", Matched: 1}); out.Failures != 0 {
		t.Fatalf("clean comparison left transient promotion evidence: %+v", out)
	}
	a110Advance(clk)
	if out := a110Observe(t, tracker, a110QuantityDiff(duplicate)); out.Failures != 1 || out.Permanent {
		t.Fatalf("post-clean same dispute did not restart at one: %+v", out)
	}
}

func TestA110OnlyFiniteCanonicalQuantitiesEarnPromotionEvidence(t *testing.T) {
	for _, invalid := range []reconcile.QuantityMismatch{
		a110Quantity("AAPL", "", "4"),
		a110Quantity("AAPL", "malformed", "4"),
		a110Quantity("AAPL", "NaN", "4"),
		a110Quantity("AAPL", "+Inf", "4"),
		a110Quantity("AAPL", "10", "-Inf"),
	} {
		t.Run(invalid.Local+"/"+invalid.Broker, func(t *testing.T) {
			clk := clock.NewFake(asOf)
			gate := noStaleGate(clk)
			tracker := a110Tracker(clk, gate, nil)
			for i := 0; i < reconcile.DefaultMaxFailures; i++ {
				out := a110Observe(t, tracker, a110QuantityDiff(invalid))
				if !out.Blocked {
					t.Fatalf("invalid quantity stopped being an ordinary fail-closed block: %+v", out)
				}
				a110Advance(clk)
			}
			if tracker.Permanent() || tracker.Failures() != 0 {
				t.Fatalf("unprovable quantity earned permanent streak: permanent=%v failures=%d", tracker.Permanent(), tracker.Failures())
			}
			if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil ||
				rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Fatalf("invalid quantity ordinary gate = %v, want mismatch block", rejected)
			}
		})
	}

	t.Run("valid sibling still earns its own streak", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		tracker := a110Tracker(clk, nil, nil)
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			a110Observe(t, tracker, a110QuantityDiff(
				a110Quantity("AAPL", "NaN", "4"),
				a110Quantity("MSFT", "7.00", "2"),
			))
			a110Advance(clk)
		}
		if !tracker.Permanent() {
			t.Fatal("valid sibling did not earn its independently repeated permanent promotion")
		}
	})

	t.Run("blank symbol remains fail-closed without identity evidence", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		gate := noStaleGate(clk)
		tracker := a110Tracker(clk, gate, nil)
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			out := a110Observe(t, tracker, a110QuantityDiff(a110Quantity("", "10", "4")))
			if !out.Blocked {
				t.Fatalf("blank-symbol mismatch stopped being fail-closed: %+v", out)
			}
			a110Advance(clk)
		}
		if tracker.Permanent() || tracker.Failures() != 0 {
			t.Fatalf("blank symbol earned permanent identity evidence: permanent=%v failures=%d", tracker.Permanent(), tracker.Failures())
		}
		if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Fatalf("blank-symbol mismatch must retain an account-safe gate: %v", rejected)
		}
	})

	t.Run("blank tracker account remains fail-closed without identity evidence", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		gate := noStaleGate(clk)
		// Do not use a110Tracker here: the tracker configuration, rather than
		// merely Diff.AccountRef, is the identity input that must be present.
		tracker := &reconcile.Tracker{
			Clock:       clk,
			Gate:        gate,
			MinInterval: 30 * time.Second,
			MaxFailures: reconcile.DefaultMaxFailures,
			AccountRef:  "",
		}
		for i := 0; i < reconcile.DefaultMaxFailures; i++ {
			out := a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
			if !out.Blocked {
				t.Fatalf("blank-account mismatch stopped being fail-closed: %+v", out)
			}
			a110Advance(clk)
		}
		if tracker.Permanent() || tracker.Failures() != 0 {
			t.Fatalf("blank tracker account earned permanent identity evidence: permanent=%v failures=%d", tracker.Permanent(), tracker.Failures())
		}
		if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Fatalf("blank-account mismatch must retain symbol fail-closed gate: %v", rejected)
		}
	})
}

func TestA110BlankSymbolJournalStateNeverRestoresAsPermanent(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"
	first := openJournalAt(t, path)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, first)

	out, _ := tracker.Observe(ctx, a110QuantityDiff(a110Quantity("", "10", "4")))
	if !out.Blocked {
		t.Fatal("blank-symbol ordinary mismatch must remain fail-closed even if it cannot be persisted at symbol scope")
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("blank-symbol ordinary mismatch gate = %v, want account-safe ordinary mismatch", rejected)
	}
	if tracker.Permanent() {
		t.Fatal("one unclassifiable blank-symbol mismatch must remain ordinary, not permanent")
	}
	if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("Tracker.EntryAllowed disagrees with fail-closed gate for blank symbol: %v", rejected)
	}
	states, err := first.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	for _, state := range states {
		if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
			t.Errorf("blank-symbol ordinary mismatch was persisted as account-wide permanent-shaped state: %+v", state)
		}
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close first journal: %v", err)
	}

	second := openJournalAt(t, path)
	t.Cleanup(func() { _ = second.Close() })
	restartedGate := noStaleGate(clk)
	restarted := a110Tracker(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if restarted.Permanent() {
		t.Fatalf("blank-symbol ordinary mismatch restored as operator-only permanent: %+v", restarted.Blocks())
	}
	if rejected := restartedGate.CheckEntry(); rejected != nil && rejected.Reason == execgw.ReasonReconcilePermanent {
		t.Fatalf("blank-symbol ordinary mismatch restored a permanent account gate: %v", rejected)
	}
}

func TestA110BlankSymbolGuardDoesNotStarveValidSiblingPersistence(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"
	first := openJournalAt(t, path)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, first)
	diff := a110QuantityDiff(
		a110Quantity("", "10", "4"),
		a110Quantity("AAPL", "7", "2"),
	)

	out, err := tracker.Observe(ctx, diff)
	if err == nil {
		t.Fatal("blank sibling must still report its non-durable scope error")
	}
	if !out.Blocked || tracker.EntryAllowed("us", "AAPL") == nil {
		t.Fatalf("mixed blank/AAPL observation was not fail-closed: outcome=%+v blocks=%+v", out, tracker.Blocks())
	}
	states, stateErr := first.ActiveReconcileStates(ctx)
	if stateErr != nil {
		t.Fatalf("ActiveReconcileStates: %v", stateErr)
	}
	var durableAAPL bool
	for _, state := range states {
		if state.Symbol == "AAPL" && state.Cause == journal.ReconcileCauseQuantityMismatch {
			durableAAPL = true
		}
		if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
			t.Errorf("blank sibling leaked into account-wide journal state: %+v", state)
		}
	}
	if !durableAAPL {
		t.Errorf("blank pending guard starved valid AAPL durable write: states=%+v", states)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close first journal: %v", err)
	}

	second := openJournalAt(t, path)
	t.Cleanup(func() { _ = second.Close() })
	restartedGate := noStaleGate(clk)
	restarted := a110Tracker(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if rejected := restarted.EntryAllowed("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("restart lost valid AAPL sibling block: %v", rejected)
	}
	if rejected := restartedGate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("restart gate lost valid AAPL sibling block: %v", rejected)
	}
}

func TestA110BlankSymbolPendingDoesNotStarveValidSiblingRelease(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	path := t.TempDir() + "/journal.db"
	first := openJournalAt(t, path)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, first)
	initial := a110QuantityDiff(
		a110Quantity("", "10", "4"),
		a110Quantity("AAPL", "7", "2"),
	)
	initial.AsOf = asOfAt(clk)

	out, err := tracker.Observe(ctx, initial)
	if err == nil {
		t.Fatal("blank sibling must report its non-durable scope error")
	}
	if !out.Blocked {
		t.Fatalf("initial blank/AAPL observation was not fail-closed: %+v", out)
	}
	states, stateErr := first.ActiveReconcileStates(ctx)
	if stateErr != nil {
		t.Fatalf("ActiveReconcileStates after enter: %v", stateErr)
	}
	if len(states) != 1 || states[0].Symbol != "AAPL" || states[0].Cause != journal.ReconcileCauseQuantityMismatch {
		t.Fatalf("initial observation did not durably enter only AAPL: %+v", states)
	}

	tracker.AdjustmentApplied(initial.AsOf, "AAPL")
	a110Advance(clk)
	out, err = tracker.Observe(ctx, cleanDiffAt(clk))
	if err == nil {
		t.Fatal("clean recheck must continue reporting the unpersistable blank pending guard")
	}
	if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "AAPL" {
		t.Errorf("blank pending guard starved the earned AAPL release: cleared=%+v err=%v", out.Cleared, err)
	}
	var blankStillPending, aaplStillBlocked bool
	for _, block := range tracker.Blocks() {
		if strings.TrimSpace(block.Symbol) == "" && !block.Permanent {
			blankStillPending = true
		}
		if block.Symbol == "AAPL" {
			aaplStillBlocked = true
		}
	}
	if !blankStillPending {
		t.Error("unpersistable blank guard disappeared after sibling release attempt")
	}
	if aaplStillBlocked {
		t.Errorf("released AAPL remained in tracker projection: %+v", tracker.Blocks())
	}
	for _, block := range gate.SymbolBlocks() {
		if block.Symbol == "AAPL" && block.Reason == execgw.ReasonReconcileMismatch {
			t.Errorf("released AAPL remained in the symbol gate: %+v", block)
		}
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("blank guard stopped providing its account-safe fail-closed gate: %v", rejected)
	}

	states, stateErr = first.ActiveReconcileStates(ctx)
	if stateErr != nil {
		t.Fatalf("ActiveReconcileStates after release: %v", stateErr)
	}
	if len(states) != 0 {
		t.Errorf("AAPL release was not durable: active=%+v", states)
	}
	history, historyErr := first.ReconcileStateHistory(ctx, "acct-7")
	if historyErr != nil {
		t.Fatalf("ReconcileStateHistory: %v", historyErr)
	}
	if len(history) != 1 || history[0].Symbol != "AAPL" || history[0].ReleaseCause != journal.ReconcileReleaseAdjustmentApplied {
		t.Errorf("durable AAPL release audit = %+v, want ADJUSTMENT_APPLIED", history)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close first journal: %v", err)
	}

	second := openJournalAt(t, path)
	t.Cleanup(func() { _ = second.Close() })
	restartedGate := noStaleGate(clk)
	restarted := a110Tracker(clk, restartedGate, second)
	if err := restarted.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if rejected := restarted.EntryAllowed("us", "AAPL"); rejected != nil {
		t.Errorf("restart resurrected durably released AAPL: %v blocks=%+v", rejected, restarted.Blocks())
	}
	if rejected := restartedGate.CheckEntryFor("us", "AAPL"); rejected != nil {
		t.Errorf("restart gate resurrected durably released AAPL: %v", rejected)
	}
}

func TestA110OperatorCanResolveKnownNondurableBlankSymbolGuard(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, journalStore)

	if _, err := tracker.Observe(ctx, a110QuantityDiff(a110Quantity("", "10", "4"))); err == nil {
		t.Fatal("blank-symbol guard must report that its ordinary scope is not journal-representable")
	}
	if tracker.EntryAllowed("us", "AAPL") == nil || gate.CheckEntry() == nil {
		t.Fatal("precondition: known-nondurable blank-symbol guard must be fail-closed")
	}
	if states, err := journalStore.ActiveReconcileStates(ctx); err != nil || len(states) != 0 {
		t.Fatalf("known-nondurable fixture has journal rows=%+v err=%v, want none", states, err)
	}
	if err := tracker.Resolve(ctx, "", "verified"); err == nil {
		t.Fatal("operator identity validation was bypassed")
	}
	if err := tracker.Resolve(ctx, "operator-7", ""); err == nil {
		t.Fatal("operator note validation was bypassed")
	}
	if err := tracker.Resolve(ctx, "operator-7", "verified unreadable broker scope out of band"); err != nil {
		t.Errorf("Resolve refused a known-nondurable guard with no journal row to release: %v", err)
	}
	if len(tracker.Blocks()) != 0 || tracker.Permanent() {
		t.Errorf("known-nondurable guard survived operator resolution: permanent=%v blocks=%+v", tracker.Permanent(), tracker.Blocks())
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Errorf("known-nondurable operator resolution left account gate closed: %v", rejected)
	}
	if rejected := tracker.EntryAllowed("us", "AAPL"); rejected != nil {
		t.Errorf("known-nondurable operator resolution left EntryAllowed blocked: %v", rejected)
	}
}

func TestA110ExactThresholdCreatesDurableOperatorOnlyPermanent(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	tracker := a110Tracker(clk, gate, journalStore)

	for i := 0; i < reconcile.DefaultMaxFailures; i++ {
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
	}
	if !tracker.Permanent() {
		t.Fatal("one exact dispute at threshold did not promote")
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("account gate = %v, want durable permanent rejection", rejected)
	}
	states, err := journalStore.ActiveReconcileStates(ctx)
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	var durable bool
	for _, state := range states {
		if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
			durable = true
		}
	}
	if !durable {
		t.Fatalf("journal states = %+v, want durable account-wide permanent row", states)
	}
	for _, block := range tracker.Blocks() {
		if block.Permanent && block.Release != reconcile.ReleaseOperatorOnly {
			t.Fatalf("permanent release = %q, want operator-only", block.Release)
		}
	}
}

func TestA110DurablePermanentEvidenceNamesTheEarningQuantityDispute(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	tracker := a110Tracker(clk, noStaleGate(clk), journalStore)

	for i := 0; i < reconcile.DefaultMaxFailures; i++ {
		a110Observe(t, tracker, mismatchDiff(" aapl ", "0010.00", "04.000"))
		a110Advance(clk)
	}
	states, err := journalStore.ActiveReconcileStates(context.Background())
	if err != nil {
		t.Fatalf("ActiveReconcileStates: %v", err)
	}
	var evidence string
	for _, state := range states {
		if state.AccountWide() && state.Cause == journal.ReconcileCauseQuantityMismatch {
			evidence = state.Evidence
			break
		}
	}
	if evidence == "" {
		t.Fatalf("journal states = %+v, want durable account permanent evidence", states)
	}
	for _, field := range []string{
		"threshold=3",
		"kind=quantity",
		"account=acct-7",
		"symbol=AAPL",
		"local=10",
		"broker=4",
	} {
		if !strings.Contains(evidence, field) {
			t.Errorf("permanent evidence %q does not identify earning field %q", evidence, field)
		}
	}
}

type a110EnterStore struct {
	reconcile.ReconcileStore
	gate                *execgw.EntryGate
	failPermanentCount  int
	failSymbolCount     int
	failSymbol          string
	permanentCalls      int
	symbolCalls         int
	gateOpenDuringEnter bool
}

type a110RawHoldingsReader struct {
	*fakeHoldings
	raw []reconcile.RawHolding
}

func (r *a110RawHoldingsReader) PositionsRaw(context.Context) ([]reconcile.RawHolding, error) {
	r.rec.log("holdings")
	return append([]reconcile.RawHolding(nil), r.raw...), nil
}

// a110DualPendingStore keeps both the ordinary AAPL row and its account-wide
// promotion non-durable through the threshold observation.  calls is the
// persistence-order ledger: totals cannot prove that the permanent proposal was
// attempted before a still-failing ordinary retry.
type a110DualPendingStore struct {
	reconcile.ReconcileStore
	calls                    []string
	remainingPermanentErrors int
}

// a110ThresholdHandoffStore fails the first account-wide promotion and records
// the earning evidence of every account attempt.  A permanent latch must already
// cover the account whenever either attempt reaches the durable store.
type a110ThresholdHandoffStore struct {
	reconcile.ReconcileStore
	gate              *execgw.EntryGate
	permanentRequests []journal.EnterReconcileRequest
	gateOpened        bool
	gateWrongReason   bool
}

// a110CommitThenTimeoutStore models the ambiguous result an I/O timeout can
// produce: the journal committed the account row, but the caller did not receive
// the acknowledgement.  A later observation must consult durable authority
// before treating that in-memory proposal as safe to withdraw.
type a110CommitThenTimeoutStore struct {
	reconcile.ReconcileStore
	returnedAmbiguousError bool
}

func (s *a110CommitThenTimeoutStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	state, entered, err := s.ReconcileStore.EnterReconcile(ctx, req)
	if err != nil {
		return state, entered, err
	}
	if req.Symbol == "" && !s.returnedAmbiguousError {
		s.returnedAmbiguousError = true
		return state, entered, errors.New("a110 injected timeout after permanent commit")
	}
	return state, entered, nil
}

// a110AuthorityReadFailureStore is the opposite ambiguity: the permanent write
// definitely did not commit, but the next observation cannot prove that because
// journal authority is unavailable.  Withdrawal must therefore fail closed.
type a110AuthorityReadFailureStore struct {
	reconcile.ReconcileStore
	authorityReads int
}

func (s *a110AuthorityReadFailureStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if req.Symbol == "" {
		return journal.ReconcileState{}, false, errors.New("a110 injected permanent enter failure before commit")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *a110AuthorityReadFailureStore) ActiveReconcileStates(context.Context) ([]journal.ReconcileState, error) {
	s.authorityReads++
	return nil, errors.New("a110 injected journal authority read failure")
}

// a110ContinuityBreakStore fails the pending authority read once and succeeds
// thereafter. The failed read cannot authorize withdrawal, but the observation
// still proved that the earning key was absent; a later reappearance must not
// revive the stale streak or retry its old permanent write.
type a110ContinuityBreakStore struct {
	reconcile.ReconcileStore
	permanentCalls     int
	authorityReads     int
	failAuthorityReads int
}

func (s *a110ContinuityBreakStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if req.Symbol == "" {
		s.permanentCalls++
		return journal.ReconcileState{}, false, errors.New("a110 injected no-commit permanent failure")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *a110ContinuityBreakStore) ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error) {
	s.authorityReads++
	if s.failAuthorityReads > 0 {
		s.failAuthorityReads--
		return nil, errors.New("a110 injected authority outage after continuity broke")
	}
	return s.ReconcileStore.ActiveReconcileStates(ctx)
}

func (s *a110ThresholdHandoffStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if req.Symbol != "" {
		return s.ReconcileStore.EnterReconcile(ctx, req)
	}
	rejected := s.gate.CheckEntry()
	if rejected == nil {
		s.gateOpened = true
	} else if rejected.Reason != execgw.ReasonReconcilePermanent {
		s.gateWrongReason = true
	}
	s.permanentRequests = append(s.permanentRequests, req)
	if len(s.permanentRequests) == 1 {
		return journal.ReconcileState{}, false, errors.New("a110 injected selected AAPL permanent failure")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *a110DualPendingStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	label := "symbol:" + req.Symbol
	if req.Symbol == "" {
		label = "account-permanent"
	}
	s.calls = append(s.calls, label)

	if req.Symbol == "" {
		if s.remainingPermanentErrors > 0 {
			s.remainingPermanentErrors--
			return journal.ReconcileState{}, false, errors.New("a110 injected permanent enter failure")
		}
		return s.ReconcileStore.EnterReconcile(ctx, req)
	}
	if req.Symbol == "AAPL" {
		return journal.ReconcileState{}, false, errors.New("a110 injected ordinary enter failure")
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func (s *a110EnterStore) EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error) {
	if s.gate != nil {
		if req.Symbol == "" {
			s.gateOpenDuringEnter = s.gateOpenDuringEnter || s.gate.CheckEntry() == nil
		} else if s.gate.CheckEntryFor("us", req.Symbol) == nil {
			s.gateOpenDuringEnter = true
		}
	}
	if req.Symbol == "" {
		s.permanentCalls++
		if s.failPermanentCount > 0 {
			s.failPermanentCount--
			return journal.ReconcileState{}, false, errors.New("a110 injected permanent enter failure")
		}
	}
	if req.Symbol == s.failSymbol {
		s.symbolCalls++
		if s.failSymbolCount > 0 {
			s.failSymbolCount--
			return journal.ReconcileState{}, false, errors.New("a110 injected ordinary enter failure")
		}
	}
	return s.ReconcileStore.EnterReconcile(ctx, req)
}

func TestA110FailedPermanentRetryRequiresTheEarningIdentity(t *testing.T) {
	t.Run("clean withdraws a non-durable account proposal", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		journalStore := openJournal(t)
		store := &a110EnterStore{ReconcileStore: journalStore, failPermanentCount: 1}
		gate := noStaleGate(clk)
		tracker := a110Tracker(clk, gate, store)
		for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
			a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
			a110Advance(clk)
		}
		a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
		if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
			t.Fatalf("failed permanent enter opened the account gate: %v", rejected)
		}
		if a110HasAccountPermanent(t, journalStore) {
			t.Fatal("failed permanent enter fabricated a durable row")
		}
		a110Advance(clk)
		a110Observe(t, tracker, reconcile.Diff{AccountRef: "acct-7", Matched: 1})
		if store.permanentCalls != 1 || a110HasAccountPermanent(t, journalStore) {
			t.Fatalf("clean retried stale permanent proposal: calls=%d durable=%v", store.permanentCalls, a110HasAccountPermanent(t, journalStore))
		}
		a110AssertNoTransientAccountPermanent(t, tracker, gate)
		if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Fatalf("clean must not release AAPL's ordinary fail-closed block: %v", rejected)
		}
	})

	t.Run("different identity withdraws a non-durable account proposal", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		journalStore := openJournal(t)
		store := &a110EnterStore{ReconcileStore: journalStore, failPermanentCount: 1}
		gate := noStaleGate(clk)
		tracker := a110Tracker(clk, gate, store)
		for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
			a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
			a110Advance(clk)
		}
		a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
		if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
			t.Fatalf("failed permanent enter opened the account gate: %v", rejected)
		}
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("MSFT", "7", "2"))
		if store.permanentCalls != 1 || a110HasAccountPermanent(t, journalStore) {
			t.Fatalf("different identity retried stale permanent proposal: calls=%d durable=%v", store.permanentCalls, a110HasAccountPermanent(t, journalStore))
		}
		a110AssertNoTransientAccountPermanent(t, tracker, gate)
		for _, symbol := range []string{"AAPL", "MSFT"} {
			if rejected := gate.CheckEntryFor("us", symbol); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
				t.Fatalf("different-key observation must retain ordinary %s block: %v", symbol, rejected)
			}
		}
	})

	t.Run("same identity retries while the entry gate stays closed", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		journalStore := openJournal(t)
		gate := noStaleGate(clk)
		store := &a110EnterStore{ReconcileStore: journalStore, gate: gate, failPermanentCount: 1}
		tracker := a110Tracker(clk, gate, store)
		for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
			a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
			a110Advance(clk)
		}
		a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
		if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
			t.Fatalf("failed permanent enter opened the account gate: %v", rejected)
		}
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		if store.permanentCalls != 2 || !a110HasAccountPermanent(t, journalStore) {
			t.Fatalf("same identity did not retry permanent proposal: calls=%d durable=%v", store.permanentCalls, a110HasAccountPermanent(t, journalStore))
		}
		if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
			t.Fatalf("durable same-key retry returned with account gate open or downgraded: %v", rejected)
		}
		if store.gateOpenDuringEnter {
			t.Fatal("journal enter observed an open gate after the mismatch was known")
		}
	})

	t.Run("earning identity retries when a new sibling joins the next diff", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		journalStore := openJournal(t)
		gate := noStaleGate(clk)
		store := &a110EnterStore{ReconcileStore: journalStore, gate: gate, failPermanentCount: 1}
		tracker := a110Tracker(clk, gate, store)
		for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
			a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
			a110Advance(clk)
		}
		a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
		a110Observe(t, tracker, a110QuantityDiff(
			a110Quantity("AAPL", "10", "4"),
			a110Quantity("MSFT", "7", "2"),
		))
		if store.permanentCalls != 2 || !a110HasAccountPermanent(t, journalStore) {
			t.Fatalf("earning identity was not retried in expanded diff: calls=%d durable=%v", store.permanentCalls, a110HasAccountPermanent(t, journalStore))
		}
		if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
			t.Fatalf("expanded-diff durable retry returned with account gate open or downgraded: %v", rejected)
		}
	})

	t.Run("ordinary pending retry remains independent", func(t *testing.T) {
		clk := clock.NewFake(asOf)
		journalStore := openJournal(t)
		gate := noStaleGate(clk)
		store := &a110EnterStore{
			ReconcileStore: journalStore, gate: gate, failSymbol: "AAPL", failSymbolCount: 1,
		}
		tracker := a110Tracker(clk, gate, store)
		a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
		if gate.CheckEntryFor("us", "AAPL") == nil {
			t.Fatal("ordinary persistence failure opened the gate")
		}
		a110Advance(clk)
		a110Observe(t, tracker, mismatchDiff("MSFT", "7", "2"))
		if store.symbolCalls != 2 {
			t.Fatalf("ordinary pending block was not retried: calls=%d", store.symbolCalls)
		}
		if store.gateOpenDuringEnter {
			t.Fatal("ordinary journal enter observed an open gate")
		}
		states, err := journalStore.ActiveReconcileStates(context.Background())
		if err != nil {
			t.Fatalf("ActiveReconcileStates: %v", err)
		}
		var durableAAPL bool
		for _, state := range states {
			if state.Symbol == "AAPL" && state.Cause == journal.ReconcileCauseQuantityMismatch {
				durableAAPL = true
			}
		}
		if !durableAAPL {
			t.Fatalf("ordinary pending retry did not create durable AAPL state: %+v", states)
		}
		if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Fatalf("ordinary pending retry must retain AAPL gate: %v", rejected)
		}
	})
}

func TestA110DualPendingPersistsPermanentBeforeRetryingOrdinary(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110DualPendingStore{
		ReconcileStore:           journalStore,
		remainingPermanentErrors: 1,
	}
	tracker := a110Tracker(clk, gate, store)
	diff := mismatchDiff("AAPL", "10", "4")

	// The ordinary row fails on observations one and two.  At the threshold,
	// the newly-added permanent proposal is ordered first but also fails, leaving
	// both proposals pending for the next observation.
	for observation := 1; observation <= reconcile.DefaultMaxFailures; observation++ {
		a110ObserveError(t, tracker, diff)
		a110Advance(clk)
	}
	if got, want := store.calls, []string{"symbol:AAPL", "symbol:AAPL", "account-permanent"}; !a110StringsEqual(got, want) {
		t.Fatalf("setup persistence order = %v, want %v", got, want)
	}
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("setup fabricated a durable permanent row")
	}

	start := len(store.calls)
	out, err := tracker.Observe(context.Background(), diff)
	if err == nil {
		t.Fatal("ordinary pending retry must still expose its injected persistence failure")
	}
	if got, want := store.calls[start:], []string{"account-permanent", "symbol:AAPL"}; !a110StringsEqual(got, want) {
		t.Fatalf("dual-pending persistence order = %v, want %v", got, want)
	}
	if !out.Permanent || !a110HasAccountPermanent(t, journalStore) {
		t.Fatalf("permanent was not durable before the later ordinary error: outcome=%+v durable=%v", out, a110HasAccountPermanent(t, journalStore))
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("ordinary retry failure downgraded/opened the durable permanent gate: %v", rejected)
	}
}

func TestA110ThresholdHandoffWithdrawsStaleWinnerWithoutStarvingSibling(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110ThresholdHandoffStore{ReconcileStore: journalStore, gate: gate}
	tracker := a110Tracker(clk, gate, store)
	both := a110QuantityDiff(
		a110Quantity("AAPL", "10", "4"),
		a110Quantity("MSFT", "7", "2"),
	)

	// Both exact keys reach two independently.
	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, both)
		a110Advance(clk)
	}
	// Both reach threshold together.  Deterministic selection chooses AAPL and
	// its durable account enter fails, leaving a fail-closed pending proposal.
	a110ObserveError(t, tracker, both)
	if len(store.permanentRequests) != 1 {
		t.Fatalf("first threshold made %d permanent attempts, want one deterministic AAPL proposal", len(store.permanentRequests))
	}
	if !a110EvidenceHas(store.permanentRequests[0].Evidence, "symbol=AAPL") {
		t.Errorf("first deterministic promotion = %+v, want AAPL earning evidence", store.permanentRequests)
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("failed AAPL promotion opened or downgraded account gate: %v", rejected)
	}

	// AAPL disappears, but MSFT already earned threshold in the preceding
	// comparison.  Withdraw the stale AAPL retry and immediately persist MSFT's
	// independent proposal; do not reset MSFT to one or retry AAPL's evidence.
	a110Advance(clk)
	a110Observe(t, tracker, mismatchDiff("MSFT", "7", "2"))
	if len(store.permanentRequests) != 2 {
		t.Fatalf("permanent attempts = %d (%+v), want failed AAPL then durable MSFT", len(store.permanentRequests), store.permanentRequests)
	}
	if !a110EvidenceHas(store.permanentRequests[1].Evidence,
		"threshold=3", "kind=quantity", "account=acct-7", "symbol=MSFT", "local=7", "broker=2") {
		t.Fatalf("second promotion evidence = %q, want independently earned MSFT threshold", store.permanentRequests[1].Evidence)
	}
	if strings.Contains(store.permanentRequests[1].Evidence, "symbol=AAPL") {
		t.Fatalf("stale AAPL earning key was retried after disappearance: %q", store.permanentRequests[1].Evidence)
	}
	if !a110HasAccountPermanent(t, journalStore) {
		t.Fatal("MSFT threshold handoff did not become durable")
	}
	if store.gateOpened || store.gateWrongReason {
		t.Fatalf("account gate during permanent enters: opened=%v wrong_reason=%v", store.gateOpened, store.gateWrongReason)
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("durable MSFT handoff returned with account gate open or downgraded: %v", rejected)
	}
}

func TestA110CommitThenTimeoutCannotBeWithdrawnByDirectCleanObservation(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110CommitThenTimeoutStore{ReconcileStore: journalStore}
	tracker := a110Tracker(clk, gate, store)

	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
	}
	a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
	if !a110HasAccountPermanent(t, journalStore) {
		t.Fatal("commit-then-timeout fixture did not durably commit the permanent row")
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("ambiguous permanent result opened the account gate: %v", rejected)
	}

	// No Refresh in between: Observe itself must reconcile the ambiguous result
	// against the journal before a clean diff can withdraw anything.
	a110Advance(clk)
	out, err := tracker.Observe(context.Background(), reconcile.Diff{AccountRef: "acct-7", Matched: 1})
	if err != nil {
		t.Fatalf("clean Observe failed to recover committed authority: %v", err)
	}
	if !out.Permanent || !tracker.Permanent() || !a110HasAccountPermanent(t, journalStore) {
		t.Fatalf("clean observation withdrew a committed permanent: outcome=%+v tracker=%v durable=%v", out, tracker.Permanent(), a110HasAccountPermanent(t, journalStore))
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("clean observation opened/downgraded committed permanent gate: %v", rejected)
	}
}

func TestA110AuthorityReadFailureKeepsUncommittedPendingPermanentFailClosed(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110AuthorityReadFailureStore{ReconcileStore: journalStore}
	tracker := a110Tracker(clk, gate, store)

	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
	}
	a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("fixture unexpectedly committed the failed permanent proposal")
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("failed permanent enter opened the pending account gate: %v", rejected)
	}

	a110Advance(clk)
	out, err := tracker.Observe(context.Background(), reconcile.Diff{AccountRef: "acct-7", Matched: 1})
	if err == nil {
		t.Fatalf("clean Observe succeeded despite unavailable journal authority: %+v", out)
	}
	if store.authorityReads != 1 {
		t.Fatalf("journal authority reads = %d, want one before pending withdrawal", store.authorityReads)
	}
	if !out.Permanent || !tracker.Permanent() {
		t.Fatalf("authority-read failure withdrew pending permanent: outcome=%+v tracker=%v", out, tracker.Permanent())
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("authority-read failure opened/downgraded pending account gate: %v", rejected)
	}
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("authority-read failure test fabricated a durable permanent row")
	}
}

func TestA110AuthorityOutageDoesNotEraseObservedContinuityBreak(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110ContinuityBreakStore{
		ReconcileStore: journalStore, failAuthorityReads: 1,
	}
	tracker := a110Tracker(clk, gate, store)

	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
	}
	a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
	if store.permanentCalls != 1 {
		t.Fatalf("threshold permanent calls = %d, want one failed no-commit attempt", store.permanentCalls)
	}

	// This authoritative comparison proves A absent, even though journal
	// authority cannot yet prove whether A's previous permanent write committed.
	a110Advance(clk)
	out, err := tracker.Observe(context.Background(), mismatchDiff("MSFT", "7", "2"))
	if err == nil {
		t.Fatalf("key-loss observation succeeded despite authority outage: %+v", out)
	}
	if !out.Permanent || gate.CheckEntry() == nil {
		t.Fatalf("authority outage did not retain pending permanent fail-closed: outcome=%+v gate=%v", out, gate.CheckEntry())
	}
	if store.permanentCalls != 1 || store.authorityReads != 1 {
		t.Fatalf("after key loss: permanent_calls=%d authority_reads=%d, want 1/1", store.permanentCalls, store.authorityReads)
	}

	// A reappears after the broken comparison. A successful authority check may
	// now withdraw the known-uncommitted pending row, but the old A streak cannot
	// be resurrected and the stale permanent write cannot be retried.
	a110Advance(clk)
	out, err = tracker.Observe(context.Background(), mismatchDiff("AAPL", "10", "4"))
	if err != nil {
		t.Errorf("fresh A observation failed after authority recovered: %v", err)
	}
	if store.permanentCalls != 1 {
		t.Errorf("stale A permanent write retried after observed continuity break: calls=%d", store.permanentCalls)
	}
	if store.authorityReads != 2 {
		t.Errorf("fresh A did not re-check pending journal authority: reads=%d", store.authorityReads)
	}
	if out.Permanent || tracker.Permanent() || out.Failures != 1 {
		t.Errorf("reappearing A inherited stale evidence: outcome=%+v tracker_permanent=%v", out, tracker.Permanent())
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Errorf("stale account permanent gate survived recovered no-commit authority: %v", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("fresh A ordinary guard missing after continuity break: %v", rejected)
	}
}

func TestA110RefreshDoesNotManufacturePermanentFromContinuityBrokenPendingProposal(t *testing.T) {
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110ContinuityBreakStore{
		ReconcileStore: journalStore, failAuthorityReads: 1,
	}
	tracker := a110Tracker(clk, gate, store)

	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
	}
	a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("fixture unexpectedly committed the failed permanent proposal")
	}

	// The changed comparison proves the earning AAPL key is absent. Journal
	// authority is temporarily unavailable, so the account pending proposal and
	// gate must stay fail-closed until a later successful authority read.
	a110Advance(clk)
	out, err := tracker.Observe(context.Background(), mismatchDiff("MSFT", "7", "2"))
	if err == nil {
		t.Fatalf("continuity-break Observe succeeded despite authority outage: %+v", out)
	}
	if !out.Permanent || !tracker.Permanent() {
		t.Fatalf("authority outage failed to retain the pending account guard: outcome=%+v tracker=%v", out, tracker.Permanent())
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("authority outage account gate = %v, want pending permanent", rejected)
	}
	if store.authorityReads != 1 {
		t.Fatalf("authority reads after continuity break = %d, want one failed read", store.authorityReads)
	}

	// Refresh is the next successful authority read. With no durable account
	// permanent row, it must discard the known-nondurable proposal and preserve
	// the observed continuity break instead of reconstructing a threshold streak.
	if err := tracker.Refresh(context.Background()); err != nil {
		t.Fatalf("Refresh after authority recovery: %v", err)
	}
	if store.authorityReads != 2 {
		t.Fatalf("authority reads after Refresh = %d, want failed Observe plus successful Refresh", store.authorityReads)
	}
	if tracker.Failures() != 0 {
		t.Errorf("Refresh manufactured compatibility failures from pending proposal: got %d want 0", tracker.Failures())
	}
	a110AssertNoTransientAccountPermanent(t, tracker, gate)
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("Refresh manufactured a durable account permanent row")
	}
	if rejected := tracker.EntryAllowed("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("Refresh lost the durable ordinary AAPL guard: %v", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("Refresh reopened AAPL despite its durable ordinary mismatch: %v", rejected)
	}
}

func TestA110AuthorityOutageStillProjectsCurrentDifferentOrdinaryMismatch(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110ContinuityBreakStore{
		ReconcileStore: journalStore, failAuthorityReads: 1,
	}
	tracker := a110Tracker(clk, gate, store)

	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiff("AAPL", "10", "4"))
		a110Advance(clk)
	}
	a110ObserveError(t, tracker, mismatchDiff("AAPL", "10", "4"))
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("fixture unexpectedly committed the failed AAPL permanent proposal")
	}

	// The pending account proposal masks every symbol at the public gate, but it
	// must not prevent this authoritative comparison's new MSFT ordinary guard
	// from entering the tracker and symbol projection before the authority read.
	a110Advance(clk)
	out, err := tracker.Observe(ctx, mismatchDiff("MSFT", "7", "2"))
	if err == nil {
		t.Fatalf("different-key Observe succeeded despite injected authority outage: %+v", out)
	}
	var trackerHasMSFT bool
	for _, block := range tracker.Blocks() {
		if block.Symbol == "MSFT" && block.Cause == journal.ReconcileCauseQuantityMismatch && !block.Permanent {
			trackerHasMSFT = true
		}
	}
	if !trackerHasMSFT {
		t.Errorf("authority outage omitted the current MSFT ordinary tracker block: %+v", tracker.Blocks())
	}
	var gateHasMSFT bool
	for _, block := range gate.SymbolBlocks() {
		if block.Symbol == "MSFT" && block.Reason == execgw.ReasonReconcileMismatch {
			gateHasMSFT = true
		}
	}
	if !gateHasMSFT {
		t.Errorf("authority outage omitted the current MSFT symbol gate: %+v", gate.SymbolBlocks())
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Errorf("authority outage did not retain the wider pending account guard: %v", rejected)
	}

	// The next successful authority refresh proves the failed account proposal
	// was not durable. It may withdraw that stale proposal, but the current MSFT
	// ordinary pending guard must survive beside the durable AAPL ordinary row.
	if err := tracker.Refresh(ctx); err != nil {
		t.Fatalf("Refresh after authority recovery: %v", err)
	}
	if tracker.Permanent() {
		t.Errorf("Refresh manufactured/retained a non-durable account permanent: %+v", tracker.Blocks())
	}
	var trackerHasAAPL bool
	trackerHasMSFT = false
	for _, block := range tracker.Blocks() {
		switch block.Symbol {
		case "AAPL":
			trackerHasAAPL = true
		case "MSFT":
			trackerHasMSFT = true
		}
	}
	if !trackerHasAAPL || !trackerHasMSFT {
		t.Errorf("Refresh lost ordinary guards: AAPL=%v MSFT=%v blocks=%+v", trackerHasAAPL, trackerHasMSFT, tracker.Blocks())
	}
	if rejected := gate.CheckEntry(); rejected != nil {
		t.Errorf("successful no-durable authority retained stale account proposal: %v", rejected)
	}
	for _, symbol := range []string{"AAPL", "MSFT"} {
		if rejected := gate.CheckEntryFor("us", symbol); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
			t.Errorf("Refresh reopened current ordinary %s guard: %v", symbol, rejected)
		}
	}
	states, stateErr := journalStore.ActiveReconcileStates(ctx)
	if stateErr != nil {
		t.Fatalf("ActiveReconcileStates: %v", stateErr)
	}
	for _, state := range states {
		if state.AccountWide() {
			t.Errorf("no-commit fixture gained an account-wide durable row: %+v", state)
		}
		if state.Symbol == "MSFT" {
			t.Errorf("authority-error MSFT block was unexpectedly durable instead of pending: %+v", state)
		}
	}
}

func TestA110AuthorityOutageStillRefutesUsableAdjustmentCredit(t *testing.T) {
	ctx := context.Background()
	clk := clock.NewFake(asOf)
	journalStore := openJournal(t)
	gate := noStaleGate(clk)
	store := &a110ContinuityBreakStore{
		ReconcileStore: journalStore, failAuthorityReads: 1,
	}
	tracker := a110Tracker(clk, gate, store)

	for i := 0; i < reconcile.DefaultMaxFailures-1; i++ {
		a110Observe(t, tracker, mismatchDiffAt(clk, "AAPL", "10", "4"))
		a110Advance(clk)
	}
	threshold := mismatchDiffAt(clk, "AAPL", "10", "4")
	// This is the production driver order: an adjustment computed from the
	// threshold comparison is credited before that same comparison is observed.
	// Equal as-of cannot spend it; only a strictly later re-read may answer it.
	tracker.AdjustmentApplied(threshold.AsOf, "AAPL")
	a110ObserveError(t, tracker, threshold)
	if a110HasAccountPermanent(t, journalStore) {
		t.Fatal("fixture unexpectedly committed the failed permanent proposal")
	}

	// The next authoritative comparison is later, changes the earning tuple and
	// still disputes AAPL. Even though journal authority fails and the wider gate
	// must remain closed, this observation has already refuted the usable credit.
	a110Advance(clk)
	changed := mismatchDiffAt(clk, "AAPL", "11", "4")
	out, err := tracker.Observe(ctx, changed)
	if err == nil {
		t.Fatalf("changed-tuple Observe succeeded despite injected authority outage: %+v", out)
	}
	if rejected := gate.CheckEntry(); rejected == nil || rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("authority outage did not retain pending account guard: %v", rejected)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil {
		t.Fatal("authority outage reopened AAPL while the changed tuple still disputed it")
	}

	if err := tracker.Refresh(ctx); err != nil {
		t.Fatalf("Refresh after authority recovery: %v", err)
	}
	if tracker.Permanent() {
		t.Fatalf("Refresh retained known-nondurable permanent proposal: %+v", tracker.Blocks())
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Fatalf("Refresh lost durable AAPL ordinary guard: %v", rejected)
	}

	// The old credit was answered by the changed-tuple disagreement, not by this
	// later clean read. It cannot now release AAPL without a new adjustment.
	a110Advance(clk)
	out, err = tracker.Observe(ctx, cleanDiffAt(clk))
	if err != nil {
		t.Fatalf("clean Observe after authority recovery: %v", err)
	}
	if len(out.Cleared) != 0 {
		t.Errorf("stale credit unsafely released AAPL after authority recovery: %+v", out.Cleared)
	}
	if len(out.AwaitingAdjustment) != 1 || out.AwaitingAdjustment[0].Symbol != "AAPL" {
		t.Errorf("clean recheck = awaiting %+v, want AAPL to require a new adjustment", out.AwaitingAdjustment)
	}
	states, stateErr := journalStore.ActiveReconcileStates(ctx)
	if stateErr != nil {
		t.Fatalf("ActiveReconcileStates: %v", stateErr)
	}
	if len(states) != 1 || states[0].Symbol != "AAPL" || states[0].Cause != journal.ReconcileCauseQuantityMismatch {
		t.Errorf("stale credit changed durable AAPL guard: %+v", states)
	}
	if rejected := gate.CheckEntryFor("us", "AAPL"); rejected == nil || rejected.Reason != execgw.ReasonReconcileMismatch {
		t.Errorf("stale credit reopened AAPL gate: %v", rejected)
	}
}

func a110EvidenceHas(evidence string, fields ...string) bool {
	for _, field := range fields {
		if !strings.Contains(evidence, field) {
			return false
		}
	}
	return true
}

func a110StringsEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
