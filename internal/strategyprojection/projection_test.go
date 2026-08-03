package strategyprojection

import (
	"reflect"
	"testing"
	"time"
)

var projectionNow = time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)

func TestDormantSnapshotContainsExactPairedHonestMarkets(t *testing.T) {
	snapshot := DormantSnapshot(projectionNow)
	if err := Validate(snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.SchemaVersion != SchemaVersion || len(snapshot.Markets) != 2 {
		t.Fatalf("snapshot=%+v", snapshot)
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		item := snapshot.Markets[market]
		if item.Market != market || item.Status != StatusUnknown || item.Error == nil || item.Error.Code != RefusalNotConfigured ||
			item.Lane.Desired != StateOff || item.Lane.Effective != StateOff || item.Evidence.Freshness != FreshnessUnobserved ||
			item.Protection.Readiness != ProtectionUnwired || item.Protection.Refusal != RefusalNotConfigured ||
			item.FirstRefusal != RefusalNotConfigured || item.ObservedAt != nil {
			t.Fatalf("market=%s item=%+v", market, item)
		}
	}
}

func TestMarketFailureReplacesOnlyExactMarketWithoutFallback(t *testing.T) {
	current := currentPair(t)
	beforeKR := current.Markets[MarketKR]
	partial := WithMarketFailure(current, MarketUS, RefusalRuntimeUnavailable, projectionNow.Add(time.Second))
	if err := Validate(partial); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(partial.Markets[MarketKR], beforeKR) {
		t.Fatal("US failure altered KR")
	}
	us := partial.Markets[MarketUS]
	if us.Status != StatusUnknown || us.Error == nil || us.Error.Code != RefusalRuntimeUnavailable ||
		us.Protection.Readiness != ProtectionUnwired || us.FirstRefusal != RefusalRuntimeUnavailable || us.Evidence.ID != nil {
		t.Fatalf("US failure laundered into defaults/current=%+v", us)
	}
}

func TestEitherMarketFailurePreservesTheExactPeer(t *testing.T) {
	for _, failed := range []Market{MarketKR, MarketUS} {
		t.Run(string(failed), func(t *testing.T) {
			current := currentPair(t)
			peer := MarketKR
			if failed == MarketKR {
				peer = MarketUS
			}
			beforePeer := current.Markets[peer]
			partial := WithMarketFailure(current, failed, RefusalRuntimeUnavailable, projectionNow.Add(time.Second))
			if err := Validate(partial); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(partial.Markets[peer], beforePeer) {
				t.Fatalf("%s failure altered %s", failed, peer)
			}
			failedProjection := partial.Markets[failed]
			if failedProjection.Status != StatusUnknown || failedProjection.Error == nil ||
				failedProjection.Error.Code != RefusalRuntimeUnavailable || failedProjection.Protection.Readiness != ProtectionUnwired {
				t.Fatalf("%s did not fail closed: %+v", failed, failedProjection)
			}
		})
	}
}

func TestValidationRejectsMissingDuplicateScopeAndInventedReadiness(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Snapshot)
	}{
		{"missing US", func(s *Snapshot) { delete(s.Markets, MarketUS) }},
		{"cross-market record", func(s *Snapshot) { item := s.Markets[MarketUS]; item.Market = MarketKR; s.Markets[MarketUS] = item }},
		{"zero market fallback", func(s *Snapshot) { s.Markets[MarketUS] = MarketProjection{} }},
		{"third readiness", func(s *Snapshot) {
			item := s.Markets[MarketUS]
			item.Protection.Readiness = ProtectionReadiness("UNKNOWN")
			s.Markets[MarketUS] = item
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := DormantSnapshot(projectionNow)
			test.edit(&snapshot)
			if err := Validate(snapshot); err == nil {
				t.Fatal("invalid projection accepted")
			}
		})
	}
}

func TestGoldenRegistryIsStableAndReturnedAsCopy(t *testing.T) {
	want := []string{"market", "status", "lane", "evidence", "campaign", "horizonRisk", "scheduler", "activation", "protection", "reconciliation", "firstRefusal", "observedAt"}
	registry := Registry()
	got := make([]string, 0, len(registry))
	for _, field := range registry {
		got = append(got, field.Key)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("registry=%#v", got)
	}
	registry[0].Key = "mutated"
	if Registry()[0].Key != "market" {
		t.Fatal("registry caller mutated shared truth")
	}
}

func currentPair(t *testing.T) Snapshot {
	t.Helper()
	markets := map[Market]MarketProjection{}
	for _, market := range []Market{MarketKR, MarketUS} {
		observed := projectionNow.Add(-time.Second)
		id, evidenceDigest := "evidence-"+string(market), digest("evidence-"+string(market))
		laneID, laneVersion := "lane-"+string(market), "v1"
		campaignID, legID := "campaign-"+string(market), "leg-1"
		bucket, policy := "SHORT", "risk-v1"
		calendarSource, calendarVersion := "official-calendar-"+string(market), "2026.08"
		activationDigest := digest("activation-" + string(market))
		markets[market] = MarketProjection{Market: market, Status: StatusCurrent,
			Lane:           LaneProjection{ID: &laneID, Version: &laneVersion, Desired: StateOn, Effective: StateOn},
			Evidence:       EvidenceProjection{ID: &id, Digest: &evidenceDigest, Freshness: FreshnessCurrent},
			Campaign:       CampaignProjection{ID: &campaignID, LegID: &legID},
			HorizonRisk:    HorizonRiskProjection{Bucket: &bucket, PolicyVersion: &policy, Status: ComponentCurrent},
			Scheduler:      SchedulerProjection{Desired: StateOn, Effective: StateOn, CalendarSource: &calendarSource, CalendarVersion: &calendarVersion, CalendarFreshness: FreshnessCurrent},
			Activation:     ActivationProjection{Desired: StateOn, Effective: StateOn, Digest: &activationDigest, Status: ActivationConfigured},
			Protection:     ProtectionProjection{Readiness: ProtectionWired, Refusal: RefusalNone},
			Reconciliation: ReconciliationProjection{Status: ReconciliationHealthy, Refusal: RefusalNone},
			FirstRefusal:   RefusalNone, ObservedAt: &observed}
	}
	snapshot := Snapshot{SchemaVersion: SchemaVersion, GeneratedAt: projectionNow, Markets: markets}
	if err := Validate(snapshot); err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func digest(value string) string {
	return digestString(value)
}
