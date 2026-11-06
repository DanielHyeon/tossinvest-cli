//go:build linux

package shadowsession

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func TestSimultaneousInitialLock(t *testing.T) {
	if os.Getenv("A169_BARRIER_CHILD") == "1" {
		var dev, ino uint64
		_, _ = fmt.Sscan(os.Getenv("A169_BARRIER_ID"), &dev, &ino)
		in := bufio.NewReader(os.Stdin)
		hooks := &ioHooks{after: func(stage string) error {
			if stage == "lock-open" {
				fmt.Println("OPEN")
				_, err := in.ReadString('\n')
				return err
			}
			return nil
		}}
		s, err := Begin(context.WithValue(context.Background(), hooksKey{}, hooks), Root{Path: os.Getenv("A169_BARRIER_ROOT"), Device: dev, Inode: ino}, "2026-11-06")
		if errors.Is(err, ErrBusy) {
			fmt.Println("BUSY")
			return
		}
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
			return
		}
		fmt.Println("WON")
		_, _ = in.ReadString('\n')
		if err := s.Close(); err != nil {
			fmt.Printf("CLOSE %v\n", err)
			return
		}
		fmt.Println("CLOSED")
		return
	}
	for _, prior := range []string{"absent", "paused"} {
		t.Run(prior, func(t *testing.T) {
			r := fixture(t)
			if prior == "paused" {
				s, err := Begin(context.Background(), r, "2026-11-06")
				if err != nil {
					t.Fatal(err)
				}
				if err := s.Pause(); err != nil {
					t.Fatal(err)
				}
			}
			type child struct {
				cmd *exec.Cmd
				in  io.WriteCloser
				out *bufio.Reader
			}
			children := make([]child, 2)
			for i := range children {
				c := exec.Command(os.Args[0], "-test.run=^TestSimultaneousInitialLock$")
				c.Env = append(os.Environ(), "A169_BARRIER_CHILD=1", "A169_BARRIER_ROOT="+r.Path, fmt.Sprintf("A169_BARRIER_ID=%d %d", r.Device, r.Inode))
				stdin, e := c.StdinPipe()
				if e != nil {
					t.Fatal(e)
				}
				stdout, e := c.StdoutPipe()
				if e != nil {
					t.Fatal(e)
				}
				if e = c.Start(); e != nil {
					t.Fatal(e)
				}
				children[i] = child{c, stdin, bufio.NewReader(stdout)}
				defer func() {
					if c.ProcessState == nil {
						_ = c.Process.Kill()
						_ = c.Wait()
					}
				}()
			}
			read := func(c child) string {
				t.Helper()
				s, e := c.out.ReadString('\n')
				if e != nil {
					t.Fatalf("child output %q: %v", s, e)
				}
				return strings.TrimSpace(s)
			}
			for _, c := range children {
				if s := read(c); s != "OPEN" {
					t.Fatalf("barrier: %q", s)
				}
			}
			for _, c := range children {
				if _, e := io.WriteString(c.in, "go\n"); e != nil {
					t.Fatal(e)
				}
			}
			winner := -1
			busy := 0
			for i, c := range children {
				switch s := read(c); s {
				case "WON":
					winner = i
				case "BUSY":
					busy++
				default:
					t.Fatalf("child result: %q", s)
				}
			}
			if winner < 0 || busy != 1 {
				t.Fatalf("winner=%d busy=%d", winner, busy)
			}
			if _, err := Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrBusy) {
				t.Fatalf("third: %v", err)
			}
			if _, err := io.WriteString(children[winner].in, "close\n"); err != nil {
				t.Fatal(err)
			}
			if s := read(children[winner]); s != "CLOSED" {
				t.Fatalf("close: %q", s)
			}
			for _, c := range children {
				if e := c.cmd.Wait(); e != nil {
					t.Fatal(e)
				}
			}
			lock, err := os.OpenFile(filepath.Join(r.Path, "2026-11-06.lock"), os.O_RDWR, 0)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
				t.Fatalf("lock not free: %v", err)
			}
			if err = unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil {
				t.Fatal(err)
			}
			if _, err = Begin(context.Background(), r, "2026-11-06"); !errors.Is(err, ErrInvalid) || errors.Is(err, ErrBusy) {
				t.Fatalf("active: %v", err)
			}
		})
	}
}

