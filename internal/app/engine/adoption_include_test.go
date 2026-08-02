package engine_test

// adoption_include_test.go covers per-symbol designation (change
// console-adoption-controls task 2.1/2.2).
//
//	opt-in adopts     an included symbol is adopted with the global switch off,
//	                  through the same gates — nothing about freshness or the
//	                  Stabiliser is relaxed for it.
//	exclusion wins    include+exclude is not adopted, and the alert names the
//	                  exclusion regardless of the global switch.
//	the alert is true a designated symbol whose cycle failed is reported as
//	                  tried-and-failed, never as "adoption is off" (review
//	                  round 1, P2-6).
//	audited           the include list is part of the startup settings audit —
//	                  a per-symbol designation turns a specific real holding
//	                  into a sell-capable managed position (P1-2).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

func includeOnly(symbols ...string) config.Adoption {
	return config.Adoption{DefaultStopPct: 0.05, IncludeSymbols: symbols}
}

// TestAnIncludedSymbolIsAdoptedWithTheSwitchOff is the opt-in itself: the
// designated holding is adopted, the undesignated one beside it stays an
// unmanaged finding.
func TestAnIncludedSymbolIsAdoptedWithTheSwitchOff(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = includeOnly("005930")
	})
	h.holds("005930", "10", "55000.0000", 70000)
	h.holds("000660", "3", "120000.0000", 131500)

	cycle := h.cycle()
	if cycle.Adopted != 1 {
		t.Fatalf("adopted = %d, want exactly the designated symbol", cycle.Adopted)
	}
	if !h.position("005930").Adopted() {
		t.Error("the designated holding was not adopted")
	}
	if h.position("000660").Adopted() {
		t.Error("an undesignated holding was adopted with the global switch off")
	}
	if cycle.Unmanaged != 1 {
		t.Errorf("unmanaged = %d, want 1: the undesignated holding is still a finding", cycle.Unmanaged)
	}
}

func TestUSIncludedSymbolFoldsAdoptsAndOpensExitT0(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = includeOnly("AAPL")
	})
	h.holdsMarket("us", "AAPL", "2", "180.0000", 200, "USD")

	cycle := h.cycle()
	if cycle.Err != nil || cycle.Folded != 1 || cycle.Adopted != 1 || cycle.Unmanaged != 0 {
		t.Fatalf("cycle=%+v", cycle)
	}
	p := h.positionMarket("us", "AAPL")
	provenance, eligibility := positionpolicy.ClassifyProvenance(p.EntryDecisionID, p.AdoptionID)
	if !p.Adopted() || !p.ExitEligible() || provenance != positionpolicy.ProvenanceExternalAdoption ||
		eligibility != positionpolicy.EligibilityExternalLifecycle {
		t.Fatalf("position=%+v provenance=%s eligibility=%s", p, provenance, eligibility)
	}
	adoption, err := h.journal.AdoptionOf(t.Context(), p.ID)
	if err != nil || adoption.ObservedPrice != "200" || adoption.SyntheticStop != "190" {
		t.Fatalf("adoption=%+v err=%v", adoption, err)
	}
	exitState, err := h.journal.ExitState(t.Context(), p.ID)
	if err != nil || exitState.EntryPrice != "200" || exitState.HighWater != "200" ||
		exitState.InitialStop != "190" || exitState.Baseline != "190" {
		t.Fatalf("exit state=%+v err=%v", exitState, err)
	}
}

func TestUSAdoptionRefusesWrongOrEmptyQuoteCurrency(t *testing.T) {
	for _, currency := range []string{"KRW", ""} {
		t.Run("currency="+currency, func(t *testing.T) {
			h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
				o.Adoption = includeOnly("AAPL")
			})
			h.holdsMarket("us", "AAPL", "2", "180", 200, currency)
			cycle := h.cycle()
			if cycle.Adopted != 0 || cycle.Deferred != 1 {
				t.Fatalf("cycle=%+v", cycle)
			}
			p := h.positionMarket("us", "AAPL")
			if p.Adopted() {
				t.Fatalf("wrong-currency quote was persisted: %+v", p)
			}
		})
	}
}

func TestUSIncludedSymbolWaitsUnderAccountWidePermanentReconcile(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = includeOnly("AAPL")
	})
	for i := 0; i < reconcile.DefaultMaxFailures; i++ {
		out, err := h.tracker.Observe(t.Context(), reconcile.Diff{
			AccountRef: reconcileAccount,
			Quantities: []reconcile.QuantityMismatch{{Symbol: "OTHER", Local: "1", Broker: "2"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if i+1 == reconcile.DefaultMaxFailures && !out.Permanent {
			t.Fatalf("final outcome=%+v", out)
		}
		h.clk.Advance(reconcile.DefaultReconcileInterval)
	}
	if rejected := h.tracker.EntryAllowed("us", "AAPL"); rejected == nil ||
		rejected.Reason != execgw.ReasonReconcilePermanent {
		t.Fatalf("precondition block=%v", rejected)
	}
	h.holdsMarket("us", "AAPL", "2", "180", 200, "USD")

	cycle := h.cycle()
	if cycle.Folded != 1 || cycle.Adopted != 0 || cycle.Unmanaged != 0 || h.prices.calls != 0 {
		t.Fatalf("cycle=%+v price calls=%d", cycle, h.prices.calls)
	}
	p := h.positionMarket("us", "AAPL")
	if p.Adopted() {
		t.Fatalf("blocked position adopted=%+v", p)
	}
}

// TestIncludeDoesNotBypassExclusion: exclude wins, and the alert names the
// exclusion even though the global switch is off — the reason a holding is
// unprotected is the list the operator wrote, not the switch.
func TestIncludeDoesNotBypassExclusion(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		a := includeOnly("005930")
		a.ExcludeSymbols = []string{"005930"}
		o.Adoption = a
	})
	h.holds("005930", "10", "55000", 70000)

	cycle := h.cycle()
	if cycle.Adopted != 0 {
		t.Fatalf("adopted = %d, want 0: exclusion wins over designation", cycle.Adopted)
	}
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("no unmanaged alert for the excluded-and-designated symbol")
	}
	if !strings.Contains(alert.Body, "exclude_symbols") {
		t.Errorf("the alert must name the exclusion regardless of the global switch: %q", alert.Body)
	}
	if h.prices.calls != 0 {
		t.Errorf("price reads = %d; an excluded symbol was never a candidate", h.prices.calls)
	}
}

