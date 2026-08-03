package ratebudget

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestLeaseExcludesAnotherOwnerAndIsCrashStyleReleaseSafe(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	first, ok, err := TryAcquire(path)
	if err != nil || !ok {
		t.Fatalf("first lease = %v/%v", ok, err)
	}
	if second, ok, err := TryAcquire(path); err != nil || ok || second != nil {
		t.Fatalf("contended lease = %v/%v/%v", second, ok, err)
	}
	first.Release()
	first.Release()
	third, ok, err := TryAcquire(path)
	if err != nil || !ok {
		t.Fatalf("lease after release = %v/%v", ok, err)
	}
	third.Release()
}

func TestAcquireHonorsContextWhileAnotherProcessOwnsTheBudget(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	first, ok, err := TryAcquire(path)
	if err != nil || !ok {
		t.Fatalf("first lease = %v/%v", ok, err)
	}
	defer first.Release()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := Acquire(ctx, path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("contended Acquire error = %v, want deadline exceeded", err)
	}
}
