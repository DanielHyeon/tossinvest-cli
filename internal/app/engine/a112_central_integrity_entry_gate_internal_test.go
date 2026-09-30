package engine

// a112 5.6.2.1 — 생산 배선: 생산이 Run 에 넣는 감독자(`NewRefreshingPairedStrategyEntrySupervisor`)는 엔진의 진입 게이트
// **바로 그것**을 받는다(다른 게이트 · 인라인 새 게이트가 아님 — 역할 확인). 게이트가 없는 Context 에서는 감독자를 만들지
// 않는다: 그 조립에서는 중앙 무결성 고장이 진입이 아니라 프로세스를 닫게 되므로(fallback), 생산이 그 길로 떨어지지 않게
// 생성자에서 거절함. 생산 기동 순서에서 이 생성자 앞의 `Recovery` 가 이미 같은 게이트를 요구한다(runtime_wiring.go).

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestTheProductionStrategySupervisorBlocksOnTheEnginesOwnEntryGate(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	c := &Context{Entry: gate}
	s, err := c.NewRefreshingPairedStrategyEntrySupervisor(clk)
	if err != nil {
		t.Fatalf("NewRefreshingPairedStrategyEntrySupervisor: %v", err)
	}
	if got, ok := s.entry.(*execgw.EntryGate); !ok || got != gate {
		t.Fatalf("supervisor entry blocker = %#v, want the engine's own entry gate %p", s.entry, gate)
	}
}

func TestTheProductionStrategySupervisorRefusesAContextWithoutAnEntryGate(t *testing.T) {
	clk := clock.NewFake(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	if _, err := (&Context{}).NewRefreshingPairedStrategyEntrySupervisor(clk); err == nil ||
		!errors.Is(err, ErrRuntimeUnavailable) {
		t.Fatalf("err = %v, want runtime unavailable — without a gate a central fault would close the process, not entry", err)
	}
}
