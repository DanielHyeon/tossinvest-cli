package main

// features_test.go 는 1.7 state 확장(공개 시장 데이터만)의 특징 계산과 조립을 고정함.
// 기대값은 손으로(별도 python 계산으로) 정한 픽스처 — 계산 코드에서 읽어 오지 않음.

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

func fp(v float64) *float64 { return &v }

func sameFloat(t *testing.T, name string, got *float64, want *float64) {
	t.Helper()
	switch {
	case got == nil && want == nil:
	case got == nil || want == nil:
		t.Errorf("%s = %v, want %v", name, deref(got), deref(want))
	case *got != *want:
		t.Errorf("%s = %v, want %v", name, *got, *want)
	}
}

func deref(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

// minuteFixture 는 09:01..10:00 KST 에 닫힌 60 봉: 종가 100+i, 고가 +1, 저가 -1, 거래량 10(마지막 5 봉 20).
// 개장 전(08:59)·판단 시각 뒤(10:01) 봉을 하나씩 섞어 세션·시각 거름을 봄.
func minuteFixture() []minuteBar {
	bars := []minuteBar{{CloseAt: at(8, 59), Open: 1, High: 999, Low: 1, Last: 1, Volume: 999}}
	for i := 1; i <= 60; i++ {
		volume := 10.0
		if i > 55 {
			volume = 20
		}
		price := 100 + float64(i)
		bars = append(bars, minuteBar{CloseAt: at(9, 0).Add(time.Duration(i) * time.Minute),
			Open: price, High: price + 1, Low: price - 1, Last: price, Volume: volume})
	}
	return append(bars, minuteBar{CloseAt: at(10, 1), Open: 1, High: 999, Low: 1, Last: 1, Volume: 999})
}

func TestIntradayFeaturesOnAHandComputedSession(t *testing.T) {
	t.Parallel()
	got := intradayFeatures(minuteFixture(), at(9, 0), at(10, 0), 160)
	if got.BarsAvailable != 60 || len(got.Bars) != 30 || got.Bars[0].T != "09:31" || got.Bars[29].T != "10:00" ||
		got.Bars[29].C != 160 || got.Bars[29].V != 20 {
		t.Fatalf("bars: available %d, recent %d, first %+v, last %+v", got.BarsAvailable, len(got.Bars), got.Bars[0], got.Bars[len(got.Bars)-1])
	}
	sameFloat(t, "return_5m", got.Return5mPct, fp(3.2258))
	sameFloat(t, "return_15m", got.Return15mPct, fp(10.3448))
	sameFloat(t, "return_30m", got.Return30mPct, fp(23.0769))
	sameFloat(t, "return_60m (no bar at or before the open)", got.Return60mPct, nil)
	sameFloat(t, "window_high", got.WindowHigh, fp(161))
	sameFloat(t, "window_low", got.WindowLow, fp(100))
	sameFloat(t, "price_in_window_range_pct", got.PriceInWindowRangePct, fp(98.36))
	sameFloat(t, "window_vwap_typical", got.WindowVWAPTypical, fp(132.6154))
	sameFloat(t, "vs_window_vwap_pct", got.VsWindowVWAPPct, fp(20.6497))
	sameFloat(t, "volume_last_5", got.VolumeLast5BarsVsEarlierAvg, fp(2))
	sameFloat(t, "volume_last_30", got.VolumeLast30BarsVsEarlierAvg, fp(1.167))
}

func TestIntradayFeaturesBoundaries(t *testing.T) {
	t.Parallel()
	t.Run("empty", func(t *testing.T) {
		got := intradayFeatures(nil, at(9, 0), at(10, 0), 160)
		if got.BarsAvailable != 0 || got.Bars == nil || len(got.Bars) != 0 || got.Return5mPct != nil || got.WindowHigh != nil ||
			got.WindowVWAPTypical != nil || got.VolumeLast5BarsVsEarlierAvg != nil {
			t.Fatalf("empty session must carry bars_available 0, an empty bar list and no features: %+v", got)
		}
	})
	t.Run("single bar", func(t *testing.T) {
		got := intradayFeatures([]minuteBar{{CloseAt: at(9, 55), Open: 155, High: 156, Low: 154, Last: 155, Volume: 7}}, at(9, 0), at(10, 0), 160)
		sameFloat(t, "return_5m", got.Return5mPct, fp(3.2258))
		sameFloat(t, "return_15m", got.Return15mPct, nil)
		sameFloat(t, "range", got.PriceInWindowRangePct, fp(300))
		sameFloat(t, "vwap", got.WindowVWAPTypical, fp(155))
		sameFloat(t, "volume_last_5 (no earlier bars)", got.VolumeLast5BarsVsEarlierAvg, nil)
	})
	t.Run("zero volume", func(t *testing.T) {
		var bars []minuteBar
		for i := 1; i <= 40; i++ {
			bars = append(bars, minuteBar{CloseAt: at(9, 0).Add(time.Duration(i) * time.Minute), Open: 100, High: 100, Low: 100, Last: 100})
		}
		got := intradayFeatures(bars, at(9, 0), at(10, 0), 100)
		if got.WindowVWAPTypical != nil || got.VolumeLast5BarsVsEarlierAvg != nil || got.VolumeLast30BarsVsEarlierAvg != nil ||
			got.PriceInWindowRangePct != nil {
			t.Fatalf("zero volume / flat range must not divide by zero: %+v", got)
		}
	})
	t.Run("lookback tolerance 5 minutes", func(t *testing.T) {
		in := intradayFeatures([]minuteBar{{CloseAt: at(9, 40), Last: 150, High: 150, Low: 150, Open: 150}}, at(9, 0), at(10, 0), 165)
		sameFloat(t, "return_15m with a bar exactly 5m before the target", in.Return15mPct, fp(10))
		out := intradayFeatures([]minuteBar{{CloseAt: at(9, 39), Last: 150, High: 150, Low: 150, Open: 150}}, at(9, 0), at(10, 0), 165)
		sameFloat(t, "return_15m with a bar 6m before the target", out.Return15mPct, nil)
	})
}

// dailyFixture 는 직전 21 거래일(종가 101..121, 고저 ±2, 거래량 1000) + 오늘 봉(09:00 KST, 거래량 500).
func dailyFixture() []domain.Candle {
	var candles []domain.Candle
	for k := 1; k <= 21; k++ {
		day := time.Date(2026, 10, 7-(22-k), 9, 0, 0, 0, kst)
		close := 100 + float64(k)
		candles = append(candles, domain.Candle{Time: day, Open: close, High: close + 2, Low: close - 2, Close: close, Volume: 1000})
	}
	candles = append(candles, domain.Candle{Time: at(9, 0), Close: 129, High: 131, Low: 120, Volume: 500}, domain.Candle{}) // 시각 없는 봉은 버림
	// 응답 순서(최신이 앞)를 흉내내려고 뒤집음.
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}
	return candles
}

