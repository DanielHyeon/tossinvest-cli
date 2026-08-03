package riskbucket

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fixedRiskSnapshotAuthoritySource struct {
	material riskSnapshotAuthorityMaterial
	err      error
}

func (s fixedRiskSnapshotAuthoritySource) loadRiskSnapshotAuthority(context.Context, RiskSnapshotScope) (riskSnapshotAuthorityMaterial, error) {
	return s.material, s.err
}

func TestRiskSnapshotAuthorityBundleLoadsExactFiveDimensionsForKRAndUS(t *testing.T) {
	for _, market := range []Market{MarketKR, MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			scope, material := riskSnapshotAuthorityFixture(t, market)
			service := newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{material: material})
			bundle, err := service.Load(context.Background(), scope)
			if err != nil {
				t.Fatal(err)
			}
			if err := bundle.Validate(scope); err != nil {
				t.Fatalf("validate sealed bundle: %v", err)
			}
			if bundle.Digest() == "" || bundle.Scope() != scope || bundle.Policy() != material.Policy {
				t.Fatalf("bundle scope=%+v digest=%q policy=%+v", bundle.Scope(), bundle.Digest(), bundle.Policy())
			}
			entries := bundle.Entries()
			if len(entries) != 5 {
				t.Fatalf("entries=%d want=5", len(entries))
			}
			for i, dimension := range RequiredDimensionOrder() {
				if entries[i].Bucket.Key.Dimension != dimension || entries[i].Reference.Key != entries[i].Bucket.Key {
					t.Fatalf("entry[%d]=%+v dimension=%s", i, entries[i], dimension)
				}
				bound := entries[i].Bucket.BoundEvidence()
				if entries[i].Reference.PolicyDigest != bound.PolicyEvidence.Digest ||
					entries[i].Reference.SnapshotDigest != bound.SnapshotEvidence.Digest ||
					entries[i].Reference.SnapshotVersion != entries[i].Bucket.SnapshotVersion {
					t.Fatalf("entry[%d] journal reference is not exact: %+v bound=%+v", i, entries[i].Reference, bound)
				}
			}
		})
	}
}

