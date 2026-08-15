package verifylive

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// m0RawConditionalPageReader is read-only. It is intentionally narrower than
// Broker so M0 recovery cannot gain any mutation capability.
type m0RawConditionalPageReader interface {
	ProtectionConditionalOrdersRaw(context.Context, string, string, string, int) (official.RawConditionalOrderList, error)
}

// M0Unsettled returns the one durable M0 owner which still needs a human.
// A parent create acknowledgement is not a terminal result: until the trigger
// step durably records its PASS, both the parent and an exposed child are held
// for manual reconciliation.  In particular, a restart must never mistake a
// parent-created line for permission to issue another conditional POST.
func M0Unsettled(entries []Entry) (M0Checkpoint, bool, error) {
	pending := map[string]M0Checkpoint{}
	parents := map[string]M0Checkpoint{}
	children := map[string]M0Checkpoint{}
	settledParents := map[string]bool{}
	for _, entry := range entries {
		if entry.Kind == KindStep && entry.StepID == StepConditionalTrigger && entry.Verdict == VerdictPass {
			for _, artifact := range entry.Artifacts {
				if artifact.Kind == KindConditional && artifact.Filled && artifact.ID != "" {
					settledParents[artifact.ID] = true
					delete(parents, artifact.ID)
					for childID, owner := range children {
						if owner.ParentConditionalID == artifact.ID {
							delete(children, childID)
						}
					}
				}
			}
			continue
		}
		if entry.Kind != KindM0Checkpoint || entry.M0Checkpoint == nil {
			continue
		}
		checkpoint := *entry.M0Checkpoint
		switch checkpoint.Kind {
		case "pending-create":
			pending[checkpoint.ClientOrderID] = checkpoint
		case "parent-created":
			if candidate, ok := pending[checkpoint.ClientOrderID]; ok && m0CheckpointScopeEqual(candidate, checkpoint) {
				delete(pending, checkpoint.ClientOrderID)
			}
			if checkpoint.ParentConditionalID != "" && !settledParents[checkpoint.ParentConditionalID] {
				parents[checkpoint.ParentConditionalID] = checkpoint
			}
		case "child-observed":
			if checkpoint.ParentConditionalID != "" && !settledParents[checkpoint.ParentConditionalID] {
				parents[checkpoint.ParentConditionalID] = checkpoint
			}
			if checkpoint.ChildOrderID != "" {
				children[checkpoint.ChildOrderID] = checkpoint
			}
		case "manual-resolved":
			delete(parents, checkpoint.ParentConditionalID)
			delete(children, checkpoint.ChildOrderID)
		}
	}
	owners := make([]M0Checkpoint, 0, len(pending)+len(parents)+len(children))
	for _, checkpoint := range pending {
		owners = append(owners, checkpoint)
	}
	for _, checkpoint := range parents {
		owners = append(owners, checkpoint)
	}
	for child, checkpoint := range children {
		if checkpoint.ParentConditionalID == "" || parents[checkpoint.ParentConditionalID].ChildOrderID != child {
			owners = append(owners, checkpoint)
		}
	}
	if len(owners) == 0 {
		return M0Checkpoint{}, false, nil
	}
	sort.Slice(owners, func(i, j int) bool {
		return owners[i].ClientOrderID+owners[i].ParentConditionalID+owners[i].ChildOrderID < owners[j].ClientOrderID+owners[j].ParentConditionalID+owners[j].ChildOrderID
	})
	if len(owners) != 1 {
		return M0Checkpoint{}, false, fmt.Errorf("verify: M0 causal receipt HOLD: multiple unresolved pending M0 owners require manual reconciliation")
	}
	return owners[0], true, nil
}

// M0ExactPrerequisites keeps trigger-only mode from becoming a fresh full
// verification run. Every other procedure step must already be a positive
// record fact; in particular no earlier mutating step may be reached after the
// operator approved the irreversible trigger redo.
func M0ExactPrerequisites(entries []Entry) error {
	for _, step := range Steps() {
		if step.ID == StepConditionalTrigger {
			continue
		}
		if !Passed(entries, step.ID) {
			return fmt.Errorf("verify: M0 trigger mode requires prior PASS for %s before conditional-trigger redo", step.ID)
		}
	}
	return nil
}

func (r *Runner) m0PendingCheckpoint() (M0Checkpoint, bool, error) {
	checkpoint, ok, err := M0Unsettled(r.prior)
	if err != nil || !ok {
		return checkpoint, ok, err
	}
	if checkpoint.Kind != "pending-create" {
		return M0Checkpoint{}, false, fmt.Errorf("verify: M0 causal receipt HOLD: unsettled %s owner requires manual reconciliation before cleanup or redo", checkpoint.Kind)
	}
	return checkpoint, true, nil
}

func m0CheckpointScopeEqual(a, b M0Checkpoint) bool {
	return a.ClientOrderID == b.ClientOrderID && a.Symbol == b.Symbol && a.Market == b.Market && a.Type == b.Type &&
		a.Quantity == b.Quantity && a.Side == b.Side && a.OrderType == b.OrderType && a.ConditionType == b.ConditionType && a.Trigger == b.Trigger && a.ExpireDate == b.ExpireDate
}

