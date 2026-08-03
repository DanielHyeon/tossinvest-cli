package riskbucket

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrRiskSnapshotAuthorityUnavailable = errors.New("risk bucket: snapshot authority loader unavailable")
	ErrRiskSnapshotScopeMismatch        = errors.New("risk bucket: snapshot authority scope mismatch")
	ErrRiskSnapshotReferenceMismatch    = errors.New("risk bucket: snapshot journal reference mismatch")
	ErrRiskSnapshotBundleTampered       = errors.New("risk bucket: snapshot authority bundle tampered")
)

// RiskSnapshotScope is the exact market and strategy identity for one
// authoritative five-bucket read. AsOf is both the policy evaluation instant
// and the freshness boundary for every policy and snapshot in the bundle.
type RiskSnapshotScope struct {
	AccountID           string
	Market              Market
	Horizon             Horizon
	StrategyRiskID      string
	StrategyRiskVersion string
	Sector              string
	Symbol              string
	AccountCurrency     string
	QuoteCurrency       string
	AsOf                time.Time
}

// RiskSnapshotJournalReference identifies the immutable journal record that
// supplied one snapshot. It must repeat the exact evidence consumed by risk
// admission; caller labels are never accepted as authority.
type RiskSnapshotJournalReference struct {
	Key                BucketKey
	SnapshotID         string
	SnapshotDigest     string
	SnapshotVersion    string
	PolicyDigest       string
	PolicyObservedAt   time.Time
	PolicyFreshUntil   time.Time
	SnapshotObservedAt time.Time
	SnapshotFreshUntil time.Time
}

type RiskSnapshotAuthorityEntry struct {
	Bucket    BucketSnapshot
	Reference RiskSnapshotJournalReference
}

type riskSnapshotAuthorityMaterialEntry struct {
	Bucket        BucketSnapshot
	Reference     RiskSnapshotJournalReference
	referenceSeal string
}

type riskSnapshotAuthorityMaterial struct {
	Scope   RiskSnapshotScope
	Policy  ReservePolicy
	Entries []riskSnapshotAuthorityMaterialEntry
}

// riskSnapshotAuthoritySource is deliberately package-private. A production
// adapter must live beside this contract and mint sealed material; arbitrary
// callers cannot substitute a source string or construct authoritative input.
type riskSnapshotAuthoritySource interface {
	loadRiskSnapshotAuthority(context.Context, RiskSnapshotScope) (riskSnapshotAuthorityMaterial, error)
}

type RiskSnapshotAuthorityService struct {
	source riskSnapshotAuthoritySource
}

// NewRiskSnapshotAuthorityService fails closed until a package-owned
// production loader is implemented. It performs no journal, broker, activation
// or other mutation.
func NewRiskSnapshotAuthorityService() (*RiskSnapshotAuthorityService, error) {
	return nil, ErrRiskSnapshotAuthorityUnavailable
}

func newRiskSnapshotAuthorityService(source riskSnapshotAuthoritySource) *RiskSnapshotAuthorityService {
	return &RiskSnapshotAuthorityService{source: source}
}

func (s *RiskSnapshotAuthorityService) Load(ctx context.Context, scope RiskSnapshotScope) (RiskSnapshotAuthorityBundle, error) {
	if err := ctx.Err(); err != nil {
		return RiskSnapshotAuthorityBundle{}, err
	}
	if s == nil || s.source == nil {
		return RiskSnapshotAuthorityBundle{}, ErrRiskSnapshotAuthorityUnavailable
	}
	if err := validateRiskSnapshotScope(scope); err != nil {
		return RiskSnapshotAuthorityBundle{}, err
	}
	material, err := s.source.loadRiskSnapshotAuthority(ctx, scope)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, fmt.Errorf("load risk snapshot authority: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return RiskSnapshotAuthorityBundle{}, err
	}
	return sealRiskSnapshotAuthorityBundle(scope, material)
}

