package operatorview

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func completeLine() exitpolicy.ExitLineSnapshot {
	return exitpolicy.ExitLineSnapshot{
		SnapshotID: "els-view-1", DecisionID: "eld-view-1",
		Policy:     exitpolicy.PolicyIdentity{ID: "COMMON_LADDER_DEFAULT", Version: "v1", Digest: "sha256:policy"},
		EntryPrice: "10000", InitialStop: "9300", ObservedPrice: "11200",
		CurrentProtection: "10100", HighWater: "11300",
		ActiveRung: 1, NextTarget: "12000", NextProtection: "10800",
		Action: exitpolicy.ActionLadderPartial, Ratio: "0.25", ProjectedQuantity: "2",
		Orderable: true,
	}
}

func TestBuildExitLinePreservesTheCanonicalSnapshot(t *testing.T) {
	line := completeLine()
	got := BuildExitLine(Source{
		Snapshot: &line, RemainingQuantity: "8", ObservationSource: "official.quote",
		ObservedAt: "2026-07-31T23:59:30Z", EffectiveSource: "recomputed",
	})

	if !got.Fresh() || got.StatusText != "평가 완료" {
		t.Fatalf("status = %q/%q, want fresh/평가 완료", got.Status, got.StatusText)
	}
	for label, pair := range map[string][2]string{
		"entry": {got.EntryPrice, "10000"}, "initial stop": {got.InitialStop, "9300"},
		"current protection": {got.CurrentProtection, "10100"}, "next target": {got.NextTarget, "12000"},
		"next protection": {got.NextProtection, "10800"}, "projected quantity": {got.ProjectedQuantity, "2"},
		"observation": {got.ObservedPrice, "11200"}, "stage": {got.Stage, "rung 1"},
		"decision": {got.DecisionID, "eld-view-1"}, "policy": {got.Policy, "COMMON_LADDER_DEFAULT · v1"},
	} {
		if pair[0] != pair[1] {
			t.Errorf("%s = %q, want %q", label, pair[0], pair[1])
		}
	}
	if got.ActionText != "중간 익절 25%" {
		t.Errorf("action text = %q, want concise Korean summary", got.ActionText)
	}
}

func TestBuildExitLineFailsClosedForStaleAndUnknownEvidence(t *testing.T) {
	line := completeLine()
	for _, tc := range []struct {
		name string
		in   Source
		want string
	}{
		{"stale", Source{Snapshot: &line, StaleReason: "observation_older_than_limit", ObservedAt: "2026-07-31T23:58:00Z"}, "오래된 평가"},
		{"unknown", Source{UnknownReason: "partial_evaluated_tuple"}, "근거 없음"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildExitLine(tc.in)
			if got.StatusText != tc.want || got.Reason == "" {
				t.Fatalf("status/reason = %q/%q, want %q and a reason", got.StatusText, got.Reason, tc.want)
			}
			for label, value := range map[string]string{
				"entry": got.EntryPrice, "initial stop": got.InitialStop,
				"current protection": got.CurrentProtection, "next target": got.NextTarget,
				"next protection": got.NextProtection, "projected quantity": got.ProjectedQuantity,
				"observation": got.ObservedPrice,
			} {
				if value != "—" {
					t.Errorf("%s = %q, want em dash for non-current evidence", label, value)
				}
			}
		})
	}
}

func TestBuildExitLineExplainsAOneShareStateOnlyPartial(t *testing.T) {
	line := completeLine()
	line.ProjectedQuantity = "0"
	line.Orderable = false
	line.StateOnly = true
	got := BuildExitLine(Source{Snapshot: &line, RemainingQuantity: "1"})
	if !got.OneShare || got.OneShareText != "중간 매도 없음 · 보호선 승격" ||
		got.FinalExitText != "최종 익절·손절 시 1주 전량" {
		t.Fatalf("one-share explanation = %v/%q/%q", got.OneShare, got.OneShareText, got.FinalExitText)
	}
	if got.ProjectedQuantity != "0 (주문 없음)" {
		t.Errorf("projected quantity = %q, want state-only wording", got.ProjectedQuantity)
	}
}
