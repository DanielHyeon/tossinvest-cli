package main

// features.go 는 1.7 에서 넓힌 state(공개 시장 데이터만)의 블록과 그 특징 계산을 담당함.
//
// 무엇을 왜 더하는가(design.md 「Jev 가 보는 state (s2)」):
//   - intraday: 오늘 세션의 1분봉(최근 30)과 5/15/30/60분 수익률·창 고저 안 위치·창 VWAP 거리·거래량 가속 —
//     J1(h분 뒤 상승?)이 가격 경로 없이 한 점만 보던 것을 경로로 보게 함.
//   - daily: 직전 거래일 봉으로 1/5/20일 수익률·20일 고저 거리·20일 평균 거래량 — 일 단위 맥락.
//     과거 봉은 하루 동안 변하지 않으므로 (시장 현지 날짜, 종목) 단위로 프로세스 안에 캐시함(호출 예산).
//   - market_context(KR): KOSPI·KOSDAQ 오늘 등락·30분 등락, 시장 전체 투자자별 순매수(원).
//
// 넣지 않은 것: 호가 깊이 — 엄격 리더(StrictOrderbookTop)는 맨 위 한 줄만 내주고, 깊이를 얻으려면 같은
// /orderbook 을 한 번 더 부르거나 internal/official 을 고쳐야 함(이 change 는 internal/ 무접촉). 체결(tape) —
// 공식 체결 응답에 매수/매도 구분이 없고 1분봉이 거래량을 이미 실음.
//   - warning_types·price_limits: 매수 유의(VI·투자경고 등)와 상하한가까지의 거리(J2 악재 판단의 재료).
//
// 단위 규칙: 모든 *_pct 는 퍼센트(1.0 = 1%), 가격은 시세 통화, 거래량은 주 수, 투자자 금액은 원(KRW,
// openapi 「모든 거래대금은 원화(KRW) 정수」). 계산 못 하는 값은 0 이 아니라 생략(null)함 — 0 은 「변화 없음」으로
// 읽히므로.

