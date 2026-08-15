package verifylive

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestM0PendingIntentIsDurableBeforeConditionalCreate(t *testing.T) {
	h := triggerHarness(t, nil)
	seedM0TriggerPrerequisites(t, h)
	runner := h.runner(t, triggerOptions(t, time.Minute))
	seen := false
	runner.m0BeforeConditionalCreate = func() error {
		entries := h.entries()
		for _, entry := range entries {
			if entry.M0Checkpoint != nil && entry.M0Checkpoint.Kind == "pending-create" {
				seen = true
			}
		}
		if n := h.broker.countRequests("POST /conditional-orders"); n != 0 {
			t.Fatalf("conditional create happened before durable pending intent: %d", n)
		}
		return errors.New("before-create killpoint")
	}
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("killpoint becomes a measured step failure, not program error: %v", err)
	}
	if !seen {
		t.Fatal("pending-create checkpoint was not durable before conditional create")
	}
}

func TestM0AfterCreateBeforeParentCheckpointRecoversUniqueParentWithoutPOST(t *testing.T) {
	h := triggerHarness(t, nil)
	seedM0TriggerPrerequisites(t, h)
	runner := h.runner(t, triggerOptions(t, time.Minute))
	runner.m0AfterConditionalCreate = func() error { return errors.New("after-create killpoint") }
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatalf("killpoint becomes a measured failure, not program error: %v", err)
	}
	if n := h.broker.countRequests("POST /conditional-orders"); n != 1 {
		t.Fatalf("create count = %d, want one accepted parent before crash", n)
	}
	entries := h.entries()
	parentRecorded := false
	for _, entry := range entries {
		if entry.M0Checkpoint != nil && entry.M0Checkpoint.Kind == "parent-created" {
			parentRecorded = true
		}
	}
	if parentRecorded {
		t.Fatal("after-create killpoint recorded parent id")
	}
	_, err := h.run(triggerOptions(t, time.Minute))
	if err == nil {
		t.Fatal("fresh unique recovery continued instead of terminal HOLD")
	}
	if n := h.broker.countRequests("POST /conditional-orders"); n != 1 {
		t.Fatalf("recovery made create POST: %d", n)
	}
	parentRecorded = false
	for _, entry := range h.entries() {
		if entry.M0Checkpoint != nil && entry.M0Checkpoint.Kind == "parent-created" {
			parentRecorded = true
		}
	}
	if !parentRecorded {
		t.Fatal("unique recovery did not durably record the recovered parent before HOLD")
	}
}

func TestM0ParentCheckpointAppendFailureIsTerminalAndResumesReadOnly(t *testing.T) {
	h := triggerHarness(t, nil)
	seedM0TriggerPrerequisites(t, h)
	runner := h.runner(t, triggerOptions(t, time.Minute))
	runner.m0AppendCheckpoint = func(entry Entry) error {
		if entry.M0Checkpoint != nil && entry.M0Checkpoint.Kind == "parent-created" {
			return errors.New("parent checkpoint writer kill")
		}
		return runner.recorder.Append(entry)
	}
	summary, err := runner.Run(context.Background())
	if err == nil || !errors.Is(err, ErrM0TerminalHold) || !summary.Halted || !strings.Contains(summary.Halt, "parent-created checkpoint") {
		t.Fatalf("append failure = summary=%+v err=%v, want terminal M0 HOLD", summary, err)
	}
	if n := h.broker.countRequests("POST /conditional-orders"); n != 1 {
		t.Fatalf("initial run conditional POSTs = %d, want one", n)
	}
	entries := h.entries()
	pending, parent := false, false
	for _, entry := range entries {
		if entry.M0Checkpoint == nil {
			continue
		}
		pending = pending || entry.M0Checkpoint.Kind == "pending-create"
		parent = parent || entry.M0Checkpoint.Kind == "parent-created"
	}
	if !pending || parent {
		t.Fatalf("checkpoints pending=%v parent=%v, want durable pending only", pending, parent)
	}
	if _, err := h.run(triggerOptions(t, time.Minute)); err == nil || !strings.Contains(err.Error(), "recovered one parent") {
		t.Fatalf("resume did not stop after read-only recovery: %v", err)
	}
	if n := h.broker.countRequests("POST /conditional-orders"); n != 1 {
		t.Fatalf("resume conditional POSTs = %d, want no second create", n)
	}
}