// m0RecoverPending is deliberately terminal even when exactly one parent is
// found: a create response may have been lost, so it checkpoints ownership and
// makes the operator resume rather than continuing into another mutation.
func (r *Runner) m0RecoverPending(ctx context.Context, pending M0Checkpoint) error {
	reader, ok := r.broker.(m0RawConditionalPageReader)
	if !ok {
		return fmt.Errorf("verify: M0 causal receipt HOLD: broker lacks read-only all-page conditional recovery")
	}
	matches := map[string]official.RawConditionalOrder{}
	for _, status := range []string{"OPEN", "CLOSED"} {
		cursor := ""
		seen := map[string]bool{}
		for page := 0; page < maxFixturePages; page++ {
			if cursor != "" && seen[cursor] {
				return fmt.Errorf("verify: M0 causal receipt HOLD: conditional recovery repeated cursor")
			}
			if cursor != "" {
				seen[cursor] = true
			}
			readCtx := r.m0ReadContext(ctx, "pending-recovery-"+strings.ToLower(status))
			result, err := reader.ProtectionConditionalOrdersRaw(readCtx, status, pending.Symbol, cursor, 100)
			if err != nil {
				return fmt.Errorf("verify: M0 causal receipt HOLD: conditional recovery %s: %w", status, err)
			}
			if r.m0ReceiptErr != nil {
				return fmt.Errorf("verify: M0 causal receipt HOLD: %w", r.m0ReceiptErr)
			}
			for _, order := range result.Orders {
				if m0PendingMatches(pending, order) {
					if strings.TrimSpace(order.ID) == "" {
						return fmt.Errorf("verify: M0 causal receipt HOLD: conditional recovery returned empty parent id")
					}
					if prior, exists := matches[order.ID]; exists && prior != order {
						return fmt.Errorf("verify: M0 causal receipt HOLD: conditional recovery repeated parent id with conflicting raw identity")
					}
					matches[order.ID] = order
				}
			}
			if !result.HasNext {
				break
			}
			if strings.TrimSpace(result.NextCursor) == "" {
				return fmt.Errorf("verify: M0 causal receipt HOLD: conditional recovery empty cursor with more pages")
			}
			cursor = result.NextCursor
			if page == maxFixturePages-1 {
				return fmt.Errorf("verify: M0 causal receipt HOLD: conditional recovery page cap reached")
			}
		}
	}
	if len(matches) != 1 {
		return fmt.Errorf("verify: M0 causal receipt HOLD: pending create matched %d conditional orders", len(matches))
	}
	var match official.RawConditionalOrder
	for _, match = range matches {
	}
	checkpoint := pending
	checkpoint.Kind = "parent-created"
	checkpoint.ParentConditionalID = match.ID
	if err := r.appendM0Checkpoint(checkpoint); err != nil {
		return fmt.Errorf("verify: M0 causal receipt HOLD: persisting recovered parent: %w", err)
	}
	if r.m0ReceiptUsable() {
		if r.m0ReceiptLease == nil {
			return fmt.Errorf("verify: M0 causal receipt HOLD: receipt run lease is unavailable")
		}
		_, err := r.m0ReceiptLease.RecordCausal("pending-create-recovered", m0CausalFieldsV1{
			ParentResponseTag: r.m0Receipt.tag(match.ID), PendingClientTag: r.m0Receipt.tag(pending.ClientOrderID), ParentClientTag: r.m0Receipt.tag(match.ClientOrderID), Symbol: match.Symbol,
			RequestedMarket: pending.Market, Market: match.Market, Type: match.Type, OrderType: match.OrderType, Quantity: match.Quantity,
			Side: match.OrderSide, RootStatus: match.Status, FirstStatus: match.FirstStatus, Condition: match.ConditionType, Trigger: match.TriggerPrice, Expiry: match.ExpireDate,
		})
		if err != nil {
			return fmt.Errorf("verify: M0 causal receipt HOLD: persisting recovered causal evidence: %w", err)
		}
	}
	return fmt.Errorf("verify: M0 causal receipt HOLD: recovered one parent from pending create; restart required before any further action")
}

func m0PendingMatches(p M0Checkpoint, order official.RawConditionalOrder) bool {
	return order.ClientOrderID == p.ClientOrderID && order.Symbol == p.Symbol && order.Market == p.Market &&
		order.Type == "SINGLE" && order.Type == p.Type && order.OrderType == "MARKET" && order.OrderType == p.OrderType &&
		order.Quantity == "1" && order.Quantity == p.Quantity && strings.EqualFold(order.OrderSide, "SELL") &&
		strings.EqualFold(order.OrderSide, p.Side) && strings.EqualFold(order.ConditionType, "STOP") && strings.EqualFold(order.ConditionType, p.ConditionType) &&
		order.TriggerPrice == p.Trigger && order.ExpireDate == p.ExpireDate
}
