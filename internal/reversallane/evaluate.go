package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
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

// ProposeKR returns the cap-free q_candidate for the KR lane. It cannot issue
// risk, journal, dispatch, or broker authority.
func ProposeKR(request KREvaluationRequest) EvaluationResult {
	metric := EvaluateKRMetric(request.Evidence, request.Config)
	return evaluateStage(request.Context, request.Evidence.CommonEnvelope, request.Config.StructuralWindow, request.Structure, metric, KRReversalLaneID, evaluationProposal)
}

// ProposeUS is the paired US q_candidate proposal path.
func ProposeUS(request USEvaluationRequest) EvaluationResult {
	metric := EvaluateUSMetric(request.Evidence, request.Config)
	return evaluateStage(request.Context, request.Evidence.CommonEnvelope, request.Config.StructuralWindow, request.Structure, metric, USReversalLaneID, evaluationProposal)
}

type evaluationStage uint8

const (
	evaluationAdmitted evaluationStage = iota
	evaluationProposal
)

func evaluate(context EvaluationContext, envelope CommonEnvelope, window time.Duration, structure StructuralConfirmation, metric MetricResult, laneID string) EvaluationResult {
	return evaluateStage(context, envelope, window, structure, metric, laneID, evaluationAdmitted)
}

func evaluateStage(context EvaluationContext, envelope CommonEnvelope, window time.Duration, structure StructuralConfirmation, metric MetricResult, laneID string, stage evaluationStage) EvaluationResult {
	lineage := decisionLineageForStage(context, envelope, structure, laneID, stage)
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
	if refusal := validateStop(context.Plan, envelope, context.SavedEffectiveStopMinor, context.StopCandidate); refusal != "" {
		return refuse(refusal)
	}
	if context.Leg.Ordinal < 1 || context.Leg.Ordinal > 3 || context.Leg.Cancelled || context.Leg.Expired {
		return refuse(RefusalLegTerminal)
	}
	if stage == evaluationAdmitted && !context.Cap.validAt(context.Plan, envelope.EvaluatedAt) {
		return refuse(RefusalCapInvalid)
	}
	if context.Leg.Ordinal == 3 {
		if structuralRefusal := ValidateStructure(structure, envelope, window); structuralRefusal != "" {
			return refuse(structuralRefusal)
		}
	}
	ceilings := context.Plan.LegCeilings()
	if context.Leg.FilledQuantity >= ceilings[context.Leg.Ordinal-1] {
		return refuse(RefusalLegTerminal)
	}
	quantity := ceilings[context.Leg.Ordinal-1] - context.Leg.FilledQuantity
	if stage == evaluationAdmitted {
		quantity = PlannedLegQuantity(context.Plan, context.Leg, context.Cap)
	}
	if quantity == 0 {
		return refuse(RefusalLegTerminal)
	}
	if stage == evaluationAdmitted {
		if context.Cap.ReservationQuantity != quantity {
			return refuse(RefusalCapInvalid)
		}
		if riskRefusal := AdmitRisk(context.Plan, context.Risk, context.Cap); riskRefusal != "" {
			return refuse(riskRefusal)
		}
	} else if riskRefusal := admitExistingRisk(context.Plan, context.Risk); riskRefusal != "" {
		return refuse(riskRefusal)
	}
	entryAuthority, stopAuthority, targetAuthority, policyDigest, termsOK := validatedExecutionTerms(context.Plan, envelope, context.ExecutionTerms, context.StopCandidate)
	if !termsOK {
		return refuse(RefusalExecutionTermsInvalid)
	}
	lineage = decisionLineageForStage(context, envelope, structure, laneID, stage)
	action := "ADD"
	if context.Leg.Ordinal == 1 {
		action = "ENTRY"
	}
	return EvaluationResult{Kind: OutcomeDecision, Action: action, Quantity: quantity, EntryPriceMinor: entryAuthority.PriceMinor, EffectiveStopMinor: stopAuthority.PriceMinor,
		TargetPriceMinor: targetAuthority.PriceMinor, EntryProvenance: entryAuthority, StopProvenance: stopAuthority, TargetProvenance: targetAuthority,
		ExecutionPolicyDigest: policyDigest, Lineage: lineage, CommonExitIndependent: true}
}

func admitExistingRisk(plan CampaignPlan, state RiskState) RefusalCode {
	if !plan.valid() || state.PlanDigest != plan.Digest() {
		return RefusalPlanInvalid
	}
	if state.Latches[LatchCampaignRiskOverage] || state.Latches[LatchUnknownActualRisk] {
		return RefusalRiskLatched
	}
	filled, filledOK := parseMinor(state.FilledMinor)
	held, heldOK := parseMinor(state.HeldMinor)
	budget, budgetOK := parseMinor(plan.request.RiskBudgetMinor)
	if !filledOK || !heldOK || !budgetOK {
		return RefusalArithmeticInvalid
	}
	used := new(big.Int).Add(filled, held)
	if used.BitLen() > maxRiskBits {
		return RefusalArithmeticInvalid
	}
	if used.Cmp(budget) > 0 {
		return RefusalRiskBudgetExceeded
	}
	return ""
}

func decisionLineageForStage(context EvaluationContext, envelope CommonEnvelope, structure StructuralConfirmation, laneID string, stage evaluationStage) DecisionLineage {
	lineage := decisionLineage(context, envelope, structure, laneID)
	if stage == evaluationProposal {
		lineage.CapSnapshotID = ""
		lineage.CapPolicyDigest = ""
	}
	return lineage
}

func validateStop(plan CampaignPlan, envelope CommonEnvelope, saved string, candidate StopCandidate) RefusalCode {
	if !candidate.valid(plan, envelope) {
		return RefusalStopRetreat
	}
	candidateMinor, candidateOK := parseMinor(candidate.priceMinor)
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
