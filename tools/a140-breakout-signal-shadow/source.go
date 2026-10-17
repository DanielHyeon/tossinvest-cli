package main

// source.go 는 브로커 표면 — 읽기 GET 다섯뿐(static_test.go 가 메서드 집합을 고정).
//
// 엔진과 같은 Open API 키를 쓰므로 모든 GET 앞에 간격을 두고, rate limit 을 이렇게 다룸(design §2.1-10):
//   - MARKET_DATA(호가 · 현재가 — 엔진의 /prices · /orderbook 과 같은 그룹): 재시도하지 않고, 한 번이라도 받으면 호가 읽기를
//     suspend 동안 멈춤. 엔진의 보호 경로 가격 읽기와 다투지 않게 하려는 것(안전 불변식 4 인접).
//   - 그 밖(봉 · 랭킹 · 달력): 정해진 지연으로 재시도, 다 쓰면 그 호출만 포기(호출자가 gap 행).

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// marketSource 는 *official.Client 가 그대로 만족하는 읽기 전용 표면.
type marketSource interface {
	Rankings(ctx context.Context, typ, marketCountry, duration string, excludeCaution bool, count int) (domain.Ranking, error)
	TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error)
	StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error)
	StrictOrderbookTop(ctx context.Context, market, symbol string) (official.StrictTopOfBook, error)
	StrictLastPrice(ctx context.Context, market, symbol string) (official.StrictLastPrice, error)
}

// rate-limit 그룹(openapi 「Rate Limits Group」 — a128 collect.go 와 같은 이름).
const (
	groupRanking         = "RANKING"
	groupMarketData      = "MARKET_DATA"
	groupMarketDataChart = "MARKET_DATA_CHART"
	groupMarketInfo      = "MARKET_INFO"
)

var (
	errQuoteSuspended  = errors.New("quote reads are suspended after a MARKET_DATA rate limit")
	errRateLimitGaveUp = errors.New("rate_limited_gave_up")
)

// pacedSource 는 간격 · 재시도 · 그룹별 계수 · 호가 정지를 씌운 표면. officialbars 의 CandleReader · QuoteReader 를 만족함.
type pacedSource struct {
	src          marketSource
	gap          time.Duration
	retryDelays  []time.Duration
	suspend      time.Duration
	now          func() time.Time
	sleep        func(context.Context, time.Duration) error
	mu           sync.Mutex
	reads        map[string]int
	quoteResumes time.Time
}

func newPacedSource(src marketSource, gap time.Duration, retryDelays []time.Duration, suspend time.Duration,
	now func() time.Time, sleep func(context.Context, time.Duration) error,
) *pacedSource {
	return &pacedSource{src: src, gap: gap, retryDelays: retryDelays, suspend: suspend, now: now, sleep: sleep, reads: map[string]int{}}
}

func (p *pacedSource) count(group string) {
	p.mu.Lock()
	p.reads[group]++
	p.mu.Unlock()
}

// readsByGroup 은 지금까지의 그룹별 GET 수를 복사해 돌려주고 계수기를 비움(budget 행 하나 = 그 사이의 GET).
func (p *pacedSource) readsByGroup() map[string]int {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := p.reads
	p.reads = map[string]int{}
	return out
}

func (p *pacedSource) pause(ctx context.Context, d time.Duration) error {
	if d <= 0 || p.sleep == nil {
		return ctx.Err()
	}
	return p.sleep(ctx, d)
}

// retrying 은 MARKET_DATA 밖 그룹의 호출 — rate limit 만 재시도.
func retrying[T any](ctx context.Context, p *pacedSource, group string, call func(context.Context) (T, error)) (T, error) {
	var zero T
	for attempt := 0; ; attempt++ {
		delay := p.gap
		if attempt > 0 {
			delay = p.retryDelays[attempt-1]
		}
		if err := p.pause(ctx, delay); err != nil {
			return zero, err
		}
		p.count(group)
		value, err := call(ctx)
		if err == nil || !errors.Is(err, official.ErrRateLimited) {
			return value, err
		}
		if attempt >= len(p.retryDelays) {
			return zero, fmt.Errorf("%w after %d attempts: %w", errRateLimitGaveUp, attempt+1, err)
		}
	}
}

// quoteRead 는 MARKET_DATA 호출 — 정지 중이면 브로커를 부르지 않고, rate limit 이면 정지를 시작함.
func quoteRead[T any](ctx context.Context, p *pacedSource, call func(context.Context) (T, error)) (T, error) {
	var zero T
	p.mu.Lock()
	suspended := p.now().Before(p.quoteResumes)
	p.mu.Unlock()
	if suspended {
		return zero, errQuoteSuspended
	}
	if err := p.pause(ctx, p.gap); err != nil {
		return zero, err
	}
	p.count(groupMarketData)
	value, err := call(ctx)
	if errors.Is(err, official.ErrRateLimited) {
		p.mu.Lock()
		p.quoteResumes = p.now().Add(p.suspend)
		p.mu.Unlock()
	}
	return value, err
}

func (p *pacedSource) Rankings(ctx context.Context, typ, country, duration string, excludeCaution bool, count int) (domain.Ranking, error) {
	return retrying(ctx, p, groupRanking, func(ctx context.Context) (domain.Ranking, error) {
		return p.src.Rankings(ctx, typ, country, duration, excludeCaution, count)
	})
}

func (p *pacedSource) TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error) {
	return retrying(ctx, p, groupMarketInfo, func(ctx context.Context) (official.MarketCalendarResponse, error) {
		return p.src.TypedMarketCalendar(ctx, country, date)
	})
}

func (p *pacedSource) StrictMinuteCandles(ctx context.Context, market, symbol string, count int, before string) (official.StrictMinutePage, error) {
	return retrying(ctx, p, groupMarketDataChart, func(ctx context.Context) (official.StrictMinutePage, error) {
		return p.src.StrictMinuteCandles(ctx, market, symbol, count, before)
	})
}

func (p *pacedSource) StrictOrderbookTop(ctx context.Context, market, symbol string) (official.StrictTopOfBook, error) {
	return quoteRead(ctx, p, func(ctx context.Context) (official.StrictTopOfBook, error) {
		return p.src.StrictOrderbookTop(ctx, market, symbol)
	})
}

func (p *pacedSource) StrictLastPrice(ctx context.Context, market, symbol string) (official.StrictLastPrice, error) {
	return quoteRead(ctx, p, func(ctx context.Context) (official.StrictLastPrice, error) {
		return p.src.StrictLastPrice(ctx, market, symbol)
	})
}
