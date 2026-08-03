package execgw_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestQFinalPrecheckCapsKRByGuardianAndAllMonetaryBucketsBeforeAtomicIssuance(t *testing.T) {
	rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
		options.NewID = fixedIDs("qfinal-kr", "qfinal-kr-nonce")
	})
	request := qFinalKRRequest(t, rig, "kr", 20)
	precheck, err := rig.guardian.PrecheckQFinalEntry(request)
	if err != nil {
		t.Fatal(err)
	}
	if precheck.QCandidate() != 20 || precheck.QFinal() != 10 || len(precheck.BindingCaps()) != 5 || rig.collections != 0 {
		t.Fatalf("candidate=%d final=%d binding=%v collections=%d", precheck.QCandidate(), precheck.QFinal(), precheck.BindingCaps(), rig.collections)
	}
	issued, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
	if err != nil {
		t.Fatal(err)
	}
	if issued.Decision.ID != "qfinal-kr" || issued.RiskBucketReceipt.QFinal != 10 || len(issued.RiskBucketReceipt.ReservationIDs) != 5 || len(issued.Reservations) != 1 {
		t.Fatalf("issued=%+v", issued)
	}
	if required, err := rig.journal.RevalidateQFinalAdmission(context.Background(), issued.Decision.ID); err != nil || !required {
		t.Fatalf("required=%v err=%v", required, err)
	}
}

func TestQFinalPrecheckUSUSDWithKRWGuardianFailsCurrencyUnresolvedAndCollectsNothing(t *testing.T) {
	rig := newGuardian(t, nil)
	request := qFinalKRRequest(t, rig, "us-unresolved", 10)
	request.Market = "US"
	request.Currency = "USD"
	request.Symbol = "AAPL"
	request.Admission.Owner.Key.Market = riskbucket.MarketUS
	request.Admission.Owner.Key.Symbol = "AAPL"
	for i := range request.Admission.Admission.Buckets {
		if request.Admission.Admission.Buckets[i].Key.Dimension == riskbucket.DimensionMarket || request.Admission.Admission.Buckets[i].Key.Dimension == riskbucket.DimensionSymbol {
			// The refusal is intentionally before monetary authority is consumed:
			// existing Guardian limits are still KRW and cannot cap raw USD safely.
			request.Admission.Admission.Buckets[i] = riskbucket.BucketSnapshot{}
		}
	}
	_, err := rig.guardian.PrecheckQFinalEntry(request)
	var refused *execgw.QFinalRefusal
	if !errors.As(err, &refused) || refused.Code != riskbucket.RefusalCurrencyUnresolved || rig.collections != 0 {
		t.Fatalf("refusal=%+v collections=%d err=%v", refused, rig.collections, err)
	}
}

func TestQFinalPrecheckRejectsCrossedMarketCurrencyPairs(t *testing.T) {
	rig := newGuardian(t, nil)
	for _, tc := range []struct {
		market, currency string
		code             riskbucket.RefusalCode
	}{
		{"KR", "USD", riskbucket.RefusalCurrencyUnresolved},
		{"US", "KRW", riskbucket.RefusalCurrencyUnresolved},
		{"CN", "CNY", riskbucket.RefusalUnknownMarket},
	} {
		request := qFinalKRRequest(t, rig, tc.market+tc.currency, 10)
		request.Market, request.Currency = tc.market, tc.currency
		_, err := rig.guardian.PrecheckQFinalEntry(request)
		var refused *execgw.QFinalRefusal
		if !errors.As(err, &refused) || refused.Code != tc.code {
			t.Fatalf("%s/%s refusal=%+v err=%v", tc.market, tc.currency, refused, err)
		}
	}
}

