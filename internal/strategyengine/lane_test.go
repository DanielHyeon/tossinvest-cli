package strategyengine

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
)

var laneEvidence = []byte("synthetic a047 lane contract fixture; never activation evidence")

func approvedFixture(t *testing.T) strategy.ApprovedCandidate {
	t.Helper()
	evidenceDigest := candidate.DigestEvidence(laneEvidence)
	document := fmt.Sprintf(`{"version":"candidate-veto-test","market":"KR","session":"regular","metrics":[{"key":"seen_late","definition":"first-sighting rank percentile","value":"80"},{"key":"extended","definition":"gain from stored first price","value":"50"},{"key":"near_high","definition":"distance below intraday high","value":"2"}],"sample_window":{"from":"2026-07-01T00:00:00Z","to":"2026-07-31T00:00:00Z"},"sample_count":100,"missing_rate":"0.1","evidence_digest":%q}`, evidenceDigest)
	scope := candidate.ThresholdScope{Market: candidate.MarketKR, Session: candidate.SessionRegular}
	setDigest, err := candidate.DigestThresholdSetDocument(strings.NewReader(document), scope)
	if err != nil {
		t.Fatal(err)
	}
	activation, err := candidate.LoadActivationRecord(strings.NewReader(fmt.Sprintf(`{"version":"candidate-veto-test","market":"KR","session":"regular","set_digest":%q,"evidence_digest":%q,"approved_at":"2026-07-31T01:00:00Z","approved_by":"synthetic-test-not-approval"}`, setDigest, evidenceDigest)))
	if err != nil {
		t.Fatal(err)
	}
	set, err := candidate.LoadThresholdSet(strings.NewReader(document), laneEvidence, activation, scope, time.Date(2026, 7, 31, 1, 0, 0, 0, time.UTC), 0)
	if err != nil {
		t.Fatal(err)
	}
	first := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	at := first.Add(time.Minute)
	approved, err := candidate.AssessApprovedCandidate(candidate.VetoInputs{
		Candidate: candidate.Candidate{Key: candidate.Key{Market: "KR", Symbol: "005930"}, FirstSeenAt: first},
		Sighting:  candidate.Sighting{Measured: true, Rank: 90, RankTotal: 100},
		Expansion: candidate.Expansion{Measured: true, FirstPrice: "100", LastPrice: "110", FirstAt: first, LastAt: at},
		Range:     candidate.RangePosition{Measured: true, High: "120", Price: "100", At: at}, At: at,
	}, set)
	if err != nil {
		t.Fatal(err)
	}
	return strategy.FromApproved(approved)
}

func passingLaneInput(t *testing.T) LaneInput {
	return LaneInput{Approved: approvedFixture(t), Market: "KR", Session: "regular", SourceVerified: true,
		RegularSession: true, BarsClosedContiguous: true, SymbolStateNormal: true, NoExistingPosition: true,
		Volume: "1000", Price: "100.10", VWAP: "100", VWAPSlopePct: "0.08", EMA9: "100", EMA20: "99",
		LVNForwardSpacePct: "1.2", BandExpansionRate: "1.8", ExpectedRR: "3", Untangled: true,
		HVNCeilingClear: true, SignalAgeSeconds: 15, EntryPriceDriftPct: "0.20"}
}

func TestParkerConservativeLaneGoldenContractFixture(t *testing.T) {
	got := (ParkerConservativeLane{}).Evaluate(passingLaneInput(t))
	if !got.Accepted || got.Reason != RefusalNone {
		t.Fatalf("decision=%+v", got)
	}
	if got.EntryPrice != "100.1" || got.StopPrice != "99.3993" || got.TargetPrice != "102.2021" || got.ExpectedRR != "3" {
		t.Fatalf("prices=%+v", got)
	}
	if got.CandidateLifeID == "" || got.ThresholdSetDigest == "" || got.EvidenceDigest == "" || got.Identity == "" {
		t.Fatalf("provenance=%+v", got)
	}
}

func TestParkerConservativeLaneUsesFrozenGateOrder(t *testing.T) {
	base := passingLaneInput(t)
	tests := []struct {
		name   string
		mutate func(*LaneInput)
		want   Refusal
	}{
		{"unsupported", func(v *LaneInput) { v.Market = "US" }, RefusalUnsupportedScope},
		{"source", func(v *LaneInput) { v.SourceVerified = false }, RefusalSource},
		{"session", func(v *LaneInput) { v.RegularSession = false }, RefusalSession},
		{"bars", func(v *LaneInput) { v.BarsClosedContiguous = false }, RefusalBarIntegrity},
		{"state", func(v *LaneInput) { v.SymbolStateNormal = false }, RefusalSymbolState},
		{"position", func(v *LaneInput) { v.NoExistingPosition = false }, RefusalExistingPosition},
		{"volume", func(v *LaneInput) { v.Volume = "0" }, RefusalZeroVolume},
		{"indicator", func(v *LaneInput) { v.EMA9 = "" }, RefusalIndicator},
		{"vwap", func(v *LaneInput) { v.Price = "99" }, RefusalVWAPAbove},
		{"slope", func(v *LaneInput) { v.VWAPSlopePct = "0.079" }, RefusalVWAPSlope},
		{"ema", func(v *LaneInput) { v.EMA9 = "99" }, RefusalEMA9Pullback},
		{"lvn", func(v *LaneInput) { v.LVNForwardSpacePct = "1.19" }, RefusalLVNSpace},
		{"tangled", func(v *LaneInput) { v.Untangled = false }, RefusalTangledBand},
		{"expansion", func(v *LaneInput) { v.BandExpansionRate = "1.81" }, RefusalBandExpansion},
		{"rr", func(v *LaneInput) { v.ExpectedRR = "1.49" }, RefusalRR},
		{"hvn", func(v *LaneInput) { v.HVNCeilingClear = false }, RefusalHVNCeiling},
		{"age", func(v *LaneInput) { v.SignalAgeSeconds = 16 }, RefusalAge},
		{"drift", func(v *LaneInput) { v.EntryPriceDriftPct = "0.201" }, RefusalDrift},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := base
			tc.mutate(&in)
			got := (ParkerConservativeLane{}).Evaluate(in)
			if got.Accepted || got.Reason != tc.want {
				t.Fatalf("got=%+v want=%s", got, tc.want)
			}
		})
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
