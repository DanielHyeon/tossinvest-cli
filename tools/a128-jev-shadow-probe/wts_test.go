package main

// wts_test.go 는 1.8(s3) WTS 조회 입력을 httptest 서버로만 시험함 — 실 WTS 는 부르지 않음.
// 기존 WTS 클라이언트(internal/client)를 그대로 쓰고 기본 주소만 mock 으로 바꿈(newWTSClient).

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	tossclient "github.com/JungHoonGhae/tossinvest-cli/internal/client"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	wtssession "github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

// 세션 쿠키·WTS 응답의 계좌 모양 표식 — 외부 전송·원장에 새지 않아야 함(시험용 문자열).
const (
	sentinelCookie     = "WTS-COOKIE-SENTINEL-6161"
	sentinelWTSAccount = "WTS-ACCT-SENTINEL-8080"
)

const (
	pathNews    = "/api/v1/dashboard/wts/news"
	pathIndices = "/api/v1/dashboard/wts/overview/indicator/index"
	pathTrend   = "/api/v1/stock-infos/trade/trend/trading-trend"
)

type wtsMock struct {
	server *httptest.Server

	mu         sync.Mutex
	calls      map[string]int
	scopes     []string
	failStatus map[string]int
	cookies    []string
}

func newWTSMock(t *testing.T) *wtsMock {
	m := &wtsMock{calls: map[string]int{}, failStatus: map[string]int{}}
	m.server = httptest.NewServer(http.HandlerFunc(m.serve))
	t.Cleanup(m.server.Close)
	return m
}

// newsBody 는 시험 시각 at(10,0)=2026-10-07 10:00 KST 기준 기사 여섯 건(시각 모양 둘 — 클라이언트 fixture 의 시간대 없는
// "2006-01-02 15:04:05" 와 RFC3339):
// 09:30 삼성 관련(실림) · 10:03 삼성 관련(판단 뒤·받은 순간 허용폭 안 — headlinesFor 가 뺌) · 전날 08:00 삼성 관련(24 시간
// 밖 — 빠짐) · 09:50 kr 시장 기사 · 09:40 us 기사(KR 시장 목록에 안 실림) · 시각 못 읽는 삼성 관련(빠짐) ·
// 11:00 삼성 관련(받은 순간 + 5분 뒤 — 시간대 오류 신호로 보관하지 않음).
func newsBody() string {
	item := func(id, title, created, nation, code string) string {
		related := `[]`
		if code != "" {
			related = `[{"stockCode":"` + code + `","stockName":"삼성전자","market":"KOSPI","fluctuation":1.2}]`
		}
		return `{"newsId":"` + id + `","title":"` + title + `","summary":"요약   ` + title + `\n본문","createdAt":"` + created +
			`","source":"통신사","newsType":"ARTICLE","nation":"` + nation + `","relatedStocks":` + related +
			`,"accountNo":"` + sentinelWTSAccount + `"}`
	}
	return `{"result":{"type":"ALL_HIGHLIGHT","title":"주요 뉴스","news":[` + strings.Join([]string{
		item("n1", "삼성 관련 아침", "2026-10-07 09:30:00", "kr", "A005930"),
		item("n2", "삼성 관련 미래", "2026-10-07 10:03:00", "kr", "A005930"),
		item("n7", "삼성 관련 먼미래", "2026-10-07 11:00:00", "kr", "A005930"),
		item("n3", "삼성 관련 어제", "2026-10-06T08:00:00+09:00", "kr", "A005930"),
		item("n4", "코스피 시장", "2026-10-07 09:50:00", "kr", ""),
		item("n5", "나스닥 시장", "2026-10-07T00:40:00Z", "us", ""),
		item("n6", "삼성 시각없음", "어제쯤", "kr", "A005930"),
	}, ",") + `]},"holdingQuantity":"` + sentinelHolding + `"}`
}

