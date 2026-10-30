package engine

import (
	"context"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

const (
	exitLowerLimitRefreshName = "exit-lower-limit-refresh"
	exitLowerLimitQuiet       = 5 * time.Minute
)

type exitPriceSource string

const (
	exitPriceSourceLowerLimit exitPriceSource = "lower_limit"
	exitPriceSourceObserved   exitPriceSource = "observed"
	exitPriceSourceBaseline   exitPriceSource = "baseline"
)

type exitLimitChoice struct {
	price  string
	source exitPriceSource
}

type lowerLimitRecord struct {
	price     string
	timestamp time.Time
}

type lowerLimitSnapshot struct {
	records map[string]lowerLimitRecord
}

type lowerLimitReader interface {
	ProtectivePriceLimits(context.Context, string) (official.ProtectivePriceLimits, error)
}

type lowerLimitPositions interface {
	Positions(context.Context, string) ([]journal.Position, error)
}

type lowerLimitAttempts interface {
	HasProtectiveExitAttempt(context.Context, string, string, string, string, []string) (bool, error)
}

// lowerLimitAdmission 은 a166이 제공할 wire 단위 관문의 존재 표식임.
// a087은 이를 호출하지 않아 같은 요청에 별도 토큰을 소비하지 않음.
type lowerLimitAdmission interface{ Admit(context.Context) error }

type exitLowerLimitBoard struct {
	clk       clock.Clock
	reader    lowerLimitReader
	positions lowerLimitPositions
	attempts  lowerLimitAttempts
	account   string
	admission lowerLimitAdmission

	write     sync.Mutex
	snapshot  atomic.Pointer[lowerLimitSnapshot]
	attempted map[string]bool
	tokens    chan time.Duration
}

func newExitLowerLimitBoard(clk clock.Clock, reader lowerLimitReader, positions lowerLimitPositions,
	attempts lowerLimitAttempts, account string, admission lowerLimitAdmission) *exitLowerLimitBoard {
	if clk == nil {
		clk = clock.System()
	}
	return &exitLowerLimitBoard{
		clk: clk, reader: reader, positions: positions, attempts: attempts,
		account: strings.TrimSpace(account), admission: admission,
		attempted: make(map[string]bool), tokens: make(chan time.Duration, 1),
	}
}

func lowerLimitKey(market, symbol string) string {
	return strings.ToLower(strings.TrimSpace(market)) + "|" + strings.ToUpper(strings.TrimSpace(symbol))
}

func (b *exitLowerLimitBoard) publish(next lowerLimitSnapshot) {
	if b == nil {
		return
	}
	clone := make(map[string]lowerLimitRecord, len(next.records))
	for key, value := range next.records {
		clone[key] = value
	}
	b.snapshot.Store(&lowerLimitSnapshot{records: clone})
}

func (b *exitLowerLimitBoard) record(market, symbol string) (lowerLimitRecord, bool) {
	if b == nil {
		return lowerLimitRecord{}, false
	}
	snapshot := b.snapshot.Load()
	if snapshot == nil {
		return lowerLimitRecord{}, false
	}
	record, ok := snapshot.records[lowerLimitKey(market, symbol)]
	return record, ok
}

func (b *exitLowerLimitBoard) attemptedOn(market, symbol string, now time.Time) bool {
	if b == nil {
		return true
	}
	day, err := clock.MarketKR.TradingDay(now)
	if err != nil {
		return true
	}
	b.write.Lock()
	defer b.write.Unlock()
	return b.attempted[day+"|"+lowerLimitKey(market, symbol)]
}

func (b *exitLowerLimitBoard) markProtectiveAttempt(market, symbol string, now time.Time) {
	if b == nil || b.admission == nil {
		return
	}
	day, err := clock.MarketKR.TradingDay(now)
	if err != nil {
		return
	}
	b.write.Lock()
	if b.attempted == nil {
		b.attempted = make(map[string]bool)
	}
	b.attempted[day+"|"+lowerLimitKey(market, symbol)] = true
	b.write.Unlock()
}

func (b *exitLowerLimitBoard) cycleComplete(interval time.Duration) {
	if b == nil || b.admission == nil {
		return
	}
	select {
	case b.tokens <- interval:
	default:
	}
}

// ExitLowerLimitRefresher 는 하한가 판의 보조 갱신기임.
// admission이 없으면 취소까지 대기하여 a166 전 broker 읽기를 0으로 유지함.
func (c *Context) ExitLowerLimitRefresher(_ clock.Clock) AuxiliaryExecutor {
	var board *exitLowerLimitBoard
	if c != nil {
		board = c.exitLowerLimits
	}
	return AuxiliaryExecutor{Name: exitLowerLimitRefreshName, Run: board.Run}
}

func (b *exitLowerLimitBoard) canRefresh() bool {
	return b != nil && b.admission != nil && b.reader != nil && b.positions != nil && b.attempts != nil
}

func (b *exitLowerLimitBoard) Run(ctx context.Context) error {
	if !b.canRefresh() {
		<-ctx.Done()
		return ctx.Err()
	}
	started := b.clk.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case interval := <-b.tokens:
			if b.clk.Now().Sub(started) < exitLowerLimitQuiet {
				continue
			}
			b.refresh(ctx, interval)
		}
	}
}

