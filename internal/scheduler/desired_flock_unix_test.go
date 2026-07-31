//go:build unix

package scheduler

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

const desiredLockHelperPath = "TOSSOS_TEST_DESIRED_LOCK_PATH"

func TestDesiredStoreSaveSerializesAcrossProcesses(t *testing.T) {
	if path := os.Getenv(desiredLockHelperPath); path != "" {
		store := NewDesiredStore(path)
		desired, err := store.Load(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		desired.Enabled = false
		desired.AutoStart = false
		if err := store.Save(context.Background(), desired); err != nil {
			t.Fatal(err)
		}
		return
	}

	path := filepath.Join(t.TempDir(), "scheduler.json")
	store := NewDesiredStore(path)
	if err := store.Save(context.Background(), approvedDesired()); err != nil {
		t.Fatal(err)
	}
	lock, err := acquireDesiredLock(context.Background(), path+".lock")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestDesiredStoreSaveSerializesAcrossProcesses$")
	cmd.Env = append(os.Environ(), desiredLockHelperPath+"="+path)
	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		lock.release()
		t.Fatal(err)
	}
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		lock.release()
		t.Fatalf("subprocess bypassed desired-state flock: %v", err)
	case <-time.After(150 * time.Millisecond):
		// The child reached Save with revision 1 and is blocked on the kernel
		// lock held by this process.
	}
	lock.release()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("subprocess save after lock release: %v", err)
		}
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("subprocess did not finish after desired-state lock release")
	}

	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled || got.AutoStart || got.Revision != 2 {
		t.Fatalf("serialized subprocess did not persist OFF: %+v", got)
	}
}

func TestDesiredStoreSaveCancelsWhileWaitingForProcessLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.json")
	store := NewDesiredStore(path)
	if err := store.Save(context.Background(), approvedDesired()); err != nil {
		t.Fatal(err)
	}
	lock, err := acquireDesiredLock(context.Background(), path+".lock")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.release()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	started := time.Now()
	desired, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	desired.Enabled = false
	desired.AutoStart = false
	if err := store.Save(ctx, desired); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Save error = %v, want deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("lock cancellation took %s", elapsed)
	}
	got, err := store.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Enabled || got.Revision != 1 {
		t.Fatalf("canceled lock wait mutated desired state: %+v", got)
	}
}

func TestDesiredLockWaitIsBoundedWithoutCallerDeadline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scheduler.json")
	lock, err := acquireDesiredLock(context.Background(), path+".lock")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.release()

	started := time.Now()
	if _, err := acquireDesiredLock(context.Background(), path+".lock"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("bounded lock error = %v, want deadline exceeded", err)
	}
	elapsed := time.Since(started)
	if elapsed < desiredLockMaxWait || elapsed > desiredLockMaxWait+time.Second {
		t.Fatalf("bounded lock wait took %s, want %s..%s", elapsed, desiredLockMaxWait, desiredLockMaxWait+time.Second)
	}
}
