package prototype

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func acquireContext(ctx context.Context, ch chan struct{}, f *os.File) (func(), error) {
	select {
	case ch <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	fail := func(err error) (func(), error) { <-ch; return nil, err }
	for f != nil {
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return fail(err)
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fail(ctx.Err())
		case <-timer.C:
		}
	}
	release := func() {
		if f != nil {
			syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		}
		<-ch
	}
	timer := time.NewTimer(350 * time.Millisecond)
	select {
	case <-ctx.Done():
		timer.Stop()
		release()
		return nil, ctx.Err()
	case <-timer.C:
	}
	if err := ctx.Err(); err != nil {
		release()
		return nil, err
	}
	return release, nil
}
func TestAcquireCancellationAtLocalFlockAndCooldown(t *testing.T) {
	for _, stage := range []string{"local", "flock", "cooldown"} {
		t.Run(stage, func(t *testing.T) {
			f, err := os.OpenFile(filepath.Join(t.TempDir(), "lock"), os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			ch := make(chan struct{}, 1)
			var other *os.File
			if stage == "local" {
				ch <- struct{}{}
			}
			if stage == "flock" {
				other, err = os.OpenFile(f.Name(), os.O_RDWR, 0600)
				if err != nil {
					t.Fatal(err)
				}
				defer other.Close()
				if err = syscall.Flock(int(other.Fd()), syscall.LOCK_EX); err != nil {
					t.Fatal(err)
				}
				defer syscall.Flock(int(other.Fd()), syscall.LOCK_UN)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			start := time.Now()
			release, err := acquireContext(ctx, ch, f)
			if release != nil {
				release()
				t.Fatal("unexpected admission")
			}
			if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > time.Second {
				t.Fatalf("cancel err=%v elapsed=%v", err, time.Since(start))
			}
		})
	}
}

func TestCanceledLockWaitDoesNotCallTransport(t *testing.T) {
	f, err := os.OpenFile(filepath.Join(t.TempDir(), "lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	calls := 0
	send := func() error {
		release, err := acquireContext(ctx, ch, f)
		if err != nil {
			return err
		}
		defer release()
		calls++
		return nil
	}
	if err := send(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("transport calls=%d", calls)
	}
}
