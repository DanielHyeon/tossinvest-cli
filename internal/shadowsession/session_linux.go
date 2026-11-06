//go:build linux

package shadowsession

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const maxBytes = 128 << 10
const ext4Magic = 0xef53

type frameBody struct {
	Version  int    `json:"version"`
	Day      string `json:"day"`
	Sequence uint64 `json:"sequence"`
	State    string `json:"state"`
}
type frame struct {
	Version  int    `json:"version"`
	Day      string `json:"day"`
	Sequence uint64 `json:"sequence"`
	State    string `json:"state"`
	Checksum string `json:"checksum"`
}

// Session owns a lock and must not be copied. Pause and Complete require the caller
// to stop admission and drain all work; no wire may follow a terminal method.
type Session struct {
	mu               sync.Mutex
	root             Root
	dir, lock, state *os.File
	day              string
	sequence         uint64
	size             int64
	closed           bool
	closeErr         error
	hooks            *ioHooks
}

// ioHooks are private, per-call fault seams; nil uses the operating system.
type ioHooks struct {
	before func(string) error
	after  func(string) error
	write  func(*os.File, []byte) (int, error)
}
type hooksKey struct{}

func (s *Session) operation(name string, fn func() error) error {
	if s.hooks != nil && s.hooks.before != nil {
		if err := s.hooks.before(name); err != nil {
			return err
		}
	}
	err := fn()
	if err == nil && s.hooks != nil && s.hooks.after != nil {
		err = s.hooks.after(name)
	}
	return err
}

func validDay(day string) bool {
	if len(day) != 10 || day[:4] == "0000" {
		return false
	}
	t, err := time.Parse("2006-01-02", day)
	return err == nil && t.Format("2006-01-02") == day
}

func openRoot(root Root) (*os.File, error) {
	if root.Device == 0 || root.Inode == 0 || !filepath.IsAbs(root.Path) || root.Path == "/" || filepath.Clean(root.Path) != root.Path {
		return nil, ErrInvalid
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(root.Path, "/"), "/") {
		next, e := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		_ = unix.Close(fd)
		if e != nil {
			return nil, e
		}
		fd = next
	}
	var st unix.Stat_t
	var fs unix.Statfs_t
	if err = unix.Fstat(fd, &st); err == nil {
		err = unix.Fstatfs(fd, &fs)
	}
	if err != nil || uint64(st.Dev) != root.Device || st.Ino != root.Inode || st.Uid != uint32(os.Geteuid()) || st.Mode&07777 != 0700 || uint64(fs.Type) != ext4Magic {
		_ = unix.Close(fd)
		return nil, errors.Join(ErrInvalid, err)
	}
	return os.NewFile(uintptr(fd), root.Path), nil
}

func checkFile(dir *os.File, name string, f *os.File) error {
	var fd, named unix.Stat_t
	if f == nil {
		return ErrInvalid
	}
	if err := unix.Fstat(int(f.Fd()), &fd); err != nil {
		return err
	}
	if err := unix.Fstatat(int(dir.Fd()), name, &named, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	if fd.Dev != named.Dev || fd.Ino != named.Ino || fd.Mode&unix.S_IFMT != unix.S_IFREG || fd.Mode&07777 != 0600 || fd.Uid != uint32(os.Geteuid()) || fd.Nlink != 1 {
		return ErrInvalid
	}
	return nil
}

func openFile(dir *os.File, name string, flags int) (*os.File, bool, error) {
	base := flags | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
	fd, err := unix.Openat(int(dir.Fd()), name, base|unix.O_CREAT|unix.O_EXCL, 0600)
	created := err == nil
	if errors.Is(err, unix.EEXIST) {
		fd, err = unix.Openat(int(dir.Fd()), name, base, 0)
	}
	if err != nil {
		return nil, false, err
	}
	f := os.NewFile(uintptr(fd), name)
	if err = checkFile(dir, name, f); err != nil {
		_ = f.Close()
		return nil, false, err
	}
	return f, created, nil
}

func (s *Session) verify() error {
	root, err := openRoot(s.root)
	if err != nil {
		return err
	}
	err = root.Close()
	if err == nil {
		err = checkFile(s.dir, s.day+".lock", s.lock)
	}
	if err == nil && s.state != nil {
		err = checkFile(s.dir, s.day+".state", s.state)
	}
	return err
}

func marshal(day string, seq uint64, state string) ([]byte, error) {
	body := frameBody{1, day, seq, state}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	b, err = json.Marshal(frame{1, day, seq, state, hex.EncodeToString(sum[:])})
	return append(b, '\n'), err
}

func readState(f *os.File, day string) (uint64, string, int64, error) {
	info, err := f.Stat()
	if err != nil {
		return 0, "", 0, err
	}
	if info.Size() <= 0 || info.Size() > maxBytes {
		return 0, "", 0, ErrInvalid
	}
	b, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return 0, "", 0, err
	}
	if int64(len(b)) != info.Size() || b[len(b)-1] != '\n' {
		return 0, "", 0, ErrInvalid
	}
	var seq uint64
	state := ""
	for _, line := range bytes.Split(b[:len(b)-1], []byte{'\n'}) {
		var x frame
		if json.Unmarshal(line, &x) != nil || x.Version != 1 || x.Day != day || seq == ^uint64(0) || x.Sequence != seq+1 {
			return 0, "", 0, ErrInvalid
		}
		if !(state == "" && x.State == "active" || state == "paused" && x.State == "active" || state == "active" && (x.State == "paused" || x.State == "aborted" || x.State == "completed")) {
			return 0, "", 0, ErrInvalid
		}
		canonical, e := marshal(day, x.Sequence, x.State)
		if e != nil || !bytes.Equal(canonical, append(append([]byte(nil), line...), '\n')) {
			return 0, "", 0, ErrInvalid
		}
		seq, state = x.Sequence, x.State
	}
	return seq, state, info.Size(), nil
}