func TestDailyFeaturesOnHandComputedBars(t *testing.T) {
	t.Parallel()
	got := dailyFeatures(dailyFixture(), at(9, 0), 130)
	if got.PriorBars != 21 || got.LastPriorBarAt != "2026-10-06T00:00:00Z" {
		t.Fatalf("prior bars %d, last prior %q", got.PriorBars, got.LastPriorBarAt)
	}
	sameFloat(t, "return_1d", got.Return1dPct, fp(7.438))
	sameFloat(t, "return_5d", got.Return5dPct, fp(11.1111))
	sameFloat(t, "return_20d", got.Return20dPct, fp(27.451))
	sameFloat(t, "high_20d", got.High20d, fp(123))
	sameFloat(t, "low_20d", got.Low20d, fp(100))
	sameFloat(t, "vs_high_20d", got.VsHigh20dPct, fp(5.6911))
	sameFloat(t, "vs_low_20d", got.VsLow20dPct, fp(30))
	sameFloat(t, "avg_volume_20d (today's bar left out)", got.AvgVolume20d, fp(1000))

	short := dailyFeatures(dailyFixture()[:8], at(9, 0), 130) // 시각 없는 봉 + 오늘 + 직전 6 일
	sameFloat(t, "return_5d with 6 prior bars", short.Return5dPct, fp(11.1111))
	sameFloat(t, "return_20d with 6 prior bars", short.Return20dPct, nil)
	sameFloat(t, "high_20d with 6 prior bars", short.High20d, nil)
	empty := dailyFeatures(nil, at(9, 0), 130)
	if empty.PriorBars != 0 || empty.Return1dPct != nil || empty.High20d != nil {
		t.Fatalf("no bars must carry no features: %+v", empty)
	}
}

