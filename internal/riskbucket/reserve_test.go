package riskbucket

import (
	"errors"
	"math/rand"
	"testing"
	"time"
)

var testNow = time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

func identityPolicy() ReservePolicy {
	return ReservePolicy{
		AccountCurrency: "KRW",
		QuoteCurrency:   "KRW",
		EvaluatedAt:     testNow,
		Price: PriceEvidence{
			WorstExecutableQuote: "5",
			Evidence:             Evidence{Source: "official-order-contract", Version: "v1", Digest: "price-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)},
		},
		FX: FXEvidence{
			RateQuoteToBase: "1",
			Haircut:         "1",
			Evidence:        Evidence{Source: "same-currency", Version: "v1", Digest: "fx-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)},
		},
		Fee: FeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "0", MinimumBaseMinor: "0", Version: "fee-v1", Digest: "fee-digest"},
	}
}

func TestReservationMinorUsesWorstPriceFXHaircutMinimumFeeAndCeil(t *testing.T) {
	p := identityPolicy()
	p.AccountCurrency = "KRW"
	p.QuoteCurrency = "USD"
	p.Price.WorstExecutableQuote = "10.01"
	p.FX.RateQuoteToBase = "1.2"
	p.FX.Haircut = "1.05"
	p.Fee = FeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "0.1", MinimumBaseMinor: "2.5", Version: "fee-v2", Digest: "fee-minimum"}

	got, err := ReservationMinor(3, p)
	if err != nil {
		t.Fatalf("ReservationMinor: %v", err)
	}
	if got != "41" {
		t.Fatalf("reserve = %s, want 41", got)
	}
}

func TestReservationMinorRejectsUnsafeEvidence(t *testing.T) {
	tests := []struct {
		name string
		edit func(*ReservePolicy)
		code RefusalCode
	}{
		{"stale price", func(p *ReservePolicy) { p.Price.FreshUntil = testNow.Add(-time.Second) }, RefusalWorstPriceUnavailable},
		{"unofficial price", func(p *ReservePolicy) { p.Price.Official = false }, RefusalWorstPriceUnavailable},
		{"unfrozen fx", func(p *ReservePolicy) { p.FX.Frozen = false }, RefusalCurrencyUnresolved},
		{"missing fx digest", func(p *ReservePolicy) { p.FX.Digest = "" }, RefusalCurrencyUnresolved},
		{"haircut below one", func(p *ReservePolicy) { p.FX.Haircut = "0.99" }, RefusalInvalidFXHaircut},
		{"same currency nonidentity fx", func(p *ReservePolicy) { p.FX.RateQuoteToBase = "1.01" }, RefusalCurrencyUnresolved},
		{"missing fee policy", func(p *ReservePolicy) { p.Fee.Digest = "" }, RefusalFeePolicyUnavailable},
		{"decimal overflow", func(p *ReservePolicy) {
			p.Price.WorstExecutableQuote = "340282366920938463463374607431768211456"
			p.MaxDecimalBits = 64
		}, RefusalRiskCalculationInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := identityPolicy()
			tt.edit(&p)
			_, err := ReservationMinor(1, p)
			var refusal *RefusalError
			if !errors.As(err, &refusal) || refusal.Code != tt.code {
				t.Fatalf("error = %v, want refusal %s", err, tt.code)
			}
		})
	}
}

func TestMaximumQuantityUsesExactNonlinearFee(t *testing.T) {
	p := identityPolicy()
	p.Fee = FeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "0", MinimumBaseMinor: "6", Version: "fee-v2", Digest: "minimum-six"}

	got, err := MaximumQuantity("21", 10, p)
	if err != nil {
		t.Fatalf("MaximumQuantity: %v", err)
	}
	if got != 3 {
		t.Fatalf("cap = %d, want 3", got)
	}
}