func (s *Session) appendState(state string) error {
	if err := s.verify(); err != nil {
		return err
	}
	if s.sequence == ^uint64(0) {
		return ErrInvalid
	}
	info, err := s.state.Stat()
	if err != nil || info.Size() != s.size {
		return errors.Join(ErrInvalid, err)
	}
	b, err := marshal(s.day, s.sequence+1, state)
	if err != nil {
		return err
	}
	if s.size+int64(len(b)) > maxBytes {
		return ErrInvalid
	}
	err = s.operation("write", func() error {
		var n int
		var e error
		if s.hooks != nil && s.hooks.write != nil {
			n, e = s.hooks.write(s.state, b)
		} else {
			n, e = s.state.Write(b)
		}
		if e == nil && n != len(b) {
			return io.ErrShortWrite
		}
		return e
	})
	if err == nil {
		err = s.operation("state-sync", s.state.Sync)
	}
	if err == nil {
		err = s.operation("state-dir-sync", s.dir.Sync)
	}
	if err != nil {
		return err
	}
	s.sequence++
	s.size += int64(len(b))
	if err = s.verify(); err != nil {
		return err
	}
	info, err = s.state.Stat()
	if err != nil || info.Size() != s.size {
		return errors.Join(ErrInvalid, err)
	}
	return nil
}

func Begin(ctx context.Context, root Root, day string) (*Session, error) {
	if ctx == nil || !validDay(day) {
		return nil, ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir, err := openRoot(root)
	if err != nil {
		return nil, err
	}
	hooks, _ := ctx.Value(hooksKey{}).(*ioHooks)
	s := &Session{root: root, dir: dir, day: day, hooks: hooks}
	fail := func(e error) (*Session, error) { return nil, errors.Join(e, s.cleanup()) }
	err = s.operation("lock-open", func() error { var e error; s.lock, _, e = openFile(dir, day+".lock", unix.O_RDWR); return e })
	if err != nil {
		return fail(err)
	}
	if err = unix.Flock(int(s.lock.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
			err = ErrBusy
		}
		return fail(err)
	}
	if err = s.verify(); err == nil {
		err = s.operation("lock-sync", s.lock.Sync)
	}
	if err == nil {
		err = s.operation("lock-dir-sync", dir.Sync)
	}
	if err != nil {
		return fail(err)
	}
	var created bool
	err = s.operation("state-open", func() error {
		var e error
		s.state, created, e = openFile(dir, day+".state", unix.O_RDWR|unix.O_APPEND)
		return e
	})
	if err != nil {
		return fail(err)
	}
	if !created {
		var state string
		s.sequence, state, s.size, err = readState(s.state, day)
		if err != nil {
			return fail(err)
		}
		if state != "paused" {
			return fail(ErrInvalid)
		}
	}
	if err = ctx.Err(); err == nil {
		err = s.appendState("active")
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return fail(err)
	}
	return s, nil
}

func (s *Session) cleanup() error {
	var errs []error
	if s.state != nil {
		// Always perform actual descriptor cleanup before injecting an error.
		var before error
		if s.hooks != nil && s.hooks.before != nil {
			before = s.hooks.before("state-close")
		}
		err := s.state.Close()
		if err == nil && s.hooks != nil && s.hooks.after != nil {
			err = s.hooks.after("state-close")
		}
		errs = append(errs, before, err)
		s.state = nil
	}
	if s.lock != nil {
		var beforeUnlock, beforeClose error
		if s.hooks != nil && s.hooks.before != nil {
			beforeUnlock = s.hooks.before("unlock")
			beforeClose = s.hooks.before("lock-close")
		}
		err := unix.Flock(int(s.lock.Fd()), unix.LOCK_UN)
		if err == nil && s.hooks != nil && s.hooks.after != nil {
			err = s.hooks.after("unlock")
		}
		closeErr := s.lock.Close()
		if closeErr == nil && s.hooks != nil && s.hooks.after != nil {
			closeErr = s.hooks.after("lock-close")
		}
		errs = append(errs, beforeUnlock, err, beforeClose, closeErr)
		s.lock = nil
	}
	if s.dir != nil {
		var before error
		if s.hooks != nil && s.hooks.before != nil {
			before = s.hooks.before("dir-close")
		}
		err := s.dir.Close()
		if err == nil && s.hooks != nil && s.hooks.after != nil {
			err = s.hooks.after("dir-close")
		}
		errs = append(errs, before, err)
		s.dir = nil
	}
	return errors.Join(errs...)
}

func (s *Session) finish(state string) error {
	if s == nil {
		return ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.dir == nil || s.lock == nil || s.state == nil {
		return ErrClosed
	}
	err := s.appendState(state)
	s.closed = true
	s.closeErr = s.cleanup()
	return errors.Join(err, s.closeErr)
}
func (s *Session) Pause() error    { return s.finish("paused") }
func (s *Session) Abort() error    { return s.finish("aborted") }
func (s *Session) Complete() error { return s.finish("completed") }
func (s *Session) Close() error {
	if s == nil {
		return ErrClosed
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return s.closeErr
	}
	if s.dir == nil || s.lock == nil || s.state == nil {
		return ErrClosed
	}
	s.closed = true
	s.closeErr = s.cleanup()
	return s.closeErr
}
