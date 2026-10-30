//go:build linux

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
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"golang.org/x/sys/unix"
)

const observationMaxBytes = 256 << 10
const observationRing = 128
const observationSlots = 256
const observationLock = "event.lock"
const observationFile = "event.state"
const observationDirty = "event.dirty"

type ObservationStore struct {
	b                 *Budget
	lockID            FileID
	boot              clock.BootSource
	syncDir           func(int) error
	testFault         func(string) error
	testAfterSnapshot func()
}

// pinObservationBudget는 dirfd를 사용하는 동안 Budget.Close를 거절하게 함.
func pinObservationBudget(b *Budget) (func(), error) {
	if b == nil {
		return nil, ErrUnavailable
	}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, ErrUnavailable
	}
	b.active++
	b.mu.Unlock()
	return func() { b.mu.Lock(); b.active--; b.mu.Unlock() }, nil
}

// ProvisionObservation는 이미 준비된 root에 별도 event 저장소를 명시적으로 한 번만 만듦.
func ProvisionObservation(b *Budget) (FileID, error) {
	unpin, err := pinObservationBudget(b)
	if err != nil {
		return FileID{}, err
	}
	defer unpin()
	if err := b.verify(); err != nil {
		return FileID{}, errors.Join(ErrUnavailable, err)
	}
	if f, _, err := checkedFile(b.dir, observationDirty, unix.O_RDONLY); err == nil {
		return FileID{}, errors.Join(ErrUnavailable, f.Close())
	} else if !errors.Is(err, unix.ENOENT) {
		return FileID{}, errors.Join(ErrUnavailable, err)
	}
	lock, st, err := checkedFile(b.dir, observationLock, unix.O_RDWR|unix.O_CREAT|unix.O_EXCL)
	if err != nil {
		return FileID{}, errors.Join(ErrUnavailable, err)
	}
	err = errors.Join(lock.Sync(), lock.Close())
	if err != nil {
		return FileID{}, errors.Join(ErrUnavailable, err)
	}
	f, _, err := checkedFile(b.dir, observationFile, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL)
	if err != nil {
		return FileID{}, errors.Join(ErrUnavailable, err)
	}
	initial := observationState{Version: 1, Instance: b.instance, Scope: b.scope, Events: []ObservationEvent{}, Gaps: []ObservationGap{}, Slots: []observationSlot{}}
	data, _ := json.Marshal(initial)
	n, err := f.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err == nil {
		err = unix.Fsync(b.dir)
	}
	if err != nil {
		return FileID{}, errors.Join(ErrUnavailable, err)
	}
	return FileID{Device: uint64(st.Dev), Inode: st.Ino}, nil
}