func observed(t *testing.T, r Root) string {
	t.Helper()
	f, e := os.Open(filepath.Join(r.Path, "2026-11-06.state"))
	if errors.Is(e, os.ErrNotExist) {
		return "absent"
	}
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	_, state, _, e := readState(f, "2026-11-06")
	if e != nil {
		return "invalid"
	}
	return state
}
func assertRestart(t *testing.T, r Root, want string) {
	t.Helper()
	if got := observed(t, r); got != want {
		t.Fatalf("record=%s want=%s", got, want)
	}
	s, e := Begin(context.Background(), r, "2026-11-06")
	if want == "absent" || want == "paused" {
		if e != nil || s == nil {
			t.Fatalf("expected restart: %v", e)
		}
		if e = s.Close(); e != nil {
			t.Fatal(e)
		}
		return
	}
	if s != nil || !errors.Is(e, ErrInvalid) || errors.Is(e, ErrBusy) {
		t.Fatalf("expected state refusal: session=%v err=%v", s, e)
	}
}
func TestBeginFaultRestartMatrix(t *testing.T) {
	sentinel := errors.New("fault")
	for _, prior := range []string{"absent", "paused"} {
		for _, stage := range []string{"lock-open", "lock-sync", "lock-dir-sync", "state-open", "write", "state-sync", "state-dir-sync"} {
			for _, when := range []string{"before", "after"} {
				t.Run(prior+"/"+stage+"/"+when, func(t *testing.T) {
					r := fixture(t)
					if prior == "paused" {
						s, e := Begin(context.Background(), r, "2026-11-06")
						if e != nil {
							t.Fatal(e)
						}
						if e = s.Pause(); e != nil {
							t.Fatal(e)
						}
					}
					fail := func(got string) error {
						if got == stage {
							return sentinel
						}
						return nil
					}
					h := &ioHooks{}
					if when == "before" {
						h.before = fail
					} else {
						h.after = fail
					}
					s, e := Begin(context.WithValue(context.Background(), hooksKey{}, h), r, "2026-11-06")
					if s != nil || !errors.Is(e, sentinel) {
						t.Fatalf("session=%v err=%v", s, e)
					}
					want := prior
					if stage == "state-open" && when == "after" && prior == "absent" || stage == "write" && when == "before" && prior == "absent" {
						want = "invalid"
					}
					if stage == "write" && when == "after" || stage == "state-sync" || stage == "state-dir-sync" {
						want = "active"
					}
					assertRestart(t, r, want)
				})
			}
		}
	}
}

func TestTerminalFaultRestartMatrix(t *testing.T) {
	sentinel := errors.New("fault")
	for _, transition := range []string{"paused", "aborted", "completed"} {
		for _, stage := range []string{"write", "state-sync", "state-dir-sync", "unlock", "state-close", "lock-close", "dir-close"} {
			for _, when := range []string{"before", "after"} {
				t.Run(transition+"/"+stage+"/"+when, func(t *testing.T) {
					r := fixture(t)
					s, e := Begin(context.Background(), r, "2026-11-06")
					if e != nil {
						t.Fatal(e)
					}
					fail := func(got string) error {
						if got == stage {
							return sentinel
						}
						return nil
					}
					s.hooks = &ioHooks{}
					if when == "before" {
						s.hooks.before = fail
					} else {
						s.hooks.after = fail
					}
					switch transition {
					case "paused":
						e = s.Pause()
					case "aborted":
						e = s.Abort()
					case "completed":
						e = s.Complete()
					}
					if !errors.Is(e, sentinel) {
						t.Fatalf("error=%v", e)
					}
					want := transition
					if stage == "write" && when == "before" {
						want = "active"
					}
					assertRestart(t, r, want)
				})
			}
		}
	}
}

