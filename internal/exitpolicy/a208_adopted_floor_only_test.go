package exitpolicy_test

// a208 F3 — HYBRID_50 편입 보호선 사본: 설정 공통 정책과 트리거 · 보호선 · 러너 트레일이 같고 부분익절만 0, 고정 정체성,
// 편입 전용, 공통 정책 등록부 밖(선택 불가). 이미 편입된 HYBRID_50 행의 의미는 바뀌지 않음.

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

const a208FloorOnlyDigest = "sha256:cb8039a3b4f20f8482d4ad3d503137f626e4fb33ce6ffc9cbe407c68498998c5"

func TestA208FloorOnlyCopyKeepsTheHybridFloors(t *testing.T) {
	t.Parallel()

	floorID := exitpolicy.CommonLadderHybrid50AdoptedFloorOnly
	copyLadder, err := exitpolicy.CommonLadderForPosition(floorID, true)
	if err != nil {
		t.Fatalf("CommonLadderForPosition(copy): %v", err)
	}
	common, ok := exitpolicy.CommonPolicyByID(exitpolicy.CommonLadderHybrid50)
	if !ok {
		t.Fatal("HYBRID_50 is not registered")
	}
	if len(copyLadder.Rungs) != len(common.Ladder.Rungs) || copyLadder.RunnerTrailPct != common.Ladder.RunnerTrailPct ||
		copyLadder.FinalTakeFull != common.Ladder.FinalTakeFull {
		t.Fatalf("copy = %+v, want the HYBRID_50 shape %+v", copyLadder, common.Ladder)
	}
	for i, rung := range copyLadder.Rungs {
		want := common.Ladder.Rungs[i]
		if rung.TargetPct != want.TargetPct || rung.StopPct != want.StopPct {
			t.Errorf("rung %d = %s/%s, want %s/%s", i, rung.TargetPct, rung.StopPct, want.TargetPct, want.StopPct)
		}
		if rung.PartialRatio != "0" {
			t.Errorf("rung %d partial = %s, want 0", i, rung.PartialRatio)
		}
	}
	identity, err := copyLadder.Identity()
	if err != nil {
		t.Fatalf("Identity: %v", err)
	}
	want := exitpolicy.PolicyIdentity{ID: floorID, Version: "1.0.0-adopted.1", Digest: a208FloorOnlyDigest}
	if identity != want {
		t.Fatalf("copy identity = %+v, want pinned %+v", identity, want)
	}
	seeded, err := exitpolicy.LegacyLadderPolicyIdentity(floorID, true)
	if err != nil || seeded != want {
		t.Fatalf("seed identity = %+v, %v, want %+v", seeded, err, want)
	}
}

func TestA208FloorOnlyCopyIsAdoptionOnly(t *testing.T) {
	t.Parallel()

	floorID := exitpolicy.CommonLadderHybrid50AdoptedFloorOnly
	if _, err := exitpolicy.CommonLadderForPosition(floorID, false); err == nil {
		t.Fatal("an engine-entered position resolved the adoption-only copy")
	}
	if _, err := exitpolicy.LegacyLadderPolicyIdentity(floorID, false); !errors.Is(err, exitpolicy.ErrPolicyIdentityConflict) {
		t.Fatalf("non-adopted seed error = %v, want policy identity conflict", err)
	}
}

// 기존 갈래를 넓히지 않음 — 편입 HYBRID_50 행은 여전히 공통 HYBRID_50 표(부분익절 포함) · 1.0.0 정체성.
func TestA208AdoptedHybridRowsKeepTheirMeaning(t *testing.T) {
	t.Parallel()

	common, _ := exitpolicy.CommonPolicyByID(exitpolicy.CommonLadderHybrid50)
	adopted, err := exitpolicy.CommonLadderForPosition(exitpolicy.CommonLadderHybrid50, true)
	if err != nil {
		t.Fatal(err)
	}
	for i, rung := range adopted.Rungs {
		if rung != common.Ladder.Rungs[i] {
			t.Errorf("adopted HYBRID_50 rung %d = %+v, want the registered %+v", i, rung, common.Ladder.Rungs[i])
		}
	}
	if adopted.PolicyVersion != common.Ladder.PolicyVersion || adopted.PolicyDigest != common.Ladder.PolicyDigest {
		t.Fatalf("adopted HYBRID_50 identity = %s/%s, want %s/%s", adopted.PolicyVersion, adopted.PolicyDigest,
			common.Ladder.PolicyVersion, common.Ladder.PolicyDigest)
	}
	identity, err := exitpolicy.LegacyLadderPolicyIdentity(exitpolicy.CommonLadderHybrid50, true)
	if err != nil || identity.Version != exitpolicy.DefaultPolicyVersion || identity.Digest != common.Ladder.PolicyDigest {
		t.Fatalf("adopted HYBRID_50 seed identity = %+v, %v", identity, err)
	}
}

// 사본은 등록부 밖 — 콘솔 · 설정 · override 의 선택지에 나타나지 않음.
func TestA208FloorOnlyCopyIsNotSelectable(t *testing.T) {
	t.Parallel()

	floorID := exitpolicy.CommonLadderHybrid50AdoptedFloorOnly
	if _, ok := exitpolicy.CommonPolicyByID(floorID); ok {
		t.Fatal("the floor-only copy is a registered common policy")
	}
	for _, policy := range exitpolicy.RegisteredCommonPolicies() {
		if policy.ID == floorID {
			t.Fatal("the floor-only copy is listed among the registered common policies")
		}
	}
	for _, option := range exitpolicy.CommonPolicyFieldDescriptor().Options {
		if option.Value == floorID {
			t.Fatal("the floor-only copy is a selectable common-policy option")
		}
	}
}