func TestLimitsAndIndexFeatures(t *testing.T) {
	t.Parallel()
	limits := priceLimitFeatures(domain.PriceLimits{UpperLimit: 130, LowerLimit: 70}, 100)
	if limits == nil || limits.UpperLimit != 130 || limits.LowerLimit != 70 {
		t.Fatalf("limits = %+v", limits)
	}
	sameFloat(t, "to upper", limits.ToUpperPct, fp(30))
	sameFloat(t, "to lower", limits.ToLowerPct, fp(-30))
	if priceLimitFeatures(domain.PriceLimits{}, 100) != nil {
		t.Fatal("null limits (US shape) must not become a band of zeros")
	}

	index := indexFeatures("KOSPI", 2600,
		[]domain.MarketIndicatorCandle{{Timestamp: "2026-10-07T09:00:00+09:00", Close: 2590}, {Timestamp: "2026-10-06T09:00:00+09:00", Close: 2500}},
		[]domain.MarketIndicatorCandle{{Timestamp: "2026-10-07T09:31:00+09:00", Close: 2595}, {Timestamp: "2026-10-07T09:30:00+09:00", Close: 2590}},
		at(9, 0), at(10, 0))
	if index.Symbol != "KOSPI" || index.Last != 2600 {
		t.Fatalf("index = %+v", index)
	}
	sameFloat(t, "index change today", index.ChangeTodayPct, fp(4))
	sameFloat(t, "index change 30m", index.Change30mPct, fp(0.3861))
	bare := indexFeatures("KOSDAQ", 800, nil, nil, at(9, 0), at(10, 0))
	if bare.ChangeTodayPct != nil || bare.Change30mPct != nil {
		t.Fatalf("an index without bars must carry no changes: %+v", bare)
	}
}

// TestAssembleStateCarriesEachNewInputOrNamesItsAbsence 는 새 입력 하나하나가 있으면 state 에 실리고,
// 없으면(nil) missing_inputs 에 이름이 남는 정책(stocks_meta 와 같음)을 봄 — 판단은 계속됨.
func TestAssembleStateCarriesEachNewInputOrNamesItsAbsence(t *testing.T) {
	t.Parallel()
	full := marketInputs{
		Market: "KR", ObservedAt: at(10, 0), Horizon: time.Hour, Session: session{Open: at(9, 0), Close: at(15, 30)},
		Rank: domain.RankingItem{Symbol: "005930", Currency: "KRW"}, LastPrice: 160,
		MinuteBars: minuteFixture(), DailyBars: dailyFixture(),
		Warnings: &domain.StockWarnings{Warnings: []domain.StockWarning{{Type: "VI_STATIC"}}},
		Limits:   &domain.PriceLimits{UpperLimit: 200, LowerLimit: 100},
		Context:  &MarketContextState{Indices: []IndexState{{Symbol: "KOSPI", Last: 1}}},
	}
	state := assembleState(full)
	if state.Intraday == nil || state.Daily == nil || state.MarketContext == nil ||
		state.PriceLimits == nil || state.WarningTypes == nil || !reflect.DeepEqual(*state.WarningTypes, []string{"VI_STATIC"}) {
		t.Fatalf("a full input must fill every new block: %+v", state)
	}
	for _, name := range []string{"minute_bars", "daily_bars", "warnings", "price_limits"} {
		if hasPrefix(state.MissingInputs, name+":") {
			t.Fatalf("%s present but named missing: %v", name, state.MissingInputs)
		}
	}

	// 경고 0 건은 빈 목록으로 명시됨(「없음」과 「못 읽음」을 가름).
	noWarnings := full
	noWarnings.Warnings = &domain.StockWarnings{}
	raw, _ := json.Marshal(assembleState(noWarnings))
	if !strings.Contains(string(raw), `"warning_types":[]`) {
		t.Fatalf("zero warnings must be an explicit empty list: %s", raw)
	}

	bare := assembleState(marketInputs{
		Market: "KR", ObservedAt: at(10, 0), Session: session{Open: at(9, 0), Close: at(15, 30)},
		Rank: domain.RankingItem{Symbol: "005930"}, LastPrice: 160,
		Gaps: []string{"minute_bars: x", "daily_bars: x", "warnings: x", "price_limits: x", "market_context: x"},
	})
	if bare.Intraday != nil || bare.Daily != nil || bare.PriceLimits != nil ||
		bare.WarningTypes != nil || bare.MarketContext != nil {
		t.Fatalf("missing inputs must leave their blocks out: %+v", bare)
	}
	for _, name := range []string{"minute_bars", "daily_bars", "warnings", "price_limits", "market_context"} {
		if !hasPrefix(bare.MissingInputs, name+":") {
			t.Fatalf("%s absence is not named: %v", name, bare.MissingInputs)
		}
	}
	// KR 종목별 투자자 매매(s3 = WTS)·뉴스를 수집 단계가 이름 없이 비워 넘겨도 조립기가 이름을 남김.
	if !contains(bare.MissingInputs, "investor_flows_per_stock: not read") || !contains(bare.MissingInputs, "news: not read") {
		t.Fatalf("KR states must name an unread per-stock investor flow and news: %v", bare.MissingInputs)
	}
}

