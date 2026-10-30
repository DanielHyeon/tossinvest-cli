//go:build linux

package queueprototype

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestProtocolChild(t *testing.T) {
	mode := os.Getenv("PROTOCOL_CHILD")
	if mode == "" {
		return
	}
	q := &Queue{Dir: os.Getenv("PROTOCOL_DIR")}
	if mode == "owner" {
		f, e := lock(q.path("wire.lock"), syscall.LOCK_EX)
		if e != nil {
			t.Fatal(e)
		}
		defer unlock(f)
		if e = os.WriteFile(q.path("owner-ready"), []byte("1"), 0600); e != nil {
			t.Fatal(e)
		}
		time.Sleep(10 * time.Second)
		return
	}
	if mode == "aux" {
		f, e := q.TryAux()
		if f != nil {
			unlock(f)
			t.Fatal("aux admitted")
		}
		if e == nil {
			t.Fatal("aux not refused")
		}
		return
	}
	w, e := q.RegisterContext(context.Background(), mode == "critical")
	if w != nil {
		w.Close()
		t.Fatal("waiter admitted")
	}
	if e == nil {
		t.Fatal("waiter not refused")
	}
}

func TestPostRenameSyncFaultRejectsOtherProcesses(t *testing.T) {
	q := setup(t)
	w, e := q.RegisterContext(context.Background(), true)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	calls := 0
	q.SyncDir = func(path string) error {
		calls++
		if calls == 2 {
			return errors.New("injected directory sync EIO")
		}
		d, e := os.Open(path)
		if e != nil {
			return e
		}
		defer d.Close()
		return d.Sync()
	}
	f, e := w.Try()
	if f != nil {
		unlock(f)
		t.Fatal("wire admitted")
	}
	if e == nil {
		t.Fatal("missing fault")
	}
	for _, mode := range []string{"aux", "normal", "critical"} {
		c := exec.Command(os.Args[0], "-test.run=^TestProtocolChild$")
		c.Env = append(os.Environ(), "PROTOCOL_CHILD="+mode, "PROTOCOL_DIR="+q.Dir)
		if output, e := c.CombinedOutput(); e != nil {
			t.Fatalf("%s: %v %s", mode, e, output)
		}
	}
}

func TestSamePriorityFIFO(t *testing.T) {
	for _, critical := range []bool{false, true} {
		q := setup(t)
		first, e := q.RegisterContext(context.Background(), critical)
		if e != nil {
			t.Fatal(e)
		}
		defer first.Close()
		second, e := q.RegisterContext(context.Background(), critical)
		if e != nil {
			t.Fatal(e)
		}
		defer second.Close()
		if f, e := second.Try(); e != nil || f != nil {
			if f != nil {
				unlock(f)
			}
			t.Fatalf("second overtook: %v", e)
		}
		f, e := first.Try()
		if e != nil || f == nil {
			t.Fatalf("first: %v", e)
		}
		unlock(f)
		f, e = second.Try()
		if e != nil || f == nil {
			t.Fatalf("second: %v", e)
		}
		unlock(f)
	}
}

func TestKilledWireOwnerReleases(t *testing.T) {
	q := setup(t)
	c := exec.Command(os.Args[0], "-test.run=^TestProtocolChild$")
	c.Env = append(os.Environ(), "PROTOCOL_CHILD=owner", "PROTOCOL_DIR="+q.Dir)
	if e := c.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if c.ProcessState == nil {
			c.Process.Kill()
			c.Wait()
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, e := os.Stat(q.path("owner-ready")); e == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owner not ready")
		}
		time.Sleep(5 * time.Millisecond)
	}
	w, e := q.RegisterContext(context.Background(), true)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	if f, e := w.Try(); e != nil || f != nil {
		if f != nil {
			unlock(f)
		}
		t.Fatalf("owner bypassed: %v", e)
	}
	if e = c.Process.Kill(); e != nil {
		t.Fatal(e)
	}
	c.Wait()
	f, e := w.Try()
	if e != nil || f == nil {
		t.Fatalf("no takeover: %v", e)
	}
	unlock(f)
}

