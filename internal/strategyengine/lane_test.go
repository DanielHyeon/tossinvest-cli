package strategyengine

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
)

var laneEvidence = []byte("synthetic a047 lane contract fixture; never activation evidence")

func approvedFixture(t *testing.T, evaluatedAt, approvedAt, lastSeenAt time.Time) strategy.ApprovedSnapshot {
	t.Helper()
	evidenceDigest := candidate.DigestEvidence(laneEvidence)
	document := fmt.Sprintf(`{"version":"candidate-veto-test","market":"KR","session":"regular","metrics":[{"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},{"key":"extended","definition":"gain from stored first price","value":"50"},{"key":"near_high","definition":"distance below intraday high","value":"2"}],"sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-07-31T00:00:00Z"},"sample_count":100,"missing_rate":"0.1","evidence_digest":%q}`, evidenceDigest)
	scope := candidate.ThresholdScope{Market: candidate.MarketKR, Session: candidate.SessionRegular}
	setDigest, err := candidate.DigestThresholdSetDocument(strings.NewReader(document), scope)
	if err != nil {
		t.Fatal(err)
	}
	activationJSON := fmt.Sprintf(`{"version":"candidate-veto-test","market":"KR","session":"regular","set_digest":%q,"evidence_digest":%q,"approved_at":%q,"approved_by":"synthetic-test-not-approval"}`, setDigest, evidenceDigest, approvedAt.UTC().Format(time.RFC3339Nano))
	activation, err := candidate.LoadActivationRecord(strings.NewReader(activationJSON))
	if err != nil {
		t.Fatal(err)
	}
	set, err := candidate.LoadThresholdSet(strings.NewReader(document), laneEvidence, activation, scope, approvedAt, 0)
	if err != nil {
		t.Fatal(err)
	}
	firstSeenAt := evaluatedAt.Add(-20 * time.Minute)
	assessedAt := lastSeenAt.Add(time.Minute)
	approved, err := candidate.AssessApprovedCandidate(candidate.VetoInputs{
		Candidate: candidate.Candidate{
			Key: candidate.Key{Market: "KR", Symbol: "005930"}, State: candidate.StateActive,
			FirstSeenAt: firstSeenAt, LastSeenAt: lastSeenAt,
		},
		Sighting:  candidate.Sighting{Measured: true, Rank: 90, RankTotal: 100},
		Expansion: candidate.Expansion{Measured: true, FirstPrice: "100", LastPrice: "110", FirstAt: firstSeenAt, LastAt: assessedAt},
		Range:     candidate.RangePosition{Measured: true, High: "120", Price: "100", At: assessedAt},
		At:        assessedAt,
	}, set)
	if err != nil {
		t.Fatal(err)
	}
	return strategy.SealApproved(approved)
}

type normalStateSource struct{ at time.Time }

func (s normalStateSource) ReadSymbolState(_ context.Context, market, symbol string) (strategymarket.StateReading, error) {
	return strategymarket.StateReading{Market: market, Symbol: symbol, State: strategymarket.StateNormal, ObservedAt: s.at, Source: strategymarket.SourceOfficialSymbolState}, nil
}

func (s normalStateSource) ReadPosition(_ context.Context, market, symbol string) (strategymarket.PositionReading, error) {
	return strategymarket.PositionReading{Market: market, Symbol: symbol, Quantity: "0", ObservedAt: s.at, Source: strategymarket.SourceOfficialPosition}, nil
}

func verifiedBarFixture(t *testing.T, hour, minute int, mutate func([]strategymarket.RawMinuteCandle)) strategymarket.VerifiedBar {
	t.Helper()
	zone := time.FixedZone("KST", 9*60*60)
	rows := make([]strategymarket.RawMinuteCandle, 5)
	for index := range rows {
		rows[index] = strategymarket.RawMinuteCandle{
			Timestamp: time.Date(2026, 7, 31, hour, minute+index, 0, 0, zone).Format(time.RFC3339),
			Open:      "100", High: "101", Low: "99.9", Close: "100.1", Volume: "200", Currency: "KRW",
		}
	}
	if mutate != nil {
		mutate(rows)
	}
	page := strategymarket.SealAdaptedOfficialMinutePage("KR", "005930", strategymarket.IntervalOneMinute, false, strategymarket.SourceOfficialOpenAPI, rows)
	bar, err := strategymarket.SealOfficialClosedKRXFiveMinute(page, time.Date(2026, 7, 31, hour, minute+5, 0, 0, zone))
	if err != nil {
		t.Fatal(err)
	}
	return bar
}

