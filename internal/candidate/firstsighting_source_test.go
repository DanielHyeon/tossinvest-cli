package candidate

// firstsighting_source_test.go is section 2: a first sighting's position is only a
// measurement when the source that produced it could say what the position meant.
//
// # The failure this closes
//
// Cooling is thirty minutes and staleness ten, so any gap longer than forty minutes
// expires the whole watchlist. Overnight, over a weekend, after a process stop, the
// next scan promotes the entire panel at once — every row of every reading is a
// candidate, `Promote` has no threshold — and each symbol records its *current*
// rank as the one we first saw it at. The top of the list is then marked "we were
// late to this" for the rest of the session, and the reason is that our process
// started.
//
// # Why the refusal is not a ratio
//
// The obvious repair is to notice the batch: "a scan that promoted more than N% of
// the panel is a bulk re-promotion, so its positions are suspect". That N has no
// source. It would be a policy number chosen by whoever wrote the line, applied to
// every market, unverified — which is the exact disease this whole change is
// treating, reintroduced by its own cure. D6 forbids it at the front door and a
// ratio here is the back one.
//
// What is used instead is a fact the source declares about itself: did it hold a
// previous reading of this list. On the first reading of a session it did not, and
// "we cannot tell whether this symbol had just joined" is then not a heuristic
// about the batch — it is the source saying it does not know.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// scriptedSource returns one reading per call, in order, and repeats the last one.
//
// It is deliberately not a copy of candidatesrc's previous-reading bookkeeping.
// What the scan-level tests here assert is that whatever a source reports about its
// reading reaches the verdict; that the real adapters compute those reports
// correctly is candidatesrc's own test.
type scriptedSource struct {
	id       SourceID
	readings [][]Row
	calls    int
}

func (s *scriptedSource) ID() SourceID { return s.id }

func (s *scriptedSource) Read(context.Context, string) (Reading, error) {
	i := s.calls
	if i >= len(s.readings) {
		i = len(s.readings) - 1
	}
	s.calls++
	return Reading{Rows: s.readings[i]}, nil
}

// unqualifiedRow is a ranked row from a source that has no previous reading: the
// first reading of a session, which is the one that used to stamp the panel.
func unqualifiedRow(symbol string, rank, total int) Row {
	return Row{
		Symbol: symbol, Rank: rank, RankTotal: total, RankRequested: total,
		NewlyListed:  NewlyListedUnknown(),
		TradingValue: "1000000", TradingVolume: "500", Price: "70000",
	}
}

// enteredRow is a ranked row the source reports as absent from its previous reading.
func enteredRow(symbol string, rank, total int) Row {
	r := unqualifiedRow(symbol, rank, total)
	r.NewlyListed = NewlyListedYes()
	return r
}

// heldRow is a ranked row the source reports as present in its previous reading.
func heldRow(symbol string, rank, total int) Row {
	r := unqualifiedRow(symbol, rank, total)
	r.NewlyListed = NewlyListedNo()
	return r
}

// seenLateOf is one symbol's seen_late verdict out of an assessment.
func seenLateOf(t *testing.T, verdicts []Verdict, symbol string) (VetoState, Sighting) {
	t.Helper()
	for _, v := range verdicts {
		if v.Summary.Symbol == symbol {
			return v.Chase.SeenLate, v.Sighting
		}
	}
	t.Fatalf("no verdict for %s among %d", symbol, len(verdicts))
	return VetoState{}, Sighting{}
}

// seenLateThresholds is a threshold high enough that nothing measured could raise
// seen_late, so that every raised verdict in these tests would have to come from the
// measurement rather than from the number.
//
// It is not a proposal. This change sets no threshold — Chase.Passed() stays
// unreachable — and the value here exists only so that AssessSeenLate gets past
// THRESHOLD_ABSENT and the reason under test is the one that appears.
func seenLateThresholds() VetoThresholds {
	return VetoThresholds{
		SeenLatePercentilePct: "80",
		NearHighDistancePct:   DefaultNearHighThresholdPct,
	}
}

// TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate is task 2.1.
func TestAPositionFromASourcesFirstReadingCannotAnswerSeenLate(t *testing.T) {
	c := aCandidate(t0)
	first := storedFirstRank(4, 100, t0, SourceOfficialTradingValue)
	first.NewlyListed = NewlyListedUnknown()

	got := MeasureFirstSighting(c, first, nil)
	if got.Measured {
		t.Fatalf("a position from a reading nobody could qualify was measured at %s%%",
			got.PercentilePct)
	}
	if got.Reason() != VetoNewEntrantUnknown {
		t.Errorf("reason = %q, want %q", got.Reason(), VetoNewEntrantUnknown)
	}
	// It still reports what it read, so a screen can show how far up the list the
	// position was even though nothing may be concluded from it.
	if got.Rank != 4 || got.RankTotal != 100 {
		t.Errorf("the refused sighting reports %d of %d, want 4 of 100", got.Rank, got.RankTotal)
	}
	if got.PercentilePct != "" {
		t.Errorf("percentile = %q, want empty — absent is not a number", got.PercentilePct)
	}
	if state := AssessSeenLate(got, seenLateThresholds()); state.Clear() || state.Dangerous() {
		t.Errorf("seen_late = %v for an unqualified position; it has to be unmeasured, and "+
			"unmeasured is not a pass (D10)", state)
	}
}

// TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps is task 2.1's naming clause.
//
// veto.go keeps many reasons rather than one because a screen full of a single
// reason has to tell an operator what to fix. The four existing "no position"
// reasons are all about this store — retention took the row, the panel has no
// ranking source, the column predates the schema. This one is about the source, and
// the action it implies is different: wait for the second reading.
func TestTheUnqualifiedReasonIsNotOneOfTheStoresOwnGaps(t *testing.T) {
	c := aCandidate(t0)
	unqualified := storedFirstRank(4, 100, t0, SourceOfficialTradingValue)
	unqualified.NewlyListed = NewlyListedUnknown()

	got := map[string]VetoUnmeasured{
		"a source with no previous reading": MeasureFirstSighting(c, unqualified, nil).Reason(),
		"nothing recorded at all":           MeasureFirstSighting(c, FirstRank{}, nil).Reason(),
		"nothing ranked": MeasureFirstSighting(c, FirstRank{}, []Observation{
			obs("005930", t0, SourceOfficialPrices, Reported{Price: "70000"}),
		}).Reason(),
		"a ranked row with no stored column": MeasureFirstSighting(c, FirstRank{},
			[]Observation{rankedObs(t0, SourceOfficialTradingValue, 4, 100)}).Reason(),
	}
	seen := map[VetoUnmeasured]string{}
	for what, reason := range got {
		if reason == "" {
			t.Errorf("%s produced an empty reason", what)
			continue
		}
		if other, clash := seen[reason]; clash {
			t.Errorf("%q is the reason for both %q and %q; the two send an operator to two "+
				"different places and a screen full of one must not read like the other",
				reason, what, other)
		}
		seen[reason] = what
	}
	if got["a source with no previous reading"] != VetoNewEntrantUnknown {
		t.Errorf("the unqualified reason = %q, want %q",
			got["a source with no previous reading"], VetoNewEntrantUnknown)
	}
}

