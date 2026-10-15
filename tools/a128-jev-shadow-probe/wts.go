package main

// wts.go 는 1.8(state_rev a128-s3)에서 더한 WTS 조회 전용 입력 세 가지를 담당함.
//
// 왜: s2 판단 전부가 같은 결손을 달고 있었음(원장 실측 2026-10-15) — 뉴스(KR·US 100%), KR 종목별 투자자 매매
// (공식 investor-trading 은 KOSPI/KOSDAQ 시장 단위뿐), US 시장 맥락(공식 지표 카탈로그에 US 지수 없음).
// 셋 다 기존 WTS 클라이언트(internal/client)의 조회 함수가 이미 읽음 — 이 파일은 그 셋만 부름.
//
// 안전:
//   - wtsSource 인터페이스가 이 도구가 부를 수 있는 WTS 표면 전부(조회 셋). 주문·계좌·보유 메서드는 인터페이스에
//     없으므로 수집 코드에서 도달 불가(static_test.go 가 메서드 집합과 이름 사용을 고정함).
//   - 뉴스 피드는 계좌 맥락이 없는 범위만 씀(newsFeedScopes) — 보유·관심·추천 개인화 범위는 계좌의 종목을
//     드러내므로 금지(시험이 요청 본문의 범위와 소스 문자열 양쪽으로 고정).
//   - 세션은 기존 세션 저장소를 기존 클라이언트가 읽음(main.go). 이 도구는 쿠키·헤더 값을 로그·원장에 쓰지 않음 —
//     state 로 옮기는 것은 공개 시장 값(제목·요약·시각·출처, 순매수 주식 수, 지수 값)뿐.
//   - 실패는 그 블록만 빠지고 missing_inputs 에 이름이 남음(지어낸 값 없음, 사이클 중단 없음).
//
// 호출 경제: 뉴스 피드는 사이클당 범위 수만큼(전역 피드 — 풀 전체가 공유), 주요 지수는 사이클당 1, 종목별 투자자
// 매매는 (시장 현지 날짜, 종목) 하루 캐시. 한 사이클에서 WTS 가 세션을 거절(401/403)하면 그 사이클의 나머지 WTS
// 읽기를 부르지 않음(만료 세션에 종목마다 실패 요청을 보내지 않게).

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	tossclient "github.com/JungHoonGhae/tossinvest-cli/internal/client"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	wtssession "github.com/JungHoonGhae/tossinvest-cli/internal/session"
)

// wtsSource 는 *tossclient.Client 가 그대로 만족하는 WTS 조회 표면(뉴스 피드·주요 지수·종목 수급).
type wtsSource interface {
	GetMarketNews(ctx context.Context, scope string, limit int) (domain.MarketNews, error)
	GetMarketIndices(ctx context.Context) (domain.MarketIndices, error)
	GetTradingFlows(ctx context.Context, symbol string, size int) (domain.TradingFlows, error)
}

const (
	// groupWTS 는 WTS 요청의 계수 열쇠 — 공유 Open API 예산(broker_reads)과 섞지 않고 wts_requests 로 따로 셈.
	groupWTS = "WTS"

	newsWindow         = 24 * time.Hour // 이보다 오래된 기사는 state 에 싣지 않고 보관소에서도 버림
	symbolHeadlineCap  = 3              // 종목 관련 기사 상한(최신순)
	marketHeadlineCap  = 5              // 시장 기사 상한(최신순, 제목만)
	titleRuneCap       = 120
	summaryRuneCap     = 120
	stockFlowFetchSize = 6 // 직전 5 거래일 + 오늘(오늘 행은 버림)
	stockFlowDays      = 5
)

// newsFeedScopes 는 사이클마다 읽는 피드 범위 — 계좌 맥락이 없는 범위만(internal/client news.go: ALL_HIGHLIGHT 는
// 「계좌 맥락에 의존하지 않는 유일한 범위」, SOARING_STOCK 은 급등 종목 기사). 범위당 최대 50 건(서버 상한).
var newsFeedScopes = []string{"ALL_HIGHLIGHT", "SOARING_STOCK"}