func TestRiskSnapshotAuthorityBundleRejectsIncompleteStaleMismatchedOrTamperedMaterial(t *testing.T) {
	baseScope, base := riskSnapshotAuthorityFixture(t, MarketKR)
	tests := []struct {
		name   string
		mutate func(*RiskSnapshotScope, *riskSnapshotAuthorityMaterial)
		check  func(error) bool
	}{
		{
			name: "missing bucket",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Entries = material.Entries[:4]
			},
			check: func(err error) bool { return IsRefusal(err, RefusalMissingBucket) },
		},
		{
			name: "duplicate dimension",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Entries[4] = material.Entries[0]
			},
			check: func(err error) bool { return IsRefusal(err, RefusalMissingBucket) },
		},
		{
			name: "stale snapshot",
			mutate: func(scope *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				entry := &material.Entries[2]
				binding := entry.Bucket.binding()
				evidence := entry.Bucket.BoundEvidence().SnapshotEvidence
				evidence.FreshUntil = scope.AsOf.Add(-time.Second)
				entry.Bucket.SnapshotProvenance = mustSnapshotAuthority(t, binding, evidence)
				entry.Reference.SnapshotFreshUntil = evidence.FreshUntil
			},
			check: func(err error) bool { return IsRefusal(err, RefusalStaleBucket) },
		},
		{
			name: "stale policy",
			mutate: func(scope *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				entry := &material.Entries[1]
				evidence := entry.Bucket.BoundEvidence().PolicyEvidence
				evidence.FreshUntil = scope.AsOf.Add(-time.Second)
				provenance, err := NewPolicyProvenance(entry.Bucket.Key, evidence)
				if err != nil {
					t.Fatal(err)
				}
				entry.Bucket.PolicyProvenance = provenance
				entry.Reference.PolicyFreshUntil = evidence.FreshUntil
			},
			check: func(err error) bool { return IsRefusal(err, RefusalPolicyProvenanceInvalid) },
		},
		{
			name: "stale reserve price evidence",
			mutate: func(scope *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Policy.Price.FreshUntil = scope.AsOf.Add(-time.Second)
			},
			check: func(err error) bool { return IsRefusal(err, RefusalWorstPriceUnavailable) },
		},
		{
			name: "currency scope mismatch",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Policy.QuoteCurrency = "USD"
			},
			check: func(err error) bool { return errors.Is(err, ErrRiskSnapshotScopeMismatch) },
		},
		{
			name: "snapshot monetary tamper",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Entries[4].Bucket.LimitMinor = "999999"
			},
			check: func(err error) bool { return IsRefusal(err, RefusalSnapshotProvenanceInvalid) },
		},
		{
			name: "journal policy digest tamper",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Entries[1].Reference.PolicyDigest = "caller-substitution"
			},
			check: func(err error) bool { return errors.Is(err, ErrRiskSnapshotReferenceMismatch) },
		},
		{
			name: "journal snapshot identity tamper",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Entries[3].Reference.SnapshotID = "different-snapshot"
			},
			check: func(err error) bool { return errors.Is(err, ErrRiskSnapshotReferenceMismatch) },
		},
		{
			name: "duplicate journal snapshot identity",
			mutate: func(_ *RiskSnapshotScope, material *riskSnapshotAuthorityMaterial) {
				material.Entries[4].Reference.SnapshotID = material.Entries[3].Reference.SnapshotID
				entry, err := newRiskSnapshotAuthorityMaterialEntry(material.Entries[4].Bucket, material.Entries[4].Reference)
				if err != nil {
					t.Fatal(err)
				}
				material.Entries[4] = entry
			},
			check: func(err error) bool { return errors.Is(err, ErrRiskSnapshotReferenceMismatch) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope := baseScope
			material := cloneRiskSnapshotAuthorityMaterial(base)
			tc.mutate(&scope, &material)
			service := newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{material: material})
			if _, err := service.Load(context.Background(), scope); err == nil || !tc.check(err) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestRiskSnapshotAuthorityBundleRejectsKRUSCrossReuse(t *testing.T) {
	krScope, krMaterial := riskSnapshotAuthorityFixture(t, MarketKR)
	usScope, _ := riskSnapshotAuthorityFixture(t, MarketUS)
	service := newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{material: krMaterial})
	if _, err := service.Load(context.Background(), usScope); !errors.Is(err, ErrRiskSnapshotScopeMismatch) {
		t.Fatalf("KR material reused for US error=%v", err)
	}
	bundle, err := service.Load(context.Background(), krScope)
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.Validate(usScope); !errors.Is(err, ErrRiskSnapshotScopeMismatch) {
		t.Fatalf("KR bundle validated for US error=%v", err)
	}
}

func TestRiskSnapshotAuthorityBundleReturnsImmutableCopies(t *testing.T) {
	scope, material := riskSnapshotAuthorityFixture(t, MarketUS)
	bundle, err := newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{material: material}).Load(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	originalDigest := bundle.Digest()
	entries := bundle.Entries()
	entries[0].Bucket.Key.Value = "tampered"
	entries[0].Reference.SnapshotDigest = "tampered"
	policy := bundle.Policy()
	policy.AccountCurrency = "XXX"
	if bundle.Digest() != originalDigest || bundle.Entries()[0].Bucket.Key.Value == "tampered" || bundle.Policy().AccountCurrency == "XXX" {
		t.Fatal("returned value copies mutated sealed bundle")
	}
	if err := bundle.Validate(scope); err != nil {
		t.Fatalf("copy mutation invalidated bundle: %v", err)
	}
	bundle.digest = "sha256:tampered"
	if err := bundle.Validate(scope); !errors.Is(err, ErrRiskSnapshotBundleTampered) {
		t.Fatalf("tampered bundle digest error=%v", err)
	}
}

