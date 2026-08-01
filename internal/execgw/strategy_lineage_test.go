package execgw

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

func TestStrategyDecisionLineagePayloadPreservesCompleteDecisionRecord(t *testing.T) {
	record := strategyengine.DecisionRecord{
		CandidateLifeID: "life", CandidateState: "active", CandidateFirstSeen: 1, CandidateLastSeen: 2,
		CandidateValidUntil: 3, CandidateApprovedAt: 2, Market: "KR", Symbol: "005930",
		LaneID: "lane", LaneVersion: "lane-v", SourceCommit: "commit", SourceDigest: "source", ConstantsDigest: "constants",
		ThresholdVersion: "threshold", ThresholdSetDigest: "threshold-digest", EvidenceDigest: "evidence",
		MarketInputVersion: "market-input", CalendarSource: "calendar-source", CalendarVersion: "calendar-v",
		ConfigSource: "config-source", ConfigVersion: "config-v", IndicatorSource: "indicator-source", IndicatorVersion: "indicator-v",
		IndicatorComputedAt: 4, TradingDay: true, SessionOpenAt: 5, SessionCloseAt: 6, NoEntryAfter: 6,
		BarSource: "bar-source", BarAdjusted: true, BarOpenAt: 7, BarClosedAt: 8, EvaluatedAt: 9, ExpiresAt: 10,
		Open: "11", High: "12", Low: "10", Close: "11.5", Volume: "13", Currency: "KRW",
		VWAP: "11.1", VWAPSlopePct: "0.1", EMA9: "11.2", LVNSpacePct: "1.2", TangledPct: "0.35",
		Expansion: "1.8", HVNAboveDistancePct: "1.3", StateSource: "state", StateAt: 14,
		PositionSource: "position", PositionAt: 15, EntryPrice: "11.5", LivePrice: "11.6",
		LivePriceObserved: true, EntryPriceDriftPct: "0.87", StopPrice: "11.4195", TargetPrice: "11.7415",
		ExpectedRR: "3", AcceptReasons: [7]string{"a", "b", "c", "d", "e", "f", "g"}, Identity: "decision",
	}
	lineage, err := strategyDecisionLineage(record, "7", "policy-v", "settings", "activation")
	if err != nil {
		t.Fatal(err)
	}
	var restored strategyengine.DecisionRecord
	if err := json.Unmarshal([]byte(lineage.DecisionPayload), &restored); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, record) {
		t.Fatalf("payload lost DecisionRecord fields:\n got=%+v\nwant=%+v", restored, record)
	}
	hash := sha256.Sum256([]byte(lineage.DecisionPayload))
	if lineage.DecisionPayloadDigest != "sha256:"+hex.EncodeToString(hash[:]) ||
		lineage.DecisionIdentity != record.Identity || lineage.CandidateLifeID != record.CandidateLifeID ||
		lineage.Market != record.Market || lineage.Symbol != record.Symbol || lineage.Quantity != "7" ||
		lineage.PolicyVersion != "policy-v" || lineage.SettingsDigest != "settings" ||
		lineage.ActivationManifestDigest != "activation" {
		t.Fatalf("lineage=%+v", lineage)
	}
}
