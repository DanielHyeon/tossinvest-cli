package main

// shadow.go 는 한 세션의 그림자 측정 루프 — 분마다 종목별로 봉을 적재하고 평가해 원장에 적음(design §2 · §2.1).

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutsignal"
	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialbars"
	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"
)

const (
	rankingType     = "MARKET_TRADING_AMOUNT"
	rankingDuration = "realtime"
	rankingFetch    = 100

	sessionSourceCalendar = "calendar"
	sessionSourceManual   = "manual-override"

	reasonLimit = 200
)

var krSymbol = regexp.MustCompile(`^[0-9]{6}$`)

// probeConfig 는 실행 매개변수(기본값은 defaultProbeConfig · main.go 플래그).
type probeConfig struct {
	Top           int
	Params        shadowParams
	ParamsSource  string
	ReadGap       time.Duration
	RetryDelays   []time.Duration
	QuoteSuspend  time.Duration
	ArmedAttempts int
	ArmedWindow   time.Duration
	GapPersist    time.Duration
	LabelGrace    time.Duration
	BudgetEvery   time.Duration
	PollLookback  time.Duration
}

func defaultProbeConfig() probeConfig {
	return probeConfig{Top: 30, ReadGap: 250 * time.Millisecond,
		RetryDelays:  []time.Duration{5 * time.Second, 10 * time.Second, 20 * time.Second},
		QuoteSuspend: 5 * time.Minute, ArmedAttempts: 5, ArmedWindow: 5 * time.Minute, GapPersist: 3 * time.Minute,
		LabelGrace: 30 * time.Minute, BudgetEvery: 10 * time.Minute, PollLookback: 5 * time.Minute}
}

// recordingStore 는 evidence 저장소에 마지막으로 넘긴 봉투를 기억함 — PollQuoteL1 이 호가 값을 돌려주지 않으므로
// 방금 적재한 호가 본문을 여기서 다시 읽음(같은 고루틴에서 순서대로 부름).
type recordingStore struct {
	*strategyevidence.Store
	mu   sync.Mutex
	last strategyevidence.Envelope
	has  bool
}

func (r *recordingStore) Append(ctx context.Context, envelope strategyevidence.Envelope) (strategyevidence.AppendResult, error) {
	r.mu.Lock()
	r.last, r.has = envelope, true
	r.mu.Unlock()
	return r.Store.Append(ctx, envelope)
}

func (r *recordingStore) lastEnvelope() (strategyevidence.Envelope, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.last, r.has
}

// symbolState 는 종목 하나의 세션 상태 — 재시작 때 원장에서 다시 세움.
type symbolState struct {
	symbol                 string
	quote                  *quoteSnap
	lastPhase, lastRefusal string
	done                   string
	attempts               int
	armedSince             time.Time
	proposal               *ledgerRow
	labeled                bool
	newestOpenMS           uint64
	gapAt, gapSince        time.Time
	gapLogged              bool
	gapped                 bool
}

type probe struct {
	paced         *pacedSource
	store         *recordingStore
	ledger        *ledger
	ledgerPath    string
	cfg           probeConfig
	now           func() time.Time
	sess          sessionInfo
	calendar      scheduler.CalendarSnapshot
	sessionSource string
	lot           string
	symbols       []*symbolState
	gapsSeen      map[string]bool
	counts        map[string]int
	lastBudget    time.Time
	collect       time.Duration
	barsSkipped   bool // 이번 분에 봉 그룹 재시도를 다 썼음 — 나머지 종목의 봉 폴을 건너뜀(§2.1-10)
}

