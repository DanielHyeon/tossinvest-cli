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

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
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
)

type IntegrityError struct {
	Kind   Refusal
	Detail string
}

func (e *IntegrityError) Error() string {
	return "strategy market: " + string(e.Kind) + ": " + e.Detail
}

type FiveMinuteBar struct {
	OpenAt, ClosedAt                         time.Time
	Open, High, Low, Close, Volume, Currency string
}

var offsetTimestamp = regexp.MustCompile(`(?:Z|[+-][0-9]{2}:[0-9]{2})$`)

func AggregateClosedKRXFiveMinute(raw []official.RawMinuteCandle, now time.Time) (FiveMinuteBar, error) {
	if len(raw) != 5 {
		return FiveMinuteBar{}, &IntegrityError{Kind: RefusalIncompleteBucket, Detail: "exactly five minutes required"}
	}
	if now.IsZero() {
		return FiveMinuteBar{}, &IntegrityError{Kind: RefusalOpenBucket, Detail: "injected now required"}
	}
	seoul, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return FiveMinuteBar{}, err
	}
	type parsedMinute struct {
		candle official.RawMinuteCandle
		local  time.Time
		values [5]*big.Rat
	}
	minutes := make([]parsedMinute, 0, 5)
	currency := ""
	for _, candle := range raw {
		if !offsetTimestamp.MatchString(candle.Timestamp) {
			return FiveMinuteBar{}, &IntegrityError{Kind: RefusalNaiveTimestamp, Detail: "RFC3339 offset required"}
		}
		parsed, parseErr := time.Parse(time.RFC3339, candle.Timestamp)
		if parseErr != nil {
			return FiveMinuteBar{}, &IntegrityError{Kind: RefusalNaiveTimestamp, Detail: "invalid RFC3339 timestamp"}
		}
		local := parsed.In(seoul)
		minuteOfDay := local.Hour()*60 + local.Minute()
		if local.Second() != 0 || local.Nanosecond() != 0 || minuteOfDay < 9*60 || minuteOfDay >= 15*60+30 {
			return FiveMinuteBar{}, &IntegrityError{Kind: RefusalOutsideRegularSession, Detail: candle.Timestamp}
		}
		if currency == "" {
			currency = candle.Currency
		}
		if candle.Currency != "KRW" || candle.Currency != currency {
			return FiveMinuteBar{}, &IntegrityError{Kind: RefusalCurrency, Detail: "KRW required"}
		}
		parsedMinute := parsedMinute{candle: candle, local: local}
		fields := []string{candle.Open, candle.High, candle.Low, candle.Close, candle.Volume}
		for j, field := range fields {
			v, ok := exactDecimal(field)
			if !ok || v.Sign() < 0 {
				return FiveMinuteBar{}, &IntegrityError{Kind: RefusalInvalidDecimal, Detail: field}
			}
			parsedMinute.values[j] = v
		}
		minutes = append(minutes, parsedMinute)
	}
	sort.Slice(minutes, func(i, j int) bool { return minutes[i].local.Before(minutes[j].local) })
	startMinute := minutes[0].local.Hour()*60 + minutes[0].local.Minute()
	if (startMinute-9*60)%5 != 0 {
		return FiveMinuteBar{}, &IntegrityError{Kind: RefusalIncompleteBucket, Detail: "not aligned to KRX five-minute boundary"}
	}
	for i := 1; i < 5; i++ {
		if !minutes[i].local.Equal(minutes[0].local.Add(time.Duration(i) * time.Minute)) {
			return FiveMinuteBar{}, &IntegrityError{Kind: RefusalMinuteGap, Detail: "minutes are not contiguous"}
		}
	}
	closedAt := minutes[0].local.Add(5 * time.Minute)
	if now.Before(closedAt) {
		return FiveMinuteBar{}, &IntegrityError{Kind: RefusalOpenBucket, Detail: "bucket is not closed"}
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
	return FiveMinuteBar{OpenAt: minutes[0].local, ClosedAt: closedAt, Open: decimalString(minutes[0].values[0]), High: decimalString(high), Low: decimalString(low), Close: decimalString(minutes[4].values[3]), Volume: decimalString(volume), Currency: currency}, nil
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
	State      SymbolState
	ObservedAt time.Time
	Authority  string
}
type SymbolStateSource interface {
	ReadSymbolState(context.Context, string, string) (StateReading, error)
}

var ErrSymbolStateNotConfigured = errors.New("strategy symbol state: authority not configured")

func RequireFreshNormalState(ctx context.Context, source SymbolStateSource, market, symbol string, now time.Time) (StateReading, error) {
	if source == nil || now.IsZero() {
		return StateReading{}, &IntegrityError{Kind: RefusalStateUnavailable, Detail: ErrSymbolStateNotConfigured.Error()}
	}
	reading, err := source.ReadSymbolState(ctx, market, symbol)
	if err != nil || reading.Authority == "" || reading.ObservedAt.IsZero() {
		return StateReading{}, &IntegrityError{Kind: RefusalStateUnavailable, Detail: fmt.Sprint(err)}
	}
	age := now.UTC().Sub(reading.ObservedAt.UTC())
	if age < 0 || age > 30*time.Second {
		return StateReading{}, &IntegrityError{Kind: RefusalStateStale, Detail: age.String()}
	}
	if reading.State != StateNormal {
		return StateReading{}, &IntegrityError{Kind: RefusalStateBlocked, Detail: string(reading.State)}
	}
	return reading, nil
}
