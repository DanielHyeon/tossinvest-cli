package breakoutlane

import (
	"reflect"
	"testing"
)

func TestFinalRedTeamDuplicateSnapshotIsIdempotent(t *testing.T) {
	terminalInput := fixtureInput(t)
	bar := terminalInput.Bars[17].value
	bar.LowMinor, bar.CloseMinor = 89, 89
	terminalInput.Bars[17] = ClosedBar{value: bar}
	preterminalInput := fixtureInput(t)
	preterminalInput.Bars = preterminalInput.Bars[:16]

	for _, tc := range []struct {
		name  string
		input EvidenceInput
	}{
		{"proposed", fixtureInput(t)},
		{"terminal", terminalInput},
		{"preterminal", preterminalInput},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sealed := snapshot(t, tc.input)
			prior := Evaluate(sealed, nil)
			got := Evaluate(sealed, &prior)
			if !reflect.DeepEqual(got, prior) {
				t.Fatalf("duplicate changed decision:\n got=%+v\nwant=%+v", got, prior)
			}
		})
	}
}

func TestFinalRedTeamLivenessRefusalsUsePublicEvaluatePath(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*EvidenceInput)
		want RefusalCode
	}{
		{"stale quote", func(i *EvidenceInput) {
			i.EvaluatedAtMS = 11
			i.Quote = redteamQuote(t, 100, 101, "USD", 0, 6)
			i.FX = redteamFX(t, "KRW", "USD", 1, 1, 1, 20)
		}, RefusalQuoteStale},
		{"future quote", func(i *EvidenceInput) {
			i.Quote = redteamQuote(t, 100, 101, "USD", 11, 11)
		}, RefusalQuoteStale},
		{"stale FX", func(i *EvidenceInput) {
			i.FX = redteamFX(t, "KRW", "USD", 1, 1, 1, 9)
		}, RefusalFXStale},
		{"future FX", func(i *EvidenceInput) {
			i.FX = redteamFX(t, "KRW", "USD", 1, 1, 11, 20)
		}, RefusalFXStale},
		{"currency mismatch", func(i *EvidenceInput) {
			i.Quote = redteamQuote(t, 100, 101, "JPY", 5, 6)
		}, RefusalFXCurrencyMismatch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := fixtureInput(t)
			tc.mut(&i)
			sealed, err := NewEvidenceSnapshot(i)
			if err != nil {
				t.Fatalf("structurally valid snapshot rejected before Evaluate: %v", err)
			}
			if got := Evaluate(sealed, nil).Refusal(); got != tc.want {
				t.Fatalf("refusal=%q want=%q", got, tc.want)
			}
		})
	}
}

func TestFinalRedTeamSameCurrencyIdentityFXAndSnapshots(t *testing.T) {
	for _, tc := range []struct {
		name, currency, session, lane string
		market                        Market
	}{
		{"KR KRW", "KRW", "KRX:2026-08-18", KRLaneID, MarketKR},
		{"US USD", "USD", "NYSE:2026-08-18", USLaneID, MarketUS},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i := fixtureInput(t)
			i.Market, i.SessionID, i.LaneID = tc.market, tc.session, tc.lane
			for n := range i.Bars {
				bar := i.Bars[n].value
				bar.SessionID = tc.session
				i.Bars[n] = ClosedBar{value: bar}
			}
			i.Quote = redteamQuote(t, 100, 101, tc.currency, 5, 6)
			i.FX = redteamFX(t, tc.currency, tc.currency, 1, 1, 1, 10)
			sealed, err := NewEvidenceSnapshot(i)
			if err != nil {
				t.Fatalf("identity snapshot rejected: %v", err)
			}
			if d := Evaluate(sealed, nil); d.Refusal() != RefusalNone || d.Phase() != "PROPOSED" {
				t.Fatalf("identity evaluation=%+v", d)
			}
		})
	}

	nonIdentity := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "KRW", Direction: FXAccountToInstrument, RateNum: 2, RateDen: 1, Scale: 2, AsOfMS: 1, FreshUntilMS: 10}
	nonIdentity.Digest = FXSealDigest(nonIdentity)
	if _, err := NewFXSeal(nonIdentity); err == nil {
		t.Fatal("non-identity same-currency FX accepted")
	}
}

