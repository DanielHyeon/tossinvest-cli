package breakoutlane

import "testing"

func TestRepairREDConstructorRejectsSkippedSequenceAndStaleSeal(t *testing.T) {
	i := fixtureInput(t)
	b := i.Bars[5].value
	b.Sequence = 7
	i.Bars[5] = ClosedBar{value: b}
	if _, err := NewEvidenceSnapshot(i); err == nil {
		t.Fatal("skipped sequence accepted")
	}
	i = fixtureInput(t)
	i.EvaluatedAtMS = 11
	sealed, err := NewEvidenceSnapshot(i)
	if err != nil || Evaluate(sealed, nil).Refusal() != RefusalFXStale {
		t.Fatalf("stale FX was not typed at Evaluate: err=%v refusal=%q", err, Evaluate(sealed, nil).Refusal())
	}
}

func TestRepairREDTimeoutBoundaryAndCorrectionRegression(t *testing.T) {
	i := fixtureInput(t)
	i.Bars = i.Bars[:16]
	for seq := uint64(17); seq <= 23; seq++ {
		i.Bars = append(i.Bars, fixtureBar(t, seq, 106, 101, 105, 1_000_000, 100_000))
	}
	i.Bars = append(i.Bars, fixtureBar(t, 24, 100, 98, 99, 1_000_000, 100_000))
	if d := Evaluate(snapshot(t, i), nil); d.Phase() != "TIMED_OUT" {
		t.Fatalf("limit retest bypass phase=%s", d.Phase())
	}
	base := fixtureInput(t)
	b := base.Bars[15].value
	b.Revision = 2
	base.Bars[15] = ClosedBar{value: b}
	prior := Evaluate(snapshot(t, base), nil)
	changed := base
	b = changed.Bars[15].value
	b.Revision = 1
	changed.Bars[15] = ClosedBar{value: b}
	if d := Evaluate(snapshot(t, changed), &prior); d.Refusal() != RefusalEvidenceInvalid {
		t.Fatalf("revision decrease accepted: %q", d.Refusal())
	}
}

func TestRepairREDSizingWorstExitCostOverflow(t *testing.T) {
	i := fixtureInput(t)
	i.Sizing.ProposedEntryMinor = maxUint64 - 2
	i.Quote = fixtureQuote(t, 5, 6)
	q := i.Quote.value
	q.AskMinor = maxUint64 - 2
	q.Digest = QuoteSealDigest(q)
	i.Quote = QuoteSeal{value: q}
	i.Sizing.StopMinor = maxUint64 - 3
	i.Sizing.TargetMinor = maxUint64
	i.Sizing.EntrySlippageMinor = 1
	i.Sizing.ExitSlippageMinor = 1
	i.Sizing.RoundTripCostAccountMinor = 1
	i.Sizing.RiskBudgetAccountMinor = maxUint64
	i.Sizing.NotionalCapAccountMinor = maxUint64
	i.Sizing.FinalCap = 1
	if got := size(i.Sizing, i.Quote, i.FX, i.EvaluatedAtMS); got.Refusal != RefusalSizingOverflow {
		t.Fatalf("overflow result=%q", got.Refusal)
	}
}

