package operatorview

import "testing"

func TestBuildExitLineReferenceAuthorityMatrix(t *testing.T) {
	raw := &StoredExitReference{EntryPrice: "70000", InitialStop: "66500", Baseline: "68000",
		HighWater: "73000", LifecycleGeneration: 1}
	tests := []struct {
		name string
		in   ExitLineReferenceSource
		kind ExitLineReferenceKind
		stop string
	}{
		{name: "fresh canonical stays actionable only", in: ExitLineReferenceSource{
			Market: "kr", CanonicalSnapshotPresent: true, Raw: raw,
			LifecycleKnown: true, CurrentLifecycleGeneration: 1,
		}},
		{name: "stale canonical remains fail closed without raw promotion", in: ExitLineReferenceSource{
			Market: "us", CanonicalSnapshotPresent: true, CanonicalSnapshotStale: true, Raw: raw,
			LifecycleKnown: true, CurrentLifecycleGeneration: 1,
		}},
		{name: "legacy raw", in: ExitLineReferenceSource{Market: "kr", Raw: raw,
			LifecycleKnown: true, CurrentLifecycleGeneration: 1,
			UnknownReason: "legacy_snapshot_absent"}, kind: ExitLineReferenceLegacyRaw},
		{name: "adoption plan", in: ExitLineReferenceSource{Market: "us",
			ManagementStatus: "RECONCILE_BLOCKED", EffectiveSettingsKnown: true,
			EffectiveStopPct: .03}, kind: ExitLineReferenceAdoptionPlan, stop: "3%"},
		{name: "runtime unknown plan stays absent", in: ExitLineReferenceSource{Market: "us",
			ManagementStatus: "UNKNOWN", EffectiveSettingsKnown: false}, kind: ExitLineReferenceRuntimeUnknown},
		{name: "unmanaged does not get plan", in: ExitLineReferenceSource{Market: "us",
			ManagementStatus: "UNMANAGED", EffectiveSettingsKnown: true, EffectiveStopPct: .03}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildExitLineReference(tc.in)
			if got.Kind != tc.kind || got.StopPercent != tc.stop {
				t.Fatalf("reference=%+v want kind=%q stop=%q", got, tc.kind, tc.stop)
			}
			if got.EffectiveKnown {
				t.Fatalf("non-actionable reference became effective: %+v", got)
			}
			if tc.kind == ExitLineReferenceAdoptionPlan &&
				(got.EntryPrice != "—" || got.InitialStop != "—" || got.Baseline != "—") {
				t.Fatalf("adoption plan synthesized a price: %+v", got)
			}
		})
	}
}

func TestBuildExitLineReferenceRejectsCrossLifecycleEvidence(t *testing.T) {
	got := BuildExitLineReference(ExitLineReferenceSource{Market: "us", LifecycleKnown: true,
		LifecycleProofRequired:     true,
		CurrentLifecycleGeneration: 2, Raw: &StoredExitReference{
			EntryPrice: "SENTINEL_ENTRY", InitialStop: "SENTINEL_STOP", Baseline: "SENTINEL_BASELINE",
			HighWater: "SENTINEL_HIGH", LifecycleGeneration: 1,
		}})
	if got.Kind != ExitLineReferenceGenerationMismatch || got.EntryPrice != "—" || got.InitialStop != "—" ||
		got.Baseline != "—" || got.HighWater != "—" || got.Reason == "" {
		t.Fatalf("cross-lifecycle evidence leaked: %+v", got)
	}
}

func TestExitLineReferenceUsesMarketNativeCurrencyWithoutChangingDecimals(t *testing.T) {
	for _, tc := range []struct{ market, currency string }{{"kr", "KRW"}, {"US", "USD"}} {
		got := BuildExitLineReference(ExitLineReferenceSource{Market: tc.market,
			Raw:           &StoredExitReference{Baseline: "3890.25", InitialStop: "3890.25", LifecycleGeneration: 1},
			UnknownReason: "legacy_snapshot_absent"})
		if got.Currency != tc.currency || got.Baseline != "3890.25" {
			t.Fatalf("market=%s reference=%+v", tc.market, got)
		}
	}
}

func TestBuildExitLineReferenceSuppressesCorruptAndUnverifiedRawEvidence(t *testing.T) {
	raw := &StoredExitReference{EntryPrice: "SENTINEL_ENTRY", InitialStop: "SENTINEL_STOP",
		Baseline: "SENTINEL_BASELINE", LifecycleGeneration: 1}
	corrupt := BuildExitLineReference(ExitLineReferenceSource{Market: "US", Raw: raw,
		UnknownReason: "invalid_effective_snapshot"})
	if corrupt.Present() {
		t.Fatalf("corrupt raw evidence became visible: %+v", corrupt)
	}
	unverified := BuildExitLineReference(ExitLineReferenceSource{Market: "US", Raw: raw,
		UnknownReason: "legacy_snapshot_absent", LifecycleProofRequired: true})
	if unverified.Kind != ExitLineReferenceLifecycleUnknown || unverified.Baseline != "—" {
		t.Fatalf("unverified lifecycle evidence leaked: %+v", unverified)
	}
	releasedCorrupt := BuildExitLineReference(ExitLineReferenceSource{Market: "US", Raw: raw,
		Released: true, UnknownReason: "partial_snapshot_tuple"})
	if releasedCorrupt.Present() {
		t.Fatalf("released corruption bypassed raw allowlist: %+v", releasedCorrupt)
	}
}