func TestFinalRedTeamClosedBarRejectsCloseOutsideRange(t *testing.T) {
	for _, closeMinor := range []uint64{89, 101} {
		i := ClosedBarInput{Sequence: 1, Revision: 1, IntervalMS: 60_000, ID: "bar-1", SessionID: "KRX:2026-08-18", RegularSession: true, Closed: true, HighMinor: 100, LowMinor: 90, CloseMinor: closeMinor, RVOLPPM: 1_500_000, UpperWickRangePPM: 100_000}
		if _, err := NewClosedBar(i); err == nil {
			t.Fatalf("close %d outside [90,100] accepted", closeMinor)
		}
	}
}

func TestFinalRedTeamSizingUsesWorstEntryInRisk(t *testing.T) {
	q := redteamQuote(t, 109, 110, "USD", 5, 6)
	f := redteamFX(t, "KRW", "USD", 1, 1, 1, 10)
	in := SizingInput{ProposedEntryMinor: 100, StopMinor: 90, TargetMinor: 200, EntrySlippageMinor: 2, ExitSlippageMinor: 3, RoundTripCostAccountMinor: 4, RiskBudgetAccountMinor: 290, NotionalCapAccountMinor: 100_000, FinalCap: 100}
	got := size(in, q, f, 10)
	if got.Refusal != RefusalNone || got.WorstEntryMinor != 112 || got.RiskPerShareMinor != 29 || got.CandidateQuantity != 10 || got.FinalQuantity != 10 {
		t.Fatalf("worst-entry sizing=%+v", got)
	}
}

func TestFinalRedTeamConfigDigestBindsEveryField(t *testing.T) {
	base := redteamConfigInput()
	if got := V1ConfigDigest(base); got != base.Digest {
		t.Fatalf("exported config digest=%q want=%q", got, base.Digest)
	}
	if _, err := NewV1Config(base); err != nil {
		t.Fatalf("canonical config rejected: %v", err)
	}
	for _, tc := range []struct {
		name string
		mut  func(*V1ConfigInput)
	}{
		{"version", func(i *V1ConfigInput) { i.Version = "v2" }},
		{"tick", func(i *V1ConfigInput) { i.TickMinor++ }},
		{"opening range", func(i *V1ConfigInput) { i.OpeningRangeMinutes++ }},
		{"breakout buffer", func(i *V1ConfigInput) { i.BreakoutBufferPPM++ }},
		{"retest tolerance", func(i *V1ConfigInput) { i.RetestTolerancePPM++ }},
		{"KR timeout", func(i *V1ConfigInput) { i.TimeoutKR++ }},
		{"US timeout", func(i *V1ConfigInput) { i.TimeoutUS++ }},
		{"RVOL", func(i *V1ConfigInput) { i.RVOLMinPPM++ }},
		{"wick", func(i *V1ConfigInput) { i.UpperWickRangeMaxPPM++ }},
		{"quote age", func(i *V1ConfigInput) { i.MaxQuoteAgeMS++ }},
		{"spread", func(i *V1ConfigInput) { i.MaxSpreadPPM++ }},
		{"drift", func(i *V1ConfigInput) { i.MaxEntryDriftPPM++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := base
			tc.mut(&changed)
			if redteamConfigDigest(changed) == base.Digest {
				t.Fatal("config digest did not change")
			}
			if _, err := NewV1Config(changed); err == nil {
				t.Fatal("config mutation retained stale digest")
			}
		})
	}

	changed := base
	changed.MaxSpreadPPM++
	changed.Digest = redteamConfigDigest(changed)
	changedConfig, err := NewV1Config(changed)
	if err != nil {
		t.Fatal(err)
	}
	baseConfig, _ := NewV1Config(base)
	first := fixtureInput(t)
	first.Config = baseConfig
	second := fixtureInput(t)
	second.Config = changedConfig
	firstSnapshot, err := NewEvidenceSnapshot(first)
	if err != nil {
		t.Fatal(err)
	}
	secondSnapshot, err := NewEvidenceSnapshot(second)
	if err != nil {
		t.Fatal(err)
	}
	if firstSnapshot.digest == secondSnapshot.digest || setupID(first) == setupID(second) {
		t.Fatal("recomputed config digest did not change snapshot/setup identity")
	}
}

