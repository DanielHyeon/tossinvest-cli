//go:build linux

package strategyobserve

import (
	"bytes"
	"encoding/json"
	"errors"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreReplayDuplicateAndTornRefusal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	line := testLine(t, "one")
	if dup, err := s.Append(line); err != nil || dup {
		t.Fatalf("append %v %v", dup, err)
	}
	if dup, err := s.Append(line); err != nil || !dup {
		t.Fatalf("duplicate %v %v", dup, err)
	}
	if _, err = Open(path); err == nil {
		t.Fatal("second writer acquired lock")
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if dup, err := s.Append(line); err != nil || !dup {
		t.Fatalf("restart duplicate %v %v", dup, err)
	}
	s.Close()
	ledger := filepath.Join(path, "ledger.jsonl")
	f, err := os.OpenFile(ledger, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"partial":true}`)
	f.Close()
	before, _ := os.ReadFile(ledger)
	if _, err = Open(path); err == nil {
		t.Fatal("torn tail accepted")
	}
	after, _ := os.ReadFile(ledger)
	if !bytes.Equal(before, after) {
		t.Fatal("torn evidence modified")
	}
}
func TestStoreRejectsSymlinkHardlinkAndForeign(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	os.Mkdir(real, 0700)
	link := filepath.Join(root, "alias")
	os.Symlink(real, link)
	if _, err := Open(filepath.Join(link, "out")); err == nil {
		t.Fatal("symlink ancestor")
	}
	foreign := filepath.Join(root, "foreign")
	os.Mkdir(foreign, 0700)
	os.WriteFile(filepath.Join(foreign, "existing"), []byte("keep"), 0600)
	if _, err := Open(foreign); err == nil {
		t.Fatal("foreign dir")
	}
	b, _ := os.ReadFile(filepath.Join(foreign, "existing"))
	if string(b) != "keep" {
		t.Fatal("foreign changed")
	}
	protected := filepath.Join(root, "a128-evidence")
	if _, err := Open(protected); err == nil {
		t.Fatal("protected path")
	}
	path := filepath.Join(root, "out")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	os.Link(filepath.Join(path, "ledger.jsonl"), filepath.Join(root, "hardlink"))
	if _, err = Open(path); err == nil {
		t.Fatal("hardlink accepted")
	}
}
func TestWriterFaultsLatchAndRecovery(t *testing.T) {
	for _, stage := range []string{"short-write", "append-sync", "checkpoint-sync", "checkpoint-rename", "directory-sync"} {
		t.Run(stage, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "out")
			s, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			if stage == "short-write" {
				s.write = func(f *os.File, b []byte) (int, error) {
					n, e := f.Write(b[:len(b)/2])
					if e != nil {
						return n, e
					}
					return n, io.ErrShortWrite
				}
			} else {
				s.fault = func(at string) error {
					if at == stage {
						return errors.New("injected " + stage)
					}
					return nil
				}
			}
			if _, err = s.Append(testLine(t, "x")); err == nil {
				t.Fatal("fault ignored")
			}
			if _, err = s.Append(testLine(t, "y")); err == nil {
				t.Fatal("writer did not latch")
			}
			s.Close()
			ledger := filepath.Join(path, "ledger.jsonl")
			before, _ := os.ReadFile(ledger)
			reopened, err := Open(path)
			if stage == "short-write" || stage == "checkpoint-sync" || stage == "checkpoint-rename" {
				if err == nil {
					reopened.Close()
					t.Fatal("ambiguous recovery accepted")
				}
			} else if err != nil {
				t.Fatal(err)
			} else {
				reopened.Close()
			}
			after, _ := os.ReadFile(ledger)
			if !bytes.Equal(before, after) {
				t.Fatal("recovery modified ledger")
			}
		})
	}
}

func TestCheckpointRecoveryAndCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	first, second := testLine(t, "first"), testLine(t, "second")
	if _, err = s.Append(first); err != nil {
		t.Fatal(err)
	}
	cpPath := filepath.Join(path, "checkpoint.json")
	old, err := os.ReadFile(cpPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Append(second); err != nil {
		t.Fatal(err)
	}
	s.Close()
	// A crash may leave a valid ledger ahead of its checkpoint.
	if err = os.WriteFile(cpPath, old, 0600); err != nil {
		t.Fatal(err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if dup, err := s.Append(second); err != nil || !dup {
		t.Fatalf("lagging checkpoint replay: %v %v", dup, err)
	}
	s.Close()
	ledgerPath := filepath.Join(path, "ledger.jsonl")
	before, _ := os.ReadFile(ledgerPath)
	for _, bad := range [][]byte{append(append([]byte(nil), old...), []byte(` {}`)...), []byte(`{"schema":"strategyobserve-v1","count":1,"last_id":"x","offset":999999,"prefix":"x"}`)} {
		if err = os.WriteFile(cpPath, bad, 0600); err != nil {
			t.Fatal(err)
		}
		if s, err = Open(path); err == nil {
			s.Close()
			t.Fatal("bad checkpoint accepted")
		}
		after, _ := os.ReadFile(ledgerPath)
		if !bytes.Equal(before, after) {
			t.Fatal("evidence altered")
		}
	}
}

func TestSameIDConflictingPayloadFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	line := testLine(t, "one")
	if _, err = s.Append(line); err != nil {
		t.Fatal(err)
	}
	var e Envelope
	if err = json.Unmarshal(line, &e); err != nil {
		t.Fatal(err)
	}
	var d Diagnostic
	if err = json.Unmarshal(e.Payload, &d); err != nil {
		t.Fatal(err)
	}
	d.Parent = "different"
	payload, _ := json.Marshal(d)
	e.Payload = payload
	e.PayloadDigest = digest(payload)
	conflict, _ := json.Marshal(e)
	conflict = append(conflict, '\n')
	if _, err = s.Append(conflict); err == nil {
		t.Fatal("conflict accepted")
	}
	if _, err = s.Append(testLine(t, "other")); err == nil {
		t.Fatal("conflict did not latch")
	}
}

func TestReviewFIFORefusedWithoutBlocking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	markerPath := filepath.Join(path, "marker")
	if err = os.Remove(markerPath); err != nil {
		t.Fatal(err)
	}
	if err = unix.Mkfifo(markerPath, 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		s, e := Open(path)
		if e == nil {
			s.Close()
		}
		done <- e
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("fifo accepted")
		}
	case <-time.After(300 * time.Millisecond):
		fd, e := unix.Open(markerPath, unix.O_WRONLY|unix.O_NONBLOCK, 0)
		if e == nil {
			unix.Close(fd)
		}
		<-done
		t.Fatal("fifo open blocked")
	}
}

func TestReviewRecoveryEstablishesDurabilityBeforeDuplicate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.fault = func(stage string) error {
		if stage == "append-sync" {
			return errors.New("sync failed")
		}
		return nil
	}
	line := testLine(t, "unconfirmed")
	if _, err = s.Append(line); err == nil {
		t.Fatal("expected failure")
	}
	s.Close()
	s, err = open(path, func(stage string) error {
		if stage == "recovery-sync" {
			return errors.New("injected recovery sync failure")
		}
		return nil
	})
	if err == nil {
		s.Close()
		t.Fatal("recovery sync failure ignored")
	}
	if _, err = os.Stat(filepath.Join(path, "checkpoint.json")); !os.IsNotExist(err) {
		t.Fatalf("checkpoint advanced after failed sync: %v", err)
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = os.Stat(filepath.Join(path, "checkpoint.json")); err != nil {
		t.Fatalf("recovery did not establish checkpoint: %v", err)
	}
	if dup, err := s.Append(line); err != nil || !dup {
		t.Fatalf("duplicate: %v %v", dup, err)
	}
}

func TestReviewMatchingCheckpointRecoverySyncFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	s.fault = func(stage string) error {
		if stage == "directory-sync" {
			return errors.New("directory sync failed")
		}
		return nil
	}
	line := testLine(t, "x")
	if _, err = s.Append(line); err == nil {
		t.Fatal("expected failure")
	}
	s.Close()
	s, err = open(path, func(stage string) error {
		if stage == "recovery-directory-sync" {
			return errors.New("recovery dir sync failed")
		}
		return nil
	})
	if err == nil {
		s.Close()
		t.Fatal("recovery barrier ignored")
	}
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if dup, err := s.Append(line); err != nil || !dup {
		t.Fatalf("duplicate: %v %v", dup, err)
	}
}
