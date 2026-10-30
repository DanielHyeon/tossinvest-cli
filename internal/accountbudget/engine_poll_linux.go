//go:build linux

package accountbudget

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"time"
)

// EnginePollKey derives an opaque slot from the observed budget's immutable identity.
func (b *Budget) EnginePollKey(accountRef, kind string) (ObservationKey, error) {
	if b == nil || accountRef == "" || kind == "" {
		return ObservationKey{}, ErrUnavailable
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed || b.observation == nil {
		return ObservationKey{}, ErrUnavailable
	}
	h := sha256.New()
	for _, field := range []string{"tossos/a166/engine-poll/v1", b.scope, b.instance, accountRef, "engine-poll", kind} {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(field)))
		_, _ = h.Write(size[:])
		_, _ = h.Write([]byte(field))
	}
	var key ObservationKey
	copy(key[:], h.Sum(nil))
	return key, nil
}

// PollRemaining borrows the observation store while the factory owns the budget.
func (b *Budget) PollRemaining(ctx context.Context, key ObservationKey) (time.Duration, error) {
	if b == nil || ctx == nil {
		return 0, ErrUnavailable
	}
	b.mu.Lock()
	store := b.observation
	closed := b.closed
	b.mu.Unlock()
	if closed || store == nil {
		return 0, ErrUnavailable
	}
	return store.PollRemaining(ctx, key)
}
