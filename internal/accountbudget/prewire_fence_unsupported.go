//go:build !linux

package accountbudget

import "context"

type ObservationFence struct{}

func (*Budget) NewObservationFence(context.Context) (*ObservationFence, error) {
	return nil, ErrUnavailable
}
func (*ObservationFence) Context(context.Context) (context.Context, error) {
	return nil, ErrUnavailable
}
