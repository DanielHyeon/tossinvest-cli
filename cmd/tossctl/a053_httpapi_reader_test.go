package main

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/httpapi"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

func TestHTTPAPIExitLineReferenceCoversLegacyAndAdoptionPlan(t *testing.T) {
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	legacy := httpapi.Position{Market: "KR", Symbol: "333430"}
	stored := journal.PositionExit{HasExit: true, Exit: journal.ExitState{EntryPrice: "4095",
		InitialStop: "3890.25", Baseline: "3890.25", HighWater: "4135", LifecycleGeneration: 1,
		Snapshot: journal.ExitSnapshotView{UnknownReason: "legacy_snapshot_absent"}}}
	applyStoredPosition(&legacy, stored, now)
	state := positionpolicy.State{Status: positionpolicy.StatusManaged, Version: 1, AdoptionGeneration: 1}
	applyManagementProjection(&legacy, true, true, false, positionpolicy.ManagementRuntime{EffectiveKnown: true})
	applyExitLineReference(&legacy, &stored, state, true, positionpolicy.ManagementRuntime{EffectiveKnown: true})
	if legacy.ExitLineReference == nil || legacy.ExitLineReference.Kind != "LEGACY_RAW" ||
		legacy.ExitLineReference.Baseline != "3890.25" || legacy.ExitLine.CurrentProtection != "—" {
		t.Fatalf("legacy reference=%+v position=%+v", legacy.ExitLineReference, legacy)
	}

	plan := httpapi.Position{Market: "US", Symbol: "IONQ"}
	runtime := positionpolicy.ManagementRuntime{EffectiveKnown: true,
		Effective: positionpolicy.NewAdoptionSettings(false, .03, []string{"IONQ"}, nil, ""),
		Blocks:    []positionpolicy.ReconcileBlock{{Scope: positionpolicy.ScopeAccount, Reason: "QUANTITY_MISMATCH"}}}
	applyManagementProjection(&plan, true, false, false, runtime)
	applyExitLineReference(&plan, nil, positionpolicy.State{}, false, runtime)
	if plan.ExitLineReference == nil || plan.ExitLineReference.Kind != "ADOPTION_PLAN" ||
		plan.ExitLineReference.StopPercent != "3%" || plan.ExitLineReference.Baseline != "—" {
		t.Fatalf("US plan reference=%+v position=%+v", plan.ExitLineReference, plan)
	}
}

func TestHTTPAPIExitLineReferenceDoesNotCrossLifecycleGeneration(t *testing.T) {
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	stored := journal.PositionExit{HasExit: true, Exit: journal.ExitState{EntryPrice: "OLD_ENTRY",
		InitialStop: "OLD_STOP", Baseline: "OLD_BASELINE", HighWater: "OLD_HIGH", LifecycleGeneration: 1,
		Snapshot: journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
			Line: exitpolicy.ExitLineSnapshot{EntryPrice: "OLD_ENTRY", InitialStop: "OLD_STOP",
				CurrentProtection: "OLD_CURRENT", NextTarget: "OLD_TARGET", SnapshotID: "OLD_SNAPSHOT",
				DecisionID: "OLD_DECISION", ObservationID: "OLD_OBSERVATION"},
			ObservationSource: "official.quote", ObservedAt: now.Format(time.RFC3339Nano),
		}}}}
	out := httpapi.Position{Market: "US", Symbol: "IONQ"}
	applyStoredPosition(&out, stored, now)
	applyExitLineReference(&out, &stored, positionpolicy.State{Status: positionpolicy.StatusManaged,
		Version: 1, AdoptionGeneration: 2}, true, positionpolicy.ManagementRuntime{EffectiveKnown: true})
	if out.ExitLineReference == nil || out.ExitLineReference.Kind != "GENERATION_MISMATCH" ||
		out.StoredExitEvidence != nil || out.ExitLine.CurrentProtection != "—" || out.ExitLine.NextTarget != "—" ||
		out.ExitLine.SnapshotID != "—" || out.ExitLine.DecisionID != "—" || out.ExitLine.ObservationID != "—" {
		t.Fatalf("cross-generation position=%+v", out)
	}
}

