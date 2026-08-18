// Package strategymarket converts official raw reads into exact, closed KRX
// strategy inputs. It owns no order or journal dependency.
package strategymarket

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Refusal string

const (
	RefusalInvalidDecimal        Refusal = "invalid_decimal"
	RefusalNaiveTimestamp        Refusal = "timezone_naive_timestamp"
	RefusalOutsideRegularSession Refusal = "outside_regular_session"
	RefusalMinuteGap             Refusal = "minute_gap"
	RefusalIncompleteBucket      Refusal = "incomplete_bucket"
	RefusalOpenBucket            Refusal = "bar_not_closed"
	RefusalCurrency              Refusal = "currency_mismatch"
	RefusalStateUnavailable      Refusal = "symbol_state_unavailable"
	RefusalStateStale            Refusal = "symbol_state_stale"
	RefusalStateBlocked          Refusal = "symbol_state_blocked"
	RefusalIdentity              Refusal = "request_identity_mismatch"
	RefusalAdjusted              Refusal = "adjusted_candles_forbidden"
	RefusalSource                Refusal = "source_unavailable_or_untrusted"
	RefusalInterval              Refusal = "interval_mismatch"
)

type MarketSource string

const (
	IntervalOneMinute                      = "1m"
	SourceOfficialOpenAPI     MarketSource = "official-open-api"
	SourceOfficialSymbolState MarketSource = "official-symbol-state"
	SourceOfficialPosition    MarketSource = "official-position"
)

type IntegrityError struct {
	Kind   Refusal
	Detail string
}

func (e *IntegrityError) Error() string {
	return "strategy market: " + string(e.Kind) + ": " + e.Detail
}

type RawMinuteCandle struct {
	Timestamp string
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	Currency  string
}

// OfficialMinutePage is an opaque adapter product. Production construction is
// statically restricted to internal/strategycandle; its scalar constructor is
// exported solely because Go has no friend-package visibility.
type OfficialMinutePage struct {
	market, symbol, interval string
	adjusted                 bool
	source                   MarketSource
	candles                  []RawMinuteCandle
	valid                    bool
}

func SealAdaptedOfficialMinutePage(market, symbol, interval string, adjusted bool, source MarketSource, candles []RawMinuteCandle) OfficialMinutePage {
	copyCandles := make([]RawMinuteCandle, len(candles))
	copy(copyCandles, candles)
	return OfficialMinutePage{market: market, symbol: symbol, interval: interval, adjusted: adjusted, source: source, candles: copyCandles, valid: true}
}

type VerifiedBar struct {
	market, symbol, source         string
	adjusted                       bool
	openAt, closedAt               time.Time
	open, high, low, close, volume string
	currency                       string
	valid                          bool
}

func (b VerifiedBar) Valid() bool         { return b.valid }
func (b VerifiedBar) Market() string      { return b.market }
func (b VerifiedBar) Symbol() string      { return b.symbol }
func (b VerifiedBar) Source() string      { return b.source }
func (b VerifiedBar) Adjusted() bool      { return b.adjusted }
func (b VerifiedBar) OpenAt() time.Time   { return b.openAt }
func (b VerifiedBar) ClosedAt() time.Time { return b.closedAt }
func (b VerifiedBar) Open() string        { return b.open }
func (b VerifiedBar) High() string        { return b.high }
func (b VerifiedBar) Low() string         { return b.low }
func (b VerifiedBar) Close() string       { return b.close }
func (b VerifiedBar) Volume() string      { return b.volume }
func (b VerifiedBar) Currency() string    { return b.currency }

var offsetTimestamp = regexp.MustCompile(`(?:Z|[+-][0-9]{2}:[0-9]{2})$`)

// AggregateClosedKRXFiveMinute no longer accepts caller-asserted raw bytes as
// verified strategy data. Use the official-page adapter boundary instead.
func AggregateClosedKRXFiveMinute(_ []RawMinuteCandle, _ time.Time) (VerifiedBar, error) {
	return VerifiedBar{}, &IntegrityError{Kind: RefusalSource, Detail: "opaque official page required"}
}

func SealOfficialClosedKRXFiveMinute(page OfficialMinutePage, now time.Time) (VerifiedBar, error) {
	return SealOfficialClosedKRXFiveMinuteFor(page.market, page.symbol, page, now)
}

