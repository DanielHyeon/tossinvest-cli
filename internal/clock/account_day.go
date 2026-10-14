package clock

import "time"

// AccountDayRollover 는 계좌 단위 「하루」가 넘어가는 KST 시각(자정 기준 오프셋)임.
//
// 두 시장을 합친 일일 한도(a136 하루 매수 총액 · a137 일일 손실)는 하루를 하나만 가져야 함. 시장별 거래일
// (Market.TradingDay)은 한 UTC 순간을 서울과 뉴욕에서 다른 날로 읽으므로 계좌의 「오늘」이 되지 못함.
// 08:00 KST 는 두 시장 정규장이 모두 닫힌 창(06:00–09:00 KST) 안이라, 서울 날짜 D 의 KR 정규장과 뉴욕
// 날짜 D 의 US 정규장(22:30/23:30 KST D → 05:00/06:00 KST D+1)이 같은 라벨 D 에 들어가고 어느 정규장도
// 둘로 쪼개지지 않음(사용자 결정 2026-10-14, a136 D1=D).
const AccountDayRollover = 8 * time.Hour

// AccountTradingDay 는 t 가 속한 계좌 단위 하루의 라벨("2006-01-02")을 돌려줌.
//
// 라벨 = 서울 달력 날짜(t − 8시간). 계좌 단위 일일 판정은 전부 이 함수 하나로 날짜를 계산해야 함 —
// 사본을 두면 두 한도가 서로 다른 하루를 쓰게 됨.
func AccountTradingDay(t time.Time) (string, error) {
	return MarketKR.TradingDay(t.Add(-AccountDayRollover))
}