func TestRuntimeIdentityReplacement(t *testing.T) {
	for _, what := range []string{"root", "lock", "state"} {
		t.Run(what, func(t *testing.T) {
			r := fixture(t)
			s, e := Begin(context.Background(), r, "2026-11-06")
			if e != nil {
				t.Fatal(e)
			}
			switch what {
			case "root":
				old := r.Path + "-old"
				if e = os.Rename(r.Path, old); e != nil {
					t.Fatal(e)
				}
				t.Cleanup(func() { _ = os.RemoveAll(old) })
				if e = os.Mkdir(r.Path, 0700); e != nil {
					t.Fatal(e)
				}
			case "lock", "state":
				name := filepath.Join(r.Path, "2026-11-06."+what)
				if e = os.Rename(name, name+".old"); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(name, []byte("replacement"), 0600); e != nil {
					t.Fatal(e)
				}
			}
			if e = s.Pause(); e == nil {
				t.Fatal("identity replacement accepted")
			}
			if _, e = Begin(context.Background(), r, "2026-11-06"); e == nil {
				t.Fatal("replacement restart accepted")
			}
		})
	}
}

func TestPausedPrefixPartialEOF(t *testing.T) {
	a, _ := marshal("2026-11-06", 1, "active")
	p, _ := marshal("2026-11-06", 2, "paused")
	next, _ := marshal("2026-11-06", 3, "active")
	for _, n := range []int{1, len(next) / 2, len(next) - 1} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			r := fixture(t)
			body := append(append(append([]byte(nil), a...), p...), next[:n]...)
			if e := os.WriteFile(filepath.Join(r.Path, "2026-11-06.state"), body, 0600); e != nil {
				t.Fatal(e)
			}
			assertRestart(t, r, "invalid")
		})
	}
}

func TestNonExt4Refused(t *testing.T) {
	var fs unix.Statfs_t
	if e := unix.Statfs("/dev/shm", &fs); e != nil || uint64(fs.Type) == ext4Magic {
		t.Skipf("non-ext4 /dev/shm unavailable: %v", e)
	}
	dir, e := os.MkdirTemp("/dev/shm", "a169-shadowsession-")
	if e != nil {
		t.Skipf("non-ext4 fixture unavailable: %v", e)
	}
	defer os.RemoveAll(dir)
	var st unix.Stat_t
	if e = unix.Stat(dir, &st); e != nil {
		t.Fatal(e)
	}
	if _, e = Begin(context.Background(), Root{Path: dir, Device: uint64(st.Dev), Inode: st.Ino}, "2026-11-06"); !errors.Is(e, ErrInvalid) {
		t.Fatalf("non-ext4 accepted: %v", e)
	}
}

func TestPauseCommitThenChildKill(t *testing.T) {
	if os.Getenv("A169_PAUSED_CHILD") == "1" {
		var dev, ino uint64
		_, _ = fmt.Sscan(os.Getenv("A169_BARRIER_ID"), &dev, &ino)
		s, err := Begin(context.Background(), Root{Path: os.Getenv("A169_BARRIER_ROOT"), Device: dev, Inode: ino}, "2026-11-06")
		if err != nil {
			fmt.Println("ERROR", err)
			return
		}
		if err = s.Pause(); err != nil {
			fmt.Println("ERROR", err)
			return
		}
		fmt.Println("PAUSED")
		_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		return
	}
	r := fixture(t)
	c := exec.Command(os.Args[0], "-test.run=^TestPauseCommitThenChildKill$")
	c.Env = append(os.Environ(), "A169_PAUSED_CHILD=1", "A169_BARRIER_ROOT="+r.Path, fmt.Sprintf("A169_BARRIER_ID=%d %d", r.Device, r.Inode))
	input, err := c.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	out, err := c.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if c.ProcessState == nil {
			_ = c.Process.Kill()
			_ = c.Wait()
		}
	}()
	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil || strings.TrimSpace(line) != "PAUSED" {
		t.Fatalf("child=%q err=%v", line, err)
	}
	if err = c.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = c.Wait()
	lock, err := os.OpenFile(filepath.Join(r.Path, "2026-11-06.lock"), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		t.Fatalf("lock: %v", err)
	}
	if err = unix.Flock(int(lock.Fd()), unix.LOCK_UN); err != nil {
		t.Fatal(err)
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	if got := observed(t, r); got != "paused" {
		t.Fatalf("state=%s", got)
	}
	s, err := Begin(context.Background(), r, "2026-11-06")
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
}
