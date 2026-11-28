package exitpolicy

// a208 F3 — 편입 전용 보호선 사본(사용자 결정 2026-11-28 「F3 2」).
//
// engine.adoption.floor_only_exit 가 참이면 새 편입은 「설정된 공통 정책」의 단계 트리거 · 보호선 · 러너 트레일을 그대로 쓰고
// 부분익절 비율만 0 인 사본으로 판정됨. 사본은 공통 정책 등록부(commonPolicies)에 넣지 않음 — 설정 · 콘솔 · override 가
// 선택 가능한 정책으로 보지 못하게 하기 위함. 이미 있는 ID 의 해소 갈래는 건드리지 않으므로 이미 편입된 행의 의미는 그대로임.

import (
	"fmt"
	"strings"
)

// CommonLadderHybrid50AdoptedFloorOnly 는 HYBRID_50 의 편입 보호선 사본 ID 임. 편입 기록과 exit state 에만 적힘.
const CommonLadderHybrid50AdoptedFloorOnly = "COMMON_LADDER_HYBRID_50_ADOPTED_FLOOR_ONLY"

const (
	// 편입 사본의 version 은 편입 RUNNER 사본과 같은 표기(「편입 사본 1 판」) — ID 가 달라 digest 는 따로 고정됨.
	adoptedFloorOnlyPolicyVersion = "1.0.0-adopted.1"
	// HYBRID_50 표(1.8/0 · 3.0/1.2 · 4.8/2.5 · 6.5/3.8, 트레일 6.5) + 부분익절 0 의 고정 digest. 표가 바뀌면 판정을 멈춤.
	hybrid50AdoptedFloorOnlyPolicyDigest = "sha256:cb8039a3b4f20f8482d4ad3d503137f626e4fb33ce6ffc9cbe407c68498998c5"
)

// AdoptedFloorOnlyPolicyID 는 설정된 공통 정책의 편입 보호선 사본 ID 를 돌려줌. 거짓이면 사본이 없어 토글을 무시해야 함.
//
//   - HYBRID_50 → 새 사본 ID(트리거 · 보호선 · 트레일 동일, 부분익절 0).
//   - RUNNER → RUNNER 그대로: 기존 편입 RUNNER 사본(1.0.0-adopted.1)이 이미 「RUNNER 표 + 부분익절 0」이라 새 정체성이 필요 없음.
//   - BALANCED → 없음: 최종 rung 이 전량 익절(FinalTakeFull)이라 부분익절만 0 으로는 「보호선만」이 성립하지 않음.
//   - 공통 정책 없음(RATCHET) → 없음: 사다리 표가 없음.
func AdoptedFloorOnlyPolicyID(commonPolicy string) (string, bool) {
	switch strings.TrimSpace(commonPolicy) {
	case CommonLadderHybrid50:
		return CommonLadderHybrid50AdoptedFloorOnly, true
	case CommonLadderRunner:
		return CommonLadderRunner, true
	}
	return "", false
}

// IsAdoptedFloorOnlyPolicyID 는 정책 ID 가 등록부 밖의 편입 보호선 사본인지 알려줌(원장의 정책 ID 문이 씀).
func IsAdoptedFloorOnlyPolicyID(id string) bool {
	return strings.TrimSpace(id) == CommonLadderHybrid50AdoptedFloorOnly
}

// adoptedFloorOnlyIdentity 는 사본의 고정 정체성임. 편입이 아닌 포지션에는 의미가 없어 거부함(fail-closed).
func adoptedFloorOnlyIdentity(adopted bool) (PolicyIdentity, error) {
	if !adopted {
		return PolicyIdentity{}, fmt.Errorf("%w: %s is an adoption-only copy and has no meaning for an engine-entered position",
			ErrPolicyIdentityConflict, CommonLadderHybrid50AdoptedFloorOnly)
	}
	identity := PolicyIdentity{
		ID: CommonLadderHybrid50AdoptedFloorOnly, Version: adoptedFloorOnlyPolicyVersion,
		Digest: hybrid50AdoptedFloorOnlyPolicyDigest,
	}
	return identity, identity.Validate()
}

// adoptedFloorOnlyLadder 는 등록된 HYBRID_50 표의 복사본에서 부분익절만 0 으로 바꾼 사본을 만듦.
// 표 값을 따로 적지 않고 등록부에서 유도하므로 보호선이 공통 정책과 갈릴 수 없고, 고정 digest 가 표류를 잡음.
func adoptedFloorOnlyLadder(adopted bool) (LadderPolicy, error) {
	identity, err := adoptedFloorOnlyIdentity(adopted)
	if err != nil {
		return LadderPolicy{}, err
	}
	source, ok := CommonPolicyByID(CommonLadderHybrid50)
	if !ok {
		return LadderPolicy{}, fmt.Errorf("%w: the floor-only copy's source %s is not registered",
			ErrPolicyIdentityConflict, CommonLadderHybrid50)
	}
	ladder := source.Ladder // CommonPolicyByID 가 rung 슬라이스까지 복사해 줌 — 전역 표 무변
	for i := range ladder.Rungs {
		ladder.Rungs[i].PartialRatio = "0"
	}
	ladder.PolicyID, ladder.PolicyVersion, ladder.PolicyDigest = identity.ID, identity.Version, identity.Digest
	// Identity 는 주장한 digest 와 다시 계산한 값이 다르면 정체성 충돌을 돌려줌 — 원본 표가 바뀌면 여기서 판정이 멈춤.
	if _, err := ladder.Identity(); err != nil {
		return LadderPolicy{}, err
	}
	return ladder, nil
}
