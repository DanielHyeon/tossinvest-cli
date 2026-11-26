package main

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

// ---- tasks 2.2: 캐시 리더 → PollClosedBars → SealBarSeries 가 과거 폴 시각으로 한 세션을 적재하는가 ----

func a203TestSession(t *testing.T) a203Session {
	t.Helper()
	src := newFakeSource(nil, func() time.Time { return testOpen })
	payload, err := src.TypedMarketCalendar(context.Background(), "KR", "2026-08-14")
	if err != nil {
		t.Fatal(err)
	}
	return a203Session{Date: "2026-08-14", Role: a203SampleMain, Open: testOpen, Close: testClose, payload: payload}
}

func flatBars(n int) []ohlcv {
	out := make([]ohlcv, n)
	for i := range out {
		out[i] = ohlcv{10000, 10050, 9950, 10000, 100}
	}
	return out
}

func TestA203PastPollLoadsAWholeSession(t *testing.T) {
	t.Parallel()
	// 가짜 브로커는 「지금」 = 2026-11-28(과거 세션 2026-08-14 를 석 달 뒤에 받음), 200 봉 페이지.
	later := time.Date(2026, 11, 28, 10, 0, 0, 0, seoul)
	for _, c := range []struct {
		name      string
		bars      int
		wantBars  int
		wantPages int
	}{
		{"with after-market successors", 390 + 30, 390, 3},
		// 마감 뒤 봉이 없으면 마지막 정규장 봉은 successor 가 없어 적재되지 않음(a140 장중 측정과 같은 생산자 성질).
		{"regular bars only", 390, 389, 2},
	} {
		src := newFakeSource(flatBars(c.bars), func() time.Time { return later })
		src.pageSize = 200
		store := openTestStore(t, func() time.Time { return later })
		s := a203TestSession(t)
		out := a203ReplaySymbol(context.Background(), src, store, s, a203Job{Symbol: "TST001", Universe: a203UniverseUC}, testParams(),
			func() time.Time { return later })
		input, err := a203PollInput(s, "TST001")
		if err != nil {
			t.Fatal(err)
		}
		series, err := store.SealBarSeries(context.Background(), strategyevidence.BarSeriesQuery{Market: "KR", Symbol: "TST001", SessionID: s.id(),
			IntervalMS: strategyevidence.ClosedBar1mIntervalMS, EvaluationAt: input.PollAt, IngestionCutoff: later, RegularSessionOnly: true})
		if err != nil {
			t.Fatal(err)
		}
		if len(series.Bars) != c.wantBars || src.count("StrictMinuteCandles") != c.wantPages {
			t.Fatalf("%s: bars %d pages %d, want %d · %d", c.name, len(series.Bars), src.count("StrictMinuteCandles"), c.wantBars, c.wantPages)
		}
		// 「평가 시각 = 마지막 봉 + 2 분 이상」을 평가기가 받음(evidence snapshot 거절 0) — 평탄한 하루는 RANGE_LOCKED 로 끝남.
		if out.Gap != "" || out.EvalErrors != 0 || out.Terminal != "RANGE_LOCKED" {
			t.Fatalf("%s: gap %q eval errors %d terminal %q", c.name, out.Gap, out.EvalErrors, out.Terminal)
		}
		// 과거 폴 시각이 적재 시각(지금)보다 이른 컷오프를 쓰면 아무것도 안 보임 — 재생 컷오프가 「지금」이어야 하는 이유.
		early, err := store.SealBarSeries(context.Background(), strategyevidence.BarSeriesQuery{Market: "KR", Symbol: "TST001", SessionID: s.id(),
			IntervalMS: strategyevidence.ClosedBar1mIntervalMS, EvaluationAt: input.PollAt, IngestionCutoff: input.PollAt, RegularSessionOnly: true})
		if err != nil || len(early.Bars) != 0 {
			t.Fatalf("%s: a past ingestion cutoff saw %d bars (%v)", c.name, len(early.Bars), err)
		}
	}
}

// ---- tasks 2.3: 대리 호가 P0 · S1 · S2 와 시도 · 라벨 시각 ----

func armedScenario(attemptOpens ...uint64) []ohlcv {
	bars := breakoutScenario()[:18]                             // 봉 17 이 되찾음 → ARMED(k = 17), 진입가 10,070
	bars = append(bars, ohlcv{10070, 10080, 10060, 10075, 100}) // 봉 k+1 — 시도 1 은 그다음 봉(k+2)의 시가
	for _, open := range attemptOpens {
		bars = append(bars, ohlcv{open, open + 10, open - 10, open, 100})
	}
	bars = append(bars, ohlcv{10070, 10400, 10060, 10390, 100}) // 목표(10,378) 를 넘음
	return bars
}

