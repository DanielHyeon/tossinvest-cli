//go:build linux

package accountbudget

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"golang.org/x/sys/unix"
)

func TestInvalidFilesFailClosed(t *testing.T) {
	for _, mode := range []string{"scope", "instance", "hardlink", "symlink", "lock-replaced", "dirty", "corrupt", "overflow"} {
		t.Run(mode, func(t *testing.T) {
			b, root := fixture(t)
			switch mode {
			case "scope":
				os.WriteFile(filepath.Join(root, "scope"), []byte("other"), 0600)
			case "instance":
				os.WriteFile(filepath.Join(root, "instance"), []byte("other"), 0600)
			case "hardlink":
				if e := os.Link(filepath.Join(root, "wire.lock"), filepath.Join(root, "copy")); e != nil {
					t.Fatal(e)
				}
			case "symlink":
				if e := os.Remove(filepath.Join(root, "state")); e != nil {
					t.Fatal(e)
				}
				if e := os.Symlink("scope", filepath.Join(root, "state")); e != nil {
					t.Fatal(e)
				}
			case "lock-replaced":
				if e := os.Rename(filepath.Join(root, "queue.lock"), filepath.Join(root, "old")); e != nil {
					t.Fatal(e)
				}
				if e := os.WriteFile(filepath.Join(root, "queue.lock"), nil, 0600); e != nil {
					t.Fatal(e)
				}
			case "dirty":
				os.WriteFile(filepath.Join(root, "dirty"), nil, 0600)
			case "corrupt":
				os.WriteFile(filepath.Join(root, "state"), []byte("{"), 0600)
			case "overflow":
				os.WriteFile(filepath.Join(root, "state"), []byte(`{"next":18446744073709551615,"tickets":[]}`), 0600)
			}
			lease, e := b.Acquire(context.Background(), Critical)
			if lease != nil {
				lease.Release()
				t.Fatal("unsafe state admitted")
			}
			if !errors.Is(e, ErrUnavailable) {
				t.Fatalf("got %v", e)
			}
		})
	}
}
func TestCloseRejectsLiveLease(t *testing.T) {
	b, _ := fixture(t)
	l, e := b.Acquire(context.Background(), Critical)
	if e != nil {
		t.Fatal(e)
	}
	if e = b.Close(); e == nil {
		t.Fatal("closed active budget")
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
}
func TestProcessHelper(t *testing.T) {
	root := os.Getenv("BUDGET_PROCESS_ROOT")
	if root == "" {
		return
	}
	b, e := Open(Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if e != nil {
		t.Fatal(e)
	}
	defer b.Close()
	l, e := b.Acquire(context.Background(), Critical)
	if e != nil {
		t.Fatal(e)
	}
	defer l.Release()
	if e = os.WriteFile(filepath.Join(root, "test-ready"), []byte("ready"), 0600); e != nil {
		t.Fatal(e)
	}
	time.Sleep(10 * time.Second)
}
func TestKilledProcessReleasesWire(t *testing.T) {
	b, root := fixture(t)
	c := exec.Command(os.Args[0], "-test.run=^TestProcessHelper$")
	c.Env = append(os.Environ(), "BUDGET_PROCESS_ROOT="+root)
	if e := c.Start(); e != nil {
		t.Fatal(e)
	}
	defer func() {
		if c.ProcessState == nil {
			c.Process.Kill()
			c.Wait()
		}
	}()
	deadline := time.Now().Add(4 * time.Second)
	for {
		if _, e := os.Stat(filepath.Join(root, "test-ready")); e == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("child not ready")
		}
		time.Sleep(5 * time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	if l, e := b.Acquire(ctx, Critical); !errors.Is(e, context.DeadlineExceeded) {
		if l != nil {
			l.Release()
		}
		t.Fatalf("busy wire: %v", e)
	}
	if e := c.Process.Kill(); e != nil {
		t.Fatal(e)
	}
	c.Wait()
	l, e := b.Acquire(context.Background(), Critical)
	if e != nil {
		t.Fatal(e)
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
}
func TestQueueBounds(t *testing.T) {
	b, root := fixture(t)
	_ = b
	entries := ""
	for i := 0; i < 257; i++ {
		if i > 0 {
			entries += ","
		}
		entries += fmt.Sprintf(`{"id":"%032x","seq":%d,"critical":false}`, i, i)
	}
	if e := os.WriteFile(filepath.Join(root, "state"), []byte(fmt.Sprintf(`{"next":257,"tickets":[%s]}`, entries)), 0600); e != nil {
		t.Fatal(e)
	}
	l, e := b.Acquire(context.Background(), Critical)
	if l != nil {
		l.Release()
		t.Fatal("oversized queue admitted")
	}
	if !errors.Is(e, ErrUnavailable) {
		t.Fatal(e)
	}
}

func TestPriorityOvertakesQueuedNormal(t *testing.T) {
	b, root := fixture(t)
	owner, e := b.Acquire(context.Background(), Normal)
	if e != nil {
		t.Fatal(e)
	}
	type result struct {
		name  string
		lease *Lease
		err   error
	}
	done := make(chan result, 2)
	start := func(name string, p Priority) {
		go func() { l, e := b.Acquire(context.Background(), p); done <- result{name, l, e} }()
	}
	await := func(n int) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for {
			data, e := os.ReadFile(filepath.Join(root, "state"))
			if e == nil {
				var s queueState
				if json.Unmarshal(data, &s) == nil && len(s.Tickets) >= n {
					return
				}
			}
			if time.Now().After(deadline) {
				t.Fatal("queue registration timeout")
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	start("normal", Normal)
	await(1)
	start("critical", Critical)
	await(2)
	if e = owner.Release(); e != nil {
		t.Fatal(e)
	}
	first := <-done
	if first.err != nil || first.name != "critical" {
		t.Fatalf("first %s %v", first.name, first.err)
	}
	if e = first.lease.Release(); e != nil {
		t.Fatal(e)
	}
	second := <-done
	if second.err != nil || second.name != "normal" {
		t.Fatalf("second %s %v", second.name, second.err)
	}
	if e = second.lease.Release(); e != nil {
		t.Fatal(e)
	}
}
func TestInstanceMismatchRefusesOpen(t *testing.T) {
	b, root := fixture(t)
	_ = b
	_, e := Open(Config{Root: root, Scope: "scope-a", Instance: "wrong-instance", Clock: clock.System()})
	if !errors.Is(e, ErrUnavailable) {
		t.Fatalf("instance mismatch %v", e)
	}
}
func TestDefaultPriorityAndInheritance(t *testing.T) {
	if PriorityFromContext(context.Background()) != Normal {
		t.Fatal("default not normal")
	}
	ctx := WithPriority(context.Background(), Critical)
	derived, cancel := context.WithCancel(ctx)
	defer cancel()
	if PriorityFromContext(derived) != Critical {
		t.Fatal("priority lost")
	}
}

func TestDirtyFaultHelper(t *testing.T) {
	root := os.Getenv("BUDGET_DIRTY_ROOT")
	if root == "" {
		return
	}
	b, e := Open(Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if b != nil {
		b.Close()
		t.Fatal("dirty state opened")
	}
	if !errors.Is(e, ErrUnavailable) {
		t.Fatalf("got %v", e)
	}
}
func TestDirtySyncFaultBlocksOtherProcess(t *testing.T) {
	b, root := fixture(t)
	calls := 0
	b.testSyncDir = func(fd int) error {
		calls++
		if calls == 2 {
			return errors.New("injected directory fsync error after rename")
		}
		return unix.Fsync(fd)
	}
	l, e := b.Acquire(context.Background(), Critical)
	if l != nil {
		l.Release()
		t.Fatal("fault admitted")
	}
	if !errors.Is(e, ErrUnavailable) {
		t.Fatalf("fault: %v", e)
	}
	if _, e = os.Stat(filepath.Join(root, "dirty")); e != nil {
		t.Fatalf("missing dirty: %v", e)
	}
	c := exec.Command(os.Args[0], "-test.run=^TestDirtyFaultHelper$")
	c.Env = append(os.Environ(), "BUDGET_DIRTY_ROOT="+root)
	if out, e := c.CombinedOutput(); e != nil {
		t.Fatalf("peer: %v %s", e, out)
	}
}
func TestMarkerSyncFaultLeavesStateUnchanged(t *testing.T) {
	b, root := fixture(t)
	before, e := os.ReadFile(filepath.Join(root, "state"))
	if e != nil {
		t.Fatal(e)
	}
	b.testSyncDir = func(int) error { return errors.New("injected marker sync error") }
	l, e := b.Acquire(context.Background(), Normal)
	if l != nil {
		l.Release()
		t.Fatal("fault admitted")
	}
	if !errors.Is(e, ErrUnavailable) {
		t.Fatal(e)
	}
	after, e := os.ReadFile(filepath.Join(root, "state"))
	if e != nil || !bytes.Equal(before, after) {
		t.Fatal("state changed before durable marker")
	}
}

var errTestDelay = errors.New("delay failure")
var errTestRelease = errors.New("wire cleanup failure")

type delayFailureClock struct{ clock.Clock }

func (delayFailureClock) Sleep(context.Context, time.Duration) error { return errTestDelay }
func TestWireCleanupErrorIsReported(t *testing.T) {
	b, _ := fixture(t)
	b.clk = delayFailureClock{clock.System()}
	b.testWireRelease = func(f *os.File) error { return errors.Join(release(f), errTestRelease) }
	l, e := b.Acquire(context.Background(), Critical)
	if l != nil {
		l.Release()
		t.Fatal("failed cleanup returned lease")
	}
	if !errors.Is(e, errTestDelay) || !errors.Is(e, errTestRelease) {
		t.Fatalf("lost errors: %v", e)
	}
}

func TestOpenContextCancelsQueueLockWait(t *testing.T) {
	b, root := fixture(t)
	held, e := b.lock(context.Background(), "queue.lock", true)
	if e != nil {
		t.Fatal(e)
	}
	defer release(held)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	peer, e := OpenContext(ctx, Config{Root: root, Scope: "scope-a", Instance: "instance-a", Clock: clock.System()})
	if peer != nil {
		peer.Close()
		t.Fatal("open ignored deadline")
	}
	if !errors.Is(e, context.DeadlineExceeded) {
		t.Fatalf("got %v", e)
	}
}

func TestReleaseRepeatsCleanupError(t *testing.T) {
	b, _ := fixture(t)
	l, e := b.Acquire(context.Background(), Critical)
	if e != nil {
		t.Fatal(e)
	}
	b.testWireRelease = func(f *os.File) error { return errors.Join(release(f), errTestRelease) }
	if e = l.Release(); !errors.Is(e, errTestRelease) {
		t.Fatalf("first release %v", e)
	}
	if e = l.Release(); !errors.Is(e, errTestRelease) {
		t.Fatalf("second release lost error: %v", e)
	}
}

func assertFullDeadQueue(t *testing.T, p Priority) {
	t.Helper()
	b, root := fixture(t)
	s := queueState{Next: maxTickets, Tickets: make([]entry, 0, maxTickets)}
	for i := 0; i < maxTickets; i++ {
		id := fmt.Sprintf("%032x", i)
		s.Tickets = append(s.Tickets, entry{ID: id, Seq: uint64(i)})
		if e := os.WriteFile(filepath.Join(root, "live-"+id), nil, 0600); e != nil {
			t.Fatal(e)
		}
	}
	data, e := json.Marshal(s)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(root, "state"), data, 0600); e != nil {
		t.Fatal(e)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	l, e := b.Acquire(ctx, p)
	if e != nil {
		t.Fatal(e)
	}
	if l == nil {
		t.Fatal("missing lease")
	}
	if e = l.Release(); e != nil {
		t.Fatal(e)
	}
}
func TestFullDeadQueueAllowsCriticalRegistration(t *testing.T) { assertFullDeadQueue(t, Critical) }
func TestFullDeadQueueAllowsNormalRegistration(t *testing.T)   { assertFullDeadQueue(t, Normal) }
func TestFullQueueUnknownLivenessFailsClosed(t *testing.T) {
	b, root := fixture(t)
	s := queueState{Next: maxTickets, Tickets: make([]entry, 0, maxTickets)}
	for i := 0; i < maxTickets; i++ {
		id := fmt.Sprintf("%032x", i)
		s.Tickets = append(s.Tickets, entry{ID: id, Seq: uint64(i)})
		if i > 0 {
			if e := os.WriteFile(filepath.Join(root, "live-"+id), nil, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}
	data, _ := json.Marshal(s)
	if e := os.WriteFile(filepath.Join(root, "state"), data, 0600); e != nil {
		t.Fatal(e)
	}
	l, e := b.Acquire(context.Background(), Critical)
	if l != nil {
		l.Release()
		t.Fatal("unknown admitted")
	}
	if !errors.Is(e, ErrUnavailable) {
		t.Fatalf("error %v", e)
	}
	after, e := os.ReadFile(filepath.Join(root, "state"))
	if e != nil || !bytes.Equal(data, after) {
		t.Fatal("uncertain queue changed")
	}
}
