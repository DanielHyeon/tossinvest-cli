package execgw

// amend_indoubt.go resolves CANCEL and AMEND attempts whose outcome is unknown
// (harden-execution-base task 2.8).
//
// A place in doubt is answered by looking for the order. A cancel or an amend
// cannot be, because neither creates an order that looks like the request:
//
//   - A cancel leaves no trace of its own. The question "did my cancel land" is
//     answered by the *original* order's derived state.
//   - An amend answers with a new order number. The original closes carrying its
//     fill and a successor carries the remainder, so the question is answered by
//     the original's state plus a symbol-scoped search for that successor.
//
// Both use brokerstate's priority table rather than the raw status field: CLOSED
// covers filled, cancelled, cancelled-after-partial and replaced, and the
// difference between those is the difference between flat and exposed.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/brokerstate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// resolveCancel settles a cancel whose outcome is unknown.
//
// The stabilisation rule is the same as for a place, and for the same reason: a
// single observation of "still open" may just mean the cancel is still working its
// way through the broker.
func (r *Resolver) resolveCancel(
	ctx context.Context,
	attempt *journal.Attempt,
	rec journal.AttemptRecord,
	intent journal.Intent,
	res Resolution,
	cfg ResolveConfig,
	clk clock.Clock,
) (Resolution, error) {
	if r.Order == nil {
		return r.park(ctx, attempt, res, "no single-order reader is configured, so the original order cannot be inspected")
	}
	target := strings.TrimSpace(rec.TargetOrderID)
	if target == "" {
		return r.park(ctx, attempt, res, "the cancel attempt records no target order id")
	}

	started := clk.Now()
	deadline := started.Add(cfg.MaxDuration)
	stableOpen := 0

	for {
		res.Observations++
		derived, err := r.deriveOrder(ctx, target)
		if err != nil {
			// Not evidence either way. Keep trying inside the budget.
			if !r.waitOrExpire(ctx, clk, cfg, deadline) {
				return r.park(ctx, attempt, res,
					fmt.Sprintf("order %s could not be read: %v", target, err))
			}
			continue
		}

		switch derived.State {
		case brokerstate.StateCancelled, brokerstate.StateCancelledPartiallyFilled:
			res.State = journal.StateConfirmed
			res.BrokerOrderID = target
			res.Reason = ReasonCode(journal.ReasonResolvedFound)
			res.Detail = fmt.Sprintf("order %s is %s at the broker, so the cancel took effect (%s filled)",
				target, derived.State, decimalString(derived.FilledQuantity))
			if err := attempt.ResolveConfirmed(ctx, target, journal.ReasonResolvedFound, res.Detail); err != nil {
				return res, err
			}
			return res, nil

		case brokerstate.StateFilled:
			// The order filled before the cancel could act. Definitive: the
			// cancel did not happen, and the position is real.
			res.State = journal.StateFailedConfirmed
			res.Reason = ReasonCode(journal.ReasonResolvedAbsent)
			res.Detail = fmt.Sprintf("order %s is fully filled, so the cancel did not take effect", target)
			if err := attempt.ResolveFailed(ctx, journal.ReasonResolvedAbsent, res.Detail); err != nil {
				return res, err
			}
			return res, nil

		case brokerstate.StateOpenUnfilled, brokerstate.StateOpenPartiallyFilled:
			stableOpen++
			if stableOpen >= cfg.StableObservations && clk.Now().Sub(started) >= cfg.MinObservation {
				res.State = journal.StateFailedConfirmed
				res.Reason = ReasonCode(journal.ReasonResolvedAbsent)
				res.Detail = fmt.Sprintf(
					"order %s is still open after %d observations over %s, so the cancel never landed",
					target, stableOpen, clk.Now().Sub(started).Round(time.Second))
				if err := attempt.ResolveFailed(ctx, journal.ReasonResolvedAbsent, res.Detail); err != nil {
					return res, err
				}
				return res, nil
			}
			if !r.waitOrExpire(ctx, clk, cfg, deadline) {
				return r.park(ctx, attempt, res, fmt.Sprintf(
					"order %s was still open when the %s resolution budget ran out", target, cfg.MaxDuration))
			}

		case brokerstate.StateReplaced:
			// An amendment replaced the order out from under the cancel. Whether
			// the cancel applied to the parent or was rejected is not something
			// the API tells us.
			return r.park(ctx, attempt, res, fmt.Sprintf(
				"order %s was replaced by %s while the cancel was in doubt",
				target, derived.SuccessorOrderID))

		default: // UNKNOWN_BROKER_STATE
			return r.park(ctx, attempt, res, fmt.Sprintf(
				"order %s derives as %s (%s): %s", target, derived.State, derived.Reason, derived.Detail))
		}
	}
}

