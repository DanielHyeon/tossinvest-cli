package positioncampaign

import (
	"fmt"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

type OrderWatermark struct {
	OrderID       string
	PredecessorID string
	CarryBaseline string
	RequestedCap  string
	Cumulative    string
	Remaining     string
	Terminal      bool
}

type LegLedger struct {
	Requested string
	Filled    string
	Residual  string
	Orders    map[string]*OrderWatermark
	Reconcile bool
}

type OrderObservation struct {
	OrderID          string
	Cumulative       string
	LineageAmbiguous bool
}

type ObservationResult struct {
	Delta        string
	LateTerminal bool
	CapExceeded  bool
	Reconcile    bool
}

func NewLegLedger(requested string) (*LegLedger, error) {
	requested, err := positiveDecimal(requested)
	if err != nil {
		return nil, fmt.Errorf("requested quantity: %w", err)
	}
	return &LegLedger{Requested: requested, Residual: requested, Filled: "0", Orders: map[string]*OrderWatermark{}}, nil
}

func (l *LegLedger) LinkOrder(orderID, predecessorID, requestedCap string) error {
	if l == nil {
		return fmt.Errorf("position campaign: nil leg ledger")
	}
	orderID = strings.TrimSpace(orderID)
	predecessorID = strings.TrimSpace(predecessorID)
	if orderID == "" {
		return fmt.Errorf("position campaign: empty order id")
	}
	cap, err := positiveDecimal(requestedCap)
	if err != nil {
		return fmt.Errorf("order %s cap: %w", orderID, err)
	}
	if existing := l.Orders[orderID]; existing != nil {
		if existing.PredecessorID == predecessorID && existing.RequestedCap == cap {
			return nil
		}
		return fmt.Errorf("position campaign: immutable order identity %s was relinked", orderID)
	}
	carry := "0"
	if predecessorID != "" {
		predecessor := l.Orders[predecessorID]
		if predecessor == nil {
			return fmt.Errorf("position campaign: predecessor %s is unknown", predecessorID)
		}
		carry = predecessor.Cumulative
	}
	remaining, err := riskcalc.SubDecimal(l.Requested, l.Filled)
	if err != nil {
		return err
	}
	remaining, err = riskcalc.MaxDecimal("0", remaining)
	if err != nil {
		return err
	}
	l.Orders[orderID] = &OrderWatermark{
		OrderID: orderID, PredecessorID: predecessorID, CarryBaseline: carry,
		RequestedCap: cap, Cumulative: "0", Remaining: remaining,
	}
	return nil
}

func (l *LegLedger) MarkTerminal(orderID string) error {
	order := l.Orders[strings.TrimSpace(orderID)]
	if order == nil {
		return fmt.Errorf("position campaign: order %s is unknown", orderID)
	}
	order.Terminal = true
	return nil
}

// Observe advances only an immutable order identity's cumulative watermark.
// Lower observations cannot retreat it. A late terminal or over-cap fill is
// preserved and latches reconciliation instead of being truncated.
func (l *LegLedger) Observe(obs OrderObservation) (ObservationResult, error) {
	order := l.Orders[strings.TrimSpace(obs.OrderID)]
	if order == nil {
		return ObservationResult{}, fmt.Errorf("position campaign: order %s is unknown", obs.OrderID)
	}
	cumulative, err := nonNegative(obs.Cumulative)
	if err != nil {
		return ObservationResult{}, fmt.Errorf("order %s cumulative: %w", obs.OrderID, err)
	}
	cmp, err := riskcalc.CompareDecimal(cumulative, order.Cumulative)
	if err != nil {
		return ObservationResult{}, err
	}
	if cmp <= 0 {
		return ObservationResult{Delta: "0", Reconcile: l.Reconcile}, nil
	}
	delta, err := riskcalc.SubDecimal(cumulative, order.Cumulative)
	if err != nil {
		return ObservationResult{}, err
	}
	capCmp, err := riskcalc.CompareDecimal(cumulative, order.RequestedCap)
	if err != nil {
		return ObservationResult{}, err
	}
	result := ObservationResult{
		Delta: delta, LateTerminal: order.Terminal, CapExceeded: capCmp > 0,
	}
	newFilled, err := riskcalc.AddDecimal(l.Filled, delta)
	if err != nil {
		return ObservationResult{}, err
	}
	newResidual, err := riskcalc.SubDecimal(l.Requested, newFilled)
	if err != nil {
		return ObservationResult{}, err
	}
	newResidual, err = riskcalc.MaxDecimal("0", newResidual)
	if err != nil {
		return ObservationResult{}, err
	}
	aggregateCmp, err := riskcalc.CompareDecimal(newFilled, l.Requested)
	if err != nil {
		return ObservationResult{}, err
	}
	// Commit the computed transition only after every decimal operation has
	// succeeded. A malformed in-memory snapshot cannot consume the watermark and
	// turn a retry into delta zero.
	order.Cumulative = cumulative
	l.Filled = newFilled
	l.Residual = newResidual
	// Every active successor derives its remaining quantity from all newly
	// known facts, including a predecessor's late fill.
	for _, item := range l.Orders {
		if item.PredecessorID == "" || item.Terminal {
			continue
		}
		item.Remaining = l.Residual
	}
	if result.LateTerminal || result.CapExceeded || obs.LineageAmbiguous || aggregateCmp > 0 {
		l.Reconcile = true
	}
	result.Reconcile = l.Reconcile
	return result, nil
}

func (l *LegLedger) OrderedWatermarks() []OrderWatermark {
	ids := make([]string, 0, len(l.Orders))
	for id := range l.Orders {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]OrderWatermark, 0, len(ids))
	for _, id := range ids {
		out = append(out, *l.Orders[id])
	}
	return out
}

func nonNegative(value string) (string, error) {
	value, err := riskcalc.CanonicalDecimal(value)
	if err != nil {
		return "", err
	}
	negative, err := riskcalc.IsNegativeDecimal(value)
	if err != nil {
		return "", err
	}
	if negative {
		return "", fmt.Errorf("negative decimal %s", value)
	}
	return value, nil
}