func TestAdmissionIntersectsFiveBucketCapsAndGuardian(t *testing.T) {
	p := identityPolicy()
	buckets := testBuckets("100")
	for i := range buckets {
		if buckets[i].Key.Dimension == DimensionSector {
			buckets[i].LimitMinor = "15"
			reattestSnapshot(&buckets[i])
		}
	}

	decision := CalculateAdmission(AdmissionRequest{QCandidate: 10, QExistingGuardian: 9, Policy: p, Buckets: buckets})
	if decision.Refusal != nil {
		t.Fatalf("unexpected refusal: %v", decision.Refusal)
	}
	if decision.QCandidate != 10 || decision.QFinal != 3 {
		t.Fatalf("quantities = (%d,%d), want (10,3)", decision.QCandidate, decision.QFinal)
	}
	if len(decision.Binding) != 1 || decision.Binding[0].Dimension != DimensionSector {
		t.Fatalf("binding = %#v, want sector", decision.Binding)
	}
	if len(decision.Caps) != len(RequiredDimensionOrder()) {
		t.Fatalf("caps = %d, want %d", len(decision.Caps), len(RequiredDimensionOrder()))
	}
}

func TestAdmissionDoesNotMisreportCandidateClippedBucketAsBinding(t *testing.T) {
	decision := CalculateAdmission(AdmissionRequest{QCandidate: 10, QExistingGuardian: 10, Policy: identityPolicy(), Buckets: testBuckets("1000")})
	if decision.Refusal != nil || decision.QFinal != 10 {
		t.Fatalf("decision = %#v", decision)
	}
	if len(decision.Binding) != 0 {
		t.Fatalf("spacious candidate-clipped buckets reported as binding: %#v", decision.Binding)
	}
	for _, cap := range decision.Caps {
		if !cap.CandidateFullyAdmitted || cap.MaxQuantityWithinCandidate != 10 {
			t.Fatalf("cap preimage = %#v", cap)
		}
	}
}

func TestAdmissionSubtractsFilledAndHeldUsage(t *testing.T) {
	buckets := testBuckets("100")
	for i := range buckets {
		buckets[i].FilledMinor = "20"
		buckets[i].HeldMinor = "30"
		reattestSnapshot(&buckets[i])
	}
	decision := CalculateAdmission(AdmissionRequest{QCandidate: 20, QExistingGuardian: 20, Policy: identityPolicy(), Buckets: buckets})
	if decision.Refusal != nil || decision.QFinal != 10 {
		t.Fatalf("decision = %#v, want remaining 50 / reserve 5 = 10", decision)
	}
}

func TestCrossCurrencyReservationRequiresOfficialFrozenFXAndRoundsUp(t *testing.T) {
	p := identityPolicy()
	p.AccountCurrency = "KRW"
	p.QuoteCurrency = "USD"
	p.Price.WorstExecutableQuote = "0.01"
	p.FX.RateQuoteToBase = "1300.25"
	p.FX.Haircut = "1.02"

	got, err := ReservationMinor(1, p)
	if err != nil {
		t.Fatal(err)
	}
	if got != "14" {
		t.Fatalf("reserve = %s, want ceil(13.26255)=14", got)
	}
}

func TestAdmissionFailsClosedForUnknownDimensionAndStaleBucket(t *testing.T) {
	tests := []struct {
		name string
		edit func([]BucketSnapshot) []BucketSnapshot
		code RefusalCode
	}{
		{"unknown sector", func(b []BucketSnapshot) []BucketSnapshot { b[3].Key.Value = ""; return b }, RefusalUnknownSector},
		{"unknown strategy", func(b []BucketSnapshot) []BucketSnapshot { b[2].Key.PolicyVersion = ""; return b }, RefusalUnknownStrategyBucket},
		{"missing market bucket", func(b []BucketSnapshot) []BucketSnapshot { return append(b[:1], b[2:]...) }, RefusalMissingBucket},
		{"stale symbol bucket", func(b []BucketSnapshot) []BucketSnapshot {
			b[4].SnapshotProvenance = mustSnapshotProvenance(b[4].binding(), Evidence{Source: RiskSnapshotAuthoritySource, Version: "snap-1", Digest: "snapshot-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(-time.Second)})
			return b
		}, RefusalStaleBucket},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := CalculateAdmission(AdmissionRequest{QCandidate: 10, QExistingGuardian: 10, Policy: identityPolicy(), Buckets: tt.edit(testBuckets("100"))})
			if decision.QFinal != 0 || decision.Refusal == nil || decision.Refusal.Code != tt.code {
				t.Fatalf("decision = %#v, want q_final=0 refusal=%s", decision, tt.code)
			}
		})
	}
}

