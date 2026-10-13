package main

// collect.go 는 수집 루프·세션 경계·rate limit 백오프·라벨러를 담당함.
//
// 브로커에는 읽기(GET)만 보냄: marketSource 인터페이스가 이 도구가 부를 수 있는 브로커 표면
// 전부이고, 그 메서드 여섯은 모두 공개 시장 데이터 GET 임(환율 포함 — 1.5 표본 선택의 상한 환산).
// 주문·계좌 메서드는 인터페이스에 없으므로 수집 코드에서 도달 불가(static_test.go 가 메서드 집합을 고정함).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
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
	ExchangeRate(ctx context.Context, base, quote string) (domain.ExchangeRate, error)
}

const (
	rankingType     = "MARKET_TRADING_AMOUNT"
	rankingDuration = "realtime"
	// 라벨이 기한보다 이만큼 넘게 늦으면 실현가를 쓰지 않고 결손으로 적음(다른 horizon 을 재게 되므로).
	labelTolerance = 5 * time.Minute
	reasonLimit    = 160

	// sessionSource* 는 세션창을 어떻게 정했는지 — state·원장에 남겨 보고서가 수동 우회가 섞인
	// 로트를 가려낼 수 있게 함(달력 오류 대응으로 넣은 --session 수단).
	sessionSourceCalendar = "calendar"
	sessionSourceManual   = "manual-override"
)

// config 는 실행 매개변수. 기본값은 main.go 플래그에 있음.
type config struct {
	Interval time.Duration
	// TopK 는 사이클당 결선 수 — 풀 판단 뒤 J2 악재 제외·J1 내림차순 상위 K(screen.go).
	TopK int
	// Pool 은 사이클당 판단하는 풀 크기(상위 100 중 상한 이하 후보에서 무작위 추출). TopK ≤ Pool ≤ 100.
	Pool        int
	Horizon     time.Duration
	WaitOpen    bool
	LabelTick   time.Duration
	ReadGap     time.Duration   // 브로커 호출 사이 간격
	RetryDelays []time.Duration // 429 재시도 지연(지수) — 길이가 재시도 횟수
	// ManualWindow 는 --session 수동 세션창 우회. nil 이면 달력을 씀(기본 동작 불변).
	ManualWindow *manualWindow
	// CapSource 는 표본 선택의 주문 상한 출처(운영 = config 파일 읽기). nil 이면 선택을 거부함.
	CapSource func(context.Context) (orderCap, error)
}

// session 은 정규장 하나의 경계. Source 는 창을 어떻게 정했는지(달력 vs 수동 우회)를 적어
// state·원장으로 흘려보냄.
type session struct {
	Open   time.Time
	Close  time.Time
	Source string
}

