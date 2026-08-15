package verifylive

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestM0ChildCheckpointThenCausalBarrierPrecedeChildGET(t *testing.T) {
	sequence := []string{}
	h := triggerHarness(t, func(f *fakeBroker) {
		f.firesOnRead(1, 1, 1)
		f.beforeOrderRawByID = func(id string) {
			if len(id) >= len("child-") && id[:len("child-")] == "child-" {
				sequence = append(sequence, "child-get")
			}
		}
	})
	seedM0TriggerPrerequisites(t, h)
	runner := h.runner(t, triggerOptions(t, time.Minute))
	runner.m0AfterChildCheckpoint = func() error { sequence = append(sequence, "checkpoint"); return nil }
	runner.m0AfterParentCausal = func() error { sequence = append(sequence, "causal"); return nil }
	if _, err := runner.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(sequence) != 3 || sequence[0] != "checkpoint" || sequence[1] != "causal" || sequence[2] != "child-get" {
		t.Fatalf("durable ordering = %v, want checkpoint then causal then child GET", sequence)
	}
	if n := h.broker.countRequests("GET /orders/child-"); n == 0 {
		t.Fatal("child GET was never attempted after both durable barriers")
	}
}

func TestM0BarrierFailurePreventsChildGET(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  func(*Runner)
	}{
		{name: "checkpoint", set: func(r *Runner) { r.m0AfterChildCheckpoint = func() error { return errors.New("checkpoint sync kill") } }},
		{name: "causal", set: func(r *Runner) { r.m0AfterParentCausal = func() error { return errors.New("causal sync kill") } }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) { f.firesOnRead(1, 1, 1) })
			seedM0TriggerPrerequisites(t, h)
			runner := h.runner(t, triggerOptions(t, time.Minute))
			tc.set(runner)
			_, _ = runner.Run(context.Background()) // outstanding parent is expected after barrier kill.
			if n := h.broker.countRequests("GET /orders/child-"); n != 0 {
				t.Fatalf("barrier failure permitted %d child GET(s)", n)
			}
		})
	}
}

func TestM0ActualCheckpointWriterAndReceiptSyncFailuresPreventChildGET(t *testing.T) {
	for _, tc := range []struct {
		name string
		set  func(*Runner)
	}{
		{name: "checkpoint-writer", set: func(r *Runner) {
			r.m0AppendCheckpoint = func(entry Entry) error {
				if entry.M0Checkpoint != nil && entry.M0Checkpoint.Kind == "child-observed" {
					return errors.New("child checkpoint writer kill")
				}
				return r.recorder.Append(entry)
			}
		}},
		{name: "parent-causal-fsync", set: func(r *Runner) {
			r.m0AfterChildCheckpoint = func() error {
				r.m0Receipt.sync = func(*os.File) error { return errors.New("parent causal fsync kill") }
				return nil
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := triggerHarness(t, func(f *fakeBroker) { f.firesOnRead(1, 1, 1) })
			seedM0TriggerPrerequisites(t, h)
			runner := h.runner(t, triggerOptions(t, time.Minute))
			tc.set(runner)
			_, _ = runner.Run(context.Background())
			if n := h.broker.countRequests("GET /orders/child-"); n != 0 {
				t.Fatalf("%s permitted %d child GET(s)", tc.name, n)
			}
		})
	}
}
