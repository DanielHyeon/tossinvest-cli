package main

// collect.go 는 수집 루프·세션 경계·rate limit 백오프·라벨러를 담당함.
//
// 브로커에는 읽기(GET)만 보냄: marketSource 인터페이스가 이 도구가 부를 수 있는 브로커 표면
// 전부이고, 그 메서드 다섯은 모두 공개 시장 데이터 GET 임. 주문·계좌 메서드는 인터페이스에
// 없으므로 수집 코드에서 도달 불가(static_test.go 가 메서드 집합을 고정함).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// marketSource 는 *official.Client 가 그대로 만족하는 읽기 전용 표면.
type marketSource interface {
	Rankings(ctx context.Context, typ, marketCountry, duration string, excludeCaution bool, count int) (domain.Ranking, error)
	Stocks(ctx context.Context, symbols []string) ([]domain.Quote, error)
	Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)
	StrictOrderbookTop(ctx context.Context, market, symbol string) (official.StrictTopOfBook, error)
	TypedMarketCalendar(ctx context.Context, country, date string) (official.MarketCalendarResponse, error)
}

const (
	rankingType     = "MARKET_TRADING_AMOUNT"
	rankingDuration = "realtime"
	// 라벨이 기한보다 이만큼 넘게 늦으면 실현가를 쓰지 않고 결손으로 적음(다른 horizon 을 재게 되므로).
	labelTolerance = 5 * time.Minute
	reasonLimit    = 160
)

// config 는 실행 매개변수. 기본값은 main.go 플래그에 있음.
type config struct {
	Interval    time.Duration
	TopK        int
	Horizon     time.Duration
	WaitOpen    bool
	LabelTick   time.Duration
	ReadGap     time.Duration   // 브로커 호출 사이 간격
	RetryDelays []time.Duration // 429 재시도 지연(지수) — 길이가 재시도 횟수
}

// session 은 정규장 하나의 경계.
type session struct {
	Open  time.Time
	Close time.Time
}

// probe 는 한 실행의 상태 전부.
type probe struct {
	src       marketSource
	judge     *judgeClient
	ledger    *ledger
	cfg       config
	questions []question
	now       func() time.Time
	sleep     func(context.Context, time.Duration) error
	logf      func(format string, args ...any)

	mu      sync.Mutex
	pending []pendingLabel
}

// pendingLabel 은 한 종목·한 사이클의 판단들(J1·J2)이 공유하는 라벨 대기 항목.
type pendingLabel struct {
	Market      string
	Symbol      string
	BasePrice   float64
	Due         time.Time
	JudgmentIDs []string
}

func newProbe(src marketSource, judge *judgeClient, l *ledger, cfg config) *probe {
	return &probe{
		src: src, judge: judge, ledger: l, cfg: cfg, questions: probeQuestions(),
		now: time.Now, sleep: sleepContext,
		logf: func(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...) },
	}
}

// withRateLimitRetry 는 official.ErrRateLimited 만 재시도함(지수 지연). 다른 오류는 즉시 돌려줌.
// 포기하면 errRateLimitGaveUp 로 감싸 원장 사유가 "포기"임을 분명히 함.
var errRateLimitGaveUp = errors.New("rate_limited_gave_up")

func withRateLimitRetry[T any](ctx context.Context, p *probe, call func(context.Context) (T, error)) (T, error) {
	var zero T
	for attempt := 0; ; attempt++ {
		if attempt > 0 || p.cfg.ReadGap > 0 {
			delay := p.cfg.ReadGap
			if attempt > 0 {
				delay = p.cfg.RetryDelays[attempt-1]
			}
			if err := p.sleep(ctx, delay); err != nil {
				return zero, err
			}
		}
		value, err := call(ctx)
		if err == nil {
			return value, nil
		}
		if !errors.Is(err, official.ErrRateLimited) {
			return zero, err
		}
		if attempt >= len(p.cfg.RetryDelays) {
			return zero, fmt.Errorf("%w after %d attempts: %v", errRateLimitGaveUp, attempt+1, err)
		}
	}
}

func reasonOf(err error) string { return truncate(err.Error(), reasonLimit) }

func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339) }

