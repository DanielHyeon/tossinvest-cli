package config

import "math"

// DefaultMaxOpenPositions 는 최대 보유 포지션 수의 기본값임.
//
// 출처: 사용자 결정 2026-10-15(change a138). 이전에는 보유 종목 수에 상한이 없었으므로 이 값의 등재는 완화가 아니라 새 검사임.
// 개수는 통화와 무관하므로 모든 통화의 게이트에 적용함(a136 하루 매수 기본값이 KRW 전용인 것과 다름).
const DefaultMaxOpenPositions = 6

// maxOpenPositionsCeiling 은 정수 판정이 받는 가장 큰 값 — int 로 안전하게 담기 위한 범위 경계일 뿐 정책 상한이 아님.
const maxOpenPositionsCeiling = math.MaxInt32

// EffectiveMaxOpenPositions 는 엔진이 실제로 쓸 최대 보유 포지션 수를 해석함.
//
// 규칙(a138 D4): 설정값이 0 이 아니면 그대로(음수 · 소수 · NaN · Inf 도 그대로 넘겨 정수 판정이 거부하게 함), 0(키 없음 포함)이면 기본값.
// Guardian 정책 유도와 인터록 대조가 모두 이 함수를 써야 단일 출처가 됨.
func (g AutomationGate) EffectiveMaxOpenPositions() float64 {
	if g.MaxOpenPositions != 0 {
		return g.MaxOpenPositions
	}
	return DefaultMaxOpenPositions
}

// WholeOpenPositions 는 개수 상한으로 쓸 수 있는 값인지 판정하는 **한 곳**임(정책 유도 · 인터록 · 콘솔 검증이 모두 이것을 씀).
//
// 1 이상의 정수이고 int 로 담을 수 있으면 (값, true), 아니면 (0, false). 무한대를 먼저 거르는 이유: math.Trunc(+Inf) == +Inf 라
// 「소수부가 없다」 검사만으로는 무한대가 정수로 통과함(freeze 리뷰 P2-3).
func WholeOpenPositions(v float64) (int, bool) {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 1 || v > maxOpenPositionsCeiling || v != math.Trunc(v) {
		return 0, false
	}
	return int(v), true
}