func newProbe(src marketSource, store *strategyevidence.Store, led *ledger, cfg probeConfig, now func() time.Time) *probe {
	sleep := func(ctx context.Context, d time.Duration) error {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
	return &probe{paced: newPacedSource(src, cfg.ReadGap, cfg.RetryDelays, cfg.QuoteSuspend, now, sleep),
		store: &recordingStore{Store: store}, ledger: led, ledgerPath: led.f.Name(), cfg: cfg, now: now,
		gapsSeen: map[string]bool{}, counts: map[string]int{}}
}

func (p *probe) stamp() string { return p.now().UTC().Format(time.RFC3339) }

func (p *probe) write(row ledgerRow) error {
	row.At, row.Session, row.Lot = p.stamp(), p.sess.ID, p.lot
	return p.ledger.write(row)
}

// gap 은 결손 행 — 종목 · 단계 · 분마다 한 행(spec 「그 종목 · 그 분의 결손 행」). 같은 분의 같은 결손만 접고, 계수는
// 10 분마다 budget 행에 실려 중간에 죽어도 남음.
func (p *probe) gap(symbol, stage string, err error) error {
	reason := err.Error()
	if len(reason) > reasonLimit {
		reason = reason[:reasonLimit]
	}
	p.counts["gap:"+stage]++
	minute := p.now().UTC().Truncate(time.Minute).Format(time.RFC3339)
	key := symbol + "\x00" + stage + "\x00" + minute
	if p.gapsSeen[key] {
		return nil
	}
	p.gapsSeen[key] = true
	return p.write(ledgerRow{Kind: kindGap, Symbol: symbol, Stage: stage, Reason: reason, Minute: minute})
}

func firstWords(s string) string {
	if fields := strings.Fields(s); len(fields) > 3 {
		return strings.Join(fields[:3], " ")
	}
	return s
}

// startSession 은 세션 신원을 정하고, 종목 집합을 랭킹에서 고르거나(처음) 원장에서 되살리고(재시작), 끝난 setup · 라벨 대기를 복원하고,
// 끝나지 않은 종목마다 구조용 호가 봉인 하나를 읽음. 같은 세션의 다른 판(lot) 행이 원장에 있으면 섞지 않고 거부함.
func (p *probe) startSession(ctx context.Context, cal scheduler.CalendarSnapshot, source string) error {
	if cal.Today.Regular == nil {
		return fmt.Errorf("the calendar carries no regular session for %s", cal.Today.Date)
	}
	p.calendar, p.sessionSource = cal, source
	p.sess = sessionInfo{ID: "KRX:" + cal.Today.Date, CalendarVersion: cal.Version, Open: cal.Today.Regular.Open, Close: cal.Today.Regular.Close}
	policy := effectivePolicyFor(p.cfg, source)
	policyDigest := policy.digest()
	sessionIdentity := sessionIdentityFor(p.sess)
	p.lot = lotForPolicy(policyDigest)
	p.lastBudget = p.now()

	all, err := readLedger(p.ledgerPath)
	if err != nil {
		return err
	}
	var rows []ledgerRow
	var universe []string
	for _, row := range all {
		if row.Session != p.sess.ID || row.Kind == kindGap && row.Lot == "" {
			continue
		}
		if row.Lot != p.lot {
			return fmt.Errorf("the ledger already holds %s rows of lot %q; this run is lot %q — use another --out instead of mixing", p.sess.ID, row.Lot, p.lot)
		}
		if row.Kind == kindUniverse {
			if row.Policy == nil || row.PolicyDigest != policyDigest || row.Policy.digest() != policyDigest {
				return fmt.Errorf("the ledger universe policy for %s is missing or does not match lot %q", p.sess.ID, p.lot)
			}
			if row.SessionIdentity == nil || *row.SessionIdentity != sessionIdentity {
				return fmt.Errorf("the ledger universe session identity for %s is missing or does not match this restart", p.sess.ID)
			}
			universe = row.Symbols
		}
		rows = append(rows, row)
	}
	if universe == nil {
		universe, err = p.selectUniverse(ctx)
		if err != nil {
			// 랭킹 포기도 결손 행으로 남김(spec) — 그다음 시작을 거부.
			if gapErr := p.gap("", "ranking", err); gapErr != nil {
				return gapErr
			}
			return err
		}
		params := p.cfg.Params
		if err := p.write(ledgerRow{Kind: kindUniverse, Symbols: universe, SessionSource: source, Params: &params,
			ParamsSource: p.cfg.ParamsSource, Policy: &policy, PolicyDigest: policyDigest, SessionIdentity: &sessionIdentity}); err != nil {
			return err
		}
	}
	states := map[string]*symbolState{}
	p.symbols = p.symbols[:0]
	for _, symbol := range universe {
		state := &symbolState{symbol: symbol}
		states[symbol] = state
		p.symbols = append(p.symbols, state)
	}
	labeled := map[string]bool{}
	for _, row := range rows {
		if row.Kind == kindLabel {
			labeled[row.SetupID] = true
		}
	}
	for i := range rows {
		row := rows[i]
		state := states[row.Symbol]
		if row.Kind != kindSignal || state == nil {
			continue
		}
		state.lastPhase, state.lastRefusal = row.Phase, row.Refusal
		if row.Attempt > state.attempts {
			state.attempts = row.Attempt
		}
		if row.ArmedSince != "" && state.armedSince.IsZero() {
			state.armedSince, _ = time.Parse(time.RFC3339, row.ArmedSince)
		}
		if row.Done != "" {
			state.done = row.Done
		}
		if row.Done == doneProposed {
			copied := row
			state.proposal, state.labeled = &copied, labeled[row.SetupID]
		}
	}
	for _, state := range p.symbols {
		if state.done == "" {
			if quote, err := p.fetchQuote(ctx, state.symbol); err != nil {
				if err := p.gap(state.symbol, "quote", err); err != nil {
					return err
				}
			} else {
				state.quote = &quote
			}
		}
	}
	return nil
}

// selectUniverse 는 D8 — 거래대금 랭킹 상위 100 중 KR 6 자리 · 1 주 가격이 명목 상한 이하인 종목을 순위대로 Top 개.
func (p *probe) selectUniverse(ctx context.Context) ([]string, error) {
	ranking, err := p.paced.Rankings(ctx, rankingType, "KR", rankingDuration, false, rankingFetch)
	if err != nil {
		return nil, fmt.Errorf("reading the ranking: %w", err)
	}
	var out []string
	for _, item := range ranking.Items {
		if len(out) >= p.cfg.Top {
			break
		}
		price := item.LastPrice
		if !krSymbol.MatchString(item.Symbol) || item.Currency != "KRW" || math.IsNaN(price) || price <= 0 ||
			price > float64(p.cfg.Params.NotionalCapKRW) {
			continue
		}
		out = append(out, item.Symbol)
	}
	if len(out) == 0 {
		return nil, errors.New("no ranked symbol fits the order cap")
	}
	return out, nil
}

// fetchQuote 는 호가 한 줄(GET 2)을 도구 evidence 에 봉인하고 그 정수 값을 돌려줌.
func (p *probe) fetchQuote(ctx context.Context, symbol string) (quoteSnap, error) {
	result, err := officialbars.PollQuoteL1(ctx, p.paced, p.store, officialbars.QuotePollInput{Market: marketclock.MarketKR, Symbol: symbol})
	if err != nil {
		return quoteSnap{}, err
	}
	if result.Outcome == officialbars.QuoteOutcomeConflict {
		return quoteSnap{}, errors.New("QUOTE_CONFLICT: another body was stored for the same instant")
	}
	envelope, ok := p.store.lastEnvelope()
	if !ok {
		return quoteSnap{}, errors.New("the quote envelope was not observed")
	}
	payload, err := strategyevidence.DecodeQuoteL1Payload(envelope.CanonicalPayload())
	if err != nil {
		return quoteSnap{}, err
	}
	// KRW 의 price_scale 은 소수 자리 수 0(strategyevidence marketMoney) — 원 단위 정수.
	if payload.Symbol != symbol || payload.Currency != "KRW" || payload.PriceScale != 0 {
		return quoteSnap{}, fmt.Errorf("the stored quote is %s %s at scale %d, not a whole-won KRW quote for %s",
			payload.Symbol, payload.Currency, payload.PriceScale, symbol)
	}
	return quoteSnap{Bid: payload.BidMinor, Ask: payload.AskMinor, Last: payload.LastMinor, ObservedMS: payload.SourceObservedAtMS,
		ReceivedMS: payload.ReceivedAtMS, Currency: payload.Currency, Digest: payload.SourceResponseDigest}, nil
}

// step 은 분 하나 — 종목마다 봉을 적재 · 평가 · 라벨.
func (p *probe) step(ctx context.Context) error {
	started := p.now()
	p.barsSkipped = false
	for _, state := range p.symbols {
		if err := ctx.Err(); err != nil {
			return err
		}
		if state.done != "" && (state.proposal == nil || state.labeled) {
			continue // 끝났고 라벨도 끝남 — 더 폴하지 않음
		}
		if !p.now().Before(p.sess.Close) && (state.proposal == nil || state.labeled) {
			continue // 마감 뒤에는 라벨 대기만 폴함(§2.1 리뷰 — 새 setup 평가 금지)
		}
		if err := p.stepSymbol(ctx, state); err != nil {
			return err
		}
	}
	p.collect += p.now().Sub(started)
	if p.now().Sub(p.lastBudget) >= p.cfg.BudgetEvery {
		return p.writeBudget()
	}
	return nil
}

func (p *probe) writeBudget() error {
	p.lastBudget = p.now()
	reads := p.paced.readsByGroup()
	seconds := p.collect.Seconds()
	p.collect = 0
	counts := make(map[string]int, len(p.counts))
	for k, v := range p.counts {
		counts[k] = v
	}
	return p.write(ledgerRow{Kind: kindBudget, Reads: reads, CollectSeconds: seconds, Counts: counts})
}

// loadSeries 는 at 시점까지 적재된 정규장 봉 — 컷오프는 폴 **뒤** 시각이어야 방금 적재한 봉이 보임(리뷰 P2-1).
func (p *probe) loadSeries(ctx context.Context, symbol string, at time.Time) (strategyevidence.BarSeries, error) {
	return p.store.SealBarSeries(ctx, strategyevidence.BarSeriesQuery{Market: "KR", Symbol: symbol, SessionID: p.sess.ID,
		IntervalMS: strategyevidence.ClosedBar1mIntervalMS, EvaluationAt: at, IngestionCutoff: at, RegularSessionOnly: true})
}

func (p *probe) stepSymbol(ctx context.Context, state *symbolState) error {
	if p.barsSkipped {
		return p.gap(state.symbol, "bars", errors.New("skipped: the bar group gave up on a rate limit earlier this minute"))
	}
	// 폴 하한 = 가장 새 저장 봉 − 5 분(§2.1-4) — 평소 1 페이지, 놓친 분은 5 분 안이면 소급 적재.
	lower := p.sess.Open
	if state.newestOpenMS > 0 {
		lower = time.UnixMilli(int64(state.newestOpenMS)).UTC().Add(-p.cfg.PollLookback)
		if lower.Before(p.sess.Open) {
			lower = p.sess.Open
		}
	}
	if _, err := officialbars.PollClosedBars(ctx, p.paced, p.store, officialbars.BarPollInput{Market: marketclock.MarketKR,
		Symbol: state.symbol, Calendar: p.calendar, PollAt: p.now(), LowerBound: lower, MaxPages: 2}); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, errRateLimitGaveUp) {
			p.barsSkipped = true
		}
		return p.gap(state.symbol, "bars", err)
	}
	afterPoll := p.now()
	series, err := p.loadSeries(ctx, state.symbol, afterPoll)
	if err != nil {
		return p.gap(state.symbol, "series", err)
	}
	if n := len(series.Bars); n > 0 {
		state.newestOpenMS = series.Bars[n-1].Payload.OpenAtMS
	}
	prefix, missing, gapped := contiguousPrefix(series, p.sess.Open)
	if err := p.trackGap(state, missing, gapped, afterPoll); err != nil {
		return err
	}
	if state.proposal != nil && !state.labeled {
		// 라벨은 연속성이 필요 없음 — 전체 묶음에서 제안 뒤에 연 봉만 봄(리뷰 P2-6).
		if err := p.label(state, series, false); err != nil {
			return err
		}
	}
	if state.done != "" || !afterPoll.Before(p.sess.Close) || len(prefix.Bars) < breakoutsignal.OpeningRangeBars {
		return nil
	}
	if state.quote == nil {
		quote, err := p.fetchQuote(ctx, state.symbol)
		if err != nil {
			return p.gap(state.symbol, "quote", err)
		}
		state.quote = &quote
	}
	ev, err := evaluateSetup(prefix, p.sess, state.symbol, p.cfg.Params, *state.quote, uint64(p.now().UnixMilli()))
	if err != nil {
		var refusal *breakoutsignal.Refusal
		if errors.As(err, &refusal) {
			return p.gap(state.symbol, "derive", err)
		}
		return p.gap(state.symbol, "evaluate", err)
	}
	if ev.Phase == "ARMED" || ev.Phase == "PROPOSED" {
		return p.armedAttempt(ctx, state, prefix, ev)
	}
	if ev.Phase == state.lastPhase && ev.Refusal == state.lastRefusal {
		return nil
	}
	row := signalRow(state.symbol, ev, *state.quote)
	switch ev.Phase {
	case "INVALIDATED":
		row.Done = doneInvalidated
	case "TIMED_OUT":
		row.Done = doneTimedOut
	}
	state.lastPhase, state.lastRefusal, state.done = ev.Phase, ev.Refusal, row.Done
	return p.write(row)
}