// TestAFailedIncludeCycleSaysTriedNotOff is the why-matrix row review round 1
// P2-6 exists for: the designated symbol's quote is missing, the adoption is
// deferred, and the operator must read "designated, tried, failed this cycle" —
// not "adoption is off", which would send them to the wrong setting.
func TestAFailedIncludeCycleSaysTriedNotOff(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = includeOnly("005930")
	})
	// A holding with no quote on the tape: a candidate whose observation never
	// arrives is deferred, not adopted.
	h.holdings.items = append(h.holdings.items, reconcile.RawHolding{
		Symbol: "005930", Market: "kr", Quantity: "10", AveragePrice: "55000",
	})

	cycle := h.cycle()
	if cycle.Adopted != 0 {
		t.Fatalf("adopted = %d with no observable price", cycle.Adopted)
	}
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("a designated holding that could not be adopted must still be reported unprotected")
	}
	if strings.Contains(alert.Body, "adoption is off") {
		t.Errorf("the alert says adoption is off about a designated symbol that was tried: %q",
			alert.Body)
	}
	if !strings.Contains(alert.Body, "designated") {
		t.Errorf("the alert must say the symbol was designated and the cycle failed: %q", alert.Body)
	}
}

// TestIncludeListIsAudited is §0.5 for the designation list (review round 1,
// P1-2): the startup settings audit carries include_symbols like it carries the
// toggle, the fraction and the exclusions.
func TestIncludeListIsAudited(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	srv, _ := interlockServer(t, "123-45")

	writeAdoptionConfig(t, dir, config.Adoption{})
	if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("first start: %v", err)
	}

	writeAdoptionConfig(t, dir, config.Adoption{
		DefaultStopPct: 0.05, IncludeSymbols: []string{"005930", "000660"},
	})
	if _, err := openGateEngine(t, dir, srv, matchedGuardian()); err != nil {
		t.Fatalf("second start: %v", err)
	}

	entries := readAudit(t, filepath.Join(dir, "audit.log"))
	includes := lastEntryFor(entries, "engine.adoption.include_symbols", audit.ActionAdoptionSetting)
	if includes == nil {
		t.Fatalf("no include_symbols audit entry; a per-symbol designation turns a real holding "+
			"into a sell-capable managed position and cannot be an unrecorded setting. entries = %+v",
			entries)
	}
	if includes.New != "000660,005930" && includes.New != "005930,000660" {
		t.Errorf("include entry new = %q, want the designated list", includes.New)
	}
}

// TestIncludeSurvivesTheConfigRoundTrip pins the wiring end of task 1.1: a
// config file carrying the list reaches the engine's Adoption options intact.
func TestIncludeSurvivesTheConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	body := `{"schema_version":5,"trading":{},
	  "engine":{"adoption":{"default_stop_pct":0.05,"include_symbols":["005930"]}}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	svc := config.NewService(filepath.Join(dir, "config.json"))
	cfg, err := svc.Load(t.Context())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	var block config.Adoption
	data, _ := json.Marshal(cfg.Engine.Adoption)
	if err := json.Unmarshal(data, &block); err != nil {
		t.Fatal(err)
	}
	if !block.Included("005930") {
		t.Errorf("the include list did not survive the load: %+v", block)
	}
}

// The remaining three why-matrix rows (독립 검증 finding 1): each body must say
// its own actionable fact.

func TestARejectedBlockAlertNamesTheRefusal(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{Rejected: "stop fraction out of band"}
	})
	h.holds("005930", "10", "55000", 70000)
	h.cycle()
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("no unmanaged alert for a rejected block")
	}
	if !strings.Contains(alert.Body, "refused") || !strings.Contains(alert.Body, "stop fraction out of band") {
		t.Errorf("the alert must name the refusal and its reason: %q", alert.Body)
	}
}

func TestAnEnabledFailedCycleSaysOnAndFailed(t *testing.T) {
	h := newDriverHarness(t, nil) // enabled=true harness default
	h.holdings.items = append(h.holdings.items, reconcile.RawHolding{
		Symbol: "005930", Market: "kr", Quantity: "10", AveragePrice: "55000",
	}) // no quote: the candidate defers
	h.cycle()
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("no unmanaged alert for an enabled-but-failed cycle")
	}
	if !strings.Contains(alert.Body, "adoption is on but") {
		t.Errorf("the alert must say adoption was on and the cycle failed: %q", alert.Body)
	}
}

func TestTheDefaultRowSaysOffAndUndesignated(t *testing.T) {
	h := newDriverHarness(t, func(o *engine.ReconcileDriverOptions) {
		o.Adoption = config.Adoption{}
	})
	h.holds("005930", "10", "55000", 70000)
	h.cycle()
	alert, ok := h.alerts.first(obs.EventExitPositionUnmanaged)
	if !ok {
		t.Fatal("no unmanaged alert with adoption off")
	}
	if !strings.Contains(alert.Body, "adoption is off and the symbol is not designated") {
		t.Errorf("the default row must say off-and-undesignated: %q", alert.Body)
	}
}