func SealOfficialClosedKRXFiveMinuteFor(market, symbol string, page OfficialMinutePage, now time.Time) (VerifiedBar, error) {
	switch {
	case !page.valid:
		return VerifiedBar{}, &IntegrityError{Kind: RefusalSource, Detail: "sealed adapter page required"}
	case market != "KR" || symbol == "" || page.market != market || page.symbol != symbol:
		return VerifiedBar{}, &IntegrityError{Kind: RefusalIdentity, Detail: "exact KRX request identity required"}
	case page.interval != IntervalOneMinute:
		return VerifiedBar{}, &IntegrityError{Kind: RefusalInterval, Detail: page.interval}
	case page.adjusted:
		return VerifiedBar{}, &IntegrityError{Kind: RefusalAdjusted, Detail: "unadjusted official candles required"}
	case page.source != SourceOfficialOpenAPI:
		return VerifiedBar{}, &IntegrityError{Kind: RefusalSource, Detail: string(page.source)}
	}
	return aggregateClosedKRXFiveMinute(market, symbol, string(page.source), page.adjusted, page.candles, now)
}

func aggregateClosedKRXFiveMinute(market, symbol, source string, adjusted bool, raw []RawMinuteCandle, now time.Time) (VerifiedBar, error) {
	if len(raw) != 5 {
		return VerifiedBar{}, &IntegrityError{Kind: RefusalIncompleteBucket, Detail: "exactly five minutes required"}
	}
	if now.IsZero() {
		return VerifiedBar{}, &IntegrityError{Kind: RefusalOpenBucket, Detail: "injected now required"}
	}
	seoul, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return VerifiedBar{}, err
	}
	type parsedMinute struct {
		candle RawMinuteCandle
		local  time.Time
		values [5]*big.Rat
	}
	minutes := make([]parsedMinute, 0, 5)
	currency := ""
	for _, candle := range raw {
		if !offsetTimestamp.MatchString(candle.Timestamp) {
			return VerifiedBar{}, &IntegrityError{Kind: RefusalNaiveTimestamp, Detail: "RFC3339 offset required"}
		}
		parsed, parseErr := time.Parse(time.RFC3339, candle.Timestamp)
		if parseErr != nil {
			return VerifiedBar{}, &IntegrityError{Kind: RefusalNaiveTimestamp, Detail: "invalid RFC3339 timestamp"}
		}
		local := parsed.In(seoul)
		minuteOfDay := local.Hour()*60 + local.Minute()
		// 브로커가 준 시각표는 봉이 "닫힌" 시각이다(a112 결정 30·31 라이브 실측).
		// 그래서 라벨이 t인 1분봉은 [t-1분, t) 구간을 담는다. 정규장(09:00~15:30) 안에
		// 있으려면 라벨이 09:01 이상 15:30 이하여야 한다. 09:00 라벨은 08:59~09:00,
		// 즉 개장 전 1분이라 여기서 걸러진다.
		if local.Second() != 0 || local.Nanosecond() != 0 || minuteOfDay < 9*60+1 || minuteOfDay > 15*60+30 {
			return VerifiedBar{}, &IntegrityError{Kind: RefusalOutsideRegularSession, Detail: candle.Timestamp}
		}
		if currency == "" {
			currency = candle.Currency
		}
		if candle.Currency != "KRW" || candle.Currency != currency {
			return VerifiedBar{}, &IntegrityError{Kind: RefusalCurrency, Detail: "KRW required"}
		}
		parsedMinute := parsedMinute{candle: candle, local: local}
		fields := []string{candle.Open, candle.High, candle.Low, candle.Close, candle.Volume}
		for j, field := range fields {
			v, ok := exactDecimal(field)
			if !ok || v.Sign() < 0 {
				return VerifiedBar{}, &IntegrityError{Kind: RefusalInvalidDecimal, Detail: field}
			}
			parsedMinute.values[j] = v
		}
		minutes = append(minutes, parsedMinute)
	}
	sort.Slice(minutes, func(i, j int) bool { return minutes[i].local.Before(minutes[j].local) })
	// 버킷이 여는 시각은 첫 봉의 라벨보다 1분 이르다. 5분 격자는 09:00에서 시작하므로
	// 정렬은 라벨이 아니라 이 여는 시각으로 따진다.
	openAt := minutes[0].local.Add(-time.Minute)
	openMinute := openAt.Hour()*60 + openAt.Minute()
	if (openMinute-9*60)%5 != 0 {
		return VerifiedBar{}, &IntegrityError{Kind: RefusalIncompleteBucket, Detail: "not aligned to KRX five-minute boundary"}
	}
	for i := 1; i < 5; i++ {
		if !minutes[i].local.Equal(minutes[0].local.Add(time.Duration(i) * time.Minute)) {
			return VerifiedBar{}, &IntegrityError{Kind: RefusalMinuteGap, Detail: "minutes are not contiguous"}
		}
	}
	// 버킷이 닫히는 시각은 여는 시각의 5분 뒤이고, 그것은 마지막 봉의 라벨과 같다.
	closedAt := openAt.Add(5 * time.Minute)
	if now.Before(closedAt) {
		return VerifiedBar{}, &IntegrityError{Kind: RefusalOpenBucket, Detail: "bucket is not closed"}
	}
	high := new(big.Rat).Set(minutes[0].values[1])
	low := new(big.Rat).Set(minutes[0].values[2])
	volume := new(big.Rat)
	for i := 0; i < 5; i++ {
		if minutes[i].values[1].Cmp(high) > 0 {
			high.Set(minutes[i].values[1])
		}
		if minutes[i].values[2].Cmp(low) < 0 {
			low.Set(minutes[i].values[2])
		}
		volume.Add(volume, minutes[i].values[4])
	}
	return VerifiedBar{
		market:   market,
		symbol:   symbol,
		source:   source,
		adjusted: adjusted,
		openAt:   openAt,
		closedAt: closedAt,
		open:     decimalString(minutes[0].values[0]),
		high:     decimalString(high),
		low:      decimalString(low),
		close:    decimalString(minutes[4].values[3]),
		volume:   decimalString(volume),
		currency: currency,
		valid:    true,
	}, nil
}