// armedAttempt 는 ARMED 시도 하나 — 새 호가로 다시 판정하고 시도마다 원장에 적음(§2.1-1).
//
// 창은 시간이다: ARMED 를 처음 본 순간부터 ArmedWindow(5 분) 안에서만 시도하고, 창이 지나면 시도 없이도 armed_expired 로
// 끝냄(호가가 정지 · 실패해 시도가 모자라도 묵은 진입가로 계속 시도하지 않게 — 리뷰 P2-2). 시도 수 상한(5)은 그 안의 분 수.
func (p *probe) armedAttempt(ctx context.Context, state *symbolState, prefix strategyevidence.BarSeries, stale evaluation) error {
	if state.armedSince.IsZero() {
		state.armedSince = p.now()
	}
	if p.now().Sub(state.armedSince) >= p.cfg.ArmedWindow {
		row := signalRow(state.symbol, stale, *state.quote)
		row.Done, row.Refusal, row.ArmedSince = doneArmedExpired, state.lastRefusal, state.armedSince.UTC().Format(time.RFC3339)
		state.done = doneArmedExpired
		return p.write(row)
	}
	fresh, err := p.fetchQuote(ctx, state.symbol)
	if err != nil {
		return p.gap(state.symbol, "armed_quote", err) // 시도로 세지 않음 — 창 안이면 다음 분에 다시
	}
	state.quote = &fresh
	// 평가 시각은 호가를 **읽은 뒤**의 지금 — 호가의 received_at 은 본문을 다 읽은 순간이라 그 앞 시각으로 평가하면
	// 평가기가 언제나 QUOTE_STALE 로 읽음(리뷰 P1-1).
	evaluatedMS := max(uint64(p.now().UnixMilli()), fresh.ReceivedMS)
	ev, err := evaluateSetup(prefix, p.sess, state.symbol, p.cfg.Params, fresh, evaluatedMS)
	if err != nil {
		return p.gap(state.symbol, "evaluate", err)
	}
	state.attempts++
	row := signalRow(state.symbol, ev, fresh)
	row.Attempt, row.ArmedSince = state.attempts, state.armedSince.UTC().Format(time.RFC3339)
	switch {
	case ev.Phase == "PROPOSED":
		row.Done = doneProposed
		// 라벨은 제안 시각 뒤에 **연** 봉부터 — 그 앞 봉은 진입 전 가격(리뷰 P1-2).
		row.LabelFromMS = uint64(time.UnixMilli(int64(evaluatedMS)).UTC().Truncate(time.Minute).Add(time.Minute).UnixMilli())
	case state.attempts >= p.cfg.ArmedAttempts:
		row.Done = doneArmedExpired
	}
	state.lastPhase, state.lastRefusal, state.done = ev.Phase, ev.Refusal, row.Done
	if row.Done == doneProposed {
		copied := row
		state.proposal = &copied
	}
	return p.write(row)
}

