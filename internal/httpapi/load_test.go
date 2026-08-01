package httpapi

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStreamCapsConcurrentClientsAtThirtyTwo(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a"})
	subscriptions := make([]*streamSubscription, 0, DefaultStreamMaxClients)
	for n := 0; n < DefaultStreamMaxClients; n++ {
		subscription, err := stream.subscribe(context.Background(), "process-a:0")
		if err != nil {
			t.Fatalf("client %d: %v", n+1, err)
		}
		subscriptions = append(subscriptions, subscription)
	}
	defer func() {
		for _, subscription := range subscriptions {
			subscription.close()
		}
	}()
	if _, err := stream.subscribe(context.Background(), "process-a:0"); !errors.Is(err, ErrStreamClientLimit) {
		t.Fatalf("33rd client error=%v", err)
	}
}

func TestQueueFullDisconnectsOnlySlowClientAndNeverBlocksProducer(t *testing.T) {
	stream, _ := newTestStream(t, StreamOptions{Epoch: "process-a", QueueSize: 2, MaxClients: 2})
	slow, err := stream.subscribe(context.Background(), "process-a:0")
	if err != nil {
		t.Fatal(err)
	}
	defer slow.close()
	fast, err := stream.subscribe(context.Background(), "process-a:0")
	if err != nil {
		t.Fatal(err)
	}
	defer fast.close()

	producerDone := make(chan error, 1)
	go func() {
		for n := 0; n < 3; n++ {
			if _, err := stream.Publish("update", []byte(`{}`)); err != nil {
				producerDone <- err
				return
			}
			select {
			case <-fast.events:
			case <-time.After(time.Second):
				producerDone <- errors.New("fast client was blocked by slow client")
				return
			}
		}
		producerDone <- nil
	}()
	select {
	case err := <-producerDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("producer blocked on a full client queue")
	}
	select {
	case <-slow.done:
	default:
		t.Fatal("queue-full slow client was not disconnected")
	}
	select {
	case <-fast.done:
		t.Fatal("queue-full disconnected the healthy client")
	default:
	}
	if got := stream.ClientCount(); got != 1 {
		t.Fatalf("remaining clients=%d, want 1", got)
	}
}
