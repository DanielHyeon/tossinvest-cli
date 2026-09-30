package execgw_test

// a112 5.6.2.1 — 전략 중앙 무결성 사유(`strategy_central_integrity`)의 자리: 진입 점검이 여러 사유 중 무엇을 먼저 말하는지는
// 운영자 런북과의 계약이다(latchOrder 머리말). 새 사유는 뒤에 붙이되 운영 모드보다 앞(원인이 결과보다 먼저), 송신자 정지보다
// 뒤(기존 쌍의 상대 순서 불변).

import (
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestTheStrategyCentralIntegrityLatchReadsBeforeTheModeAndAfterSenderDown(t *testing.T) {
	gate := func() *execgw.EntryGate {
		return execgw.NewEntryGate(clock.NewFake(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)),
			map[execgw.RequiredQuery]time.Duration{})
	}
	g := gate()
	g.Block(execgw.ReasonOperatingModeBlocked, "ENTRY_BLOCKED")
	g.Block(execgw.ReasonStrategyCentralIntegrity, "strategy central")
	if r := g.CheckEntry(); r == nil || r.Reason != execgw.ReasonStrategyCentralIntegrity {
		t.Errorf("with the mode also latched CheckEntry = %v, want the strategy central integrity cause first", r)
	}
	g = gate()
	g.Block(execgw.ReasonStrategyCentralIntegrity, "strategy central")
	g.Block(execgw.ReasonAlertSenderDown, "sender down")
	if r := g.CheckEntry(); r == nil || r.Reason != execgw.ReasonAlertSenderDown {
		t.Errorf("with sender-down also latched CheckEntry = %v, want sender-down first (existing precedence unchanged)", r)
	}
}
