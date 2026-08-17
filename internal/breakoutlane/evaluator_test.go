package breakoutlane

import "testing"

func fixtureConfig(t *testing.T, _ string) V1Config {
	t.Helper()
	input := V1ConfigInput{Version: "v1", TickMinor: 1, OpeningRangeMinutes: 15, BreakoutBufferPPM: 100_000, RetestTolerancePPM: 100_000, TimeoutKR: 8, TimeoutUS: 10, RVOLMinPPM: 1_500_000, UpperWickRangeMaxPPM: 350_000, MaxQuoteAgeMS: 10, MaxSpreadPPM: 100_000, MaxEntryDriftPPM: 100_000}
	input.Digest = V1ConfigDigest(input)
	c, err := NewV1Config(input)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func fixtureBar(t *testing.T, seq, high, low, close, rvol, wick uint64) ClosedBar {
	t.Helper()
	b, err := NewClosedBar(ClosedBarInput{Sequence: seq, Revision: 1, IntervalMS: 60_000, ID: "bar-" + u(seq), SessionID: "KRX:2026-08-18", RegularSession: true, Closed: true, HighMinor: high, LowMinor: low, CloseMinor: close, RVOLPPM: rvol, UpperWickRangePPM: wick})
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func fixtureQuote(t *testing.T, source, received uint64) QuoteSeal {
	t.Helper()
	i := QuoteSealInput{BidMinor: 100, AskMinor: 101, LastMinor: 101, Currency: "USD", SourceObservedAtMS: source, ReceivedAtMS: received}
	i.Digest = QuoteSealDigest(i)
	q, err := NewQuoteSeal(i)
	if err != nil {
		t.Fatal(err)
	}
	return q
}
func fixtureFX(t *testing.T, until uint64) FXSeal {
	t.Helper()
	i := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 2, AsOfMS: 1, FreshUntilMS: until}
	i.Digest = FXSealDigest(i)
	f, err := NewFXSeal(i)
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func fixtureInput(t *testing.T) EvidenceInput {
	t.Helper()
	bars := make([]ClosedBar, 0, 18)
	for i := uint64(1); i <= 15; i++ {
		bars = append(bars, fixtureBar(t, i, 100, 90, 95, 1_000_000, 100_000))
	}
	bars = append(bars, fixtureBar(t, 16, 111, 99, 101, 1_500_000, 350_000), fixtureBar(t, 17, 100, 98, 99, 1_000_000, 100_000), fixtureBar(t, 18, 102, 99, 101, 1_000_000, 100_000))
	return EvidenceInput{Market: MarketKR, Symbol: "005930", SessionID: "KRX:2026-08-18", CalendarVersion: "krx-calendar-v1", LaneID: KRLaneID, LaneVersion: "v1", Config: fixtureConfig(t, "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), Bars: bars, ATRMinor: 10, EvaluatedAtMS: 10, Quote: fixtureQuote(t, 5, 6), FX: fixtureFX(t, 10), Sizing: SizingInput{ProposedEntryMinor: 101, StopMinor: 95, TargetMinor: 140, RiskBudgetAccountMinor: 1_000, NotionalCapAccountMinor: 10_000, FinalCap: 10}}
}
func snapshot(t *testing.T, input EvidenceInput) EvidenceSnapshot {
	t.Helper()
	s, err := NewEvidenceSnapshot(input)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestEvaluatorDerivesOnlyCompleteClosedBarPathToProposal(t *testing.T) {
	d := Evaluate(snapshot(t, fixtureInput(t)), nil)
	if d.Phase() != "PROPOSED" || d.Refusal() != RefusalNone || d.ConfigDigest() != fixtureInput(t).Config.Digest() || d.FinalQuantity() == 0 || d.CandidateQuantity() < d.FinalQuantity() {
		t.Fatalf("decision=%+v", d)
	}
	p := d.Provenance()
	if len(p.Transitions) != 7 || !p.RVOLAdmission || !p.RVOLAt1200000 || p.RVOLAt2000000 || p.RVOLAt2500000 {
		t.Fatalf("provenance=%+v", p)
	}
}

func TestAdversarialForgedDecisionCannotBypassEvaluator(t *testing.T) {
	forged := Decision{phase: "PROPOSED", proposalID: "forged", seal: "forged"}
	got := Evaluate(snapshot(t, fixtureInput(t)), &forged)
	if got.Refusal() != RefusalLineageSealMismatch || got.ProposalID() != "" {
		t.Fatalf("forged prior bypassed evaluator: %+v", got)
	}
}
func TestAdversarialFirstTouchMissingRangeAndBadBarCannotPropose(t *testing.T) {
	input := fixtureInput(t)
	input.Bars[15] = fixtureBar(t, 16, 111, 99, 100, 1_500_000, 350_000)
	d := Evaluate(snapshot(t, input), nil)
	if d.Phase() != "RANGE_LOCKED" || d.Refusal() != RefusalFirstTouch || d.ProposalID() != "" {
		t.Fatalf("first touch=%+v", d)
	}
	input = fixtureInput(t)
	input.Bars = input.Bars[:14]
	if _, err := NewEvidenceSnapshot(input); err == nil {
		t.Fatal("missing opening range accepted")
	}
	input = fixtureInput(t)
	bad := input.Bars[16].value
	bad.Closed = false
	input.Bars[16] = ClosedBar{value: bad}
	if _, err := NewEvidenceSnapshot(input); err == nil {
		t.Fatal("unfinished bar snapshot accepted")
	}
}
func TestAdversarialCorrectionReplaysPreTerminalAndPreservesProposed(t *testing.T) {
	input := fixtureInput(t)
	input.Bars = input.Bars[:16]
	before := Evaluate(snapshot(t, input), nil)
	if before.Phase() != "RETEST_WAIT" {
		t.Fatal(before.Phase())
	}
	corrected := input
	bar := corrected.Bars[15].value
	bar.Revision = 2
	bar.CloseMinor = 100
	corrected.Bars[15] = ClosedBar{value: bar}
	after := Evaluate(snapshot(t, corrected), &before)
	if after.Phase() != "RANGE_LOCKED" || after.SnapshotDigest() == before.SnapshotDigest() {
		t.Fatalf("correction did not replay=%+v", after)
	}
	proposed := Evaluate(snapshot(t, fixtureInput(t)), nil)
	changed := fixtureInput(t)
	bar = changed.Bars[15].value
	bar.Revision = 2
	bar.CloseMinor = 100
	changed.Bars[15] = ClosedBar{value: bar}
	held := Evaluate(snapshot(t, changed), &proposed)
	if held.Phase() != "PROPOSED" || held.ProposalID() != proposed.ProposalID() || held.Diagnostic() != DiagnosticCorrectionAfterProposal {
		t.Fatalf("proposed correction=%+v", held)
	}
}

func TestEvaluatorInvalidationTimeoutAndQuoteVetoEdges(t *testing.T) {
	invalid := fixtureInput(t)
	bad := invalid.Bars[17].value
	bad.CloseMinor, bad.LowMinor = 89, 89
	invalid.Bars[17] = ClosedBar{value: bad}
	if d := Evaluate(snapshot(t, invalid), nil); d.Phase() != "INVALIDATED" {
		t.Fatalf("invalidation=%+v", d)
	}
	timeout := fixtureInput(t)
	timeout.Bars = timeout.Bars[:16]
	for seq := uint64(17); seq <= 24; seq++ {
		timeout.Bars = append(timeout.Bars, fixtureBar(t, seq, 106, 101, 105, 1_000_000, 100_000))
	}
	if d := Evaluate(snapshot(t, timeout), nil); d.Phase() != "TIMED_OUT" {
		t.Fatalf("timeout=%+v", d)
	}
	wide := fixtureInput(t)
	q := wide.Quote.value
	q.BidMinor = 1
	q.Digest = QuoteSealDigest(q)
	wide.Quote = QuoteSeal{value: q}
	if d := Evaluate(snapshot(t, wide), nil); d.Refusal() != RefusalSpreadTooWide {
		t.Fatalf("wide=%q", d.Refusal())
	}
	drift := fixtureInput(t)
	drift.Sizing.ProposedEntryMinor = 50
	if d := Evaluate(snapshot(t, drift), nil); d.Refusal() != RefusalEntryDriftExceeded {
		t.Fatalf("drift=%q", d.Refusal())
	}
}

func TestTerminalCorrectionsRetainTerminalAuthority(t *testing.T) {
	invalid := fixtureInput(t)
	bar := invalid.Bars[17].value
	bar.CloseMinor, bar.LowMinor = 89, 89
	invalid.Bars[17] = ClosedBar{value: bar}
	first := Evaluate(snapshot(t, invalid), nil)
	changed := invalid
	bar = changed.Bars[15].value
	bar.Revision = 2
	changed.Bars[15] = ClosedBar{value: bar}
	held := Evaluate(snapshot(t, changed), &first)
	if held.Phase() != "INVALIDATED" || held.Diagnostic() != DiagnosticCorrectionAfterProposal || held.ProposalID() != "" {
		t.Fatalf("invalid terminal correction=%+v", held)
	}
	timeout := fixtureInput(t)
	timeout.Bars = timeout.Bars[:16]
	for seq := uint64(17); seq <= 24; seq++ {
		timeout.Bars = append(timeout.Bars, fixtureBar(t, seq, 106, 101, 105, 1_000_000, 100_000))
	}
	first = Evaluate(snapshot(t, timeout), nil)
	changed = timeout
	bar = changed.Bars[15].value
	bar.Revision = 2
	changed.Bars[15] = ClosedBar{value: bar}
	held = Evaluate(snapshot(t, changed), &first)
	if held.Phase() != "TIMED_OUT" || held.Diagnostic() != DiagnosticCorrectionAfterProposal || held.ProposalID() != "" {
		t.Fatalf("timeout terminal correction=%+v", held)
	}
}

func TestArithmeticAndConservativeSizingOverflowEdges(t *testing.T) {
	if _, overflow := mulDivCeil(maxUint64, maxUint64, 1); !overflow {
		t.Fatal("ceil overflow accepted")
	}
	if _, refusal := mulDivFloor(maxUint64, maxUint64, 1); refusal != RefusalSizingOverflow {
		t.Fatalf("floor overflow=%q", refusal)
	}
	if _, overflow := checkedAdd(maxUint64, 1); !overflow {
		t.Fatal("addition overflow accepted")
	}
	if BreakoutBufferMinor(1, maxUint64) == 0 || BreakoutCloseQualifies(100, 100, maxUint64, fixtureConfig(t, "cfg")) {
		t.Fatal("overflow breakout buffer accepted")
	}
	if RetestQualifies(maxUint64, 0, 1, fixtureConfig(t, "cfg")) {
		t.Fatal("overflow retest accepted")
	}
	if got := (Decision{}).SetupID(); got != "" {
		t.Fatal(got)
	}
}

func TestFXConversionAndSnapshotValidationFailureEdges(t *testing.T) {
	if _, refusal := convertCapacity(1, FXSeal{}, 1); refusal != RefusalFXInvalidRate {
		t.Fatalf("missing FX=%q", refusal)
	}
	future := fixtureFX(t, 10)
	f := future.value
	f.AsOfMS = 2
	f.Digest = FXSealDigest(f)
	future = FXSeal{value: f}
	if _, refusal := convertCapacity(1, future, 1); refusal != RefusalFXStale {
		t.Fatalf("future capacity=%q", refusal)
	}
	overflowFX := fixtureFX(t, 10)
	overflowFields := overflowFX.value
	overflowFields.RateNum = 2
	overflowFields.Digest = FXSealDigest(overflowFields)
	overflowFX = FXSeal{value: overflowFields}
	if _, refusal := convertCost(maxUint64, overflowFX, 1); refusal != RefusalSizingOverflow {
		t.Fatalf("cost overflow=%q", refusal)
	}
	i := fixtureInput(t)
	i.LaneID = "wrong"
	if _, err := NewEvidenceSnapshot(i); err == nil {
		t.Fatal("wrong descriptor snapshot accepted")
	}
	if got := size(fixtureInput(t).Sizing, fixtureInput(t).Quote, FXSeal{}, 10); got.Refusal != RefusalFXMissing {
		t.Fatalf("missing sizing FX=%+v", got)
	}
}
func TestAdversarialQuoteFXAndSizingRefusals(t *testing.T) {
	cases := []struct {
		name   string
		mut    func(*EvidenceInput)
		want   RefusalCode
		strict bool
	}{
		{"missing ask", func(i *EvidenceInput) { q := i.Quote.value; q.AskMinor = 0; i.Quote = QuoteSeal{value: q} }, RefusalEvidenceInvalid, true},
		{"zero final cap", func(i *EvidenceInput) { i.Sizing.FinalCap = 0 }, RefusalZeroQuantity, false},
		{"cost inclusive target", func(i *EvidenceInput) {
			i.Sizing.TargetMinor = 102
			i.Sizing.EntrySlippageMinor = 1
			i.Sizing.ExitSlippageMinor = 1
			i.Sizing.RoundTripCostAccountMinor = 1
		}, RefusalNonProtectiveTarget, false},
		{"stale fx", func(i *EvidenceInput) { i.FX = fixtureFX(t, 9) }, RefusalFXStale, false},
		{"currency mismatch", func(i *EvidenceInput) {
			q := i.Quote.value
			q.Currency = "KRW"
			q.Digest = QuoteSealDigest(q)
			i.Quote = QuoteSeal{value: q}
		}, RefusalFXCurrencyMismatch, false},
		{"quote stale", func(i *EvidenceInput) { i.Quote = fixtureQuote(t, 0, 6) }, RefusalQuoteStale, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			i := fixtureInput(t)
			tc.mut(&i)
			if tc.name == "quote stale" {
				i.EvaluatedAtMS = 11
				i.FX = fixtureFX(t, 11)
			}
			if tc.strict {
				if _, err := NewEvidenceSnapshot(i); err == nil {
					t.Fatal("strict snapshot accepted")
				}
				return
			}
			d := Evaluate(snapshot(t, i), nil)
			if d.Refusal() != tc.want {
				t.Fatalf("got %q want %q", d.Refusal(), tc.want)
			}
		})
	}
}
func TestSetupIDKnownVectorAndNineFieldSensitivity(t *testing.T) {
	i := fixtureInput(t)
	i.Config = V1Config{value: V1ConfigInput{Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}}
	first := i.Bars[0].value
	first.ID = "bar-20260818-0900"
	i.Bars[0] = ClosedBar{value: first}
	last := i.Bars[14].value
	last.ID = "bar-20260818-0914"
	i.Bars[14] = ClosedBar{value: last}
	if got := setupID(i); got != "sha256:d2d5e2da006b841e45a1d2991624c5316b520a3900bb780653c79f455eeef04c" {
		t.Fatalf("vector=%s", got)
	}
	base := setupID(i)
	mutators := []func(*EvidenceInput){func(v *EvidenceInput) { v.Market = MarketUS }, func(v *EvidenceInput) { v.Symbol = "005931" }, func(v *EvidenceInput) { v.SessionID = "KRX:next" }, func(v *EvidenceInput) { v.CalendarVersion = "next" }, func(v *EvidenceInput) { b := v.Bars[0].value; b.ID = "first-x"; v.Bars[0] = ClosedBar{value: b} }, func(v *EvidenceInput) { b := v.Bars[14].value; b.ID = "last-x"; v.Bars[14] = ClosedBar{value: b} }, func(v *EvidenceInput) { v.LaneID = "other" }, func(v *EvidenceInput) { v.LaneVersion = "v2" }, func(v *EvidenceInput) {
		v.Config = V1Config{value: V1ConfigInput{Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}}
	}}
	for n, m := range mutators {
		copy := i
		m(&copy)
		if setupID(copy) == base {
			t.Fatalf("field %d insensitive", n)
		}
	}
}
func TestAdversarialInvalidUTF8ConfigAndFXDirectionScale(t *testing.T) {
	i := fixtureInput(t)
	i.Symbol = string([]byte{0xff})
	s := EvidenceSnapshot{value: i, digest: "forged"}
	if d := Evaluate(s, nil); d.Refusal() != RefusalEvidenceInvalid {
		t.Fatal(d.Refusal())
	}
	raw := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXInstrumentToAccount, RateNum: 2, RateDen: 1, Scale: 6, AsOfMS: 1, FreshUntilMS: 10}
	f, err := NewFXSeal(raw)
	if err != nil || f.value.Direction != FXAccountToInstrument || f.value.RateNum != 1 || f.value.RateDen != 2 || f.value.Scale != 6 {
		t.Fatalf("fx=%+v err=%v", f, err)
	}
}

func TestAdversarialQuoteAndFXSealTimeOrderDigestCurrencyDirectionScale(t *testing.T) {
	quote := QuoteSealInput{BidMinor: 100, AskMinor: 101, LastMinor: 101, Currency: "USD", SourceObservedAtMS: 1, ReceivedAtMS: 1, Digest: "wrong"}
	if _, err := NewQuoteSeal(quote); err == nil {
		t.Fatal("bad quote digest accepted")
	}
	quote.Digest = QuoteSealDigest(quote)
	sealedQuote, err := NewQuoteSeal(quote)
	if err != nil {
		t.Fatal(err)
	}
	i := fixtureInput(t)
	i.Quote = sealedQuote
	i.EvaluatedAtMS = 0
	sealed, err := NewEvidenceSnapshot(i)
	if err != nil || Evaluate(sealed, nil).Refusal() != RefusalQuoteStale {
		t.Fatalf("future quote was not typed at Evaluate: err=%v refusal=%q", err, Evaluate(sealed, nil).Refusal())
	}
	badOrder := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 2, AsOfMS: 11, FreshUntilMS: 10}
	badOrder.Digest = FXSealDigest(badOrder)
	if _, err := NewFXSeal(badOrder); err == nil {
		t.Fatal("bad FX order accepted")
	}
	badScale := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 0, AsOfMS: 1, FreshUntilMS: 10}
	badScale.Digest = FXSealDigest(badScale)
	if _, err := NewFXSeal(badScale); err == nil {
		t.Fatal("zero FX scale accepted")
	}
	badDigest := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 2, AsOfMS: 1, FreshUntilMS: 10, Digest: "bad"}
	if _, err := NewFXSeal(badDigest); err == nil {
		t.Fatal("bad FX digest accepted")
	}
	future := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 2, AsOfMS: 11, FreshUntilMS: 20}
	future.Digest = FXSealDigest(future)
	sealedFX, err := NewFXSeal(future)
	if err != nil {
		t.Fatal(err)
	}
	i = fixtureInput(t)
	i.FX = sealedFX
	sealed, err = NewEvidenceSnapshot(i)
	if err != nil || Evaluate(sealed, nil).Refusal() != RefusalFXStale {
		t.Fatalf("future FX was not typed at Evaluate: err=%v refusal=%q", err, Evaluate(sealed, nil).Refusal())
	}
	identity := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "KRW", Direction: FXAccountToInstrument, RateNum: 1, RateDen: 1, Scale: 2, AsOfMS: 1, FreshUntilMS: 10}
	identity.Digest = FXSealDigest(identity)
	if _, err := NewFXSeal(identity); err != nil {
		t.Fatalf("same-currency identity FX rejected: %v", err)
	}
}
