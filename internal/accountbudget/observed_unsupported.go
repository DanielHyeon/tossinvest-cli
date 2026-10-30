//go:build !linux

package accountbudget

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func OpenObserved(context.Context, Config, FileID, clock.BootSource) (*Budget, error) {
	return nil, ErrUnavailable
}

func (*Budget) EnginePollKey(string, string) (ObservationKey, error) {
	return ObservationKey{}, ErrUnavailable
}
func (*Budget) PollRemaining(context.Context, ObservationKey) (time.Duration, error) {
	return 0, ErrUnavailable
}