func (p *probe) trackGap(state *symbolState, missing time.Time, gapped bool, now time.Time) error {
	if !gapped {
		state.gapped, state.gapLogged = false, false
		return nil
	}
	if !state.gapped || !state.gapAt.Equal(missing) {
		state.gapped, state.gapAt, state.gapSince, state.gapLogged = true, missing, now, false
	}
	if !state.gapLogged && now.Sub(state.gapSince) >= p.cfg.GapPersist {
		state.gapLogged = true
		return p.gap(state.symbol, "bars", fmt.Errorf("%s: the minute opening at %s is still missing after %s",
			breakoutsignal.RefusalMinuteGap, missing.Format(time.RFC3339), p.cfg.GapPersist))
	}
	return nil
}

func (p *probe) label(state *symbolState, series strategyevidence.BarSeries, final bool) error {
	row := state.proposal
	plan := sizingPlan{Entry: row.Entry, Ask: row.Ask, Worst: row.Worst, Stop: row.Stop, RoundTrip: row.RoundTrip, Risk: row.Risk, Target: row.Target}
	result, ok := labelProposal(series, row.LabelFromMS, plan, final)
	if !ok {
		return nil
	}
	state.labeled = true
	return p.write(ledgerRow{Kind: kindLabel, Symbol: state.symbol, SetupID: row.SetupID, Outcome: result.Outcome,
		ExitMinor: result.ExitMinor, ExitOpenAt: time.UnixMilli(int64(result.ExitOpenMS)).UTC().Format(time.RFC3339), RPPM: result.RPPM})
}