func TestDirtyStages(t *testing.T) {
	for _, failAt := range []int{1, 2, 3} {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			q := setup(t)
			w, e := q.RegisterContext(context.Background(), true)
			if e != nil {
				t.Fatal(e)
			}
			defer w.Close()
			before, e := os.ReadFile(q.path("state"))
			if e != nil {
				t.Fatal(e)
			}
			count := 0
			q.SyncDir = func(path string) error {
				count++
				if count == failAt {
					return errors.New("injected sync")
				}
				d, e := os.Open(path)
				if e != nil {
					return e
				}
				defer d.Close()
				return d.Sync()
			}
			f, e := w.Try()
			if f != nil {
				unlock(f)
				t.Fatal("wire admitted")
			}
			if e == nil {
				t.Fatal("fault ignored")
			}
			after, e := os.ReadFile(q.path("state"))
			if e != nil {
				t.Fatal(e)
			}
			_, marker := os.Stat(q.path("dirty"))
			if failAt == 1 && !bytes.Equal(before, after) {
				t.Fatal("state modified before marker durable")
			}
			if failAt < 3 && marker != nil {
				t.Fatal("dirty marker missing")
			}
			if failAt == 3 && !errors.Is(marker, os.ErrNotExist) {
				t.Fatal("marker should be removed")
			}
			peer := &Queue{Dir: q.Dir}
			got, e := peer.TryAux()
			if failAt < 3 && (e == nil || got != nil) {
				if got != nil {
					unlock(got)
				}
				t.Fatal("dirty peer admitted")
			}
			if failAt == 3 && (e != nil || got == nil) {
				t.Fatalf("durable state should allow peer: %v", e)
			}
			if got != nil {
				unlock(got)
			}
		})
	}
}

func TestCancelWhileWireBusy(t *testing.T) {
	q := setup(t)
	owner, e := lock(q.path("wire.lock"), syscall.LOCK_EX)
	if e != nil {
		t.Fatal(e)
	}
	defer unlock(owner)
	w, e := q.RegisterContext(context.Background(), true)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		f, e := w.Wait(ctx)
		if f != nil {
			unlock(f)
			done <- errors.New("canceled wait acquired wire")
			return
		}
		done <- e
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case e := <-done:
		if !errors.Is(e, context.Canceled) {
			t.Fatalf("got %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel timeout")
	}
}

func TestCloseRemovesTicketAndLiveness(t *testing.T) {
	q := setup(t)
	w, e := q.RegisterContext(context.Background(), true)
	if e != nil {
		t.Fatal(e)
	}
	id := w.id
	w.Close()
	if _, e := os.Stat(q.path("live-" + id)); !errors.Is(e, os.ErrNotExist) {
		t.Fatalf("liveness remains: %v", e)
	}
	f, e := q.TryAux()
	if e != nil || f == nil {
		t.Fatalf("stale canceled ticket: %v", e)
	}
	unlock(f)
}

func TestDeadTicketReclaimsLivenessAfterStateCommit(t *testing.T) {
	q := setup(t)
	dead, e := q.RegisterContext(context.Background(), true)
	if e != nil {
		t.Fatal(e)
	}
	id := dead.id
	unlock(dead.live)
	dead.live = nil
	w, e := q.RegisterContext(context.Background(), false)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	f, e := w.Try()
	if e != nil || f == nil {
		t.Fatalf("takeover: %v", e)
	}
	unlock(f)
	if _, e = os.Stat(q.path("live-" + id)); !errors.Is(e, os.ErrNotExist) {
		t.Fatalf("dead liveness remains: %v", e)
	}
}

func TestSelectedCloseDoesNotWaitForQueue(t *testing.T) {
	q := setup(t)
	w, e := q.RegisterContext(context.Background(), true)
	if e != nil {
		t.Fatal(e)
	}
	wire, e := w.Try()
	if e != nil || wire == nil {
		t.Fatalf("select: %v", e)
	}
	defer unlock(wire)
	held, e := lock(q.path("queue.lock"), syscall.LOCK_EX)
	if e != nil {
		t.Fatal(e)
	}
	done := make(chan struct{})
	go func() { w.Close(); close(done) }()
	select {
	case <-done:
		unlock(held)
	case <-time.After(100 * time.Millisecond):
		unlock(held)
		<-done
		t.Fatal("selected close waited for queue while holding wire")
	}
}
