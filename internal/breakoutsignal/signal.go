// Package breakoutsignal 은 저장된 공식 닫힌 1분봉 묶음에서 breakout v1 평가기가 요구하는 파생 지표 넷을 만듦.
//
// 왜 있는가: 생산 제안 적재기는 breakout 범위를 레인 ID 만 보고 거절함(strategyproposal.buildLaneInput 의
// ErrBreakoutEvidenceUnavailable) — ATR · RVOL · 윗꼬리 비율 · 거래량 확장을 만드는 코드가 없었기 때문.
// 이 패키지가 그 계산 하나만 맡음. 생산 호출자는 아직 없음(a140 은 그림자 측정 도구만 부름).
//
// 계약(openspec breakout-retest-strategy-lane 「breakout 파생 지표는 저장된 공식 닫힌 1분봉에서 결정적으로 만들어진다」):
//   - 정수 산술만 · 시계 · 네트워크 · 부동소수 없음(signal_test.go 가 AST 로 고정).
//   - 입력 봉은 정규장 개장 분부터 빈 분 없이 이어져야 함 — 평가기는 봉 번호로 opening range 와 timeout 을 셈.
//   - 못 만드는 값은 0 이 아니라 typed 거절(*Refusal).
package breakoutsignal

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// Version 은 지표 정의의 판본 — 원장 행과 digest 에 박혀 정의가 바뀌면 모집단이 갈리게 함.
const Version = "breakoutsignal-v1"

const (
	// OpeningRangeBars 는 breakout v1 의 opening range 봉 수(평가기 동결 상수와 같아야 함 — 시험이 행동으로 대조).
	OpeningRangeBars = 15
	// atrPeriod 는 Wilder ATR 의 기간.
	atrPeriod = 14
	// rvolWindow 는 RVOL 기준 거래량의 직전 봉 수.
	rvolWindow = 15
	ppmScale   = 1_000_000
	barMS      = 60_000
)

// RefusalCode 는 지표를 만들 수 없는 이유의 이름.
type RefusalCode string

const (
	RefusalSeriesInvalid          RefusalCode = "SERIES_INVALID"
	RefusalBarInvalid             RefusalCode = "BAR_INVALID"
	RefusalNotAtOpen              RefusalCode = "FIRST_BAR_NOT_AT_OPEN"
	RefusalMinuteGap              RefusalCode = "MINUTE_GAP"
	RefusalOpeningRangeIncomplete RefusalCode = "OPENING_RANGE_INCOMPLETE"
	RefusalATRZero                RefusalCode = "ATR_ZERO"
	RefusalRVOLUndefined          RefusalCode = "RVOL_UNDEFINED"
	RefusalDenominatorZero        RefusalCode = "DENOMINATOR_ZERO"
	RefusalOverflow               RefusalCode = "OVERFLOW"
)

// Refusal 은 typed 거절. Detail 은 진단용 문장(계약 아님).
type Refusal struct {
	Code   RefusalCode
	Detail string
}

func (r *Refusal) Error() string { return "breakoutsignal: " + string(r.Code) + ": " + r.Detail }

func refuse(code RefusalCode, detail string) error { return &Refusal{Code: code, Detail: detail} }

// Bar 는 지표 계산에 쓰는 봉 값(minor 단위 가격 · 주 단위 거래량).
type Bar struct {
	OpenAtMS                                   uint64
	OpenMinor, HighMinor, LowMinor, CloseMinor uint64
	Volume                                     uint64
}

func barFromPayload(p strategyevidence.ClosedBar1mPayload) Bar {
	return Bar{OpenAtMS: p.OpenAtMS, OpenMinor: p.OpenMinor, HighMinor: p.HighMinor, LowMinor: p.LowMinor,
		CloseMinor: p.CloseMinor, Volume: p.Volume}
}

func validBar(b Bar) bool {
	return b.OpenMinor > 0 && b.HighMinor > 0 && b.LowMinor > 0 && b.CloseMinor > 0 &&
		b.LowMinor <= b.HighMinor && b.OpenMinor >= b.LowMinor && b.OpenMinor <= b.HighMinor &&
		b.CloseMinor >= b.LowMinor && b.CloseMinor <= b.HighMinor
}

