package execgw

import "time"

// Strategy authority reads are market-local. They deliberately do not appear
// in DefaultStaleness: that map is account-wide, so putting US FX there would
// block an otherwise healthy KR evaluation and every other new-entry caller.
const (
	QueryAccountIdentity RequiredQuery = "strategy_account_identity"
	QueryExchangeRate    RequiredQuery = "strategy_exchange_rate"
)

// StrategyReadBudget is the immutable operational contract for one market's
// authority read. It describes no monetary value and grants no read, evidence,
// journal, broker or activation capability.
type StrategyReadBudget struct {
	Market                  string
	EvidenceSource          string
	EndpointSource          string
	Endpoint                string
	Query                   RequiredQuery
	StaleAfter              time.Duration
	Retry                   RetryPolicy
	SoakMaxRequestsPerCycle int
}

// StrategyReadBudgets returns independent KR identity and US FX read budgets.
// A fresh map is returned so a diagnostic/test caller cannot rewrite the next
// evaluation's policy. Production adapters must still validate source evidence
// freshness at the consuming clock; this catalog is an operational upper bound,
// not authority to mint or extend evidence.
func StrategyReadBudgets() map[RequiredQuery]StrategyReadBudget {
	retry := DefaultRetryPolicy()
	return map[RequiredQuery]StrategyReadBudget{
		QueryAccountIdentity: {
			Market: "KR", EvidenceSource: "same-currency", EndpointSource: "official-open-api",
			Endpoint: "GET /api/v1/accounts", Query: QueryAccountIdentity,
			StaleAfter: 5 * time.Minute, Retry: retry, SoakMaxRequestsPerCycle: 1,
		},
		QueryExchangeRate: {
			Market: "US", EvidenceSource: "official-fx", EndpointSource: "official-open-api",
			Endpoint: "GET /api/v1/exchange-rate", Query: QueryExchangeRate,
			StaleAfter: 15 * time.Second, Retry: retry, SoakMaxRequestsPerCycle: 1,
		},
	}
}
