package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func EvaluateKR(request KREvaluationRequest) EvaluationResult {
	metric := EvaluateKRMetric(request.Evidence, request.Config)
	return evaluate(request.Context, request.Evidence.CommonEnvelope, request.Config.StructuralWindow, request.Structure, metric, KRReversalLaneID)
}

func EvaluateUS(request USEvaluationRequest) EvaluationResult {
	metric := EvaluateUSMetric(request.Evidence, request.Config)
	return evaluate(request.Context, request.Evidence.CommonEnvelope, request.Config.StructuralWindow, request.Structure, metric, USReversalLaneID)
}

func evaluate(context EvaluationContext, envelope CommonEnvelope, window time.Duration, structure StructuralConfirmation, metric MetricResult, laneID string) EvaluationResult {
	lineage := decisionLineage(context, envelope, structure, laneID)
	refuse := func(code RefusalCode) EvaluationResult {
		return EvaluationResult{Kind: OutcomeRefusal, Code: code, Lineage: lineage, CommonExitIndependent: true}
	}
	if !context.Enabled {
		return refuse(RefusalDisabled)
	}
	if context.Invalidation.Structural || context.Invalidation.Code != "" {
		return EvaluationResult{Kind: OutcomeInvalidation, Code: RefusalInvalidated, Lineage: lineage, CommonExitIndependent: true}
	}
	if !context.Plan.valid() || context.CandidateID == "" || context.Plan.LaneID() != laneID || context.Plan.Market() != envelope.Market || context.Plan.ConfigDigest() != envelope.ConfigDigest || context.Plan.PositionGeneration() != envelope.PositionGeneration || context.Plan.request.AccountRef != envelope.AccountRef || context.Plan.request.Symbol != envelope.Symbol {
		return refuse(RefusalPlanInvalid)
	}
	if context.Plan.request.FX != nil && !context.Plan.request.FX.validAt(envelope.EvaluatedAt) {
		return refuse(RefusalPlanInvalid)
	}
	if !metric.Accepted {
		if metric.Refusal == "" {
			return refuse(RefusalStrictSchema)
		}
		return refuse(metric.Refusal)
	}
	if context.Risk.Latches[LatchCampaignRiskOverage] || context.Risk.Latches[LatchUnknownActualRisk] {
		return refuse(RefusalRiskLatched)
	}
	if refusal := validateStop(context.SavedEffectiveStopMinor, context.StopCandidate, envelope.EvaluatedAt); refusal != "" {
		return refuse(refusal)
	}
	if context.Leg.Ordinal < 1 || context.Leg.Ordinal > 3 || context.Leg.Cancelled || context.Leg.Expired {
		return refuse(RefusalLegTerminal)
	}
	if !context.Cap.validAt(context.Plan, envelope.EvaluatedAt) {
		return refuse(RefusalCapInvalid)
	}
	if context.Leg.Ordinal == 3 {
		if structuralRefusal := ValidateStructure(structure, envelope, window); structuralRefusal != "" {
			return refuse(structuralRefusal)
		}
	}
	quantity := PlannedLegQuantity(context.Plan, context.Leg, context.Cap)
	if quantity == 0 {
		return refuse(RefusalLegTerminal)
	}
	if context.Cap.ReservationQuantity != quantity {
		return refuse(RefusalCapInvalid)
	}
	if riskRefusal := AdmitRisk(context.Plan, context.Risk, context.Cap); riskRefusal != "" {
		return refuse(riskRefusal)
	}
	entryPrice, effectiveStop, targetPrice, termsOK := validatedExecutionTerms(context.Plan, context.ExecutionTerms, context.StopCandidate.PriceMinor)
	if !termsOK {
		return refuse(RefusalExecutionTermsInvalid)
	}
	lineage = decisionLineage(context, envelope, structure, laneID)
	action := "ADD"
	if context.Leg.Ordinal == 1 {
		action = "ENTRY"
	}
	return EvaluationResult{Kind: OutcomeDecision, Action: action, Quantity: quantity, EntryPriceMinor: entryPrice, EffectiveStopMinor: effectiveStop,
		TargetPriceMinor: targetPrice, Lineage: lineage, CommonExitIndependent: true}
}

func validateStop(saved string, candidate StopCandidate, evaluatedAt time.Time) RefusalCode {
	if !candidate.Valid || candidate.Source == "" || candidate.Policy == "" || candidate.Digest == "" || candidate.ObservedAt.IsZero() || candidate.ObservedAt.After(evaluatedAt) {
		return RefusalStopRetreat
	}
	candidateMinor, candidateOK := parseMinor(candidate.PriceMinor)
	if !candidateOK {
		return RefusalStopRetreat
	}
	if saved == "" {
		return ""
	}
	savedMinor, savedOK := parseMinor(saved)
	if !savedOK || candidateMinor.Cmp(savedMinor) < 0 {
		return RefusalStopRetreat
	}
	return ""
}

func decisionLineage(context EvaluationContext, envelope CommonEnvelope, structure StructuralConfirmation, laneID string) DecisionLineage {
	ceiling := uint64(0)
	if context.Leg.Ordinal >= 1 && context.Leg.Ordinal <= 3 {
		ceiling = context.Plan.LegCeilings()[context.Leg.Ordinal-1]
	}
	structureDigest := ""
	if structure.Sweep.RecordID != "" || structure.Break.RecordID != "" || structure.Reclaim.RecordID != "" {
		structureDigest = structuralDigest(structure)
	}
	return DecisionLineage{
		Market: envelope.Market, LaneID: laneID, LaneVersion: LaneVersionV1, CandidateID: context.CandidateID,
		SchemaVersion: envelope.SchemaVersion, ConfigDigest: envelope.ConfigDigest, MetricEvidenceDigest: envelope.SourceDigest,
		StructuralDigest: structureDigest, CampaignID: context.Plan.CampaignID(), PositionGeneration: envelope.PositionGeneration,
		RiskBudgetDigest: riskBudgetDigest(context.Plan), LegOrdinal: context.Leg.Ordinal, PlannedCeiling: ceiling,
		CapSnapshotID: context.Cap.SnapshotID, CapPolicyDigest: context.Cap.PolicyDigest,
	}
}

func riskBudgetDigest(plan CampaignPlan) string {
	if !plan.valid() {
		return ""
	}
	preimage := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s", plan.Digest(), plan.request.RiskBudgetMinor, plan.request.PerShareRiskMinor, plan.request.PlannedQuantity, plan.request.PolicyDigest)
	sum := sha256.Sum256([]byte(preimage))
	return hex.EncodeToString(sum[:])
}
