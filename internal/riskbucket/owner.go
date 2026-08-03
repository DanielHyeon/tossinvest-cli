package riskbucket

import "time"

type OwnerKey struct {
	AccountID             string
	Market                Market
	Symbol                string
	ProspectiveGeneration string
}

type Owner struct {
	LaneID           string
	CampaignID       string
	ActualGeneration string
}

type OwnerClaim struct {
	Key        OwnerKey
	LaneID     string
	CampaignID string
}

type OwnerState struct {
	Owners map[OwnerKey]Owner
}

type OwnerAcquireResult struct {
	Acquired bool
	Reused   bool
}

func AcquireOwner(state OwnerState, claim OwnerClaim) (OwnerState, OwnerAcquireResult, error) {
	next := cloneOwnerState(state)
	if claim.Key.AccountID == "" || claim.Key.Symbol == "" || claim.Key.ProspectiveGeneration == "" ||
		(claim.Key.Market != MarketKR && claim.Key.Market != MarketUS) || claim.LaneID == "" || claim.CampaignID == "" {
		return cloneOwnerState(state), OwnerAcquireResult{}, refusal(RefusalOwnerConflict, "owner_identity", nil)
	}
	scopeCount := ownerScopeCount(next, claim.Key)
	if scopeCount > 1 {
		return cloneOwnerState(state), OwnerAcquireResult{}, refusal(RefusalReconstructionMismatch, "duplicate_owner_scope", nil)
	}
	if scopeCount == 1 {
		current, exact := next.Owners[claim.Key]
		if exact && current.LaneID == claim.LaneID && current.CampaignID == claim.CampaignID {
			return next, OwnerAcquireResult{Reused: true}, nil
		}
		return cloneOwnerState(state), OwnerAcquireResult{}, refusal(RefusalOwnerConflict, "owner", nil)
	}
	next.Owners[claim.Key] = Owner{LaneID: claim.LaneID, CampaignID: claim.CampaignID}
	return next, OwnerAcquireResult{Acquired: true}, nil
}

func BindActualGeneration(state OwnerState, key OwnerKey, actualGeneration string) (OwnerState, error) {
	next := cloneOwnerState(state)
	if ownerScopeCount(next, key) > 1 {
		return cloneOwnerState(state), refusal(RefusalReconstructionMismatch, "duplicate_owner_scope", nil)
	}
	owner, exists := next.Owners[key]
	if !exists {
		return cloneOwnerState(state), refusal(RefusalOwnerNotFound, "owner", nil)
	}
	if actualGeneration == "" {
		return cloneOwnerState(state), refusal(RefusalOwnerConflict, "actual_generation", nil)
	}
	if owner.ActualGeneration != "" && owner.ActualGeneration != actualGeneration {
		return cloneOwnerState(state), refusal(RefusalOwnerConflict, "actual_generation", nil)
	}
	owner.ActualGeneration = actualGeneration
	next.Owners[key] = owner
	return next, nil
}

type ClaimState string

const (
	ClaimClean   ClaimState = "CLEAN"
	ClaimActive  ClaimState = "ACTIVE"
	ClaimPending ClaimState = "PENDING"
	ClaimUnknown ClaimState = "UNKNOWN"
	ClaimStale   ClaimState = "STALE"
)

type ReleaseEvidence struct {
	OwnerKey         OwnerKey
	LaneID           string
	CampaignID       string
	ActualGeneration string
	EvaluatedAt      time.Time
	Attestation      ReleaseAttestation
	PositionClosed   bool
	PositionQuantity uint64
	PendingEntry     bool
	HeldMinor        string
	BrokerReconciled bool
	BrokerQuantity   uint64
	ProtectionOrder  ClaimState
	ProtectionSaga   ClaimState
	SellClaim        ClaimState
	SellMutation     ClaimState
	UnresolvedFill   ClaimState
}

type OwnerReleaseResult struct {
	Released        bool
	AlreadyReleased bool
	BlockingField   string
}

func ReleaseOwner(state OwnerState, key OwnerKey, evidence ReleaseEvidence) (OwnerState, OwnerReleaseResult, error) {
	next := cloneOwnerState(state)
	scopeCount := ownerScopeCount(next, key)
	if scopeCount > 1 {
		return cloneOwnerState(state), OwnerReleaseResult{}, refusal(RefusalReconstructionMismatch, "duplicate_owner_scope", nil)
	}
	owner, exists := next.Owners[key]
	if !exists && scopeCount == 0 {
		return next, OwnerReleaseResult{AlreadyReleased: true}, nil
	}
	if !exists {
		return cloneOwnerState(state), OwnerReleaseResult{}, refusal(RefusalOwnerConflict, "owner_generation", nil)
	}
	if evidence.OwnerKey != key || evidence.LaneID != owner.LaneID || evidence.CampaignID != owner.CampaignID ||
		owner.ActualGeneration == "" || evidence.ActualGeneration != owner.ActualGeneration {
		return cloneOwnerState(state), OwnerReleaseResult{}, refusal(RefusalReleaseEvidenceInvalid, "owner_binding", nil)
	}
	if !evidence.Attestation.validAt(evidence.EvaluatedAt, key, owner) {
		return cloneOwnerState(state), OwnerReleaseResult{}, refusal(RefusalReleaseEvidenceInvalid, "release_attestation", nil)
	}
	clean, blocker, err := cleanReleasePredicate(evidence)
	if err != nil {
		return cloneOwnerState(state), OwnerReleaseResult{}, err
	}
	if !clean {
		return next, OwnerReleaseResult{BlockingField: blocker}, nil
	}
	delete(next.Owners, key)
	return next, OwnerReleaseResult{Released: true}, nil
}

func ownerScopeCount(state OwnerState, target OwnerKey) int {
	count := 0
	for key := range state.Owners {
		if key.AccountID == target.AccountID && key.Market == target.Market && key.Symbol == target.Symbol {
			count++
		}
	}
	return count
}

func cleanReleasePredicate(evidence ReleaseEvidence) (bool, string, error) {
	if !evidence.PositionClosed {
		return false, "position_closed", nil
	}
	if evidence.PositionQuantity != 0 {
		return false, "position_quantity", nil
	}
	if evidence.PendingEntry {
		return false, "pending_entry", nil
	}
	held, err := parseMinor(evidence.HeldMinor, 0)
	if err != nil {
		return false, "held_minor", refusal(RefusalRiskCalculationInvalid, "held_minor", err)
	}
	if held.Sign() != 0 {
		return false, "held_minor", nil
	}
	if !evidence.BrokerReconciled {
		return false, "broker_reconciled", nil
	}
	if evidence.BrokerQuantity != 0 {
		return false, "broker_quantity", nil
	}
	claims := []struct {
		name  string
		state ClaimState
	}{
		{"protection_order", evidence.ProtectionOrder},
		{"protection_saga", evidence.ProtectionSaga},
		{"sell_claim", evidence.SellClaim},
		{"sell_mutation", evidence.SellMutation},
		{"unresolved_fill", evidence.UnresolvedFill},
	}
	for _, claim := range claims {
		if claim.state != ClaimClean {
			return false, claim.name, nil
		}
	}
	return true, "", nil
}

func cloneOwnerState(in OwnerState) OwnerState {
	out := OwnerState{Owners: make(map[OwnerKey]Owner, len(in.Owners))}
	for key, owner := range in.Owners {
		out.Owners[key] = owner
	}
	return out
}
