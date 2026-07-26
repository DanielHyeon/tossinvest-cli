package official

import (
	"context"
	"net/url"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// CancelConditionalOrder cancels a conditional order by id (account-scoped).
func (c *Client) CancelConditionalOrder(ctx context.Context, id string) error {
	return c.deleteAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id), nil)
}

// ConditionLegBody mirrors ConditionRequest (string decimals; orderPrice
// omitted for MARKET). Exported so the hybrid layer can construct it.
type ConditionLegBody struct {
	OrderSide    string `json:"orderSide"`
	TriggerPrice string `json:"triggerPrice"`
	OrderPrice   string `json:"orderPrice,omitempty"`
}

// ConditionalCreateBody mirrors ConditionalOrderCreateRequest.
type ConditionalCreateBody struct {
	Symbol                string            `json:"symbol"`
	Type                  string            `json:"type"`
	Quantity              string            `json:"quantity"`
	OrderType             string            `json:"orderType"`
	ClientOrderID         string            `json:"clientOrderId,omitempty"`
	ExpireDate            string            `json:"expireDate"`
	First                 ConditionLegBody  `json:"first"`
	Second                *ConditionLegBody `json:"second,omitempty"`
	ConfirmHighValueOrder bool              `json:"confirmHighValueOrder"`
}

// apiConditionalCreateResponse mirrors ConditionalOrderCreateResponse (unwrapped result).
type apiConditionalCreateResponse struct {
	ConditionalOrderID string `json:"conditionalOrderId"`
	ClientOrderID      string `json:"clientOrderId"`
}

// CreateConditionalOrder creates a conditional order (account-scoped, POST).
func (c *Client) CreateConditionalOrder(ctx context.Context, body ConditionalCreateBody) (domain.ConditionalOrderRef, error) {
	var raw apiConditionalCreateResponse
	if err := c.postAcct(ctx, "/api/v1/conditional-orders", body, &raw); err != nil {
		return domain.ConditionalOrderRef{}, err
	}
	return domain.ConditionalOrderRef{ID: raw.ConditionalOrderID, ClientOrderID: raw.ClientOrderID}, nil
}

// ConditionalModifyBody mirrors ConditionalOrderModifyRequest (no symbol/clientOrderId).
type ConditionalModifyBody struct {
	Type                  string            `json:"type"`
	Quantity              string            `json:"quantity"`
	OrderType             string            `json:"orderType"`
	ExpireDate            string            `json:"expireDate"`
	First                 ConditionLegBody  `json:"first"`
	Second                *ConditionLegBody `json:"second,omitempty"`
	ConfirmHighValueOrder bool              `json:"confirmHighValueOrder"`
}

// ModifyConditionalOrder modifies an existing conditional order (POST /modify).
//
// It discards the response. Callers that need the identifier the broker issues —
// see ModifyConditionalOrderRef and the note there — must use that method
// instead; this one is left decoding nothing so a null result stays a success for
// the callers that only ask whether the call worked.
func (c *Client) ModifyConditionalOrder(ctx context.Context, id string, body ConditionalModifyBody) error {
	return c.postAcct(ctx, "/api/v1/conditional-orders/"+url.PathEscape(id)+"/modify", body, nil)
}

// ModifyConditionalOrderRef modifies a conditional order and returns the
// identifier the broker issued for the result.
//
// The identifier matters because a modify is not an edit: "수정은 기존 조건주문을
// 취소하고 새 조건주문을 생성하는 방식으로 동작합니다. 따라서 수정 후에는 새로운
// conditionalOrderId 가 발급되고 기존 ID 는 무효화됩니다" (POST
// /api/v1/conditional-orders/{conditionalOrderId}/modify,
// docs/migration/openapi.latest.json). Anything that has to keep tracking the
// conditional after modifying it — cancel it, poll it, reconcile it against a
// position — is tracking the wrong object the moment it keeps the old id.
//
// A separate method rather than a change to ModifyConditionalOrder: that one has
// callers (internal/hybrid, internal/app/engine) that pass no output and would
// start failing on a broker response this build did not expect, and widening a
// live mutation's failure surface to add a return value nobody there reads is not
// a trade worth making.
func (c *Client) ModifyConditionalOrderRef(ctx context.Context, id string, body ConditionalModifyBody) (domain.ConditionalOrderRef, error) {
	var raw apiConditionalCreateResponse
	path := "/api/v1/conditional-orders/" + url.PathEscape(id) + "/modify"
	if err := c.postAcct(ctx, path, body, &raw); err != nil {
		return domain.ConditionalOrderRef{}, err
	}
	return domain.ConditionalOrderRef{ID: raw.ConditionalOrderID, ClientOrderID: raw.ClientOrderID}, nil
}