// resolveSession 은 market-calendar 의 오늘 정규장을 읽음. 없으면(휴장) 거부.
func (p *probe) resolveSession(ctx context.Context, market string) (session, error) {
	calendar, err := withRateLimitRetry(ctx, p, func(ctx context.Context) (official.MarketCalendarResponse, error) {
		return p.src.TypedMarketCalendar(ctx, market, "")
	})
	if err != nil {
		return session{}, fmt.Errorf("%s market calendar: %w", market, err)
	}
	regular := calendar.Today.RegularMarket
	if regular == nil {
		return session{}, fmt.Errorf("%s has no regular session today (%s) per the market calendar", market, calendar.Today.Date)
	}
	if !regular.EndTime.After(regular.StartTime) {
		return session{}, fmt.Errorf("%s regular session %s..%s is not a forward interval", market, regular.StartTime, regular.EndTime)
	}
	return session{Open: regular.StartTime, Close: regular.EndTime}, nil
}

// runMarket 은 세션 안에서 간격마다 한 사이클씩 돎. 판단 시각 + horizon 이 장 마감을 넘으면
// 더 판단하지 않음 — 라벨이 장 밖 가격을 재지 않게 하기 위함(설계상 표본 창).
func (p *probe) runMarket(ctx context.Context, market string) error {
	sess, err := p.resolveSession(ctx, market)
	if err != nil {
		return err
	}
	now := p.now()
	if !now.Before(sess.Close) {
		return fmt.Errorf("%s regular session already closed at %s", market, stamp(sess.Close))
	}
	if now.Before(sess.Open) {
		if !p.cfg.WaitOpen {
			return fmt.Errorf("%s regular session opens at %s; pass --wait-open to wait for it", market, stamp(sess.Open))
		}
		p.logf("%s: waiting for the open at %s", market, stamp(sess.Open))
		if err := p.sleep(ctx, sess.Open.Sub(now)); err != nil {
			return err
		}
	}
	cycles := 0
	for {
		started := p.now()
		if started.Add(p.cfg.Horizon).After(sess.Close) {
			p.logf("%s: sampling window closed (%d cycles); %s + horizon passes the close %s",
				market, cycles, stamp(started), stamp(sess.Close))
			return nil
		}
		if err := p.cycle(ctx, market, sess, started); err != nil {
			return err
		}
		cycles++
		if err := p.sleep(ctx, started.Add(p.cfg.Interval).Sub(p.now())); err != nil {
			return err
		}
	}
}

