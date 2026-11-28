package exitpolicy

// a208 F3 — 편입 전용 보호선 사본(사용자 결정 2026-11-28 「F3 2」).
//
// engine.adoption.floor_only_exit 가 참이면 새 편입은 「설정된 공통 정책」의 단계 트리거 · 보호선 · 러너 트레일을 그대로 쓰고
// 부분익절 비율만 0 인 사본으로 판정됨. 사본은 공통 정책 등록부(commonPolicies)에 넣지 않음 — 설정 · 콘솔 · override 가
// 선택 가능한 정책으로 보지 못하게 하기 위함.

// CommonLadderHybrid50AdoptedFloorOnly 는 HYBRID_50 의 편입 보호선 사본 ID 임. 편입 기록과 exit state 에만 적힘.
const CommonLadderHybrid50AdoptedFloorOnly = "COMMON_LADDER_HYBRID_50_ADOPTED_FLOOR_ONLY"
