package httpapi

import (
	"context"
	"testing"
)

func TestLastEventIDResumesOnlyContiguousSameEpochEvents(t *testing.T) {
	stream, fixture := newTestStream(t, StreamOptions{Epoch: "process-a", QueueSize: 4})
	for _, data := range []string{`{"n":1}`, `{"n":2}`, `{"n":3}`} {
		if _, err := stream.Publish("update", []byte(data)); err != nil {
			t.Fatal(err)
		}
	}

	subscription, err := stream.subscribe(context.Background(), "process-a:1")
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.close()
	if fixture.calls != 0 {
		t.Fatalf("continuous resume generated %d full snapshots", fixture.calls)
	}
	if len(subscription.initial) != 2 || subscription.initial[0].ID != "process-a:2" || subscription.initial[1].ID != "process-a:3" {
		t.Fatalf("resume events=%+v", subscription.initial)
	}
}

func TestEpochMismatchUnknownAndGapConvergeWithFullSnapshot(t *testing.T) {
	tests := []struct {
		name        string
		lastEventID string
	}{
		{"epoch mismatch", "old-process:3"},
		{"unknown future", "process-a:99"},
		{"malformed", "not-an-event-id"},
		{"gap", "process-a:1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream, fixture := newTestStream(t, StreamOptions{Epoch: "process-a", QueueSize: 2})
			for n := 0; n < 4; n++ {
				if _, err := stream.Publish("update", []byte(`{}`)); err != nil {
					t.Fatal(err)
				}
			}
			subscription, err := stream.subscribe(context.Background(), test.lastEventID)
			if err != nil {
				t.Fatal(err)
			}
			defer subscription.close()
			if fixture.calls != 1 || len(subscription.initial) != 1 {
				t.Fatalf("snapshot calls/events=%d/%d", fixture.calls, len(subscription.initial))
			}
			event := subscription.initial[0]
			if event.Type != StreamEventSnapshot || event.ID != "process-a:4" || string(event.Data) != string(fixture.data) {
				t.Fatalf("full snapshot=%+v", event)
			}
		})
	}
}

func TestProcessRestartForcesNewEpochFullSnapshot(t *testing.T) {
	old, _ := newTestStream(t, StreamOptions{Epoch: "process-a"})
	event, err := old.Publish("update", []byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	restarted, fixture := newTestStream(t, StreamOptions{Epoch: "process-b"})
	subscription, err := restarted.subscribe(context.Background(), event.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.close()
	if fixture.calls != 1 || len(subscription.initial) != 1 || subscription.initial[0].ID != "process-b:0" {
		t.Fatalf("restart convergence=%+v calls=%d", subscription.initial, fixture.calls)
	}
}
