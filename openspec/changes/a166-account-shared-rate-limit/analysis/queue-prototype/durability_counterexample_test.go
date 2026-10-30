package queueprototype

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

// 디렉터리 fsync 실패가 rename 뒤에 발생한 상태를 모사함.
func TestDirectorySyncFailureBlocksOtherProcesses(t *testing.T) {
	q := setup(t)
	w, err := q.Register(true)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	q.Write = func(path string, b []byte) error {
		f, err := os.CreateTemp(q.Dir, "state-fault-")
		if err != nil {
			return err
		}
		defer os.Remove(f.Name())
		if _, err = f.Write(b); err != nil {
			f.Close()
			return err
		}
		if err = f.Sync(); err != nil {
			f.Close()
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
		if err = os.Rename(f.Name(), path); err != nil {
			return err
		}
		return errors.New("injected directory fsync EIO after rename")
	}
	if f, err := w.Try(); err == nil || f != nil {
		t.Fatal("selection must refuse")
	}
	other := &Queue{Dir: q.Dir}
	if f, err := other.TryAux(); err == nil && f != nil {
		unlock(f)
		t.Fatal("another participant admitted after uncertain directory fsync")
	}
}

func TestCanceledWaitDoesNotAcquireWire(t *testing.T) {
	q := setup(t)
	w, err := q.Register(true)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	f, err := w.Wait(ctx)
	if f != nil {
		unlock(f)
		t.Fatal("canceled waiter acquired wire")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestWaitCancelsWhileQueueLockHeld(t *testing.T) {
	q := setup(t)
	w, err := q.Register(true)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()
	held, err := lock(q.path("queue.lock"), syscall.LOCK_EX)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock(held)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		f, err := w.Wait(ctx)
		if f != nil {
			unlock(f)
		}
		done <- err
	}()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("queue lock ignored deadline")
	}
}
