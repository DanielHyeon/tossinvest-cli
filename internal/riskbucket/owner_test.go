package riskbucket

import (
	"reflect"
	"testing"
	"time"
)

func TestAcquireOwnerAllowsOneOwnerAndIdempotentScaleIn(t *testing.T) {
	state := OwnerState{Owners: make(map[OwnerKey]Owner)}
	key := OwnerKey{AccountID: "acct-1", Market: MarketKR, Symbol: "005930", ProspectiveGeneration: "prospective-1"}
	claim := OwnerClaim{Key: key, LaneID: "kr-short", CampaignID: "campaign-1"}

	owned, result, err := AcquireOwner(state, claim)
	if err != nil || !result.Acquired {
		t.Fatalf("first acquire: result=%#v err=%v", result, err)
	}
	same, result, err := AcquireOwner(owned, claim)
	if err != nil || !result.Reused || len(same.Owners) != 1 {
		t.Fatalf("same owner scale-in: result=%#v err=%v", result, err)
	}

	conflict := claim
	conflict.LaneID = "kr-medium"
	afterConflict, _, err := AcquireOwner(owned, conflict)
	if !IsRefusal(err, RefusalOwnerConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if len(afterConflict.Owners) != 1 || afterConflict.Owners[key].LaneID != claim.LaneID {
		t.Fatal("owner conflict changed authoritative owner")
	}

	differentProspective := claim
	differentProspective.Key.ProspectiveGeneration = "prospective-2"
	differentProspective.CampaignID = "campaign-2"
	afterProspectiveConflict, _, err := AcquireOwner(owned, differentProspective)
	if !IsRefusal(err, RefusalOwnerConflict) || len(afterProspectiveConflict.Owners) != 1 {
		t.Fatalf("different prospective token bypassed one-symbol owner: err=%v owners=%d", err, len(afterProspectiveConflict.Owners))
	}
}

func TestBindActualGenerationIsSetOnce(t *testing.T) {
	state := OwnerState{Owners: make(map[OwnerKey]Owner)}
	key := OwnerKey{AccountID: "acct-1", Market: MarketUS, Symbol: "AAPL", ProspectiveGeneration: "prospective-1"}
	state, _, _ = AcquireOwner(state, OwnerClaim{Key: key, LaneID: "us-short", CampaignID: "campaign-1"})

	bound, err := BindActualGeneration(state, key, "actual-42")
	if err != nil || bound.Owners[key].ActualGeneration != "actual-42" {
		t.Fatalf("bind result=%#v err=%v", bound.Owners[key], err)
	}
	retry, err := BindActualGeneration(bound, key, "actual-42")
	if err != nil || retry.Owners[key].ActualGeneration != "actual-42" {
		t.Fatalf("idempotent bind failed: %v", err)
	}
	conflict, err := BindActualGeneration(bound, key, "actual-99")
	if !IsRefusal(err, RefusalOwnerConflict) || conflict.Owners[key].ActualGeneration != "actual-42" {
		t.Fatalf("set-once conflict result=%#v err=%v", conflict.Owners[key], err)
	}
}

func TestDuplicateOwnerScopeAlwaysFailsClosedAsReconstructionMismatch(t *testing.T) {
	key1 := OwnerKey{AccountID: "acct-1", Market: MarketKR, Symbol: "005930", ProspectiveGeneration: "prospective-1"}
	key2 := OwnerKey{AccountID: "acct-1", Market: MarketKR, Symbol: "005930", ProspectiveGeneration: "prospective-2"}
	state := OwnerState{Owners: map[OwnerKey]Owner{
		key1: {LaneID: "kr-short", CampaignID: "campaign-1", ActualGeneration: "actual-1"},
		key2: {LaneID: "kr-medium", CampaignID: "campaign-2", ActualGeneration: "actual-2"},
	}}
	original := cloneOwnerState(state)

	for i := 0; i < 100; i++ {
		next, _, err := AcquireOwner(state, OwnerClaim{Key: key1, LaneID: "kr-short", CampaignID: "campaign-1"})
		if !IsRefusal(err, RefusalReconstructionMismatch) || !reflect.DeepEqual(next, original) {
			t.Fatalf("acquire iteration %d: next=%#v err=%v", i, next, err)
		}
	}
	if next, err := BindActualGeneration(state, key1, "actual-1"); !IsRefusal(err, RefusalReconstructionMismatch) || !reflect.DeepEqual(next, original) {
		t.Fatalf("bind duplicate scope: next=%#v err=%v", next, err)
	}
	if next, _, err := ReleaseOwner(state, key1, cleanReleaseEvidence(key1, state.Owners[key1])); !IsRefusal(err, RefusalReconstructionMismatch) || !reflect.DeepEqual(next, original) {
		t.Fatalf("release duplicate scope: next=%#v err=%v", next, err)
	}
}

func TestReleaseOwnerRequiresEveryCleanPredicate(t *testing.T) {
	state := OwnerState{Owners: make(map[OwnerKey]Owner)}
	key := OwnerKey{AccountID: "acct-1", Market: MarketKR, Symbol: "005930", ProspectiveGeneration: "prospective-1"}
	state, _, _ = AcquireOwner(state, OwnerClaim{Key: key, LaneID: "kr-short", CampaignID: "campaign-1"})
	state, err := BindActualGeneration(state, key, "actual-1")
	if err != nil {
		t.Fatal(err)
	}
	clean := cleanReleaseEvidence(key, state.Owners[key])

	tests := []struct {
		name string
		edit func(*ReleaseEvidence)
	}{
		{"position open", func(e *ReleaseEvidence) { e.PositionClosed = false }},
		{"quantity nonzero", func(e *ReleaseEvidence) { e.PositionQuantity = 1 }},
		{"pending entry", func(e *ReleaseEvidence) { e.PendingEntry = true }},
		{"held remains", func(e *ReleaseEvidence) { e.HeldMinor = "1" }},
		{"broker not reconciled", func(e *ReleaseEvidence) { e.BrokerReconciled = false }},
		{"broker nonzero", func(e *ReleaseEvidence) { e.BrokerQuantity = 1 }},
		{"protection active", func(e *ReleaseEvidence) { e.ProtectionOrder = ClaimActive }},
		{"protection unknown", func(e *ReleaseEvidence) { e.ProtectionSaga = ClaimUnknown }},
		{"sell claim stale", func(e *ReleaseEvidence) { e.SellClaim = ClaimStale }},
		{"sell mutation pending", func(e *ReleaseEvidence) { e.SellMutation = ClaimPending }},
		{"fill unresolved", func(e *ReleaseEvidence) { e.UnresolvedFill = ClaimActive }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence := clean
			tt.edit(&evidence)
			next, result, err := ReleaseOwner(state, key, evidence)
			if err != nil || result.Released || len(next.Owners) != 1 {
				t.Fatalf("unsafe release: result=%#v err=%v owners=%d", result, err, len(next.Owners))
			}
		})
	}

	released, result, err := ReleaseOwner(state, key, clean)
	if err != nil || !result.Released || len(released.Owners) != 0 {
		t.Fatalf("clean release: result=%#v err=%v owners=%d", result, err, len(released.Owners))
	}
	retry, result, err := ReleaseOwner(released, key, clean)
	if err != nil || !result.AlreadyReleased || len(retry.Owners) != 0 {
		t.Fatalf("release retry: result=%#v err=%v", result, err)
	}
}