// manualWindow 는 --session 으로 사람이 공급한 세션창(시장 현지 시각). 달력이 틀렸을 때만 씀.
// 파싱·역전창 검사는 시작 전 parseManualWindow 에서 끝나므로 여기 담긴 값은 이미 유효함.
type manualWindow struct {
	StartHour, StartMin int
	EndHour, EndMin     int
	Raw                 string // "HH:MM-HH:MM" 원문 — 원장 우회 행에 그대로 적음
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
	seed      func() uint64 // 사이클별 표본 선택 시드(시험이 고정함)

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
		now: time.Now, sleep: sleepContext, seed: newSelectionSeed,
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

// resolveSession 은 세션창을 정함. 기본은 market-calendar 의 오늘 정규장을 읽음(없으면 휴장으로 거부).
// --session 수동 우회가 있으면 달력을 전혀 호출하지 않고 그 창을 쓰며, 우회 사실을 원장에 남김
// (달력이 틀렸을 때 사람이 명시적으로 세션창을 공급하는 수단 — 2026-10-12 KR 달력 오응답 대응).
func (p *probe) resolveSession(ctx context.Context, market string) (session, error) {
	if p.cfg.ManualWindow != nil {
		return p.manualSession(market)
	}
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
	return session{Open: regular.StartTime, Close: regular.EndTime, Source: sessionSourceCalendar}, nil
}

// manualSession 은 수동 창을 시장 현지 "오늘"(p.now() 의 시장 현지 날짜)에 붙여 세션으로 만들고,
// 달력 우회를 원장에 1행 남김(침묵 생략 금지). 달력은 한 번도 호출하지 않음.
func (p *probe) manualSession(market string) (session, error) {
	loc, err := marketLocation(market)
	if err != nil {
		return session{}, err
	}
	w := p.cfg.ManualWindow
	nowLocal := p.now().In(loc)
	y, m, d := nowLocal.Date()
	openAt := time.Date(y, m, d, w.StartHour, w.StartMin, 0, 0, loc)
	closeAt := time.Date(y, m, d, w.EndHour, w.EndMin, 0, 0, loc)
	// 명시적 우회를 원장에 1행 남김 — 보고서가 gap 단계 "session" 으로 override 혼입을 봄.
	if err := p.ledger.gap(gapRow{
		At: stamp(p.now()), Market: market, Stage: "session",
		Reason: fmt.Sprintf("calendar bypassed by --session, window=%s, reason=operator", w.Raw),
	}); err != nil {
		return session{}, err
	}
	return session{Open: openAt, Close: closeAt, Source: sessionSourceManual}, nil
}

// marketLocation 은 수동 창 해석에 쓰는 시장 현지 시간대. US 는 DST(EDT/EST)를 그대로 반영함.
func marketLocation(market string) (*time.Location, error) {
	name := map[string]string{"KR": "Asia/Seoul", "US": "America/New_York"}[market]
	if name == "" {
		return nil, fmt.Errorf("no timezone known for market %q", market)
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("loading %s timezone %q: %w", market, name, err)
	}
	return loc, nil
}

// parseManualWindow 는 "HH:MM-HH:MM"(시장 현지 시각)를 검증해 manualWindow 로 만듦. 시작 전에 부름 —
// 파싱 실패·역전창(start≥end)은 여기서 거부되어 브로커를 한 번도 부르지 않음.
func parseManualWindow(raw string) (*manualWindow, error) {
	start, end, ok := strings.Cut(raw, "-")
	if !ok {
		return nil, fmt.Errorf(`--session must be "HH:MM-HH:MM", got %q`, raw)
	}
	sh, sm, err := parseHHMM(start)
	if err != nil {
		return nil, fmt.Errorf("--session start: %w", err)
	}
	eh, em, err := parseHHMM(end)
	if err != nil {
		return nil, fmt.Errorf("--session end: %w", err)
	}
	if sh*60+sm >= eh*60+em {
		return nil, fmt.Errorf("--session window %q is not a forward interval (start must precede end)", raw)
	}
	return &manualWindow{StartHour: sh, StartMin: sm, EndHour: eh, EndMin: em, Raw: raw}, nil
}

// parseHHMM 은 "15:04" 레이아웃으로 시·분을 뽑음. 레이아웃이 시(0..23)·분(0..59) 범위를 함께 검증함.
func parseHHMM(value string) (int, int, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(value))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid HH:MM %q", strings.TrimSpace(value))
	}
	return t.Hour(), t.Minute(), nil
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

