package continuationlane

import (
	"math/rand"
	"testing"
)

func TestAllocationAndAdmissionNeverIncreaseProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(6701))
	for i := 0; i < 5000; i++ {
		q := uint64(rng.Int63())
		plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, q, "100000000000000000000000000000")
		var sum uint64
		for ordinal, ceiling := range plan.LegCeilings {
			sum += ceiling
			filled := uint64(0)
			if ceiling > 0 {
				filled = uint64(rng.Int63n(int64(minU64(ceiling, uint64(^uint64(0)>>1))) + 1))
			}
			cap := uint64(rng.Int63())
			remaining := ceiling - filled
			got := PlannedLegQuantity(plan, LegProgress{Ordinal: ordinal + 1, FilledQuantity: filled}, cap)
			if got > remaining || got > cap || got > ceiling {
				t.Fatalf("iteration=%d Q=%d ordinal=%d got=%d remaining=%d cap=%d ceiling=%d", i, q, ordinal+1, got, remaining, cap, ceiling)
			}
		}
		if sum != q {
			t.Fatalf("iteration=%d ceilings=%v sum=%d Q=%d", i, plan.LegCeilings, sum, q)
		}
	}
}

func FuzzEightFourTwoConservesQuantity(f *testing.F) {
	for _, q := range []uint64{0, 1, 2, 7, 14, 15, 100, ^uint64(0)} {
		f.Add(q)
	}
	f.Fuzz(func(t *testing.T, q uint64) {
		ceilings := AllocateEightFourTwo(q)
		if ceilings[0]+ceilings[1]+ceilings[2] != q || ceilings[0] > q || ceilings[1] > q || ceilings[2] > q {
			t.Fatalf("Q=%d ceilings=%v", q, ceilings)
		}
	})
}

func FuzzRiskFillRetryIsIdempotent(f *testing.F) {
	for _, qty := range []uint64{0, 1, 2, 100, ^uint64(0)} {
		f.Add(qty)
	}
	f.Fuzz(func(t *testing.T, qty uint64) {
		plan := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "100000000000000000000000000000")
		state := NewRiskState(plan)
		state.HeldMinor = "100000000000000000000000"
		event := scopedFillRiskEvent(plan, FillRiskEvent{FillID: "fill", Quantity: qty, TransferredReservationMinor: "1", EntryPriceMinor: "2", EffectiveStopMinor: "1", EntryFeesMinor: "0", EstimatedExitFeesLeviesMinor: "0"})
		first, result := ApplyFillRisk(state, plan, event)
		if qty == 0 {
			if result.Applied || result.Duplicate || first.Fills[event.FillID].Applied || first.FilledMinor != state.FilledMinor || first.HeldMinor != state.HeldMinor || !first.Latches[LatchUnknownActualRisk] {
				t.Fatal("zero fill was not preserved as non-applied evidence")
			}
			second, result := ApplyFillRisk(first, plan, event)
			if result.Applied || !result.Duplicate || second.FilledMinor != first.FilledMinor || second.HeldMinor != first.HeldMinor || len(second.Fills) != len(first.Fills) {
				t.Fatal("zero fill retry was not idempotent")
			}
			return
		}
		if !result.Applied {
			t.Fatal("authoritative fill not applied")
		}
		second, result := ApplyFillRisk(first, plan, event)
		if !result.Duplicate || second.FilledMinor != first.FilledMinor || len(second.Fills) != len(first.Fills) {
			t.Fatalf("retry moved risk: first=%+v second=%+v result=%+v", first, second, result)
		}
	})
}

func minU64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