// RiskSnapshotAuthorityBundle contains one exact, ordered, immutable-by-copy
// authority snapshot. All fields remain private so callers can only obtain
// value copies.
type RiskSnapshotAuthorityBundle struct {
	scope   RiskSnapshotScope
	policy  ReservePolicy
	entries [5]riskSnapshotAuthorityMaterialEntry
	digest  string
}

func (b RiskSnapshotAuthorityBundle) Scope() RiskSnapshotScope { return b.scope }

func (b RiskSnapshotAuthorityBundle) Policy() ReservePolicy { return b.policy }

func (b RiskSnapshotAuthorityBundle) Digest() string { return b.digest }

func (b RiskSnapshotAuthorityBundle) Entries() []RiskSnapshotAuthorityEntry {
	out := make([]RiskSnapshotAuthorityEntry, len(b.entries))
	for i := range b.entries {
		out[i] = RiskSnapshotAuthorityEntry{Bucket: b.entries[i].Bucket, Reference: b.entries[i].Reference}
	}
	return out
}

func (b RiskSnapshotAuthorityBundle) Validate(scope RiskSnapshotScope) error {
	if err := validateRiskSnapshotScope(scope); err != nil {
		return err
	}
	if !sameRiskSnapshotScope(b.scope, scope) {
		return ErrRiskSnapshotScopeMismatch
	}
	material := riskSnapshotAuthorityMaterial{
		Scope:   b.scope,
		Policy:  b.policy,
		Entries: append([]riskSnapshotAuthorityMaterialEntry(nil), b.entries[:]...),
	}
	ordered, err := validateRiskSnapshotAuthorityMaterial(scope, material)
	if err != nil {
		return err
	}
	digest, err := digestRiskSnapshotAuthority(scope, b.policy, ordered)
	if err != nil {
		return err
	}
	if b.digest == "" || digest != b.digest {
		return ErrRiskSnapshotBundleTampered
	}
	return nil
}

func sealRiskSnapshotAuthorityBundle(scope RiskSnapshotScope, material riskSnapshotAuthorityMaterial) (RiskSnapshotAuthorityBundle, error) {
	ordered, err := validateRiskSnapshotAuthorityMaterial(scope, material)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, err
	}
	digest, err := digestRiskSnapshotAuthority(scope, material.Policy, ordered)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, err
	}
	bundle := RiskSnapshotAuthorityBundle{scope: scope, policy: material.Policy, digest: digest}
	copy(bundle.entries[:], ordered)
	return bundle, nil
}

func validateRiskSnapshotAuthorityMaterial(scope RiskSnapshotScope, material riskSnapshotAuthorityMaterial) ([]riskSnapshotAuthorityMaterialEntry, error) {
	if err := validateRiskSnapshotScope(material.Scope); err != nil || !sameRiskSnapshotScope(scope, material.Scope) {
		return nil, ErrRiskSnapshotScopeMismatch
	}
	if material.Policy.AccountCurrency != scope.AccountCurrency || material.Policy.QuoteCurrency != scope.QuoteCurrency ||
		!material.Policy.EvaluatedAt.Equal(scope.AsOf) {
		return nil, ErrRiskSnapshotScopeMismatch
	}
	if _, _, _, _, _, _, err := validateReservePolicy(material.Policy); err != nil {
		return nil, err
	}

	buckets, entriesByDimension, err := indexRiskSnapshotEntries(material.Entries)
	if err != nil {
		return nil, err
	}
	orderedBuckets, err := validateAndOrderBuckets(buckets, scope.AsOf, material.Policy.MaxDecimalBits)
	if err != nil {
		return nil, err
	}
	return validateOrderedRiskSnapshotEntries(scope, orderedBuckets, entriesByDimension)
}

