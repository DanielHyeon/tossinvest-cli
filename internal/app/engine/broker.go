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
// so the engine derives the answer from OrderRawByID + the brokerstate priority
// table (task 4.1). The whole implementation, and the reason it cannot return an
// error, is in precheck.go.
func (b *officialBroker) GetOrderAvailableActions(ctx context.Context, orderID string) (map[string]any, error) {
	return b.availableActions(ctx, orderID), nil
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
