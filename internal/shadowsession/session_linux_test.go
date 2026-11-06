//go:build linux

package shadowsession

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func fixture(t *testing.T) Root {
	t.Helper()
	parent, specified := os.LookupEnv("A169_SHADOWSESSION_TEST_ROOT")
	if !specified {
		parent = os.TempDir()
	}
	if specified && (!filepath.IsAbs(parent) || filepath.Clean(parent) != parent) {
		t.Fatalf("A169_SHADOWSESSION_TEST_ROOT must be an absolute, clean path: %q", parent)
	}
	var fs unix.Statfs_t
	if err := unix.Statfs(parent, &fs); err != nil || uint64(fs.Type) != ext4Magic {
		if specified {
			t.Fatalf("invalid A169_SHADOWSESSION_TEST_ROOT %q: filesystem=%x err=%v", parent, fs.Type, err)
		}
		t.Skipf("default temp directory is unavailable or not ext4: filesystem=%x err=%v", fs.Type, err)
	}
	dir, err := os.MkdirTemp(parent, "shadowsession-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		t.Fatal(err)
	}
	return Root{Path: dir, Device: uint64(st.Dev), Inode: st.Ino}
}

func TestLifecycle(t *testing.T) {
	r := fixture(t)
	s, err := Begin(context.Background(), r, "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrBusy) {
		t.Fatalf("busy: %v", err)
	}
	if err = s.Pause(); err != nil {
		t.Fatal(err)
	}
	s, err = Begin(context.Background(), r, "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Complete(); err != nil {
		t.Fatal(err)
	}
	if _, err = Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("completed reopened")
	}
}

