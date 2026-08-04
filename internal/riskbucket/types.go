// Package riskbucket provides pure, fail-closed calculations and state
// transitions for multi-dimensional monetary risk buckets. Persistence and
// runtime admission wiring deliberately live outside this leaf package.
package riskbucket

import (
	"errors"
	"fmt"
	"time"
)

type Horizon string

const (
	HorizonShort  Horizon = "SHORT"
	HorizonMedium Horizon = "MEDIUM"
)

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type Dimension string

const (
	DimensionHorizon  Dimension = "horizon"
	DimensionMarket   Dimension = "market"
	DimensionStrategy Dimension = "strategy"
	DimensionSector   Dimension = "sector"
	DimensionSymbol   Dimension = "symbol"
)

var requiredDimensions = [...]Dimension{
	DimensionHorizon,
	DimensionMarket,
	DimensionStrategy,
	DimensionSector,
	DimensionSymbol,
}

func RequiredDimensionOrder() []Dimension {
	out := make([]Dimension, len(requiredDimensions))
	copy(out, requiredDimensions[:])
	return out
}

type BucketKey struct {
	Dimension     Dimension
	Value         string
	PolicyVersion string
}

type RefusalCode string

const (
	RefusalZeroQuantity              RefusalCode = "ZERO_QUANTITY"
	RefusalMissingBucket             RefusalCode = "MISSING_BUCKET"
	RefusalStaleBucket               RefusalCode = "STALE_BUCKET"
	RefusalUnknownHorizon            RefusalCode = "UNKNOWN_HORIZON"
	RefusalUnknownMarket             RefusalCode = "UNKNOWN_MARKET"
	RefusalUnknownStrategyBucket     RefusalCode = "UNKNOWN_STRATEGY_BUCKET"
	RefusalUnknownSector             RefusalCode = "UNKNOWN_SECTOR"
	RefusalUnknownSymbol             RefusalCode = "UNKNOWN_SYMBOL"
	RefusalWorstPriceUnavailable     RefusalCode = "WORST_PRICE_UNAVAILABLE"
	RefusalCurrencyUnresolved        RefusalCode = "CURRENCY_UNRESOLVED"
	RefusalInvalidFXHaircut          RefusalCode = "INVALID_FX_HAIRCUT"
	RefusalFeePolicyUnavailable      RefusalCode = "FEE_POLICY_UNAVAILABLE"
	RefusalRiskCalculationInvalid    RefusalCode = "RISK_CALCULATION_INVALID"
	RefusalBucketCapExhausted        RefusalCode = "BUCKET_CAP_EXHAUSTED"
	RefusalExistingGuardianCap       RefusalCode = "EXISTING_GUARDIAN_CAP"
	RefusalOwnerConflict             RefusalCode = "OWNER_CONFLICT"
	RefusalOwnerNotFound             RefusalCode = "OWNER_NOT_FOUND"
	RefusalReconstructionMismatch    RefusalCode = "RECONSTRUCTION_MISMATCH"
	RefusalPolicyProvenanceInvalid   RefusalCode = "POLICY_PROVENANCE_INVALID"
	RefusalSnapshotProvenanceInvalid RefusalCode = "SNAPSHOT_PROVENANCE_INVALID"
	RefusalReleaseEvidenceInvalid    RefusalCode = "RELEASE_EVIDENCE_INVALID"
	RefusalFillEvidenceInconsistent  RefusalCode = "FILL_EVIDENCE_INCONSISTENT"
)

type RefusalError struct {
	Code  RefusalCode
	Field string
	Cause error
}

func (e *RefusalError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s (%s): %v", e.Code, e.Field, e.Cause)
	}
	if e.Field != "" {
		return fmt.Sprintf("%s (%s)", e.Code, e.Field)
	}
	return string(e.Code)
}

func (e *RefusalError) Unwrap() error { return e.Cause }

func IsRefusal(err error, code RefusalCode) bool {
	var refusal *RefusalError
	return errors.As(err, &refusal) && refusal.Code == code
}

func refusal(code RefusalCode, field string, cause error) *RefusalError {
	return &RefusalError{Code: code, Field: field, Cause: cause}
}

type Evidence struct {
	Source     string
	Version    string
	Digest     string
	Official   bool
	Frozen     bool
	ObservedAt time.Time
	FreshUntil time.Time
}

func (e Evidence) validAt(at time.Time) bool {
	return e.Source != "" && e.Version != "" && e.Digest != "" && e.Official && e.Frozen &&
		!at.IsZero() && !e.ObservedAt.IsZero() && !e.FreshUntil.IsZero() &&
		!e.ObservedAt.After(at) && !e.FreshUntil.Before(at)
}

const (
	RiskPolicyAuthoritySource   = "risk-policy-authority"
	RiskSnapshotAuthoritySource = "risk-snapshot-authority"
	OwnerReleaseAuthoritySource = "owner-release-authority"
)

