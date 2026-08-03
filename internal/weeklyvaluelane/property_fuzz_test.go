package weeklyvaluelane

import (
	"math/rand"
	"testing"
)

func TestSevenLegQuantityNeverExceedsImmutableCeilingOrCapProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(6901))
	for i := 0; i < 5000; i++ {
		q := uint64(rng.Int63()) + 1
		plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, q, "100000000000000000000000000000")
		for index, ceiling := range plan.LegCeilings() {
			filled := uint64(0)
			if ceiling > 0 {
				filled = uint64(rng.Int63n(int64(min64(ceiling, uint64(^uint64(0)>>1))) + 1))
			}
			cap := uint64(rng.Int63())
			got := PlannedLegQuantity(plan, LegProgress{Ordinal: index + 1, FilledQuantity: filled}, cap)
			if got > ceiling-filled || got > cap {
				t.Fatalf("Q=%d ordinal=%d got=%d ceiling=%d filled=%d cap=%d", q, index+1, got, ceiling, filled, cap)
			}
		}
	}
}

func FuzzAllocateSevenConservesQuantity(f *testing.F) {
	for _, q := range []uint64{0, 1, 7, 14, 15, 100, ^uint64(0)} {
		f.Add(q)
	}
	f.Fuzz(func(t *testing.T, q uint64) {
		got := AllocateSeven(q)
		if sumSeven(got) != q {
			t.Fatalf("Q=%d allocation=%v", q, got)
		}
	})
}

func FuzzMissingFillRetryIsIdempotent(f *testing.F) {
	for _, qty := range []uint64{0, 1, 2, 100, ^uint64(0)} {
		f.Add(qty)
	}
	f.Fuzz(func(t *testing.T, qty uint64) {
		plan := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "100000000000000000000000000000")
		state := mustRiskState(t, plan, "0", "100000000000000000000000")
		event := validFill(plan, "")
		event.Quantity = qty
		first, _ := ApplyFillRisk(state, plan, event)
		second, result := ApplyFillRisk(first, plan, event)
		if !result.Duplicate || second.FilledMinor() != first.FilledMinor() || second.HeldMinor() != first.HeldMinor() || second.FillCount() != first.FillCount() {
			t.Fatalf("retry moved state: first=%+v second=%+v result=%+v", first, second, result)
		}
	})
}

func min64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