// cycle 은 rankings → stocks → prices → orderbook(종목별) → 판단 → 원장. 원장 쓰기 실패만
// 오류로 올림(측정 자체가 무너진 것). 읽기·판단 실패는 결손 행으로 남기고 계속함.
func (p *probe) cycle(ctx context.Context, market string, sess session, at time.Time) error {
	gap := func(stage, symbol string, err error) error {
		p.logf("%s %s %s: %s", market, stage, symbol, reasonOf(err))
		return p.ledger.gap(gapRow{At: stamp(at), Market: market, Stage: stage, Symbol: symbol, Reason: reasonOf(err)})
	}

	ranking, err := withRateLimitRetry(ctx, p, func(ctx context.Context) (domain.Ranking, error) {
		return p.src.Rankings(ctx, rankingType, market, rankingDuration, false, p.cfg.TopK)
	})
	if err != nil {
		return gap("rankings", "", err)
	}
	items := ranking.Items
	if len(items) > p.cfg.TopK {
		items = items[:p.cfg.TopK]
	}
	symbols := make([]string, 0, len(items))
	for _, item := range items {
		symbols = append(symbols, item.Symbol)
	}
	if len(symbols) == 0 {
		return gap("rankings", "", errors.New("ranking returned no symbols"))
	}

	// 메타 결손은 판단을 막지 않음 — 이름 없이도 판단은 가능하고, 결손은 행과 state 양쪽에 남음.
	var cycleGaps []string
	meta := map[string]domain.Quote{}
	if stocks, err := withRateLimitRetry(ctx, p, func(ctx context.Context) ([]domain.Quote, error) {
		return p.src.Stocks(ctx, symbols)
	}); err != nil {
		cycleGaps = append(cycleGaps, "stocks_meta: "+reasonOf(err))
	} else {
		for _, quote := range stocks {
			meta[quote.Symbol] = quote
		}
	}

	// 기준가가 없으면 라벨을 만들 수 없으므로 판단(비용)을 쓰지 않고 결손으로 남김.
	prices, err := withRateLimitRetry(ctx, p, func(ctx context.Context) ([]domain.Quote, error) {
		return p.src.Prices(ctx, symbols)
	})
	if err != nil {
		return gap("prices", "", err)
	}
	last := map[string]float64{}
	for _, quote := range prices {
		last[quote.Symbol] = quote.Last
	}

	for _, item := range items {
		base := last[item.Symbol]
		if base <= 0 {
			if err := gap("prices", item.Symbol, errors.New("no positive last price for the symbol")); err != nil {
				return err
			}
			continue
		}
		gaps := append([]string{}, cycleGaps...)
		var metaPtr *domain.Quote
		if quote, found := meta[item.Symbol]; found {
			metaPtr = &quote
		} else if len(cycleGaps) == 0 {
			gaps = append(gaps, "stocks_meta: symbol missing from the response")
		}
		var bookPtr *official.StrictTopOfBook
		book, err := withRateLimitRetry(ctx, p, func(ctx context.Context) (official.StrictTopOfBook, error) {
			return p.src.StrictOrderbookTop(ctx, market, item.Symbol)
		})
		if err != nil {
			gaps = append(gaps, "top_of_book: "+reasonOf(err))
		} else {
			bookPtr = &book
		}

		state := assembleState(marketInputs{
			Market: market, ObservedAt: at, Horizon: p.cfg.Horizon, Session: sess,
			RankingType: rankingType, Rank: item, Meta: metaPtr, LastPrice: base, Book: bookPtr, Gaps: gaps,
		})
		digest, err := p.ledger.storeState(state)
		if err != nil {
			return err
		}
		result, err := p.judge.judge(ctx, state, p.questions)
		if err != nil {
			if err := gap("judgment", item.Symbol, err); err != nil {
				return err
			}
			continue
		}
		due := at.Add(p.cfg.Horizon)
		pending := pendingLabel{Market: market, Symbol: item.Symbol, BasePrice: base, Due: due}
		for _, q := range p.questions {
			id := fmt.Sprintf("%s-%s-%s-%s", at.UTC().Format("20060102T150405Z"), market, item.Symbol, q.ID)
			if err := p.ledger.judgment(judgmentRow{
				ID: id, At: stamp(at), Market: market, Symbol: item.Symbol,
				QuestionID: q.ID, QuestionRev: q.Rev, QuestionDigest: q.digest(),
				ModelRequested: p.judge.model, ModelAnswered: result.Model,
				StateDigest: digest, P: result.P[q.ID], BasePrice: base, Currency: state.Currency,
				HorizonMinutes: state.HorizonMinutes, LabelDueAt: stamp(due),
				InputTokens: result.InputTokens, CollectionGaps: state.MissingInputs,
			}); err != nil {
				return err
			}
			pending.JudgmentIDs = append(pending.JudgmentIDs, id)
		}
		p.addPending(pending)
	}
	return nil
}

func (p *probe) addPending(item pendingLabel) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pending = append(p.pending, item)
}

func (p *probe) pendingCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pending)
}

// restorePending 은 재시작 시 원장에서 라벨 없는 판단을 다시 대기열에 올림.
// 기한이 한참 지난 것은 다음 tick 에서 "missed_due_window" 결손 라벨이 됨.
func (p *probe) restorePending(contents ledgerContents) {
	groups := map[string]*pendingLabel{}
	var order []string
	for _, row := range contents.Judgments {
		if _, labeled := contents.Labels[row.ID]; labeled {
			continue
		}
		due, err := time.Parse(time.RFC3339, row.LabelDueAt)
		if err != nil {
			continue // 기한을 못 읽는 행은 보고서에서 "미라벨"로 드러남
		}
		key := row.Market + "|" + row.Symbol + "|" + row.LabelDueAt
		if groups[key] == nil {
			groups[key] = &pendingLabel{Market: row.Market, Symbol: row.Symbol, BasePrice: row.BasePrice, Due: due}
			order = append(order, key)
		}
		groups[key].JudgmentIDs = append(groups[key].JudgmentIDs, row.ID)
	}
	for _, key := range order {
		p.addPending(*groups[key])
	}
}

