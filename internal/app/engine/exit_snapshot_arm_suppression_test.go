package engine_test

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func TestUnclearedWorkingOrderPersistsExactSnapshotAndTypedArmSuppression(t *testing.T) {
	h := newExitHarness(t, nil)
	position := h.entry("005930", "10", "70000", "68000", "70000")
	h.workingEntry("005930", "5", "69500")
	h.submit.cancelFails = true
	h.quote("005930", 67900)

	h.observe()
	state := h.state(position.ID)
	if state.Snapshot.Snapshot == nil || !state.Snapshot.Snapshot.Line.Orderable || state.Pending() {
		t.Fatalf("orderable evaluation must persist unchanged without arming: %+v", state)
	}
	events, err := h.journal.ExitEvents(context.Background(), position.ID)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.ArmSuppressedReason != journal.ArmSuppressedWorkingOrder ||
		last.Evaluation.Recomputed.Snapshot == nil || !last.Evaluation.Recomputed.Snapshot.Line.Orderable {
		t.Fatalf("arm-suppression audit evidence = %+v", last)
	}
}