// TestCycleReadsTheRicherStateWithinItsCallBudget 는 사이클의 브로커 GET 을 경로·rate-limit 그룹별로 셈(KR 종목 2):
// /orderbook 은 종목당 한 번(엄격 맨 위 한 줄만), 지수 시세는 사이클당 한 번 묶음, 일봉·가격제한폭·지수 일봉·
// 시장 투자자 매매는 하루 한 번(둘째 사이클은 캐시 적중), 체결·호가 깊이는 부르지 않음.
func TestCycleReadsTheRicherStateWithinItsCallBudget(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	want := map[string]int{
		"/api/v1/rankings": 1, "/api/v1/stocks": 1, "/api/v1/prices": 1,
		"/api/v1/orderbook": 2, "/api/v1/candles": 4, "/api/v1/trades": 0, "/api/v1/price-limits": 2,
		"/api/v1/stocks/005930/warnings": 1, "/api/v1/stocks/000660/warnings": 1,
		"/api/v1/market-indicators/prices": 1, "/api/v1/market-indicators/KOSPI/candles": 2,
		"/api/v1/market-indicators/KOSDAQ/candles": 2, "/api/v1/market-indicators/KOSPI/investor-trading": 1,
		"/api/v1/market-indicators/KOSDAQ/investor-trading": 1,
	}
	for path, n := range want {
		if got := rig.official.callCount(path); got != n {
			t.Errorf("%s called %d times, want %d", path, got, n)
		}
	}
	contents := rig.contents(t)
	first := onlySelection(t, contents)
	wantGroups := readGroups{Ranking: 1, Stock: 3, MarketData: 5, MarketDataChart: 4, MarketIndicator: 3, MarketIndicatorChart: 4}
	if first.ReadsByGroup != wantGroups || first.BrokerReads != 20 || first.CacheHits != 0 || first.IntervalSeconds != 600 {
		t.Fatalf("first cycle reads = %d %+v, cache hits %d, interval %d", first.BrokerReads, first.ReadsByGroup, first.CacheHits, first.IntervalSeconds)
	}
	for _, row := range contents.Judgments {
		if row.StateRev != stateRev {
			t.Fatalf("judgment row state_rev = %q, want %q", row.StateRev, stateRev)
		}
	}
	var request struct {
		State PublicState `json:"state"`
	}
	_ = json.Unmarshal(rig.typeSafe.requests()[0], &request)
	got := request.State
	if got.Intraday == nil || got.Intraday.BarsAvailable != 60 || len(got.Intraday.Bars) != 30 || got.Daily == nil ||
		got.Daily.PriorBars != 22 || got.PriceLimits == nil || got.WarningTypes == nil ||
		!reflect.DeepEqual(*got.WarningTypes, []string{"VI_STATIC"}) ||
		got.MarketContext == nil || len(got.MarketContext.Indices) != 2 || len(got.MarketContext.InvestorFlows) != 2 {
		t.Fatalf("outbound state lacks the richer blocks: %+v", got)
	}
	flow := got.MarketContext.InvestorFlows[0]
	if flow.Market != "KOSPI" || flow.IndividualNetKRW != 200 || flow.ForeignerNetKRW != -150 || flow.InstitutionNetKRW != 50 {
		t.Fatalf("investor flow = %+v", flow)
	}

	// 둘째 사이클(같은 날): 하루 캐시 8 적중(종목 일봉 2·가격제한폭 2·지수 일봉 2·투자자 2), GET 12.
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 10)); err != nil {
		t.Fatal(err)
	}
	second := *rig.contents(t).Judgments[len(rig.contents(t).Judgments)-1].Selection
	wantGroups = readGroups{Ranking: 1, Stock: 3, MarketData: 3, MarketDataChart: 2, MarketIndicator: 1, MarketIndicatorChart: 2}
	if second.ReadsByGroup != wantGroups || second.BrokerReads != 12 || second.CacheHits != 8 {
		t.Fatalf("second cycle reads = %d %+v, cache hits %d", second.BrokerReads, second.ReadsByGroup, second.CacheHits)
	}
	if n := rig.official.callCount("/api/v1/price-limits"); n != 2 {
		t.Fatalf("price limits read %d times over two cycles of one day, want 2", n)
	}

	// US: 지수·투자자 맥락·가격제한폭은 부르지 않고(공식 표면에 없음) 결손으로 이름을 남김.
	us := newTestRig(t)
	us.official.setRanking(rankingRow{"AAPL", "100", "USD"})
	if err := us.probe.cycle(context.Background(), "US", us.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/price-limits", "/api/v1/market-indicators/prices", "/api/v1/market-indicators/KOSPI/candles"} {
		if n := us.official.callCount(path); n != 0 {
			t.Errorf("US cycle called %s %d times", path, n)
		}
	}
	var usRequest struct {
		State PublicState `json:"state"`
	}
	_ = json.Unmarshal(us.typeSafe.requests()[0], &usRequest)
	if usRequest.State.MarketContext != nil || !hasPrefix(usRequest.State.MissingInputs, "market_context:") ||
		usRequest.State.PriceLimits != nil || usRequest.State.Intraday == nil {
		t.Fatalf("US state must name the missing market context and carry intraday: %+v", usRequest.State)
	}
}

