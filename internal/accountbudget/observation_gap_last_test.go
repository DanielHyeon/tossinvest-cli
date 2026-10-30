//go:build linux

package accountbudget

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestA166ObservationRejectsRetainedGapAsNoResponse(t *testing.T) {
	b, s, _ := a166Store(t)
	state := observationState{Version: 1, Scope: "scope", Instance: "instance", Attempt: 2, Gap: 1, Last: &ObservationCompletion{Attempt: 2, Kind: ObservationOther, Result: ObservationNoResponse}, Events: []ObservationEvent{}, Gaps: []ObservationGap{{Sequence: 1, Attempt: 2, Kind: ObservationOther}}, Slots: []observationSlot{}}
	raw, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(b.root, observationFile), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Snapshot(context.Background(), ObservationCursor{}); err == nil {
		t.Fatal("retained gap accepted as completed no-response")
	}
}
