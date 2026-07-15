package client

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

type lendingExpectedRaw struct {
	ExpectedAmountUsdOneMonth float64 `json:"expectedAmountUsdOneMonth"`
	ExpectedAmountUsdOneYear  float64 `json:"expectedAmountUsdOneYear"`
	Items                     []struct {
		Guid      string  `json:"guid"`
		StockName string  `json:"stockName"`
		Amount    float64 `json:"amount"`
	} `json:"items"`
}

// GetLendingExpected fetches projected share-lending (대주) income for the
// account: monthly/yearly USD totals plus a per-stock breakdown. WTS-only.
func (c *Client) GetLendingExpected(ctx context.Context) (domain.LendingExpected, error) {
	if err := c.requireSession(); err != nil {
		return domain.LendingExpected{}, err
	}
	var env quoteEnvelope[lendingExpectedRaw]
	endpoint := c.certBaseURL + "/api/v1/lending/revenue/account/expected"
	if err := c.getJSON(ctx, endpoint, &env); err != nil {
		return domain.LendingExpected{}, err
	}
	out := domain.LendingExpected{
		OneMonthUSD: env.Result.ExpectedAmountUsdOneMonth,
		OneYearUSD:  env.Result.ExpectedAmountUsdOneYear,
		FetchedAt:   time.Now(),
	}
	for _, it := range env.Result.Items {
		out.Stocks = append(out.Stocks, domain.LendingExpectedStock{
			ProductCode: it.Guid,
			Name:        it.StockName,
			AmountUSD:   it.Amount,
		})
	}
	return out, nil
}