func indexRiskSnapshotEntries(entries []riskSnapshotAuthorityMaterialEntry) ([]BucketSnapshot, map[Dimension]riskSnapshotAuthorityMaterialEntry, error) {
	buckets := make([]BucketSnapshot, len(entries))
	byDimension := make(map[Dimension]riskSnapshotAuthorityMaterialEntry, len(entries))
	for i, entry := range entries {
		buckets[i] = entry.Bucket
		if _, duplicate := byDimension[entry.Bucket.Key.Dimension]; duplicate {
			return nil, nil, refusal(RefusalMissingBucket, string(entry.Bucket.Key.Dimension), fmt.Errorf("duplicate dimension"))
		}
		byDimension[entry.Bucket.Key.Dimension] = entry
	}
	return buckets, byDimension, nil
}

func validateOrderedRiskSnapshotEntries(scope RiskSnapshotScope, buckets []BucketSnapshot, entries map[Dimension]riskSnapshotAuthorityMaterialEntry) ([]riskSnapshotAuthorityMaterialEntry, error) {
	ordered := make([]riskSnapshotAuthorityMaterialEntry, 0, len(requiredDimensions))
	seenSnapshotIDs := make(map[string]struct{}, len(requiredDimensions))
	for _, bucket := range buckets {
		entry := entries[bucket.Key.Dimension]
		if err := validateRiskSnapshotEntry(scope, entry); err != nil {
			return nil, err
		}
		if _, duplicate := seenSnapshotIDs[entry.Reference.SnapshotID]; duplicate {
			return nil, ErrRiskSnapshotReferenceMismatch
		}
		seenSnapshotIDs[entry.Reference.SnapshotID] = struct{}{}
		seal, err := digestRiskSnapshotMaterialEntry(entry.Bucket, entry.Reference)
		if err != nil || seal != entry.referenceSeal {
			return nil, ErrRiskSnapshotReferenceMismatch
		}
		ordered = append(ordered, entry)
	}
	return ordered, nil
}

func validateRiskSnapshotEntry(scope RiskSnapshotScope, entry riskSnapshotAuthorityMaterialEntry) error {
	expected, ok := expectedRiskSnapshotBucket(scope, entry.Bucket.Key.Dimension)
	if !ok || entry.Bucket.Key.Value != expected ||
		(entry.Bucket.Key.Dimension == DimensionStrategy && entry.Bucket.Key.PolicyVersion != scope.StrategyRiskVersion) {
		return ErrRiskSnapshotScopeMismatch
	}
	bound := entry.Bucket.BoundEvidence()
	ref := entry.Reference
	if ref.Key != entry.Bucket.Key || ref.SnapshotID == "" || !canonicalIdentity(ref.SnapshotID) ||
		ref.PolicyDigest != bound.PolicyEvidence.Digest ||
		ref.SnapshotDigest != bound.SnapshotEvidence.Digest ||
		ref.SnapshotVersion != entry.Bucket.SnapshotVersion ||
		!ref.PolicyObservedAt.Equal(bound.PolicyEvidence.ObservedAt) ||
		!ref.PolicyFreshUntil.Equal(bound.PolicyEvidence.FreshUntil) ||
		!ref.SnapshotObservedAt.Equal(bound.SnapshotEvidence.ObservedAt) ||
		!ref.SnapshotFreshUntil.Equal(bound.SnapshotEvidence.FreshUntil) {
		return ErrRiskSnapshotReferenceMismatch
	}
	return nil
}

func newRiskSnapshotAuthorityMaterialEntry(bucket BucketSnapshot, reference RiskSnapshotJournalReference) (riskSnapshotAuthorityMaterialEntry, error) {
	seal, err := digestRiskSnapshotMaterialEntry(bucket, reference)
	if err != nil {
		return riskSnapshotAuthorityMaterialEntry{}, err
	}
	return riskSnapshotAuthorityMaterialEntry{Bucket: bucket, Reference: reference, referenceSeal: seal}, nil
}

