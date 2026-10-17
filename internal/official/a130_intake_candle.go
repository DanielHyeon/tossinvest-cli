package official

// a130_intake_candle.go 는 a130 backfill 도구가 늘어난 편입 묶음의 편입가로 쓰는 공식 1분봉 하나를 읽음(design D1). 읽기 전용 GET 이고
// 주문 능력이 아님. 원장 패키지는 네트워크를 부르지 않으므로 이 읽기는 도구(tools/a130-outcome-backfill)에서만 부름.
//
// 새 전송 경로를 만들지 않고 a112 의 엄격한 1분봉 리더(StrictMinuteCandles — KR/US 공통, adjusted=false, 응답 문법 · 통화 · 소수 ·
// 분 경계 검증, 재시도 없음)를 count=1 로 부름. 이 파일이 더하는 것은 「돌려받은 봉의 이름이 요청한 분과 같을 때만」 이라는 판정 하나.

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// intakeCandleBeforeZone 은 before 를 적는 시간대임 — 엄격 리더의 before 문법은 숫자 오프셋을 요구하고, 생산 소비자
// (internal/officialbars)도 서울 시각 오프셋으로 보냄. 같은 순간이므로 판정은 시간대와 무관함.
var intakeCandleBeforeZone = time.FixedZone("KST", 9*60*60)

// MinuteCandleClosingAt 은 이름(봉 **종료** 시각)이 label 인 1분봉 하나를 읽어 종가와 **돌려받은 봉의 이름**을 돌려줌.
//
// 공식 분봉의 timestamp 는 봉이 닫히는 시각이고 t 봉은 [t−1분, t) 를 덮음(a112 결정 30 · 31 실측). before 는 포함 상한이므로
// before=label · count=1 이 그 봉을 돌려줌. 그 분에 체결이 없어 봉이 없으면 API 는 더 이른 봉을 돌려주므로, 돌려받은 봉의 이름이 label 과
// 같을 때만 ok — 다른 분의 봉을 이 분의 가격으로 쓰지 않음. 호출자는 돌려받은 이름을 그대로 기록하고 원장 판정(priceLots)이 그것을
// 묶음의 분과 다시 대조함(두 번째 가드).
func (c *Client) MinuteCandleClosingAt(ctx context.Context, market, symbol string, label time.Time) (
	closePrice string, barLabel time.Time, ok bool, err error) {
	if c == nil {
		return "", time.Time{}, false, fmt.Errorf("official minute candle: client is required")
	}
	page, err := c.StrictMinuteCandles(ctx, strings.ToUpper(strings.TrimSpace(market)), strings.TrimSpace(symbol), 1,
		label.In(intakeCandleBeforeZone).Format("2006-01-02T15:04:05-07:00"))
	if err != nil {
		return "", time.Time{}, false, err
	}
	for _, candle := range page.Candles {
		at, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(candle.Timestamp))
		if err != nil {
			return "", time.Time{}, false, fmt.Errorf("official minute candle: unreadable timestamp %q", candle.Timestamp)
		}
		if at.Equal(label) {
			return strings.TrimSpace(candle.Close), at.UTC(), strings.TrimSpace(candle.Close) != "", nil
		}
		return "", at.UTC(), false, nil
	}
	return "", time.Time{}, false, nil
}
