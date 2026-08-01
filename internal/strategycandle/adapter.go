// Package strategycandle is the only production DTO bridge from official
// Open API candle pages to the order-free strategy market verifier.
package strategycandle

import (
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategymarket"
)

func AdaptOfficialMinutePage(page official.RawMinutePage) (strategymarket.OfficialMinutePage, error) {
	if !page.Valid() || page.Market() != "KR" || page.Interval() != official.RawCandleIntervalOneMinute ||
		page.Source() != official.RawCandleSourceOfficialOpenAPI || page.Adjusted() {
		return strategymarket.OfficialMinutePage{}, fmt.Errorf("strategy candle adapter: valid unadjusted official KRX page required")
	}
	raw := page.Candles()
	candles := make([]strategymarket.RawMinuteCandle, len(raw))
	for index, candle := range raw {
		candles[index] = strategymarket.RawMinuteCandle{
			Timestamp: candle.Timestamp, Open: candle.Open, High: candle.High,
			Low: candle.Low, Close: candle.Close, Volume: candle.Volume, Currency: candle.Currency,
		}
	}
	return strategymarket.SealAdaptedOfficialMinutePage(
		page.Market(), page.Symbol(), page.Interval(), page.Adjusted(),
		strategymarket.SourceOfficialOpenAPI, candles,
	), nil
}

func AdaptAndSealClosedKRXFiveMinute(market, symbol string, page official.RawMinutePage, now time.Time) (strategymarket.VerifiedBar, error) {
	adapted, err := AdaptOfficialMinutePage(page)
	if err != nil {
		return strategymarket.VerifiedBar{}, err
	}
	return strategymarket.SealOfficialClosedKRXFiveMinuteFor(market, symbol, adapted, now)
}
