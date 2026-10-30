//go:build !linux

package accountbudget

import (
	"context"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func OpenObserved(context.Context, Config, FileID, clock.BootSource) (*Budget, error) {
	return nil, ErrUnavailable
}
