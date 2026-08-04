package strategyprojection

import (
	"context"
	"errors"
	"sync"
)

// Store is the engine-owned, read-only-at-the-boundary projection source.
// Replace is intentionally available only to in-process assembly code; the
// Unix transport exposes Store solely through Reader.
type Store struct {
	mu       sync.RWMutex
	snapshot Snapshot
}

func NewStore(initial Snapshot) (*Store, error) {
	if err := Validate(initial); err != nil {
		return nil, err
	}
	return &Store{snapshot: Clone(initial)}, nil
}

func (s *Store) Read(ctx context.Context) (Snapshot, error) {
	if s == nil {
		return Snapshot{}, errors.New("strategy projection: store unavailable")
	}
	if ctx == nil {
		return Snapshot{}, errors.New("strategy projection: context unavailable")
	}
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Clone(s.snapshot), nil
}

func (s *Store) Replace(next Snapshot) error {
	if s == nil {
		return errors.New("strategy projection: store unavailable")
	}
	if err := Validate(next); err != nil {
		return err
	}
	s.mu.Lock()
	s.snapshot = Clone(next)
	s.mu.Unlock()
	return nil
}

var _ Reader = (*Store)(nil)