func (b *exitLowerLimitBoard) refresh(ctx context.Context, interval time.Duration) {
	readLimit := lowerLimitReadTimeout(interval)
	prefetchLimit := interval - readLimit
	if prefetchLimit <= 0 {
		prefetchLimit = readLimit
	}
	prefetchCtx, cancelPrefetch := context.WithTimeout(ctx, prefetchLimit)
	positions, err := b.positions.Positions(prefetchCtx, b.account)
	if err != nil || prefetchCtx.Err() != nil {
		cancelPrefetch()
		return
	}
	defer cancelPrefetch()
	now := b.clk.Now()
	if !inKoreanLimitWindow(now) {
		return
	}
	day, err := clock.AccountTradingDay(now)
	if err != nil {
		return
	}
	for _, position := range positions {
		if !strings.EqualFold(position.Market, "kr") || position.State == journal.PositionClosed {
			continue
		}
		if record, exists := b.record(position.Market, position.Symbol); exists && recordValidNow(now, record) {
			continue
		}
		attempted, err := b.attempts.HasProtectiveExitAttempt(prefetchCtx, b.account, position.Market, day, position.Symbol, protectiveActionValues())
		if err != nil || prefetchCtx.Err() != nil {
			return
		}
		if attempted {
			b.markProtectiveAttempt(position.Market, position.Symbol, now)
			continue
		}
		readCtx, cancel := context.WithTimeout(ctx, readLimit)
		raw, err := b.reader.ProtectivePriceLimits(readCtx, position.Symbol)
		cancel()
		if err != nil {
			return
		}
		record, ok := validLowerLimitRecord(now, raw)
		if !ok {
			return
		}
		current := b.snapshot.Load()
		next := lowerLimitSnapshot{records: map[string]lowerLimitRecord{}}
		if current != nil {
			for key, value := range current.records {
				next.records[key] = value
			}
		}
		next.records[lowerLimitKey(position.Market, position.Symbol)] = record
		b.publish(next)
		return
	}
}

func lowerLimitReadTimeout(interval time.Duration) time.Duration {
	limit := 2 * time.Second
	if interval > 0 && interval/2 < limit {
		limit = interval / 2
	}
	if limit <= 0 {
		return time.Millisecond
	}
	return limit
}

func validLowerLimitRecord(now time.Time, raw official.ProtectivePriceLimits) (lowerLimitRecord, bool) {
	if raw.LowerLimitPrice == nil || raw.Timestamp == nil || !positiveCanonicalInteger(*raw.LowerLimitPrice) {
		return lowerLimitRecord{}, false
	}
	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(*raw.Timestamp))
	if err != nil {
		return lowerLimitRecord{}, false
	}
	sameDay, err := clock.MarketKR.SameTradingDay(now, timestamp)
	if err != nil || !sameDay {
		return lowerLimitRecord{}, false
	}
	location, err := clock.MarketKR.Location()
	if err != nil {
		return lowerLimitRecord{}, false
	}
	local := timestamp.In(location)
	if local.Hour() < 9 {
		return lowerLimitRecord{}, false
	}
	return lowerLimitRecord{price: *raw.LowerLimitPrice, timestamp: timestamp}, true
}

func exitLimitPrice(now time.Time, m managed, proposal exitpolicy.Proposal, observed string,
	board *exitLowerLimitBoard) exitLimitChoice {
	fallback := exitFallbackLimit(m, observed)
	if !isProtective(proposal) || !inKoreanLimitWindow(now) || !strings.EqualFold(strings.TrimSpace(m.position.Market), "kr") {
		return fallback
	}
	record, ok := board.record(m.position.Market, m.position.Symbol)
	if !ok || board.attemptedOn(m.position.Market, m.position.Symbol, now) || !recordValidNow(now, record) {
		return fallback
	}
	if !positiveCanonicalInteger(record.price) || !decimalAtMost(record.price, strings.TrimSpace(observed)) {
		return fallback
	}
	return exitLimitChoice{price: record.price, source: exitPriceSourceLowerLimit}
}

func exitFallbackLimit(m managed, observed string) exitLimitChoice {
	if price := strings.TrimSpace(observed); price != "" {
		return exitLimitChoice{price: price, source: exitPriceSourceObserved}
	}
	return exitLimitChoice{price: strings.TrimSpace(m.state.Baseline), source: exitPriceSourceBaseline}
}

func inKoreanLimitWindow(now time.Time) bool {
	inSession, err := clock.MarketKR.InRegularSession(now)
	return err == nil && inSession
}

func recordValidNow(now time.Time, record lowerLimitRecord) bool {
	if record.timestamp.IsZero() {
		return false
	}
	sameDay, err := clock.MarketKR.SameTradingDay(now, record.timestamp)
	if err != nil || !sameDay {
		return false
	}
	location, err := clock.MarketKR.Location()
	return err == nil && record.timestamp.In(location).Hour() >= 9
}

func positiveCanonicalInteger(value string) bool {
	if len(value) == 0 || value[0] == '0' {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func decimalAtMost(lower, observed string) bool {
	if observed == "" {
		return false
	}
	lowerRat, ok := new(big.Rat).SetString(lower)
	if !ok || lowerRat.Sign() <= 0 {
		return false
	}
	observedRat, ok := new(big.Rat).SetString(observed)
	return ok && observedRat.Sign() > 0 && lowerRat.Cmp(observedRat) <= 0
}

func protectiveActionValues() []string {
	return []string{string(exitpolicy.ActionBaselineBreach), string(exitpolicy.ActionLadderStop)}
}
