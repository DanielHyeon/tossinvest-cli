package main

// a102_ready_wiring_test.go is the seam between boot step 6 and boot step 7
// (openspec change a102, task 3.9 ②, design D5c).
//
// It is its own file rather than another block in engine_test.go for the reason
// this repository has learned twice: appending to an existing Go file makes the
// neighbouring function look "modified" to tools/logic-map/check_analysis.py, and
// the evidence bundle it then demands is about a function nobody edited.

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/testenv"
)

// TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam is the wiring A2 found
// unguarded (a102 §3.9, D5c).
//
// The seam crosses two boot steps: step 6 holds the marker and step 7 assembles
// the loops, and the runtime calls Recover before the first loop starts. Every
// unit test of recoverThenReady proves what happens *given* a ready function;
// none of them proves that runEngineRun hands one over, or that calling it writes
// anything. A2 emptied the closure and passed nil in its place and the whole
// suite stayed green.
//
// The assertion is the marker file, read from inside the factory — after the
// command returns the marker is gone, which is the other half of the contract.
func TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam(t *testing.T) {
	dir := testenv.Isolate(t)
	stubAssembly(t, verifiedEngine(), nil)

	markerPath := enginelock.MarkerPath(dir)
	var readyWasNil bool
	var beforeRecovery, afterRecovery enginelock.Status
	stubRuntimeWithReady(t, func(_ *engine.Context, ready func()) (*engine.Runtime, error) {
		readyWasNil = ready == nil
		beforeRecovery = enginelock.Read(markerPath, time.Now())
		if ready != nil {
			ready()
		}
		afterRecovery = enginelock.Read(markerPath, time.Now())
		return nil, errStubRuntimeReached
	})

	if _, _, err := runCLI(t, "--config-dir", dir, "engine", "run"); !errors.Is(err, errStubRuntimeReached) {
		t.Fatalf("err = %v", err)
	}
	if readyWasNil {
		t.Fatal("the runtime was assembled with no ready seam; nothing can publish the signal")
	}
	if beforeRecovery.Ready() {
		t.Error("the marker already said ready before recovery ran — the signal is not the recovery's")
	}
	if !afterRecovery.Ready() {
		t.Fatal("calling the runtime's ready seam did not publish ready_at into the marker")
	}
	if afterRecovery.Marker.PID != os.Getpid() {
		t.Errorf("the ready marker names pid %d, want this process", afterRecovery.Marker.PID)
	}
	if status := enginelock.Read(markerPath, time.Now()); status.Ready() {
		t.Error("the ready marker survived the command returning")
	}
}
