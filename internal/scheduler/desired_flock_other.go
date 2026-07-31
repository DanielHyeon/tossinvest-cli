//go:build !unix

package scheduler

import (
	"context"
	"errors"
)

type desiredLock struct{}

func (desiredLock) release() {}

func acquireDesiredLock(ctx context.Context, _ string) (desiredLock, error) {
	if err := ctx.Err(); err != nil {
		return desiredLock{}, err
	}
	return desiredLock{}, errors.New("scheduler: saving desired state is unsupported without cross-process locking")
}