// resolveAmend settles an amend whose outcome is unknown.
func (r *Resolver) resolveAmend(
	ctx context.Context,
	attempt *journal.Attempt,
	rec journal.AttemptRecord,
	intent journal.Intent,
	res Resolution,
	cfg ResolveConfig,
	clk clock.Clock,
) (Resolution, error) {
	if r.Order == nil {
		return r.park(ctx, attempt, res, "no single-order reader is configured, so the original order cannot be inspected")
	}
	target := strings.TrimSpace(rec.TargetOrderID)
	if target == "" {
		return r.park(ctx, attempt, res, "the amend attempt records no target order id")
	}
	successorMatcher, err := newMatcher(intent, rec, cfg)
	if err != nil {
		return res, err
	}

	started := clk.Now()
	deadline := started.Add(cfg.MaxDuration)
	stableOpen := 0

	for {
		res.Observations++
		derived, err := r.deriveOrder(ctx, target)
		if err != nil {
			if !r.waitOrExpire(ctx, clk, cfg, deadline) {
				return r.park(ctx, attempt, res,
					fmt.Sprintf("order %s could not be read: %v", target, err))
			}
			continue
		}

		switch derived.State {
		case brokerstate.StateOpenUnfilled, brokerstate.StateOpenPartiallyFilled:
			// The original is still live. If it stays that way it was never
			// replaced — the amend did not land.
			stableOpen++
			if stableOpen >= cfg.StableObservations && clk.Now().Sub(started) >= cfg.MinObservation {
				res.State = journal.StateFailedConfirmed
				res.Reason = ReasonCode(journal.ReasonResolvedAbsent)
				res.Detail = fmt.Sprintf(
					"order %s is still open and unreplaced after %d observations over %s, so the amend never landed",
					target, stableOpen, clk.Now().Sub(started).Round(time.Second))
				if err := attempt.ResolveFailed(ctx, journal.ReasonResolvedAbsent, res.Detail); err != nil {
					return res, err
				}
				return res, nil
			}
			if !r.waitOrExpire(ctx, clk, cfg, deadline) {
				return r.park(ctx, attempt, res, fmt.Sprintf(
					"order %s was still open when the %s resolution budget ran out", target, cfg.MaxDuration))
			}

		case brokerstate.StateFilled:
			res.State = journal.StateFailedConfirmed
			res.Reason = ReasonCode(journal.ReasonResolvedAbsent)
			res.Detail = fmt.Sprintf("order %s filled in full, so there was nothing left to amend", target)
			if err := attempt.ResolveFailed(ctx, journal.ReasonResolvedAbsent, res.Detail); err != nil {
				return res, err
			}
			return res, nil

		case brokerstate.StateCancelled, brokerstate.StateCancelledPartiallyFilled, brokerstate.StateReplaced:
			// The original is closed. An amend that landed produced a successor;
			// find it, or admit we cannot account for the exposure.
			successors, scanErr := r.findSuccessors(ctx, successorMatcher, target, cfg)
			if scanErr != nil {
				if !r.waitOrExpire(ctx, clk, cfg, deadline) {
					return r.park(ctx, attempt, res,
						fmt.Sprintf("the successor scan failed: %v", scanErr))
				}
				continue
			}
			switch len(successors) {
			case 1:
				child := successors[0]
				res.State = journal.StateConfirmed
				res.BrokerOrderID = child.OrderID
				res.Reason = ReasonCode(journal.ReasonResolvedFound)
				res.Detail = fmt.Sprintf(
					"order %s closed with %s filled and order %s carries the remainder",
					target, decimalString(derived.FilledQuantity), child.OrderID)
				edge := journal.LineageEdge{
					ParentOrderID:        target,
					ChildOrderID:         child.OrderID,
					Relation:             journal.RelationReplaces,
					ParentFilledQuantity: decimalString(derived.FilledQuantity),
					RequestedQuantity:    decimalString(child.Quantity),
				}
				if err := attempt.ResolveConfirmedWithLineage(ctx, edge,
					journal.ReasonResolvedFound, res.Detail); err != nil {
					return res, err
				}
				return res, nil

			case 0:
				// Closed with no successor: either the amend was rejected and
				// something else cancelled the order, or the successor is not
				// visible to us. Those have opposite exposure implications.
				return r.park(ctx, attempt, res, fmt.Sprintf(
					"order %s is %s but no successor carrying the amended order was found",
					target, derived.State))

			default:
				ids := make([]string, 0, len(successors))
				for _, s := range successors {
					ids = append(ids, s.OrderID)
				}
				return r.park(ctx, attempt, res, fmt.Sprintf(
					"order %s is %s and more than one order could be its successor (%s)",
					target, derived.State, strings.Join(ids, ", ")))
			}

		default: // UNKNOWN_BROKER_STATE
			return r.park(ctx, attempt, res, fmt.Sprintf(
				"order %s derives as %s (%s): %s", target, derived.State, derived.Reason, derived.Detail))
		}
	}
}