// NewsState 는 뉴스 블록. Symbol = 관련 종목에 이 종목이 있는 기사, Market = 이 시장 국적의 최신 기사(제목만).
// 둘 다 판단 시각 이전 24 시간 안에 만들어진 기사만. 빈 목록 = 읽었고 해당 기사 없음.
type NewsState struct {
	Symbol []HeadlineState `json:"symbol"`
	Market []HeadlineState `json:"market"`
}

// HeadlineState 는 기사 한 건 — 글자 수 상한으로 잘림(토큰 예산).
type HeadlineState struct {
	Title     string `json:"title"`
	Summary   string `json:"summary,omitempty"`
	CreatedAt string `json:"created_at"`
	Source    string `json:"source,omitempty"`
}

// StockFlowState 는 KR 종목 하나의 하루 투자자별 순매수(주 수, WTS trading-trend 의 *BuyVolume). 직전 거래일까지만.
type StockFlowState struct {
	Date                 string  `json:"date"`
	IndividualNetShares  float64 `json:"individual_net_shares"`
	ForeignerNetShares   float64 `json:"foreigner_net_shares"`
	InstitutionNetShares float64 `json:"institution_net_shares"`
}

// MajorIndexState 는 WTS 주요 지수 하나(코스피·코스닥·나스닥·S&P 500·필라델피아 반도체·VIX — 2026-10-15 실측 6 종).
// ChangePctVsBase = latest / base - 1(%) — base 는 WTS basePrice. KR 두 지수에서 공식 직전 종가 기준 등락과 소수 넷째
// 자리까지 같음을 실측(2026-10-15) → 직전 종가. US 지수의 base 는 같은 필드라고 봄(별도 미실측).
type MajorIndexState struct {
	Code            string   `json:"code,omitempty"`
	Name            string   `json:"name"`
	Nation          string   `json:"nation,omitempty"`
	Last            float64  `json:"last"`
	ChangePctVsBase *float64 `json:"change_pct_vs_base,omitempty"`
}

// newWTSClient 는 기존 WTS 클라이언트를 세션(없으면 nil)과 함께 만들고, 요청마다 사이클 계수기에 groupWTS 로 셈.
// baseURL 이 비면 운영 주소(시험은 httptest 주소를 넘김). TradingPolicy 는 비움 — 주문 경로를 쓰지 않음.
func newWTSClient(sess *wtssession.Session, baseURL string) *tossclient.Client {
	return tossclient.New(tossclient.Config{
		HTTPClient: &http.Client{Timeout: 15 * time.Second, Transport: countingTransport{base: http.DefaultTransport}},
		APIBaseURL: baseURL, InfoBaseURL: baseURL, CertBaseURL: baseURL,
		Session: sess,
	})
}

// countingTransport 는 WTS HTTP 요청 하나마다 계수함 — GetTradingFlows 처럼 함수 하나가 요청 둘을 내는 경우도 정확히 셈.
type countingTransport struct{ base http.RoundTripper }

func (t countingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	countRead(req.Context(), groupWTS)
	return t.base.RoundTrip(req)
}

// wtsCycle 은 사이클 하나의 WTS 상태 — 세션 거절을 한 번 보면 나머지 WTS 읽기를 건너뜀.
type wtsCycle struct {
	rejected error
	// flowFailures 는 연속된 서버 쪽 수급 읽기 실패 수 — flowFailureStop 에 닿으면 그 사이클의 나머지 수급 읽기를 멈춤
	// (5xx·429 에 종목마다 실패 요청 둘을 보내지 않게). 성공하면 0 으로 되돌림.
	flowFailures int
	flowsStopped error
}

// flowFailureStop 은 한 사이클에서 수급 읽기를 멈추는 연속 서버 쪽 실패 수.
const flowFailureStop = 3

// newsFutureSkew 는 기사 시각이 피드를 받은 순간보다 이만큼 넘게 늦으면 시각 해석(시간대)이 틀린 것으로 보는 허용폭.
const newsFutureSkew = 5 * time.Minute

var errWTSNotConfigured = errors.New("WTS client not configured")

