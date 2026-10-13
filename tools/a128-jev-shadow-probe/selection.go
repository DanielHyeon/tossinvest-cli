package main

// selection.go 는 사이클마다 판단할 종목을 고르는 표본 선택 규칙을 담당함(tasks.md 1.5·1.6).
//
// 규칙 jev_screen<K>_of_random<POOL>_of_top100_le_order_cap (1.6, 기본 K=10·POOL=30):
//  1. official MARKET_TRADING_AMOUNT 랭킹 상위 100 을 받음.
//  2. 1주 가격(랭킹 행의 lastPrice)을 주문 상한 통화로 환산해 상한 이하인 종목만 남김 —
//     엔진이 실제로 살 수 있는 집합에서 표본을 뜨기 위함. 상한 = config
//     `engine.automation_gate.max_order_notional`(+ `limit_currency`), Guardian 이 1회 주문에 거는
//     금액 상한. 환산 방향도 Guardian 과 같음: 가격 × 환율(시세 통화 → 상한 통화) ≤ 상한.
//  3. 남은 후보 중 무작위 POOL(후보가 POOL 이하이면 전부). 시드는 사이클마다 새로 뽑아
//     judgment 행의 selection 에 적음 — 같은 후보 목록과 시드로 같은 풀을 재현할 수 있음.
//  4. 풀 전체를 Jev 로 판단하고 J2 악재 제외 → J1 상위 K 를 결선으로 표시함(screen.go).
//
// 1.5 규칙 random10_of_top100_le_order_cap(무작위 10 을 그대로 판단)과 그 이전 top10 은 더 만들지 않지만,
// 원장의 옛 행은 그 이름의 로트로 보고서에 그대로 남음.
//
// 상한·환율을 못 읽으면 조용히 옛 규칙(상위 10)으로 돌아가지 않고 그 사이클을 gap 행
// (stage=selection)으로 남기고 판단하지 않음. 계좌 API 는 쓰지 않음 — config 파일 읽기만.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"strings"

	tossconfig "github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

const (
	// selectionRule 은 기본값(--top 10 · --pool 30)의 규칙 id. 다른 K·POOL 은 selectionRuleFor 가 id 에 넣어
	// 다른 로트로 갈라짐(모집단이 다르므로 같은 이름을 쓰면 안 됨).
	selectionRule = "jev_screen10_of_random30_of_top100_le_order_cap"
	// legacyRandomRule 은 1.5 규칙(선별 없이 무작위 10 을 판단)의 로트 이름 — 더 만들지 않고 옛 행만 이 이름.
	legacyRandomRule = "random10_of_top100_le_order_cap"
	// legacySelectionLot 는 selection 필드가 없는 옛 행(1.5 이전, 상위 10 고정)의 로트 이름.
	legacySelectionLot = "top10"
	// rankingFetch 는 필터 전에 받는 랭킹 행 수(상위 100).
	rankingFetch = 100
	// selectionPriceSource 는 필터가 쓰는 가격의 출처 — 랭킹 응답의 lastPrice.
	// 100 종목 /prices 한 번을 더 부르지 않으려는 선택이고, 기준가(base_price)는 여전히 /prices 값임.
	selectionPriceSource = "rankings.last_price"
	// selectionPCGStream 은 PCG 둘째 상태 단어. 시드 하나만 원장에 적으므로 이 값은 고정이어야 재현됨.
	selectionPCGStream uint64 = 0xa128_0105
	// seedMask 는 시드를 53 비트로 자름 — JSON 숫자를 float64 로 읽는 도구(jq 등)에서도 정확히 보존되게.
	seedMask uint64 = 1<<53 - 1
)

// marketCurrency 는 시장의 시세 통화. 랭킹 행 통화가 이와 다르면 가격을 모르는 것으로 셈.
var marketCurrency = map[string]string{"KR": "KRW", "US": "USD"}

// orderCap 은 1회 주문 금액 상한과 그 통화(config 에서 읽은 값).
type orderCap struct {
	Value    float64
	Currency string
}

