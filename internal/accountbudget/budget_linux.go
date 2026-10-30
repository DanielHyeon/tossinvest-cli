//go:build linux

// Package accountbudget contains an unwired Linux-only shared admission candidate.
package accountbudget

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"golang.org/x/sys/unix"
)

var ErrUnavailable = errors.New("account budget unavailable")
var ErrNotSent = errors.New("account budget: request not sent")

const interval = 350 * time.Millisecond
const maxTickets = 256

type Priority uint8

const (
	Normal Priority = iota
	Critical
	Auxiliary
)

type Config struct {
	Root     string
	Scope    string
	Instance string
	Clock    clock.Clock
}
type Budget struct {
	observation     *ObservationStore
	root            string
	scope           string
	instance        string
	clk             clock.Clock
	dir             int
	dirStat         unix.Stat_t
	locks           map[string]unix.Stat_t
	mu              sync.Mutex
	closed          bool
	active          int
	testSyncDir     func(int) error
	testWireRelease func(*os.File) error
}
type entry struct {
	ID       string `json:"id"`
	Seq      uint64 `json:"seq"`
	Critical bool   `json:"critical"`
}
type queueState struct {
	Next    uint64  `json:"next"`
	Tickets []entry `json:"tickets"`
}

// Provision는 빈 전용 디렉터리만 초기화함. 운영 설정에는 자동 호출하지 않음.
func Provision(root, scope, instance string) error {
	if !validScope(scope) || !validScope(instance) {
		return fmt.Errorf("%w: invalid scope", ErrUnavailable)
	}
	d, e := checkedDir(root)
	if e != nil {
		return errors.Join(ErrUnavailable, e)
	}
	defer unix.Close(d)
	for _, item := range []struct {
		name string
		data []byte
	}{{"scope", []byte(scope)}, {"instance", []byte(instance)}, {"queue.lock", nil}, {"wire.lock", nil}, {"state", []byte(`{"next":0,"tickets":[]}`)}} {
		f, _, err := checkedFile(d, item.name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL)
		if err != nil {
			return errors.Join(ErrUnavailable, err)
		}
		if len(item.data) > 0 {
			_, err = f.Write(item.data)
		}
		if err == nil {
			err = f.Sync()
		}
		err = errors.Join(err, f.Close())
		if err != nil {
			return errors.Join(ErrUnavailable, err)
		}
	}
	if e = unix.Fsync(d); e != nil {
		return errors.Join(ErrUnavailable, e)
	}
	return nil
}
func Open(c Config) (*Budget, error) { return OpenContext(context.Background(), c) }