// finish 는 세션 끝 — 남은 제안을 마감가로 라벨하고, 끝까지 메워지지 않은 빈 분 · 끝나지 않은 ARMED 를 퍼널에 세고, 요약 행을 씀.
func (p *probe) finish(ctx context.Context) error {
	now := p.now()
	for _, state := range p.symbols {
		series, err := p.loadSeries(ctx, state.symbol, now)
		if err != nil {
			if err := p.gap(state.symbol, "series", err); err != nil {
				return err
			}
			continue
		}
		_, missing, gapped := contiguousPrefix(series, p.sess.Open)
		if state.proposal != nil && !state.labeled {
			if err := p.label(state, series, true); err != nil {
				return err
			}
		}
		if state.done == "" && state.lastPhase == "ARMED" {
			p.counts["funnel_armed_unfinished"]++
		}
		if gapped && state.done == "" {
			p.counts["funnel_gap"]++
			if err := p.write(ledgerRow{Kind: kindGap, Symbol: state.symbol, Stage: "session_end", Reason: string(breakoutsignal.RefusalMinuteGap) +
				": the session ended with the minute opening at " + missing.Format(time.RFC3339) + " missing"}); err != nil {
				return err
			}
		}
	}
	if err := p.writeBudget(); err != nil {
		return err
	}
	return p.write(ledgerRow{Kind: kindSession, Counts: p.counts})
}