func marketFields(bar strategymarket.VerifiedBar, evaluatedAt time.Time) MarketInputFields {
	zone := time.FixedZone("KST", 9*60*60)
	return MarketInputFields{
		Version: MarketInputVersion, Market: "KR", Symbol: "005930",
		CalendarSource: CalendarSource, CalendarVersion: "krx-calendar:2026-07-31",
		ConfigSource: ConfigSource, ConfigVersion: ConfigVersion,
		IndicatorSource: IndicatorSource, IndicatorVersion: IndicatorVersion,
		EvaluatedAt: evaluatedAt, IndicatorComputedAt: evaluatedAt, TradingDay: true,
		SessionOpenAt:  time.Date(2026, 7, 31, 9, 0, 0, 0, zone),
		SessionCloseAt: time.Date(2026, 7, 31, 15, 30, 0, 0, zone),
		NoEntryAfter:   time.Date(2026, 7, 31, 15, 20, 0, 0, zone),
		VWAP:           "100", VWAPSlopePct: "0.08", EMA9: "100",
		LVNForwardSpacePct: "1.2", TangledScorePct: "0.35",
		BandExpansionRate: "1.8", HVNAboveDistancePct: "1.2", CurrentPrice: "100.3002",
	}
}

func laneInputAt(t *testing.T, bar strategymarket.VerifiedBar, evaluatedAt time.Time, mutate func(*MarketInputFields)) LaneInput {
	t.Helper()
	fields := marketFields(bar, evaluatedAt)
	if mutate != nil {
		mutate(&fields)
	}
	market, err := SealVersionedMarketInput(fields)
	if err != nil {
		t.Fatal(err)
	}
	state, err := strategymarket.RequireFreshNormalState(context.Background(), normalStateSource{at: evaluatedAt}, "KR", "005930", evaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	position, err := strategymarket.RequireNoPosition(context.Background(), normalStateSource{at: evaluatedAt}, "KR", "005930", evaluatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return LaneInput{
		Approved: approvedFixture(t, evaluatedAt, evaluatedAt.Add(-time.Minute), evaluatedAt.Add(-5*time.Minute)),
		// Same-package fixture is parity-only. The real source manifest is unavailable,
		// and VerifyFrozenSource therefore never reports fake success.
		Source: SourceProof{digest: FrozenSourceSetDigest, valid: true},
		Bar:    bar, State: state, Position: position, Market: market,
	}
}

func passingLaneInput(t *testing.T) LaneInput {
	t.Helper()
	bar := verifiedBarFixture(t, 9, 10, nil)
	return laneInputAt(t, bar, bar.ClosedAt(), nil)
}

func TestParkerConservativeLaneGoldenContractFixture(t *testing.T) {
	got := (ParkerConservativeLane{}).Evaluate(passingLaneInput(t))
	if !got.Decision.Valid() || got.Reason != RefusalNone {
		t.Fatalf("decision=%+v", got)
	}
	record := got.Decision.Record()
	if record.EntryPrice != "100.1" || record.StopPrice != "99.3993" || record.TargetPrice != "102.2021" ||
		record.ExpectedRR != "1.7142857142857142857142857143" || record.LivePrice != "100.3002" || record.EntryPriceDriftPct != "0.2" {
		t.Fatalf("derived prices=%+v", record)
	}
	wantReasons := [7]string{"VWAP_ABOVE", "VWAP_SLOPE_UP", "EMA9_PULLBACK_CONFIRMED", "VOLUME_PROFILE_SPACE_OK", "RR_GE_2", "NOT_TANGLED", "NOT_AFTER_ENTRY_CUTOFF"}
	if record.AcceptReasons != wantReasons || record.CalendarSource != CalendarSource || record.ConfigVersion != ConfigVersion ||
		record.IndicatorSource != IndicatorSource || record.BarSource != string(strategymarket.SourceOfficialOpenAPI) ||
		record.CandidateState != "active" || record.CandidateLastSeen >= record.CandidateValidUntil {
		t.Fatalf("provenance/reasons=%+v", record)
	}
}

func TestParkerFrozenGateBoundariesAndRefusalOrder(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*MarketInputFields)
		want   Refusal
	}{
		{"vwap above", func(v *MarketInputFields) { v.VWAP = "100.1" }, RefusalVWAPAbove},
		{"negative slope is gate refusal", func(v *MarketInputFields) { v.VWAPSlopePct = "-0.01" }, RefusalVWAPSlope},
		{"slope threshold", func(v *MarketInputFields) { v.VWAPSlopePct = "0.079" }, RefusalVWAPSlope},
		{"ema9 pullback", func(v *MarketInputFields) { v.EMA9 = "100.2" }, RefusalEMA9Pullback},
		{"negative LVN is space refusal", func(v *MarketInputFields) { v.LVNForwardSpacePct = "-0.1" }, RefusalLVNSpace},
		{"LVN threshold", func(v *MarketInputFields) { v.LVNForwardSpacePct = "1.199" }, RefusalLVNSpace},
		{"tangled below threshold", func(v *MarketInputFields) { v.TangledScorePct = "0.349" }, RefusalTangledBand},
		{"tangled exact threshold accepted", func(v *MarketInputFields) { v.TangledScorePct = "0.35" }, RefusalNone},
		{"optional expansion", func(v *MarketInputFields) { v.BandExpansionRate = "1.801" }, RefusalBandExpansion},
		{"HVN below forward space", func(v *MarketInputFields) { v.HVNAboveDistancePct = "1.199" }, RefusalHVNCeiling},
		{"HVN exact forward space", func(v *MarketInputFields) { v.HVNAboveDistancePct = "1.2" }, RefusalNone},
		{"drift above limit", func(v *MarketInputFields) { v.CurrentPrice = "100.300201" }, RefusalDrift},
		{"nonpositive live price is drift refusal", func(v *MarketInputFields) { v.CurrentPrice = "0" }, RefusalDrift},
		{"missing live price falls back to close", func(v *MarketInputFields) { v.CurrentPrice = "" }, RefusalNone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := passingLaneInput(t)
			fields := marketFields(base.Bar, base.Bar.ClosedAt())
			test.mutate(&fields)
			sealed, err := SealVersionedMarketInput(fields)
			if err != nil {
				t.Fatal(err)
			}
			base.Market = sealed
			got := (ParkerConservativeLane{}).Evaluate(base)
			if got.Reason != test.want {
				t.Fatalf("got=%s want=%s", got.Reason, test.want)
			}
		})
	}
	marketType := reflect.TypeOf(MarketInputFields{})
	for _, forbidden := range []string{"EMA20", "ExpectedRR", "HVNCeilingPrice", "SignalAgeSeconds", "EntryPriceDriftPct"} {
		if _, present := marketType.FieldByName(forbidden); present {
			t.Errorf("caller-asserted field %s remains public", forbidden)
		}
	}
}