func TestTimeoutExactBoundaryKRAndUS(t *testing.T) {
	for _, market := range []struct {
		market        Market
		limit         uint64
		session, lane string
	}{{MarketKR, 8, "KRX:2026-08-18", KRLaneID}, {MarketUS, 10, "NYSE:2026-08-18", USLaneID}} {
		t.Run(string(market.market), func(t *testing.T) {
			base := fixtureInput(t)
			base.Market, base.LaneID, base.SessionID = market.market, market.lane, market.session
			for n := range base.Bars {
				b := base.Bars[n].value
				b.SessionID = market.session
				base.Bars[n] = ClosedBar{value: b}
			}
			oneBefore := base
			oneBefore.Bars = oneBefore.Bars[:16]
			for seq := uint64(17); seq < 15+market.limit; seq++ {
				b := fixtureBar(t, seq, 106, 101, 105, 1_000_000, 100_000).value
				b.SessionID = market.session
				oneBefore.Bars = append(oneBefore.Bars, ClosedBar{value: b})
			}
			b := fixtureBar(t, 15+market.limit, 100, 98, 99, 1_000_000, 100_000).value
			b.SessionID = market.session
			oneBefore.Bars = append(oneBefore.Bars, ClosedBar{value: b})
			b = fixtureBar(t, 16+market.limit, 102, 99, 101, 1_000_000, 100_000).value
			b.SessionID = market.session
			oneBefore.Bars = append(oneBefore.Bars, ClosedBar{value: b})
			if d := Evaluate(snapshot(t, oneBefore), nil); d.Phase() != "PROPOSED" {
				t.Fatalf("retest before/reclaim at limit=%+v", d)
			}
			exact := base
			exact.Bars = exact.Bars[:16]
			for seq := uint64(17); seq < 16+market.limit; seq++ {
				b := fixtureBar(t, seq, 106, 101, 105, 1_000_000, 100_000).value
				b.SessionID = market.session
				exact.Bars = append(exact.Bars, ClosedBar{value: b})
			}
			b = fixtureBar(t, 16+market.limit, 100, 98, 99, 1_000_000, 100_000).value
			b.SessionID = market.session
			exact.Bars = append(exact.Bars, ClosedBar{value: b})
			if d := Evaluate(snapshot(t, exact), nil); d.Phase() != "TIMED_OUT" {
				t.Fatalf("first retest at limit=%+v", d)
			}
			b = fixtureBar(t, 17+market.limit, 102, 99, 101, 1_000_000, 100_000).value
			b.SessionID = market.session
			exact.Bars = append(exact.Bars, ClosedBar{value: b})
			if d := Evaluate(snapshot(t, exact), nil); d.Phase() != "TIMED_OUT" {
				t.Fatalf("reclaim after limit=%+v", d)
			}
			after := base
			after.Bars = after.Bars[:16]
			for seq := uint64(17); seq < 15+market.limit; seq++ {
				b := fixtureBar(t, seq, 106, 101, 105, 1_000_000, 100_000).value
				b.SessionID = market.session
				after.Bars = append(after.Bars, ClosedBar{value: b})
			}
			b = fixtureBar(t, 15+market.limit, 100, 98, 99, 1_000_000, 100_000).value
			b.SessionID = market.session
			after.Bars = append(after.Bars, ClosedBar{value: b})
			b = fixtureBar(t, 16+market.limit, 100, 98, 99, 1_000_000, 100_000).value
			b.SessionID = market.session
			after.Bars = append(after.Bars, ClosedBar{value: b})
			if d := Evaluate(snapshot(t, after), nil); d.Phase() != "TIMED_OUT" {
				t.Fatalf("reclaim one-after limit=%+v", d)
			}
		})
	}
}

func TestCorrectionLineageRejectsEqualOrDecreaseAndAcceptsHigherOrAppend(t *testing.T) {
	base := fixtureInput(t)
	base.Bars = base.Bars[:16]
	prior := Evaluate(snapshot(t, base), nil)
	if d := Evaluate(snapshot(t, base), &prior); d.Refusal() != RefusalNone || d.SnapshotDigest() != prior.SnapshotDigest() || d.Phase() != prior.Phase() {
		t.Fatalf("duplicate snapshot changed decision=%+v", d)
	}
	higher := base
	higher.Bars = append([]ClosedBar(nil), base.Bars...)
	b := higher.Bars[15].value
	b.Revision = 2
	higher.Bars[15] = ClosedBar{value: b}
	if d := Evaluate(snapshot(t, higher), &prior); d.Refusal() != RefusalNone || d.SnapshotDigest() == prior.SnapshotDigest() {
		t.Fatalf("higher correction=%+v", d)
	}
	prior = Evaluate(snapshot(t, higher), nil)
	decrease := higher
	decrease.Bars = append([]ClosedBar(nil), higher.Bars...)
	b = decrease.Bars[15].value
	b.Revision = 1
	decrease.Bars[15] = ClosedBar{value: b}
	if d := Evaluate(snapshot(t, decrease), &prior); d.Refusal() != RefusalEvidenceInvalid {
		t.Fatalf("decrease correction=%q", d.Refusal())
	}
	appendOnly := base
	appendOnly.Bars = append(appendOnly.Bars, fixtureBar(t, 17, 106, 101, 105, 1_000_000, 100_000))
	basePrior := Evaluate(snapshot(t, base), nil)
	if d := Evaluate(snapshot(t, appendOnly), &basePrior); d.Refusal() != RefusalNone {
		t.Fatalf("append correction=%+v", d)
	}
}

func TestSnapshotRejectsSkippedOpeningAndPostBreakoutSequence(t *testing.T) {
	for _, index := range []int{4, 15} {
		t.Run(u(uint64(index)), func(t *testing.T) {
			i := fixtureInput(t)
			b := i.Bars[index].value
			b.Sequence++
			i.Bars[index] = ClosedBar{value: b}
			if _, err := NewEvidenceSnapshot(i); err == nil {
				t.Fatalf("accepted skipped sequence at index %d", index)
			}
		})
	}
}
