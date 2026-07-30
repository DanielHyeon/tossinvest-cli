//go:build unix

package main

import (
	"os"
	"syscall"
	"testing"
)

func TestConsoleTerminationSignalsIncludeInterruptAndContainerTERM(t *testing.T) {
	got := consoleTerminationSignals()
	seen := map[os.Signal]bool{}
	for _, sig := range got {
		seen[sig] = true
	}
	if !seen[os.Interrupt] || !seen[syscall.SIGTERM] {
		t.Fatalf("signals = %v, want Interrupt and SIGTERM", got)
	}
}
