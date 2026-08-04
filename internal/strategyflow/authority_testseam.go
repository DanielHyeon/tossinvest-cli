//go:build tossos_testseams

package strategyflow

import (
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// AcceptedResultForAuthorityTest creates a sealed accepted result with a
// realistic validity window only in explicitly tagged test binaries.
func AcceptedResultForAuthorityTest(descriptor Descriptor, accountRef, symbol, campaignID string, quantity uint64,
	entryMinor, stopMinor, targetMinor string, observedAt, validUntil time.Time,
) (Result, error) {
	registered := false
	for _, value := range pairedDescriptors {
		registered = registered || value == descriptor
	}
	if !registered || observedAt.IsZero() || !observedAt.Before(validUntil) {
		return Result{}, fmt.Errorf("strategyflow authority test seam: invalid input")
	}
	currency := map[strategyrouter.Market]string{strategyrouter.MarketKR: "KRW", strategyrouter.MarketUS: "USD"}[descriptor.Market]
	scale := map[strategyrouter.Market]int{strategyrouter.MarketKR: 0, strategyrouter.MarketUS: 2}[descriptor.Market]
	price := func(value, source, digit string) PriceProvenance {
		return PriceProvenance{priceMinor: value, source: source, version: "authority-test-v1", digest: "sha256:" + strings.Repeat(digit, 64),
			asOf: observedAt.UTC().Format(time.RFC3339Nano), currency: currency, minorScale: scale, unitVersion: "minor-v1"}
	}
	policy := ExecutionPolicy{identity: "strategy-execution-policy:test:v1:sha256:" + strings.Repeat(map[strategyrouter.Market]string{
		strategyrouter.MarketKR: "1", strategyrouter.MarketUS: "2"}[descriptor.Market], 64)}
	lineage := sealLineage(Lineage{AccountRef: accountRef, Market: descriptor.Market, Symbol: symbol, PositionGeneration: 1,
		CandidateState: "ACTIVE", CandidateLifeID: "candidate-life:v1:sha256:" + strings.Repeat("3", 64),
		CandidateFirstSeenNS: observedAt.Add(-time.Minute).UnixNano(), CandidateLastSeenNS: observedAt.UnixNano(),
		CandidateValidUntilNS: validUntil.UnixNano(), CandidateApprovedAtNS: observedAt.Add(-time.Second).UnixNano(),
		ThresholdVersion: "threshold-v1", ThresholdSetDigest: "sha256:" + strings.Repeat("4", 64),
		CandidateEvidenceDigest: "sha256:" + strings.Repeat("5", 64), RouterEvidenceDigest: "sha256:" + strings.Repeat("6", 64),
		LaneEvidenceDigest: "sha256:" + strings.Repeat("7", 64), RouterID: strategyrouter.RouterID, RouterRelease: strategyrouter.RouterRelease,
		Horizon: descriptor.Horizon, LaneID: descriptor.LaneID, LaneVersion: descriptor.LaneVersion, LaneRelease: descriptor.Release,
		ConfigDigest: "sha256:" + strings.Repeat("8", 64), CampaignID: campaignID, LegOrdinal: 1, PlannedCeiling: quantity,
		RiskBudgetDigest: "sha256:" + strings.Repeat("9", 64), ExecutionPolicyDigest: policy.identity, Complete: true})
	evaluation := laneEvaluation{accepted: true, quantity: quantity, entry: price(entryMinor, "entry-contract", "a"),
		stop: price(stopMinor, "stop-contract", "b"), target: price(targetMinor, "target-contract", "c"), policy: policy}
	terms, ok := sealExecutionTerms(lineage, evaluation)
	if !ok {
		return Result{}, fmt.Errorf("strategyflow authority test seam: invalid execution terms")
	}
	return sealProposalResult(Result{Quantity: quantity, ExecutionTerms: terms, Lineage: lineage, CommonSafetyIndependent: true}), nil
}
