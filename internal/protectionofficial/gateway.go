// Package protectionofficial is the only broker adapter for resident protection.
// It depends exclusively on internal/official and has no WTS or hybrid fallback.
// No app or command package constructs it in this change, so runtime activation
// remains OFF/UNWIRED.
package protectionofficial

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
)

var ErrAmbiguousConditional = errors.New("protection official: conditional state is ambiguous")

type client interface {
	CreateConditionalOrder(context.Context, official.ConditionalCreateBody) (domain.ConditionalOrderRef, error)
	ModifyConditionalOrderRef(context.Context, string, official.ConditionalModifyBody) (domain.ConditionalOrderRef, error)
	CancelConditionalOrder(context.Context, string) error
	ConditionalOrderRaw(context.Context, string) (official.RawConditionalOrder, error)
	ProtectionConditionalOrdersRaw(context.Context, string, string, string, int) (official.RawConditionalOrderList, error)
	SellableQuantityRaw(context.Context, string) (string, time.Time, error)
}

type Gateway struct {
	client client
	scope  protection.Scope
	now    func() time.Time
}

func New(client client, scope protection.Scope, now func() time.Time) (*Gateway, error) {
	if client == nil || scope.Validate() != nil || now == nil {
		return nil, errors.New("protection official: client, exact scope, and clock are required")
	}
	return &Gateway{client: client, scope: scope, now: now}, nil
}

func (g *Gateway) Create(ctx context.Context, body protection.ConditionalBody) (protection.BrokerProtection, error) {
	if err := g.validateBody(body); err != nil {
		return protection.BrokerProtection{}, err
	}
	ref, err := g.client.CreateConditionalOrder(ctx, official.ConditionalCreateBody{
		Symbol: body.Symbol, Type: body.ConditionalType, Quantity: strconv.FormatInt(body.Quantity, 10),
		OrderType: body.OrderType, ClientOrderID: body.ClientOrderID, ExpireDate: body.ExpireDate,
		First: official.ConditionLegBody{OrderSide: body.Side, TriggerPrice: strconv.FormatInt(body.Trigger, 10)},
	})
	if err != nil {
		return protection.BrokerProtection{}, err
	}
	return g.confirm(ctx, ref.ID, body.ClientOrderID)
}

func (g *Gateway) Replace(ctx context.Context, id string, body protection.ConditionalBody) (protection.BrokerProtection, error) {
	if strings.TrimSpace(id) == "" {
		return protection.BrokerProtection{}, ErrAmbiguousConditional
	}
	if err := g.validateBody(body); err != nil {
		return protection.BrokerProtection{}, err
	}
	ref, err := g.client.ModifyConditionalOrderRef(ctx, id, official.ConditionalModifyBody{
		Type: body.ConditionalType, Quantity: strconv.FormatInt(body.Quantity, 10),
		OrderType: body.OrderType, ExpireDate: body.ExpireDate,
		First: official.ConditionLegBody{OrderSide: body.Side, TriggerPrice: strconv.FormatInt(body.Trigger, 10)},
	})
	if err != nil {
		return protection.BrokerProtection{}, err
	}
	return g.confirm(ctx, ref.ID, body.ClientOrderID)
}

func (g *Gateway) Cancel(ctx context.Context, id string) (protection.CancelObservation, error) {
	if strings.TrimSpace(id) == "" {
		return protection.CancelObservation{}, ErrAmbiguousConditional
	}
	if err := g.client.CancelConditionalOrder(ctx, id); err != nil {
		return protection.CancelObservation{}, err
	}
	// A successful DELETE alone is not terminal evidence: the measured API also
	// drops an order that raced to trigger. Require an authoritative read/list
	// state and fail closed when the identifier simply disappears.
	raw, err := g.client.ConditionalOrderRaw(ctx, id)
	if err == nil {
		return cancelObservation(g.scope, raw, g.now()), nil
	}
	for _, group := range []string{"OPEN", "CLOSED"} {
		cursor := ""
		for page := 0; page < 10; page++ {
			result, listErr := g.client.ProtectionConditionalOrdersRaw(ctx, group, g.scope.Symbol, cursor, 100)
			if listErr != nil {
				return protection.CancelObservation{}, fmt.Errorf("%w: post-cancel %s read: %v", ErrAmbiguousConditional, group, listErr)
			}
			for _, item := range result.Orders {
				if item.ID == id {
					return cancelObservation(g.scope, item, g.now()), nil
				}
			}
			if !result.HasNext || result.NextCursor == "" {
				break
			}
			cursor = result.NextCursor
		}
	}
	return protection.CancelObservation{}, fmt.Errorf("%w: %s disappeared after cancel", ErrAmbiguousConditional, id)
}

func (g *Gateway) Get(ctx context.Context, id string) (protection.BrokerProtection, error) {
	return g.confirm(ctx, id, "")
}