func OpenContext(ctx context.Context, c Config) (*Budget, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if !validScope(c.Scope) || !validScope(c.Instance) || c.Clock == nil {
		return nil, fmt.Errorf("%w: invalid config", ErrUnavailable)
	}
	d, e := checkedDir(c.Root)
	if e != nil {
		return nil, errors.Join(ErrUnavailable, e)
	}
	b := &Budget{root: c.Root, scope: c.Scope, instance: c.Instance, clk: c.Clock, dir: d, locks: make(map[string]unix.Stat_t)}
	fail := func(err error) (*Budget, error) { unix.Close(d); return nil, errors.Join(ErrUnavailable, err) }
	if e = unix.Fstat(d, &b.dirStat); e != nil {
		return fail(e)
	}
	f, _, e := checkedFile(d, "scope", unix.O_RDONLY)
	if e != nil {
		return fail(e)
	}
	v, e := readBounded(f, 128)
	e = errors.Join(e, f.Close())
	if e != nil || string(v) != c.Scope {
		return fail(errors.Join(errors.New("scope mismatch"), e))
	}
	f, _, e = checkedFile(d, "instance", unix.O_RDONLY)
	if e != nil {
		return fail(e)
	}
	v, e = readBounded(f, 128)
	e = errors.Join(e, f.Close())
	if e != nil || string(v) != c.Instance {
		return fail(errors.Join(errors.New("instance mismatch"), e))
	}
	for _, name := range []string{"queue.lock", "wire.lock"} {
		f, st, err := checkedFile(d, name, unix.O_RDWR)
		if err != nil {
			return fail(err)
		}
		if e = f.Close(); e != nil {
			return fail(e)
		}
		b.locks[name] = st
	}
	l, e := b.lock(ctx, "queue.lock", true)
	if e != nil {
		return fail(e)
	}
	_, e = b.read()
	e = errors.Join(e, release(l))
	if e != nil {
		return fail(e)
	}
	return b, nil
}
func (b *Budget) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	if b.active != 0 {
		return errors.New("budget has active owners")
	}
	b.closed = true
	return unix.Close(b.dir)
}
func (b *Budget) verify() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return errors.New("closed")
	}
	d, e := checkedDir(b.root)
	if e != nil {
		return e
	}
	defer unix.Close(d)
	var st unix.Stat_t
	if e = unix.Fstat(d, &st); e != nil {
		return e
	}
	if st.Dev != b.dirStat.Dev || st.Ino != b.dirStat.Ino {
		return errors.New("root replaced")
	}
	f, _, e := checkedFile(b.dir, "scope", unix.O_RDONLY)
	if e != nil {
		return e
	}
	v, e := readBounded(f, 128)
	e = errors.Join(e, f.Close())
	if e != nil {
		return e
	}
	if string(v) != b.scope {
		return errors.New("scope changed")
	}
	f, _, e = checkedFile(b.dir, "instance", unix.O_RDONLY)
	if e != nil {
		return e
	}
	v, e = readBounded(f, 128)
	e = errors.Join(e, f.Close())
	if e != nil {
		return e
	}
	if string(v) != b.instance {
		return errors.New("instance changed")
	}
	for n, stat := range b.locks {
		if e = checkIdentity(b.dir, n, stat); e != nil {
			return e
		}
	}
	return nil
}
func (b *Budget) lock(ctx context.Context, name string, wait bool) (*os.File, error) {
	if e := b.verify(); e != nil {
		return nil, errors.Join(ErrUnavailable, e)
	}
	f, st, e := checkedFile(b.dir, name, unix.O_RDWR)
	if e != nil {
		return nil, errors.Join(ErrUnavailable, e)
	}
	expected := b.locks[name]
	if st.Dev != expected.Dev || st.Ino != expected.Ino {
		f.Close()
		return nil, ErrUnavailable
	}
	for {
		if e = ctx.Err(); e != nil {
			f.Close()
			return nil, e
		}
		e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if e == nil {
			if e = ctx.Err(); e != nil {
				release(f)
				return nil, e
			}
			return f, nil
		}
		if !errors.Is(e, unix.EWOULDBLOCK) && !errors.Is(e, unix.EAGAIN) {
			f.Close()
			return nil, errors.Join(ErrUnavailable, e)
		}
		if !wait {
			f.Close()
			return nil, nil
		}
		if e = b.clk.Sleep(ctx, 5*time.Millisecond); e != nil {
			f.Close()
			return nil, e
		}
	}
}
func release(f *os.File) error {
	if f == nil {
		return nil
	}
	return errors.Join(unix.Flock(int(f.Fd()), unix.LOCK_UN), f.Close())
}
func (b *Budget) read() (queueState, error) {
	var s queueState
	f, _, e := checkedFile(b.dir, "dirty", unix.O_RDONLY)
	if e == nil {
		f.Close()
		return s, errors.New("dirty state")
	}
	if !errors.Is(e, unix.ENOENT) {
		return s, e
	}
	f, _, e = checkedFile(b.dir, "state", unix.O_RDONLY)
	if e != nil {
		return s, e
	}
	data, e := readBounded(f, maxStateBytes)
	e = errors.Join(e, f.Close())
	if e != nil {
		return s, e
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if e = dec.Decode(&s); e != nil {
		return s, e
	}
	if dec.Decode(new(any)) != io.EOF {
		return s, errors.New("trailing state")
	}
	canonical, e := json.Marshal(s)
	if e != nil || !bytes.Equal(canonical, data) || s.Tickets == nil || len(s.Tickets) > maxTickets {
		return s, errors.New("invalid state")
	}
	ids := map[string]bool{}
	seq := map[uint64]bool{}
	for _, t := range s.Tickets {
		raw, err := hex.DecodeString(t.ID)
		if err != nil || len(raw) != 16 || hex.EncodeToString(raw) != t.ID || t.Seq >= s.Next || ids[t.ID] || seq[t.Seq] {
			return s, errors.New("invalid ticket")
		}
		ids[t.ID] = true
		seq[t.Seq] = true
	}
	return s, nil
}
func (b *Budget) syncDir() error {
	if b.testSyncDir != nil {
		return b.testSyncDir(b.dir)
	}
	return unix.Fsync(b.dir)
}
func (b *Budget) write(s queueState) (result error) {
	data, e := json.Marshal(s)
	if e != nil || len(data) > maxStateBytes {
		return errors.New("invalid state size")
	}
	marker, _, e := checkedFile(b.dir, "dirty", unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL)
	if e != nil {
		return e
	}
	e = errors.Join(marker.Sync(), marker.Close())
	if e != nil {
		return e
	}
	if e = b.syncDir(); e != nil {
		return e
	}
	var random [16]byte
	if _, e = rand.Read(random[:]); e != nil {
		return e
	}
	name := "temp-" + hex.EncodeToString(random[:])
	f, _, e := checkedFile(b.dir, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL)
	if e != nil {
		return e
	}
	renamed := false
	defer func() {
		if !renamed {
			result = errors.Join(result, unix.Unlinkat(b.dir, name, 0))
		}
	}()
	n, e := f.Write(data)
	if e == nil && n != len(data) {
		e = io.ErrShortWrite
	}
	if e == nil {
		e = f.Sync()
	}
	e = errors.Join(e, f.Close())
	if e != nil {
		return e
	}
	if e = unix.Renameat(b.dir, name, b.dir, "state"); e != nil {
		return e
	}
	renamed = true
	if e = b.syncDir(); e != nil {
		return e
	}
	if e = unix.Unlinkat(b.dir, "dirty", 0); e != nil {
		return e
	}
	return b.syncDir()
}
func (b *Budget) createLive() (string, *os.File, error) {
	var id [16]byte
	if _, e := rand.Read(id[:]); e != nil {
		return "", nil, e
	}
	name := "live-" + hex.EncodeToString(id[:])
	f, _, e := checkedFile(b.dir, name, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL)
	if e != nil {
		return "", nil, e
	}
	if e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); e != nil {
		f.Close()
		unix.Unlinkat(b.dir, name, 0)
		return "", nil, e
	}
	return hex.EncodeToString(id[:]), f, nil
}
func (b *Budget) clean(s queueState, self string) (queueState, []string, bool, error) {
	out := make([]entry, 0, len(s.Tickets))
	dead := []string{}
	found := false
	for _, t := range s.Tickets {
		if t.ID == self {
			found = true
			out = append(out, t)
			continue
		}
		f, _, e := checkedFile(b.dir, "live-"+t.ID, unix.O_RDWR)
		if e != nil {
			return s, nil, false, e
		}
		e = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if e == nil {
			if e = release(f); e != nil {
				return s, nil, false, e
			}
			dead = append(dead, t.ID)
			continue
		}
		ce := f.Close()
		if !errors.Is(e, unix.EWOULDBLOCK) && !errors.Is(e, unix.EAGAIN) {
			return s, nil, false, errors.Join(e, ce)
		}
		if ce != nil {
			return s, nil, false, ce
		}
		out = append(out, t)
	}
	s.Tickets = out
	return s, dead, found, nil
}
func (b *Budget) persist(s queueState, dead []string) error {
	if e := b.write(s); e != nil {
		return e
	}
	for _, id := range dead {
		if e := unix.Unlinkat(b.dir, "live-"+id, 0); e != nil {
			return e
		}
	}
	if len(dead) > 0 {
		return unix.Fsync(b.dir)
	}
	return nil
}

