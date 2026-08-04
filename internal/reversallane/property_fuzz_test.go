package reversallane

import (
	"math/rand"
	"testing"
)

func TestTwoFourEightCapNeverIncreasesProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(6801))
	for i := 0; i < 5000; i++ {
		q := uint64(rng.Int63())
		if q == 0 {
			continue
		}
		plan := mustPlan(t, MarketKR, KRReversalLaneID, q, "100000000000000000000000000000", "KRW", "KRW", nil)
		ceilings := plan.LegCeilings()
		if ceilings[0]+ceilings[1]+ceilings[2] != q {
			t.Fatalf("iteration=%d Q=%d ceilings=%v", i, q, ceilings)
		}
		for index, ceiling := range ceilings {
			filled := uint64(0)
			if ceiling > 0 {
				filled = uint64(rng.Int63n(int64(min64(ceiling, uint64(^uint64(0)>>1))) + 1))
			}
			cap := uint64(rng.Int63())
			got := PlannedLegQuantity(plan, LegProgress{Ordinal: index + 1, FilledQuantity: filled}, RiskCap{QFinal: cap, ReservationMinor: "1", Official: true, Frozen: true})
			if got > ceiling-filled || got > cap || got > ceiling {
				t.Fatalf("iteration=%d ordinal=%d got=%d ceiling=%d filled=%d cap=%d", i, index+1, got, ceiling, filled, cap)
			}
		}
	}
}

func FuzzTwoFourEightConservesQuantity(f *testing.F) {
	for _, q := range []uint64{0, 1, 7, 14, 15, 100, ^uint64(0)} {
		f.Add(q)
	}
	f.Fuzz(func(t *testing.T, q uint64) {
		got := AllocateTwoFourEight(q)
		if got[0]+got[1]+got[2] != q || got[0] > q || got[1] > q || got[2] > q {
			t.Fatalf("Q=%d allocation=%v", q, got)
		}
	})
}

func FuzzFillRetryIsIdempotent(f *testing.F) {
	for _, qty := range []uint64{1, 2, 100, ^uint64(0)} {
		f.Add(qty)
	}
	f.Fuzz(func(t *testing.T, qty uint64) {
		plan := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100000000000000000000000000000", "KRW", "KRW", nil)
		state := NewRiskState(plan)
		state.HeldMinor = "100000000000000000000000"
		event := validFillEvent(plan, "fill")
		event.Quantity = qty
		event.TransferredReservationMinor = "1"
		event.EntryPriceMinor = "2"
		event.EffectiveStopMinor = "1"
		event.EntryFeesMinor = "0"
		event.EstimatedExitFeesLeviesMinor = "0"
		first, result := ApplyFillRisk(state, plan, event)
		if qty == 0 {
			if result.Applied || result.Duplicate || first.Fills[event.FillID].Applied || !first.Latches[LatchUnknownActualRisk] {
				t.Fatal("zero fill was not preserved as non-applied evidence")
			}
			second, result := ApplyFillRisk(first, plan, event)
			if !result.Duplicate || result.Applied || second.FilledMinor != first.FilledMinor || len(second.Fills) != len(first.Fills) {
				t.Fatal("zero fill retry was not idempotent")
			}
			return
		}
		if !result.Applied {
			t.Fatal("authoritative fill not applied")
		}
		second, result := ApplyFillRisk(first, plan, event)
		if !result.Duplicate || second.FilledMinor != first.FilledMinor || len(second.Fills) != len(first.Fills) {
			t.Fatal("duplicate fill moved risk")
		}
	})
}

func min64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
