//go:build linux

package accountbudget

import (
	"context"
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// OpenObserved binds a provisioned observation store before publishing its Budget.
// Deployment identity validation is the caller's responsibility.
func OpenObserved(ctx context.Context, cfg Config, expected FileID, boot clock.BootSource) (*Budget, error) {
	b, err := OpenContext(ctx, cfg)
	if err != nil {
		return nil, err
	}
	s, err := OpenObservation(ctx, b, expected, boot)
	if err != nil {
		return nil, errors.Join(err, b.Close())
	}
	b.observation = s
	return b, nil
}
