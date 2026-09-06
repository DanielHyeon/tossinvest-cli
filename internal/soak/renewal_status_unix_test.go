//go:build linux || darwin

package soak

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestLoadRenewalStatusRejectsFIFOWithoutWaitingForWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "renewal-status.fifo")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := LoadRenewalStatus(path, time.Now().UTC()); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO accepted as renewal status")
		}
	case <-time.After(time.Second):
		t.Fatal("FIFO status read waited for a writer")
	}
}
