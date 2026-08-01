package official

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// RawMinuteCandle preserves every official decimal and timestamp byte used by
// the a047 exact 1m-to-5m aggregator. It is a read DTO, not an order capability.
type RawMinuteCandle struct {
	Timestamp, Open, High, Low, Close, Volume, Currency string
}

type RawMinutePage struct {
	Candles    []RawMinuteCandle
	NextBefore string
}

// RawMinuteCandles reads only the official 1m endpoint and does not convert
// decimal strings to float64. The caller validates KST/session/bar integrity.
func (c *Client) RawMinuteCandles(ctx context.Context, symbol string, count int, before string, adjusted bool) (RawMinutePage, error) {
	if c == nil {
		return RawMinutePage{}, fmt.Errorf("official raw candles: client is required")
	}
	if symbol == "" {
		return RawMinutePage{}, fmt.Errorf("official raw candles: symbol is required")
	}
	q := url.Values{}
	q.Set("symbol", symbol)
	q.Set("interval", "1m")
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
	out := RawMinutePage{Candles: make([]RawMinuteCandle, 0, len(raw.Candles)), NextBefore: raw.NextBefore}
	for _, candle := range raw.Candles {
		out.Candles = append(out.Candles, RawMinuteCandle{Timestamp: candle.Timestamp, Open: candle.OpenPrice, High: candle.HighPrice, Low: candle.LowPrice, Close: candle.ClosePrice, Volume: candle.Volume, Currency: candle.Currency})
	}
	return out, nil
}
