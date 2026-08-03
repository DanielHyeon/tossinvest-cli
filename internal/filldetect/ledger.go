package filldetect

// ledger.go binds the detector to the journal's fill tables
// (harden-execution-base task 3.2).
//
// It is deliberately thin. Every decision that matters — is this a new fill, is
// it a contradiction, has it been seen before — is made inside the journal's
// transaction, because the answer depends on the stored prior and two cycles must
// not be able to read the same prior and each claim "their" delta. What is left
// here is translation: a Snapshot into what the table stores, and a FillResult
// back into what the entry gate understands.

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// JournalLedger is the durable Ledger.
type JournalLedger struct {
	Journal    *journal.Journal
	AccountRef string
	// Refresh republishes journal RECONCILE authority after the fill transaction
	// commits. Production wires Tracker.Refresh; nil is allowed for isolated
	// ledger tests.
	Refresh func(context.Context) error
}

// Apply records one cumulative snapshot.
func (l JournalLedger) Apply(ctx context.Context, snap Snapshot) (Applied, error) {
	if l.Journal == nil {
		return Applied{}, fmt.Errorf("filldetect: the ledger has no journal")
	}

	obs := journal.FillObservation{
		OrderID:        snap.OrderID,
		Symbol:         snap.Symbol,
		Market:         snap.Market,
		TradingDay:     snap.TradingDay,
		Side:           snap.Side,
		State:          string(snap.Derived.State),
		Terminal:       snap.Derived.Terminal,
		FailClosed:     snap.Derived.FailClosed,
		Reason:         string(snap.Derived.Reason),
		Detail:         snap.Derived.Detail,
		Quantity:       decimalString(snap.Quantity),
		FilledQuantity: decimalString(snap.FilledQuantity),
		AveragePrice:   snap.AveragePrice,
		FilledAmount:   snap.FilledAmount,
		AccountRef:     strings.TrimSpace(l.AccountRef),
		ObservedAt:     journal.RFC3339(snap.ObservedAt),
	}
	// Only a timestamp the *broker* supplied is passed through. The detector's
	// fallback (the previous poll's instant) is a measurement aid, and letting it
	// into the ordering check would make a real, older execution.filledAt look
	// like an out-of-order snapshot and fail closed on a healthy account.
	if snap.HasBrokerTime {
		obs.BrokerVisibleAt = journal.RFC3339(snap.BrokerVisibleAt)
	}

	res, err := l.Journal.RecordFill(ctx, obs)
	if err != nil {
		return Applied{}, err
	}
	if l.Refresh != nil {
		if err := l.Refresh(ctx); err != nil {
			return Applied{}, fmt.Errorf("filldetect: refreshing RECONCILE authority after the fill: %w", err)
		}
	}

	applied := Applied{
		OrderID:    res.OrderID,
		Delta:      res.DeltaQuantity,
		Changed:    res.Changed,
		Corrected:  res.Corrected,
		FailClosed: res.FailClosed,
	}
	if res.FailClosed {
		// The gate's vocabulary stays the stable one; the specific rule that
		// fired travels in the detail an operator reads.
		applied.Reason = execgw.ReasonBrokerStateUnknown
		applied.Detail = res.Reason
		if res.Detail != "" {
			applied.Detail = res.Reason + ": " + res.Detail
		}
	}
	if ts, perr := time.Parse(time.RFC3339, res.CommittedAt); perr == nil {
		applied.CommittedAt = ts.UTC()
	}
	return applied, nil
}

// JournalTracked is the durable TrackedSource: the orders that are not terminal,
// plus the ones a confirmed attempt named and no poll has seen yet.
type JournalTracked struct {
	Journal    *journal.Journal
	AccountRef string
}

// TrackedOrders lists what the next cycle must read by id.
func (t JournalTracked) TrackedOrders(ctx context.Context) ([]TrackedOrder, error) {
	if t.Journal == nil {
		return nil, fmt.Errorf("filldetect: the tracked-order source has no journal")
	}
	if strings.TrimSpace(t.AccountRef) == "" {
		return nil, fmt.Errorf("filldetect: the tracked-order source has no selected account")
	}
	rows, err := t.Journal.TrackedFillOrders(ctx, t.AccountRef)
	if err != nil {
		return nil, err
	}
	out := make([]TrackedOrder, 0, len(rows))
	for _, row := range rows {
		out = append(out, TrackedOrder{
			OrderID: row.OrderID,
			Symbol:  row.Symbol,
			Market:  row.Market,
			Lineage: brokerstate.Lineage{SuccessorOrderID: row.SuccessorOrderID},
		})
	}
	return out, nil
}

// Compile-time proof that the durable implementations satisfy what the detector
// asks for, so a signature drift fails the build here rather than at wiring time.
var (
	_ Ledger        = JournalLedger{}
	_ TrackedSource = JournalTracked{}
)

func decimalString(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }
