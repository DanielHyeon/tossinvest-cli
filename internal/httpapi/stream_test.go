package httpapi

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func newTestStream(t *testing.T, options StreamOptions) (*Stream, *snapshotFixture) {
	t.Helper()
	fixture := &snapshotFixture{data: []byte(`{"schema_version":"v1","full":true}`)}
	stream, err := NewStream(options, fixture.read)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(stream.Close)
	return stream, fixture
}

type snapshotFixture struct {
	calls int
	data  []byte
	err   error
}

func (f *snapshotFixture) read(context.Context) ([]byte, error) {
	f.calls++
	return append([]byte(nil), f.data...), f.err
}

func TestStreamEventIDsContainEpochAndMonotonicSequence(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a"})

	first, err := stream.Publish("engine", []byte(`{"running":false}`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := stream.Publish("positions", []byte(`[]`))
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != "process-a:1" || first.Sequence != 1 {
		t.Fatalf("first event=%+v", first)
	}
	if second.ID != "process-a:2" || second.Sequence != 2 {
		t.Fatalf("second event=%+v", second)
	}
}

func TestStreamSequenceNeverWrapsOrMutatesPublishedData(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a", QueueSize: 1})
	subscription, err := stream.subscribe(context.Background(), "process-a:0")
	if err != nil {
		t.Fatal(err)
	}
	defer subscription.close()

	input := []byte(`{"value":"original"}`)
	event, err := stream.Publish("update", input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = '!'
	event.Data[1] = '!'
	delivered := <-subscription.events
	if string(delivered.Data) != `{"value":"original"}` {
		t.Fatalf("published data aliased caller memory: %q", delivered.Data)
	}

	stream.mu.Lock()
	stream.sequence = math.MaxUint64
	stream.mu.Unlock()
	if _, err := stream.Publish("update", nil); !errors.Is(err, ErrStreamExhausted) {
		t.Fatalf("sequence exhaustion error=%v", err)
	}
}

func TestStreamRejectsLimitsWiderThanServerContract(t *testing.T) {
	tests := []struct {
		name    string
		options StreamOptions
	}{
		{"clients", StreamOptions{MaxClients: DefaultStreamMaxClients + 1}},
		{"negative clients", StreamOptions{MaxClients: -1}},
		{"queue", StreamOptions{QueueSize: DefaultStreamQueueSize + 1}},
		{"negative queue", StreamOptions{QueueSize: -1}},
		{"heartbeat", StreamOptions{Heartbeat: DefaultStreamHeartbeat + time.Second}},
		{"negative heartbeat", StreamOptions{Heartbeat: -time.Second}},
		{"invalid epoch", StreamOptions{Epoch: "bad:epoch"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewStream(test.options, func(context.Context) ([]byte, error) { return []byte(`{}`), nil }); err == nil {
				t.Fatal("NewStream accepted a wider or invalid server-owned limit")
			}
		})
	}
	if _, err := NewStream(StreamOptions{}, nil); err == nil {
		t.Fatal("NewStream accepted a nil snapshot provider")
	}
}

func TestSnapshotFailureDoesNotLeakAClientSlot(t *testing.T) {
	stream, fixture := newTestStream(t, StreamOptions{Epoch: "process-a", MaxClients: 1})
	fixture.err = errors.New("snapshot unavailable")

	if _, err := stream.subscribe(context.Background(), "different:9"); err == nil {
		t.Fatal("subscribe accepted a failed full snapshot")
	}
	if got := stream.ClientCount(); got != 0 {
		t.Fatalf("client count=%d after snapshot failure", got)
	}
}

func TestHeartbeatUsesInjectedClockAndFiresEveryFifteenSeconds(t *testing.T) {
	fake := clock.NewFake(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a", Clock: fake})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	beats := stream.heartbeats(ctx)

	if !fake.WaitForSleepers(1, time.Second) {
		t.Fatal("heartbeat did not wait on the injected clock")
	}
	fake.Advance(DefaultStreamHeartbeat - time.Nanosecond)
	select {
	case <-beats:
		t.Fatal("heartbeat fired before 15 seconds")
	default:
	}
	fake.Advance(time.Nanosecond)
	select {
	case <-beats:
	case <-time.After(time.Second):
		t.Fatal("heartbeat did not fire at 15 seconds")
	}
	if !fake.WaitForSleepers(1, time.Second) {
		t.Fatal("heartbeat did not schedule the next interval")
	}
}

func TestCloseDisconnectsClientsAndRejectsFurtherPublish(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a"})
	subscription, err := stream.subscribe(context.Background(), "process-a:0")
	if err != nil {
		t.Fatal(err)
	}

	stream.Close()
	stream.Close()
	select {
	case <-subscription.done:
	default:
		t.Fatal("Close did not disconnect the active client")
	}
	if stream.ClientCount() != 0 {
		t.Fatalf("client count after Close=%d", stream.ClientCount())
	}
	if _, err := stream.Publish("update", nil); !errors.Is(err, ErrStreamClosed) {
		t.Fatalf("publish after Close error=%v", err)
	}
}

func TestPublishRejectsUnsafeEventTypeWithoutAdvancingSequence(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a"})
	if _, err := stream.Publish("bad\nevent", nil); err == nil {
		t.Fatal("Publish accepted an unsafe event type")
	}
	event, err := stream.Publish("update", nil)
	if err != nil {
		t.Fatal(err)
	}
	if event.Sequence != 1 {
		t.Fatalf("rejected event advanced sequence to %d", event.Sequence)
	}
}
