package trading

// conditional.go owns conditional orders (stop / OCO / OTO): the write gate, the
// intent assembly, and the Service methods that combine them with a broker call.
//
// All three used to live in cmd/tossctl — the gate in conditional_gate.go and the
// intent assembly inline in the cobra RunE bodies. That made cobra the only
// caller able to enforce the policy, so a second surface (MCP, or TossOS's
// trading engine) would have had to re-implement the same three checks from
// scratch. Issue #111 is the record of exactly that going wrong for
// order-lineage recording, and the lesson applied here: write policy and the
// bookkeeping around it belong to the module that owns writes.
//
// Conditional orders keep their own gate rather than reusing Service.guard: the
// per-action toggles guard is driven by Action (place/cancel/amend), while every
// conditional verb is gated by the single trading.conditional toggle, and the
// caller distinguishes "show a preview" from "refused" by sentinel rather than by
// inspecting messages.

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// ErrConditionalPreviewOnly is returned when the gate passes policy but the
// caller did not ask to execute. It is a signal, not a failure: the caller
// renders a preview and the confirm token.
var ErrConditionalPreviewOnly = errors.New("preview only")

// ErrConditionalDisabled is returned when config does not permit live
// conditional-order actions. Distinct from ErrConditionalPreviewOnly so a
// refusal can never be mistaken for a preview.
var ErrConditionalDisabled = errors.New(
	"conditional live actions are disabled; set `trading.conditional=true` and `trading.allow_live_order_actions=true` in config.json")

// ErrConditionalConfirmMismatch is returned when --confirm does not match the
// canonical intent's token.
var ErrConditionalConfirmMismatch = errors.New(
	"confirmation token mismatch; rerun without --execute to get the current --confirm token")

// ConditionalBroker is the conditional-order write surface. Both wiring profiles
// satisfy it: *hybrid.Client for the CLI (official-only reads/writes with a
// clear error when no key is configured) and the engine's official-only broker.
type ConditionalBroker interface {
	CreateConditionalOrder(ctx context.Context, intent orderintent.ConditionalPlaceIntent) (domain.ConditionalOrderRef, error)
	CancelConditionalOrder(ctx context.Context, intent orderintent.ConditionalCancelIntent) error
	ModifyConditionalOrder(ctx context.Context, intent orderintent.ConditionalModifyIntent) error
}

// WithConditional attaches the conditional-order broker. Optional, like
// WithLineage: a Service without one previews normally and returns
// ErrLiveMutationPending instead of mutating, which keeps read-only callers and
// the many tests that build a bare Service working.
func (s *Service) WithConditional(b ConditionalBroker) *Service {
	s.conditional = b
	return s
}

// ConditionalGate enforces the conditional-order write policy: config must
// enable conditional live actions, execute must be set, and confirm must match
// the canonical confirm token.
//
// Returns nil when the mutation may proceed, ErrConditionalPreviewOnly when the
// caller should render a preview instead, and a refusal error otherwise.
func ConditionalGate(policy config.Trading, canonical string, execute bool, confirm string) error {
	if !policy.Conditional || !policy.AllowLiveOrderActions {
		return ErrConditionalDisabled
	}
	if !execute {
		return ErrConditionalPreviewOnly
	}
	token := orderintent.ConfirmToken(canonical)
	if subtle.ConstantTimeCompare([]byte(confirm), []byte(token)) != 1 {
		return ErrConditionalConfirmMismatch
	}
	return nil
}

// --- previews ---------------------------------------------------------------

func (s *Service) PreviewConditionalPlace(intent orderintent.ConditionalPlaceIntent) Preview {
	return s.conditionalPreview("conditional_place", orderintent.CanonicalConditionalPlace(intent))
}

func (s *Service) PreviewConditionalCancel(intent orderintent.ConditionalCancelIntent) Preview {
	return s.conditionalPreview("conditional_cancel", orderintent.CanonicalConditionalCancel(intent))
}

func (s *Service) PreviewConditionalModify(intent orderintent.ConditionalModifyIntent) Preview {
	return s.conditionalPreview("conditional_modify", orderintent.CanonicalConditionalModify(intent))
}

// conditionalPreview reports what the gate would decide, and the token that
// would open it. The token is derived from the same canonical string the gate
// compares against, so a preview can never advertise a token the gate rejects.
func (s *Service) conditionalPreview(kind, canonical string) Preview {
	var warnings []string
	if !s.policy.Conditional {
		warnings = append(warnings, "Config currently disables conditional order actions.")
	}
	if !s.policy.AllowLiveOrderActions {
		warnings = append(warnings, "Config currently disables live order actions.")
	}
	return Preview{
		Kind:          kind,
		Canonical:     canonical,
		ConfirmToken:  orderintent.ConfirmToken(canonical),
		Warnings:      warnings,
		LiveReady:     true,
		MutationReady: s.policy.Conditional && s.policy.AllowLiveOrderActions,
	}
}

// --- execution --------------------------------------------------------------