func validateRiskSnapshotScope(scope RiskSnapshotScope) error {
	if !canonicalIdentity(scope.AccountID) || !canonicalIdentity(scope.StrategyRiskID) ||
		!canonicalIdentity(scope.StrategyRiskVersion) || !canonicalIdentity(scope.Sector) ||
		!canonicalIdentity(scope.Symbol) || scope.AsOf.IsZero() || scope.AsOf.Location() != time.UTC {
		return ErrRiskSnapshotScopeMismatch
	}
	if scope.Market != MarketKR && scope.Market != MarketUS {
		return ErrRiskSnapshotScopeMismatch
	}
	if scope.Horizon != HorizonShort && scope.Horizon != HorizonMedium {
		return ErrRiskSnapshotScopeMismatch
	}
	if !canonicalCurrency(scope.AccountCurrency) || !canonicalCurrency(scope.QuoteCurrency) {
		return ErrRiskSnapshotScopeMismatch
	}
	if scope.Market == MarketKR && scope.QuoteCurrency != "KRW" {
		return ErrRiskSnapshotScopeMismatch
	}
	if scope.Market == MarketUS && scope.QuoteCurrency != "USD" {
		return ErrRiskSnapshotScopeMismatch
	}
	return nil
}

func canonicalIdentity(value string) bool {
	if value == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

func sameRiskSnapshotScope(a, b RiskSnapshotScope) bool {
	return a.AccountID == b.AccountID && a.Market == b.Market && a.Horizon == b.Horizon &&
		a.StrategyRiskID == b.StrategyRiskID && a.StrategyRiskVersion == b.StrategyRiskVersion &&
		a.Sector == b.Sector && a.Symbol == b.Symbol && a.AccountCurrency == b.AccountCurrency &&
		a.QuoteCurrency == b.QuoteCurrency && a.AsOf.Equal(b.AsOf)
}

func expectedRiskSnapshotBucket(scope RiskSnapshotScope, dimension Dimension) (string, bool) {
	switch dimension {
	case DimensionHorizon:
		return string(scope.Horizon), true
	case DimensionMarket:
		return string(scope.Market), true
	case DimensionStrategy:
		return scope.StrategyRiskID, true
	case DimensionSector:
		return scope.Sector, true
	case DimensionSymbol:
		return scope.Symbol, true
	default:
		return "", false
	}
}

type riskSnapshotEntryDigest struct {
	Binding          BucketSnapshotBinding
	PolicyEvidence   Evidence
	SnapshotEvidence Evidence
	Reference        RiskSnapshotJournalReference
}

func digestRiskSnapshotMaterialEntry(bucket BucketSnapshot, reference RiskSnapshotJournalReference) (string, error) {
	bound := bucket.BoundEvidence()
	return sha256JSON(riskSnapshotEntryDigest{
		Binding: bucket.binding(), PolicyEvidence: bound.PolicyEvidence,
		SnapshotEvidence: bound.SnapshotEvidence, Reference: reference,
	})
}

type riskSnapshotBundleDigest struct {
	Scope   RiskSnapshotScope
	Policy  ReservePolicy
	Entries []riskSnapshotEntryDigest
}

func digestRiskSnapshotAuthority(scope RiskSnapshotScope, policy ReservePolicy, entries []riskSnapshotAuthorityMaterialEntry) (string, error) {
	canonical := riskSnapshotBundleDigest{Scope: scope, Policy: policy, Entries: make([]riskSnapshotEntryDigest, len(entries))}
	for i, entry := range entries {
		bound := entry.Bucket.BoundEvidence()
		canonical.Entries[i] = riskSnapshotEntryDigest{
			Binding: entry.Bucket.binding(), PolicyEvidence: bound.PolicyEvidence,
			SnapshotEvidence: bound.SnapshotEvidence, Reference: entry.Reference,
		}
	}
	return sha256JSON(canonical)
}

func sha256JSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal risk snapshot authority: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
