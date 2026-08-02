package console

import "strings"

func (r positionRow) HasExitReference() bool { return r.ExitReference.Present() }

func (r positionRow) MarketLabel() string {
	if market := strings.ToUpper(strings.TrimSpace(r.Market)); market != "" {
		return market
	}
	return "—"
}

func (r positionRow) CurrencyLabel() string {
	switch r.MarketLabel() {
	case "KR", "KRX", "KOSPI", "KOSDAQ":
		return "KRW"
	case "US", "NASDAQ", "NYSE", "AMEX":
		return "USD"
	default:
		return "—"
	}
}
