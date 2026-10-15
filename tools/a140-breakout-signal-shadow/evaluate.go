package main

// evaluate.go 는 봉 묶음 하나를 동결 breakout v1 평가기에 넣어 단계 · 거절 · 가상 제안을 얻음 — 브로커 · 시계 · 파일 없음.

import (
	"errors"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutsignal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// shadowParams 는 그림자 설정(D21 기본값 · D7 가상 사이징).
type shadowParams struct {
	RetestTolerancePPM, MaxQuoteAgeMS, MaxSpreadPPM, MaxEntryDriftPPM uint64
	RiskBudgetKRW, NotionalCapKRW                                     uint64
	CostPPM, MinRRPPM                                                 uint64
}

// quoteSnap 은 봉인된 호가 한 줄(evidence 본문에서 옮긴 정수 값).
type quoteSnap struct {
	Bid, Ask, Last         uint64
	ObservedMS, ReceivedMS uint64
	Currency, Digest       string
}

// sessionInfo 는 그 세션의 신원과 정규장 창.
type sessionInfo struct {
	ID, CalendarVersion string
	Open, Close         time.Time
}

// evaluation 은 평가 한 번의 결과 — 원장 signal 행의 재료.
type evaluation struct {
	Phase, Refusal                        string
	SetupID, SnapshotDigest, ConfigDigest string
	MetricsDigest                         string
	Transitions                           []string
	Bars                                  int
	Tick, ATRMinor, Resistance, RangeLow  uint64
	RVOL12, RVOL20, RVOL25                bool
	ArmedIndex                            int
	Plan                                  sizingPlan
	Quantity                              uint64
	SpreadPPM, DriftPPM, AgeMS            uint64
}

// contiguousPrefix 는 개장 분부터 빈 분 없이 이어진 앞부분과, 끊겼다면 처음 빠진 분을 돌려줌(design §2.1-3).
// 지표 생산은 빈 분을 엄격히 거절하므로 도구는 이 앞부분만 평가하고, 다음 폴이 빈 분을 채우면 앞부분이 다시 길어짐.
func contiguousPrefix(series strategyevidence.BarSeries, open time.Time) (strategyevidence.BarSeries, time.Time, bool) {
	expected := uint64(open.UnixMilli())
	for i, record := range series.Bars {
		if record.Payload.OpenAtMS != expected {
			prefix := series
			prefix.Bars = series.Bars[:i]
			return prefix, time.UnixMilli(int64(expected)).UTC(), true
		}
		expected += 60_000
	}
	return series, time.Time{}, false
}

// evaluateSetup 은 묶음 하나를 평가함. ARMED 면 되찾은 봉을 찾아 그 종가로 가상 사이징을 세우고 같은 호가로 최종 판정까지 함.
// 호가가 묵었으면 결과는 QUOTE_STALE — 그 결과를 「시도」로 셀지는 호출자가 정함(새 호가일 때만).
func evaluateSetup(series strategyevidence.BarSeries, sess sessionInfo, symbol string, p shadowParams, q quoteSnap, nowMS uint64) (evaluation, error) {
	metrics, err := breakoutsignal.Derive(series, sess.Open)
	if err != nil {
		return evaluation{}, err
	}
	lane, err := metrics.LaneBars()
	if err != nil {
		return evaluation{}, err
	}
	resistance := uint64(0)
	for _, b := range metrics.Bars[:breakoutsignal.OpeningRangeBars] {
		resistance = max(resistance, b.Record.Payload.HighMinor)
	}
	// 틱은 opening range 저항의 호가 단위로 세션 동안 고정(§2.1-2) — 장중에 바꾸면 config digest → setupID 가 갈림.
	tick := krTickMinor(resistance)
	cfgIn := breakoutlane.V1ConfigInput{Version: breakoutlane.LaneVersionV1, TickMinor: tick, OpeningRangeMinutes: policyOpeningRangeMinutes,
		BreakoutBufferPPM: policyBreakoutBufferPPM, RetestTolerancePPM: p.RetestTolerancePPM, TimeoutKR: policyTimeoutKR, TimeoutUS: policyTimeoutUS,
		RVOLMinPPM: policyRVOLMinPPM, UpperWickRangeMaxPPM: policyUpperWickRangeMaxPPM,
		MaxQuoteAgeMS: p.MaxQuoteAgeMS, MaxSpreadPPM: p.MaxSpreadPPM, MaxEntryDriftPPM: p.MaxEntryDriftPPM}
	cfgIn.Digest = breakoutlane.V1ConfigDigest(cfgIn)
	cfg, err := breakoutlane.NewV1Config(cfgIn)
	if err != nil {
		return evaluation{}, err
	}
	quoteIn := breakoutlane.QuoteSealInput{BidMinor: q.Bid, AskMinor: q.Ask, LastMinor: q.Last,
		SourceObservedAtMS: q.ObservedMS, ReceivedAtMS: q.ReceivedMS, Currency: q.Currency}
	quoteIn.Digest = breakoutlane.QuoteSealDigest(quoteIn)
	quote, err := breakoutlane.NewQuoteSeal(quoteIn)
	if err != nil {
		return evaluation{}, err
	}
	// KRW 계좌 · KRW 종목이라 1/1, 신선 창 = 세션 개장~마감(§2.1-2).
	fxIn := breakoutlane.FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "KRW", Direction: breakoutlane.FXAccountToInstrument,
		RateNum: 1, RateDen: 1, Scale: 1, AsOfMS: uint64(sess.Open.UnixMilli()), FreshUntilMS: uint64(sess.Close.UnixMilli())}
	fxIn.Digest = breakoutlane.FXSealDigest(fxIn)
	fx, err := breakoutlane.NewFXSeal(fxIn)
	if err != nil {
		return evaluation{}, err
	}
	input := func(bars []breakoutlane.ClosedBar, sizing breakoutlane.SizingInput) breakoutlane.EvidenceInput {
		return breakoutlane.EvidenceInput{Market: breakoutlane.MarketKR, Symbol: symbol, SessionID: sess.ID,
			CalendarVersion: sess.CalendarVersion, LaneID: breakoutlane.KRLaneID, LaneVersion: breakoutlane.LaneVersionV1,
			Config: cfg, Bars: bars, ATRMinor: metrics.ATRMinor, EvaluatedAtMS: nowMS, Quote: quote, FX: fx, Sizing: sizing}
	}
	decide := func(bars []breakoutlane.ClosedBar, sizing breakoutlane.SizingInput) (breakoutlane.Decision, error) {
		snapshot, err := breakoutlane.NewEvidenceSnapshot(input(bars, sizing))
		if err != nil {
			return breakoutlane.Decision{}, err
		}
		return breakoutlane.Evaluate(snapshot, nil), nil
	}

	decision, err := decide(lane, breakoutlane.SizingInput{})
	if err != nil {
		return evaluation{}, err
	}
	out := evaluation{Bars: len(lane), Tick: tick, ATRMinor: metrics.ATRMinor, MetricsDigest: metrics.Digest,
		ArmedIndex: -1, ConfigDigest: cfg.Digest()}
	if decision.Phase() == "ARMED" {
		// 되찾은 봉 = 앞부분 길이를 늘려 가며 처음 ARMED 가 나오는 길이의 마지막 봉(§2.1-5). 평가기는 순수 함수라 GET 0.
		for n := breakoutsignal.OpeningRangeBars + 1; n <= len(lane); n++ {
			prefix, err := decide(lane[:n], breakoutlane.SizingInput{})
			if err != nil {
				return evaluation{}, err
			}
			if prefix.Phase() == "ARMED" {
				out.ArmedIndex = n - 1
				break
			}
		}
		if out.ArmedIndex < 0 {
			return evaluation{}, errors.New("the evaluator is ARMED on the series but on none of its prefixes")
		}
		entry := metrics.Bars[out.ArmedIndex].Record.Payload.CloseMinor
		stop := decision.Provenance().RangeLowMinor
		plan, err := planSizing(entry, q.Ask, stop, p.CostPPM, p.MinRRPPM)
		if err == nil {
			out.Plan = plan
			decision, err = decide(lane, breakoutlane.SizingInput{ProposedEntryMinor: entry, StopMinor: stop, TargetMinor: plan.Target,
				RoundTripCostAccountMinor: plan.RoundTrip, RiskBudgetAccountMinor: p.RiskBudgetKRW,
				NotionalCapAccountMinor: p.NotionalCapKRW, FinalCap: policyFinalCap, MinRiskRewardPPM: p.MinRRPPM})
			if err != nil {
				return evaluation{}, err
			}
		}
		// 원장에 남길 실측값 — 평가기 validateQuote 와 같은 식(D21 한계를 바꿔 재계산할 수 있게).
		out.SpreadPPM, out.DriftPPM, out.AgeMS = quoteMeasurements(q, entry, nowMS)
	}
	prov := decision.Provenance()
	out.Phase, out.Refusal = decision.Phase(), string(decision.Refusal())
	out.SetupID, out.SnapshotDigest = decision.SetupID(), decision.SnapshotDigest()
	out.Transitions = append([]string(nil), prov.Transitions...)
	out.Resistance, out.RangeLow = prov.ResistanceMinor, prov.RangeLowMinor
	out.RVOL12, out.RVOL20, out.RVOL25 = prov.RVOLAt1200000, prov.RVOLAt2000000, prov.RVOLAt2500000
	out.Quantity = decision.FinalQuantity()
	return out, nil
}

// quoteMeasurements 는 스프레드 · 진입 괴리(ppm, 올림)와 호가 나이(ms) — 평가기의 판정식 그대로.
func quoteMeasurements(q quoteSnap, entry, nowMS uint64) (spread, drift, age uint64) {
	mid := q.Bid/2 + q.Ask/2 + (q.Bid%2+q.Ask%2)/2
	if q.Ask >= q.Bid && mid > 0 {
		spread, _ = mulDivCeil(q.Ask-q.Bid, ppmScale, mid)
	}
	if entry > 0 {
		diff := q.Ask - entry
		if entry > q.Ask {
			diff = entry - q.Ask
		}
		drift, _ = mulDivCeil(diff, ppmScale, entry)
	}
	if nowMS >= q.ObservedMS {
		age = nowMS - q.ObservedMS
	}
	return spread, drift, age
}