// TestInactiveListingSkipsPerSymbolReads 는 /stocks 상장 상태가 ACTIVE 가 아니면 종목별 읽기를 건너뛰고
// 이름으로 남기되 판단은 함(결정은 Jev 와 J2 가 함 — 프로브는 비용만 아낌).
func TestInactiveListingSkipsPerSymbolReads(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.stockStatus = map[string]string{"000660": "DELISTED"}
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if n := rig.official.callCount("/api/v1/stocks/000660/warnings"); n != 0 {
		t.Fatalf("a delisted symbol's warnings were read %d times", n)
	}
	if n := rig.official.callCount("/api/v1/candles"); n != 2 {
		t.Fatalf("candles read %d times, want 2 (only the active symbol)", n)
	}
	for _, row := range rig.contents(t).Judgments {
		skipped := hasPrefix(row.CollectionGaps, "per_symbol_reads: skipped, /stocks listing status is DELISTED")
		if skipped != (row.Symbol == "000660") {
			t.Fatalf("%s skip named = %v: %v", row.Symbol, skipped, row.CollectionGaps)
		}
	}
}

// TestReadFailuresBecomeMissingInputsNotGaps 는 새 읽기가 실패해도 판단은 계속되고 실패가 이름으로 남음을 봄.
func TestReadFailuresBecomeMissingInputsNotGaps(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	for _, path := range []string{"/api/v1/candles", "/api/v1/stocks/005930/warnings", "/api/v1/stocks/000660/warnings",
		"/api/v1/price-limits", "/api/v1/market-indicators/prices"} {
		rig.official.failStatus[path] = 503
	}
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	if len(contents.Judgments) != 8 || len(contents.Gaps) != 0 { // 2 종목 × 질문 4(s4)
		t.Fatalf("read failures must not stop the judgment: %d judgments, gaps %+v", len(contents.Judgments), contents.Gaps)
	}
	for _, name := range []string{"minute_bars:", "daily_bars:", "warnings:", "price_limits:", "market_context:"} {
		if !hasPrefix(contents.Judgments[0].CollectionGaps, name) {
			t.Fatalf("%s failure is not named on the row: %v", name, contents.Judgments[0].CollectionGaps)
		}
	}
}

