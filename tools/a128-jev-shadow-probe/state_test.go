package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// TestPublicStateKeySetIsPinned 는 외부로 나가는 state 의 열쇠 집합을 고정함.
// 필드를 더하면 이 목록도 손으로 바꿔야 하고, 그 diff 가 리뷰에서 보임.
func TestPublicStateKeySetIsPinned(t *testing.T) {
	t.Parallel()
	want := []string{
		"base_price", "change_pct_vs_reference", "currency", "horizon_minutes", "listing_market", "market",
		"minutes_since_open", "minutes_to_close", "missing_inputs", "name", "news", "observed_at",
		"ranking_rank", "ranking_type", "reference_price", "session_source", "symbol", "top_of_book", "trading_amount",
		"trading_volume",
		"top_of_book.best_ask", "top_of_book.best_ask_volume", "top_of_book.best_bid",
		"top_of_book.best_bid_volume", "top_of_book.source_age_seconds", "top_of_book.spread_bps",
		// 1.7(state_rev a128-s2) — features.go.
		"intraday", "intraday.bars_available", "intraday.bars_1m", "intraday.bars_1m.t", "intraday.bars_1m.o",
		"intraday.bars_1m.h", "intraday.bars_1m.l", "intraday.bars_1m.c", "intraday.bars_1m.v",
		"intraday.return_5m_pct", "intraday.return_15m_pct", "intraday.return_30m_pct", "intraday.return_60m_pct",
		"intraday.window_high", "intraday.window_low", "intraday.price_in_window_range_pct", "intraday.window_vwap_typical",
		"intraday.vs_window_vwap_pct", "intraday.volume_last_5_bars_vs_earlier_avg", "intraday.volume_last_30_bars_vs_earlier_avg",
		"daily", "daily.prior_bars", "daily.last_prior_bar_at", "daily.return_1d_pct", "daily.return_5d_pct",
		"daily.return_20d_pct", "daily.high_20d", "daily.low_20d", "daily.vs_high_20d_pct", "daily.vs_low_20d_pct",
		"daily.avg_volume_20d",
		"market_context", "market_context.indices", "market_context.indices.symbol", "market_context.indices.last",
		"market_context.indices.change_today_pct", "market_context.indices.change_30m_pct",
		"market_context.investor_flows", "market_context.investor_flows.market", "market_context.investor_flows.date",
		"market_context.investor_flows.updated_at", "market_context.investor_flows.individual_net_krw",
		"market_context.investor_flows.foreigner_net_krw", "market_context.investor_flows.institution_net_krw",
		"warning_types", "price_limits", "price_limits.upper_limit", "price_limits.lower_limit",
		"price_limits.to_upper_pct", "price_limits.to_lower_pct",
		// 1.8(state_rev a128-s3) — wts.go.
		"market_context.major_indices", "market_context.major_indices.code", "market_context.major_indices.name",
		"market_context.major_indices.nation", "market_context.major_indices.last", "market_context.major_indices.change_pct_vs_base",
		"investor_flows_per_stock", "investor_flows_per_stock.date", "investor_flows_per_stock.individual_net_shares",
		"investor_flows_per_stock.foreigner_net_shares", "investor_flows_per_stock.institution_net_shares",
		"news_headlines", "news_headlines.symbol", "news_headlines.symbol.title", "news_headlines.symbol.summary",
		"news_headlines.symbol.created_at", "news_headlines.symbol.source", "news_headlines.market",
		"news_headlines.market.title", "news_headlines.market.summary", "news_headlines.market.created_at",
		"news_headlines.market.source",
	}
	var got []string
	var collect func(prefix string, typ reflect.Type)
	collect = func(prefix string, typ reflect.Type) {
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				t.Fatalf("PublicState field %s has no explicit json name", field.Name)
			}
			got = append(got, prefix+name)
			inner := field.Type
			// 포인터·슬라이스 안의 구조체까지 내려감 — 1분봉·지수 목록의 열쇠도 고정 목록에 들어야 함.
			for inner.Kind() == reflect.Pointer || inner.Kind() == reflect.Slice {
				inner = inner.Elem()
			}
			if inner.Kind() == reflect.Struct {
				collect(prefix+name+".", inner)
			}
		}
	}
	collect("", reflect.TypeOf(PublicState{}))
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PublicState keys changed:\n got %v\nwant %v", got, want)
	}
	// 고정 목록의 어느 열쇠도 금지 정규식에 걸리지 않아야 가드가 정상 state 를 막지 않음.
	for _, key := range want {
		if forbiddenStateKey.MatchString(key) {
			t.Fatalf("public key %q matches the forbidden pattern; the guard would refuse every state", key)
		}
	}
}