// cycle 은 상한·환율 → rankings(상위 100) → 상한 필터·무작위 풀 → stocks → prices →
// orderbook(종목별) → 풀 전체 판단 → 선별(결선 표시) → 원장. 원장 쓰기 실패만 오류로 올림(측정 자체가
// 무너진 것). 읽기·판단 실패는 결손 행으로 남기고 계속함.
//
// 판단 행은 풀 전체를 판단한 뒤에 한꺼번에 씀 — 결선 여부와 사이클의 선별 수(selection)가 판단이 다
// 끝나야 정해지기 때문. 그래서 사이클 도중 프로세스가 죽으면 그 사이클의 판단 행은 남지 않음
// (state 파일과 판단 gap 행은 이미 쓰였을 수 있음 — 판단 행 없는 state 파일은 측정에 무해함).
func (p *probe) cycle(ctx context.Context, market string, sess session, at time.Time) error {
	gap := func(stage, symbol string, err error) error {
		p.logf("%s %s %s: %s", market, stage, symbol, reasonOf(err))
		return p.ledger.gap(gapRow{At: stamp(at), Market: market, Stage: stage, Symbol: symbol, Reason: reasonOf(err)})
	}

	// 상한·환율을 못 읽으면 옛 규칙으로 폴백하지 않고 이 사이클을 판단 없이 gap 으로 남김.
	inputs, err := p.resolveSelectionInputs(ctx, market)
	if err != nil {
		return gap("selection", "", err)
	}
	ranking, err := withRateLimitRetry(ctx, p, func(ctx context.Context) (domain.Ranking, error) {
		return p.src.Rankings(ctx, rankingType, market, rankingDuration, false, rankingFetch)
	})
	if err != nil {
		return gap("rankings", "", err)
	}
	received := ranking.Items
	if len(received) > rankingFetch {
		received = received[:rankingFetch]
	}
	if len(received) == 0 {
		return gap("rankings", "", errors.New("ranking returned no symbols"))
	}
	candidates, overCap, unknownPrice := filterByCap(received, inputs)
	selection := &selectionRecord{
		Rule: selectionRuleFor(p.cfg.TopK, p.cfg.Pool), Seed: p.seed(), CapValue: inputs.cap.Value, CapCurrency: inputs.cap.Currency,
		PriceCurrency: inputs.priceCurrency, PriceSource: selectionPriceSource,
		RankingRequested: rankingFetch, RankingReceived: len(received),
		ExcludedOverCap: overCap, ExcludedUnknownPrice: unknownPrice, Candidates: len(candidates),
		Pool: p.cfg.Pool, FinalistsRequested: p.cfg.TopK,
	}
	if inputs.fxPair != "" {
		selection.FXRate, selection.FXPair = inputs.fxRate, inputs.fxPair
	}
	if len(candidates) == 0 {
		return gap("selection", "", fmt.Errorf("no ranking row within the order cap %v %s (received %d, over cap %d, unknown price %d)",
			inputs.cap.Value, inputs.cap.Currency, len(received), overCap, unknownPrice))
	}
	items := pickSample(candidates, p.cfg.Pool, selection.Seed)
	selection.Selected = len(items)
	symbols := make([]string, 0, len(items))
	for _, item := range items {
		symbols = append(symbols, item.Symbol)
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

	// judgedSymbol 은 판단에 성공한 풀 종목 하나 — 선별이 끝날 때까지 메모리에 둠(랭킹 순서 유지).
	type judgedSymbol struct {
		symbol string
		base   float64
		digest string
		state  PublicState
		result judgeResult
	}
	var judged []judgedSymbol
	var screenInputs []screenInput
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
		// 선별 입력 p 가 응답에 없으면(질문 집합이 바뀐 경우) 0 으로 읽지 않고 측정을 멈춤.
		j1, haveJ1 := result.P[screenUpQuestion]
		j2, haveJ2 := result.P[screenRedFlagQuestion]
		if !haveJ1 || !haveJ2 {
			return fmt.Errorf("screen inputs %q/%q missing from the judgment of %s", screenUpQuestion, screenRedFlagQuestion, item.Symbol)
		}
		judged = append(judged, judgedSymbol{symbol: item.Symbol, base: base, digest: digest, state: state, result: result})
		screenInputs = append(screenInputs, screenInput{J1: j1, J2: j2})
	}

	// 선별은 판단이 다 끝난 뒤 한 번 — 사이클의 selection 수치를 행을 쓰기 전에 확정함.
	screens := screenFinalists(screenInputs, p.cfg.TopK)
	selection.Judged = len(judged)
	for _, screen := range screens {
		if screen.RedFlagExcluded {
			selection.RedFlagExcluded++
		}
		if screen.Finalist {
			selection.Finalists++
		}
	}

	// 결선·비결선 모두 행을 쓰고 라벨 대기에 올림 — 비결선은 보고서의 대조군.
	due := at.Add(p.cfg.Horizon)
	for index, entry := range judged {
		screen := screens[index]
		pending := pendingLabel{Market: market, Symbol: entry.symbol, BasePrice: entry.base, Due: due}
		for _, q := range p.questions {
			id := fmt.Sprintf("%s-%s-%s-%s", at.UTC().Format("20060102T150405Z"), market, entry.symbol, q.ID)
			if err := p.ledger.judgment(judgmentRow{
				ID: id, At: stamp(at), Market: market, Symbol: entry.symbol,
				QuestionID: q.ID, QuestionRev: q.Rev, QuestionDigest: q.digest(),
				ModelRequested: p.judge.model, ModelAnswered: entry.result.Model,
				StateDigest: entry.digest, P: entry.result.P[q.ID], BasePrice: entry.base, Currency: entry.state.Currency,
				HorizonMinutes: entry.state.HorizonMinutes, LabelDueAt: stamp(due),
				InputTokens: entry.result.InputTokens, CollectionGaps: entry.state.MissingInputs,
				Selection: selection, Screen: &screen,
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