// selectionRecord 는 judgment 행에 붙는 표본 선택 기록. 한 사이클의 모든 판단 행이 같은 값을 공유함.
type selectionRecord struct {
	Rule             string  `json:"rule"`
	Seed             uint64  `json:"seed"`
	CapValue         float64 `json:"cap_value"`
	CapCurrency      string  `json:"cap_currency"`
	PriceCurrency    string  `json:"price_currency"`
	PriceSource      string  `json:"price_source"`
	FXRate           float64 `json:"fx_rate,omitempty"` // 시세 통화 → 상한 통화, 같은 통화이면 생략(=1)
	FXPair           string  `json:"fx_pair,omitempty"`
	RankingRequested int     `json:"ranking_requested"`
	RankingReceived  int     `json:"ranking_received"`
	ExcludedOverCap  int     `json:"excluded_over_cap"`
	// ExcludedUnknownPrice 는 가격이 0 이하·비유한이거나 통화가 비었거나 시장 통화와 달라 상한 비교를 못 한 행 수.
	ExcludedUnknownPrice int `json:"excluded_unknown_price"`
	Candidates           int `json:"candidates"`
	// Pool 은 요청한 풀 크기(--pool), Selected 는 실제로 뽑은 풀(후보가 적으면 Pool 미만).
	Pool     int `json:"pool"`
	Selected int `json:"selected"`
	// Judged 는 풀 중 판단에 성공한 종목 수(선별 대상). 실패 종목은 gap(judgment) 행.
	Judged          int `json:"judged"`
	RedFlagExcluded int `json:"red_flag_excluded"`
	// FinalistsRequested 는 --top, Finalists 는 실제 결선 수(생존이 적으면 미만).
	FinalistsRequested int `json:"finalists_requested"`
	Finalists          int `json:"finalists"`
	// 1.7 — 공유 Open API 예산 관측: 사이클의 브로커 GET 수(재시도 포함, rate-limit 그룹별)·하루 캐시 적중 수·
	// 상한 읽기부터 판단 끝까지의 수집 시간·간격.
	BrokerReads     int        `json:"broker_reads"`
	ReadsByGroup    readGroups `json:"broker_reads_by_group"`
	CacheHits       int        `json:"day_cache_hits"`
	CollectSeconds  float64    `json:"collect_seconds"`
	IntervalSeconds int        `json:"interval_seconds"`
}

// readGroups 는 rate-limit 그룹별 GET 수. 고정 필드라 selectionRecord 가 비교 가능하게 남음.
type readGroups struct {
	Ranking              int `json:"RANKING"`
	Stock                int `json:"STOCK"`
	MarketData           int `json:"MARKET_DATA"`
	MarketDataChart      int `json:"MARKET_DATA_CHART"`
	MarketInfo           int `json:"MARKET_INFO"`
	MarketIndicator      int `json:"MARKET_INDICATOR"`
	MarketIndicatorChart int `json:"MARKET_INDICATOR_CHART"`
}

func readGroupsOf(counts map[string]int) readGroups {
	return readGroups{
		Ranking: counts[groupRanking], Stock: counts[groupStock], MarketData: counts[groupMarketData],
		MarketDataChart: counts[groupMarketDataChart], MarketInfo: counts[groupMarketInfo],
		MarketIndicator: counts[groupMarketIndicator], MarketIndicatorChart: counts[groupMarketIndicatorChart],
	}
}

func (g readGroups) total() int {
	return g.Ranking + g.Stock + g.MarketData + g.MarketDataChart + g.MarketInfo + g.MarketIndicator + g.MarketIndicatorChart
}

// loadOrderCap 은 tossctl config 파일에서 주문 상한을 읽음(읽기 전용 — Status 는 파일을 쓰지 않음).
// 파일이 없으면 config 로더는 기본값을 돌려주므로 Exists 로 따로 거름 — 기본값(상한 0)을 상한으로
// 오인하지 않게 하기 위함.
func loadOrderCap(ctx context.Context, configFile string) (orderCap, error) {
	status, err := tossconfig.NewService(configFile).Status(ctx)
	if err != nil {
		return orderCap{}, fmt.Errorf("reading the config %s: %w", configFile, err)
	}
	if !status.Exists {
		return orderCap{}, fmt.Errorf("config %s not found", configFile)
	}
	gate := status.Engine.AutomationGate
	value := gate.MaxOrderNotional
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return orderCap{}, fmt.Errorf("config engine.automation_gate.max_order_notional is not a positive finite number (%v)", value)
	}
	currency := strings.ToUpper(strings.TrimSpace(gate.LimitCurrency))
	if currency != "KRW" && currency != "USD" {
		return orderCap{}, fmt.Errorf("config engine.automation_gate.limit_currency %q is not KRW or USD", gate.LimitCurrency)
	}
	return orderCap{Value: value, Currency: currency}, nil
}

// errNoCapSource 는 상한 출처가 배선되지 않은 probe — 상한 없이 표본을 뜨지 않게 거부함.
var errNoCapSource = errors.New("no order cap source is configured")

// selectionInputs 는 사이클 시작 시 정한 상한과 환율(같은 통화이면 1).
type selectionInputs struct {
	cap           orderCap
	priceCurrency string
	fxRate        float64
	fxPair        string
}

