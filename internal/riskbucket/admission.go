package riskbucket

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"
)

type BucketSnapshot struct {
	Key                BucketKey
	LimitMinor         string
	FilledMinor        string
	HeldMinor          string
	SnapshotVersion    string
	PolicyProvenance   PolicyProvenance
	SnapshotProvenance SnapshotProvenance
}

type BucketSnapshotBinding struct {
	Key             BucketKey
	LimitMinor      string
	FilledMinor     string
	HeldMinor       string
	SnapshotVersion string
}

func (b BucketSnapshot) binding() BucketSnapshotBinding {
	return BucketSnapshotBinding{
		Key:             b.Key,
		LimitMinor:      b.LimitMinor,
		FilledMinor:     b.FilledMinor,
		HeldMinor:       b.HeldMinor,
		SnapshotVersion: b.SnapshotVersion,
	}
}

type AdmissionRequest struct {
	QCandidate        uint64
	QExistingGuardian uint64
	Policy            ReservePolicy
	Buckets           []BucketSnapshot
}

type BucketCap struct {
	Key                        BucketKey
	AvailableMinor             string
	MaxQuantityWithinCandidate uint64
	CandidateFullyAdmitted     bool
	ReservationAtFinal         string
	SnapshotVersion            string
}

type AdmissionDecision struct {
	QCandidate        uint64
	QExistingGuardian uint64
	QFinal            uint64
	PolicyPreimage    ReservePolicy
	Caps              []BucketCap
	Binding           []BucketKey
	Refusal           *RefusalError
}

func CalculateAdmission(request AdmissionRequest) AdmissionDecision {
	decision := AdmissionDecision{QCandidate: request.QCandidate, QExistingGuardian: request.QExistingGuardian, PolicyPreimage: request.Policy}
	if request.QCandidate == 0 {
		decision.Refusal = refusal(RefusalZeroQuantity, "q_candidate", nil)
		return decision
	}
	if request.QExistingGuardian == 0 {
		decision.Refusal = refusal(RefusalExistingGuardianCap, "q_existing_guardian", nil)
		return decision
	}
	ordered, err := validateAndOrderBuckets(request.Buckets, request.Policy.EvaluatedAt, request.Policy.MaxDecimalBits)
	if err != nil {
		decision.Refusal = asRefusal(err)
		return decision
	}
	if _, _, _, _, _, _, err := validateReservePolicy(request.Policy); err != nil {
		decision.Refusal = asRefusal(err)
		return decision
	}
	qFinal := minUint64(request.QCandidate, request.QExistingGuardian)
	for _, bucket := range ordered {
		available, err := availableMinor(bucket, request.Policy.MaxDecimalBits)
		if err != nil {
			decision.Refusal = asRefusal(err)
			decision.Caps = nil
			return decision
		}
		capQuantity, err := MaximumQuantity(available, request.QCandidate, request.Policy)
		if err != nil {
			decision.Refusal = asRefusal(err)
			decision.Caps = nil
			return decision
		}
		decision.Caps = append(decision.Caps, BucketCap{
			Key:                        bucket.Key,
			AvailableMinor:             available,
			MaxQuantityWithinCandidate: capQuantity,
			CandidateFullyAdmitted:     capQuantity == request.QCandidate,
			SnapshotVersion:            bucket.SnapshotVersion,
		})
		qFinal = minUint64(qFinal, capQuantity)
	}
	decision.QFinal = qFinal
	if qFinal == 0 {
		decision.Refusal = refusal(RefusalBucketCapExhausted, "monetary_bucket", nil)
		return decision
	}
	for i := range decision.Caps {
		reserve, err := ReservationMinor(qFinal, request.Policy)
		if err != nil {
			decision.QFinal = 0
			decision.Refusal = asRefusal(err)
			return decision
		}
		decision.Caps[i].ReservationAtFinal = reserve
		if !decision.Caps[i].CandidateFullyAdmitted && decision.Caps[i].MaxQuantityWithinCandidate == qFinal {
			decision.Binding = append(decision.Binding, decision.Caps[i].Key)
		}
	}
	return decision
}