func exactDecimal(raw string) (*big.Rat, bool) {
	if raw == "" || strings.TrimSpace(raw) != raw || strings.ContainsAny(raw, "eE/") {
		return nil, false
	}
	v, ok := new(big.Rat).SetString(raw)
	return v, ok
}
func decimalString(v *big.Rat) string {
	numerator := new(big.Int).Set(v.Num())
	denominator := new(big.Int).Set(v.Denom())
	two, five := big.NewInt(2), big.NewInt(5)
	remainder := new(big.Int)
	twos, fives := 0, 0
	for {
		q := new(big.Int)
		q.QuoRem(denominator, two, remainder)
		if remainder.Sign() != 0 {
			break
		}
		denominator = q
		twos++
	}
	for {
		q := new(big.Int)
		q.QuoRem(denominator, five, remainder)
		if remainder.Sign() != 0 {
			break
		}
		denominator = q
		fives++
	}
	if denominator.Cmp(big.NewInt(1)) != 0 {
		return v.RatString()
	}
	scale := max(twos, fives)
	numerator.Mul(numerator, new(big.Int).Exp(two, big.NewInt(int64(scale-twos)), nil))
	numerator.Mul(numerator, new(big.Int).Exp(five, big.NewInt(int64(scale-fives)), nil))
	negative := numerator.Sign() < 0
	numerator.Abs(numerator)
	digits := numerator.String()
	if scale == 0 {
		if negative {
			return "-" + digits
		}
		return digits
	}
	if len(digits) <= scale {
		digits = strings.Repeat("0", scale-len(digits)+1) + digits
	}
	point := len(digits) - scale
	out := strings.TrimRight(digits[:point]+"."+digits[point:], "0")
	out = strings.TrimSuffix(out, ".")
	if negative && out != "0" {
		out = "-" + out
	}
	return out
}

type SymbolState string

const (
	StateNormal  SymbolState = "NORMAL"
	StateHalt    SymbolState = "HALT"
	StateLimit   SymbolState = "LIMIT"
	StateManaged SymbolState = "MANAGED"
)

