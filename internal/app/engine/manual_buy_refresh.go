package engine

// manual_buy_refresh.go 는 엔진 밖 수동 매수(토스 앱 등)를 읽는 보조 실행자임(a136 D9=b).
//
// 왜 감독 루프가 아니라 보조 실행자인가: 이 읽기가 멈춰도 손절 · 청산 · 대사는 계속되어야 함(안전 불변식 4). 멈추면 보드가
// 낡아 journal 이 매수만 거부함 — 그것이 이 실행자가 가질 수 있는 영향의 전부임.
//
// 왜 precheck 밖인가: 첫 다리 precheck 는 「journal 변경 · 브로커 호출 없음」 계약이고 PrecheckQFinalEntry 는 ctx 가 없음.
// 그래서 읽기는 여기서 주기적으로 하고, 발급 문은 게시된 스냅숏만 읽음.
//
// 브로커 호출 예산(§0.4): 주기(30초)마다 GET /api/v1/orders 를 status=OPEN 한 번(목록 전체 — limit · cursor 무시) +
// status=CLOSED 를 from/to(계좌 하루 앞뒤 하루씩)로 페이지 끝까지. 평소 2회/30초, 페이지 상한 20 을 넘으면 산정 불가.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/manualbuy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	// manualBuyRefreshName 은 보조 실행자 이름(감독 루프 아님 — Runtime.LoopNames 에 나오면 안 됨).
	manualBuyRefreshName = "manual-buy-refresh"
	// manualBuyRefreshInterval 은 읽기 주기 — journal.ManualBuyStaleness(2분)의 1/4.
	manualBuyRefreshInterval = 30 * time.Second
	// manualBuyClosedPageCap 은 CLOSED 조회 페이지 상한. 넘으면 끝을 못 본 것이므로 산정 불가.
	manualBuyClosedPageCap = 20
	// manualBuyClosedPageLimit 은 CLOSED 페이지 크기 — 문서상 최대(기본 20, 최대 100). 20 페이지 × 100 = 3일치 종결 2,000 건.
	manualBuyClosedPageLimit = 100
)

// manualBuyOrderPages 는 이 실행자가 쓰는 유일한 브로커 읽기(OfficialReads 의 부분집합 — 시험이 대신 넣을 수 있게).
type manualBuyOrderPages interface {
	OrdersPageRaw(ctx context.Context, filter official.OrdersFilter, cursor string) (official.RawOrderPage, error)
}

// manualBuyEngineOrders 는 엔진 자신의 주문 id 원천(journal).
type manualBuyEngineOrders interface {
	EngineBrokerOrderIDs(ctx context.Context, accountRef string) (map[string]bool, error)
}

type manualBuyRefresh struct {
	board    *execgw.ManualBuyBoard
	reads    manualBuyOrderPages
	orders   manualBuyEngineOrders
	account  string
	clock    clock.Clock
	interval time.Duration
	log      *obs.Logger
}

// ManualBuyRefresher 는 a136 수동 매수 refresher 보조 실행자를 돌려줌. 게이트가 꺼져 생산 Guardian 이 없으면 아무것도 읽지 않고
// 취소를 기다림(읽을 이유가 없음 — 진입 발급자가 없음).
func (c *Context) ManualBuyRefresher(clk clock.Clock) AuxiliaryExecutor {
	r := &manualBuyRefresh{interval: manualBuyRefreshInterval, clock: clk}
	if c != nil {
		if guardian, ok := c.Guardian.(*execgw.RiskGuardian); ok {
			r.board = guardian.ManualBuyBoard()
		}
		if c.Official != nil {
			r.reads = c.Official
		}
		if c.Journal != nil {
			r.orders = c.Journal
		}
		r.account = c.AccountRef
		r.log = c.Log
		if r.clock == nil {
			r.clock = c.clock()
		}
	}
	return AuxiliaryExecutor{Name: manualBuyRefreshName, Run: r.Run, StopEvent: obs.EventManualBuyRefreshFailed}
}

// Run 은 기동 직후 한 번, 그 뒤 interval 마다 읽음. 취소되면 ctx.Err() 를 돌려줌(런타임은 그것을 정상 정지로 읽음).
func (r *manualBuyRefresh) Run(ctx context.Context) error {
	if r.board == nil || r.reads == nil || r.orders == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		r.refreshOnce(ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// refreshOnce 는 한 번 읽어 보드에 게시함. 실패도 게시함(산정 불가 스냅숏) — 이전 성공을 그대로 두면 낡을 때까지 그 값이 쓰임.
func (r *manualBuyRefresh) refreshOnce(ctx context.Context) {
	started := r.clock.Now().UTC()
	day, err := clock.AccountTradingDay(started)
	snapshot := execgw.ManualBuySnapshot{AsOf: started, Day: day}
	if err == nil {
		snapshot.ByCurrency, err = r.collect(ctx, day)
	}
	if err != nil {
		if ctx.Err() != nil {
			return // 정지 중 — 게시하지 않음(보드는 곧 버려짐).
		}
		snapshot.Err = err
		snapshot.ByCurrency = nil
		if r.log != nil {
			r.log.Error(obs.EventManualBuyRefreshFailed, err, "detail",
				"buys are refused until the manual buy snapshot can be computed again; sells and exits are unaffected")
		}
	}
	r.board.Publish(snapshot)
}

func (r *manualBuyRefresh) collect(ctx context.Context, day string) (map[string]string, error) {
	engineOrders, err := r.orders.EngineBrokerOrderIDs(ctx, r.account)
	if err != nil {
		return nil, err
	}
	open, err := r.reads.OrdersPageRaw(ctx, official.OrdersFilter{Status: "OPEN"}, "")
	if err != nil {
		return nil, fmt.Errorf("reading open orders: %w", err)
	}
	raws := append([]json.RawMessage(nil), open.Orders...)
	from, to, err := manualBuyDateRange(day)
	if err != nil {
		return nil, err
	}
	cursor := ""
	for page := 0; ; page++ {
		if page >= manualBuyClosedPageCap {
			return nil, fmt.Errorf("%w: closed orders run past %d pages", manualbuy.ErrUnvaluable, manualBuyClosedPageCap)
		}
		closed, err := r.reads.OrdersPageRaw(ctx, official.OrdersFilter{Status: "CLOSED", From: from, To: to, Limit: manualBuyClosedPageLimit}, cursor)
		if err != nil {
			return nil, fmt.Errorf("reading closed orders: %w", err)
		}
		raws = append(raws, closed.Orders...)
		if !closed.HasNext {
			break
		}
		if closed.NextCursor == "" || closed.NextCursor == cursor {
			return nil, errors.New("closed orders report a next page with no new cursor")
		}
		cursor = closed.NextCursor
	}
	return manualbuy.Summarize(raws, engineOrders, day)
}

// manualBuyDateRange 는 CLOSED 조회의 from/to(OpenAPI: orderedAt 의 KST 날짜) — 라벨 D 의 하루(서울 D 08:00 ~ D+1 08:00)는 KST 날짜
// D · D+1 에 걸치므로 D 는 반드시, 여유로 앞뒤 하루씩 넓힘. 정확한 귀속은 Summarize 가 주문 시각으로 다시 함.
func manualBuyDateRange(day string) (string, string, error) {
	d, err := time.Parse("2006-01-02", day)
	if err != nil {
		return "", "", fmt.Errorf("account day %q: %w", day, err)
	}
	return d.AddDate(0, 0, -1).Format("2006-01-02"), d.AddDate(0, 0, 1).Format("2006-01-02"), nil
}

var _ manualBuyEngineOrders = (*journal.Journal)(nil)