func validateAndOrderBuckets(buckets []BucketSnapshot, at time.Time, maxBits uint) ([]BucketSnapshot, error) {
	byDimension := make(map[Dimension]BucketSnapshot, len(buckets))
	for _, bucket := range buckets {
		if _, exists := byDimension[bucket.Key.Dimension]; exists {
			return nil, refusal(RefusalMissingBucket, string(bucket.Key.Dimension), fmt.Errorf("duplicate dimension"))
		}
		byDimension[bucket.Key.Dimension] = bucket
	}
	ordered := make([]BucketSnapshot, 0, len(requiredDimensions))
	for _, dimension := range requiredDimensions {
		bucket, exists := byDimension[dimension]
		if !exists {
			return nil, refusal(RefusalMissingBucket, string(dimension), nil)
		}
		if err := validateBucketIdentity(bucket.Key); err != nil {
			return nil, err
		}
		if !bucket.PolicyProvenance.validAt(at, bucket.Key) {
			return nil, refusal(RefusalPolicyProvenanceInvalid, string(dimension), nil)
		}
		if bucket.SnapshotVersion == "" {
			return nil, refusal(RefusalSnapshotProvenanceInvalid, string(dimension), nil)
		}
		if !bucket.SnapshotProvenance.validAt(at, bucket.binding()) {
			if !at.IsZero() && bucket.SnapshotProvenance.evidence.Source == RiskSnapshotAuthoritySource && bucket.SnapshotProvenance.evidence.Version == bucket.SnapshotVersion && bucket.SnapshotProvenance.evidence.FreshUntil.Before(at) {
				return nil, refusal(RefusalStaleBucket, string(dimension), nil)
			}
			return nil, refusal(RefusalSnapshotProvenanceInvalid, string(dimension), nil)
		}
		if _, err := availableMinor(bucket, maxBits); err != nil {
			return nil, err
		}
		ordered = append(ordered, bucket)
	}
	if len(byDimension) != len(requiredDimensions) {
		return nil, refusal(RefusalMissingBucket, "unexpected_dimension", nil)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return dimensionRank(ordered[i].Key.Dimension) < dimensionRank(ordered[j].Key.Dimension)
	})
	return ordered, nil
}

func validateBucketIdentity(key BucketKey) error {
	if key.Value == "" {
		switch key.Dimension {
		case DimensionHorizon:
			return refusal(RefusalUnknownHorizon, "horizon", nil)
		case DimensionMarket:
			return refusal(RefusalUnknownMarket, "market", nil)
		case DimensionStrategy:
			return refusal(RefusalUnknownStrategyBucket, "strategy", nil)
		case DimensionSector:
			return refusal(RefusalUnknownSector, "sector", nil)
		case DimensionSymbol:
			return refusal(RefusalUnknownSymbol, "symbol", nil)
		default:
			return refusal(RefusalMissingBucket, "dimension", nil)
		}
	}
	if key.PolicyVersion == "" {
		if key.Dimension == DimensionStrategy {
			return refusal(RefusalUnknownStrategyBucket, "strategy_policy_version", nil)
		}
		return refusal(RefusalStaleBucket, string(key.Dimension)+"_policy_version", nil)
	}
	if key.Dimension == DimensionHorizon && key.Value != string(HorizonShort) && key.Value != string(HorizonMedium) {
		return refusal(RefusalUnknownHorizon, "horizon", nil)
	}
	if key.Dimension == DimensionMarket && key.Value != string(MarketKR) && key.Value != string(MarketUS) {
		return refusal(RefusalUnknownMarket, "market", nil)
	}
	return nil
}

func availableMinor(bucket BucketSnapshot, maxBits uint) (string, error) {
	limit, err := parseMinor(bucket.LimitMinor, maxBits)
	if err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, string(bucket.Key.Dimension)+".limit", err)
	}
	filled, err := parseMinor(bucket.FilledMinor, maxBits)
	if err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, string(bucket.Key.Dimension)+".filled", err)
	}
	held, err := parseMinor(bucket.HeldMinor, maxBits)
	if err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, string(bucket.Key.Dimension)+".held", err)
	}
	used, err := addMinor(filled, held, maxBits)
	if err != nil {
		return "", refusal(RefusalRiskCalculationInvalid, string(bucket.Key.Dimension)+".usage", err)
	}
	remaining := new(big.Int).Sub(limit, used)
	if remaining.Sign() < 0 {
		return "0", nil
	}
	return remaining.String(), nil
}

func dimensionRank(dimension Dimension) int {
	for i, required := range requiredDimensions {
		if required == dimension {
			return i
		}
	}
	return len(requiredDimensions)
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func asRefusal(err error) *RefusalError {
	var typed *RefusalError
	if errors.As(err, &typed) {
		return typed
	}
	return refusal(RefusalRiskCalculationInvalid, "internal", err)
}