// UpperWickRangePPM 은 ceil((high − max(open, close)) × 1e6 / (high − low)). high = low 면 0.
// 올림인 이유: 평가기는 이 값을 상한(35%) 거부권으로 씀 — 올림은 거부 쪽으로 틀림(보수).
func UpperWickRangePPM(b Bar) (uint64, error) {
	if !validBar(b) {
		return 0, refuse(RefusalBarInvalid, "open/high/low/close are not a positive, ordered bar")
	}
	if b.HighMinor == b.LowMinor {
		return 0, nil
	}
	body := max(b.OpenMinor, b.CloseMinor)
	wick := new(big.Int).SetUint64(b.HighMinor - body)
	rng := new(big.Int).SetUint64(b.HighMinor - b.LowMinor)
	return ratioCeil(wick, rng)
}

// trueRange 는 Wilder 의 TR. 첫 봉(이전 종가 없음)은 high − low.
func trueRange(b Bar, previous *Bar) uint64 {
	tr := b.HighMinor - b.LowMinor
	if previous == nil {
		return tr
	}
	prevClose := previous.CloseMinor
	if up := absDiff(b.HighMinor, prevClose); up > tr {
		tr = up
	}
	if down := absDiff(b.LowMinor, prevClose); down > tr {
		tr = down
	}
	return tr
}

// OpeningRangeATR 은 opening range 15 봉의 Wilder ATR(14) 를 유리수로 계산한 뒤 minor 내림.
//
//	ATR₁₄ = (TR₁ + … + TR₁₄)/14,  ATR₁₅ = (13·ATR₁₄ + TR₁₅)/14 = (13·ΣTR₁..₁₄ + 14·TR₁₅)/196
//
// 창을 opening range 로 고정하는 이유: 평가기는 매 분 전체 봉에서 다시 유도하므로 ATR 이 분마다 바뀌면 이미 판정된
// 돌파 · 되돌림이 소급해 뒤집힘. setup 은 RANGE_LOCKED(15 번째 봉)에서 정의되므로 그 순간의 ATR 이 고정점.
// 중간에 내리지 않는 이유: 먼저 내리면 마지막 단계에서 1 minor 가 사라질 수 있음(시험이 그 차이를 잡음).
func OpeningRangeATR(bars []Bar) (uint64, error) {
	if len(bars) < OpeningRangeBars {
		return 0, refuse(RefusalOpeningRangeIncomplete, strconv.Itoa(len(bars))+" bars, the range needs "+strconv.Itoa(OpeningRangeBars))
	}
	sum := new(big.Int)
	for i := 0; i < atrPeriod; i++ {
		if !validBar(bars[i]) {
			return 0, refuse(RefusalBarInvalid, "opening-range bar "+strconv.Itoa(i)+" is not a positive, ordered bar")
		}
		var previous *Bar
		if i > 0 {
			previous = &bars[i-1]
		}
		sum.Add(sum, new(big.Int).SetUint64(trueRange(bars[i], previous)))
	}
	last := bars[OpeningRangeBars-1]
	if !validBar(last) {
		return 0, refuse(RefusalBarInvalid, "the last opening-range bar is not a positive, ordered bar")
	}
	numerator := new(big.Int).Mul(sum, big.NewInt(atrPeriod-1))
	numerator.Add(numerator, new(big.Int).Mul(big.NewInt(atrPeriod), new(big.Int).SetUint64(trueRange(last, &bars[OpeningRangeBars-2]))))
	atr := new(big.Int).Quo(numerator, big.NewInt(atrPeriod*atrPeriod))
	if !atr.IsUint64() {
		return 0, refuse(RefusalOverflow, "the ATR does not fit an unsigned 64-bit minor amount")
	}
	if atr.Sign() == 0 {
		return 0, refuse(RefusalATRZero, "the opening range has no true range")
	}
	return atr.Uint64(), nil
}

