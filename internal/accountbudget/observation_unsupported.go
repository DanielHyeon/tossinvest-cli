//go:build !linux

package accountbudget

import (
	"context"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"time"
)

type ObservationStore struct{}

func ProvisionObservation(*Budget) (FileID, error) { return FileID{}, ErrUnavailable }
func OpenObservation(context.Context, *Budget, FileID, clock.BootSource) (*ObservationStore, error) {
	return nil, ErrUnavailable
}
func (*ObservationStore) Begin(context.Context, *Lease, ObservationKind, ObservationKey) (uint64, error) {
	return 0, ErrUnavailable
}
func (*ObservationStore) Finish(context.Context, *Lease, uint64, ObservationFinish, *time.Duration) error {
	return ErrUnavailable
}
func (*ObservationStore) PollRemaining(context.Context, ObservationKey) (time.Duration, error) {
	return 0, ErrUnavailable
}
func (*ObservationStore) Snapshot(context.Context, ObservationCursor) (ObservationView, error) {
	return ObservationView{}, ErrUnavailable
}
