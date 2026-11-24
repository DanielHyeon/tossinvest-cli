package engine

// a197 — 단계의 순수 판정 둘을 직접 잼: 열 조건(§4-8 · §4-9)과 실행 직전 재검증(§4-15a, Manager 결정 M-3).

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestA197TheStageOpensOnlyOnAProtectiveOppositePendingRefusal(t *testing.T) {
	stop := exitpolicy.Proposal{Action: exitpolicy.ActionBaselineBreach}
	ladderStop := exitpolicy.Proposal{Action: exitpolicy.ActionLadderStop}
	takeProfit := exitpolicy.Proposal{Action: exitpolicy.ActionLadderTakeProfit}
	opposite := execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonOppositePendingOrder}
	for name, tc := range map[string]struct {
		p    exitpolicy.Proposal
		out  execgw.Outcome
		want bool
	}{
		"stop refused for an opposite order":      {stop, opposite, true},
		"ladder stop refused for an opposite order": {ladderStop, opposite, true},
		"take profit refused for an opposite order": {takeProfit, opposite, false},
		"stop refused for another reason":          {stop, execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonCode(journal.ReasonDispatchRejected)}, false},
		"stop not dispatched with that reason":     {stop, execgw.Outcome{State: journal.StateNotDispatched, Reason: execgw.ReasonOppositePendingOrder}, false},
	} {
		if got := externalBuyOpens(tc.p, tc.out); got != tc.want {
			t.Errorf("%s: opens = %v, want %v", name, got, tc.want)
		}
	}
}

func TestA197TheItemMustStillBeTheSameOpenInstance(t *testing.T) {
	item := externalBuyItem{positionID: "p-1", instanceSeq: 2, market: "kr", symbol: "005930"}
	open := journal.Position{ID: "p-1", InstanceSeq: 2, State: "OPEN", Quantity: "10"}
	for name, tc := range map[string]struct {
		positions []journal.Position
		want      bool
	}{
		"same open instance": {[]journal.Position{open}, true},
		"closed":             {[]journal.Position{{ID: "p-1", InstanceSeq: 2, State: journal.PositionClosed, Quantity: "0"}}, false},
		"zero quantity":      {[]journal.Position{{ID: "p-1", InstanceSeq: 2, State: "OPEN", Quantity: "0"}}, false},
		"new instance":       {[]journal.Position{{ID: "p-1", InstanceSeq: 3, State: "OPEN", Quantity: "10"}}, false},
		"gone":               {nil, false},
		"another position":   {[]journal.Position{{ID: "p-2", InstanceSeq: 2, State: "OPEN", Quantity: "10"}}, false},
	} {
		if got := externalBuyItemCurrent(item, tc.positions); got != tc.want {
			t.Errorf("%s: current = %v, want %v", name, got, tc.want)
		}
	}
}