// run 은 세션 하나를 처음부터 끝까지 돎: 개장 30 초 뒤 시작, 매 분 5 초에 step, 마감 + 라벨 유예 뒤 finish.
func (p *probe) run(ctx context.Context, cal scheduler.CalendarSnapshot, source string) error {
	if cal.Today.Regular == nil {
		return fmt.Errorf("no regular session on %s", cal.Today.Date)
	}
	open, closeAt := cal.Today.Regular.Open, cal.Today.Regular.Close
	if !p.now().Before(closeAt.Add(p.cfg.LabelGrace)) {
		return fmt.Errorf("the session of %s is over", cal.Today.Date)
	}
	if err := p.paced.pause(ctx, open.Add(30*time.Second).Sub(p.now())); err != nil {
		return err
	}
	if err := p.startSession(ctx, cal, source); err != nil {
		return err
	}
	for {
		next := p.now().Truncate(time.Minute).Add(time.Minute + 5*time.Second)
		if !next.Before(closeAt.Add(p.cfg.LabelGrace)) {
			break
		}
		if err := p.paced.pause(ctx, next.Sub(p.now())); err != nil {
			return err
		}
		if err := p.step(ctx); err != nil {
			return err
		}
	}
	return p.finish(ctx)
}

// signalRow 는 평가 결과를 signal 행으로 옮김.
func signalRow(symbol string, ev evaluation, q quoteSnap) ledgerRow {
	row := ledgerRow{Kind: kindSignal, Symbol: symbol, Phase: ev.Phase, Refusal: ev.Refusal, SetupID: ev.SetupID,
		SnapshotDigest: ev.SnapshotDigest, ConfigDigest: ev.ConfigDigest, MetricsDigest: ev.MetricsDigest,
		MetricsVersion: breakoutsignal.Version, Transitions: ev.Transitions, Bars: ev.Bars, Tick: ev.Tick, ATR: ev.ATRMinor,
		Resistance: ev.Resistance, RangeLow: ev.RangeLow, RVOL12: ev.RVOL12, RVOL20: ev.RVOL20, RVOL25: ev.RVOL25}
	if ev.ArmedIndex >= 0 {
		row.ArmedIndex = ev.ArmedIndex
		row.Entry, row.Ask, row.Bid, row.Worst, row.Stop = ev.Plan.Entry, q.Ask, q.Bid, ev.Plan.Worst, ev.Plan.Stop
		row.Target, row.RoundTrip, row.Risk, row.Quantity = ev.Plan.Target, ev.Plan.RoundTrip, ev.Plan.Risk, ev.Quantity
		row.SpreadPPM, row.DriftPPM, row.QuoteAgeMS, row.QuoteDigest = ev.SpreadPPM, ev.DriftPPM, ev.AgeMS, q.Digest
	}
	return row
}