func TestA203AttemptUsesTheMinuteOpenAndLabelsFromTheNextBar(t *testing.T) {
	t.Parallel()
	series := seriesOf(armedScenario(10070))
	trade, log := a203ArmedAttempts(series, series, 17, testSessionInfo(), "TST001", testParams(), a203AssumptionP0)
	if trade == nil || trade.Attempt != 1 {
		t.Fatalf("trade %+v log %+v", trade, log)
	}
	k := series.Bars[17].Payload.OpenAtMS
	if log[0].AtMS != k+2*minuteMS || log[0].Quote.Ask != 10070 || trade.LabelFromMS != k+3*minuteMS {
		t.Fatalf("attempt 1 at %d ask %d label from %d, want bar k+2 open · its open price · bar k+3", log[0].AtMS-k, log[0].Quote.Ask, trade.LabelFromMS-k)
	}
	if !trade.Labeled || trade.Label.Outcome != "target" || trade.Label.ExitOpenMS != k+3*minuteMS {
		t.Fatalf("label %+v", trade.Label)
	}
	// r = (exit − worst − 왕복) / worst — 목표 10,378 · worst 10,070 · 왕복 51.
	if want := float64(10378-10070-51) / 10070; trade.Return != want {
		t.Fatalf("return %v, want %v", trade.Return, want)
	}
}

func TestA203DriftRefusalMovesToTheNextAttempt(t *testing.T) {
	t.Parallel()
	// 봉 19(k+2) 시가 10,200 → 괴리 12,910 ppm > 5,000 → ENTRY_DRIFT_EXCEEDED, 봉 20(k+3) 시가 10,070 → 시도 2 에서 PROPOSED.
	series := seriesOf(armedScenario(10200, 10070))
	trade, log := a203ArmedAttempts(series, series, 17, testSessionInfo(), "TST001", testParams(), a203AssumptionP0)
	if len(log) < 2 || log[0].Eval.Refusal != "ENTRY_DRIFT_EXCEEDED" || trade == nil || trade.Attempt != 2 {
		t.Fatalf("log %+v trade %+v", log, trade)
	}
	if trade.LabelFromMS != series.Bars[17].Payload.OpenAtMS+4*minuteMS {
		t.Fatalf("label start %d, want bar k+2+i with i=2", trade.LabelFromMS)
	}
	// S1(ask = 진입가)은 시도 1 에서 통과, S2(ask = 고가 10,210)는 시도 1 이 거절.
	if s1, _ := a203ArmedAttempts(series, series, 17, testSessionInfo(), "TST001", testParams(), a203AssumptionS1); s1 == nil || s1.Attempt != 1 || s1.Plan.Worst != 10070 {
		t.Fatalf("S1 %+v", s1)
	}
	if s2, _ := a203ArmedAttempts(series, series, 17, testSessionInfo(), "TST001", testParams(), a203AssumptionS2); s2 == nil || s2.Attempt != 2 || s2.Plan.Ask != 10080 {
		t.Fatalf("S2 %+v", s2)
	}
	// 다섯 시도 모두 괴리면 제안 없음(armed_expired).
	far := seriesOf(armedScenario(10200, 10200, 10200, 10200, 10200))
	if none, log := a203ArmedAttempts(far, far, 17, testSessionInfo(), "TST001", testParams(), a203AssumptionP0); none != nil || len(log) != 5 {
		t.Fatalf("expired: %+v %d", none, len(log))
	}
}

func TestA203OneTickProxySpreadStaysInsideTheLimit(t *testing.T) {
	t.Parallel()
	// C4: 가격 1,000 원 이상 · 1 틱이면 스프레드 ≤ 2,504 ppm < MaxSpreadPPM 3,000. 구간 하단(틱이 커지는 가격)이 최악.
	for _, ask := range []uint64{1000, 1999, 2000, 4999, 5000, 20000, 50000, 200000, 500000} {
		q := a203ProxyQuote(ask, 1, a203AssumptionP0)
		spread, drift, age := quoteMeasurements(q, ask, 1)
		if spread > testParams().MaxSpreadPPM || drift != 0 || age != 0 || q.Bid != ask-krTickMinor(ask) || q.Last != ask {
			t.Errorf("ask %d: spread %d drift %d age %d bid %d", ask, spread, drift, age, q.Bid)
		}
	}
	if spread, _, _ := quoteMeasurements(a203ProxyQuote(2000, 1, ""), 2000, 1); spread != 2504 {
		t.Errorf("ask 2,000 spread %d, want the C4 worst case 2,504", spread)
	}
}