func TestAdmissionRejectsMissingOrMismatchedAuthoritativeProvenance(t *testing.T) {
	tests := []struct {
		name string
		edit func([]BucketSnapshot) []BucketSnapshot
		code RefusalCode
	}{
		{"missing policy provenance", func(b []BucketSnapshot) []BucketSnapshot {
			b[2].PolicyProvenance = PolicyProvenance{}
			return b
		}, RefusalPolicyProvenanceInvalid},
		{"strategy identity substitution", func(b []BucketSnapshot) []BucketSnapshot {
			b[2].Key.Value = "strategy-beta"
			return b
		}, RefusalPolicyProvenanceInvalid},
		{"policy version mismatch", func(b []BucketSnapshot) []BucketSnapshot {
			otherKey := b[2].Key
			otherKey.PolicyVersion = "strategy-v8"
			b[2].PolicyProvenance = mustPolicyProvenance(otherKey, Evidence{Source: RiskPolicyAuthoritySource, Version: "strategy-v8", Digest: "policy-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)})
			return b
		}, RefusalPolicyProvenanceInvalid},
		{"missing snapshot provenance", func(b []BucketSnapshot) []BucketSnapshot {
			b[4].SnapshotProvenance = SnapshotProvenance{}
			return b
		}, RefusalSnapshotProvenanceInvalid},
		{"snapshot monetary mutation", func(b []BucketSnapshot) []BucketSnapshot {
			b[4].LimitMinor = "999999"
			return b
		}, RefusalSnapshotProvenanceInvalid},
		{"snapshot version mismatch", func(b []BucketSnapshot) []BucketSnapshot {
			otherBinding := b[4].binding()
			otherBinding.SnapshotVersion = "snap-2"
			b[4].SnapshotProvenance = mustSnapshotProvenance(otherBinding, Evidence{Source: RiskSnapshotAuthoritySource, Version: "snap-2", Digest: "snapshot-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)})
			return b
		}, RefusalSnapshotProvenanceInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := CalculateAdmission(AdmissionRequest{QCandidate: 10, QExistingGuardian: 10, Policy: identityPolicy(), Buckets: tt.edit(testBuckets("100"))})
			if decision.QFinal != 0 || decision.Refusal == nil || decision.Refusal.Code != tt.code {
				t.Fatalf("decision = %#v, want q_final=0 refusal=%s", decision, tt.code)
			}
		})
	}
}

func TestProvenanceConstructorsRejectCallerDeclaredAuthority(t *testing.T) {
	key := BucketKey{Dimension: DimensionStrategy, Value: "strategy-alpha", PolicyVersion: "strategy-v7"}
	evidence := Evidence{Source: "caller-declared", Version: key.PolicyVersion, Digest: "digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)}
	if _, err := NewPolicyProvenance(key, evidence); !IsRefusal(err, RefusalPolicyProvenanceInvalid) {
		t.Fatalf("caller-declared policy authority accepted: %v", err)
	}
	binding := BucketSnapshotBinding{Key: key, LimitMinor: "100", FilledMinor: "0", HeldMinor: "0", SnapshotVersion: "snap-1"}
	evidence.Source = RiskSnapshotAuthoritySource
	evidence.Version = binding.SnapshotVersion
	evidence.Frozen = false
	if _, err := NewSnapshotProvenance(binding, evidence); !IsRefusal(err, RefusalSnapshotProvenanceInvalid) {
		t.Fatalf("unfrozen snapshot authority accepted: %v", err)
	}
}

func TestAdmissionQuantityNeverIncreasesProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(6601))
	for i := 0; i < 2_000; i++ {
		candidate := uint64(rng.Intn(500) + 1)
		guardian := uint64(rng.Intn(500))
		buckets := testBuckets("100000")
		for j := range buckets {
			buckets[j].LimitMinor = uintString(uint64(rng.Intn(100_000)))
			reattestSnapshot(&buckets[j])
		}
		decision := CalculateAdmission(AdmissionRequest{QCandidate: candidate, QExistingGuardian: guardian, Policy: identityPolicy(), Buckets: buckets})
		if decision.QFinal > candidate || decision.QFinal > guardian {
			t.Fatalf("iteration %d increased quantity: candidate=%d guardian=%d final=%d", i, candidate, guardian, decision.QFinal)
		}
		for _, cap := range decision.Caps {
			if decision.QFinal > cap.MaxQuantityWithinCandidate {
				t.Fatalf("iteration %d exceeded %s cap: final=%d cap=%d", i, cap.Key.Dimension, decision.QFinal, cap.MaxQuantityWithinCandidate)
			}
		}
	}
}

func FuzzReservationIsMonotone(f *testing.F) {
	for _, seed := range []uint64{0, 1, 2, 99, 1_000_000} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, q uint64) {
		if q == ^uint64(0) {
			t.Skip()
		}
		p := identityPolicy()
		a, err := ReservationMinor(q, p)
		if err != nil {
			t.Fatalf("reserve(%d): %v", q, err)
		}
		b, err := ReservationMinor(q+1, p)
		if err != nil {
			t.Fatalf("reserve(%d): %v", q+1, err)
		}
		if compareMinor(a, b) > 0 {
			t.Fatalf("reserve decreased: q=%d a=%s b=%s", q, a, b)
		}
	})
}