// ConditionalPlace creates a conditional order once the gate permits it.
func (s *Service) ConditionalPlace(ctx context.Context, intent orderintent.ConditionalPlaceIntent, opts ExecuteOptions) (domain.ConditionalOrderRef, error) {
	if err := ConditionalGate(s.policy, orderintent.CanonicalConditionalPlace(intent), opts.Execute, opts.Confirm); err != nil {
		return domain.ConditionalOrderRef{}, err
	}
	if s.conditional == nil {
		return domain.ConditionalOrderRef{}, ErrLiveMutationPending
	}
	return s.conditional.CreateConditionalOrder(ctx, intent)
}

// ConditionalCancel cancels a conditional order once the gate permits it.
func (s *Service) ConditionalCancel(ctx context.Context, intent orderintent.ConditionalCancelIntent, opts ExecuteOptions) error {
	if err := ConditionalGate(s.policy, orderintent.CanonicalConditionalCancel(intent), opts.Execute, opts.Confirm); err != nil {
		return err
	}
	if s.conditional == nil {
		return ErrLiveMutationPending
	}
	return s.conditional.CancelConditionalOrder(ctx, intent)
}

// ConditionalModify modifies a conditional order once the gate permits it.
func (s *Service) ConditionalModify(ctx context.Context, intent orderintent.ConditionalModifyIntent, opts ExecuteOptions) error {
	if err := ConditionalGate(s.policy, orderintent.CanonicalConditionalModify(intent), opts.Execute, opts.Confirm); err != nil {
		return err
	}
	if s.conditional == nil {
		return ErrLiveMutationPending
	}
	return s.conditional.ModifyConditionalOrder(ctx, intent)
}

// --- intent assembly --------------------------------------------------------

// ConditionalPlaceInput is the flat, flag-shaped input a surface collects for a
// conditional place. Assembling the intent here (rather than in each surface)
// keeps the OCO/OTO second-leg requirement in one place: a two-leg order type
// with no second leg must be refused before a confirm token is ever computed,
// or the user would be handed a token for an order the broker will reject.
type ConditionalPlaceInput struct {
	Symbol           string
	Type             string // SINGLE | OCO | OTO
	OrderType        string // LIMIT | MARKET
	ExpireDate       string
	Quantity         float64
	ClientOrderID    string
	ConfirmHighValue bool

	FirstSide       string
	FirstTrigger    float64
	FirstOrderPrice float64

	SecondSide       string
	SecondTrigger    float64
	SecondOrderPrice float64
}

// ConditionalModifyInput is the same for a modify, keyed by conditional order id.
type ConditionalModifyInput struct {
	ID               string
	Type             string
	OrderType        string
	ExpireDate       string
	Quantity         float64
	ConfirmHighValue bool

	FirstSide       string
	FirstTrigger    float64
	FirstOrderPrice float64

	SecondSide       string
	SecondTrigger    float64
	SecondOrderPrice float64
}

// NewConditionalPlaceIntent assembles a place intent, refusing a two-leg type
// without a second leg.
func NewConditionalPlaceIntent(in ConditionalPlaceInput) (orderintent.ConditionalPlaceIntent, error) {
	intent := orderintent.ConditionalPlaceIntent{
		Symbol:           in.Symbol,
		Type:             in.Type,
		OrderType:        in.OrderType,
		ExpireDate:       in.ExpireDate,
		Quantity:         in.Quantity,
		ClientOrderID:    in.ClientOrderID,
		ConfirmHighValue: in.ConfirmHighValue,
		First:            orderintent.ConditionLeg{OrderSide: in.FirstSide, TriggerPrice: in.FirstTrigger, OrderPrice: in.FirstOrderPrice},
	}
	second, err := secondLeg(in.Type, in.SecondSide, in.SecondTrigger, in.SecondOrderPrice)
	if err != nil {
		return orderintent.ConditionalPlaceIntent{}, err
	}
	intent.Second = second
	return intent, nil
}

// NewConditionalModifyIntent assembles a modify intent under the same rule.
func NewConditionalModifyIntent(in ConditionalModifyInput) (orderintent.ConditionalModifyIntent, error) {
	intent := orderintent.ConditionalModifyIntent{
		ID:               in.ID,
		Type:             in.Type,
		OrderType:        in.OrderType,
		ExpireDate:       in.ExpireDate,
		Quantity:         in.Quantity,
		ConfirmHighValue: in.ConfirmHighValue,
		First:            orderintent.ConditionLeg{OrderSide: in.FirstSide, TriggerPrice: in.FirstTrigger, OrderPrice: in.FirstOrderPrice},
	}
	second, err := secondLeg(in.Type, in.SecondSide, in.SecondTrigger, in.SecondOrderPrice)
	if err != nil {
		return orderintent.ConditionalModifyIntent{}, err
	}
	intent.Second = second
	return intent, nil
}

// secondLeg returns the second condition leg for two-leg order types, or nil for
// SINGLE. The error text names the flags a CLI user must add, which is what the
// cobra command used to print.
func secondLeg(orderType, side string, trigger, orderPrice float64) (*orderintent.ConditionLeg, error) {
	if orderType != "OCO" && orderType != "OTO" {
		return nil, nil
	}
	if side == "" {
		return nil, fmt.Errorf("--second-side/--second-trigger required for %s", orderType)
	}
	return &orderintent.ConditionLeg{OrderSide: side, TriggerPrice: trigger, OrderPrice: orderPrice}, nil
}
