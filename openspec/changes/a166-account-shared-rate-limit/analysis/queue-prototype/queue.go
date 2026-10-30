// Package queueprototype explores ordering only; it is not production code.
package queueprototype

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type ticket struct {
	ID       string
	Critical bool
	Seq      uint64
}
type state struct {
	Next    uint64
	Tickets []ticket
}
type Queue struct {
	Dir     string
	Write   func(string, []byte) error
	SyncDir func(string) error
}
type Waiter struct {
	q        *Queue
	id       string
	live     *os.File
	selected bool
}

func lock(path string, flags int) (*os.File, error) {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = syscall.Flock(int(f.Fd()), flags); e != nil {
		f.Close()
		return nil, e
	}
	return f, nil
}
func unlock(f *os.File)               { syscall.Flock(int(f.Fd()), syscall.LOCK_UN); f.Close() }
func (q *Queue) path(s string) string { return filepath.Join(q.Dir, s) }
func (q *Queue) read() (state, error) {
	var s state
	if _, e := os.Lstat(q.path("dirty")); e == nil {
		return s, errors.New("dirty queue")
	} else if !errors.Is(e, os.ErrNotExist) {
		return s, e
	}
	b, e := os.ReadFile(q.path("state"))
	if e != nil {
		return s, e
	}
	e = json.Unmarshal(b, &s)
	if e != nil {
		return s, e
	}
	seen := map[string]bool{}
	for _, t := range s.Tickets {
		if t.ID == "" || t.Seq >= s.Next || seen[t.ID] {
			return s, errors.New("bad state")
		}
		seen[t.ID] = true
	}
	return s, nil
}
func (q *Queue) syncDir() error {
	if q.SyncDir != nil {
		return q.SyncDir(q.Dir)
	}
	d, e := os.Open(q.Dir)
	if e != nil {
		return e
	}
	e = d.Sync()
	if c := d.Close(); e == nil {
		e = c
	}
	return e
}

// 큐 잠금 아래 dirty 선기록으로 불확실한 갱신을 다른 참여자에게 드러냄.
func (q *Queue) write(s state) error {
	b, e := json.Marshal(s)
	if e != nil {
		return e
	}
	dirty, e := os.OpenFile(q.path("dirty"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	e = dirty.Sync()
	if c := dirty.Close(); e == nil {
		e = c
	}
	if e != nil {
		return e
	}
	if e = q.syncDir(); e != nil {
		return e
	}
	if q.Write != nil {
		e = q.Write(q.path("state"), b)
	} else {
		var f *os.File
		f, e = os.CreateTemp(q.Dir, "state-")
		if e != nil {
			return e
		}
		defer os.Remove(f.Name())
		if _, e = f.Write(b); e == nil {
			e = f.Sync()
		}
		if c := f.Close(); e == nil {
			e = c
		}
		if e == nil {
			e = os.Rename(f.Name(), q.path("state"))
		}
	}
	if e != nil {
		return e
	}
	if e = q.syncDir(); e != nil {
		return e
	}
	if e = os.Remove(q.path("dirty")); e != nil {
		return e
	}
	return q.syncDir()
}
func (q *Queue) Register(critical bool) (*Waiter, error) {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return nil, e
	}
	id := hex.EncodeToString(b)
	f, e := os.OpenFile(q.path("live-"+id), os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if e != nil {
		return nil, e
	}
	if e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); e != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, e
	}
	w := &Waiter{q: q, id: id, live: f}
	l, e := lock(q.path("queue.lock"), syscall.LOCK_EX)
	if e != nil {
		unlock(w.live)
		os.Remove(w.live.Name())
		return nil, e
	}
	defer unlock(l)
	s, e := q.read()
	if e == nil {
		s.Tickets = append(s.Tickets, ticket{id, critical, s.Next})
		s.Next++
		e = q.write(s)
	}
	if e != nil {
		unlock(w.live)
		return nil, e
	}
	return w, nil
}
func (w *Waiter) Close() {
	if w.live == nil {
		return
	}
	if w.selected {
		os.Remove(w.live.Name())
		unlock(w.live)
		w.live = nil
		return
	}
	// 큐 기록이 확정적으로 제거된 경우에만 liveness 파일을 지움.
	l, e := lock(w.q.path("queue.lock"), syscall.LOCK_EX)
	if e == nil {
		s, err := w.q.read()
		if err == nil {
			found := false
			for i, t := range s.Tickets {
				if t.ID == w.id {
					s.Tickets = append(s.Tickets[:i], s.Tickets[i+1:]...)
					found = true
					break
				}
			}
			if found {
				err = w.q.write(s)
			}
			if err == nil {
				os.Remove(w.live.Name())
			}
		}
		unlock(l)
	}
	unlock(w.live)
	w.live = nil
}
func (w *Waiter) Try() (*os.File, error) {
	return w.tryContext(context.Background())
}