func TestHTTPAPIReleasedCanonicalExitStillReturnsHistoricalReference(t *testing.T) {
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	stored := journal.PositionExit{HasExit: true, Exit: journal.ExitState{EntryPrice: "200", InitialStop: "190",
		Baseline: "195", HighWater: "210", LifecycleGeneration: 1,
		Snapshot: journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
			Line: exitpolicy.ExitLineSnapshot{CurrentProtection: "195", NextTarget: "220",
				SnapshotID: "released-snapshot", DecisionID: "released-decision"},
			ObservationSource: "official.quote", ObservedAt: now.Format(time.RFC3339Nano),
		}}}}
	out := httpapi.Position{Market: "US", Symbol: "AAPL"}
	applyStoredPosition(&out, stored, now)
	state := positionpolicy.State{Status: positionpolicy.StatusReleased, Version: 1, AdoptionGeneration: 1}
	applyReleasedExitTruth(&out, stored)
	applyManagementProjection(&out, true, false, true, positionpolicy.ManagementRuntime{EffectiveKnown: true})
	applyExitLineReference(&out, &stored, state, true, positionpolicy.ManagementRuntime{EffectiveKnown: true})
	if out.ExitLineReference == nil || out.ExitLineReference.Kind != "LEGACY_RAW" ||
		out.ExitLineReference.Label != "저장된 과거 원장 기준선 · 관리 해제됨" ||
		out.ExitLine.CurrentProtection != "—" || out.ExitLine.SnapshotID != "—" {
		t.Fatalf("released historical projection=%+v", out)
	}
}

func TestHTTPAPILifecycleUnknownClearsCanonicalPricesAndIdentity(t *testing.T) {
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	stored := journal.PositionExit{HasExit: true, Exit: journal.ExitState{EntryPrice: "OLD_ENTRY",
		InitialStop: "OLD_STOP", Baseline: "OLD_BASELINE", HighWater: "OLD_HIGH", LifecycleGeneration: 1,
		Snapshot: journal.ExitSnapshotView{Snapshot: &journal.StoredExitSnapshot{
			Line: exitpolicy.ExitLineSnapshot{EntryPrice: "OLD_ENTRY", InitialStop: "OLD_STOP",
				CurrentProtection: "OLD_CURRENT", NextTarget: "OLD_TARGET", SnapshotID: "OLD_SNAPSHOT",
				DecisionID: "OLD_DECISION", ObservationID: "OLD_OBSERVATION"},
			ObservationSource: "official.quote", ObservedAt: now.Format(time.RFC3339Nano),
		}}}}
	out := httpapi.Position{Market: "US", Symbol: "IONQ"}
	applyStoredPosition(&out, stored, now)
	applyExitLineReference(&out, &stored, positionpolicy.State{}, false,
		positionpolicy.ManagementRuntime{EffectiveKnown: true})
	if out.ExitLineReference == nil || out.ExitLineReference.Kind != "LIFECYCLE_UNKNOWN" ||
		out.StoredExitEvidence != nil || out.ExitLine.CurrentProtection != "—" || out.ExitLine.NextTarget != "—" ||
		out.ExitLine.SnapshotID != "—" || out.ExitLine.DecisionID != "—" || out.ExitLine.ObservationID != "—" {
		t.Fatalf("unverified canonical projection=%+v", out)
	}
}

func TestHTTPAPIReleasedCorruptRawEvidenceStaysHidden(t *testing.T) {
	now := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	stored := journal.PositionExit{HasExit: true, Exit: journal.ExitState{EntryPrice: "SENTINEL_ENTRY",
		InitialStop: "SENTINEL_STOP", Baseline: "SENTINEL_BASELINE", LifecycleGeneration: 1,
		Snapshot: journal.ExitSnapshotView{UnknownReason: "invalid_effective_snapshot"}}}
	out := httpapi.Position{Market: "US", Symbol: "IONQ"}
	applyStoredPosition(&out, stored, now)
	state := positionpolicy.State{Status: positionpolicy.StatusReleased, Version: 1, AdoptionGeneration: 1}
	applyReleasedExitTruth(&out, stored)
	applyExitLineReference(&out, &stored, state, true, positionpolicy.ManagementRuntime{EffectiveKnown: true})
	if out.ExitLineReference != nil || out.StoredExitEvidence != nil || out.ExitLine.CurrentProtection != "—" {
		t.Fatalf("released corrupt evidence leaked: %+v", out)
	}
}
