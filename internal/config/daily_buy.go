package config

import "strings"

// DefaultMaxDailyBuyNotionalKRW 는 하루 매수 총액 상한의 기본값(KRW)임.
//
// 출처: 사용자 결정 2026-10-14(change a136). 이전에는 하루 매수 총액에 상한이 없었으므로 이 값의 등재는 완화가
// 아니라 새 검사임. KRW 게이트에만 적용 — 다른 통화로 환산해 쓰지 않음(spec: 코드 환율 금지).
const DefaultMaxDailyBuyNotionalKRW = 10_000_000

// EffectiveMaxDailyBuyNotional 은 엔진이 실제로 쓸 하루 매수 총액 상한을 해석함.
//
// 규칙(a136 D5):
//   - 설정값이 0 이 아니면 그대로 돌려줌 — 음수·NaN·Inf 도 그대로 넘겨 정책 검증이 기동을 거부하게 함.
//   - 0(키 없음 포함)이고 통화가 KRW 면 기본값.
//   - 0 이고 통화가 KRW 가 아니면 0(미설정) — 승인된 다른 통화 기본값이 없으므로 기동 거부로 이어짐.
//
// Guardian 정책 유도와 인터록 대조가 모두 이 함수를 써야 단일 출처가 됨.
func (g AutomationGate) EffectiveMaxDailyBuyNotional() float64 {
	if g.MaxDailyBuyNotional != 0 {
		return g.MaxDailyBuyNotional
	}
	if strings.EqualFold(strings.TrimSpace(g.LimitCurrency), "KRW") {
		return DefaultMaxDailyBuyNotionalKRW
	}
	return 0
}