func (w *Waiter) tryContext(ctx context.Context) (*os.File, error) {
	q := w.q
	var l *os.File
	var e error
	for {
		if e = ctx.Err(); e != nil {
			return nil, e
		}
		l, e = lock(q.path("queue.lock"), syscall.LOCK_EX|syscall.LOCK_NB)
		if !errors.Is(e, syscall.EWOULDBLOCK) {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
	}
	if e != nil {
		return nil, e
	}
	defer unlock(l)
	s, e := q.read()
	if e != nil {
		return nil, e
	}
	var dead []string
	live := s.Tickets[:0]
	found := false
	for _, t := range s.Tickets {
		if t.ID == w.id {
			found = true
			live = append(live, t)
			continue
		}
		f, e := os.OpenFile(q.path("live-"+t.ID), os.O_RDWR, 0)
		if e != nil {
			return nil, e
		}
		e = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if e == nil {
			unlock(f)
			dead = append(dead, t.ID)
			continue
		}
		f.Close()
		if !errors.Is(e, syscall.EWOULDBLOCK) {
			return nil, e
		}
		live = append(live, t)
	}
	s.Tickets = live
	if !found {
		return nil, errors.New("ticket missing")
	}
	selected := s.Tickets[0]
	for _, t := range s.Tickets[1:] {
		if (t.Critical && !selected.Critical) || (t.Critical == selected.Critical && t.Seq < selected.Seq) {
			selected = t
		}
	}
	if selected.ID != w.id {
		if e = q.writeClean(s, dead); e != nil {
			return nil, e
		}
		return nil, nil
	}
	wire, e := lock(q.path("wire.lock"), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(e, syscall.EWOULDBLOCK) {
		if e = q.writeClean(s, dead); e != nil {
			return nil, e
		}
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	for i, t := range s.Tickets {
		if t.ID == w.id {
			s.Tickets = append(s.Tickets[:i], s.Tickets[i+1:]...)
			break
		}
	}
	if e = q.writeClean(s, dead); e != nil {
		unlock(wire)
		return nil, e
	}
	w.selected = true
	return wire, nil
}
func (q *Queue) writeClean(s state, dead []string) error {
	if e := q.write(s); e != nil {
		return e
	}
	for _, id := range dead {
		_ = os.Remove(q.path("live-" + id))
	}
	return nil
}
func (w *Waiter) Wait(ctx context.Context) (*os.File, error) {
	for {
		f, e := w.tryContext(ctx)
		if e != nil {
			return nil, e
		}
		if f != nil {
			if e = ctx.Err(); e != nil {
				unlock(f)
				return nil, e
			}
			return f, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Millisecond * 5):
		}
	}
}

// TryAux does not publish a ticket and never waits for wire ownership.
func (q *Queue) TryAux() (*os.File, error) {
	l, e := lock(q.path("queue.lock"), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(e, syscall.EWOULDBLOCK) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	defer unlock(l)
	s, e := q.read()
	if e != nil {
		return nil, e
	}
	if len(s.Tickets) != 0 {
		return nil, nil
	}
	f, e := lock(q.path("wire.lock"), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(e, syscall.EWOULDBLOCK) {
		return nil, nil
	}
	return f, e
}