// wtsRead 는 WTS 읽기 하나: 클라이언트가 없거나 이 사이클에 세션이 이미 거절됐으면 부르지 않고 사유를 돌려줌.
// 401/403(AuthError)만 거절로 침 — 클라이언트 쪽 ErrNoSession(요청 없음)이나 종목별 오류는 다른 읽기를 막지 않음.
func wtsRead[T any](ctx context.Context, p *probe, w *wtsCycle, call func(context.Context, wtsSource) (T, error)) (T, error) {
	var zero T
	if p.wts == nil {
		return zero, errWTSNotConfigured
	}
	if w.rejected != nil {
		return zero, fmt.Errorf("skipped, the WTS session was rejected earlier this cycle: %v", w.rejected)
	}
	if p.cfg.ReadGap > 0 {
		if err := p.sleep(ctx, p.cfg.ReadGap); err != nil {
			return zero, err
		}
	}
	value, err := call(ctx, p.wts)
	var rejected *tossclient.AuthError
	if errors.As(err, &rejected) {
		w.rejected = err
	}
	return value, err
}

// storedNews 는 프로세스 안 기사 보관소의 한 건 — 사이클마다 피드를 읽어 쌓고 24 시간 지난 것은 버림.
// 피드가 최신 ≤50 건만 주므로 사이클을 거치며 쌓아야 종목 관련 기사가 늘어남(추가 호출 0).
type storedNews struct {
	item    domain.NewsItem
	created time.Time
}

// refreshNews 는 이 사이클의 피드를 읽어 보관소에 더함. 한 범위라도 읽었으면 ok — 이 사이클에 하나도 못 읽으면
// 보관소의 이전 기사도 쓰지 않음(낡은 피드를 「있음」으로 내보내지 않음). 실패한 범위는 이름으로 돌려줌.
func (p *probe) refreshNews(ctx context.Context, w *wtsCycle, at time.Time) (bool, []string) {
	var gaps []string
	ok := false
	for _, scope := range newsFeedScopes {
		feed, err := wtsRead(ctx, p, w, func(ctx context.Context, s wtsSource) (domain.MarketNews, error) {
			return s.GetMarketNews(ctx, scope, tossclient.MaxNewsLimit)
		})
		if err != nil {
			gaps = append(gaps, "news: "+scope+": "+reasonOf(err))
			continue
		}
		fetchedAt := p.now()
		unparseable, future, newestRaw := 0, 0, ""
		p.newsMu.Lock()
		for _, item := range feed.Items {
			created, err := newsTime(item.CreatedAt)
			if err != nil {
				unparseable++
				continue // 시각을 못 읽는 기사는 최근성·선행(판단 뒤 기사) 여부를 확인할 수 없으므로 싣지 않음
			}
			// 받은 순간보다 뒤의 기사는 있을 수 없음 — 시간대 해석이 틀렸다는 신호이므로 싣지 않고 셈.
			if created.After(fetchedAt.Add(newsFutureSkew)) {
				future++
				continue
			}
			if newestRaw == "" || item.CreatedAt > newestRaw {
				newestRaw = item.CreatedAt
			}
			key := item.ID
			if key == "" {
				key = item.Title + "|" + item.CreatedAt
			}
			p.newsStore[key] = storedNews{item: item, created: created}
		}
		for key, stored := range p.newsStore {
			if at.Sub(stored.created) > newsWindow {
				delete(p.newsStore, key)
			}
		}
		stored := len(p.newsStore)
		p.newsMu.Unlock()
		// 피드 관측(제목·본문 없이 수와 최신 시각 원문만) — 시각 해석(시간대)을 벽시계와 대조할 수 있게 함.
		p.logf("news feed %s: items %d · created_at unparseable %d · after fetch time %d · newest raw %q · store %d",
			scope, len(feed.Items), unparseable, future, newestRaw, stored)
		// 피드를 받았어도 쓸 수 있는 기사(시각을 읽었고 받은 순간 이전)가 0 이면 그 범위는 못 읽은 것으로 침 — 빈 목록을
		// 「기사 없음」으로 내보내지 않음(2026-10-15 실측: RFC3339 만 받던 첫 판에서 53/53 이 해석 실패해 news=present·빈 목록이
		// 됨). 0 건 응답도 같음(성한 피드는 언제나 기사가 있음 — 빈 200 은 열화로 봄).
		if usable := len(feed.Items) - unparseable - future; usable == 0 {
			gaps = append(gaps, fmt.Sprintf("news: %s: no usable item (received %d, created_at unparseable %d, after fetch time %d)",
				scope, len(feed.Items), unparseable, future))
			continue
		}
		ok = true
	}
	return ok, gaps
}

