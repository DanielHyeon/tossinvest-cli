//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestOpenRejectsUnprovisionedScope(t *testing.T) {
	_, err := Open(Config{Root: t.TempDir(), Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unprovisioned root accepted: %v", err)
	}
}
func TestWaitRespectsCallerDeadline(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := Provision(root, "scope-a", "instance-a"); err != nil {
		t.Fatal(err)
	}
	b, err := Open(Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	first, err := b.Acquire(context.Background(), Critical)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err = b.Acquire(ctx, Critical)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline: %v", err)
	}
}
