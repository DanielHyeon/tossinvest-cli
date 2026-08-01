package official

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

// RawMinuteCandle preserves every official decimal and timestamp byte used by
// the a047 exact 1m-to-5m aggregator. It is a read DTO, not an order capability.
type RawMinuteCandle struct {
	Timestamp, Open, High, Low, Close, Volume, Currency string
}

const (
	RawCandleIntervalOneMinute     = "1m"
	RawCandleSourceOfficialOpenAPI = "official-open-api"
)

var krxRawCandleSymbol = regexp.MustCompile(`^[0-9]{6}$`)

type RawMinutePage struct {
	market, symbol, interval, source string
	adjusted                         bool
	candles                          []RawMinuteCandle
	nextBefore                       string
	valid                            bool
}

func (p RawMinutePage) Valid() bool        { return p.valid }
func (p RawMinutePage) Market() string     { return p.market }
func (p RawMinutePage) Symbol() string     { return p.symbol }
func (p RawMinutePage) Interval() string   { return p.interval }
func (p RawMinutePage) Adjusted() bool     { return p.adjusted }
func (p RawMinutePage) Source() string     { return p.source }
func (p RawMinutePage) NextBefore() string { return p.nextBefore }
func (p RawMinutePage) Candles() []RawMinuteCandle {
	out := make([]RawMinuteCandle, len(p.candles))
	copy(out, p.candles)
	return out
}

// RawMinuteCandles reads only the official 1m endpoint and does not convert
// decimal strings to float64. The caller validates KST/session/bar integrity.
func (c *Client) RawMinuteCandles(ctx context.Context, market, symbol string, count int, before string, adjusted bool) (RawMinutePage, error) {
	if c == nil {
		return RawMinutePage{}, fmt.Errorf("official raw candles: client is required")
	}
	if market != "KR" || !krxRawCandleSymbol.MatchString(symbol) {
		return RawMinutePage{}, fmt.Errorf("official raw candles: canonical KRX market/symbol is required")
	}
	q := url.Values{}
	q.Set("symbol", symbol)
	q.Set("interval", RawCandleIntervalOneMinute)
	if count > 0 {
		q.Set("count", strconv.Itoa(count))
	}
	if before != "" {
		q.Set("before", before)
	}
	q.Set("adjusted", strconv.FormatBool(adjusted))
	var raw apiCandlePage
	if err := c.get(ctx, "/api/v1/candles", q, &raw); err != nil {
		return RawMinutePage{}, err
	}
	out := RawMinutePage{
		market: market, symbol: symbol, interval: RawCandleIntervalOneMinute,
		adjusted: adjusted, source: RawCandleSourceOfficialOpenAPI,
		candles: make([]RawMinuteCandle, 0, len(raw.Candles)), nextBefore: raw.NextBefore,
		valid: true,
	}
	for _, candle := range raw.Candles {
		out.candles = append(out.candles, RawMinuteCandle{Timestamp: candle.Timestamp, Open: candle.OpenPrice, High: candle.HighPrice, Low: candle.LowPrice, Close: candle.ClosePrice, Volume: candle.Volume, Currency: candle.Currency})
	}
	return out, nil
}
