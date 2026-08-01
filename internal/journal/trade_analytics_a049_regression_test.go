package journal

import (
	"encoding/json"
	"testing"
)

// TestA049DoesNotChangePortfolioAggregateBytes pins the legacy portfolio answer
// independently of the new performance.db read model. Lane attribution may
// segment the same outcomes, but it cannot redefine win rate, profit factor,
// close-ordered drawdown, or realised R.
func TestA049DoesNotChangePortfolioAggregateBytes(t *testing.T) {
	outcomes := []TradeOutcome{
		{RealizedPnLAfterCosts: "300", RealizedR: "1.5"},
		{RealizedPnLAfterCosts: "-100", RealizedR: "-0.5"},
		{RealizedPnLAfterCosts: "-100", RealizedR: "-0.5"},
		{RealizedPnLAfterCosts: "100", RealizedR: "0.5"},
	}
	got, err := json.Marshal(AggregateTradeOutcomes(outcomes))
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"Trades":4,"Wins":2,"Losses":2,"WinRate":"0.5","ProfitFactor":"2","GrossProfit":"400","GrossLoss":"200","NetPnL":"200","MaxDrawdown":"200","SumRealizedR":"1"}`
	if string(got) != want {
		t.Fatalf("portfolio aggregate bytes changed\n got: %s\nwant: %s", got, want)
	}
}