func TestQFinalPrecheckSealsValidSameQFinalEvidenceSubstitution(t *testing.T) {
	rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
		options.NewID = fixedIDs("qfinal-mutated", "qfinal-mutated-nonce")
	})
	request := qFinalKRRequest(t, rig, "mutated", 20)
	precheck, err := rig.guardian.PrecheckQFinalEntry(request)
	if err != nil {
		t.Fatal(err)
	}
	// Substitute a fully valid Horizon authority with a different immutable
	// policy/snapshot identity but the same limit and therefore the same q_final.
	// A q_final-only seal would miss this attack; the evidence slices themselves
	// must have been copied before the opaque precheck was returned.
	bucket := request.Admission.Admission.Buckets[0]
	bucket.Key.PolicyVersion = "policy-substituted-v2"
	policyEvidence := riskbucket.Evidence{Source: riskbucket.RiskPolicyAuthoritySource, Version: bucket.Key.PolicyVersion, Digest: "policy-substituted", Official: true, Frozen: true, ObservedAt: fixedNow.Add(-time.Second), FreshUntil: fixedNow.Add(time.Minute)}
	bucket.PolicyProvenance, err = riskbucket.NewPolicyProvenance(bucket.Key, policyEvidence)
	if err != nil {
		t.Fatal(err)
	}
	bucket.SnapshotVersion = "snapshot-substituted-v2"
	binding := riskbucket.BucketSnapshotBinding{Key: bucket.Key, LimitMinor: bucket.LimitMinor, FilledMinor: bucket.FilledMinor, HeldMinor: bucket.HeldMinor, SnapshotVersion: bucket.SnapshotVersion}
	snapshotEvidence := riskbucket.Evidence{Source: riskbucket.RiskSnapshotAuthoritySource, Version: bucket.SnapshotVersion, Digest: "snapshot-substituted", Official: true, Frozen: true, ObservedAt: fixedNow.Add(-time.Second), FreshUntil: fixedNow.Add(time.Minute)}
	bucket.SnapshotProvenance, err = riskbucket.NewSnapshotProvenance(binding, snapshotEvidence)
	if err != nil {
		t.Fatal(err)
	}
	request.Admission.Admission.Buckets[0] = bucket
	request.Admission.Snapshots[0] = journal.RiskBucketSnapshotReference{Key: bucket.Key, SnapshotID: "snapshot-substituted", SnapshotDigest: snapshotEvidence.Digest, SnapshotVersion: bucket.SnapshotVersion, PolicyDigest: policyEvidence.Digest, ObservedAt: snapshotEvidence.ObservedAt, FreshUntil: snapshotEvidence.FreshUntil}
	request.Admission.Admission.QExistingGuardian = 14
	if substituted := riskbucket.CalculateAdmission(request.Admission.Admission); substituted.Refusal != nil || substituted.QFinal != precheck.QFinal() {
		t.Fatalf("substitute was not a valid same-q_final authority: %+v", substituted)
	}
	issued, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
	if err != nil {
		t.Fatal(err)
	}
	state, err := rig.journal.ReadRiskBucketState(context.Background(), riskbucket.OwnerKey{AccountID: "acct-7", Market: riskbucket.MarketKR, Symbol: "005930", ProspectiveGeneration: "prospective-mutated"})
	if err != nil {
		t.Fatal(err)
	}
	if _, substitutedPersisted := state.Reservations[bucket.Key]; substitutedPersisted {
		t.Fatal("post-precheck same-q_final evidence substitution became persisted authority")
	}
	originalKey := bucket.Key
	originalKey.PolicyVersion = "policy-v1"
	if _, sealedPersisted := state.Reservations[originalKey]; !sealedPersisted || issued.Decision.ID != "qfinal-mutated" {
		t.Fatalf("sealed authority missing: state=%+v issued=%+v", state, issued)
	}
}

func TestQFinalFinalIssuanceRechecksExpiredEvidenceAtGuardianClock(t *testing.T) {
	rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
		options.NewID = fixedIDs("qfinal-expired", "qfinal-expired-nonce")
	})
	request := qFinalKRRequest(t, rig, "expired", 10)
	precheck, err := rig.guardian.PrecheckQFinalEntry(request)
	if err != nil {
		t.Fatal(err)
	}
	rig.clock.Advance(2 * time.Minute)
	if _, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck); err == nil {
		t.Fatal("expired monetary evidence issued authority")
	}
	if _, err := rig.journal.LookupDecision(context.Background(), "qfinal-expired"); !errors.Is(err, journal.ErrDecisionNotFound) {
		t.Fatalf("expired evidence left decision: %v", err)
	}
}

func TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker(t *testing.T) {
	broker := &fakeBroker{}
	gw, j, clk := newGateway(t, broker)
	version, err := j.ReservationVersion(context.Background(), "acct-7")
	if err != nil {
		t.Fatal(err)
	}
	policyVersion, err := journal.QFinalPolicyVersion("guardian-v1", "missing-admission")
	if err != nil {
		t.Fatal(err)
	}
	limitsJSON, err := execgw.EncodeLimits(testLimits())
	if err != nil {
		t.Fatal(err)
	}
	result, err := j.RecordDecisionAndReserve(context.Background(), journal.IssueRequest{
		Decision: journal.DecisionRequest{
			ID: "qfinal-gateway-missing", AccountRef: "acct-7", SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
			Preimage:   journal.RiskIntent{AccountRef: "acct-7", Market: "kr", Symbol: "005930", Side: "BUY", Quantity: "2", EntryPrice: "70000", StopPrice: "69000", TargetPrice: "72000", PolicyVersion: policyVersion},
			LimitsJSON: limitsJSON, Nonce: "qfinal-gateway-nonce", IssuedAt: clk.Now(), ExpiresAt: clk.Now().Add(time.Minute),
		},
		Reserve: journal.ReserveRequest{
			SnapshotAsOf: clk.Now(), ObservedVersion: version,
			SnapshotUsage: []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: "0", Currency: "KRW"}},
			Limits:        []journal.AggregateAmount{{Kind: journal.ReservationKindOpenExposure, Amount: "5000000", Currency: "KRW"}},
			Reservations:  []journal.ReservationRequest{{ID: "qfinal-gateway-hold", Kind: journal.ReservationKindOpenExposure, Amount: "140000", Currency: "KRW"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit", Quantity: 2, Price: 70000, CurrencyMode: "KRW"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = gw.Place(context.Background(), execgw.PlaceRequest{Intent: intent, Decision: execgw.GuardianDecision{ID: result.Decision.ID, Generation: result.Decision.Generation}})
	var rejected *execgw.RejectedError
	places, _, _ := broker.totals()
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianRiskBucketMismatch || places != 0 {
		t.Fatalf("rejected=%+v places=%d err=%v", rejected, places, err)
	}
}

func TestGatewayLastMomentQFinalBarrierRefusesHoldReleaseAfterInitialAdmissionCheck(t *testing.T) {
	rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
		options.NewID = fixedIDs("qfinal-race", "qfinal-race-nonce")
	})
	precheck, err := rig.guardian.PrecheckQFinalEntry(qFinalKRRequest(t, rig, "race", 20))
	if err != nil {
		t.Fatal(err)
	}
	issued, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
	if err != nil {
		t.Fatal(err)
	}
	broker := &fakeBroker{}
	checks := 0
	var releaseErr error
	opts := execgw.Options{
		Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker), Clock: rig.clock,
		AccountRef: "acct-7", Source: "qfinal-last-moment-test",
	}
	opts.SetMarketProtectionForTest(func(market string, check int) (bool, string) {
		checks = check
		if check == 2 {
			_, releaseErr = rig.journal.OperatorReleaseReservation(context.Background(), journal.OperatorReleaseRequest{
				ReservationID: issued.Reservations[0].ID, Operator: "race-test-operator",
				Reason: "prove final q_final barrier", Evidence: "controlled release after initial check",
				Auditor: &modeAuditor{},
			})
		}
		return true, market + ":stable-protection"
	})
	gw, err := execgw.New(opts)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := orderintent.NormalizePlace(orderintent.PlaceInput{Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit", Quantity: 10, Price: 70000, CurrencyMode: "KRW"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = gw.Place(context.Background(), execgw.PlaceRequest{Intent: intent, Decision: issued.Decision})
	var rejected *execgw.RejectedError
	places, _, _ := broker.totals()
	if releaseErr != nil || checks != 2 || !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonGuardianReservationMissing || places != 0 {
		t.Fatalf("releaseErr=%v checks=%d rejected=%+v brokerPlaces=%d err=%v", releaseErr, checks, rejected, places, err)
	}
}

func qFinalKRRequest(t *testing.T, rig *guardianRig, suffix string, candidate uint64) execgw.QFinalEntryIssuance {
	t.Helper()
	now := fixedNow
	policy := riskbucket.ReservePolicy{
		AccountCurrency: "KRW", QuoteCurrency: "KRW", EvaluatedAt: now,
		Price: riskbucket.PriceEvidence{WorstExecutableQuote: "70000", Evidence: riskbucket.Evidence{Source: "official-order-contract", Version: "price-v1", Digest: "price-" + suffix, Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}},
		FX:    riskbucket.FXEvidence{RateQuoteToBase: "1", Haircut: "1", Evidence: riskbucket.Evidence{Source: "same-currency", Version: "fx-v1", Digest: "fx-" + suffix, Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}},
		Fee:   riskbucket.FeePolicy{FixedBaseMinor: "0", PerUnitBaseMinor: "0", MinimumBaseMinor: "0", Version: "fee-v1", Digest: "fee-" + suffix},
	}
	values := map[riskbucket.Dimension]string{
		riskbucket.DimensionHorizon: "SHORT", riskbucket.DimensionMarket: "KR", riskbucket.DimensionStrategy: "strategy-alpha",
		riskbucket.DimensionSector: "sector-tech", riskbucket.DimensionSymbol: "005930",
	}
	var buckets []riskbucket.BucketSnapshot
	var references []journal.RiskBucketSnapshotReference
	for _, dimension := range riskbucket.RequiredDimensionOrder() {
		key := riskbucket.BucketKey{Dimension: dimension, Value: values[dimension], PolicyVersion: "policy-v1"}
		policyEvidence := riskbucket.Evidence{Source: riskbucket.RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: "policy-" + suffix + "-" + string(dimension), Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
		policyProvenance, err := riskbucket.NewPolicyProvenance(key, policyEvidence)
		if err != nil {
			t.Fatal(err)
		}
		binding := riskbucket.BucketSnapshotBinding{Key: key, LimitMinor: "700000", FilledMinor: "0", HeldMinor: "0", SnapshotVersion: "snapshot-v1"}
		snapshotEvidence := riskbucket.Evidence{Source: riskbucket.RiskSnapshotAuthoritySource, Version: binding.SnapshotVersion, Digest: "snapshot-" + suffix + "-" + string(dimension), Official: true, Frozen: true, ObservedAt: now.Add(-time.Second), FreshUntil: now.Add(time.Minute)}
		snapshotProvenance, err := riskbucket.NewSnapshotProvenance(binding, snapshotEvidence)
		if err != nil {
			t.Fatal(err)
		}
		buckets = append(buckets, riskbucket.BucketSnapshot{Key: key, LimitMinor: binding.LimitMinor, FilledMinor: binding.FilledMinor, HeldMinor: binding.HeldMinor, SnapshotVersion: binding.SnapshotVersion, PolicyProvenance: policyProvenance, SnapshotProvenance: snapshotProvenance})
		references = append(references, journal.RiskBucketSnapshotReference{Key: key, SnapshotID: "snapshot-" + suffix + "-" + string(dimension), SnapshotDigest: snapshotEvidence.Digest, SnapshotVersion: binding.SnapshotVersion, PolicyDigest: policyEvidence.Digest, ObservedAt: policyEvidence.ObservedAt, FreshUntil: policyEvidence.FreshUntil})
	}
	return execgw.QFinalEntryIssuance{
		Market: "KR", Currency: "KRW", Symbol: "005930", QCandidate: candidate,
		EntryPrice: "70000", StopPrice: "69000", TargetPrice: "72000", Account: guardianAccount(), Collect: rig.collect,
		Admission: journal.RiskBucketAdmissionPlan{
			TransactionID: "qfinal-tx-" + suffix,
			Admission:     riskbucket.AdmissionRequest{QCandidate: candidate, QExistingGuardian: 999, Policy: policy, Buckets: buckets},
			Owner:         riskbucket.OwnerClaim{Key: riskbucket.OwnerKey{AccountID: "acct-7", Market: riskbucket.MarketKR, Symbol: "005930", ProspectiveGeneration: "prospective-" + suffix}, LaneID: "lane-short", CampaignID: "campaign-" + suffix},
			Snapshots:     references,
		},
		ExpectedPolicyVersion: rig.guardian.PolicyVersion(), ExpectedLimitsDigest: rig.guardian.LimitsDigest(),
	}
}
