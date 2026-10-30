//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/sys/unix"
)

func a166FailLastDirectorySync(s *ObservationStore) error {
	sentinel := errors.New("final directory sync failure")
	count := 0
	s.syncDir = func(fd int) error {
		count++
		if count == 3 {
			return sentinel
		}
		return unix.Fsync(fd)
	}
	return sentinel
}

func TestA166ObservationPostcommitBeginFailureBecomesGap(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	var key ObservationKey
	key[0] = 42
	fault := a166FailLastDirectorySync(s)
	seq, err := s.Begin(ctx, lease, ObservationPoll, key)
	writes := 0
	if err == nil {
		writes++
	}
	if !errors.Is(err, fault) || seq != 0 || writes != 0 {
		t.Fatalf("begin seq=%d writes=%d err=%v", seq, writes, err)
	}
	s.syncDir = unix.Fsync
	view, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || view.Pending == nil || view.PendingStatus != ObservationPendingBusy {
		t.Fatalf("committed pending: %+v %v", view, err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	next, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer next.Release()
	seq, err = s.Begin(ctx, next, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	view, err = s.Snapshot(ctx, ObservationCursor{})
	if err != nil || len(view.Gaps) != 1 || view.Gaps[0].Attempt != 1 || view.Pending == nil || view.Pending.Sequence != seq {
		t.Fatalf("gap=%+v err=%v", view, err)
	}
	if _, err = s.PollRemaining(ctx, key); !errors.Is(err, ErrObservationPoisoned) {
		t.Fatalf("poison: %v", err)
	}
}

func TestA166ObservationPostcommitFinishFailureNoDuplicate(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	seq, err := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	fault := a166FailLastDirectorySync(s)
	if err = s.Finish(ctx, lease, seq, ObservationRateLimited, nil); !errors.Is(err, fault) {
		t.Fatalf("finish error=%v", err)
	}
	s.syncDir = unix.Fsync
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	view, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || view.Pending != nil || len(view.Events) != 1 || view.Events[0].Attempt != seq {
		t.Fatalf("committed event=%+v err=%v", view, err)
	}
	next, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer next.Release()
	n, err := s.Begin(ctx, next, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Finish(ctx, next, n, ObservationNoResponse, nil); err != nil {
		t.Fatal(err)
	}
	view, err = s.Snapshot(ctx, ObservationCursor{})
	if err != nil || view.Pending != nil || len(view.Events) != 1 || len(view.Gaps) != 0 || view.Events[0].Attempt != seq {
		t.Fatalf("duplicate or gap=%+v err=%v", view, err)
	}
}
