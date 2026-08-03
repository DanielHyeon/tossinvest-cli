package strategyflow

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// Evaluate is the fixed production composition. Test substitution remains
// package-private so callers cannot forge router or lane acceptance.
func Evaluate(request Request) Result {
	return evaluateWith(request, strategyrouter.Route, defaultRegistry())
}

func evaluateWith(request Request, route func(strategyrouter.RouteRequest) strategyrouter.RouteResult, lanes registry) Result {
	result := Result{CommonSafetyIndependent: true}
	if !validApproved(request.Approved) {
		result.Code = RefusalInvalidCandidate
		return result
	}
	lineage := candidateLineage(request.Approved, request.Router.Key)
	if !validScope(request.Router.Key, request.Approved) || route == nil {
		result.Code = RefusalInvalidScope
		result.Lineage = sealLineage(lineage)
		return result
	}

	routed := route(request.Router)
	if routed.Code != strategyrouter.RefusalNone {
		result.Code = RefusalRouter
		result.NativeCode = string(routed.Code)
		result.Lineage = sealLineage(lineage)
		return result
	}
	decision := routed.Decision
	if decision.Key != request.Router.Key || decision.LaneID == "" || decision.LaneVersion == "" || !validRouteLineage(decision) {
		result.Code = RefusalLineageMismatch
		result.Lineage = sealLineage(lineage)
		return result
	}
	descriptor, ok := canonicalDescriptor(decision)
	if !ok {
		result.Code = RefusalUnsupportedBinding
		result.Lineage = sealLineage(lineage)
		return result
	}
	lineage.RouterRelease = strategyrouter.RouterRelease
	lineage.Horizon = descriptor.Horizon
	lineage.LaneID = descriptor.LaneID
	lineage.LaneVersion = descriptor.LaneVersion
	lineage.LaneRelease = descriptor.Release
	lineage.RouterEvidenceDigest = decision.EvidenceDigest
	lineage.ConfigDigest = decision.ConfigDigest
	if !request.Lane.matches(descriptor) {
		result.Code = RefusalUnsupportedBinding
		result.Lineage = sealLineage(lineage)
		return result
	}
	binding, ok := lanes.lookup(descriptor)
	if !ok {
		result.Code = RefusalUnsupportedBinding
		result.Lineage = sealLineage(lineage)
		return result
	}

	evaluated := binding.evaluate(request.Lane)
	if !evaluated.accepted {
		result.Code = RefusalLane
		result.NativeCode = evaluated.nativeCode
		result.Lineage = sealLineage(lineage)
		return result
	}
	if !matchingLaneLineage(evaluated.lineage, lineage) {
		result.Code = RefusalLineageMismatch
		result.Lineage = sealLineage(lineage)
		return result
	}
	if decision.ExistingOwner && evaluated.lineage.CampaignID != decision.CampaignID {
		result.Code = RefusalLineageMismatch
		result.Lineage = sealLineage(lineage)
		return result
	}
	lineage.CampaignID = evaluated.lineage.CampaignID
	lineage.ConfigDigest = evaluated.lineage.ConfigDigest
	lineage.LaneEvidenceDigest = evaluated.lineage.EvidenceDigest
	lineage.LegOrdinal = evaluated.lineage.LegOrdinal
	lineage.PlannedCeiling = evaluated.lineage.PlannedCeiling
	lineage.RiskBudgetDigest = evaluated.lineage.RiskBudgetDigest
	lineage.ExecutionPolicyDigest = evaluated.policy.identity
	if evaluated.quantity == 0 || evaluated.lineage.CampaignID == "" || evaluated.lineage.LegOrdinal < 1 ||
		evaluated.lineage.PlannedCeiling == 0 || evaluated.lineage.RiskBudgetDigest == "" || evaluated.policy.identity == "" {
		result.Code = RefusalLineageIncomplete
		result.Lineage = sealLineage(lineage)
		return result
	}
	lineage.Complete = true
	completeLineage := sealLineage(lineage)
	terms, ok := sealExecutionTerms(completeLineage, evaluated)
	if !ok {
		lineage.Complete = false
		result.Code = RefusalExecutionTermsInvalid
		result.Lineage = sealLineage(lineage)
		return result
	}
	result.Quantity = evaluated.quantity
	result.ExecutionTerms = terms
	result.Lineage = completeLineage
	return result
}

func validRouteLineage(decision strategyrouter.RouteDecision) bool {
	if decision.ExistingOwner {
		return decision.CampaignID != "" && decision.EvidenceDigest == "" && decision.ConfigDigest == ""
	}
	return decision.CampaignID == "" && decision.EvidenceDigest != "" && decision.ConfigDigest != ""
}

func validApproved(approved strategy.ApprovedSnapshot) bool {
	return approved.Valid() && (approved.Market() == string(strategyrouter.MarketKR) || approved.Market() == string(strategyrouter.MarketUS)) &&
		approved.Symbol() != "" && approved.Symbol() == strings.ToUpper(strings.TrimSpace(approved.Symbol())) &&
		approved.State() != "" && approved.CandidateLifeID() != "" && approved.ThresholdVersion() != "" &&
		approved.SetDigest() != "" && approved.EvidenceDigest() != "" && approved.FirstSeenUnixNano() > 0 &&
		approved.LastSeenUnixNano() >= approved.FirstSeenUnixNano() && approved.ValidUntilUnixNano() > approved.LastSeenUnixNano() &&
		approved.ApprovedAtUnixNano() > 0
}

func validScope(key strategyrouter.OwnerKey, approved strategy.ApprovedSnapshot) bool {
	canonical, err := strategyrouter.NewOwnerKey(key.AccountRef, key.Market, key.Symbol, key.PositionGeneration)
	return err == nil && canonical == key && string(key.Market) == approved.Market() && key.Symbol == approved.Symbol()
}

func candidateLineage(approved strategy.ApprovedSnapshot, key strategyrouter.OwnerKey) Lineage {
	return Lineage{AccountRef: key.AccountRef, Market: key.Market, Symbol: key.Symbol, PositionGeneration: key.PositionGeneration,
		CandidateState: approved.State(), CandidateLifeID: approved.CandidateLifeID(), CandidateFirstSeenNS: approved.FirstSeenUnixNano(),
		CandidateLastSeenNS: approved.LastSeenUnixNano(), CandidateValidUntilNS: approved.ValidUntilUnixNano(), CandidateApprovedAtNS: approved.ApprovedAtUnixNano(),
		ThresholdVersion: approved.ThresholdVersion(), ThresholdSetDigest: approved.SetDigest(), CandidateEvidenceDigest: approved.EvidenceDigest()}
}

func matchingLaneLineage(got laneLineage, want Lineage) bool {
	return got.AccountRef == want.AccountRef && got.Market == want.Market && got.Symbol == want.Symbol &&
		got.PositionGeneration == want.PositionGeneration && got.LaneID == want.LaneID && got.LaneVersion == want.LaneVersion &&
		got.CandidateID == want.CandidateLifeID && got.EvidenceDigest != "" &&
		(want.RouterEvidenceDigest == "" || got.EvidenceDigest == want.RouterEvidenceDigest) && got.ConfigDigest != "" &&
		(want.ConfigDigest == "" || got.ConfigDigest == want.ConfigDigest)
}
