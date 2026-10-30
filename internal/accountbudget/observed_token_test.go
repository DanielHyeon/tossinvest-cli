package accountbudget

import (
	"context"
	"testing"
	"time"
)

func TestA166TokenObservationRetainsContext(t *testing.T) {
	var key ObservationKey
	key[0] = 1
	collector := &ObservationCollector{}
	deadline := time.Now().Add(time.Hour)
	parent, cancel := context.WithDeadline(WithPriority(context.Background(), Critical), deadline)
	defer cancel()
	parent = WithObservation(parent, ObservationPoll, key, collector)
	child := WithTokenObservation(parent)
	scope := requestObservation(child)
	gotDeadline, ok := child.Deadline()
	if !ok || !gotDeadline.Equal(deadline) || PriorityFromContext(child) != Critical || scope.kind != ObservationToken || scope.key != (ObservationKey{}) || scope.collector != collector {
		t.Fatal("token scope lost parent properties")
	}
	cancel()
	if child.Err() != context.Canceled {
		t.Fatal("token scope lost cancellation")
	}
}
