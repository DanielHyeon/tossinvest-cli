package strategyflow

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

func defaultRegistry() registry {
	return newRegistry([]binding{
		{descriptor: pairedDescriptors[0], evaluate: evaluateContinuationKR},
		{descriptor: pairedDescriptors[1], evaluate: evaluateContinuationUS},
		{descriptor: pairedDescriptors[2], evaluate: evaluateReversalKR},
		{descriptor: pairedDescriptors[3], evaluate: evaluateReversalUS},
		{descriptor: pairedDescriptors[4], evaluate: evaluateWeeklyKR},
		{descriptor: pairedDescriptors[5], evaluate: evaluateWeeklyUS},
	})
}

func evaluateContinuationKR(input LaneInput) laneEvaluation {
	return fromContinuation(continuationlane.EvaluateKR(input.continuationKR))
}

func evaluateContinuationUS(input LaneInput) laneEvaluation {
	return fromContinuation(continuationlane.EvaluateUS(input.continuationUS))
}

func fromContinuation(outcome continuationlane.Outcome) laneEvaluation {
	lineage := outcome.Lineage
	return laneEvaluation{accepted: outcome.Kind == continuationlane.OutcomeDecision && outcome.Code == continuationlane.RefusalNone,
		nativeCode: string(outcome.Code), quantity: outcome.Quantity, entry: continuationPrice(outcome.EntryProvenance), stop: continuationPrice(outcome.StopProvenance),
		target: continuationPrice(outcome.TargetProvenance), policy: ExecutionPolicy{identity: outcome.ExecutionPolicyDigest}, lineage: laneLineage{
			AccountRef: lineage.AccountRef, Market: strategyrouter.Market(lineage.Market), Symbol: lineage.Symbol,
			PositionGeneration: uint64(lineage.PositionGeneration), LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion,
			CandidateID: lineage.CandidateID, EvidenceDigest: lineage.EvidenceDigest, ConfigDigest: lineage.ConfigDigest,
			CampaignID: lineage.CampaignID, LegOrdinal: lineage.LegOrdinal, PlannedCeiling: lineage.PlannedCeiling,
			RiskBudgetDigest: lineage.RiskBudgetDigest,
		}}
}

func evaluateReversalKR(input LaneInput) laneEvaluation {
	outcome := reversallane.EvaluateKR(input.reversalKR)
	return fromReversal(outcome, input.reversalKR.Evidence.CommonEnvelope)
}

func evaluateReversalUS(input LaneInput) laneEvaluation {
	outcome := reversallane.EvaluateUS(input.reversalUS)
	return fromReversal(outcome, input.reversalUS.Evidence.CommonEnvelope)
}

func fromReversal(outcome reversallane.EvaluationResult, envelope reversallane.CommonEnvelope) laneEvaluation {
	lineage := outcome.Lineage
	return laneEvaluation{accepted: outcome.Kind == reversallane.OutcomeDecision && outcome.Code == "",
		nativeCode: string(outcome.Code), quantity: outcome.Quantity, entry: reversalPrice(outcome.EntryProvenance), stop: reversalPrice(outcome.StopProvenance),
		target: reversalPrice(outcome.TargetProvenance), policy: ExecutionPolicy{identity: outcome.ExecutionPolicyDigest}, lineage: laneLineage{
			AccountRef: envelope.AccountRef, Market: strategyrouter.Market(lineage.Market), Symbol: envelope.Symbol,
			PositionGeneration: lineage.PositionGeneration, LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion,
			CandidateID: lineage.CandidateID, EvidenceDigest: lineage.MetricEvidenceDigest, ConfigDigest: lineage.ConfigDigest,
			CampaignID: lineage.CampaignID, LegOrdinal: lineage.LegOrdinal, PlannedCeiling: lineage.PlannedCeiling,
			RiskBudgetDigest: lineage.RiskBudgetDigest,
		}}
}

func evaluateWeeklyKR(input LaneInput) laneEvaluation {
	return fromWeekly(weeklyvaluelane.EvaluateKR(input.weeklyKR), input.weeklyKR)
}

func evaluateWeeklyUS(input LaneInput) laneEvaluation {
	return fromWeekly(weeklyvaluelane.EvaluateUS(input.weeklyUS), input.weeklyUS)
}

func fromWeekly(outcome weeklyvaluelane.Outcome, request weeklyvaluelane.EvaluationRequest) laneEvaluation {
	lineage := outcome.Lineage
	return laneEvaluation{accepted: outcome.Kind == weeklyvaluelane.OutcomeDecision && outcome.Code == "",
		nativeCode: string(outcome.Code), quantity: outcome.Quantity, entry: weeklyPrice(outcome.EntryProvenance), stop: weeklyPrice(outcome.StopProvenance),
		target: weeklyPrice(outcome.TargetProvenance), policy: weeklyPolicy(outcome.ExecutionPolicy), lineage: laneLineage{
			AccountRef: request.Plan.AccountRef(), Market: strategyrouter.Market(lineage.Market), Symbol: lineage.Symbol,
			PositionGeneration: lineage.PositionGeneration, LaneID: lineage.LaneID, LaneVersion: lineage.LaneVersion,
			CandidateID: lineage.CandidateID, EvidenceDigest: lineage.EvidenceDigest, ConfigDigest: lineage.ConfigDigest,
			CampaignID: lineage.CampaignID, LegOrdinal: lineage.PlannedLegOrdinal, PlannedCeiling: lineage.PlannedLegCeiling,
			RiskBudgetDigest: weeklyRiskBudgetDigest(lineage.PlanDigest, lineage.RiskBudgetMinor),
		}}
}

func continuationPrice(p continuationlane.PriceProvenance) PriceProvenance {
	return PriceProvenance{priceMinor: p.PriceMinor, source: p.Source, version: p.Version, digest: p.Digest, asOf: p.AsOf, currency: p.Currency, minorScale: p.MinorScale, unitVersion: p.UnitVersion}
}
func reversalPrice(p reversallane.PriceProvenance) PriceProvenance {
	return PriceProvenance{priceMinor: p.PriceMinor, source: p.Source, version: p.Version, digest: p.Digest, asOf: p.AsOf, currency: p.Currency, minorScale: p.MinorScale, unitVersion: p.UnitVersion}
}
func weeklyPrice(p weeklyvaluelane.PriceProvenance) PriceProvenance {
	return PriceProvenance{priceMinor: p.PriceMinor, source: p.Source, version: p.Version, digest: p.Digest, asOf: p.AsOf, currency: p.Currency, minorScale: p.MinorScale, unitVersion: p.UnitVersion}
}
func weeklyPolicy(p weeklyvaluelane.RRExecutionPolicy) ExecutionPolicy {
	return ExecutionPolicy{stagedTargetMinor: p.StagedTargetMinor, fairValueMinor: p.FairValueMinor, entryCostsMinor: p.EntryCostsMinor, exitCostsMinor: p.ExitCostsMinor, minimumRRPPM: p.MinimumRRPPM, decisionDigest: p.DecisionDigest, calendarDigest: p.CalendarDigest, capSnapshotID: p.CapSnapshotID, identity: p.Identity}
}

func weeklyRiskBudgetDigest(planDigest, riskBudgetMinor string) string {
	if planDigest == "" || riskBudgetMinor == "" {
		return ""
	}
	sum := sha256.Sum256([]byte("weekly-risk-v1\x00" + planDigest + "\x00" + riskBudgetMinor))
	return hex.EncodeToString(sum[:])
}