// RVOLPPM 은 opening range 뒤의 봉 index 에 대해 floor(volume × 1e6 × 15 / 직전 15 봉 거래량 합).
// opening range 안의 봉은 정의되지 않음(직전 15 봉이 같은 세션에 없음) — 거절.
func RVOLPPM(bars []Bar, index int) (uint64, error) {
	if index < 0 || index >= len(bars) {
		return 0, refuse(RefusalSeriesInvalid, "bar index "+strconv.Itoa(index)+" is outside the series")
	}
	if index < OpeningRangeBars {
		return 0, refuse(RefusalRVOLUndefined, "bar "+strconv.Itoa(index)+" is inside the opening range")
	}
	window := new(big.Int)
	for i := index - rvolWindow; i < index; i++ {
		window.Add(window, new(big.Int).SetUint64(bars[i].Volume))
	}
	if window.Sign() == 0 {
		return 0, refuse(RefusalDenominatorZero, "the fifteen bars before bar "+strconv.Itoa(index)+" traded no volume")
	}
	numerator := new(big.Int).Mul(new(big.Int).SetUint64(bars[index].Volume), big.NewInt(ppmScale*rvolWindow))
	value := numerator.Quo(numerator, window)
	if !value.IsUint64() {
		return 0, refuse(RefusalOverflow, "RVOL does not fit an unsigned 64-bit ppm value")
	}
	return value.Uint64(), nil
}

// VolumeExpanded 는 RVOL ≥ 1.0×. 평가기는 되돌림 중 「거래량 늘며 저항 아래 마감」을 실패로 읽음.
func VolumeExpanded(rvolPPM uint64) bool { return rvolPPM >= ppmScale }

// BarMetrics 는 봉 하나의 지표와 그 봉의 신원.
type BarMetrics struct {
	Record            strategyevidence.ClosedBarRecord
	RVOLPPM           uint64
	RVOLDefined       bool
	UpperWickRangePPM uint64
	VolumeExpanded    bool
}

// Metrics 는 한 세션 묶음의 지표 전부.
type Metrics struct {
	Version   string
	Market    string
	Symbol    string
	SessionID string
	ATRMinor  uint64
	Bars      []BarMetrics
	Digest    string
}

// Derive 는 봉인된 묶음과 그 세션의 정규장 개장 시각으로 지표를 만듦.
//
// 묶음 검사(전부 typed 거절): 같은 시장 · 종목 · 세션, 1분 · 닫힘 · 정규장 봉, 첫 봉 = 개장 분, 빈 분 없음,
// opening range 완성. 봉 하나라도 값이 불가능하면 묶음 전체를 거절(일부만 쓰면 봉 번호가 분과 어긋남).
func Derive(series strategyevidence.BarSeries, regularOpen time.Time) (Metrics, error) {
	records := series.Bars
	if len(records) == 0 {
		return Metrics{}, refuse(RefusalOpeningRangeIncomplete, "the series is empty")
	}
	first := records[0].Payload
	openMS := uint64(regularOpen.UnixMilli())
	if regularOpen.IsZero() || first.OpenAtMS != openMS {
		return Metrics{}, refuse(RefusalNotAtOpen, "the first bar opens at "+formatMS(first.OpenAtMS)+", the session opens at "+formatMS(openMS))
	}
	bars := make([]Bar, len(records))
	for i, record := range records {
		p := record.Payload
		switch {
		case p.Market != first.Market || p.Symbol != first.Symbol || p.SessionID != first.SessionID:
			return Metrics{}, refuse(RefusalSeriesInvalid, "bar "+strconv.Itoa(i)+" belongs to another market, symbol or session")
		case p.IntervalMS != barMS || !p.Closed || !p.RegularSession:
			return Metrics{}, refuse(RefusalSeriesInvalid, "bar "+strconv.Itoa(i)+" is not a closed regular-session one-minute bar")
		case strings.TrimSpace(record.EvidenceID) == "":
			return Metrics{}, refuse(RefusalSeriesInvalid, "bar "+strconv.Itoa(i)+" has no evidence identity")
		}
		expected := openMS + uint64(i)*barMS
		if p.OpenAtMS != expected {
			return Metrics{}, refuse(RefusalMinuteGap, "the minute opening at "+formatMS(expected)+" is missing")
		}
		bars[i] = barFromPayload(p)
		if !validBar(bars[i]) {
			return Metrics{}, refuse(RefusalBarInvalid, "bar "+strconv.Itoa(i)+" is not a positive, ordered bar")
		}
	}
	atr, err := OpeningRangeATR(bars)
	if err != nil {
		return Metrics{}, err
	}
	out := Metrics{Version: Version, Market: first.Market, Symbol: first.Symbol, SessionID: first.SessionID,
		ATRMinor: atr, Bars: make([]BarMetrics, len(bars))}
	for i := range bars {
		wick, err := UpperWickRangePPM(bars[i])
		if err != nil {
			return Metrics{}, err
		}
		metric := BarMetrics{Record: records[i], UpperWickRangePPM: wick}
		// opening range 봉은 RVOL 이 정의되지 않음 — 칸은 0 · 거짓으로 남고 평가기는 읽지 않음(bars[15:] 만 봄).
		if i >= OpeningRangeBars {
			rvol, err := RVOLPPM(bars, i)
			if err != nil {
				return Metrics{}, err
			}
			metric.RVOLPPM, metric.RVOLDefined, metric.VolumeExpanded = rvol, true, VolumeExpanded(rvol)
		}
		out.Bars[i] = metric
	}
	out.Digest = digestOf(out)
	return out, nil
}