func TestFinalRedTeamCorrectionLineageBindsBarContent(t *testing.T) {
	base := fixtureInput(t)
	base.Bars = base.Bars[:16]
	prior := Evaluate(snapshot(t, base), nil)
	for _, tc := range []struct {
		name string
		mut  func(*ClosedBarInput)
	}{
		{"OHLC", func(b *ClosedBarInput) { b.HighMinor-- }},
		{"RVOL", func(b *ClosedBarInput) { b.RVOLPPM++ }},
		{"flag", func(b *ClosedBarInput) { b.VolumeExpanded = !b.VolumeExpanded }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := base
			changed.Bars = append([]ClosedBar(nil), base.Bars...)
			bar := changed.Bars[0].value
			tc.mut(&bar)
			changed.Bars[0] = ClosedBar{value: bar}
			changed.Bars = append(changed.Bars, fixtureBar(t, 17, 106, 101, 105, 1_000_000, 100_000))
			if d := Evaluate(snapshot(t, changed), &prior); d.Refusal() != RefusalEvidenceInvalid {
				t.Fatalf("same-revision content mutation accepted: %+v", d)
			}
		})
	}

	appendOnly := base
	appendOnly.Bars = append(append([]ClosedBar(nil), base.Bars...), fixtureBar(t, 17, 106, 101, 105, 1_000_000, 100_000))
	if d := Evaluate(snapshot(t, appendOnly), &prior); d.Refusal() != RefusalNone {
		t.Fatalf("unchanged-prefix append rejected: %+v", d)
	}
	higher := base
	higher.Bars = append([]ClosedBar(nil), base.Bars...)
	bar := higher.Bars[0].value
	bar.Revision++
	bar.RVOLPPM++
	higher.Bars[0] = ClosedBar{value: bar}
	if d := Evaluate(snapshot(t, higher), &prior); d.Refusal() != RefusalNone {
		t.Fatalf("higher-revision correction rejected: %+v", d)
	}
}

func redteamQuote(t *testing.T, bid, ask uint64, currency string, source, received uint64) QuoteSeal {
	t.Helper()
	i := QuoteSealInput{BidMinor: bid, AskMinor: ask, LastMinor: ask, Currency: currency, SourceObservedAtMS: source, ReceivedAtMS: received}
	i.Digest = QuoteSealDigest(i)
	q, err := NewQuoteSeal(i)
	if err != nil {
		t.Fatal(err)
	}
	return q
}

func redteamFX(t *testing.T, account, instrument string, num, den, asOf, freshUntil uint64) FXSeal {
	t.Helper()
	i := FXSealInput{AccountCurrency: account, InstrumentCurrency: instrument, Direction: FXAccountToInstrument, RateNum: num, RateDen: den, Scale: 2, AsOfMS: asOf, FreshUntilMS: freshUntil}
	i.Digest = FXSealDigest(i)
	f, err := NewFXSeal(i)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func redteamConfigInput() V1ConfigInput {
	i := V1ConfigInput{Version: LaneVersionV1, TickMinor: 1, OpeningRangeMinutes: v1OpeningRangeMinutes, BreakoutBufferPPM: v1BreakoutBufferPPM, RetestTolerancePPM: v1RetestToleranceMinPPM, TimeoutKR: v1TimeoutKRClosedBars, TimeoutUS: v1TimeoutUSClosedBars, RVOLMinPPM: v1RVOLMinPPM, UpperWickRangeMaxPPM: v1UpperWickRangeMaxPPM, MaxQuoteAgeMS: 10, MaxSpreadPPM: 100_000, MaxEntryDriftPPM: 100_000}
	i.Digest = redteamConfigDigest(i)
	return i
}

func redteamConfigDigest(i V1ConfigInput) string {
	return "sha256:" + hashFields("tossos.breakout.config.v1", i.Version, u(i.TickMinor), u(i.OpeningRangeMinutes), u(i.BreakoutBufferPPM), u(i.RetestTolerancePPM), u(i.TimeoutKR), u(i.TimeoutUS), u(i.RVOLMinPPM), u(i.UpperWickRangeMaxPPM), u(i.MaxQuoteAgeMS), u(i.MaxSpreadPPM), u(i.MaxEntryDriftPPM))
}
