//go:build linux

package queueprototype

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func liveCount(t *testing.T, q *Queue) int {
	t.Helper()
	f, e := filepath.Glob(q.path("live-*"))
	if e != nil {
		t.Fatal(e)
	}
	return len(f)
}
func TestRegisterContextLockCancellation(t *testing.T) {
	q := setup(t)
	held, e := lock(q.path("queue.lock"), syscall.LOCK_EX)
	if e != nil {
		t.Fatal(e)
	}
	defer unlock(held)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		w, e := q.RegisterContext(ctx, true)
		if w != nil {
			w.Close()
		}
		done <- e
	}()
	select {
	case e := <-done:
		if !errors.Is(e, context.DeadlineExceeded) {
			t.Fatalf("error %v", e)
		}
	case <-time.After(time.Second):
		t.Fatal("registration ignored cancellation")
	}
	if n := liveCount(t, q); n != 0 {
		t.Fatalf("unpublished liveness: %d", n)
	}
}
func TestRegisterFailuresPreserveOnlyUncertainLiveness(t *testing.T) {
	for _, mode := range []string{"missing", "dirty", "write", "published"} {
		t.Run(mode, func(t *testing.T) {
			q := setup(t)
			switch mode {
			case "missing":
				if e := os.Remove(q.path("state")); e != nil {
					t.Fatal(e)
				}
			case "dirty":
				if e := os.WriteFile(q.path("dirty"), nil, 0600); e != nil {
					t.Fatal(e)
				}
			case "published":
				q.Write = func(path string, b []byte) error {
					tmp := path + ".tmp"
					if err := os.WriteFile(tmp, b, 0600); err != nil {
						return err
					}
					if err := os.Rename(tmp, path); err != nil {
						return err
					}
					return errors.New("injected post-rename error")
				}
			case "write":
				q.Write = func(string, []byte) error { return errors.New("injected write") }
			}
			w, e := q.RegisterContext(context.Background(), true)
			if w != nil || e == nil {
				t.Fatal("registration should fail")
			}
			want := 0
			if mode == "write" || mode == "published" {
				want = 1
			}
			if n := liveCount(t, q); n != want {
				t.Fatalf("live files %d want %d", n, want)
			}
		})
	}
}
func TestAuxReclaimsDeadButRefusesUnknown(t *testing.T) {
	for _, mode := range []string{"dead", "missing", "dirty"} {
		t.Run(mode, func(t *testing.T) {
			q := setup(t)
			w, e := q.Register(true)
			if e != nil {
				t.Fatal(e)
			}
			id := w.id
			unlock(w.live)
			w.live = nil
			switch mode {
			case "missing":
				if e = os.Remove(q.path("live-" + id)); e != nil {
					t.Fatal(e)
				}
			case "dirty":
				if e = os.WriteFile(q.path("dirty"), nil, 0600); e != nil {
					t.Fatal(e)
				}
			}
			f, e := q.TryAux()
			if mode == "dead" {
				if e != nil || f == nil {
					t.Fatalf("dead ticket: %v", e)
				}
				unlock(f)
				if n := liveCount(t, q); n != 0 {
					t.Fatal("dead file remains")
				}
			} else {
				if f != nil {
					unlock(f)
					t.Fatal("unsafe admission")
				}
				if e == nil {
					t.Fatal("missing error")
				}
			}
		})
	}
}
