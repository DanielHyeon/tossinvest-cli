package strategyengine

import (
	"strings"
	"testing"
)

func TestDormantRuntimeDescriptorSeparatesSectionsAndNeverPretendsActivation(t *testing.T) {
	d := DormantRuntimeDescriptor()
	if d.Category != "strategy-runtime" {
		t.Fatal(d.Category)
	}
	wantSections := [4]string{"parameters", "lane", "autostart", "live"}
	for i, section := range d.Sections {
		if section.ID != wantSections[i] || section.Label == "" || section.ActionOwner != "a050" {
			t.Fatalf("section[%d]=%+v", i, section)
		}
	}

	wantFields := [12]string{
		"min_vwap_slope_pct", "ema_touch_tolerance_pct", "min_forward_space_pct", "min_expected_rr",
		"tangled_band_pct", "max_band_expansion_rate", "hard_stop_pct", "partial_take_profit_at_r",
		"skip_open_minutes", "max_signal_age_seconds", "max_entry_price_drift_pct", "symbol_state_stale_seconds",
	}
	for i, field := range d.Fields {
		if field.Key != wantFields[i] || field.Label == "" || field.Help == "" || field.Default == "" ||
			field.Desired == "" || field.Effective != "미구성" || field.Unit == "" || field.Range == "" ||
			field.Provenance == "" || field.ApplyTiming == "" {
			t.Fatalf("field[%d] incomplete=%+v", i, field)
		}
		if !strings.Contains(field.Provenance, SourceCommit) || !strings.Contains(field.Provenance, FrozenSourceSetDigest) {
			t.Fatalf("field[%d] has abbreviated provenance=%q", i, field.Provenance)
		}
	}
	if got := d.Fields[4].Help; !strings.Contains(got, "0.35 미만") || !strings.Contains(got, "작을수록 분리가 부족") {
		t.Fatalf("tangled help=%q", got)
	}

	wantBlockers := [9]string{
		"source", "candidate", "scheduler", "protection", "guardian", "reconciliation",
		"operating-mode", "kill-switch", "activation-manifest",
	}
	for i, blocker := range d.Blockers {
		if blocker.Key != wantBlockers[i] || blocker.Label == "" || !blocker.Desired.Valid() ||
			!blocker.Effective.Valid() || !blocker.Freshness.Valid() || !blocker.Reason.Valid() {
			t.Fatalf("blocker[%d]=%+v", i, blocker)
		}
	}
	if d.Blockers[0].Effective != RuntimeStateNotConfigured || d.Blockers[3].Effective != RuntimeStateUnwired {
		t.Fatalf("fail-closed blockers=%+v", d.Blockers)
	}
}

func TestRuntimeStateAndRefusalVocabulariesRejectUnknownValues(t *testing.T) {
	if RuntimeState("ACTIVE").Valid() || RuntimeRefusal("surprise").Valid() {
		t.Fatal("unknown runtime vocabulary was accepted")
	}
	if !RuntimeStateOff.Valid() || !RuntimeStateVerified.Valid() || !RuntimeRefusalNone.Valid() || !RuntimeRefusalReadFailed.Valid() {
		t.Fatal("closed vocabulary rejected a declared value")
	}
}
