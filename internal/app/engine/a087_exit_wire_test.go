package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

func TestA087EligibleLowerLimitReachesPlaceIntentAsLimit(t *testing.T) {
	h := newExitHarness(t, func(options *engine.ExitObserverOptions) {
		engine.SetA087LowerLimitForTest(options, "kr", "005930", "50400", exitNow)
	})
	a087Breach(h)
	if len(h.submit.places) != 1 {
		t.Fatalf("Place 호출 = %d, want 1", len(h.submit.places))
	}
	intent := h.submit.places[0].Intent
	if intent.OrderType != "limit" || intent.Price != 50400 {
		t.Fatalf("PlaceIntent = %+v, want LIMIT 50400", intent)
	}
	if got := a087Wire(t, intent); got != `{"symbol":"005930","side":"SELL","orderType":"LIMIT","quantity":"10","price":"50400","timeInForce":"DAY","confirmHighValueOrder":false}` {
		t.Fatalf("lower-limit wire = %s", got)
	}
	recorded, err := h.journal.LookupIntent(context.Background(), h.submit.places[0].IntentID)
	if err != nil || recorded.Price != "50400" {
		t.Fatalf("intents.price = %q, %v; want 50400, nil", recorded.Price, err)
	}
	if !engine.A087LowerLimitAttemptedForTest(h.observer, "kr", "005930", h.clk.Now()) {
		t.Fatal("하한가 Place 뒤 첫 시도 금지 표시가 없음")
	}
}

func TestA087DormantBoardKeepsLegacyPlaceIntent(t *testing.T) {
	h := newExitHarness(t, nil)
	a087Breach(h)
	if len(h.submit.places) != 1 {
		t.Fatalf("Place 호출 = %d, want 1", len(h.submit.places))
	}
	intent := h.submit.places[0].Intent
	if intent.OrderType != "limit" || intent.Price != 67900 {
		t.Fatalf("PlaceIntent = %+v, want legacy LIMIT 67900", intent)
	}
	if got := a087Wire(t, intent); got != `{"symbol":"005930","side":"SELL","orderType":"LIMIT","quantity":"10","price":"67900","timeInForce":"DAY","confirmHighValueOrder":false}` {
		t.Fatalf("dormant wire = %s", got)
	}
	recorded, err := h.journal.LookupIntent(context.Background(), h.submit.places[0].IntentID)
	if err != nil || recorded.Price != "67900" {
		t.Fatalf("휴면 intents.price = %q, %v; want 67900, nil", recorded.Price, err)
	}

}

func TestA087AnyLowerLimitPlaceOutcomeClosesFirstAttempt(t *testing.T) {
	for _, state := range []journal.AttemptState{
		journal.StateConfirmed,
		journal.StateInDoubt,
		journal.StateFailedConfirmed,
	} {
		t.Run(string(state), func(t *testing.T) {
			h := newExitHarness(t, func(options *engine.ExitObserverOptions) {
				engine.SetA087LowerLimitForTest(options, "kr", "005930", "50400", exitNow)
			})
			h.submit.placeOutcome = &execgw.Outcome{State: state}
			a087Breach(h)
			if len(h.submit.places) != 1 {
				t.Fatalf("%s Place 호출 = %d, want 1", state, len(h.submit.places))
			}
			if !engine.A087LowerLimitAttemptedForTest(h.observer, "kr", "005930", h.clk.Now()) {
				t.Fatalf("%s 결과 뒤 첫 시도 금지 표시가 없음", state)
			}
		})
	}
}

func TestA087FallbackProtectivePlaceStillClosesFirstAttempt(t *testing.T) {
	h := newExitHarness(t, func(options *engine.ExitObserverOptions) {
		engine.SetA087LowerLimitForTest(options, "kr", "005930", "50400", exitNow)
	})
	h.submit.placeOutcome = &execgw.Outcome{State: journal.StateFailedConfirmed, Reason: execgw.ReasonBrokerRejected}
	a087BreachAt(h, 50399)
	if len(h.submit.places) != 1 || h.submit.places[0].Intent.Price != 50399 {
		t.Fatalf("fallback Place = %+v, want observed LIMIT 50399", h.submit.places)
	}
	if !engine.A087LowerLimitAttemptedForTest(h.observer, "kr", "005930", h.clk.Now()) {
		t.Fatal("하한가가 아닌 첫 보호 매도 뒤 첫 시도 금지 표시가 없음")
	}
}

func a087Breach(h *exitHarness) {
	a087BreachAt(h, 67900)
}

func a087BreachAt(h *exitHarness, observed float64) {
	h.t.Helper()
	h.entry("005930", "10", "70000", "68000", "70000")
	h.quote("005930", 70100)
	h.observe()
	h.quote("005930", observed)
	h.clk.Advance(5 * time.Second)
	h.observe()
}

func a087Wire(t *testing.T, intent orderintent.PlaceIntent) string {
	t.Helper()
	wire, err := execgw.PlaceWireBody(intent)
	if err != nil {
		t.Fatalf("PlaceWireBody: %v", err)
	}
	return wire
}