// TestASessionStartDoesNotStampThePanelAsSeenLate is task 2.2, end to end.
//
// The first scan promotes everything the source returned — that is the design, and
// Promote has no threshold — and every one of those candidates records its current
// position. Before this change each of them was then measurable, so the top of the
// list was raised on the first scan of every session. Now the whole panel is
// unmeasured under one named reason, and none of it is counted as having passed.
//
// # Corrected 2026-07-28: which reason, and for how long
//
// This asserted NEW_ENTRANT_UNKNOWN on the first scan, which meant the unqualified
// position had been stored — and first_rank is written once, so it stayed stored for
// the rest of each candidate's life. The refusal was correct and permanent, and
// permanent was not the intention: veto.go told the operator to wait for the second
// reading, and waiting did nothing.
//
// recordFirsts now declines to store a position whose reading carries neither
// qualifier, so the first scan leaves NO_FIRST_RANK — the recoverable state — and
// the second scan's qualified reading becomes the first sighting for every one of
// these candidates, still inside the ±TTL identity window. What the session start
// costs is one tick, not a session.
func TestASessionStartDoesNotStampThePanelAsSeenLate(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	src := &scriptedSource{
		id: SourceOfficialTradingValue,
		readings: [][]Row{
			// The first reading of the session: the source has nothing to compare
			// against, so it qualifies nothing. Three symbols at the very top of a
			// hundred-row list — the ones that would be stamped.
			{
				unqualifiedRow("005930", 1, 100),
				unqualifiedRow("000660", 2, 100),
				unqualifiedRow("035720", 3, 100),
			},
			// The second reading: now it can answer. Two were there before; one has
			// just joined, at the top.
			{
				heldRow("005930", 1, 100),
				heldRow("000660", 2, 100),
				enteredRow("068270", 3, 100),
			},
		},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	// The counter that says so, asserted here because nothing else asserts it.
	//
	// Added 2026-07-28. The hold *behaviour* was pinned by the rest of this test —
	// removing the write left the store's columns filled and the reasons wrong — but
	// removing only `result.FirstRanksHeld++` left the whole suite green, so the
	// count could stop counting while the refusal kept working. It is the number
	// `tossctl candidate scan` prints to say a one-shot command cannot qualify
	// anything, and a scan that held three positions and reported none of them is
	// indistinguishable from a scan that had nothing to hold.
	if res.Scan.FirstRanksHeld != 3 {
		t.Errorf("the session's first scan held %d positions, want 3 — one for each candidate "+
			"whose only reading was the one the source could not qualify", res.Scan.FirstRanksHeld)
	}
	if res.Scan.FirstRanks != 0 {
		t.Errorf("%d first ranks were stored on the session's first scan", res.Scan.FirstRanks)
	}
	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: t0, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if len(verdicts) != 3 {
		t.Fatalf("%d candidates after the first scan, want 3", len(verdicts))
	}
	for _, v := range verdicts {
		state := v.Chase.SeenLate
		if state.Measured() {
			t.Errorf("%s: seen_late is %v after the session's first scan. Its position is the "+
				"one the panel happened to be in when the process came up, and the source had "+
				"no previous reading to say whether the symbol had just joined",
				v.Summary.Symbol, state)
		}
		// NO_FIRST_RANK rather than NEW_ENTRANT_UNKNOWN, and the difference is the
		// whole of the repair: nothing was stored, so nothing is fixed in place. The
		// reason an operator sees is the recoverable one, and the next tick recovers
		// it.
		if state.Reason() != VetoNoFirstRank {
			t.Errorf("%s: reason = %q, want %q", v.Summary.Symbol, state.Reason(),
				VetoNoFirstRank)
		}
	}
	// And nothing was written, which is the fact behind the reason above.
	stored, _, err := s.FirstRank(ctx, MarketKR, "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if stored.Recorded() {
		t.Errorf("the session's first scan stored %d of %d from a reading it could not "+
			"qualify. first_rank is written once, so that position would answer for the "+
			"rest of this candidate's life and nothing could fill the columns beside it",
			stored.Rank, stored.Total)
	}
	tally, _, _ := TallyVerdicts(verdicts)
	if tally.Passed != 0 {
		t.Errorf("passed = %d after a scan in which nothing could be measured", tally.Passed)
	}
	if tally.NotMeasured[VetoSeenLate] != 3 {
		t.Errorf("seen_late not-measured = %d of %d candidates",
			tally.NotMeasured[VetoSeenLate], tally.Total)
	}

	// The second scan, one tick later. The symbol that has just joined the list is
	// measurable, because the source held a previous reading and said so.
	clk.Advance(20 * time.Second)
	at := clk.Now()
	opts.Schedule = NewSchedule(nil)
	res, err = Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	// And the counter goes back to zero, which is the other half of what it means.
	// A watch loop reports a held count on its first turn and none afterwards; a
	// count that stayed at three would say the panel never becomes measurable.
	if res.Scan.FirstRanksHeld != 0 {
		t.Errorf("the second scan held %d positions; its reading carries both qualifiers for "+
			"every row, so there is nothing left to hold", res.Scan.FirstRanksHeld)
	}
	if res.Scan.FirstRanks != 3 {
		t.Errorf("%d first ranks stored on the second scan, want 3 — the three symbols this "+
			"reading listed, each getting the position it was first qualified at",
			res.Scan.FirstRanks)
	}
	verdicts, err = Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: at, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("second Assess: %v", err)
	}
	state, sighting := seenLateOf(t, verdicts, "068270")
	if !sighting.Measured {
		t.Fatalf("the symbol that joined on the second reading is still unmeasured (%s); the "+
			"source answered `yes` about it and that is exactly the answer seen_late needs",
			sighting.Reason())
	}
	if !sighting.NewlyListed.Yes() {
		t.Errorf("the new entrant's stored fact = %s, want yes", sighting.NewlyListed)
	}
	if sighting.PercentilePct != "97" {
		t.Errorf("percentile = %q, want 97 (3rd of 100)", sighting.PercentilePct)
	}
	if !state.Dangerous() {
		t.Errorf("seen_late = %v for a symbol first seen 3rd of a hundred; measurable now "+
			"means measured, not cleared", state)
	}

	// And the two that were already on the list are measurable too, from the second
	// reading — because the first one stored nothing for them.
	//
	// Corrected 2026-07-28. This used to assert that they stayed NEW_ENTRANT_UNKNOWN
	// for good, on the grounds that "a later reading does not retroactively qualify"
	// the stored position. That is still true of a stored position; what changed is
	// that no position was stored. The second reading's is the first one this
	// candidate has, it is inside the identity window, and its `no` is an honest
	// answer about the list — so the candidate is measurable one tick after the
	// process came up rather than never.
	//
	// Note what is *not* claimed: that this is the position the symbol had when the
	// process started. It is the position at the second reading, and the reason that
	// is acceptable is the same reason the ±TTL window exists — inside it, a reading
	// is the one this life was first seen in.
	held, heldSighting := seenLateOf(t, verdicts, "005930")
	if !heldSighting.Measured {
		t.Fatalf("a candidate that was on the list at process start is unmeasured (%s) one "+
			"tick later; the first scan stored nothing for it, so the second reading's "+
			"qualified position is the first one it has", heldSighting.Reason())
	}
	if !heldSighting.NewlyListed.No() {
		t.Errorf("its stored fact = %s, want no — it was in the source's previous reading",
			heldSighting.NewlyListed)
	}
	if heldSighting.PercentilePct != "99" {
		t.Errorf("percentile = %q, want 99 (1st of 100)", heldSighting.PercentilePct)
	}
	if !held.Dangerous() {
		t.Errorf("seen_late = %v for a candidate first seen at the top of a hundred-row "+
			"list", held)
	}
}