func TestRiskSnapshotAuthorityServiceRejectsInvalidScopeAndUnavailableSource(t *testing.T) {
	scope, material := riskSnapshotAuthorityFixture(t, MarketKR)
	service := newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{material: material})
	tests := []struct {
		name   string
		mutate func(*RiskSnapshotScope)
	}{
		{name: "empty account", mutate: func(scope *RiskSnapshotScope) { scope.AccountID = "" }},
		{name: "control character", mutate: func(scope *RiskSnapshotScope) { scope.Symbol = "005930\n" }},
		{name: "unknown market", mutate: func(scope *RiskSnapshotScope) { scope.Market = Market("CN") }},
		{name: "unknown horizon", mutate: func(scope *RiskSnapshotScope) { scope.Horizon = Horizon("LONG") }},
		{name: "noncanonical currency", mutate: func(scope *RiskSnapshotScope) { scope.AccountCurrency = "krw" }},
		{name: "wrong KR quote currency", mutate: func(scope *RiskSnapshotScope) { scope.QuoteCurrency = "USD" }},
		{name: "non UTC as-of", mutate: func(scope *RiskSnapshotScope) { scope.AsOf = scope.AsOf.In(time.FixedZone("KST", 9*60*60)) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			invalid := scope
			tc.mutate(&invalid)
			if _, err := service.Load(context.Background(), invalid); !errors.Is(err, ErrRiskSnapshotScopeMismatch) {
				t.Fatalf("invalid scope error=%v", err)
			}
		})
	}
	var nilService *RiskSnapshotAuthorityService
	if _, err := nilService.Load(context.Background(), scope); !errors.Is(err, ErrRiskSnapshotAuthorityUnavailable) {
		t.Fatalf("nil service error=%v", err)
	}
}

func TestProductionRiskSnapshotAuthorityConstructorFailsClosedWithoutLoader(t *testing.T) {
	service, err := NewRiskSnapshotAuthorityService()
	if service != nil || !errors.Is(err, ErrRiskSnapshotAuthorityUnavailable) {
		t.Fatalf("service=%v err=%v", service, err)
	}
}

func TestRiskSnapshotAuthorityServicePropagatesCancellationAndSourceFailure(t *testing.T) {
	scope, material := riskSnapshotAuthorityFixture(t, MarketKR)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	service := newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{material: material})
	if _, err := service.Load(cancelled, scope); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled load error=%v", err)
	}
	sourceErr := errors.New("snapshot store unavailable")
	service = newRiskSnapshotAuthorityService(fixedRiskSnapshotAuthoritySource{err: sourceErr})
	if _, err := service.Load(context.Background(), scope); !errors.Is(err, sourceErr) {
		t.Fatalf("source error=%v", err)
	}
}

