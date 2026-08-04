//go:build tossos_testseams

package strategyflow

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// AcceptedResultForJournalTest creates sealed evidence only in explicitly
// tagged test builds. Production binaries do not contain this constructor.
func AcceptedResultForJournalTest(descriptor Descriptor, accountRef, symbol, campaignID string, quantity uint64, entryMinor, stopMinor, targetMinor string) (Result, error) {
	matched := false
	for _, registered := range pairedDescriptors {
		if descriptor == registered {
			matched = true
			break
		}
	}
	if !matched {
		return Result{}, fmt.Errorf("strategyflow test seam: descriptor is not registered")
	}
	currency := map[strategyrouter.Market]string{strategyrouter.MarketKR: "KRW", strategyrouter.MarketUS: "USD"}[descriptor.Market]
	scale := map[strategyrouter.Market]int{strategyrouter.MarketKR: 0, strategyrouter.MarketUS: 2}[descriptor.Market]
	price := func(value, source string) PriceProvenance {
		return PriceProvenance{priceMinor: value, source: source, version: "test-v1", digest: "sha256:" + strings.Repeat(map[string]string{"entry": "a", "stop": "b", "target": "c"}[source], 64),
			asOf: "2026-08-04T00:00:00Z", currency: currency, minorScale: scale, unitVersion: "minor-v1"}
	}
	policy := ExecutionPolicy{stagedTargetMinor: targetMinor, fairValueMinor: targetMinor, entryCostsMinor: "1", exitCostsMinor: "1", minimumRRPPM: 1,
		decisionDigest: "sha256:" + strings.Repeat("d", 64), calendarDigest: "sha256:" + strings.Repeat("e", 64), capSnapshotID: "cap-test",
		identity: "strategy-execution-policy:test:v1:sha256:" + strings.Repeat(map[strategyrouter.Market]string{strategyrouter.MarketKR: "1", strategyrouter.MarketUS: "2"}[descriptor.Market], 64)}
	lineage := sealLineage(Lineage{AccountRef: accountRef, Market: descriptor.Market, Symbol: symbol, PositionGeneration: 1,
		CandidateState: "ACTIVE", CandidateLifeID: "candidate-life:v1:sha256:" + strings.Repeat("3", 64), CandidateFirstSeenNS: 1,
		CandidateLastSeenNS: 2, CandidateValidUntilNS: 4, CandidateApprovedAtNS: 3, ThresholdVersion: "threshold-v1",
		ThresholdSetDigest: "sha256:" + strings.Repeat("4", 64), CandidateEvidenceDigest: "sha256:" + strings.Repeat("5", 64),
		RouterEvidenceDigest: "sha256:" + strings.Repeat("6", 64), LaneEvidenceDigest: "sha256:" + strings.Repeat("7", 64),
		RouterID: strategyrouter.RouterID, RouterRelease: strategyrouter.RouterRelease, Horizon: descriptor.Horizon,
		LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion, LaneRelease: descriptor.Release, ConfigDigest: "sha256:" + strings.Repeat("8", 64),
		CampaignID: campaignID, LegOrdinal: 1, PlannedCeiling: quantity, RiskBudgetDigest: "sha256:" + strings.Repeat("9", 64),
		ExecutionPolicyDigest: policy.identity, Complete: true})
	evaluation := laneEvaluation{accepted: true, quantity: quantity, entry: price(entryMinor, "entry"), stop: price(stopMinor, "stop"), target: price(targetMinor, "target"), policy: policy}
	terms, ok := sealExecutionTerms(lineage, evaluation)
	if !ok {
		return Result{}, fmt.Errorf("strategyflow test seam: invalid execution terms")
	}
	return Result{Quantity: quantity, ExecutionTerms: terms, Lineage: lineage, CommonSafetyIndependent: true}, nil
}
