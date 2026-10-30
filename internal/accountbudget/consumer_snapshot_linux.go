//go:build linux

package accountbudget

import "context"

// ObservationSnapshot reads the observed budget without transferring its ownership.
func (b *Budget) ObservationSnapshot(ctx context.Context, cursor ObservationCursor) (ObservationView, error) {
	if b == nil || ctx == nil {
		return ObservationView{}, ErrUnavailable
	}
	b.mu.Lock()
	store, closed := b.observation, b.closed
	b.mu.Unlock()
	if closed || store == nil {
		return ObservationView{}, ErrUnavailable
	}
	return store.Snapshot(ctx, cursor)
}