// deriveOrder reads one order and runs it through the priority table, with the
// lineage the journal already knows about.
func (r *Resolver) deriveOrder(ctx context.Context, orderID string) (brokerstate.Derived, error) {
	raw, err := r.Order.OrderRaw(ctx, orderID)
	if err != nil {
		return brokerstate.Derived{}, err
	}
	lineage, err := r.knownLineage(ctx, orderID)
	if err != nil {
		return brokerstate.Derived{}, err
	}
	return brokerstate.DeriveOfficialOrder(raw, lineage), nil
}

// knownLineage asks the journal whether this order already has a recorded
// successor. The derivation needs it: CLOSED + a successor is REPLACED, not
// CANCELLED, and reporting a replace as a cancellation would tell the engine it is
// flat while a live order still carries the remainder.
func (r *Resolver) knownLineage(ctx context.Context, orderID string) (brokerstate.Lineage, error) {
	edges, err := r.Journal.LineageChildren(ctx, orderID)
	if err != nil {
		return brokerstate.Lineage{}, err
	}
	for _, edge := range edges {
		if edge.Relation == journal.RelationReplaces {
			return brokerstate.Lineage{SuccessorOrderID: edge.ChildOrderID}, nil
		}
	}
	return brokerstate.Lineage{}, nil
}

// findSuccessors scans the symbol's OPEN and CLOSED lists for orders that could be
// the amended order, excluding the original itself.
//
// Symbol-scoped rather than account-wide on purpose: the engine holds one
// in-flight mutation per symbol, so within a symbol the amended order is unique;
// widening the scan would only add candidates that cannot be it.
func (r *Resolver) findSuccessors(ctx context.Context, m *matcher, originalID string, cfg ResolveConfig) ([]orderFacts, error) {
	var out []orderFacts
	for _, status := range []string{"OPEN", "CLOSED"} {
		raws, err := ScanOrders(ctx, r.Orders, OrderQuery{Status: status, Symbol: m.symbol}, cfg.MaxPages)
		if err != nil {
			return nil, fmt.Errorf("scanning the %s list: %w", status, err)
		}
		for _, raw := range raws {
			facts, err := parseOrderFacts(raw)
			if err != nil {
				return nil, fmt.Errorf("an order in the %s list could not be read: %w", status, err)
			}
			if facts.OrderID == originalID {
				continue
			}
			if m.matches(facts) {
				out = append(out, facts)
			}
		}
	}
	return out, nil
}