// TestARePromotionAfterExpiryIsQualifiedByTheReadingThatSawTheSymbolReturn is task
// 2.3.
//
// This is the other half of the session-start rule and it is why the refusal is not
// simply "the first scan is untrustworthy". A candidate that expires and crosses
// again gets a new first_seen_at and a new first rank, and that reading can be
// qualified — so the refusal costs the session's first tick and not the re-entries.
//
// # Corrected 2026-07-28: which answer a re-promotion actually gets
//
// This test used to script the return as `no` and say in its failure message that
// "the source has been reading this list for forty minutes", which nothing
// established: the scripted source hands back whatever the fixture spells, and
// design D3's sentence — that a re-promotion is qualified because the source still
// holds a previous reading — was an assumption about the adapter rather than a
// property of it.
//
// It is now a property, and it points the other way. candidatesrc's memory expires
// at previousReadingTTL, so a forty-minute gap in *reading* leaves the source with
// no answer at all. What is left is the case that reaches here in production: the
// symbol leaves the list, the source keeps reading it every tick, the candidate
// cools and expires with nothing to promote it, and then the symbol comes back — and
// against a fresh previous reading that is `yes`. A candidate cannot be cooled while
// a source it has is still listing it, so `no` on a re-promotion would require a
// forty-minute gap in the reading, which is exactly the gap the memory no longer
// survives.
//
// `yes` is the stronger answer anyway: it says the symbol joined the list between
// two consecutive readings and we saw where it landed.
func TestARePromotionAfterExpiryIsQualifiedByTheReadingThatSawTheSymbolReturn(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	// Life one, promoted from a reading the source could qualify.
	src := &scriptedSource{
		id:       SourceOfficialTradingValue,
		readings: [][]Row{{heldRow("005930", 60, 100)}},
	}
	opts := cycleOpts(MarketKR, src)
	opts.Thresholds = seenLateThresholds()
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}

	// Cooling plus staleness, and then some. The candidate expires, which is what
	// clears the stored first rank and starts a new life on the next crossing.
	clk.Advance(DefaultCoolingTTL + DefaultStalenessTTL + time.Minute)
	reborn := clk.Now()
	src.readings = [][]Row{{enteredRow("005930", 4, 100)}}
	opts.Schedule = NewSchedule(nil)
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("second Cycle: %v", err)
	}

	verdicts, err := Assess(ctx, s, AssessOptions{
		Market: MarketKR, At: reborn, Thresholds: seenLateThresholds(),
	})
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	state, sighting := seenLateOf(t, verdicts, "005930")
	if !sighting.Measured {
		t.Fatalf("the re-promotion is unmeasured (%s); the reading that saw the symbol come "+
			"back answered about its own previous reading, which is an answer",
			sighting.Reason())
	}
	if !sighting.NewlyListed.Yes() {
		t.Errorf("the re-promotion's stored fact = %s, want yes", sighting.NewlyListed)
	}
	if sighting.Rank != 4 || sighting.PercentilePct != "96" {
		t.Errorf("the new life's sighting = %d of %d at %s%%, want 4 of 100 at 96%% — the new "+
			"life's own position, not the dead life's 60th",
			sighting.Rank, sighting.RankTotal, sighting.PercentilePct)
	}
	if !state.Dangerous() {
		t.Errorf("seen_late = %v; a candidate whose life began at 4th of a hundred is one we "+
			"arrived late to", state)
	}
}

// TestTheRefusalCountsNoRatioOfThePanel is task 2.4, and it is a test about what is
// *not* in the source.
//
// The alternative to asking the source was to notice the batch: "one scan promoted
// N% of the panel, so treat its positions as bulk". Every version of that rule needs
// an N, and an N here would be a policy number with no source, no market and no
// verification state — applied to a measurement, which is worse than applying one to
// a threshold because nothing downstream would ever show it.
//
// The guard is written against the source rather than against behaviour because the
// behaviour of a ratio rule and of this rule agree on the session-start case, which
// is the case anybody would test.
func TestTheRefusalCountsNoRatioOfThePanel(t *testing.T) {
	const path = "veto.go"
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, src, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	var body *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "MeasureFirstSighting" {
			body = fn
		}
	}
	if body == nil {
		t.Fatal("MeasureFirstSighting is not in veto.go; this guard is reading nothing")
	}

	// A bar and a share are the two shapes the alternative takes, and both need
	// arithmetic this function must not contain: a quantity compared against a
	// number somebody chose, or a count divided by another count.
	//
	// Zero is the one number allowed, and it is not a bar. `len(observations) > 0`
	// asks whether there is anything at all — emptiness, which is a property of the
	// slice rather than a level somebody picked — and the same zero appears as the
	// "no position recorded" test. Every other literal is a decision.
	ast.Inspect(body, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			if v.Kind == token.FLOAT || (v.Kind == token.INT && v.Value != "0") {
				t.Errorf("MeasureFirstSighting names the numeric literal %s. The two refusals "+
					"it makes are the source's own answers — did you hold a previous reading, "+
					"how many rows did you ask for — and any other number here is a bar "+
					"somebody picked, applied to every market, with no verification state",
					v.Value)
			}
		case *ast.BinaryExpr:
			if v.Op == token.QUO || v.Op == token.REM {
				t.Errorf("MeasureFirstSighting divides (%s); a share of the panel is the ratio "+
					"rule this change refuses to reintroduce while curing the disease it comes "+
					"from", v.Op)
			}
		}
		return true
	})

	// And the reason it is refused is written down where somebody adding one would
	// be reading.
	if !strings.Contains(string(src), "ratio") && !strings.Contains(string(src), "N%") {
		t.Error("veto.go no longer says why the batch is not counted; the argument against " +
			"the ratio is the thing that stops it coming back")
	}
}