func TestReleaseOwnerRetainsOwnerForStaleOrMismatchedAttestation(t *testing.T) {
	key := OwnerKey{AccountID: "acct-1", Market: MarketUS, Symbol: "AAPL", ProspectiveGeneration: "prospective-1"}
	owner := Owner{LaneID: "us-short", CampaignID: "campaign-1", ActualGeneration: "actual-42"}
	state := OwnerState{Owners: map[OwnerKey]Owner{key: owner}}
	original := cloneOwnerState(state)

	tests := []struct {
		name string
		edit func(*ReleaseEvidence)
	}{
		{"wrong lane", func(e *ReleaseEvidence) { e.LaneID = "us-medium" }},
		{"wrong campaign", func(e *ReleaseEvidence) { e.CampaignID = "campaign-2" }},
		{"wrong actual generation", func(e *ReleaseEvidence) { e.ActualGeneration = "actual-99" }},
		{"wrong prospective generation", func(e *ReleaseEvidence) { e.OwnerKey.ProspectiveGeneration = "prospective-2" }},
		{"stale attestation", func(e *ReleaseEvidence) {
			e.Attestation = mustReleaseAttestation(e.OwnerKey, Owner{LaneID: e.LaneID, CampaignID: e.CampaignID, ActualGeneration: e.ActualGeneration}, Evidence{Source: OwnerReleaseAuthoritySource, Version: "release-v1", Digest: "release-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(-time.Second)})
		}},
		{"attestation bound to another owner", func(e *ReleaseEvidence) {
			otherKey := e.OwnerKey
			otherKey.ProspectiveGeneration = "prospective-2"
			e.Attestation = mustReleaseAttestation(otherKey, owner, Evidence{Source: OwnerReleaseAuthoritySource, Version: "release-v1", Digest: "release-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence := cleanReleaseEvidence(key, owner)
			tt.edit(&evidence)
			next, result, err := ReleaseOwner(state, key, evidence)
			if err == nil || result.Released || !reflect.DeepEqual(next, original) {
				t.Fatalf("unsafe release: result=%#v next=%#v err=%v", result, next, err)
			}
		})
	}
}

func cleanReleaseEvidence(key OwnerKey, owner Owner) ReleaseEvidence {
	return ReleaseEvidence{
		OwnerKey:         key,
		LaneID:           owner.LaneID,
		CampaignID:       owner.CampaignID,
		ActualGeneration: owner.ActualGeneration,
		EvaluatedAt:      testNow,
		Attestation:      mustReleaseAttestation(key, owner, Evidence{Source: OwnerReleaseAuthoritySource, Version: "release-v1", Digest: "release-digest", Official: true, Frozen: true, ObservedAt: testNow.Add(-time.Minute), FreshUntil: testNow.Add(time.Minute)}),
		PositionClosed:   true,
		PositionQuantity: 0,
		PendingEntry:     false,
		HeldMinor:        "0",
		BrokerReconciled: true,
		BrokerQuantity:   0,
		ProtectionOrder:  ClaimClean,
		ProtectionSaga:   ClaimClean,
		SellClaim:        ClaimClean,
		SellMutation:     ClaimClean,
		UnresolvedFill:   ClaimClean,
	}
}

func mustReleaseAttestation(key OwnerKey, owner Owner, evidence Evidence) ReleaseAttestation {
	attestation, err := NewReleaseAttestation(key, owner, evidence)
	if err != nil {
		panic(err)
	}
	return attestation
}