// TestStateGuardRefusesAccountShapedKeys 는 계좌·보유·잔고·토큰 열쇠가 어디에 있든 전송을 거부함을 봄.
func TestStateGuardRefusesAccountShapedKeys(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`{"state":{"accountNo":"1"}}`,
		`{"state":{"top_of_book":{"holdingQuantity":"3"}}}`,
		`{"state":{"items":[{"cash_balance":1}]}}`,
		`{"questions":{"q":{"instructions":"x"}},"accessToken":"t"}`,
		`{"state":{"api_key":"k"}}`,
		`{"state":{"X-Tossinvest-Account":"5"}}`,
		`{"state":{"positions":[]}}`,
	} {
		if err := assertPublicJSON([]byte(raw)); err == nil {
			t.Errorf("guard let %s through", raw)
		}
	}
	state := assembleState(marketInputs{Market: "KR", Rank: domain.RankingItem{Symbol: "005930"}})
	request, _ := json.Marshal(systemOneRequest{State: state, Model: "m", Questions: map[string]noulQuestion{
		"j1_up_within_horizon": probeQuestions()[0].Body, "j2_hold_off_red_flag": probeQuestions()[1].Body,
	}})
	if err := assertPublicJSON(request); err != nil {
		t.Fatalf("guard refuses the probe's own request: %v", err)
	}
}

// TestAccountFieldsInBrokerResponsesNeverReachTypeSafe 는 spec 시나리오
// 「state 조립기에 계좌 필드가 들어온다」를 전송 바이트 수준에서 고정함:
// mock official 응답에 계좌·보유·잔고 필드와 토큰을 섞어도, TypeSafe 로 나간 본문과
// 원장 디렉터리 어느 파일에도 그 값·열쇠가 없어야 함.
func TestAccountFieldsInBrokerResponsesNeverReachTypeSafe(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	requests := rig.typeSafe.requests()
	if len(requests) != 2 {
		t.Fatalf("expected one judgment request per symbol (2), got %d", len(requests))
	}
	var exposed []string
	for _, body := range requests {
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
		for _, forbidden := range append([]string{"accountNo", "holdingQuantity", "accountBalance", "access_token"}, sentinels...) {
			if strings.Contains(text, forbidden) {
				t.Fatalf("%q leaked into an outbound request or a ledger file:\n%s", forbidden, text)
			}
		}
	}
	// 키는 Authorization 헤더로만 나감.
	for _, header := range rig.typeSafe.auth {
		if header != "Bearer "+testTypeSafeKey {
			t.Fatalf("Authorization header = %q", header)
		}
	}
}

// TestNewsAbsenceIsExplicitInStateAndRow 는 spec 시나리오 「뉴스 소스가 없는 세션의 판단」:
// 전송 state 와 원장 행 양쪽에 뉴스 결손이 명시돼야 함.
func TestNewsAbsenceIsExplicitInStateAndRow(t *testing.T) {
	t.Parallel()
	rig := newTestRig(t)
	if err := rig.probe.cycle(context.Background(), "KR", rig.session(), at(10, 0)); err != nil {
		t.Fatal(err)
	}
	for _, body := range rig.typeSafe.requests() {
		var request struct {
			State PublicState `json:"state"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			t.Fatal(err)
		}
		if request.State.News != newsAbsent || request.State.NewsHeadlines != nil || !hasPrefix(request.State.MissingInputs, "news:") {
			t.Fatalf("outbound state does not name the news gap: news=%q missing=%v", request.State.News, request.State.MissingInputs)
		}
	}
	contents := rig.contents(t)
	if len(contents.Judgments) != 8 {
		t.Fatalf("expected 8 judgment rows (2 symbols × 4 questions), got %d", len(contents.Judgments))
	}
	for _, row := range contents.Judgments {
		if !hasPrefix(row.CollectionGaps, "news:") {
			t.Fatalf("judgment row %s does not name the news gap: %v", row.ID, row.CollectionGaps)
		}
	}
}

func TestAssembleStateNamesMissingInputs(t *testing.T) {
	t.Parallel()
	state := assembleState(marketInputs{
		Market: "KR", ObservedAt: at(10, 0), Session: session{Open: at(9, 0), Close: at(15, 30)},
		Rank:      domain.RankingItem{Rank: 3, Symbol: "005930", Currency: "KRW", LastPrice: 70000, BasePrice: 70000},
		LastPrice: 70000, Gaps: []string{"top_of_book: rate_limited_gave_up"},
	})
	if state.TopOfBook != nil || state.Name != "" {
		t.Fatalf("absent inputs must stay absent: %+v", state)
	}
	if !contains(state.MissingInputs, "top_of_book: rate_limited_gave_up") || !contains(state.MissingInputs, "news: not read") {
		t.Fatalf("missing inputs not named: %v", state.MissingInputs)
	}
	if state.ChangePctVsReference == nil || *state.ChangePctVsReference != 0 {
		t.Fatalf("change vs reference should be computed as 0: %v", state.ChangePctVsReference)
	}
	if state.MinutesSinceOpen != 60 || state.MinutesToClose != 330 {
		t.Fatalf("session position = %d/%d", state.MinutesSinceOpen, state.MinutesToClose)
	}
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}