func TestParkerSessionUsesInjectedEvaluationTimeAndInclusiveCutoff(t *testing.T) {
	zone := time.FixedZone("KST", 9*60*60)
	beforeOpenSkip := verifiedBarFixture(t, 9, 0, nil)
	before := laneInputAt(t, beforeOpenSkip, time.Date(2026, 7, 31, 9, 10, 0, -1, zone), nil)
	if got := (ParkerConservativeLane{}).Evaluate(before); got.Reason != RefusalSession {
		t.Fatalf("09:10 minus 1ns reason=%s", got.Reason)
	}
	atBoundaryBar := verifiedBarFixture(t, 9, 5, nil)
	atBoundary := laneInputAt(t, atBoundaryBar, time.Date(2026, 7, 31, 9, 10, 0, 0, zone), nil)
	if got := (ParkerConservativeLane{}).Evaluate(atBoundary); got.Reason != RefusalNone {
		t.Fatalf("09:10 exact reason=%s", got.Reason)
	}
	cutoffBar := verifiedBarFixture(t, 15, 15, nil)
	cutoff := laneInputAt(t, cutoffBar, time.Date(2026, 7, 31, 15, 20, 0, 0, zone), nil)
	if got := (ParkerConservativeLane{}).Evaluate(cutoff); got.Reason != RefusalNone {
		t.Fatalf("cutoff exact reason=%s", got.Reason)
	}
	after := laneInputAt(t, cutoffBar, time.Date(2026, 7, 31, 15, 20, 0, 1, zone), nil)
	if got := (ParkerConservativeLane{}).Evaluate(after); got.Reason != RefusalSession {
		t.Fatalf("cutoff plus 1ns reason=%s", got.Reason)
	}
}

