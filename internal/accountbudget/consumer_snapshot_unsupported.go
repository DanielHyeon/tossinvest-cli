//go:build !linux

package accountbudget

import "context"

func (*Budget) ObservationSnapshot(context.Context, ObservationCursor) (ObservationView, error) {
	return ObservationView{}, ErrUnavailable
}
