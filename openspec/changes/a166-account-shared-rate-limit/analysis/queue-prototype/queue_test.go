package queueprototype

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func setup(t *testing.T) *Queue {
	t.Helper()
	d := t.TempDir()
	if e := os.WriteFile(filepath.Join(d, "state"), []byte(`{"Next":0,"Tickets":[]}`), 0600); e != nil {
		t.Fatal(e)
	}
	return &Queue{Dir: d}
}
func TestPriorityAndLiveOwner(t *testing.T) {
	q := setup(t)
	owner, e := lock(q.path("wire.lock"), syscall.LOCK_EX)
	if e != nil {
		t.Fatal(e)
	}
	n, e := q.Register(false)
	if e != nil {
		t.Fatal(e)
	}
	defer n.Close()
	c, e := q.Register(true)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	if f, e := n.Try(); e != nil || f != nil {
		t.Fatalf("normal overtook: %v", e)
	}
	unlock(owner)
	if f, e := n.Try(); e != nil || f != nil {
		t.Fatalf("normal overtook: %v", e)
	}
	f, e := c.Try()
	if e != nil || f == nil {
		t.Fatalf("critical: %v", e)
	}
	unlock(f)
	f, e = n.Try()
	if e != nil || f == nil {
		t.Fatalf("normal: %v", e)
	}
	unlock(f)
}
func TestFaultDoesNotTransferWire(t *testing.T) {
	q := setup(t)
	w, e := q.Register(true)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	q.Write = func(string, []byte) error { return errors.New("write fault") }
	if f, e := w.Try(); e == nil || f != nil {
		t.Fatal("expected refusal")
	}
	f, e := lock(q.path("wire.lock"), syscall.LOCK_EX|syscall.LOCK_NB)
	if e != nil {
		t.Fatal(e)
	}
	unlock(f)
}
func TestCancelAndDeadTicket(t *testing.T) {
	q := setup(t)
	w, e := q.Register(true)
	if e != nil {
		t.Fatal(e)
	}
	w.Close()
	n, e := q.Register(false)
	if e != nil {
		t.Fatal(e)
	}
	defer n.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	f, e := n.Wait(ctx)
	if e != nil {
		t.Fatal(e)
	}
	unlock(f)
}

func TestChildProcess(t *testing.T) {
	if os.Getenv("QUEUE_CHILD") == "" {
		return
	}
	q := &Queue{Dir: os.Getenv("QUEUE_DIR")}
	w, e := q.Register(os.Getenv("QUEUE_CHILD") == "critical")
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	if e = os.WriteFile(q.path("ready-"+os.Getenv("QUEUE_CHILD")), []byte("1"), 0600); e != nil {
		t.Fatal(e)
	}
	if os.Getenv("QUEUE_HOLD_AFTER_REGISTER") == "1" {
		time.Sleep(10 * time.Second)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	f, e := w.Wait(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer unlock(f)
	if os.Getenv("QUEUE_CHILD") == "normal" {
		if _, e := os.Stat(q.path("won-critical")); e != nil {
			t.Fatal("normal overtook critical")
		}
	}
	if e = os.WriteFile(q.path("won-"+os.Getenv("QUEUE_CHILD")), []byte("1"), 0600); e != nil {
		t.Fatal(e)
	}
}
func TestProcessesPriorityAndDeath(t *testing.T) {
	q := setup(t)
	owner, e := lock(q.path("wire.lock"), syscall.LOCK_EX)
	if e != nil {
		t.Fatal(e)
	}
	defer func() {
		if owner != nil {
			unlock(owner)
		}
	}()
	start := func(kind string) *exec.Cmd {
		c := exec.Command(os.Args[0], "-test.run=^TestChildProcess$")
		c.Env = append(os.Environ(), "QUEUE_CHILD="+kind, "QUEUE_DIR="+q.Dir)
		if e := c.Start(); e != nil {
			t.Fatal(e)
		}
		t.Cleanup(func() {
			if c.ProcessState == nil {
				c.Process.Kill()
				c.Wait()
			}
		})
		deadline := time.Now().Add(3 * time.Second)
		for {
			if _, e := os.Stat(q.path("ready-" + kind)); e == nil {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("child failed to register")
			}
			time.Sleep(5 * time.Millisecond)
		}
		return c
	}
	normal := start("normal")
	critical := start("critical")
	if _, e := os.Stat(q.path("won-critical")); e == nil {
		t.Fatal("owner bypassed")
	}
	unlock(owner)
	owner = nil
	if e := critical.Wait(); e != nil {
		t.Fatal(e)
	}
	if e := normal.Wait(); e != nil {
		t.Fatal(e)
	}
}

func TestKilledWaiterCleanup(t *testing.T) {
	q := setup(t)
	owner, e := lock(q.path("wire.lock"), syscall.LOCK_EX)
	if e != nil {
		t.Fatal(e)
	}
	defer unlock(owner)
	c := exec.Command(os.Args[0], "-test.run=^TestChildProcess$")
	c.Env = append(os.Environ(), "QUEUE_CHILD=critical", "QUEUE_DIR="+q.Dir, "QUEUE_HOLD_AFTER_REGISTER=1")
	if e = c.Start(); e != nil {
		t.Fatal(e)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, e = os.Stat(q.path("ready-critical")); e == nil {
			break
		}
		if time.Now().After(deadline) {
			c.Process.Kill()
			c.Wait()
			t.Fatal("no registration")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if e = c.Process.Kill(); e != nil {
		t.Fatal(e)
	}
	c.Wait()
	w, e := q.Register(false)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	if f, e := w.Try(); e != nil || f != nil {
		t.Fatalf("busy wire: %v", e)
	}
	s, e := q.read()
	if e != nil || len(s.Tickets) != 1 || s.Tickets[0].ID != w.id {
		t.Fatalf("dead ticket remains: %v", e)
	}
}

func TestAuxRefusesQueuedWaiter(t *testing.T) {
	q := setup(t)
	w, e := q.Register(true)
	if e != nil {
		t.Fatal(e)
	}
	defer w.Close()
	f, e := q.TryAux()
	if e != nil || f != nil {
		t.Fatalf("aux passed queued waiter: %v", e)
	}
}
func TestCorruptStateRefuses(t *testing.T) {
	q := setup(t)
	if e := os.WriteFile(q.path("state"), []byte("{"), 0600); e != nil {
		t.Fatal(e)
	}
	if f, e := q.TryAux(); e == nil || f != nil {
		t.Fatal("corruption admitted")
	}
}
