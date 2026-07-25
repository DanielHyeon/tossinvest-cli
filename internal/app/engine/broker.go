package engine

import (
	"context"
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// officialBroker adapts *official.Client to trading.Broker (and to
// ConditionalMutator). It is the engine's only order path.
//
// Compare hybrid.hybridBroker, which decides per call whether to use the
// official API or the web session: there is no decision here, and no branch that
// could take one. That absence is the safety property.
type officialBroker struct {
	off *official.Client
}

func (b *officialBroker) PlacePendingOrder(ctx context.Context, intent orderintent.PlaceIntent) (trading.MutationResult, error) {
	return b.off.PlaceOrder(ctx, intent)
}

func (b *officialBroker) CancelPendingOrder(ctx context.Context, intent orderintent.CancelIntent) (trading.MutationResult, error) {
	return b.off.CancelOrder(ctx, intent.OrderID)
}

func (b *officialBroker) AmendPendingOrder(ctx context.Context, intent orderintent.AmendIntent) (trading.MutationResult, error) {
	return b.off.ModifyOrder(ctx, intent)
}

// GetOrderAvailableActions is the pre-cancel/pre-amend check trading.Service
// performs before it mutates. The official Open API has no equivalent endpoint,
// and the hybrid broker answers it from the web session — which would make every
// engine cancel depend on a live WTS session (engine-safety: "WTS 세션이 없거나
// 만료된 상태에서도 엔진의 cancel/amend는 동작해야 한다").
//
// So the engine treats the pre-check as broker-optional, which the spec allows,
// and returns no actions without spending an API call. Nothing is weakened by
// this: an order that cannot be cancelled is rejected by the cancel call itself,
// which is the authoritative answer anyway. Deriving real pre-check state from
// OrderByID is task 4.1 — deferred because it adds a mandatory API call per
// mutation and §0.4 requires that to be accounted for in the retry-matrix
// budget (task 2.6) first.
func (b *officialBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

// --- conditional orders -----------------------------------------------------
//
// The official conditional endpoints take decimal strings, so the intent →
// request-body mapping mirrors hybrid's (internal/hybrid/client.go): same field
// mapping, same "orderPrice omitted when 0 means MARKET" rule.

func (b *officialBroker) CreateConditionalOrder(ctx context.Context, intent orderintent.ConditionalPlaceIntent) (domain.ConditionalOrderRef, error) {
	body := official.ConditionalCreateBody{
		Symbol:                intent.Symbol,
		Type:                  intent.Type,
		Quantity:              decimalString(intent.Quantity),
		OrderType:             intent.OrderType,
		ClientOrderID:         intent.ClientOrderID,
		ExpireDate:            intent.ExpireDate,
		First:                 conditionLegBody(intent.First),
		ConfirmHighValueOrder: intent.ConfirmHighValue,
	}
	if intent.Second != nil {
		second := conditionLegBody(*intent.Second)
		body.Second = &second
	}
	return b.off.CreateConditionalOrder(ctx, body)
}

func (b *officialBroker) CancelConditionalOrder(ctx context.Context, intent orderintent.ConditionalCancelIntent) error {
	return b.off.CancelConditionalOrder(ctx, intent.ID)
}

func (b *officialBroker) ModifyConditionalOrder(ctx context.Context, intent orderintent.ConditionalModifyIntent) error {
	body := official.ConditionalModifyBody{
		Type:                  intent.Type,
		Quantity:              decimalString(intent.Quantity),
		OrderType:             intent.OrderType,
		ExpireDate:            intent.ExpireDate,
		First:                 conditionLegBody(intent.First),
		ConfirmHighValueOrder: intent.ConfirmHighValue,
	}
	if intent.Second != nil {
		second := conditionLegBody(*intent.Second)
		body.Second = &second
	}
	return b.off.ModifyConditionalOrder(ctx, intent.ID, body)
}

func conditionLegBody(leg orderintent.ConditionLeg) official.ConditionLegBody {
	body := official.ConditionLegBody{
		OrderSide:    leg.OrderSide,
		TriggerPrice: decimalString(leg.TriggerPrice),
	}
	if leg.OrderPrice > 0 {
		body.OrderPrice = decimalString(leg.OrderPrice)
	}
	return body
}

func decimalString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
