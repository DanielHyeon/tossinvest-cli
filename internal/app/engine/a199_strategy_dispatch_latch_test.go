//go:build tossos_testseams

package engine

// a199 구현 리뷰 처분 1 — 청산 전용 기동에서 전략 진입을 막는 것은 `Verified` 가 아니라 **진입 래치 하나**다.
//
// 실제 런타임의 전략 worker(NewRefreshingPairedStrategyEntrySupervisor → productionStrategyCycle →
// runProductionStrategyMarketCycle → fresh.dispatch)는 `Automation.Verified` 를 읽지 않는다. dispatch 주기가 첫 레그를 내기
// 전에 묻는 것은 게이트웨이의 진입 관문(ObserveStrategyEntryGate → EntryGate.CheckEntryFor, 송신 직전에는
// withStrategyEntryGateAuthority → checkEntry)이고, 그 관문이 읽는 것이 a199 래치다. 그래서 여기서는 dispatch 주기의 진입 관문을
// **실제 게이트웨이**(래치를 단 EntryGate)로 돌리고 송신만 스파이로 센다.
//
// 대조군(양성): 같은 주기 · 같은 계약에 래치가 없으면 첫 레그가 송신된다 — 거절이 래치 때문임을 보인다.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// a199RealEntryGateGateway 는 진입 관문만 실제 게이트웨이에 맡기고 보호 관측 · 송신은 스파이 그대로 두는 dispatch Gateway 다.
type a199RealEntryGateGateway struct {
	*strategyDispatchGatewaySpy
	real *execgw.Gateway
}

func (gateway a199RealEntryGateGateway) ObserveStrategyEntryGate(ctx context.Context, market, symbol string) (execgw.StrategyEntryGateAuthority, error) {
	return gateway.real.ObserveStrategyEntryGate(ctx, market, symbol)
}

func TestA199TheStaleLatchStopsTheRealStrategyDispatchBeforeTheBroker(t *testing.T) {
	for _, tc := range []struct {
		name   string
		latch  bool
		placed int
	}{
		{"control: no stale latch", false, 1},
		{"exit-only start: stale latch", true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newA112TradingFixture(t, a112TradingOptions{})
			gate := execgw.NewEntryGate(clock.NewFake(fixture.now), map[execgw.RequiredQuery]time.Duration{})
			if tc.latch {
				gate.Block(execgw.ReasonCapabilityAttestationStale, "capability attestation is stale (expired)")
			}
			real, err := execgw.New(execgw.Options{Journal: fixture.journal, Trading: trading.NewService(config.Trading{}, nil),
				Clock: clock.NewFake(fixture.now), AccountRef: "acct-risk-loader", Source: "a199-r1", Entry: gate})
			if err != nil {
				t.Fatal(err)
			}
			fixture.cycle.gateway = a199RealEntryGateGateway{strategyDispatchGatewaySpy: fixture.spy, real: real}
			err = fixture.deliverKR(t)
			if got := len(fixture.placedSymbols()); got != tc.placed {
				t.Fatalf("strategy first legs sent=%d err=%v, want %d", got, err, tc.placed)
			}
			if tc.latch && (err == nil || !strings.Contains(err.Error(), string(execgw.ReasonCapabilityAttestationStale))) {
				t.Fatalf("err=%v, want the cycle refused by the stale-attestation latch", err)
			}
		})
	}
}