// TestReportSplitsLotsByStateRev 는 같은 규칙이라도 state_rev 가 다르면 다른 로트임을 봄.
func TestReportSplitsLotsByStateRev(t *testing.T) {
	t.Parallel()
	old := screenedLedger(1, 0, 0)
	fresh := screenedLedger(1, 0, 0)
	contents := ledgerContents{Labels: map[string]labelRow{}}
	contents.Judgments = append(contents.Judgments, old.Judgments...)
	for id, label := range old.Labels {
		contents.Labels[id] = label
	}
	for _, row := range fresh.Judgments {
		label := fresh.Labels[row.ID]
		row.ID, label.JudgmentID, row.StateRev = "s2-"+row.ID, "s2-"+row.ID, stateRev
		contents.Judgments = append(contents.Judgments, row)
		contents.Labels[row.ID] = label
	}
	report := renderCalibrationReport(contents, "l")
	if !strings.Contains(report, "표본 선택 로트: "+selectionRule+"=2, "+selectionRule+"@"+stateRev+"=2") {
		t.Fatalf("state_rev must split the lot:\n%s", report)
	}
	if !strings.Contains(report, "## 결선 vs 비결선 ("+screenUpQuestion+") — KR · 로트 "+selectionRule+"@"+stateRev) {
		t.Fatalf("the new lot gets its own comparison:\n%s", report)
	}
}

// TestReportShowsCycleBudgetPerLot 는 로트마다 사이클 수·GET 수(중앙·최대)·수집 시간(중앙·최대)과 간격 절반을
// 넘은 사이클 수를 보고서가 싣음을 봄 — 공유 Open API 예산을 사람이 원장에서 확인할 수 있게.
func TestReportShowsCycleBudgetPerLot(t *testing.T) {
	t.Parallel()
	contents := ledgerContents{Labels: map[string]labelRow{}}
	for index, cycle := range []struct {
		reads   int
		seconds float64
	}{{160, 400}, {96, 100}, {98, 120}} {
		record := &selectionRecord{Rule: selectionRule, BrokerReads: cycle.reads, CollectSeconds: cycle.seconds, IntervalSeconds: 600}
		for _, questionID := range []string{screenUpQuestion, screenRedFlagQuestion} {
			contents.Judgments = append(contents.Judgments, judgmentRow{
				ID: fmt.Sprintf("c%d-%s", index, questionID), At: fmt.Sprintf("2026-10-14T0%d:00:00Z", index), Market: "KR",
				QuestionID: questionID, Selection: record, StateRev: stateRev,
			})
		}
	}
	report := renderCalibrationReport(contents, "l")
	want := "- 사이클 예산 " + selectionRule + "@" + stateRev + ": 사이클 3 · broker GET 중앙 98 · 최대 160 · 수집 중앙 120s · 최대 400s · 간격 절반 초과 1"
	if !strings.Contains(report, want) {
		t.Fatalf("report lacks %q:\n%s", want, report)
	}
}

// TestMarketContextGapDoesNotMaskAMissingStockMeta 는 시장 맥락 결손(US 는 언제나)이 「/stocks 응답에 종목 없음」
// 결손 이름을 가리지 않음을 봄 — 둘은 서로 다른 결손.
func TestMarketContextGapDoesNotMaskAMissingStockMeta(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.official.setRanking(rankingRow{"AAPL", "100", "USD"}) // mock /stocks 는 AAPL 을 답하지 않음
	if err := rig.probe.cycle(context.Background(), "US", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	gaps := rig.contents(t).Judgments[0].CollectionGaps
	if !hasPrefix(gaps, "stocks_meta: symbol missing") || !hasPrefix(gaps, "market_context:") {
		t.Fatalf("both gaps must be named: %v", gaps)
	}
}
