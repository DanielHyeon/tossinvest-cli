//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestA166EnginePollKeyScopedAndObserved(t *testing.T) {
	b, store := observedFixture(t)
	a, err := b.EnginePollKey("account-a", "price")
	if err != nil {
		t.Fatal(err)
	}
	again, err := b.EnginePollKey("account-a", "price")
	if err != nil || a != again {
		t.Fatalf("unstable key %v", err)
	}
	other, err := b.EnginePollKey("account-a", "holdings")
	if err != nil || a == other {
		t.Fatal("kind collision")
	}
	other, err = b.EnginePollKey("account-b", "price")
	if err != nil || a == other {
		t.Fatal("account collision")
	}
	b2, err := OpenObserved(context.Background(), Config{Root: b.root, Scope: b.scope, Instance: b.instance, Clock: b.clk}, store.lockID, store.boot)
	if err != nil {
		t.Fatal(err)
	}
	defer b2.Close()
	restart, err := b2.EnginePollKey("account-a", "price")
	if err != nil || restart != a {
		t.Fatalf("second factory key %v", err)
	}
	pair1, err := b.EnginePollKey("ab", "c")
	if err != nil {
		t.Fatal(err)
	}
	pair2, err := b.EnginePollKey("a", "bc")
	if err != nil || pair1 == pair2 {
		t.Fatal("ambiguous field encoding")
	}
	remaining, err := b.PollRemaining(context.Background(), a)
	if err != nil || remaining != 0 {
		t.Fatalf("remaining=%v err=%v", remaining, err)
	}
	plain, _ := fixture(t)
	if _, err = plain.EnginePollKey("account-a", "price"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unobserved %v", err)
	}
}

func TestA166EnginePollRemainingExpires(t *testing.T) {
	b, store := observedFixture(t)
	key, err := b.EnginePollKey("account-a", "price")
	if err != nil {
		t.Fatal(err)
	}
	lease, err := b.Acquire(context.Background(), Normal)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := store.Begin(context.Background(), lease, ObservationPoll, key)
	if err != nil {
		t.Fatal(err)
	}
	hint := 2 * time.Second
	if err = store.Finish(context.Background(), lease, seq, ObservationRateLimited, &hint); err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	left, err := b.PollRemaining(context.Background(), key)
	if err != nil || left != hint {
		t.Fatalf("remaining=%v err=%v", left, err)
	}
	store.boot.(*a166Boot).stamp.Uptime += 3 * time.Second
	left, err = b.PollRemaining(context.Background(), key)
	if err != nil || left != 0 {
		t.Fatalf("expired=%v err=%v", left, err)
	}
}
