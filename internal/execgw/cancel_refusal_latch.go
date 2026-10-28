package execgw

import (
	"context"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// cancelRefusalLatch 는 살아 있는 대상 주문의 R3 종결 거절만 진입에서 막음.
func (g *Gateway) cancelRefusalLatch(ctx context.Context, plan mutationPlan) (*RejectedError, error) {
	latches, err := g.journal.FailedConfirmedCancelsForSymbol(ctx, g.accountRef, plan.market, plan.symbol)
	if err != nil {
		return nil, fmt.Errorf("execgw: reading cancel refusal latches on %s: %w", plan.symbol, err)
	}
	if len(latches) == 0 {
		return nil, nil
	}
	live, err := g.journal.LiveOrdersForSymbol(ctx, g.accountRef, plan.market, plan.symbol)
	if err != nil {
		return nil, fmt.Errorf("execgw: reading live orders for cancel refusal latches on %s: %w", plan.symbol, err)
	}
	for _, latch := range latches {
		if !isCancelRefusalLatch(latch) || !liveOrderID(live, latch.TargetOrderID) {
			continue
		}
		if _, released, err := g.journal.ConfirmedCancelAfterLatch(ctx, latch.ID); err != nil {
			return nil, fmt.Errorf("execgw: reading cancellation release for latch %s: %w", latch.ID, err)
		} else if released {
			continue
		}
		return reject(ReasonSymbolInFlight,
			"cancel of order %s was definitively refused while still working (attempt %s)", latch.TargetOrderID, latch.ID), nil
	}
	return nil, nil
}

// isCancelRefusalLatch 는 쓰기와 같은 꼬리표 한 덩어리만 R3 latch로 읽음.
func isCancelRefusalLatch(record journal.AttemptRecord) bool {
	detail := strings.TrimSpace(record.Detail)
	start := strings.LastIndex(detail, cancelAlreadyProcessingMarker)
	if start < 0 || !strings.HasSuffix(detail, "]") {
		return false
	}
	fields := strings.Fields(strings.TrimSuffix(detail[start+len(cancelAlreadyProcessingMarker):], "]"))
	seen := map[string]string{}
	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return false
		}
		seen[key] = value
	}
	return seen["branch"] == "cancel_already_processing" && seen["outcome"] == "definitive"
}

// liveOrderID 는 주문 번호의 원문 바이트 동일성만 허용함.
func liveOrderID(orders []journal.LiveOrder, target string) bool {
	for _, order := range orders {
		if order.OrderID == target {
			return true
		}
	}
	return false
}