func (m *wtsMock) serve(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.calls[r.URL.Path]++
	if cookie, err := r.Cookie("SESSION"); err == nil {
		m.cookies = append(m.cookies, cookie.Value)
	}
	if r.URL.Path == pathNews {
		body, _ := io.ReadAll(r.Body)
		var request struct {
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &request)
		m.scopes = append(m.scopes, request.Type)
	}
	status := m.failStatus[r.URL.Path]
	if strings.HasPrefix(r.URL.Path, "/api/v2/stock-infos/") && m.failStatus["stock-infos"] != 0 {
		status = m.failStatus["stock-infos"]
	}
	m.mu.Unlock()
	if status != 0 {
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `{"error":"mock failure"}`)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.URL.Path == pathNews:
		_, _ = io.WriteString(w, newsBody())
	case r.URL.Path == pathIndices:
		_, _ = io.WriteString(w, `{"result":{"majorIndicatorInfos":[`+
			`{"code":"COMP.NAI","displayName":"나스닥","nation":"us","price":{"latestPrice":18200,"basePrice":18000}},`+
			`{"code":"VIX","displayName":"VIX","nation":"us","price":{"latestPrice":15,"basePrice":0}},`+
			`{"code":"DEAD","displayName":"없는지수","nation":"us","price":{"latestPrice":0,"basePrice":10}}],`+
			`"accountNo":"`+sentinelWTSAccount+`"}}`)
	case strings.HasPrefix(r.URL.Path, "/api/v2/stock-infos/"):
		_, _ = io.WriteString(w, `{"result":{"symbol":"`+strings.TrimPrefix(r.URL.Path, "/api/v2/stock-infos/A")+
			`","name":"종목","currency":"KRW","status":"ACTIVE"}}`)
	case r.URL.Path == pathTrend:
		// 오늘(2026-10-07) 행은 잠정치라 버려야 함 — 직전 6 거래일 중 최신 5 만 실림.
		rows := []string{}
		for _, day := range []string{"2026-10-07", "2026-10-06", "2026-10-02", "2026-10-01", "2026-09-30", "2026-09-29", "2026-09-26"} {
			rows = append(rows, `{"baseDate":"`+day+`","netIndividualsBuyVolume":100,"netForeignerBuyVolume":-40,`+
				`"netInstitutionBuyVolume":-60,"holdingQuantity":"`+sentinelHolding+`"}`)
		}
		_, _ = io.WriteString(w, `{"result":{"body":[`+strings.Join(rows, ",")+`]}}`)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *wtsMock) count(path string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls[path]
}

func (m *wtsMock) stockInfoCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	total := 0
	for path, n := range m.calls {
		if strings.HasPrefix(path, "/api/v2/stock-infos/") {
			total += n
		}
	}
	return total
}

// withWTS 는 rig 에 WTS mock 을 붙임. withSession=false 면 세션 없는 클라이언트(쿠키 없음).
func (r *testRig) withWTS(t *testing.T, withSession bool) *wtsMock {
	t.Helper()
	mock := newWTSMock(t)
	var sess *wtssession.Session
	if withSession {
		sess = &wtssession.Session{Provider: "test", Cookies: map[string]string{"SESSION": sentinelCookie}}
	}
	r.probe.wts = newWTSClient(sess, mock.server.URL)
	return mock
}