// headlinesFor 는 보관소에서 이 종목·이 시장의 기사를 고름. 판단 시각보다 늦은 기사는 넣지 않음(선행 정보 금지 —
// 라벨이 재는 미래를 state 가 미리 보게 됨). 종목 일치 = 관련 종목 코드가 "A"+심볼(KR 상품 코드)·심볼 그대로이거나
// 종목명이 공식 /stocks 이름과 같음(US 상품 코드는 티커가 아니므로 이름으로만 맞음).
func (p *probe) headlinesFor(market, symbol, name string, at time.Time) *NewsState {
	p.newsMu.Lock()
	var recent []storedNews
	for _, stored := range p.newsStore {
		// 보관소 정리는 정리한 시장의 at 기준이라 다른 시장의 at 으로는 24 시간을 조금 넘는 기사가 남을 수 있음 → 고르는
		// 쪽에서도 이 판단의 at 으로 창을 잼(정리는 메모리 상한, 이 검사가 state 의 규칙).
		if !stored.created.After(at) && at.Sub(stored.created) <= newsWindow {
			recent = append(recent, stored)
		}
	}
	p.newsMu.Unlock()
	sort.Slice(recent, func(a, b int) bool {
		if !recent[a].created.Equal(recent[b].created) {
			return recent[a].created.After(recent[b].created)
		}
		return recent[a].item.ID < recent[b].item.ID
	})
	out := &NewsState{Symbol: []HeadlineState{}, Market: []HeadlineState{}}
	nation := strings.ToLower(market)
	for _, stored := range recent {
		item := stored.item
		if len(out.Symbol) < symbolHeadlineCap && relatesTo(item, symbol, name) {
			out.Symbol = append(out.Symbol, HeadlineState{
				Title: clip(item.Title, titleRuneCap), Summary: clip(item.Summary, summaryRuneCap),
				CreatedAt: stored.created.UTC().Format(time.RFC3339), Source: clip(item.Source, 40),
			})
			continue // 종목 기사로 실은 것은 시장 목록에 겹쳐 싣지 않음
		}
		if len(out.Market) < marketHeadlineCap && strings.EqualFold(item.Nation, nation) {
			out.Market = append(out.Market, HeadlineState{
				Title: clip(item.Title, titleRuneCap), CreatedAt: stored.created.UTC().Format(time.RFC3339), Source: clip(item.Source, 40),
			})
		}
	}
	return out
}

// newsTime 은 피드의 createdAt 을 읽음. 시간대가 있으면 그대로, 없으면(클라이언트 fixture 모양 "2006-01-02 15:04:05")
// Asia/Seoul 로 읽음 — 토스 화면 시각이 KST 라는 가정이며 실측 전(첫 s3 사이클 로그의 newest raw 를 벽시계와 대조할 것).
func newsTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	seoul, err := marketLocation("KR")
	if err != nil {
		return time.Time{}, err
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04"} {
		if parsed, err := time.ParseInLocation(layout, raw, seoul); err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("created_at %q is not parseable", raw)
}

func relatesTo(item domain.NewsItem, symbol, name string) bool {
	for _, stock := range item.Stocks {
		code := strings.ToUpper(strings.TrimSpace(stock.Code))
		if code == "A"+symbol || code == strings.ToUpper(symbol) || (name != "" && strings.TrimSpace(stock.Name) == name) {
			return true
		}
	}
	return false
}

// clip 은 공백을 한 칸으로 접고 글자(rune) 수 상한으로 자름 — 잘렸으면 "…" 로 끝냄.
func clip(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return string(runes[:limit-1]) + "…"
}