func TestParkerSignalAgeUsesNanosecondHalfOpenExpiry(t *testing.T) {
	bar := verifiedBarFixture(t, 9, 10, nil)
	for _, test := range []struct {
		name string
		age  time.Duration
		want Refusal
	}{{"exact 15 seconds", 15 * time.Second, RefusalNone}, {"15 seconds plus 1ns", 15*time.Second + time.Nanosecond, RefusalAge}} {
		t.Run(test.name, func(t *testing.T) {
			input := laneInputAt(t, bar, bar.ClosedAt().Add(test.age), nil)
			got := (ParkerConservativeLane{}).Evaluate(input)
			if got.Reason != test.want {
				t.Fatalf("reason=%s want=%s", got.Reason, test.want)
			}
			if test.want == RefusalNone && got.Decision.Record().ExpiresAt != bar.ClosedAt().UTC().Add(15*time.Second+time.Nanosecond).UnixNano() {
				t.Fatalf("expires=%d", got.Decision.Record().ExpiresAt)
			}
		})
	}
}

func TestParkerRequiresApprovalActivatedAndCandidateCurrentlyActive(t *testing.T) {
	base := passingLaneInput(t)
	evaluatedAt := base.Market.evaluatedAt
	tests := []struct {
		name       string
		approvedAt time.Time
		lastSeenAt time.Time
		want       Refusal
	}{
		{"current", evaluatedAt.Add(-time.Minute), evaluatedAt.Add(-5 * time.Minute), RefusalNone},
		{"future threshold activation", evaluatedAt.Add(time.Nanosecond), evaluatedAt.Add(-5 * time.Minute), RefusalCandidate},
		{"candidate exact exclusive expiry", evaluatedAt.Add(-time.Minute), evaluatedAt.Add(-candidate.DefaultStalenessTTL), RefusalCandidate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := base
			input.Approved = approvedFixture(t, evaluatedAt, test.approvedAt, test.lastSeenAt)
			if got := (ParkerConservativeLane{}).Evaluate(input); got.Reason != test.want {
				t.Fatalf("reason=%s want=%s", got.Reason, test.want)
			}
		})
	}
}

func TestParkerRejectsZeroAndForgedProofsWithoutPanic(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*LaneInput)
		want   Refusal
	}{
		{"zero source", func(input *LaneInput) { input.Source = SourceProof{} }, RefusalSource},
		{"forged source", func(input *LaneInput) { input.Source = SourceProof{digest: strings.Repeat("0", 64), valid: true} }, RefusalSource},
		{"zero market bundle", func(input *LaneInput) { input.Market = VersionedMarketInput{} }, RefusalIndicator},
		{"zero bar", func(input *LaneInput) { input.Bar = strategymarket.VerifiedBar{} }, RefusalBarIntegrity},
		{"zero state", func(input *LaneInput) { input.State = strategymarket.FreshNormalState{} }, RefusalSymbolState},
		{"zero position", func(input *LaneInput) { input.Position = strategymarket.NoPositionProof{} }, RefusalExistingPosition},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := passingLaneInput(t)
			test.mutate(&input)
			got := (ParkerConservativeLane{}).Evaluate(input)
			if got.Decision.Valid() || got.Reason != test.want {
				t.Fatalf("got=%+v want=%s", got, test.want)
			}
		})
	}
}

func TestVersionedMarketInputRejectsCallerProvenanceLaundering(t *testing.T) {
	bar := verifiedBarFixture(t, 9, 10, nil)
	base := marketFields(bar, bar.ClosedAt())
	for _, mutate := range []func(*MarketInputFields){
		func(v *MarketInputFields) { v.CalendarSource = "caller-claimed" },
		func(v *MarketInputFields) { v.ConfigVersion = "caller-config" },
		func(v *MarketInputFields) { v.IndicatorSource = "caller-indicator" },
		func(v *MarketInputFields) { v.IndicatorComputedAt = v.EvaluatedAt.Add(-time.Nanosecond) },
	} {
		fields := base
		mutate(&fields)
		if sealed, err := SealVersionedMarketInput(fields); err == nil || sealed.valid {
			t.Fatalf("laundered bundle accepted: %+v", fields)
		}
	}
}

func TestSourceManifestFailsClosedWithoutFrozenRowsOrExactDigest(t *testing.T) {
	if _, err := VerifyFrozenSource(nil); err == nil {
		t.Fatal("missing manifest accepted")
	}
	if _, err := VerifyFrozenSource([]SourceBlob{{Path: "a.go", BlobSHA256: strings.Repeat("0", 64)}}); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("mismatch err=%v", err)
	}
	if _, err := VerifyFrozenSource([]SourceBlob{{Path: "b.go", BlobSHA256: strings.Repeat("0", 64)}, {Path: "a.go", BlobSHA256: strings.Repeat("1", 64)}}); err == nil {
		t.Fatal("unsorted manifest accepted")
	}
}