// OpenObservation는 외부에서 pin한 event lock과 기존 Budget을 결속함.
func OpenObservation(ctx context.Context, b *Budget, expected FileID, boot clock.BootSource) (*ObservationStore, error) {
	if b == nil || boot == nil || expected.Device == 0 || expected.Inode == 0 {
		return nil, ErrUnavailable
	}
	unpin, err := pinObservationBudget(b)
	if err != nil {
		return nil, err
	}
	defer unpin()
	s := &ObservationStore{b: b, lockID: expected, boot: boot, syncDir: unix.Fsync}
	f, err := s.lock(ctx, true)
	if err != nil {
		return nil, err
	}
	_, err = s.read()
	err = errors.Join(err, release(f))
	if err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	return s, nil
}
func (s *ObservationStore) lock(ctx context.Context, wait bool) (*os.File, error) {
	if s == nil || s.b == nil || ctx == nil {
		return nil, ErrUnavailable
	}
	if err := s.b.verify(); err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	f, st, err := checkedFile(s.b.dir, observationLock, unix.O_RDWR)
	if err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	if uint64(st.Dev) != s.lockID.Device || st.Ino != s.lockID.Inode {
		return nil, errors.Join(ErrUnavailable, f.Close())
	}
	for {
		if err = ctx.Err(); err != nil {
			return nil, errors.Join(err, f.Close())
		}
		err = unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			if err = ctx.Err(); err != nil {
				return nil, errors.Join(err, release(f))
			}
			return f, nil
		}
		if !errors.Is(err, unix.EAGAIN) && !errors.Is(err, unix.EWOULDBLOCK) {
			return nil, errors.Join(ErrUnavailable, err, f.Close())
		}
		if !wait {
			return nil, errors.Join(ErrUnavailable, f.Close())
		}
		if err = s.b.clk.Sleep(ctx, 5*time.Millisecond); err != nil {
			return nil, errors.Join(err, f.Close())
		}
	}
}
func validObservationKind(k ObservationKind) bool {
	return k == ObservationPoll || k == ObservationToken || k == ObservationOther
}
func validObservationKey(k ObservationKey) bool { return k != ObservationKey{} }
func validObservationPending(p observationPending) bool {
	return p.Sequence > 0 && validObservationKind(p.Kind) && ((p.Kind == ObservationPoll && validObservationKey(p.Key)) || (p.Kind != ObservationPoll && !validObservationKey(p.Key)))
}
func validObservationStamp(st clock.BootStamp) bool {
	_, err := clock.RemainingBootHint(0, st, st, nil)
	return err == nil
}
func validateObservationState(s observationState) error {
	if s.Version != 1 || !validScope(s.Instance) || !validScope(s.Scope) || s.Events == nil || s.Gaps == nil || s.Slots == nil || len(s.Events) > observationRing || len(s.Gaps) > observationRing || len(s.Slots) > observationSlots || (s.Pending != nil && (!validObservationPending(*s.Pending) || s.Pending.Sequence != s.Attempt)) || s.Event < uint64(len(s.Events)) || s.Gap < uint64(len(s.Gaps)) {
		return errors.New("invalid observation state")
	}
	if uint64(len(s.Events)) != min(s.Event, uint64(observationRing)) || uint64(len(s.Gaps)) != min(s.Gap, uint64(observationRing)) {
		return errors.New("incomplete observation ring")
	}
	if s.Event > s.Attempt || s.Gap > s.Attempt-s.Event || (s.Pending != nil && s.Event+s.Gap >= s.Attempt) {
		return errors.New("invalid observation counters")
	}
	seenAttempts := map[uint64]bool{}
	var previous uint64
	for i, e := range s.Events {
		if e.Sequence != s.Event-uint64(len(s.Events))+uint64(i)+1 || e.Attempt == 0 || e.Attempt > s.Attempt || e.Attempt <= previous || (s.Pending != nil && e.Attempt == s.Pending.Sequence) || !validObservationKind(e.Kind) || !validObservationStamp(e.Stamp) {
			return errors.New("invalid observation event")
		}
		previous = e.Attempt
		seenAttempts[e.Attempt] = true
	}
	previous = 0
	for i, g := range s.Gaps {
		if g.Sequence != s.Gap-uint64(len(s.Gaps))+uint64(i)+1 || g.Attempt == 0 || g.Attempt > s.Attempt || g.Attempt <= previous || seenAttempts[g.Attempt] || (s.Pending != nil && g.Attempt == s.Pending.Sequence) || !validObservationKind(g.Kind) {
			return errors.New("invalid observation gap")
		}
		previous = g.Attempt
		seenAttempts[g.Attempt] = true
	}
	completed := s.Attempt - s.Gap
	if s.Pending != nil {
		completed--
	}
	if (completed == 0) != (s.Last == nil) {
		return errors.New("invalid completion presence")
	}
	if s.Last != nil {
		last := s.Last
		if last.Attempt == 0 || last.Attempt > s.Attempt || !validObservationKind(last.Kind) || (last.Result != ObservationNoResponse && last.Result != ObservationResponse && last.Result != ObservationRateLimited) || (s.Pending != nil && last.Attempt >= s.Pending.Sequence) {
			return errors.New("invalid completion")
		}
		if len(s.Events) > 0 && last.Attempt < s.Events[len(s.Events)-1].Attempt {
			return errors.New("completion predates latest event")
		}
		if last.Result == ObservationRateLimited {
			if len(s.Events) == 0 || s.Events[len(s.Events)-1].Attempt != last.Attempt || s.Events[len(s.Events)-1].Kind != last.Kind {
				return errors.New("missing completion event")
			}
		} else if seenAttempts[last.Attempt] {
			return errors.New("completion contradicts record")
		}
	}
	seen := map[ObservationKey]bool{}
	for _, slot := range s.Slots {
		if !validObservationKey(slot.Key) || seen[slot.Key] || (!slot.Poisoned && slot.Hint == nil) {
			return errors.New("invalid observation slot")
		}
		seen[slot.Key] = true
		if slot.Hint != nil {
			h := slot.Hint
			if h.Duration <= 0 || h.Duration > 30*time.Second || !validObservationStamp(h.Stamp) || (h.Anchor != nil && !validObservationStamp(*h.Anchor)) {
				return errors.New("invalid observation hint")
			}
			if h.Anchor != nil {
				if _, err := clock.RemainingBootHint(h.Duration, h.Stamp, *h.Anchor, h.Anchor); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (s *ObservationStore) read() (observationState, error) {
	var state observationState
	f, _, err := checkedFile(s.b.dir, observationDirty, unix.O_RDONLY)
	if err == nil {
		return state, errors.Join(errors.New("dirty observation"), f.Close())
	}
	if !errors.Is(err, unix.ENOENT) {
		return state, err
	}
	f, _, err = checkedFile(s.b.dir, observationFile, unix.O_RDONLY)
	if err != nil {
		return state, err
	}
	data, err := readBounded(f, observationMaxBytes)
	err = errors.Join(err, f.Close())
	if err != nil {
		return state, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&state); err != nil {
		return state, err
	}
	canonical, err := json.Marshal(state)
	if err != nil || !bytes.Equal(canonical, data) {
		return state, errors.New("noncanonical observation state")
	}
	if err = validateObservationState(state); err != nil {
		return state, err
	}
	if state.Instance != s.b.instance || state.Scope != s.b.scope {
		return state, errors.New("observation instance mismatch")
	}
	return state, nil
}
func (s *ObservationStore) fault(stage string) error {
	if s.testFault != nil {
		return s.testFault(stage)
	}
	return nil
}
func (s *ObservationStore) write(state observationState) (out error) {
	if state.Instance != s.b.instance || state.Scope != s.b.scope {
		return errors.New("observation instance mismatch")
	}
	if state.Generation == math.MaxUint64 {
		return errors.New("observation generation overflow")
	}
	// pending이 남는 갱신은 Finish에 필요한 마지막 세대를 남겨 둠.
	if state.Pending != nil && state.Generation >= math.MaxUint64-1 {
		return errors.New("observation finish generation unavailable")
	}
	state.Generation++
	if err := validateObservationState(state); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil || len(data) > observationMaxBytes {
		return errors.New("observation state too large")
	}
	marker, _, err := checkedFile(s.b.dir, observationDirty, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL)
	if err != nil {
		return err
	}
	err = s.fault("marker-sync")
	if err == nil {
		err = marker.Sync()
	}
	err = errors.Join(err, marker.Close())
	if err != nil {
		return err
	}
	if err = s.syncDir(s.b.dir); err != nil {
		return err
	}
	var token [16]byte
	if _, err = rand.Read(token[:]); err != nil {
		return err
	}
	name := "event-temp-" + hex.EncodeToString(token[:])
	f, _, err := checkedFile(s.b.dir, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL)
	if err != nil {
		return err
	}
	renamed := false
	defer func() {
		if !renamed {
			out = errors.Join(out, unix.Unlinkat(s.b.dir, name, 0))
		}
	}()
	n := 0
	err = s.fault("state-write")
	if err == nil {
		n, err = f.Write(data)
	}
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = s.fault("state-sync")
	}
	if err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	if err = s.fault("rename"); err != nil {
		return err
	}
	if err = unix.Renameat(s.b.dir, name, s.b.dir, observationFile); err != nil {
		return err
	}
	renamed = true
	if err = s.syncDir(s.b.dir); err != nil {
		return err
	}
	if err = unix.Unlinkat(s.b.dir, observationDirty, 0); err != nil {
		return err
	}
	return s.syncDir(s.b.dir)
}
func (s *ObservationStore) withLease(l *Lease, fn func() error) error {
	if s == nil || l == nil {
		return ErrUnavailable
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.b != s.b || l.released || l.wire == nil {
		return ErrUnavailable
	}
	return fn()
}
func (s *ObservationStore) sample() (clock.BootStamp, error) {
	st, err := s.boot.Snapshot()
	if err != nil {
		return st, err
	}
	if !validObservationStamp(st) {
		return st, clock.ErrBootStamp
	}
	return st, nil
}
func observationSlotIndex(slots []observationSlot, key ObservationKey) int {
	for i := range slots {
		if slots[i].Key == key {
			return i
		}
	}
	return -1
}

// pruneObservationSlots는 만료된 정상 hint만 제거하고 poison은 보존함.
func pruneObservationSlots(state *observationState, now clock.BootStamp) error {
	kept := state.Slots[:0]
	for _, slot := range state.Slots {
		if !slot.Poisoned && slot.Hint != nil {
			result, err := clock.RemainingBootHint(slot.Hint.Duration, slot.Hint.Stamp, now, slot.Hint.Anchor)
			if err != nil {
				return err
			}
			if result.NeedsAnchor {
				anchor := now
				slot.Hint.Anchor = &anchor
			}
			if result.Remaining == 0 {
				continue
			}
		}
		kept = append(kept, slot)
	}
	state.Slots = kept
	return nil
}
func (s *ObservationStore) Begin(ctx context.Context, l *Lease, kind ObservationKind, key ObservationKey) (sequence uint64, err error) {
	err = s.withLease(l, func() (out error) {
		p := observationPending{Sequence: 1, Kind: kind, Key: key}
		if l.observationStarted {
			return ErrUnavailable
		}
		l.observationStarted = true
		if !validObservationPending(p) {
			return ErrUnavailable
		}
		lock, e := s.lock(ctx, true)
		if e != nil {
			return e
		}
		defer func() {
			out = errors.Join(out, release(lock))
			if out == nil {
				l.observationSequence = sequence
			}
		}()
		state, e := s.read()
		if e != nil {
			return errors.Join(ErrUnavailable, e)
		}
		if state.Attempt == math.MaxUint64 || state.Event == math.MaxUint64 || state.Gap == math.MaxUint64 || state.Generation >= math.MaxUint64-1 {
			return ErrUnavailable
		}
		if state.Pending != nil {
			old := *state.Pending
			if old.Kind == ObservationPoll {
				i := observationSlotIndex(state.Slots, old.Key)
				if i < 0 {
					if len(state.Slots) >= observationSlots {
						now, err := s.sample()
						if err != nil {
							return errors.Join(ErrUnavailable, err)
						}
						if err = pruneObservationSlots(&state, now); err != nil {
							return errors.Join(ErrUnavailable, err)
						}
						if len(state.Slots) >= observationSlots {
							return ErrUnavailable
						}
					}
					state.Slots = append(state.Slots, observationSlot{Key: old.Key, Poisoned: true})
				} else {
					state.Slots[i].Poisoned = true
				}
			}
			state.Gap++
			state.Gaps = append(state.Gaps, ObservationGap{Sequence: state.Gap, Attempt: old.Sequence, Kind: old.Kind})
			if len(state.Gaps) > observationRing {
				state.Gaps = state.Gaps[1:]
			}
		}
		state.Attempt++
		sequence = state.Attempt
		state.Pending = &observationPending{Sequence: sequence, Kind: kind, Key: key}
		if e = ctx.Err(); e != nil {
			return e
		}
		if e = s.write(state); e != nil {
			return errors.Join(ErrUnavailable, e)
		}
		return ctx.Err()
	})
	if err != nil {
		return 0, err
	}
	return sequence, nil
}
func (s *ObservationStore) Finish(ctx context.Context, l *Lease, sequence uint64, result ObservationFinish, hint *time.Duration) (err error) {
	return s.withLease(l, func() (out error) {
		if result != ObservationNoResponse && result != ObservationResponse && result != ObservationRateLimited {
			return ErrUnavailable
		}
		if hint != nil && (*hint < 0 || *hint > 30*time.Second || result != ObservationRateLimited) {
			// 잘못된 hint는 버리되 실제 429 사건의 기록을 막지 않음.
			hint = nil
		}
		lock, e := s.lock(ctx, true)
		if e != nil {
			return e
		}
		defer func() { out = errors.Join(out, release(lock)) }()
		state, e := s.read()
		if e != nil {
			return errors.Join(ErrUnavailable, e)
		}
		if !l.observationStarted || l.observationSequence != sequence || state.Pending == nil || state.Pending.Sequence != sequence {
			return ErrUnavailable
		}
		pending := *state.Pending
		if result == ObservationRateLimited {
			if state.Event == math.MaxUint64 {
				return ErrUnavailable
			}
			now, e := s.sample()
			if e != nil {
				return errors.Join(ErrUnavailable, e)
			}
			if hint != nil && pending.Kind == ObservationPoll && *hint > 0 {
				i := observationSlotIndex(state.Slots, pending.Key)
				if i < 0 {
					if len(state.Slots) >= observationSlots {
						if err := pruneObservationSlots(&state, now); err != nil {
							return errors.Join(ErrUnavailable, err)
						}
						if len(state.Slots) >= observationSlots {
							return ErrUnavailable
						}
					}
					state.Slots = append(state.Slots, observationSlot{Key: pending.Key})
					i = len(state.Slots) - 1
				}
				slot := &state.Slots[i]
				remaining := time.Duration(0)
				if slot.Hint != nil {
					r, e := clock.RemainingBootHint(slot.Hint.Duration, slot.Hint.Stamp, now, slot.Hint.Anchor)
					if e != nil {
						return errors.Join(ErrUnavailable, e)
					}
					remaining = r.Remaining
					if r.NeedsAnchor {
						a := now
						slot.Hint.Anchor = &a
					}
				}
				if *hint > remaining {
					slot.Hint = &observationHint{Stamp: now, Duration: *hint}
				}
			}
			state.Event++
			state.Events = append(state.Events, ObservationEvent{Sequence: state.Event, Attempt: sequence, Kind: pending.Kind, Stamp: now})
			if len(state.Events) > observationRing {
				state.Events = state.Events[1:]
			}
		}
		state.Last = &ObservationCompletion{Attempt: sequence, Kind: pending.Kind, Result: result}
		state.Pending = nil
		if e = ctx.Err(); e != nil {
			return e
		}
		if e = s.write(state); e != nil {
			return errors.Join(ErrUnavailable, e)
		}
		return ctx.Err()
	})
}
func (s *ObservationStore) PollRemaining(ctx context.Context, key ObservationKey) (remaining time.Duration, err error) {
	if s == nil {
		return 0, ErrUnavailable
	}
	unpin, err := pinObservationBudget(s.b)
	if err != nil {
		return 0, err
	}
	defer unpin()
	if !validObservationKey(key) {
		return 0, ErrUnavailable
	}
	lock, e := s.lock(ctx, true)
	if e != nil {
		return 0, e
	}
	defer func() { err = errors.Join(err, release(lock)) }()
	state, e := s.read()
	if e != nil {
		return 0, errors.Join(ErrUnavailable, e)
	}
	i := observationSlotIndex(state.Slots, key)
	if i < 0 {
		return 0, nil
	}
	slot := &state.Slots[i]
	if slot.Poisoned {
		return 0, ErrObservationPoisoned
	}
	if slot.Hint == nil {
		return 0, ErrUnavailable
	}
	now, e := s.sample()
	if e != nil {
		return 0, errors.Join(ErrUnavailable, e)
	}
	r, e := clock.RemainingBootHint(slot.Hint.Duration, slot.Hint.Stamp, now, slot.Hint.Anchor)
	if e != nil {
		return 0, errors.Join(ErrUnavailable, e)
	}
	if r.NeedsAnchor {
		a := now
		slot.Hint.Anchor = &a
		if e = s.write(state); e != nil {
			return 0, errors.Join(ErrUnavailable, e)
		}
	} else if r.Remaining == 0 {
		state.Slots = append(state.Slots[:i], state.Slots[i+1:]...)
		if e = s.write(state); e != nil {
			return 0, errors.Join(ErrUnavailable, e)
		}
	}
	return r.Remaining, nil
}
func (s *ObservationStore) Snapshot(ctx context.Context, cursor ObservationCursor) (view ObservationView, err error) {
	if s == nil {
		return view, ErrUnavailable
	}
	unpin, err := pinObservationBudget(s.b)
	if err != nil {
		return view, err
	}
	defer unpin()
	if s.b == nil || (cursor.Instance == "" && (cursor.Scope != "" || cursor.Event != 0 || cursor.Gap != 0)) || (cursor.Instance != "" && (cursor.Instance != s.b.instance || cursor.Scope != s.b.scope)) {
		return view, ErrUnavailable
	}
	eventCursor, gapCursor := cursor.Event, cursor.Gap
	lock, e := s.lock(ctx, true)
	if e != nil {
		return view, e
	}
	state, e := s.read()
	e = errors.Join(e, release(lock))
	if e != nil {
		return view, errors.Join(ErrUnavailable, e)
	}
	if eventCursor > state.Event || gapCursor > state.Gap {
		return view, ErrUnavailable
	}
	view = observationView(state, eventCursor, gapCursor)
	view.Cursor = ObservationCursor{Scope: s.b.scope, Instance: s.b.instance, Event: state.Event, Gap: state.Gap}
	if state.Pending == nil {
		return view, nil
	}
	if s.testAfterSnapshot != nil {
		s.testAfterSnapshot()
	}
	wire, e := s.b.lock(ctx, "wire.lock", false)
	if e != nil {
		return ObservationView{}, e
	}
	if wire != nil {
		defer func() { err = errors.Join(err, release(wire)) }()
	}
	lock, e = s.lock(ctx, true)
	if e != nil {
		return ObservationView{}, e
	}
	second, e := s.read()
	e = errors.Join(e, release(lock))
	if e != nil {
		return ObservationView{}, errors.Join(ErrUnavailable, e)
	}
	if second.Generation != state.Generation {
		return ObservationView{}, fmt.Errorf("%w: observation generation changed", ErrUnavailable)
	}
	if wire == nil {
		view.PendingStatus = ObservationPendingBusy
	} else {
		view.PendingStatus = ObservationPendingOrphan
	}
	return view, nil
}
func observationView(s observationState, ec, gc uint64) ObservationView {
	v := ObservationView{Generation: s.Generation, Last: s.Last}
	if s.Pending != nil {
		v.Pending = &ObservationPending{Sequence: s.Pending.Sequence, Kind: s.Pending.Kind}
	}
	if len(s.Events) > 0 && ec < s.Events[0].Sequence-1 {
		v.EventOverrun = true
	}
	if len(s.Gaps) > 0 && gc < s.Gaps[0].Sequence-1 {
		v.GapOverrun = true
	}
	for _, e := range s.Events {
		if e.Sequence > ec {
			v.Events = append(v.Events, e)
		}
	}
	for _, g := range s.Gaps {
		if g.Sequence > gc {
			v.Gaps = append(v.Gaps, g)
		}
	}
	return v
}
