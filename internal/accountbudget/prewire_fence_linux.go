//go:build linux

package accountbudget

import (
	"context"
	"errors"
	"io"
	"sync"
)

type observationFenceKey struct{}

// ObservationFence binds one stable observation cursor to its creating Budget.
// Its first integrity failure is retained for all requests using this fence.
type ObservationFence struct {
	budget     *Budget
	cursor     ObservationCursor
	generation uint64
	mu         sync.Mutex
	failure    error
}

func (b *Budget) NewObservationFence(ctx context.Context) (*ObservationFence, error) {
	if b == nil || ctx == nil {
		return nil, ErrUnavailable
	}
	view, err := b.ObservationSnapshot(ctx, ObservationCursor{})
	if err != nil {
		return nil, errors.Join(ErrUnavailable, err)
	}
	if view.Pending != nil || view.PendingStatus != ObservationPendingNone {
		return nil, ErrUnavailable
	}
	return &ObservationFence{budget: b, cursor: view.Cursor, generation: view.Generation}, nil
}
func (f *ObservationFence) Context(ctx context.Context) (context.Context, error) {
	if f == nil || f.budget == nil || ctx == nil || f.cursor.Scope == "" || f.cursor.Instance == "" {
		return nil, ErrUnavailable
	}
	return context.WithValue(ctx, observationFenceKey{}, f), nil
}
func (f *ObservationFence) latchLocked(err error) {
	if err != nil && f.failure == nil {
		f.failure = err
	}
}
func (f *ObservationFence) latch(err error) {
	if f == nil || err == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.latchLocked(err)
}
func (f *ObservationFence) initial(b *Budget) error {
	if f == nil || f.budget == nil || f.budget != b || f.cursor.Scope == "" || f.cursor.Instance == "" {
		return ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.failure
}
func (f *ObservationFence) check(ctx context.Context, b *Budget, l *Lease) error {
	if f == nil || f.budget != b || b == nil || b.observation == nil {
		return ErrUnavailable
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failure != nil {
		return f.failure
	}
	s := b.observation
	err := s.withLease(l, func() error {
		if l.observationStarted {
			return ErrUnavailable
		}
		lock, e := s.lock(ctx, true)
		if e != nil {
			return e
		}
		state, e := s.read()
		e = errors.Join(e, release(lock))
		if e != nil {
			return e
		}
		if state.Scope != f.cursor.Scope || state.Instance != f.cursor.Instance || state.Generation < f.generation || state.Event != f.cursor.Event || state.Gap != f.cursor.Gap || state.Pending != nil {
			return ErrUnavailable
		}
		return nil
	})
	if err != nil {
		err = errors.Join(ErrUnavailable, err)
		f.latchLocked(err)
	}
	return err
}
func (f *ObservationFence) release(l *Lease, known error) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.latchLocked(known)
	err := l.Release()
	f.latchLocked(err)
	return err
}

type observationFenceBody struct {
	io.ReadCloser
	fence     *ObservationFence
	collector *ObservationCollector
}

func (b *observationFenceBody) Close() error {
	err := b.ReadCloser.Close()
	b.fence.latch(err)
	b.collector.record(false, err)
	return err
}