type StateReading struct {
	Market     string
	Symbol     string
	State      SymbolState
	ObservedAt time.Time
	Source     MarketSource
}
type FreshNormalState struct {
	market, symbol, authority string
	observedAt                time.Time
	valid                     bool
}

func (s FreshNormalState) Valid() bool           { return s.valid }
func (s FreshNormalState) Market() string        { return s.market }
func (s FreshNormalState) Symbol() string        { return s.symbol }
func (s FreshNormalState) ObservedAt() time.Time { return s.observedAt }
func (s FreshNormalState) Authority() string     { return s.authority }

type SymbolStateSource interface {
	ReadSymbolState(context.Context, string, string) (StateReading, error)
}

var ErrSymbolStateNotConfigured = errors.New("strategy symbol state: authority not configured")

func RequireFreshNormalState(ctx context.Context, source SymbolStateSource, market, symbol string, now time.Time) (FreshNormalState, error) {
	if source == nil || now.IsZero() {
		return FreshNormalState{}, &IntegrityError{Kind: RefusalStateUnavailable, Detail: ErrSymbolStateNotConfigured.Error()}
	}
	reading, err := source.ReadSymbolState(ctx, market, symbol)
	if err != nil || reading.Market != market || reading.Symbol != symbol ||
		reading.Source != SourceOfficialSymbolState || reading.ObservedAt.IsZero() {
		return FreshNormalState{}, &IntegrityError{Kind: RefusalStateUnavailable, Detail: fmt.Sprint(err)}
	}
	age := now.UTC().Sub(reading.ObservedAt.UTC())
	if age < 0 || age > 30*time.Second {
		return FreshNormalState{}, &IntegrityError{Kind: RefusalStateStale, Detail: age.String()}
	}
	if reading.State != StateNormal {
		return FreshNormalState{}, &IntegrityError{Kind: RefusalStateBlocked, Detail: string(reading.State)}
	}
	return FreshNormalState{
		market:     market,
		symbol:     symbol,
		authority:  string(reading.Source),
		observedAt: reading.ObservedAt.UTC(),
		valid:      true,
	}, nil
}

type PositionReading struct {
	Market     string
	Symbol     string
	Quantity   string
	OpenOrders int
	ObservedAt time.Time
	Source     MarketSource
}

type NoPositionProof struct {
	market, symbol, authority string
	observedAt                time.Time
	valid                     bool
}

func (p NoPositionProof) Valid() bool           { return p.valid }
func (p NoPositionProof) Market() string        { return p.market }
func (p NoPositionProof) Symbol() string        { return p.symbol }
func (p NoPositionProof) Authority() string     { return p.authority }
func (p NoPositionProof) ObservedAt() time.Time { return p.observedAt }

type PositionSource interface {
	ReadPosition(context.Context, string, string) (PositionReading, error)
}

func RequireNoPosition(ctx context.Context, source PositionSource, market, symbol string, now time.Time) (NoPositionProof, error) {
	if source == nil || now.IsZero() || strings.TrimSpace(market) == "" || strings.TrimSpace(symbol) == "" {
		return NoPositionProof{}, &IntegrityError{Kind: RefusalStateUnavailable, Detail: "position authority not configured"}
	}
	reading, err := source.ReadPosition(ctx, market, symbol)
	if err != nil || reading.Market != market || reading.Symbol != symbol ||
		reading.Source != SourceOfficialPosition || reading.ObservedAt.IsZero() {
		return NoPositionProof{}, &IntegrityError{Kind: RefusalStateUnavailable, Detail: fmt.Sprint(err)}
	}
	age := now.UTC().Sub(reading.ObservedAt.UTC())
	quantity, validQuantity := exactDecimal(reading.Quantity)
	if age < 0 || age > 30*time.Second {
		return NoPositionProof{}, &IntegrityError{Kind: RefusalStateStale, Detail: age.String()}
	}
	if !validQuantity || decimalString(quantity) != reading.Quantity || quantity.Sign() != 0 || reading.OpenOrders != 0 {
		return NoPositionProof{}, &IntegrityError{Kind: RefusalStateBlocked, Detail: "position or open order exists"}
	}
	return NoPositionProof{market: market, symbol: symbol, authority: string(reading.Source), observedAt: reading.ObservedAt.UTC(), valid: true}, nil
}