import (
	"math"
	"sort"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

const (
	// stateRev 는 state 모양의 판본 — 바뀌면 p 의 모집단이 달라지므로 원장 행에 적고 보고서가 로트를 가름.
	// s1(필드 없음) = 1.6 까지의 state, s2 = 1.7 확장.
	stateRev = "a128-s2"

	recentBarCount     = 30  // state 에 싣는 최근 1분봉 수
	minuteFetchCount   = 120 // 1분봉 요청 수(60분 수익률 + 창 고저·VWAP 용)
	dailyFetchCount    = 25  // 일봉 요청 수(직전 20 거래일 + 오늘 + 여유) — 하루 한 번(캐시)
	lookbackTolerance  = 5 * time.Minute
	perStockInvestorNA = "investor_trading_per_stock: the official investor-trading endpoint is market-wide only " +
		"(KOSPI/KOSDAQ); market_context.investor_flows carries the market-wide flow"
)

// marketIndices 는 공식 지표 카탈로그(openapi Market Indicators 「심볼 카탈로그」 8종)의 지수 — US 지수는 카탈로그에 없음.
var marketIndices = map[string][]string{"KR": {"KOSPI", "KOSDAQ"}}

// minuteBar 는 파싱한 1분봉. CloseAt 은 봉이 닫힌 순간(공식 1분봉 timestamp 의 실측 의미, a112 결정 30).
type minuteBar struct {
	CloseAt                       time.Time
	Open, High, Low, Last, Volume float64
}

// MinuteBarState 는 state 에 싣는 1분봉 한 개 — 열쇠를 짧게 해 토큰을 아낌. T = 닫힌 시각(시장 현지 HH:MM).
type MinuteBarState struct {
	T string  `json:"t"`
	O float64 `json:"o"`
	H float64 `json:"h"`
	L float64 `json:"l"`
	C float64 `json:"c"`
	V float64 `json:"v"`
}

// IntradayState 는 오늘 세션(개장 뒤 ~ 판단 시각)에 닫힌 1분봉에서 낸 특징. 창 = 받은 봉 중 세션 안의 것(최대 120).
type IntradayState struct {
	BarsAvailable int              `json:"bars_available"`
	Bars          []MinuteBarState `json:"bars_1m"`
	Return5mPct   *float64         `json:"return_5m_pct,omitempty"`
	Return15mPct  *float64         `json:"return_15m_pct,omitempty"`
	Return30mPct  *float64         `json:"return_30m_pct,omitempty"`
	Return60mPct  *float64         `json:"return_60m_pct,omitempty"`
	WindowHigh    *float64         `json:"window_high,omitempty"`
	WindowLow     *float64         `json:"window_low,omitempty"`
	// PriceInWindowRangePct 는 (기준가 - 창 저가) / (창 고가 - 창 저가) × 100 — 0 = 저가, 100 = 고가.
	PriceInWindowRangePct *float64 `json:"price_in_window_range_pct,omitempty"`
	// WindowVWAPTypical 은 창 봉의 (고+저+종)/3 을 거래량으로 가중한 근사 VWAP — 세션 전체 VWAP 이 아님.
	WindowVWAPTypical            *float64 `json:"window_vwap_typical,omitempty"`
	VsWindowVWAPPct              *float64 `json:"vs_window_vwap_pct,omitempty"`
	VolumeLast5BarsVsEarlierAvg  *float64 `json:"volume_last_5_bars_vs_earlier_avg,omitempty"`
	VolumeLast30BarsVsEarlierAvg *float64 `json:"volume_last_30_bars_vs_earlier_avg,omitempty"`
}

// DailyState 는 일봉(수정주가) 맥락. 직전 거래일 봉 = 시각이 오늘 개장보다 이른 봉.
type DailyState struct {
	PriorBars      int      `json:"prior_bars"`
	LastPriorBarAt string   `json:"last_prior_bar_at,omitempty"`
	Return1dPct    *float64 `json:"return_1d_pct,omitempty"`
	Return5dPct    *float64 `json:"return_5d_pct,omitempty"`
	Return20dPct   *float64 `json:"return_20d_pct,omitempty"`
	High20d        *float64 `json:"high_20d,omitempty"`
	Low20d         *float64 `json:"low_20d,omitempty"`
	VsHigh20dPct   *float64 `json:"vs_high_20d_pct,omitempty"`
	VsLow20dPct    *float64 `json:"vs_low_20d_pct,omitempty"`
	AvgVolume20d   *float64 `json:"avg_volume_20d,omitempty"`
}

// PriceLimitState 는 KR 상하한가와 기준가에서의 거리(퍼센트, 하한은 음수).
type PriceLimitState struct {
	UpperLimit float64  `json:"upper_limit"`
	LowerLimit float64  `json:"lower_limit"`
	ToUpperPct *float64 `json:"to_upper_pct,omitempty"`
	ToLowerPct *float64 `json:"to_lower_pct,omitempty"`
}

// MarketContextState 는 사이클당 한 번 읽는 시장 맥락(KR 만 — 공식 카탈로그에 US 지수가 없음).
type MarketContextState struct {
	Indices       []IndexState        `json:"indices"`
	InvestorFlows []InvestorFlowState `json:"investor_flows"`
}

// IndexState 는 지수 하나 — 오늘 등락(직전 거래일 종가 대비)과 30분 등락.
type IndexState struct {
	Symbol         string   `json:"symbol"`
	Last           float64  `json:"last"`
	ChangeTodayPct *float64 `json:"change_today_pct,omitempty"`
	Change30mPct   *float64 `json:"change_30m_pct,omitempty"`
}

// InvestorFlowState 는 시장 전체(KOSPI/KOSDAQ) 투자자별 순매수(매수 - 매도, 원). 당일 값은 장중 잠정치(openapi).
type InvestorFlowState struct {
	Market            string  `json:"market"`
	Date              string  `json:"date"`
	UpdatedAt         string  `json:"updated_at"`
	IndividualNetKRW  float64 `json:"individual_net_krw"`
	ForeignerNetKRW   float64 `json:"foreigner_net_krw"`
	InstitutionNetKRW float64 `json:"institution_net_krw"`
}

func pctChange(now, then float64) *float64 {
	if then <= 0 || now <= 0 {
		return nil
	}
	value := round((now/then-1)*100, 4)
	return &value
}

func ptr(value float64) *float64 { return &value }

// intradayFeatures 는 개장 뒤·판단 시각 이하에 닫힌 봉만 씀(전일·장전 봉이 섞이면 「오늘」 특징이 아님).
func intradayFeatures(bars []minuteBar, open, at time.Time, base float64) *IntradayState {
	var session []minuteBar
	for _, bar := range bars {
		if bar.CloseAt.After(open) && !bar.CloseAt.After(at) {
			session = append(session, bar)
		}
	}
	sort.Slice(session, func(a, b int) bool { return session[a].CloseAt.Before(session[b].CloseAt) })
	out := &IntradayState{BarsAvailable: len(session), Bars: []MinuteBarState{}}
	for _, bar := range session[max(0, len(session)-recentBarCount):] {
		out.Bars = append(out.Bars, MinuteBarState{
			T: bar.CloseAt.In(open.Location()).Format("15:04"), O: bar.Open, H: bar.High, L: bar.Low, C: bar.Last, V: bar.Volume,
		})
	}
	if len(session) == 0 {
		return out
	}
	// N분 수익률: 닫힌 시각이 [판단-N-5분, 판단-N] 안인 가장 늦은 봉의 종가 대비 — 봉이 비는 저유동 종목에서
	// 훨씬 오래된 봉을 N분 전으로 오인하지 않게 5분 허용폭을 둠.
	lookback := func(minutes int) *float64 {
		target := at.Add(-time.Duration(minutes) * time.Minute)
		for index := len(session) - 1; index >= 0; index-- {
			closeAt := session[index].CloseAt
			if closeAt.After(target) {
				continue
			}
			if closeAt.Before(target.Add(-lookbackTolerance)) {
				return nil
			}
			return pctChange(base, session[index].Last)
		}
		return nil
	}
	out.Return5mPct, out.Return15mPct, out.Return30mPct, out.Return60mPct = lookback(5), lookback(15), lookback(30), lookback(60)

	high, low := math.Inf(-1), math.Inf(1)
	var weighted, volume float64
	for _, bar := range session {
		high, low = math.Max(high, bar.High), math.Min(low, bar.Low)
		weighted += (bar.High + bar.Low + bar.Last) / 3 * bar.Volume
		volume += bar.Volume
	}
	out.WindowHigh, out.WindowLow = ptr(high), ptr(low)
	if high > low {
		out.PriceInWindowRangePct = ptr(round((base-low)/(high-low)*100, 2))
	}
	if volume > 0 {
		vwap := weighted / volume
		out.WindowVWAPTypical, out.VsWindowVWAPPct = ptr(round(vwap, 4)), pctChange(base, vwap)
	}
	out.VolumeLast5BarsVsEarlierAvg = volumeAcceleration(session, 5)
	out.VolumeLast30BarsVsEarlierAvg = volumeAcceleration(session, 30)
	return out
}

// volumeAcceleration 은 마지막 n 봉의 봉당 평균 거래량 ÷ 그 앞 봉들의 봉당 평균. 앞 봉이 없거나 평균이 0 이면 생략.
func volumeAcceleration(session []minuteBar, n int) *float64 {
	if len(session) <= n {
		return nil
	}
	var recent, earlier float64
	for _, bar := range session[len(session)-n:] {
		recent += bar.Volume
	}
	for _, bar := range session[:len(session)-n] {
		earlier += bar.Volume
	}
	earlierAvg := earlier / float64(len(session)-n)
	if earlierAvg <= 0 {
		return nil
	}
	return ptr(round(recent/float64(n)/earlierAvg, 3))
}

// dailyFeatures 는 직전 거래일 봉(시각 < 오늘 개장)만 씀 — 오늘 봉은 캐시된 응답에서 낡으므로 버림(오늘의
// 움직임은 intraday 가 실음).
// 일봉 timestamp 의 의미(여는/닫는 시각)는 실측되지 않았음 — 이 경계는 「닫는 시각(전일 장 마감 또는 자정)」과
// 「개장 시각」 두 해석 모두에서 같은 답을 냄. 「자기 날짜의 자정」 표기라면 오늘 봉이 직전으로 섞이므로
// last_prior_bar_at 을 state 에 남겨 원장의 state 파일로 다시 검증할 수 있게 함.
func dailyFeatures(candles []domain.Candle, open time.Time, base float64) *DailyState {
	var prior []domain.Candle
	for _, candle := range candles {
		if !candle.Time.IsZero() && candle.Time.Before(open) {
			prior = append(prior, candle)
		}
	}
	sort.Slice(prior, func(a, b int) bool { return prior[a].Time.Before(prior[b].Time) })
	n := len(prior)
	out := &DailyState{PriorBars: n}
	if n == 0 {
		return out
	}
	out.LastPriorBarAt = prior[n-1].Time.UTC().Format(time.RFC3339)
	back := func(sessions int) *float64 {
		if n < sessions {
			return nil
		}
		return pctChange(base, prior[n-sessions].Close)
	}
	out.Return1dPct, out.Return5dPct, out.Return20dPct = back(1), back(5), back(20)
	if n >= 20 {
		high, low, volume := math.Inf(-1), math.Inf(1), 0.0
		for _, candle := range prior[n-20:] {
			high, low = math.Max(high, candle.High), math.Min(low, candle.Low)
			volume += candle.Volume
		}
		out.High20d, out.Low20d = ptr(high), ptr(low)
		out.VsHigh20dPct, out.VsLow20dPct = pctChange(base, high), pctChange(base, low)
		out.AvgVolume20d = ptr(round(volume/20, 2))
	}
	return out
}

// priceLimitFeatures 는 상하한가가 둘 다 있을 때만 블록을 만듦(US 는 null → 0 으로 오므로 0 띠를 만들지 않음).
func priceLimitFeatures(limits domain.PriceLimits, base float64) *PriceLimitState {
	if limits.UpperLimit <= 0 || limits.LowerLimit <= 0 {
		return nil
	}
	return &PriceLimitState{
		UpperLimit: limits.UpperLimit, LowerLimit: limits.LowerLimit,
		ToUpperPct: pctChange(limits.UpperLimit, base), ToLowerPct: pctChange(limits.LowerLimit, base),
	}
}

// indexFeatures 는 지수 하나. 오늘 등락 = 시각이 개장보다 이른 가장 늦은 일봉 종가 대비, 30분 등락 = 닫힌 시각이
// [판단-35분, 판단-30분] 안이고 개장 뒤인 가장 늦은 1분봉 종가 대비(지수 봉 timestamp 의미는 종목 봉과 같다고
// 보지만 실측되지 않음 — 1분 차이).
func indexFeatures(symbol string, last float64, daily, minute []domain.MarketIndicatorCandle, open, at time.Time) IndexState {
	out := IndexState{Symbol: symbol, Last: last}
	var priorAt time.Time
	var priorClose float64
	for _, candle := range daily {
		stamp, err := time.Parse(time.RFC3339, candle.Timestamp)
		if err == nil && stamp.Before(open) && stamp.After(priorAt) {
			priorAt, priorClose = stamp, candle.Close
		}
	}
	out.ChangeTodayPct = pctChange(last, priorClose)
	target := at.Add(-30 * time.Minute)
	var pickAt time.Time
	var pickClose float64
	for _, candle := range minute {
		stamp, err := time.Parse(time.RFC3339, candle.Timestamp)
		if err != nil || !stamp.After(open) || stamp.After(target) || stamp.Before(target.Add(-lookbackTolerance)) {
			continue
		}
		if stamp.After(pickAt) {
			pickAt, pickClose = stamp, candle.Close
		}
	}
	out.Change30mPct = pctChange(last, pickClose)
	return out
}