// LaneBars 는 지표를 동결 평가기의 봉 입력으로 옮김. Sequence 는 개장부터 1, 2, 3 …(빈 분이 없음은 Derive 가 보장).
func (m Metrics) LaneBars() ([]breakoutlane.ClosedBar, error) {
	out := make([]breakoutlane.ClosedBar, len(m.Bars))
	for i, metric := range m.Bars {
		p := metric.Record.Payload
		bar, err := breakoutlane.NewClosedBar(breakoutlane.ClosedBarInput{
			Sequence: uint64(i + 1), Revision: p.Revision, IntervalMS: p.IntervalMS,
			HighMinor: p.HighMinor, LowMinor: p.LowMinor, CloseMinor: p.CloseMinor,
			RVOLPPM: metric.RVOLPPM, UpperWickRangePPM: metric.UpperWickRangePPM,
			ID: LaneBarID(p), SessionID: p.SessionID,
			RegularSession: p.RegularSession, Closed: p.Closed, VolumeExpanded: metric.VolumeExpanded,
		})
		if err != nil {
			return nil, refuse(RefusalBarInvalid, "bar "+strconv.Itoa(i)+" is refused by the evaluator: "+err.Error())
		}
		out[i] = bar
	}
	return out, nil
}

// LaneBarID 는 판(revision)과 무관한 봉 신원 「세션 ID:여는 시각(ms)」.
// 평가기의 setupID 가 첫 봉 · 15 번째 봉의 ID 를 쓰므로(breakoutlane setupID), 판마다 바뀌는 EvidenceID 를 쓰면
// 봉 정정 하나로 같은 setup 이 둘이 됨. 판은 ClosedBarInput.Revision 으로만 넘김.
func LaneBarID(p strategyevidence.ClosedBar1mPayload) string {
	return p.SessionID + ":" + u(p.OpenAtMS)
}

// digestOf 는 판본 · 신원 · 봉 신원 · 지표를 NUL 로 이은 sha256. 같은 입력 = 같은 바이트.
func digestOf(m Metrics) string {
	parts := []string{"tossos.breakoutsignal.metrics", m.Version, m.Market, m.Symbol, m.SessionID, u(m.ATRMinor)}
	for _, b := range m.Bars {
		p := b.Record.Payload
		parts = append(parts, b.Record.EvidenceID, b.Record.RevisionIdentity, b.Record.PayloadDigest, u(p.OpenAtMS),
			u(p.Volume), u(b.RVOLPPM), strconv.FormatBool(b.RVOLDefined), u(b.UpperWickRangePPM), strconv.FormatBool(b.VolumeExpanded))
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ratioCeil 은 ceil(n × 1e6 / d), d > 0.
func ratioCeil(n, d *big.Int) (uint64, error) {
	scaled := new(big.Int).Mul(n, big.NewInt(ppmScale))
	q, r := new(big.Int).QuoRem(scaled, d, new(big.Int))
	if r.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsUint64() {
		return 0, refuse(RefusalOverflow, "the ratio does not fit an unsigned 64-bit ppm value")
	}
	return q.Uint64(), nil
}

func absDiff(a, b uint64) uint64 {
	if a > b {
		return a - b
	}
	return b - a
}

func u(v uint64) string { return strconv.FormatUint(v, 10) }

// formatMS 는 진단 문장용 UTC 시각.
func formatMS(ms uint64) string {
	return time.UnixMilli(int64(ms)).UTC().Format(time.RFC3339)
}