func (g *Gateway) List(ctx context.Context, scope protection.Scope) ([]protection.BrokerProtection, error) {
	if scope != g.scope {
		return nil, protection.ErrMixedScope
	}
	var out []protection.BrokerProtection
	for _, status := range []string{"OPEN", "CLOSED"} {
		cursor := ""
		complete := false
		for page := 0; page < 10; page++ {
			result, err := g.client.ProtectionConditionalOrdersRaw(ctx, status, scope.Symbol, cursor, 100)
			if err != nil {
				return nil, err
			}
			for _, raw := range result.Orders {
				parsed, err := g.adapt(raw)
				if err != nil {
					return nil, err
				}
				out = append(out, parsed)
			}
			if !result.HasNext || result.NextCursor == "" {
				complete = true
				break
			}
			cursor = result.NextCursor
		}
		if !complete {
			return nil, fmt.Errorf("%w: conditional %s pagination exceeded 10 pages", ErrAmbiguousConditional, status)
		}
	}
	return out, nil
}

func (g *Gateway) Sellable(ctx context.Context, scope protection.Scope, brokerID string) (protection.SellableObservation, error) {
	if scope != g.scope || strings.TrimSpace(brokerID) == "" {
		return protection.SellableObservation{}, protection.ErrMixedScope
	}
	raw, at, err := g.client.SellableQuantityRaw(ctx, scope.Symbol)
	if err != nil {
		return protection.SellableObservation{}, err
	}
	quantity, err := positiveOrZeroInteger(raw)
	if err != nil {
		return protection.SellableObservation{}, err
	}
	return protection.SellableObservation{Scope: scope, BrokerID: brokerID, Quantity: quantity, At: at}, nil
}

func (g *Gateway) validateBody(body protection.ConditionalBody) error {
	if _, err := body.CanonicalJSON(); err != nil {
		return err
	}
	if body.AccountRef != g.scope.AccountRef || body.Market != string(g.scope.Market) || body.Symbol != g.scope.Symbol {
		return protection.ErrMixedScope
	}
	return nil
}

func (g *Gateway) confirm(ctx context.Context, id, clientOrderID string) (protection.BrokerProtection, error) {
	if strings.TrimSpace(id) == "" {
		return protection.BrokerProtection{}, ErrAmbiguousConditional
	}
	raw, err := g.client.ConditionalOrderRaw(ctx, id)
	if err != nil {
		return protection.BrokerProtection{}, err
	}
	if clientOrderID != "" && raw.ClientOrderID != clientOrderID {
		return protection.BrokerProtection{}, ErrAmbiguousConditional
	}
	return g.adapt(raw)
}

func (g *Gateway) adapt(raw official.RawConditionalOrder) (protection.BrokerProtection, error) {
	quantity, err := positiveInteger(raw.Quantity)
	if err != nil {
		return protection.BrokerProtection{}, err
	}
	trigger, err := positiveInteger(raw.TriggerPrice)
	if err != nil {
		return protection.BrokerProtection{}, err
	}
	if raw.Symbol != g.scope.Symbol || raw.Market != string(g.scope.Market) || raw.Type != "SINGLE" || raw.OrderType != "MARKET" || raw.ConditionType != "STOP" {
		return protection.BrokerProtection{}, ErrAmbiguousConditional
	}
	status := strings.ToUpper(strings.TrimSpace(raw.Status))
	triggered := strings.TrimSpace(raw.TriggeredOrderID) != ""
	terminal, err := lifecycle(status, triggered)
	if err != nil {
		return protection.BrokerProtection{}, err
	}
	return protection.BrokerProtection{Scope: g.scope, ID: raw.ID, ClientOrderID: raw.ClientOrderID,
		Quantity: quantity, Trigger: trigger, Terminal: terminal, Triggered: triggered}, nil
}

func cancelObservation(scope protection.Scope, raw official.RawConditionalOrder, at time.Time) protection.CancelObservation {
	status := strings.ToUpper(strings.TrimSpace(raw.Status))
	triggered := strings.TrimSpace(raw.TriggeredOrderID) != ""
	terminal, _ := lifecycle(status, triggered)
	return protection.CancelObservation{Scope: scope, BrokerID: raw.ID, Terminal: terminal, Triggered: triggered, At: at}
}

func lifecycle(status string, triggered bool) (bool, error) {
	switch status {
	case "WATCHING", "PAUSED", "ORDERING", "ORDERED":
		return triggered, nil
	case "COMPLETED", "EXPIRED", "CANCELLED", "CANCELED":
		return true, nil
	default:
		return false, fmt.Errorf("%w: unknown lifecycle %q", ErrAmbiguousConditional, status)
	}
}

func positiveInteger(raw string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 1 || strconv.FormatInt(value, 10) != raw {
		return 0, fmt.Errorf("%w: non-canonical positive integer %q", ErrAmbiguousConditional, raw)
	}
	return value, nil
}

func positiveOrZeroInteger(raw string) (int64, error) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 || strconv.FormatInt(value, 10) != raw {
		return 0, fmt.Errorf("%w: non-canonical nonnegative integer %q", ErrAmbiguousConditional, raw)
	}
	return value, nil
}

var _ protection.ExecutionGateway = (*Gateway)(nil)
var _ client = (*official.Client)(nil)