func outboundStates(t *testing.T, r *testRig) []PublicState {
	t.Helper()
	var states []PublicState
	for _, body := range r.typeSafe.requests() {
		var request struct {
			State PublicState `json:"state"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatal(err)
		}
		states = append(states, request.State)
	}
	return states
}

// TestWTSInputsReachTheStateWithinTheirBudget 는 s3 의 세 입력이 state 에 실리고(선행 기사·24 시간 밖·시각 없는 기사·
// 오늘 수급 행 제외), WTS 요청 수가 사이클당 뉴스 범위 수 + 지수 1 + (처음 보는 KR 종목 × 2) 이고 둘째 사이클은 수급이
// 캐시됨을 봄. Open API GET 수는 s2 와 같음(20).
func TestWTSInputsReachTheStateWithinTheirBudget(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, true)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	first := onlySelection(t, rig.contents(t))
	if first.BrokerReads != 20 || first.WTSRequests != len(newsFeedScopes)+1+2*2 {
		t.Fatalf("first cycle: broker GETs %d (want 20, unchanged from s2), WTS requests %d (want %d)",
			first.BrokerReads, first.WTSRequests, len(newsFeedScopes)+5)
	}
	if !reflect.DeepEqual(wts.scopes, newsFeedScopes) {
		t.Fatalf("news scopes requested = %v, want %v", wts.scopes, newsFeedScopes)
	}
	for _, row := range rig.contents(t).Judgments {
		if row.StateRev != "a128-s3" {
			t.Fatalf("state_rev = %q, want a128-s3", row.StateRev)
		}
		for _, prefix := range []string{"news:", "investor_flows_per_stock:"} {
			if hasPrefix(row.CollectionGaps, prefix) {
				t.Fatalf("%s %s still named missing: %v", row.Symbol, prefix, row.CollectionGaps)
			}
		}
		// 남는 시장 맥락 결손은 값 0 인 지수의 이름 하나뿐(빠진 지수를 조용히 지우지 않음).
		var contextGaps []string
		for _, gap := range row.CollectionGaps {
			if strings.HasPrefix(gap, "market_context:") {
				contextGaps = append(contextGaps, gap)
			}
		}
		if !reflect.DeepEqual(contextGaps, []string{"market_context: major_indices: 없는지수 has no positive value"}) {
			t.Fatalf("%s market context gaps = %v", row.Symbol, contextGaps)
		}
	}
	for _, state := range outboundStates(t, rig) {
		if state.News != newsPresent || state.NewsHeadlines == nil {
			t.Fatalf("%s news = %q %+v", state.Symbol, state.News, state.NewsHeadlines)
		}
		var symbolTitles, marketTitles []string
		for _, h := range state.NewsHeadlines.Symbol {
			symbolTitles = append(symbolTitles, h.Title)
		}
		for _, h := range state.NewsHeadlines.Market {
			marketTitles = append(marketTitles, h.Title)
			if h.Summary != "" {
				t.Fatalf("market headlines carry titles only: %+v", h)
			}
		}
		wantSymbol := []string(nil)
		wantMarket := []string{"코스피 시장"}
		if state.Symbol == "005930" {
			wantSymbol = []string{"삼성 관련 아침"}
		} else {
			wantMarket = []string{"코스피 시장", "삼성 관련 아침"} // 다른 종목에는 삼성 기사도 kr 시장 기사로 보임
		}
		if !reflect.DeepEqual(symbolTitles, wantSymbol) || !reflect.DeepEqual(marketTitles, wantMarket) {
			t.Fatalf("%s headlines symbol=%v market=%v, want %v / %v", state.Symbol, symbolTitles, marketTitles, wantSymbol, wantMarket)
		}
		if state.Symbol == "005930" {
			h := state.NewsHeadlines.Symbol[0]
			if h.Summary != "요약 삼성 관련 아침 본문" || h.CreatedAt != "2026-10-07T00:30:00Z" || h.Source != "통신사" {
				t.Fatalf("symbol headline = %+v", h)
			}
		}
		if state.StockFlows == nil || len(*state.StockFlows) != stockFlowDays {
			t.Fatalf("%s stock flows = %+v", state.Symbol, state.StockFlows)
		}
		flows := *state.StockFlows
		if flows[0] != (StockFlowState{Date: "2026-10-06", IndividualNetShares: 100, ForeignerNetShares: -40, InstitutionNetShares: -60}) ||
			flows[4].Date != "2026-09-29" {
			t.Fatalf("stock flows must be prior days newest first: %+v", flows)
		}
		major := state.MarketContext.MajorIndices
		if len(major) != 2 || major[0].Name != "나스닥" || major[0].ChangePctVsBase == nil ||
			*major[0].ChangePctVsBase != 1.1111 || major[1].ChangePctVsBase != nil {
			t.Fatalf("major indices = %+v (zero value dropped, missing base → no change)", major)
		}
		if len(state.MarketContext.Indices) != 2 {
			t.Fatalf("official KR indices must stay alongside: %+v", state.MarketContext)
		}
	}

	// 둘째 사이클(같은 날): 수급은 하루 캐시 → WTS = 뉴스 범위 + 지수 1.
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 10)); err != nil {
		t.Fatal(err)
	}
	judgments := rig.contents(t).Judgments
	second := judgments[len(judgments)-1].Selection
	if second.WTSRequests != len(newsFeedScopes)+1 || wts.count(pathTrend) != 2 || wts.stockInfoCalls() != 2 {
		t.Fatalf("second cycle WTS requests %d, trend reads %d, stock-info reads %d — the per-stock flow must be day-cached",
			second.WTSRequests, wts.count(pathTrend), wts.stockInfoCalls())
	}
}

// TestUSStateCarriesMajorIndicesAndNoStockFlow 는 US 가 WTS 주요 지수로 시장 맥락을 얻고(공식 지수 무호출), 종목 수급은
// 부르지 않음(KRX 전용)을 봄 — 남는 결손은 시장 단위 투자자 매매 하나.
func TestUSStateCarriesMajorIndicesAndNoStockFlow(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, true)
	rig.official.setRanking(rankingRow{"AAPL", "100", "USD"})
	if err := rig.probe.cycle(context.Background(), "US", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if wts.count(pathTrend) != 0 || rig.official.callCount("/api/v1/market-indicators/prices") != 0 {
		t.Fatalf("US read the KRX flow %d / official indices %d times", wts.count(pathTrend), rig.official.callCount("/api/v1/market-indicators/prices"))
	}
	state := outboundStates(t, rig)[0]
	if state.MarketContext == nil || len(state.MarketContext.MajorIndices) != 2 || state.MarketContext.Indices != nil ||
		state.StockFlows != nil || state.News != newsPresent {
		t.Fatalf("US state = %+v", state)
	}
	if len(state.NewsHeadlines.Market) != 1 || state.NewsHeadlines.Market[0].Title != "나스닥 시장" {
		t.Fatalf("US market headlines must be the us-nation items: %+v", state.NewsHeadlines.Market)
	}
	if !contains(state.MissingInputs, "market_context: no US market-wide investor flow on any surface (official investor-trading is KOSPI/KOSDAQ only)") ||
		hasPrefix(state.MissingInputs, "investor_flows_per_stock:") {
		t.Fatalf("US missing inputs = %v", state.MissingInputs)
	}
}

// TestWTSFailuresBecomeNamedGapsNotFabrications 는 WTS 읽기 실패(5xx)가 블록 생략 + 이름이 되고 판단은 계속됨을 봄 —
// 0·빈 목록을 지어내지 않음. 실패한 수급은 캐시되지 않아 다음 사이클에 다시 읽음.
func TestWTSFailuresBecomeNamedGapsNotFabrications(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, true)
	for _, path := range []string{pathNews, pathIndices, pathTrend} {
		wts.failStatus[path] = 503
	}
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	contents := rig.contents(t)
	if len(contents.Judgments) != 4 || len(contents.Gaps) != 0 {
		t.Fatalf("WTS failures must not stop the judgment: %d judgments, gaps %+v", len(contents.Judgments), contents.Gaps)
	}
	for _, state := range outboundStates(t, rig) {
		if state.News != newsAbsent || state.NewsHeadlines != nil || state.StockFlows != nil || state.MarketContext.MajorIndices != nil {
			t.Fatalf("failed WTS reads must leave their blocks out: %+v", state)
		}
		for _, name := range []string{"news: ALL_HIGHLIGHT:", "news: SOARING_STOCK:", "market_context: major_indices:", "investor_flows_per_stock:"} {
			if !hasPrefix(state.MissingInputs, name) {
				t.Fatalf("%s failure is not named: %v", name, state.MissingInputs)
			}
		}
	}
	wts.mu.Lock()
	wts.failStatus = map[string]int{}
	wts.mu.Unlock()
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 10)); err != nil {
		t.Fatal(err)
	}
	last := outboundStates(t, rig)
	if state := last[len(last)-1]; state.StockFlows == nil || state.News != newsPresent {
		t.Fatalf("recovered WTS must fill the blocks on the next cycle: %+v", state)
	}
}

// TestWTSSessionRejectionStopsTheCycleWTSReads 는 세션 거절(401)을 한 번 보면 그 사이클의 나머지 WTS 읽기를 부르지 않고
// (만료 세션에 종목마다 실패 요청 금지) 사유를 이름으로 남기며, 다음 사이클은 다시 시도함을 봄.
func TestWTSSessionRejectionStopsTheCycleWTSReads(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, true)
	wts.failStatus[pathIndices] = 401 // 사이클의 첫 WTS 읽기(시장 맥락) — 뒤의 뉴스·수급은 부르지 않아야 함
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if wts.count(pathIndices) != 1 || wts.count(pathNews) != 0 || wts.count(pathTrend) != 0 || wts.stockInfoCalls() != 0 {
		t.Fatalf("after a 401 the cycle sent news %d / trend %d / stock-info %d more WTS requests",
			wts.count(pathNews), wts.count(pathTrend), wts.stockInfoCalls())
	}
	if got := onlySelection(t, rig.contents(t)).WTSRequests; got != 1 {
		t.Fatalf("WTS requests = %d, want 1", got)
	}
	state := outboundStates(t, rig)[0]
	for _, name := range []string{"news: ALL_HIGHLIGHT: skipped, the WTS session was rejected", "investor_flows_per_stock: skipped, the WTS session was rejected"} {
		if !hasPrefix(state.MissingInputs, name) {
			t.Fatalf("%q not named: %v", name, state.MissingInputs)
		}
	}
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 10)); err != nil {
		t.Fatal(err)
	}
	if wts.count(pathIndices) != 2 {
		t.Fatalf("the next cycle must try again: indices read %d times", wts.count(pathIndices))
	}
}

// TestNoSessionStillReadsWhatNeedsNoSession 는 세션이 없으면 뉴스(클라이언트가 세션을 요구 — 요청 0)만 결손이고,
// 세션 없이 답하는 읽기(mock 은 지수·수급)는 그대로 실림을 봄 — 세션 부재가 사이클 전체의 WTS 를 막지 않음.
func TestNoSessionStillReadsWhatNeedsNoSession(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, false)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if wts.count(pathNews) != 0 {
		t.Fatalf("news was requested %d times without a session", wts.count(pathNews))
	}
	state := outboundStates(t, rig)[0]
	if state.News != newsAbsent || !hasPrefix(state.MissingInputs, "news: ALL_HIGHLIGHT: no active session") ||
		state.StockFlows == nil || state.MarketContext.MajorIndices == nil {
		t.Fatalf("no-session state = news %q flows %v major %v missing %v", state.News, state.StockFlows,
			state.MarketContext.MajorIndices, state.MissingInputs)
	}
}

// TestWTSAccountFieldsAndSessionNeverLeave 는 WTS 응답에 섞은 계좌 모양 표식과 세션 쿠키 값이 TypeSafe 요청·원장
// 디렉터리 어느 파일에도 없음을 봄(쿠키는 WTS 요청에만 실려야 함).
func TestWTSAccountFieldsAndSessionNeverLeave(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, true)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	if len(wts.cookies) == 0 || wts.cookies[0] != sentinelCookie {
		t.Fatalf("the WTS mock never saw the session cookie — the leak check would measure nothing")
	}
	var exposed []string
	for _, body := range rig.typeSafe.requests() {
		exposed = append(exposed, string(body))
	}
	_ = filepath.Walk(rig.dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			raw, _ := os.ReadFile(path)
			exposed = append(exposed, string(raw))
		}
		return nil
	})
	for _, text := range exposed {
		for _, forbidden := range []string{sentinelCookie, sentinelWTSAccount, sentinelHolding, "accountNo", "holdingQuantity"} {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%q leaked into an outbound request or a ledger file", forbidden)
			}
		}
	}
}

func TestPriorDayFlowsAndClip(t *testing.T) {
	t.Parallel()
	flows, err := priorDayFlows([]domain.TradingFlow{{Date: "20261006", NetIndividuals: 1}, {Date: "2026-10-07"}, {Date: "2026-10-02T00:00:00+09:00"}}, "2026-10-07")
	if err != nil || len(flows) != 2 || flows[0].Date != "2026-10-06" || flows[1].Date != "2026-10-02" {
		t.Fatalf("flows = %+v, %v", flows, err)
	}
	if _, err := priorDayFlows([]domain.TradingFlow{{Date: "어제"}}, "2026-10-07"); err == nil {
		t.Fatal("an unparseable date must fail the read, not drop the row silently")
	}
	empty, err := priorDayFlows(nil, "2026-10-07")
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("zero rows must be an explicit empty list: %v %v", empty, err)
	}
	if got := clip("  가나  다라\n마 ", 4); got != "가나 …" {
		t.Fatalf("clip = %q", got)
	}
	if got := clip("짧음", 4); got != "짧음" {
		t.Fatalf("clip = %q", got)
	}
	us := domain.NewsItem{Stocks: []domain.NewsRelatedStock{{Code: "US19801212001", Name: "애플"}}}
	if !relatesTo(us, "AAPL", "애플") || relatesTo(us, "AAPL", "") || relatesTo(us, "MSFT", "마이크로소프트") {
		t.Fatal("US news must match by the official name only")
	}
}

// TestReportShowsWTSRequestsInTheCycleBudget 는 s3 로트의 예산 줄이 WTS 요청 수(중앙·최대)를 함께 싣음을 봄.
func TestReportShowsWTSRequestsInTheCycleBudget(t *testing.T) {
	t.Parallel()
	contents := ledgerContents{Labels: map[string]labelRow{}}
	for index, wtsRequests := range []int{63, 3, 5} {
		record := &selectionRecord{Rule: selectionRule, BrokerReads: 96, CollectSeconds: 60, IntervalSeconds: 600, WTSRequests: wtsRequests}
		contents.Judgments = append(contents.Judgments, judgmentRow{
			ID: "c" + string(rune('0'+index)), At: "2026-10-15T0" + string(rune('0'+index)) + ":00:00Z", Market: "KR",
			QuestionID: screenUpQuestion, Selection: record, StateRev: stateRev,
		})
	}
	report := renderCalibrationReport(contents, "l")
	if want := "· 간격 절반 초과 0 · WTS 요청 중앙 5 · 최대 63"; !strings.Contains(report, want) {
		t.Fatalf("report lacks %q:\n%s", want, report)
	}
}

// fakeNewsOnly 는 뉴스 피드만 답하는 in-process wtsSource(시각 해석 시험용).
type fakeNewsOnly struct{ items []domain.NewsItem }

func (f fakeNewsOnly) GetMarketNews(context.Context, string, int) (domain.MarketNews, error) {
	return domain.MarketNews{Items: f.items}, nil
}
func (fakeNewsOnly) GetMarketIndices(context.Context) (domain.MarketIndices, error) {
	return domain.MarketIndices{}, nil
}
func (fakeNewsOnly) GetTradingFlows(context.Context, string, int) (domain.TradingFlows, error) {
	return domain.TradingFlows{}, nil
}

// TestFeedWithNoParseableTimeIsAGapNotEmptyNews 는 피드를 받았어도 시각을 읽은 기사가 0 이면 news=absent + 이름이
// 됨을 봄(2026-10-15 dry run: 53/53 해석 실패가 news=present·빈 목록으로 나간 결함의 회귀 시험).
func TestFeedWithNoParseableTimeIsAGapNotEmptyNews(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.probe.wts = fakeNewsOnly{items: []domain.NewsItem{{ID: "x", Title: "t", CreatedAt: "10월 7일"}, {ID: "y", CreatedAt: ""}}}
	ok, gaps := rig.probe.refreshNews(context.Background(), &wtsCycle{}, at(10, 0))
	if ok || !hasPrefix(gaps, "news: ALL_HIGHLIGHT: no usable item (received 2, created_at unparseable 2, after fetch time 0)") {
		t.Fatalf("ok=%v gaps=%v", ok, gaps)
	}
	for _, raw := range []string{"2026-10-07 09:30:00", "2026-10-07T09:30:00", "2026-10-07T00:30:00Z", "2026-10-07T09:30:00+09:00"} {
		got, err := newsTime(raw)
		if err != nil || !got.Equal(at(9, 30)) {
			t.Fatalf("newsTime(%q) = %v, %v — zone-less times are Asia/Seoul", raw, got, err)
		}
	}
}

// TestEmptyFeedAndEmptyFlowsAreGaps 는 리뷰 지적(2026-10-15) 둘: 0 건 피드(열화된 200)가 news=present·빈 목록이 되지 않고,
// 직전 거래일 행이 0 인 수급이 「읽었고 0 행」으로 하루 캐시되지 않음(이름 + 다음 사이클 재시도)을 봄.
func TestEmptyFeedAndEmptyFlowsAreGaps(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	fake := &fakeFlows{}
	rig.probe.wts = fake
	ok, gaps := rig.probe.refreshNews(context.Background(), &wtsCycle{}, at(10, 0))
	if ok || !hasPrefix(gaps, "news: ALL_HIGHLIGHT: no usable item (received 0") {
		t.Fatalf("empty feed: ok=%v gaps=%v", ok, gaps)
	}
	fake.flows = []domain.TradingFlow{{Date: "2026-10-07"}} // 오늘 행뿐
	for cycle := 1; cycle <= 2; cycle++ {
		if _, err := rig.probe.readStockFlows(context.Background(), &wtsCycle{}, "005930", "2026-10-07"); err == nil ||
			!strings.Contains(err.Error(), "no prior-day row in 1 rows") {
			t.Fatalf("cycle %d: empty prior flows must be an error, got %v", cycle, err)
		}
	}
	if fake.flowCalls != 2 {
		t.Fatalf("an empty result must not be day-cached: %d reads over two cycles", fake.flowCalls)
	}
}

// TestConsecutiveServerFailuresStopTheCycleFlowReads 는 수급 읽기가 서버 쪽에서 연속 3 번 실패하면(5xx — 거절 아님) 그
// 사이클의 나머지 종목은 요청하지 않고 이름을 남김을 봄(풀 30 × 요청 2 의 실패 폭주 방지).
func TestConsecutiveServerFailuresStopTheCycleFlowReads(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	wts := rig.withWTS(t, true)
	wts.failStatus[pathTrend] = 503
	cycle := &wtsCycle{}
	var errs []error
	for _, symbol := range []string{"000001", "000002", "000003", "000004", "000005"} {
		_, err := rig.probe.readStockFlows(context.Background(), cycle, symbol, "2026-10-07")
		errs = append(errs, err)
	}
	if wts.count(pathTrend) != flowFailureStop {
		t.Fatalf("trend requested %d times, want %d then stop", wts.count(pathTrend), flowFailureStop)
	}
	if !strings.Contains(errs[4].Error(), "skipped after 3 consecutive WTS failures this cycle") {
		t.Fatalf("the stopped read must be named: %v", errs[4])
	}
	// 요청 없이 나는 종목별 오류(영숫자 코드 — 클라이언트가 국내 종목으로 보지 않음)는 세지 않음.
	fresh := &wtsCycle{}
	for index := 0; index < 5; index++ {
		_, _ = rig.probe.readStockFlows(context.Background(), fresh, "0167A0", "2026-10-07")
	}
	if fresh.flowsStopped != nil {
		t.Fatalf("client-side symbol errors must not stop the cycle: %v", fresh.flowsStopped)
	}
}

// TestHeadlinesApplyTheWindowAtTheirOwnJudgment 는 보관소에 다른 시장의 정리 기준으로 남은 24 시간 밖 기사가
// 이 판단의 at 으로 다시 걸러짐을 봄(정리는 메모리 상한, 고르는 쪽이 규칙).
func TestHeadlinesApplyTheWindowAtTheirOwnJudgment(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	rig.probe.newsStore["old"] = storedNews{item: domain.NewsItem{ID: "old", Title: "낡음", Nation: "kr"}, created: at(9, 0).Add(-25 * time.Hour)}
	rig.probe.newsStore["new"] = storedNews{item: domain.NewsItem{ID: "new", Title: "새것", Nation: "kr"}, created: at(9, 0)}
	got := rig.probe.headlinesFor("KR", "005930", "", at(10, 0))
	if len(got.Market) != 1 || got.Market[0].Title != "새것" {
		t.Fatalf("market headlines = %+v", got.Market)
	}
}

type fakeFlows struct {
	fakeNewsOnly
	flows     []domain.TradingFlow
	flowCalls int
}

func (f *fakeFlows) GetTradingFlows(context.Context, string, int) (domain.TradingFlows, error) {
	f.flowCalls++
	return domain.TradingFlows{Flows: f.flows}, nil
}

// seqFlows 는 수급 읽기 결과를 차례로 돌려주는 fake(nil = 성공).
type seqFlows struct {
	fakeNewsOnly
	results []error
}

func (f *seqFlows) GetTradingFlows(context.Context, string, int) (domain.TradingFlows, error) {
	err := f.results[0]
	f.results = f.results[1:]
	return domain.TradingFlows{Flows: []domain.TradingFlow{{Date: "2026-10-06"}}}, err
}

// TestFlowFailureCountIsConsecutive 는 성공이 연속 실패 수를 되돌림을 봄 — 실패 2·성공·실패 2 는 멈추지 않음.
func TestFlowFailureCountIsConsecutive(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	serverErr := &tossclient.StatusError{StatusCode: 503, Endpoint: "trend"}
	rig.probe.wts = &seqFlows{results: []error{serverErr, serverErr, nil, serverErr, serverErr}}
	cycle := &wtsCycle{}
	for index, symbol := range []string{"000001", "000002", "000003", "000004", "000005"} {
		_, _ = rig.probe.readStockFlows(context.Background(), cycle, symbol, "2026-10-07")
		if cycle.flowsStopped != nil {
			t.Fatalf("stopped after read %d although no three failures were consecutive", index+1)
		}
	}
}
