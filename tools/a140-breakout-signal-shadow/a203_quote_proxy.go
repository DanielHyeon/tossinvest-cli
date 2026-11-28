package main

// a203_quote_proxy.go 는 과거 호가가 없는 재생의 대리 호가와 ARMED 시도(design §3-5 · §3-6 · §4).
//
// 주 가정 P0: 시도 i 의 평가 시각 = 봉 k+1+i 의 여는 시각, ask = 그 봉 시가, bid = ask − 1 틱, last = ask, 관측 = 수신 = 평가 시각.
// 민감도(판정 아님): S1 ask = 진입가, S2 ask = 봉 k+1+i 의 고가. 호가 나이 · 스프레드 관문은 구성상 통과하고(C4) 괴리만 실제로 판정됨.

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

const (
	a203AssumptionP0 = "P0"
	a203AssumptionS1 = "S1"
	a203AssumptionS2 = "S2"
	a203Attempts     = 5 // a140 ArmedAttempts 와 같은 수(분 단위 창 5 분)
	minuteMS         = uint64(60_000)
)

var a203Assumptions = []string{a203AssumptionP0, a203AssumptionS1, a203AssumptionS2}

// a203ProxyQuote 는 ask 하나로 만든 봉인 가능한 호가(C3 — bid · ask · last 모두 양수, bid ≤ ask, KRW).
func a203ProxyQuote(ask, atMS uint64, assumption string) quoteSnap {
	return quoteSnap{Bid: ask - krTickMinor(ask), Ask: ask, Last: ask, ObservedMS: atMS, ReceivedMS: atMS, Currency: "KRW",
		Digest: "a203-proxy:" + assumption}
}

// a203ProxyAsk 는 가정별 ask.
func a203ProxyAsk(assumption string, bar strategyevidence.ClosedBar1mPayload, entry uint64) uint64 {
	switch assumption {
	case a203AssumptionS1:
		return entry
	case a203AssumptionS2:
		return bar.HighMinor
	default:
		return bar.OpenMinor
	}
}

// a203Attempt 는 시도 한 번의 기록(원장 signal 행 재료).
type a203Attempt struct {
	N     int
	AtMS  uint64
	Quote quoteSnap
	Eval  evaluation
	Err   string
}

// a203Trade 는 PROPOSED 한 건과 그 라벨.
type a203Trade struct {
	Session, Symbol, Share, Assumption string
	Attempt                            int
	Plan                               sizingPlan
	Quantity                           uint64
	SetupID                            string
	LabelFromMS                        uint64
	Labeled                            bool
	Label                              labelResult
	Return                             float64 // 건당 순 수익률 r = (exit − worst − 왕복) / worst(design §7)
}

// a203ArmedAttempts 는 ARMED 가 처음 난 봉 k 뒤 시도 1 … 5(design §3-5). 앞부분은 언제나 0..k 로 고정해 「평가기는 ARMED 에서
// 멈춘다」는 미검증 인용에 기대지 않음. 그 분의 봉이 없으면(빈 분 · 세션 끝) 그 시도는 건너뜀 — 창은 분으로 잼(a140 과 같음).
// 처음 PROPOSED 면 라벨(봉 k+2+i 의 여는 시각부터 — C6)하고 끝, 다섯 번 모두 아니면 nil(armed_expired).
func a203ArmedAttempts(prefix, full strategyevidence.BarSeries, k int, sess sessionInfo, symbol string, params shadowParams,
	assumption string) (*a203Trade, []a203Attempt) {
	armed := prefix.Bars[k].Payload
	head := prefix
	head.Bars = prefix.Bars[:k+1]
	byOpen := make(map[uint64]strategyevidence.ClosedBar1mPayload, len(full.Bars))
	for _, record := range full.Bars {
		byOpen[record.Payload.OpenAtMS] = record.Payload
	}
	var log []a203Attempt
	for i := 1; i <= a203Attempts; i++ {
		bar, ok := byOpen[armed.OpenAtMS+uint64(1+i)*minuteMS]
		if !ok {
			continue
		}
		at := bar.OpenAtMS
		quote := a203ProxyQuote(a203ProxyAsk(assumption, bar, armed.CloseMinor), at, assumption)
		ev, err := evaluateSetup(head, sess, symbol, params, quote, at)
		attempt := a203Attempt{N: i, AtMS: at, Quote: quote, Eval: ev}
		if err != nil {
			attempt.Err = err.Error()
			log = append(log, attempt)
			continue
		}
		log = append(log, attempt)
		if ev.Phase != "PROPOSED" {
			continue
		}
		trade := &a203Trade{Session: sess.ID, Symbol: symbol, Assumption: assumption, Attempt: i, Plan: ev.Plan, Quantity: ev.Quantity,
			SetupID: ev.SetupID, LabelFromMS: at + minuteMS}
		trade.Label, trade.Labeled = labelProposal(full, trade.LabelFromMS, ev.Plan, true)
		if trade.Labeled {
			trade.Return = a203NetReturn(trade.Label.ExitMinor, ev.Plan)
		}
		return trade, log
	}
	return nil, log
}

// a203NetReturn 은 a140 보고서 「가상 손익」과 같은 식(a140 archive report_supplement.py:76-79).
func a203NetReturn(exit uint64, plan sizingPlan) float64 {
	if plan.Worst == 0 {
		return 0
	}
	return (float64(exit) - float64(plan.Worst) - float64(plan.RoundTrip)) / float64(plan.Worst)
}
