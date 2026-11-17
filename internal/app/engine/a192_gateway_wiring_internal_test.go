package engine

// a192_gateway_wiring_internal_test.go — 운영 조립(buildGateway)이 추적기에 종결 증명 장부를 실제로 넘기는지 고정함
// (a192 design D11). 이 배선이 빠지면 운영 엔진은 조용히 오늘의 행동(운영자 해제만)으로 돌아감.

import (
	"context"
	"testing"
)

func TestA192TheProductionTrackerReadsTerminalProofFromItsJournal(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	if err := bindApplyHooks(j); err != nil {
		t.Fatal(err)
	}
	wiring := a098BuildGateway(t, ctx, j)
	if wiring.reconcile == nil {
		t.Fatal("the assembly built no reconcile tracker")
	}
	if wiring.reconcile.Terminal != j {
		t.Fatalf("tracker.Terminal = %v, want the engine's own journal", wiring.reconcile.Terminal)
	}
}