type Lease struct {
	b                   *Budget
	wire                *os.File
	mu                  sync.Mutex
	released            bool
	releaseErr          error
	observationStarted  bool
	observationSequence uint64
}

func (l *Lease) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.released {
		return l.releaseErr
	}
	l.released = true
	err := l.b.releaseWire(l.wire)
	l.releaseErr = err
	l.b.mu.Lock()
	l.b.active--
	l.b.mu.Unlock()
	return err
}
func (b *Budget) releaseWire(f *os.File) error {
	if f == nil {
		return nil
	}
	if b.testWireRelease != nil {
		return b.testWireRelease(f)
	}
	return release(f)
}
func (b *Budget) Acquire(ctx context.Context, p Priority) (lease *Lease, err error) {
	if p != Normal && p != Critical && p != Auxiliary {
		return nil, ErrUnavailable
	}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, ErrUnavailable
	}
	b.active++
	b.mu.Unlock()
	defer func() {
		if lease == nil {
			b.mu.Lock()
			b.active--
			b.mu.Unlock()
		}
	}()
	if e := b.verify(); e != nil {
		return nil, errors.Join(ErrUnavailable, e)
	}
	if p == Auxiliary {
		return b.tryAux(ctx)
	}
	id, live, e := b.createLive()
	if e != nil {
		return nil, errors.Join(ErrUnavailable, e)
	}
	published := false
	defer func() {
		if live == nil {
			return
		}
		if !published {
			err = errors.Join(err, release(live), unix.Unlinkat(b.dir, "live-"+id, 0))
			return
		}
		err = errors.Join(err, b.cancel(id, live))
	}()
	q, e := b.lock(ctx, "queue.lock", true)
	if e != nil {
		return nil, e
	}
	s, e := b.read()
	if e == nil {
		e = ctx.Err()
	}
	var deadBeforeRegister []string
	if e == nil && len(s.Tickets) >= maxTickets {
		s, deadBeforeRegister, _, e = b.clean(s, "")
	}
	if e == nil && (len(s.Tickets) >= maxTickets || s.Next == math.MaxUint64) {
		e = errors.New("queue full")
	}
	if e == nil {
		s.Tickets = append(s.Tickets, entry{id, s.Next, p == Critical})
		s.Next++
		published = true
		e = b.persist(s, deadBeforeRegister)
	}
	e = errors.Join(e, release(q))
	if e != nil {
		return nil, errors.Join(ErrUnavailable, e)
	}
	for {
		q, e = b.lock(ctx, "queue.lock", true)
		if e != nil {
			return nil, e
		}
		s, e = b.read()
		var dead []string
		var found bool
		if e == nil {
			s, dead, found, e = b.clean(s, id)
		}
		if e == nil && !found {
			e = errors.New("ticket missing")
		}
		var wire *os.File
		if e == nil {
			sel := s.Tickets[0]
			for _, t := range s.Tickets[1:] {
				if t.Critical && !sel.Critical || t.Critical == sel.Critical && t.Seq < sel.Seq {
					sel = t
				}
			}
			if sel.ID == id {
				wire, e = b.lock(ctx, "wire.lock", false)
			}
		}
		if e == nil && wire != nil {
			for i, t := range s.Tickets {
				if t.ID == id {
					s.Tickets = append(s.Tickets[:i], s.Tickets[i+1:]...)
					break
				}
			}
		}
		if e == nil && (wire != nil || len(dead) > 0) {
			e = b.persist(s, dead)
		}
		e = errors.Join(e, release(q))
		if e != nil {
			return nil, errors.Join(ErrUnavailable, e, b.releaseWire(wire))
		}
		if wire != nil {
			published = false
			ce := errors.Join(release(live), unix.Unlinkat(b.dir, "live-"+id, 0))
			live = nil
			if ce != nil {
				return nil, errors.Join(ErrUnavailable, ce, b.releaseWire(wire))
			}
			if e = b.clk.Sleep(ctx, interval); e != nil {
				return nil, errors.Join(e, b.releaseWire(wire))
			}
			return &Lease{b: b, wire: wire}, nil
		}
		if e = b.clk.Sleep(ctx, 5*time.Millisecond); e != nil {
			return nil, e
		}
	}
}
func (b *Budget) cancel(id string, live *os.File) error {
	q, e := b.lock(context.Background(), "queue.lock", false)
	if e != nil || q == nil {
		// 잠금 중인 peer가 죽은 waiter를 회수하도록 파일과 ticket은 유지함.
		return errors.Join(e, release(live))
	}
	s, e := b.read()
	if e == nil {
		for i, t := range s.Tickets {
			if t.ID == id {
				s.Tickets = append(s.Tickets[:i], s.Tickets[i+1:]...)
				e = b.write(s)
				break
			}
		}
	}
	if e == nil {
		e = unix.Unlinkat(b.dir, "live-"+id, 0)
	}
	return errors.Join(e, release(q), release(live))
}
func (b *Budget) tryAux(ctx context.Context) (*Lease, error) {
	q, e := b.lock(ctx, "queue.lock", false)
	if e != nil || q == nil {
		return nil, e
	}
	s, e := b.read()
	var dead []string
	if e == nil {
		s, dead, _, e = b.clean(s, "")
	}
	if e == nil && len(dead) > 0 {
		e = b.persist(s, dead)
	}
	var wire *os.File
	if e == nil && len(s.Tickets) == 0 {
		wire, e = b.lock(ctx, "wire.lock", false)
	}
	e = errors.Join(e, release(q))
	if e != nil {
		return nil, errors.Join(ErrUnavailable, e, b.releaseWire(wire))
	}
	if wire == nil {
		return nil, nil
	}
	if e = b.clk.Sleep(ctx, interval); e != nil {
		return nil, errors.Join(e, b.releaseWire(wire))
	}
	return &Lease{b: b, wire: wire}, nil
}