func testBuckets(limit string) []BucketSnapshot {
	keys := []BucketKey{
		{Dimension: DimensionHorizon, Value: string(HorizonShort), PolicyVersion: "h-v1"},
		{Dimension: DimensionMarket, Value: string(MarketKR), PolicyVersion: "m-v1"},
		{Dimension: DimensionStrategy, Value: "strategy-alpha", PolicyVersion: "strategy-v7"},
		{Dimension: DimensionSector, Value: "semiconductor", PolicyVersion: "sector-v2"},
		{Dimension: DimensionSymbol, Value: "005930", PolicyVersion: "symbol-v3"},
	}
	out := make([]BucketSnapshot, 0, len(keys))
	for _, key := range keys {
		bucket := BucketSnapshot{Key: key, LimitMinor: limit, FilledMinor: "0", HeldMinor: "0", SnapshotVersion: "snap-1"}
		bucket.PolicyProvenance = mustPolicyProvenance(key, Evidence{Source: RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: "policy-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)})
		bucket.SnapshotProvenance = mustSnapshotProvenance(bucket.binding(), Evidence{Source: RiskSnapshotAuthoritySource, Version: "snap-1", Digest: "snapshot-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)})
		out = append(out, bucket)
	}
	return out
}

func mustPolicyProvenance(key BucketKey, evidence Evidence) PolicyProvenance {
	provenance, err := NewPolicyProvenance(key, evidence)
	if err != nil {
		panic(err)
	}
	return provenance
}

func mustSnapshotProvenance(binding BucketSnapshotBinding, evidence Evidence) SnapshotProvenance {
	provenance, err := NewSnapshotProvenance(binding, evidence)
	if err != nil {
		panic(err)
	}
	return provenance
}

func reattestSnapshot(bucket *BucketSnapshot) {
	bucket.SnapshotProvenance = mustSnapshotProvenance(bucket.binding(), Evidence{Source: RiskSnapshotAuthoritySource, Version: bucket.SnapshotVersion, Digest: "snapshot-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)})
}