// PolicyProvenance is an immutable, authority-typed policy attestation. Its
// fields are intentionally private so callers cannot mutate a validated value.
type PolicyProvenance struct {
	evidence Evidence
	key      BucketKey
}

func NewPolicyProvenance(key BucketKey, evidence Evidence) (PolicyProvenance, error) {
	if err := validateFrozenAuthorityEvidence(evidence, RiskPolicyAuthoritySource); err != nil {
		return PolicyProvenance{}, refusal(RefusalPolicyProvenanceInvalid, "policy_provenance", err)
	}
	if err := validateBucketIdentity(key); err != nil || evidence.Version != key.PolicyVersion {
		return PolicyProvenance{}, refusal(RefusalPolicyProvenanceInvalid, "policy_binding", err)
	}
	return PolicyProvenance{evidence: evidence, key: key}, nil
}

func (p PolicyProvenance) validAt(at time.Time, key BucketKey) bool {
	return p.key == key && p.evidence.Source == RiskPolicyAuthoritySource && p.evidence.Version == key.PolicyVersion && p.evidence.validAt(at)
}

// SnapshotProvenance is an immutable, authority-typed bucket snapshot
// attestation. Snapshot freshness is derived only from this value.
type SnapshotProvenance struct {
	evidence Evidence
	binding  BucketSnapshotBinding
}

func NewSnapshotProvenance(binding BucketSnapshotBinding, evidence Evidence) (SnapshotProvenance, error) {
	if err := validateFrozenAuthorityEvidence(evidence, RiskSnapshotAuthoritySource); err != nil {
		return SnapshotProvenance{}, refusal(RefusalSnapshotProvenanceInvalid, "snapshot_provenance", err)
	}
	if err := validateBucketIdentity(binding.Key); err != nil || binding.SnapshotVersion == "" || evidence.Version != binding.SnapshotVersion {
		return SnapshotProvenance{}, refusal(RefusalSnapshotProvenanceInvalid, "snapshot_binding", err)
	}
	if _, err := parseMinor(binding.LimitMinor, 0); err != nil {
		return SnapshotProvenance{}, refusal(RefusalSnapshotProvenanceInvalid, "snapshot_limit", err)
	}
	if _, err := parseMinor(binding.FilledMinor, 0); err != nil {
		return SnapshotProvenance{}, refusal(RefusalSnapshotProvenanceInvalid, "snapshot_filled", err)
	}
	if _, err := parseMinor(binding.HeldMinor, 0); err != nil {
		return SnapshotProvenance{}, refusal(RefusalSnapshotProvenanceInvalid, "snapshot_held", err)
	}
	return SnapshotProvenance{evidence: evidence, binding: binding}, nil
}

func (p SnapshotProvenance) validAt(at time.Time, binding BucketSnapshotBinding) bool {
	return p.binding == binding && p.evidence.Source == RiskSnapshotAuthoritySource && p.evidence.Version == binding.SnapshotVersion && p.evidence.validAt(at)
}

// ReleaseAttestation is immutable, authority-typed evidence for one owner
// release evaluation.
type ReleaseAttestation struct {
	evidence         Evidence
	ownerKey         OwnerKey
	laneID           string
	campaignID       string
	actualGeneration string
}

func NewReleaseAttestation(key OwnerKey, owner Owner, evidence Evidence) (ReleaseAttestation, error) {
	if err := validateFrozenAuthorityEvidence(evidence, OwnerReleaseAuthoritySource); err != nil {
		return ReleaseAttestation{}, refusal(RefusalReleaseEvidenceInvalid, "release_attestation", err)
	}
	if key.AccountID == "" || (key.Market != MarketKR && key.Market != MarketUS) || key.Symbol == "" || key.ProspectiveGeneration == "" ||
		owner.LaneID == "" || owner.CampaignID == "" || owner.ActualGeneration == "" {
		return ReleaseAttestation{}, refusal(RefusalReleaseEvidenceInvalid, "release_binding", nil)
	}
	return ReleaseAttestation{evidence: evidence, ownerKey: key, laneID: owner.LaneID, campaignID: owner.CampaignID, actualGeneration: owner.ActualGeneration}, nil
}

func (a ReleaseAttestation) validAt(at time.Time, key OwnerKey, owner Owner) bool {
	return a.ownerKey == key && a.laneID == owner.LaneID && a.campaignID == owner.CampaignID && a.actualGeneration == owner.ActualGeneration &&
		a.evidence.Source == OwnerReleaseAuthoritySource && a.evidence.validAt(at)
}

func validateFrozenAuthorityEvidence(evidence Evidence, source string) error {
	if evidence.Source != source || evidence.Version == "" || evidence.Digest == "" || !evidence.Official || !evidence.Frozen ||
		evidence.ObservedAt.IsZero() || evidence.FreshUntil.IsZero() || evidence.ObservedAt.After(evidence.FreshUntil) {
		return fmt.Errorf("invalid frozen authority evidence")
	}
	return nil
}