func riskSnapshotAuthorityFixture(t *testing.T, market Market) (RiskSnapshotScope, riskSnapshotAuthorityMaterial) {
	t.Helper()
	now := time.Date(2026, 8, 4, 10, 0, 0, 0, time.UTC)
	scope := RiskSnapshotScope{
		AccountID: "acct-authority", Market: market, Horizon: HorizonShort,
		StrategyRiskID: "continuation", StrategyRiskVersion: "strategy-v1", Sector: "technology",
		AccountCurrency: "KRW", AsOf: now,
	}
	if market == MarketKR {
		scope.Symbol, scope.QuoteCurrency = "005930", "KRW"
	} else {
		scope.Symbol, scope.QuoteCurrency = "AAPL", "USD"
	}
	policy := ReservePolicy{
		AccountCurrency: scope.AccountCurrency, QuoteCurrency: scope.QuoteCurrency, EvaluatedAt: now,
		Price: PriceEvidence{WorstExecutableQuote: "100", Evidence: Evidence{Source: "official-order-contract", Version: "price-v1", Digest: "price-digest-" + string(market), Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)}},
		Fee:   FeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "1", MinimumBaseMinor: "1", Version: "fee-v1", Digest: "fee-digest-" + string(market)},
	}
	if market == MarketKR {
		policy.FX = FXEvidence{RateQuoteToBase: "1", Haircut: "1", Evidence: Evidence{Source: "same-currency", Version: "fx-v1", Digest: "fx-digest-KR", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)}}
	} else {
		policy.FX = FXEvidence{RateQuoteToBase: "1400", Haircut: "1.05", Evidence: Evidence{Source: "official-fx", Version: "fx-v1", Digest: "fx-digest-US", Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)}}
	}
	values := map[Dimension]string{
		DimensionHorizon: string(scope.Horizon), DimensionMarket: string(scope.Market),
		DimensionStrategy: scope.StrategyRiskID, DimensionSector: scope.Sector, DimensionSymbol: scope.Symbol,
	}
	entries := make([]riskSnapshotAuthorityMaterialEntry, 0, 5)
	for _, dimension := range RequiredDimensionOrder() {
		version := "policy-v1"
		if dimension == DimensionStrategy {
			version = scope.StrategyRiskVersion
		}
		key := BucketKey{Dimension: dimension, Value: values[dimension], PolicyVersion: version}
		policyEvidence := Evidence{Source: RiskPolicyAuthoritySource, Version: version, Digest: "policy-" + string(market) + "-" + string(dimension), Official: true, Frozen: true, ObservedAt: now.Add(-time.Minute), FreshUntil: now.Add(time.Minute)}
		policyProvenance, err := NewPolicyProvenance(key, policyEvidence)
		if err != nil {
			t.Fatal(err)
		}
		binding := BucketSnapshotBinding{Key: key, LimitMinor: "1000000", FilledMinor: "100", HeldMinor: "50", SnapshotVersion: "snapshot-" + string(market) + "-" + string(dimension) + "-v1"}
		snapshotEvidence := Evidence{Source: RiskSnapshotAuthoritySource, Version: binding.SnapshotVersion, Digest: "snapshot-digest-" + string(market) + "-" + string(dimension), Official: true, Frozen: true, ObservedAt: now.Add(-30 * time.Second), FreshUntil: now.Add(time.Minute)}
		snapshotProvenance, err := NewSnapshotProvenance(binding, snapshotEvidence)
		if err != nil {
			t.Fatal(err)
		}
		entry, err := newRiskSnapshotAuthorityMaterialEntry(
			BucketSnapshot{Key: key, LimitMinor: binding.LimitMinor, FilledMinor: binding.FilledMinor, HeldMinor: binding.HeldMinor, SnapshotVersion: binding.SnapshotVersion, PolicyProvenance: policyProvenance, SnapshotProvenance: snapshotProvenance},
			RiskSnapshotJournalReference{
				Key: key, SnapshotID: "journal-" + binding.SnapshotVersion, SnapshotDigest: snapshotEvidence.Digest,
				SnapshotVersion: binding.SnapshotVersion, PolicyDigest: policyEvidence.Digest,
				PolicyObservedAt: policyEvidence.ObservedAt, PolicyFreshUntil: policyEvidence.FreshUntil,
				SnapshotObservedAt: snapshotEvidence.ObservedAt, SnapshotFreshUntil: snapshotEvidence.FreshUntil,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		entries = append(entries, entry)
	}
	return scope, riskSnapshotAuthorityMaterial{Scope: scope, Policy: policy, Entries: entries}
}

func mustSnapshotAuthority(t *testing.T, binding BucketSnapshotBinding, evidence Evidence) SnapshotProvenance {
	t.Helper()
	provenance, err := NewSnapshotProvenance(binding, evidence)
	if err != nil {
		t.Fatal(err)
	}
	return provenance
}

func cloneRiskSnapshotAuthorityMaterial(in riskSnapshotAuthorityMaterial) riskSnapshotAuthorityMaterial {
	out := in
	out.Entries = append([]riskSnapshotAuthorityMaterialEntry(nil), in.Entries...)
	return out
}
