package domain

// MutationResult is the outcome of an order mutation (place / cancel / amend)
// as reported back to a caller.
//
// It lives in domain rather than in internal/trading because it is a value
// object shared by every layer that touches an order: the WTS client and the
// official Open API client produce it, the hybrid router and trading policy
// service pass it through, and output/MCP serialize it. Engine-side packages
// (journal, execution gateway) need to name it too, and none of them should
// have to depend on the CLI's trading-policy module to do so.
//
// internal/trading keeps `type MutationResult = domain.MutationResult` so all
// existing references stay valid.
//
// The json tags are a public contract: `tossctl order ... --output json` and the
// MCP order tools serialize this struct verbatim.
type MutationResult struct {
	Kind                  string   `json:"kind"`
	Status                string   `json:"status"`
	OrderID               string   `json:"order_id,omitempty"`
	OriginalOrderID       string   `json:"original_order_id,omitempty"`
	CurrentOrderID        string   `json:"current_order_id,omitempty"`
	Symbol                string   `json:"symbol,omitempty"`
	Market                string   `json:"market,omitempty"`
	Quantity              float64  `json:"quantity,omitempty"`
	FilledQuantity        float64  `json:"filled_quantity,omitempty"`
	Price                 float64  `json:"price,omitempty"`
	AverageExecutionPrice float64  `json:"average_execution_price,omitempty"`
	OrderDate             string   `json:"order_date,omitempty"`
	Warnings              []string `json:"warnings,omitempty"`
}