// resolveSelectionInputs 는 상한을 읽고, 시세 통화가 상한 통화와 다르면 official 환율 GET 으로
// 환산 계수를 얻음. 어느 하나라도 실패하면 오류 — 호출자가 gap 행으로 남김.
func (p *probe) resolveSelectionInputs(ctx context.Context, market string) (selectionInputs, error) {
	if p.cfg.CapSource == nil {
		return selectionInputs{}, errNoCapSource
	}
	limit, err := p.cfg.CapSource(ctx)
	if err != nil {
		return selectionInputs{}, fmt.Errorf("order cap: %w", err)
	}
	priceCurrency := marketCurrency[market]
	if priceCurrency == "" {
		return selectionInputs{}, fmt.Errorf("no quote currency known for market %q", market)
	}
	in := selectionInputs{cap: limit, priceCurrency: priceCurrency, fxRate: 1}
	if priceCurrency == limit.Currency {
		return in, nil
	}
	// ExchangeRate(base, quote) 는 "1 base = rate quote" — 시세 통화 1 단위가 상한 통화로 얼마인지.
	rate, err := withRateLimitRetry(ctx, p, groupMarketInfo, func(ctx context.Context) (domain.ExchangeRate, error) {
		return p.src.ExchangeRate(ctx, priceCurrency, limit.Currency)
	})
	if err != nil {
		return selectionInputs{}, fmt.Errorf("exchange rate %s/%s: %w", priceCurrency, limit.Currency, err)
	}
	// 짝이 뒤집힌 응답을 그대로 곱하면 상한이 환율² 배로 틀어지므로 짝을 확인함.
	if !strings.EqualFold(rate.BaseCurrency, priceCurrency) || !strings.EqualFold(rate.QuoteCurrency, limit.Currency) {
		return selectionInputs{}, fmt.Errorf("exchange rate answered pair %s/%s, want %s/%s",
			rate.BaseCurrency, rate.QuoteCurrency, priceCurrency, limit.Currency)
	}
	if math.IsNaN(rate.Close) || math.IsInf(rate.Close, 0) || rate.Close <= 0 {
		return selectionInputs{}, fmt.Errorf("exchange rate %s/%s is not a positive finite number (%v)", priceCurrency, limit.Currency, rate.Close)
	}
	in.fxRate, in.fxPair = rate.Close, priceCurrency+"/"+limit.Currency
	return in, nil
}

// filterByCap 은 랭킹 행을 순서대로 보며 상한 이하만 후보로 남기고, 탈락 사유별 수를 셈.
// 경계: 환산 가격 == 상한은 통과(Guardian 도 "초과"만 거부함).
func filterByCap(items []domain.RankingItem, in selectionInputs) (candidates []domain.RankingItem, overCap, unknown int) {
	for _, item := range items {
		price := item.LastPrice
		// 통화가 비었거나 다르면 환산 근거가 없으므로 가격을 모르는 것으로 셈(추정 금지).
		if !strings.EqualFold(item.Currency, in.priceCurrency) || math.IsNaN(price) || math.IsInf(price, 0) || price <= 0 {
			unknown++
			continue
		}
		if price*in.fxRate <= in.cap.Value {
			candidates = append(candidates, item)
		} else {
			overCap++
		}
	}
	return candidates, overCap, unknown
}

// pickSample 은 후보에서 k 개를 시드로 무작위 추출함(1.6 에서는 k = 풀 크기). 후보가 k 이하이면 전부(순서 유지).
// 재현 절차: 후보(랭킹 순) → rand.New(rand.NewPCG(seed, selectionPCGStream)).Perm(n) 앞 k 개 →
// 지표 오름차순 정렬(처리 순서가 랭킹 순이 되게).
func pickSample(candidates []domain.RankingItem, k int, seed uint64) []domain.RankingItem {
	if len(candidates) <= k {
		return append([]domain.RankingItem(nil), candidates...)
	}
	indexes := rand.New(rand.NewPCG(seed, selectionPCGStream)).Perm(len(candidates))[:k]
	sort.Ints(indexes)
	picked := make([]domain.RankingItem, 0, k)
	for _, index := range indexes {
		picked = append(picked, candidates[index])
	}
	return picked
}

// selectionRuleFor 는 결선 k·풀 pool 의 규칙 id — (10, 30) 이면 selectionRule 과 같음.
func selectionRuleFor(k, pool int) string {
	return fmt.Sprintf("jev_screen%d_of_random%d_of_top%d_le_order_cap", k, pool, rankingFetch)
}

// newSelectionSeed 는 사이클마다 새 시드를 뽑음(전역 생성기는 프로세스마다 무작위 시드됨).
func newSelectionSeed() uint64 { return rand.Uint64() & seedMask }

// selectionLot 은 보고서 로트 이름 — selection 이 없는 옛 행은 상위 10 고정 로트. state_rev 가 있으면
// 「규칙@state_rev」(1.7) — 같은 규칙이라도 Jev 가 본 state 모양이 다르면 p 의 모집단이 다름.
func selectionLot(row judgmentRow) string {
	lot := legacySelectionLot
	if row.Selection != nil {
		lot = row.Selection.Rule
	}
	if row.StateRev != "" {
		lot += "@" + row.StateRev
	}
	return lot
}
