package strategyruntime

import (
	"testing"
	"time"
)

func TestCentralFaultPlansEntryIncapableSafetyFallbackWithinSixtySeconds(t *testing.T) {
	manifest, err := newSafetyFallbackManifest(safetyFallbackManifestInput{Release: RuntimeRelease, RTO: 45 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	oldOwner, _ := newOwnerFence(7, "old-token")
	newOwner, _ := newOwnerFence(8, "safety-token")
	fault := mustCentralFault(t, FaultOwnerFenceCorrupt, runtimeNow, oldOwner)
	plan := PlanSafetyFallback(manifest, fault, newOwner, newTrustedTime(runtimeNow.Add(30*time.Second)))
	if plan.Status != FallbackStarted || plan.EntryAllowed || plan.LeaseIssuanceAllowed || !plan.Safety.AllEnabled() || plan.RTO > 60*time.Second || !plan.OldOwnerFenced || !plan.PreserveBrokerProtection {
		t.Fatalf("unsafe fallback=%+v", plan)
	}
}

func TestLateOrInvalidFallbackStaysNoEntryAndCritical(t *testing.T) {
	if _, err := newSafetyFallbackManifest(safetyFallbackManifestInput{Release: RuntimeRelease, RTO: 61 * time.Second}); err == nil {
		t.Fatal("RTO over 60 seconds accepted")
	}
	manifest, _ := newSafetyFallbackManifest(safetyFallbackManifestInput{Release: RuntimeRelease, RTO: 60 * time.Second})
	oldOwner, _ := newOwnerFence(7, "old-token")
	staleOwner, _ := newOwnerFence(7, "another-token")
	plan := PlanSafetyFallback(manifest, mustCentralFault(t, FaultJournalCorrupt, runtimeNow, oldOwner), staleOwner, newTrustedTime(runtimeNow.Add(61*time.Second)))
	if plan.Status != FallbackUnavailable || plan.EntryAllowed || plan.LeaseIssuanceAllowed || !plan.CriticalAlert || !plan.PreserveBrokerProtection {
		t.Fatalf("late/stale fallback failed open=%+v", plan)
	}
}
