package execgw_test

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestPairedStrategyReadBudgetsRemainMarketLocal(t *testing.T) {
	budgets := execgw.StrategyReadBudgets()
	if len(budgets) != 2 {
		t.Fatalf("strategy read budgets=%v, want exact KR identity and US FX", budgets)
	}
	want := map[execgw.RequiredQuery]struct {
		market, evidence, endpoint string
		stale                      time.Duration
	}{
		execgw.QueryAccountIdentity: {"KR", "same-currency", "GET /api/v1/accounts", 5 * time.Minute},
		execgw.QueryExchangeRate:    {"US", "official-fx", "GET /api/v1/exchange-rate", 15 * time.Second},
	}
	for query, expected := range want {
		budget, ok := budgets[query]
		if !ok {
			t.Fatalf("missing strategy query %q: %v", query, budgets)
		}
		if budget.Query != query || budget.Market != expected.market || budget.EvidenceSource != expected.evidence ||
			budget.EndpointSource != "official-open-api" || budget.Endpoint != expected.endpoint || budget.StaleAfter != expected.stale {
			t.Errorf("budget[%s]=%+v, want market=%s evidence=%s endpoint=%s stale=%s",
				query, budget, expected.market, expected.evidence, expected.endpoint, expected.stale)
		}
		if budget.SoakMaxRequestsPerCycle != 1 {
			t.Errorf("budget[%s] soak requests=%d, want 1", query, budget.SoakMaxRequestsPerCycle)
		}
		if budget.Retry.MaxAttempts != 3 || budget.Retry.Budget != 8*time.Second ||
			budget.Retry.BaseBackoff != 400*time.Millisecond || budget.Retry.MaxBackoff != 3*time.Second ||
			budget.Retry.MaxRetryAfter != 30*time.Second || budget.Retry.JitterFraction != 0.25 {
			t.Errorf("budget[%s] retry=%+v, want bounded default read policy", query, budget.Retry)
		}
	}

	if global := execgw.DefaultStaleness(); global[execgw.QueryAccountIdentity] != 0 || global[execgw.QueryExchangeRate] != 0 {
		t.Fatalf("market-local strategy reads leaked into account-wide entry gate: %v", global)
	}

	budgets[execgw.QueryExchangeRate] = execgw.StrategyReadBudget{}
	if fresh := execgw.StrategyReadBudgets()[execgw.QueryExchangeRate]; fresh.Market != "US" || fresh.Endpoint == "" {
		t.Fatalf("caller mutation changed strategy read catalog: %+v", fresh)
	}
}