// readStockFlows 는 KR 종목 하나의 투자자별 순매수를 하루 캐시로 읽음. 오늘 행은 장중 잠정치이고 하루 캐시가 그 값을
// 얼리므로 버리고 직전 거래일까지만 실음(그래서 캐시가 정확함).
// 직전 거래일 행이 0 이면 결손(오류)으로 돌려줌 — 빈 목록을 하루 동안 캐시해 「읽었고 0 행」으로 내보내지 않음.
func (p *probe) readStockFlows(ctx context.Context, w *wtsCycle, symbol, day string) ([]StockFlowState, error) {
	return dayCached(ctx, p, "KR|"+day+"|stock-flows|"+symbol, func() ([]StockFlowState, error) {
		if w.flowsStopped != nil {
			return nil, fmt.Errorf("skipped after %d consecutive WTS failures this cycle: %v", flowFailureStop, w.flowsStopped)
		}
		flows, err := wtsRead(ctx, p, w, func(ctx context.Context, s wtsSource) (domain.TradingFlows, error) {
			return s.GetTradingFlows(ctx, symbol, stockFlowFetchSize)
		})
		if err != nil {
			// 서버 쪽 실패(HTTP 상태·전송 오류)만 셈 — 영숫자 코드처럼 요청 없이 나는 종목별 오류는 다른 종목을 막지 않음.
			var status *tossclient.StatusError
			var transport *url.Error
			if errors.As(err, &status) || errors.As(err, &transport) {
				if w.flowFailures++; w.flowFailures >= flowFailureStop {
					w.flowsStopped = err
				}
			}
			return nil, err
		}
		w.flowFailures = 0
		prior, err := priorDayFlows(flows.Flows, day)
		if err == nil && len(prior) == 0 {
			err = fmt.Errorf("no prior-day row in %d rows", len(flows.Flows))
		}
		return prior, err
	})
}

// priorDayFlows 는 날짜가 today(시장 현지 YYYY-MM-DD)보다 이른 행만 최신순 stockFlowDays 개. 날짜를 하나라도 못 읽으면
// 전체를 결손으로 봄(어느 날의 흐름인지 모르는 값을 싣지 않음).
func priorDayFlows(flows []domain.TradingFlow, today string) ([]StockFlowState, error) {
	out := []StockFlowState{}
	for _, flow := range flows {
		date, err := flowDate(flow.Date)
		if err != nil {
			return nil, err
		}
		if date >= today {
			continue
		}
		out = append(out, StockFlowState{
			Date: date, IndividualNetShares: flow.NetIndividuals, ForeignerNetShares: flow.NetForeigner,
			InstitutionNetShares: flow.NetInstitution,
		})
	}
	sort.Slice(out, func(a, b int) bool { return out[a].Date > out[b].Date })
	if len(out) > stockFlowDays {
		out = out[:stockFlowDays]
	}
	return out, nil
}

func flowDate(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{"2006-01-02", "20060102", time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			if layout == time.RFC3339 {
				return raw[:10], nil // 날짜 부분만 — 시장 현지 표기라고 봄(시각 표기는 미실측)
			}
			return parsed.Format("2006-01-02"), nil
		}
	}
	return "", fmt.Errorf("trading-flow date %q is not parseable", raw)
}

// readMajorIndices 는 WTS 주요 지수를 사이클당 한 번 읽음(실시간 값이라 캐시하지 않음). 실패하면 nil 과 사유,
// 값이 0 이하인 지수는 빼고 이름을 남김.
func (p *probe) readMajorIndices(ctx context.Context, w *wtsCycle) ([]MajorIndexState, []string) {
	indices, err := wtsRead(ctx, p, w, func(ctx context.Context, s wtsSource) (domain.MarketIndices, error) {
		return s.GetMarketIndices(ctx)
	})
	if err != nil {
		return nil, []string{"market_context: major_indices: " + reasonOf(err)}
	}
	out := []MajorIndexState{}
	var gaps []string
	for _, index := range indices.Indices {
		if index.Latest <= 0 {
			gaps = append(gaps, "market_context: major_indices: "+index.Name+" has no positive value")
			continue
		}
		out = append(out, MajorIndexState{
			Code: index.Code, Name: index.Name, Nation: index.Nation, Last: index.Latest,
			ChangePctVsBase: pctChange(index.Latest, index.Base),
		})
	}
	if len(out) == 0 {
		return nil, append(gaps, "market_context: major_indices: the response carried no positive index value")
	}
	return out, gaps
}
