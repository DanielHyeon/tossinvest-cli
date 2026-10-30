package main

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestA166ConsoleRelaunchOwnerFailure(t *testing.T) {
	sentinel := errors.New("fake owner close failure")
	calls := 0
	err := consoleRelaunchWithExec(io.Discard, func() error { return sentinel }, func(string, []string) error { calls++; return nil })(12345)
	if !errors.Is(err, sentinel) || calls != 0 {
		t.Fatalf("owner failure lost: %v", err)
	}
}

func TestA166ConsoleRelaunchExecFailureAllowsOwnerReclose(t *testing.T) {
	owner := newRootBudgetOwner(func(string) (string, bool) { return "", false }, io.Discard)
	if _, err := owner.get(context.Background()); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("fake exec failure")
	calls := 0
	err := consoleRelaunchWithExec(io.Discard, owner.Close, func(string, []string) error { calls++; return sentinel })(12345)
	if !errors.Is(err, sentinel) || calls != 1 {
		t.Fatalf("exec err=%v calls=%d", err, calls)
	}
	if err := owner.Close(); err != nil {
		t.Fatalf("owner reclose: %v", err)
	}
}
