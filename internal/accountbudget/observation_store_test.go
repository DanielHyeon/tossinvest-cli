//go:build linux

package accountbudget

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"golang.org/x/sys/unix"
)

type a166Boot struct{ stamp clock.BootStamp }

func (b *a166Boot) Snapshot() (clock.BootStamp, error) { return b.stamp, nil }
func a166Store(t *testing.T) (*Budget, *ObservationStore, *a166Boot) {
	t.Helper()
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := Provision(root, "scope", "instance"); err != nil {
		t.Fatal(err)
	}
	b, err := Open(Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Close() })
	id, err := ProvisionObservation(b)
	if err != nil {
		t.Fatal(err)
	}
	boot := &a166Boot{stamp: clock.BootStamp{BootID: "123e4567-e89b-42d3-a456-426614174000", Namespace: clock.BootNamespace{Device: 1, Inode: 1}, Uptime: time.Second}}
	store, err := OpenObservation(context.Background(), b, id, boot)
	if err != nil {
		t.Fatal(err)
	}
	return b, store, boot
}
func TestA166ObservationPendingAndFinish(t *testing.T) {
	b, s, _ := a166Store(t)
	lease, err := b.Acquire(context.Background(), Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	var key ObservationKey
	key[0] = 1
	attempt, err := s.Begin(context.Background(), lease, ObservationPoll, key)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || snap.Pending == nil || snap.Pending.Sequence != attempt {
		t.Fatalf("pending=%+v err=%v", snap, err)
	}
	hint := 10 * time.Second
	if err = s.Finish(context.Background(), lease, attempt, ObservationRateLimited, &hint); err != nil {
		t.Fatal(err)
	}
	remaining, err := s.PollRemaining(context.Background(), key)
	if err != nil || remaining != hint {
		t.Fatalf("remaining=%v err=%v", remaining, err)
	}
	snap, err = s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || snap.Pending != nil || len(snap.Events) != 1 {
		t.Fatalf("snapshot=%+v err=%v", snap, err)
	}
}

func TestA166ObservationLeaseAndOrphan(t *testing.T) {
	b, s, boot := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	var key ObservationKey
	key[0] = 1
	n, err := s.Begin(ctx, lease, ObservationPoll, key)
	if err != nil {
		t.Fatal(err)
	}
	other, err := OpenObservation(ctx, b, s.lockID, boot)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = other.Begin(ctx, lease, ObservationOther, ObservationKey{}); err == nil {
		t.Fatal("same lease began twice")
	}
	if err = other.Finish(ctx, &Lease{}, n, ObservationResponse, nil); err == nil {
		t.Fatal("forged lease finished")
	}
	snap, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || snap.PendingStatus != ObservationPendingBusy {
		t.Fatalf("live pending=%+v err=%v", snap, err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	if err = s.Finish(ctx, lease, n, ObservationNoResponse, nil); err == nil {
		t.Fatal("released lease finished")
	}
	snap, err = s.Snapshot(ctx, ObservationCursor{})
	if err != nil || snap.PendingStatus != ObservationPendingOrphan {
		t.Fatalf("orphan=%+v err=%v", snap, err)
	}
	next, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer next.Release()
	n2, err := s.Begin(ctx, next, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	snap, err = s.Snapshot(ctx, ObservationCursor{})
	if err != nil || len(snap.Gaps) != 1 || snap.Gaps[0].Attempt != n {
		t.Fatalf("gap=%+v err=%v", snap, err)
	}
	if _, err = s.PollRemaining(ctx, key); !errors.Is(err, ErrObservationPoisoned) {
		t.Fatalf("poison=%v", err)
	}
	if err = s.Finish(ctx, next, n2, ObservationNoResponse, nil); err != nil {
		t.Fatal(err)
	}
	snap, err = s.Snapshot(ctx, ObservationCursor{})
	if err != nil || snap.Pending != nil || len(snap.Gaps) != 1 {
		t.Fatalf("gap vanished: %+v %v", snap, err)
	}
}

func TestA166ObservationHintMaxAndRestore(t *testing.T) {
	b, s, boot := a166Store(t)
	ctx := context.Background()
	var key ObservationKey
	key[0] = 2
	send := func(hint time.Duration) {
		t.Helper()
		lease, e := b.Acquire(ctx, Critical)
		if e != nil {
			t.Fatal(e)
		}
		n, e := s.Begin(ctx, lease, ObservationPoll, key)
		if e != nil {
			t.Fatal(e)
		}
		if e = s.Finish(ctx, lease, n, ObservationRateLimited, &hint); e != nil {
			t.Fatal(e)
		}
		if e = lease.Release(); e != nil {
			t.Fatal(e)
		}
	}
	send(10 * time.Second)
	boot.stamp.Uptime += 3 * time.Second
	send(5 * time.Second)
	if got, e := s.PollRemaining(ctx, key); e != nil || got != 7*time.Second {
		t.Fatalf("max=%v %v", got, e)
	}
	boot.stamp.BootID = "123e4567-e89b-42d3-a456-426614174001"
	boot.stamp.Uptime = 500 * time.Hour
	if got, e := s.PollRemaining(ctx, key); e != nil || got != 10*time.Second {
		t.Fatalf("reboot=%v %v", got, e)
	}
	boot.stamp.Uptime += 2 * time.Second
	if got, e := s.PollRemaining(ctx, key); e != nil || got != 8*time.Second {
		t.Fatalf("anchor=%v %v", got, e)
	}
	boot.stamp.Uptime += 9 * time.Second
	if got, e := s.PollRemaining(ctx, key); e != nil || got != 0 {
		t.Fatalf("expired=%v %v", got, e)
	}
}

func TestA166ObservationFinishContextAndNoResponse(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	n, err := s.Begin(ctx, lease, ObservationToken, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Finish(ctx, lease, n, ObservationNoResponse, nil); err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || snap.Pending != nil || len(snap.Gaps) != 0 || len(snap.Events) != 0 || snap.Last == nil || snap.Last.Result != ObservationNoResponse || snap.Last.Attempt != n {
		t.Fatalf("normal no-response: %+v %v", snap, err)
	}
	lease, err = b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Begin(ctx, lease, ObservationToken, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err = s.Finish(canceled, lease, n, ObservationNoResponse, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel=%v", err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	snap, err = s.Snapshot(ctx, ObservationCursor{})
	if err != nil || snap.PendingStatus != ObservationPendingOrphan {
		t.Fatalf("pending lost: %+v %v", snap, err)
	}
}

func TestA166ObservationDiskFaults(t *testing.T) {
	for _, stage := range []string{"marker-sync", "state-write", "state-sync", "rename", "directory-before", "directory-after-rename", "directory-after-unlink"} {
		t.Run(stage, func(t *testing.T) {
			b, s, _ := a166Store(t)
			ctx := context.Background()
			lease, err := b.Acquire(ctx, Critical)
			if err != nil {
				t.Fatal(err)
			}
			defer lease.Release()
			sentinel := errors.New("injected fault")
			n := 0
			s.testFault = func(got string) error {
				if got == stage {
					return sentinel
				}
				return nil
			}
			s.syncDir = func(fd int) error {
				n++
				if (stage == "directory-before" && n == 1) || (stage == "directory-after-rename" && n == 2) || (stage == "directory-after-unlink" && n == 3) {
					return sentinel
				}
				return unix.Fsync(fd)
			}
			if _, err = s.Begin(ctx, lease, ObservationOther, ObservationKey{}); !errors.Is(err, sentinel) {
				t.Fatalf("fault lost: %v", err)
			}
			if _, err = s.Begin(ctx, lease, ObservationOther, ObservationKey{}); err == nil {
				t.Fatal("same lease retried")
			}
			s.testFault = nil
			s.syncDir = unix.Fsync
			_, err = s.Snapshot(ctx, ObservationCursor{})
			if stage == "directory-after-unlink" {
				if err != nil {
					t.Fatalf("durable state should remain readable: %v", err)
				}
			} else if err == nil {
				t.Fatal("dirty state accepted")
			}
		})
	}
}

func TestA166ObservationCursorAndRing(t *testing.T) {
	b, s, boot := a166Store(t)
	ctx := context.Background()
	state := observationState{Version: 1, Instance: "instance", Scope: "scope", Attempt: 128, Event: 128, Last: &ObservationCompletion{Attempt: 128, Kind: ObservationOther, Result: ObservationRateLimited}, Events: make([]ObservationEvent, 128), Gaps: []ObservationGap{}, Slots: []observationSlot{}}
	for i := range state.Events {
		state.Events[i] = ObservationEvent{Sequence: uint64(i + 1), Attempt: uint64(i + 1), Kind: ObservationOther, Stamp: boot.stamp}
	}
	if err := s.write(state); err != nil {
		t.Fatal(err)
	}
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	seq, err := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Finish(ctx, lease, seq, ObservationRateLimited, nil); err != nil {
		t.Fatal(err)
	}
	view, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || !view.EventOverrun || len(view.Events) != 128 || view.Events[0].Sequence != 2 || view.Cursor.Event != 129 {
		t.Fatalf("ring=%+v err=%v", view, err)
	}
	if _, err = s.Snapshot(ctx, ObservationCursor{Instance: "different", Event: 129}); err == nil {
		t.Fatal("foreign instance cursor accepted")
	}
	if _, err = s.Snapshot(ctx, ObservationCursor{Event: 1}); err == nil {
		t.Fatal("unbound cursor accepted")
	}
	if _, err = s.Snapshot(ctx, ObservationCursor{Scope: "scope", Instance: "instance", Event: 130}); err == nil {
		t.Fatal("future cursor accepted")
	}
}

func TestA166ObservationReleaseWaitsForBegin(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	resume := make(chan struct{})
	s.testFault = func(stage string) error {
		if stage == "marker-sync" {
			close(entered)
			<-resume
		}
		return nil
	}
	beginDone := make(chan error, 1)
	go func() { _, e := s.Begin(ctx, lease, ObservationOther, ObservationKey{}); beginDone <- e }()
	<-entered
	releaseDone := make(chan error, 1)
	go func() { releaseDone <- lease.Release() }()
	select {
	case err := <-releaseDone:
		t.Fatalf("release passed checkpoint: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(resume)
	if err = <-beginDone; err != nil {
		t.Fatal(err)
	}
	if err = <-releaseDone; err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || snap.PendingStatus != ObservationPendingOrphan {
		t.Fatalf("pending=%+v err=%v", snap, err)
	}
}

func TestA166ObservationProcessCrash(t *testing.T) {
	if os.Getenv("A166_OBS_CHILD") == "1" {
		root := os.Getenv("A166_OBS_ROOT")
		dev, e := strconv.ParseUint(os.Getenv("A166_OBS_DEV"), 10, 64)
		if e != nil {
			t.Fatal(e)
		}
		ino, e := strconv.ParseUint(os.Getenv("A166_OBS_INO"), 10, 64)
		if e != nil {
			t.Fatal(e)
		}
		b, e := Open(Config{Root: root, Scope: "scope", Instance: "instance", Clock: clock.System()})
		if e != nil {
			t.Fatal(e)
		}
		boot := &a166Boot{stamp: clock.BootStamp{BootID: "123e4567-e89b-42d3-a456-426614174000", Namespace: clock.BootNamespace{Device: 1, Inode: 1}, Uptime: time.Second}}
		s, e := OpenObservation(context.Background(), b, FileID{Device: dev, Inode: ino}, boot)
		if e != nil {
			t.Fatal(e)
		}
		lease, e := b.Acquire(context.Background(), Critical)
		if e != nil {
			t.Fatal(e)
		}
		var key ObservationKey
		key[0] = 7
		if _, e = s.Begin(context.Background(), lease, ObservationPoll, key); e != nil {
			t.Fatal(e)
		}
		os.Exit(0)
	}
	b, s, _ := a166Store(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestA166ObservationProcessCrash$")
	cmd.Env = append(os.Environ(), "A166_OBS_CHILD=1", "A166_OBS_ROOT="+b.root, fmt.Sprintf("A166_OBS_DEV=%d", s.lockID.Device), fmt.Sprintf("A166_OBS_INO=%d", s.lockID.Inode))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child: %v %s", err, out)
	}
	snap, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || snap.PendingStatus != ObservationPendingOrphan {
		t.Fatalf("orphan=%+v err=%v", snap, err)
	}
	lease, err := b.Acquire(context.Background(), Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err = s.Begin(context.Background(), lease, ObservationOther, ObservationKey{}); err != nil {
		t.Fatal(err)
	}
	snap, err = s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || len(snap.Gaps) != 1 {
		t.Fatalf("gap=%+v err=%v", snap, err)
	}
	var key ObservationKey
	key[0] = 7
	if _, err = s.PollRemaining(context.Background(), key); !errors.Is(err, ErrObservationPoisoned) {
		t.Fatalf("poison: %v", err)
	}
}

func TestA166ObservationRootAndLockSafety(t *testing.T) {
	b, s, boot := a166Store(t)
	ctx := context.Background()
	if _, err := OpenObservation(ctx, b, FileID{Device: s.lockID.Device, Inode: s.lockID.Inode + 1}, boot); err == nil {
		t.Fatal("wrong event inode accepted")
	}
	if _, err := ProvisionObservation(b); err == nil {
		t.Fatal("reprovision accepted")
	}
	if err := os.Rename(filepath.Join(b.root, observationLock), filepath.Join(b.root, "old-event.lock")); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(b.root, observationLock), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err = s.Snapshot(ctx, ObservationCursor{}); err == nil {
		t.Fatal("replaced event lock accepted")
	}
}
func TestA166ObservationRejectsCorruptionAndOverflow(t *testing.T) {
	t.Run("oversize", func(t *testing.T) {
		b, s, _ := a166Store(t)
		if err := os.WriteFile(filepath.Join(b.root, observationFile), bytes.Repeat([]byte("x"), observationMaxBytes+1), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Snapshot(context.Background(), ObservationCursor{}); err == nil {
			t.Fatal("oversize accepted")
		}
	})
	t.Run("overflow", func(t *testing.T) {
		b, s, _ := a166Store(t)
		state := observationState{Version: 1, Instance: "instance", Scope: "scope", Attempt: math.MaxUint64, Last: &ObservationCompletion{Attempt: math.MaxUint64, Kind: ObservationOther, Result: ObservationNoResponse}, Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: []observationSlot{}}
		if err := s.write(state); err != nil {
			t.Fatal(err)
		}
		lease, err := b.Acquire(context.Background(), Critical)
		if err != nil {
			t.Fatal(err)
		}
		defer lease.Release()
		if _, err = s.Begin(context.Background(), lease, ObservationOther, ObservationKey{}); err == nil {
			t.Fatal("attempt overflow accepted")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		_, _, boot := a166Store(t)
		state := observationState{Version: 1, Instance: "instance", Scope: "scope", Attempt: 2, Event: 2, Last: &ObservationCompletion{Attempt: 1, Kind: ObservationOther, Result: ObservationRateLimited}, Events: []ObservationEvent{{Sequence: 1, Attempt: 1, Kind: ObservationOther, Stamp: boot.stamp}, {Sequence: 2, Attempt: 1, Kind: ObservationOther, Stamp: boot.stamp}}, Gaps: []ObservationGap{}, Slots: []observationSlot{}}
		if err := validateObservationState(state); err == nil {
			t.Fatal("duplicate attempt accepted")
		}
	})
}
func TestA166ObservationExpiredSlotsAndPoisonCapacity(t *testing.T) {
	for _, poisoned := range []bool{false, true} {
		t.Run(fmt.Sprintf("poisoned-%v", poisoned), func(t *testing.T) {
			b, s, boot := a166Store(t)
			state := observationState{Version: 1, Instance: "instance", Scope: "scope", Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: make([]observationSlot, 256)}
			for i := range state.Slots {
				var key ObservationKey
				key[0] = byte(i)
				key[1] = 1
				state.Slots[i] = observationSlot{Key: key, Poisoned: poisoned, Hint: &observationHint{Stamp: boot.stamp, Duration: time.Second}}
			}
			if err := s.write(state); err != nil {
				t.Fatal(err)
			}
			boot.stamp.Uptime += 2 * time.Second
			lease, err := b.Acquire(context.Background(), Critical)
			if err != nil {
				t.Fatal(err)
			}
			defer lease.Release()
			var key ObservationKey
			key[0] = 3
			key[1] = 2
			n, err := s.Begin(context.Background(), lease, ObservationPoll, key)
			if err != nil {
				t.Fatal(err)
			}
			hint := time.Second
			err = s.Finish(context.Background(), lease, n, ObservationRateLimited, &hint)
			if poisoned {
				if err == nil {
					t.Fatal("poison discarded")
				}
			} else if err != nil {
				t.Fatalf("expired slots not reclaimed: %v", err)
			}
		})
	}
}

func TestA166ObservationSnapshotGenerationRace(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	n, err := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	read := make(chan struct{})
	resume := make(chan struct{})
	s.testAfterSnapshot = func() { close(read); <-resume }
	result := make(chan error, 1)
	go func() { _, e := s.Snapshot(ctx, ObservationCursor{}); result <- e }()
	<-read
	if err = s.Finish(ctx, lease, n, ObservationNoResponse, nil); err != nil {
		t.Fatal(err)
	}
	close(resume)
	if err = <-result; !errors.Is(err, ErrUnavailable) {
		t.Fatalf("generation drift accepted: %v", err)
	}
}
func TestA166ObservationReaderPinsBudget(t *testing.T) {
	b, s, _ := a166Store(t)
	held, err := s.lock(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, e := s.Snapshot(ctx, ObservationCursor{}); done <- e }()
	deadline := time.Now().Add(time.Second)
	active := false
	for time.Now().Before(deadline) {
		b.mu.Lock()
		active = b.active > 0
		b.mu.Unlock()
		if active {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !active {
		cancel()
		release(held)
		t.Fatal("reader did not pin")
	}
	if err = b.Close(); err == nil {
		cancel()
		release(held)
		t.Fatal("close passed active reader")
	}
	cancel()
	release(held)
	if err = <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel=%v", err)
	}
}

func TestA166ObservationGapRingOverrun(t *testing.T) {
	b, s, _ := a166Store(t)
	state := observationState{Version: 1, Instance: "instance", Scope: "scope", Attempt: 129, Gap: 128, Pending: &observationPending{Sequence: 129, Kind: ObservationOther}, Events: []ObservationEvent{}, Gaps: make([]ObservationGap, 128), Slots: []observationSlot{}}
	for i := range state.Gaps {
		state.Gaps[i] = ObservationGap{Sequence: uint64(i + 1), Attempt: uint64(i + 1), Kind: ObservationOther}
	}
	if err := s.write(state); err != nil {
		t.Fatal(err)
	}
	lease, err := b.Acquire(context.Background(), Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err = s.Begin(context.Background(), lease, ObservationOther, ObservationKey{}); err != nil {
		t.Fatal(err)
	}
	view, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || !view.GapOverrun || len(view.Gaps) != 128 || view.Gaps[0].Sequence != 2 || view.Gaps[127].Attempt != 129 {
		t.Fatalf("gap ring=%+v err=%v", view, err)
	}
}
func TestA166ObservationFinishDiskFailureStaysUnavailable(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	n, err := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("finish sync fault")
	s.testFault = func(stage string) error {
		if stage == "state-sync" {
			return sentinel
		}
		return nil
	}
	if err = s.Finish(ctx, lease, n, ObservationRateLimited, nil); !errors.Is(err, sentinel) {
		t.Fatalf("finish error=%v", err)
	}
	s.testFault = nil
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Snapshot(ctx, ObservationCursor{}); err == nil {
		t.Fatal("dirty finish accepted")
	}
	next, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer next.Release()
	if _, err = s.Begin(ctx, next, ObservationOther, ObservationKey{}); err == nil {
		t.Fatal("dirty finish permitted next checkpoint")
	}
}

func TestA166ObservationBeginCancellationDuringSync(t *testing.T) {
	b, s, _ := a166Store(t)
	lease, err := b.Acquire(context.Background(), Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	entered := make(chan struct{})
	resume := make(chan struct{})
	s.testFault = func(stage string) error {
		if stage == "marker-sync" {
			close(entered)
			<-resume
		}
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, e := s.Begin(ctx, lease, ObservationOther, ObservationKey{}); done <- e }()
	<-entered
	cancel()
	close(resume)
	if err = <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel hidden: %v", err)
	}
	view, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || view.Pending == nil {
		t.Fatalf("checkpoint lost: %+v %v", view, err)
	}
	if err = s.Finish(context.Background(), lease, view.Pending.Sequence, ObservationNoResponse, nil); err == nil {
		t.Fatal("failed Begin allowed guessed Finish")
	}
}

func TestA166ObservationMissingHintsAndPoisonStay(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	var key ObservationKey
	key[0] = 4
	send := func(result ObservationFinish, hint *time.Duration) {
		t.Helper()
		lease, e := b.Acquire(ctx, Critical)
		if e != nil {
			t.Fatal(e)
		}
		n, e := s.Begin(ctx, lease, ObservationPoll, key)
		if e != nil {
			t.Fatal(e)
		}
		if e = s.Finish(ctx, lease, n, result, hint); e != nil {
			t.Fatal(e)
		}
		if e = lease.Release(); e != nil {
			t.Fatal(e)
		}
	}
	hint := 12 * time.Second
	send(ObservationRateLimited, &hint)
	send(ObservationRateLimited, nil)
	send(ObservationResponse, nil)
	if got, err := s.PollRemaining(ctx, key); err != nil || got != hint {
		t.Fatalf("hint lost: %v %v", got, err)
	}
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Begin(ctx, lease, ObservationPoll, key); err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	lease, err = b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	n, err := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Finish(ctx, lease, n, ObservationNoResponse, nil); err != nil {
		t.Fatal(err)
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	send(ObservationRateLimited, &hint)
	if _, err = s.PollRemaining(ctx, key); !errors.Is(err, ErrObservationPoisoned) {
		t.Fatalf("poison cleared: %v", err)
	}
	view, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || len(view.Events) != 3 || len(view.Gaps) != 1 {
		t.Fatalf("events/gaps=%+v %v", view, err)
	}
}

func TestA166ObservationMissingRingIsCorrupt(t *testing.T) {
	state := observationState{Version: 1, Instance: "instance", Scope: "scope", Attempt: 1, Event: 1, Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: []observationSlot{}}
	if err := validateObservationState(state); err == nil {
		t.Fatal("missing event ring accepted")
	}
	state.Event = 0
	state.Gap = 1
	if err := validateObservationState(state); err == nil {
		t.Fatal("missing gap ring accepted")
	}
}

func TestA166ObservationCrossProcessPending(t *testing.T) {
	if os.Getenv("A166_OBS_LIVE_CHILD") == "1" {
		dev, e := strconv.ParseUint(os.Getenv("A166_OBS_DEV"), 10, 64)
		if e != nil {
			t.Fatal(e)
		}
		ino, e := strconv.ParseUint(os.Getenv("A166_OBS_INO"), 10, 64)
		if e != nil {
			t.Fatal(e)
		}
		b, e := Open(Config{Root: os.Getenv("A166_OBS_ROOT"), Scope: "scope", Instance: "instance", Clock: clock.System()})
		if e != nil {
			t.Fatal(e)
		}
		boot := &a166Boot{stamp: clock.BootStamp{BootID: "123e4567-e89b-42d3-a456-426614174000", Namespace: clock.BootNamespace{Device: 1, Inode: 1}, Uptime: time.Second}}
		s, e := OpenObservation(context.Background(), b, FileID{Device: dev, Inode: ino}, boot)
		if e != nil {
			t.Fatal(e)
		}
		lease, e := b.Acquire(context.Background(), Critical)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = s.Begin(context.Background(), lease, ObservationOther, ObservationKey{}); e != nil {
			t.Fatal(e)
		}
		fmt.Fprintln(os.Stdout, "ready")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(0)
	}
	b, s, _ := a166Store(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestA166ObservationCrossProcessPending$")
	cmd.Env = append(os.Environ(), "A166_OBS_LIVE_CHILD=1", "A166_OBS_ROOT="+b.root, fmt.Sprintf("A166_OBS_DEV=%d", s.lockID.Device), fmt.Sprintf("A166_OBS_INO=%d", s.lockID.Inode))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { stdin.Close(); cmd.Wait() }()
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil || line != "ready\n" {
		t.Fatalf("child readiness: %q %v", line, err)
	}
	view, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || view.PendingStatus != ObservationPendingBusy {
		t.Fatalf("live pending=%+v err=%v", view, err)
	}
	if _, err = fmt.Fprintln(stdin, "exit"); err != nil {
		t.Fatal(err)
	}
	if err = cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	view, err = s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || view.PendingStatus != ObservationPendingOrphan {
		t.Fatalf("dead pending=%+v err=%v", view, err)
	}
	lease, err := b.Acquire(context.Background(), Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err = s.Begin(context.Background(), lease, ObservationOther, ObservationKey{}); err != nil {
		t.Fatal(err)
	}
	view, err = s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || len(view.Gaps) != 1 {
		t.Fatalf("gap=%+v err=%v", view, err)
	}
}

func TestA166ObservationForeignStateRejected(t *testing.T) {
	b, s, _ := a166Store(t)
	state := observationState{Version: 1, Instance: "other", Scope: "scope", Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: []observationSlot{}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(b.root, observationFile), data, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Snapshot(context.Background(), ObservationCursor{}); err == nil {
		t.Fatal("foreign state accepted")
	}
}

func TestA166ObservationInvalidHintStillRecords429(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	var key ObservationKey
	key[0] = 9
	send := func(result ObservationFinish, hint *time.Duration) {
		t.Helper()
		lease, e := b.Acquire(ctx, Critical)
		if e != nil {
			t.Fatal(e)
		}
		defer lease.Release()
		n, e := s.Begin(ctx, lease, ObservationPoll, key)
		if e != nil {
			t.Fatal(e)
		}
		if e = s.Finish(ctx, lease, n, result, hint); e != nil {
			t.Fatal(e)
		}
	}
	valid := 12 * time.Second
	negative := -time.Second
	tooLong := 31 * time.Second
	zero := time.Duration(0)
	send(ObservationRateLimited, &valid)
	send(ObservationRateLimited, &negative)
	send(ObservationRateLimited, &tooLong)
	send(ObservationRateLimited, &zero)
	send(ObservationResponse, &tooLong)
	if got, err := s.PollRemaining(ctx, key); err != nil || got != valid {
		t.Fatalf("existing hint lost: %v %v", got, err)
	}
	view, err := s.Snapshot(ctx, ObservationCursor{})
	if err != nil || len(view.Events) != 4 || view.Pending != nil {
		t.Fatalf("events=%+v err=%v", view, err)
	}
}

func TestA166ObservationReservesFinishGeneration(t *testing.T) {
	b, s, _ := a166Store(t)
	ctx := context.Background()
	state := observationState{Version: 1, Scope: "scope", Instance: "instance", Generation: math.MaxUint64 - 2, Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: []observationSlot{}}
	if err := s.write(state); err != nil {
		t.Fatal(err)
	} // generation max-1
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	if _, err = s.Begin(ctx, lease, ObservationOther, ObservationKey{}); err == nil {
		t.Fatal("Begin exhausted Finish generation")
	}
}

func TestA166ObservationPendingReservesFinishGeneration(t *testing.T) {
	b, s, boot := a166Store(t)
	ctx := context.Background()
	var key ObservationKey
	key[0] = 1
	state := observationState{Version: 1, Scope: "scope", Instance: "instance", Generation: math.MaxUint64 - 3, Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: []observationSlot{{Key: key, Hint: &observationHint{Stamp: boot.stamp, Duration: time.Second}}}}
	if err := s.write(state); err != nil {
		t.Fatal(err)
	}
	boot.stamp.Uptime += 2 * time.Second
	lease, err := b.Acquire(ctx, Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release()
	n, err := s.Begin(ctx, lease, ObservationOther, ObservationKey{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.PollRemaining(ctx, key); err == nil {
		t.Fatal("reader spent finish generation")
	}
	if err = s.Finish(ctx, lease, n, ObservationNoResponse, nil); err != nil {
		t.Fatalf("reserved finish failed: %v", err)
	}
}

func TestA166ObservationFullBoundedState(t *testing.T) {
	_, s, boot := a166Store(t)
	state := observationState{Version: 1, Scope: "scope", Instance: "instance", Attempt: 256, Event: 128, Gap: 128, Last: &ObservationCompletion{Attempt: 128, Kind: ObservationPoll, Result: ObservationRateLimited}, Events: make([]ObservationEvent, 128), Gaps: make([]ObservationGap, 128), Slots: make([]observationSlot, 256)}
	for i := 0; i < 128; i++ {
		state.Events[i] = ObservationEvent{Sequence: uint64(i + 1), Attempt: uint64(i + 1), Kind: ObservationPoll, Stamp: boot.stamp}
		state.Gaps[i] = ObservationGap{Sequence: uint64(i + 1), Attempt: uint64(i + 129), Kind: ObservationOther}
	}
	for i := range state.Slots {
		var key ObservationKey
		key[0] = byte(i)
		key[1] = 1
		state.Slots[i] = observationSlot{Key: key, Hint: &observationHint{Stamp: boot.stamp, Duration: 30 * time.Second}}
	}
	raw, err := json.Marshal(state)
	if err != nil || len(raw) > observationMaxBytes {
		t.Fatalf("bounded state size=%d err=%v", len(raw), err)
	}
	if err = s.write(state); err != nil {
		t.Fatal(err)
	}
	view, err := s.Snapshot(context.Background(), ObservationCursor{})
	if err != nil || len(view.Events) != 128 || len(view.Gaps) != 128 {
		t.Fatalf("full state: %v", err)
	}
}