// labelTick 은 기한이 된 항목을 시장별로 모아 /prices 한 번으로 라벨링함.
func (p *probe) labelTick(ctx context.Context) error {
	now := p.now()
	p.mu.Lock()
	var due, waiting []pendingLabel
	for _, item := range p.pending {
		if !item.Due.After(now) {
			due = append(due, item)
		} else {
			waiting = append(waiting, item)
		}
	}
	p.pending = waiting
	p.mu.Unlock()
	if len(due) == 0 {
		return nil
	}

	byMarket := map[string][]pendingLabel{}
	for _, item := range due {
		byMarket[item.Market] = append(byMarket[item.Market], item)
	}
	markets := make([]string, 0, len(byMarket))
	for market := range byMarket {
		markets = append(markets, market)
	}
	sort.Strings(markets)

	for _, market := range markets {
		items := byMarket[market]
		var onTime []pendingLabel
		for _, item := range items {
			if now.Sub(item.Due) > labelTolerance {
				if err := p.writeLabels(item, now, nil, fmt.Sprintf("missed_due_window: labeled %s after due, tolerance %s",
					now.Sub(item.Due).Round(time.Second), labelTolerance)); err != nil {
					return err
				}
				continue
			}
			onTime = append(onTime, item)
		}
		if len(onTime) == 0 {
			continue
		}
		symbolSet := map[string]bool{}
		var symbols []string
		for _, item := range onTime {
			if !symbolSet[item.Symbol] {
				symbolSet[item.Symbol] = true
				symbols = append(symbols, item.Symbol)
			}
		}
		prices, err := withRateLimitRetry(ctx, p, func(ctx context.Context) ([]domain.Quote, error) {
			return p.src.Prices(ctx, symbols)
		})
		readAt := p.now()
		for _, item := range onTime {
			if err != nil {
				if werr := p.writeLabels(item, readAt, nil, "price_read_failed: "+reasonOf(err)); werr != nil {
					return werr
				}
				continue
			}
			var realized float64
			for _, quote := range prices {
				if quote.Symbol == item.Symbol {
					realized = quote.Last
				}
			}
			switch {
			case realized <= 0:
				err := p.writeLabels(item, readAt, nil, "realized_price_missing: no positive last price for the symbol")
				if err != nil {
					return err
				}
			case item.BasePrice <= 0:
				if err := p.writeLabels(item, readAt, nil, "base_price_invalid"); err != nil {
					return err
				}
			default:
				if err := p.writeLabels(item, readAt, &realized, ""); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *probe) writeLabels(item pendingLabel, at time.Time, realized *float64, gapReason string) error {
	for _, id := range item.JudgmentIDs {
		row := labelRow{JudgmentID: id, At: stamp(at), DueAt: stamp(item.Due), LabelGap: gapReason}
		if realized != nil {
			price := *realized
			ret := round((price-item.BasePrice)/item.BasePrice, 8)
			row.RealizedPrice, row.RealizedReturn = &price, &ret
		}
		if err := p.ledger.label(row); err != nil {
			return err
		}
	}
	return nil
}

// run 은 시장별 수집을 병렬로 돌리고, 같은 고루틴에서 라벨 tick 을 돌린 뒤
// 수집이 끝나고 대기열이 빌 때까지 기다림.
func (p *probe) run(ctx context.Context, markets []string) error {
	var wg sync.WaitGroup
	errs := make([]error, len(markets))
	done := make(chan struct{})
	for index, market := range markets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[index] = p.runMarket(ctx, market)
		}()
	}
	go func() { wg.Wait(); close(done) }()

	collecting := true
	for {
		if err := p.labelTick(ctx); err != nil {
			return err
		}
		if collecting {
			select {
			case <-done:
				collecting = false
			default:
			}
		}
		if !collecting && p.pendingCount() == 0 {
			return errors.Join(errs...)
		}
		if err := p.sleep(ctx, p.cfg.LabelTick); err != nil {
			return err
		}
	}
}