func TestActiveAndInputsFailClosed(t *testing.T) {
	r := fixture(t)
	s, err := Begin(context.Background(), r, "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("active reopened")
	}
	for _, day := range []string{"2026-02-29", "0000-01-01", "../x", "2026-1-01"} {
		if _, err = Begin(context.Background(), r, day); err == nil {
			t.Fatalf("accepted %q", day)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = Begin(ctx, r, "2026-11-07"); err == nil {
		t.Fatal("cancelled")
	}
	if _, err = Begin(context.Background(), Root{Path: filepath.Clean(r.Path)}, "2026-11-07"); err == nil {
		t.Fatal("zero identity")
	}
	var nilSession *Session
	if err := nilSession.Pause(); !errors.Is(err, ErrClosed) {
		t.Fatalf("nil: %v", err)
	}
	var zero Session
	if err := zero.Close(); !errors.Is(err, ErrClosed) {
		t.Fatalf("zero: %v", err)
	}
}

func TestCorruptAndIdentity(t *testing.T) {
	for _, body := range []string{"", "{}\n", "{\"version\":1}\n", "partial", strings.Repeat("x", maxBytes+1)} {
		t.Run(fmt.Sprint(len(body), "/", body[:min(len(body), 3)]), func(t *testing.T) {
			r := fixture(t)
			p := filepath.Join(r.Path, "2026-11-06.state")
			if err := os.WriteFile(p, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
				t.Fatal("corrupt accepted")
			}
		})
	}
	r := fixture(t)
	r.Inode++
	if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("wrong inode")
	}
	r = fixture(t)
	if err := os.Symlink("/dev/null", filepath.Join(r.Path, "2026-11-06.lock")); err != nil {
		t.Fatal(err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("symlink")
	}
}

func TestPostCommitAndShortWrite(t *testing.T) {
	injected := errors.New("injected")
	for _, stage := range []string{"state-sync", "state-dir-sync", "unlock", "state-close", "lock-close", "dir-close"} {
		t.Run(stage, func(t *testing.T) {
			r := fixture(t)
			s, err := Begin(context.Background(), r, "2026-11-06")
			if err != nil {
				t.Fatal(err)
			}
			s.hooks = &ioHooks{after: func(got string) error {
				if got == stage {
					return injected
				}
				return nil
			}}
			if err := s.Pause(); !errors.Is(err, injected) {
				t.Fatalf("pause: %v", err)
			}
			resumed, err := Begin(context.Background(), r, "2026-11-06")
			if err != nil {
				t.Fatalf("resume after committed pause: %v", err)
			}
			if err := resumed.Abort(); err != nil {
				t.Fatal(err)
			}
			if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
				t.Fatal("aborted resumed")
			}
		})
	}
	r := fixture(t)
	h := &ioHooks{write: func(f *os.File, p []byte) (int, error) { return f.Write(p[:len(p)/2]) }}
	if _, err := Begin(context.WithValue(context.Background(), hooksKey{}, h), r, "2026-11-06"); !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("short: %v", err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("partial reopened")
	}
}

func TestPrecommitFailuresAndCancellation(t *testing.T) {
	injected := errors.New("injected")
	for _, stage := range []string{"lock-open", "lock-sync", "lock-dir-sync", "state-open", "write", "state-sync", "state-dir-sync"} {
		t.Run(stage, func(t *testing.T) {
			r := fixture(t)
			h := &ioHooks{before: func(got string) error {
				if got == stage {
					return injected
				}
				return nil
			}}
			s, err := Begin(context.WithValue(context.Background(), hooksKey{}, h), r, "2026-11-06")
			if s != nil || !errors.Is(err, injected) {
				t.Fatalf("s=%v err=%v", s, err)
			}
		})
	}
	r := fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	h := &ioHooks{after: func(got string) error {
		if got == "state-dir-sync" {
			cancel()
		}
		return nil
	}}
	if s, err := Begin(context.WithValue(ctx, hooksKey{}, h), r, "2026-11-06"); s != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("s=%v err=%v", s, err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("cancelled active reopened")
	}
}

func TestChildLockAndCrash(t *testing.T) {
	if os.Getenv("SHADOWSESSION_CHILD") == "1" {
		var dev, ino uint64
		_, _ = fmt.Sscan(os.Getenv("SHADOWSESSION_ID"), &dev, &ino)
		s, err := Begin(context.Background(), Root{Path: os.Getenv("SHADOWSESSION_ROOT"), Device: dev, Inode: ino}, "2026-11-06")
		if err != nil {
			os.Exit(3)
		}
		fmt.Println("READY")
		os.Stdout.Sync()
		_ = s
		select {}
	}
	r := fixture(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestChildLockAndCrash$")
	cmd.Env = append(os.Environ(), "SHADOWSESSION_CHILD=1", "SHADOWSESSION_ROOT="+r.Path, fmt.Sprintf("SHADOWSESSION_ID=%d %d", r.Device, r.Inode))
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil || strings.TrimSpace(line) != "READY" {
		t.Fatalf("child %q %v", line, err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrBusy) {
		t.Fatalf("lock: %v", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = cmd.Wait()
	lock, err := os.OpenFile(filepath.Join(r.Path, "2026-11-06.lock"), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Fatalf("lock not released: %v", err)
	}
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil {
		t.Fatal(err)
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(filepath.Join(r.Path, "2026-11-06.state"))
	if err != nil {
		t.Fatal(err)
	}
	_, state, _, err := readState(f, "2026-11-06")
	_ = f.Close()
	if err != nil || state != "active" {
		t.Fatalf("state=%q err=%v", state, err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrInvalid) || errors.Is(err, ErrBusy) {
		t.Fatalf("crashed active: %v", err)
	}
}

func TestCanonicalAndTransitions(t *testing.T) {
	good, err := marshal("2026-11-06", 1, "active")
	if err != nil {
		t.Fatal(err)
	}
	paused, _ := marshal("2026-11-06", 2, "paused")
	invalidNext, _ := marshal("2026-11-06", 3, "completed")
	cases := [][]byte{
		append(append([]byte(nil), good...), bytes.Replace(paused, []byte(`"checksum":"`), []byte(`"checksum":"0`), 1)...),
		bytes.Replace(good, []byte(`"checksum":"`), []byte(`"checksum":"0`), 1),
		bytes.Replace(good, []byte(`"version":1`), []byte(`"version":1,"version":1`), 1),
		bytes.Replace(good, []byte(`"version":1`), []byte(`"version":1,"extra":1`), 1),
		bytes.Replace(good, []byte(`"active"`), []byte(`"paused"`), 1),
		append(append(append([]byte(nil), good...), paused...), invalidNext...),
		bytes.Replace(good, []byte(`"sequence":1`), []byte(`"sequence":18446744073709551616`), 1),
	}
	for i, data := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			r := fixture(t)
			if err := os.WriteFile(filepath.Join(r.Path, "2026-11-06.state"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
				t.Fatal("invalid accepted")
			}
		})
	}
}

func TestRootAndFileGuards(t *testing.T) {
	r := fixture(t)
	target := filepath.Join(r.Path, "target")
	if err := os.WriteFile(target, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(target, filepath.Join(r.Path, "2026-11-06.lock")); err != nil {
		t.Fatal(err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("hardlink accepted")
	}
	r = fixture(t)
	if err := os.Chmod(r.Path, 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := Begin(context.Background(), r, "2026-11-06"); err == nil {
		t.Fatal("permissions accepted")
	}
	r = fixture(t)
	s, err := Begin(context.Background(), r, "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(filepath.Join(r.Path, "2026-11-06.state"), filepath.Join(r.Path, "old")); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(r.Path, "2026-11-06.state"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = s.Pause(); err == nil {
		t.Fatal("replacement accepted")
	}
}

func TestConcurrentFinish(t *testing.T) {
	r := fixture(t)
	s, err := Begin(context.Background(), r, "2026-11-06")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- s.Pause() }()
	}
	wg.Wait()
	close(results)
	var ok, closed int
	for e := range results {
		if e == nil {
			ok++
		} else if errors.Is(e, ErrClosed) {
			closed++
		} else {
			t.Fatal(e)
		}
	}
	if ok != 1 || closed != 1 {
		t.Fatalf("success=%d closed=%d", ok, closed)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
}
