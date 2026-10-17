//go:build tossos_testseams

package execgw

// HoldEntryGateLockForTest 는 전략 진입 dispatch 가 브로커 전송 동안 게이트 잠금(g.mu)을 쥐는 모양을
// 흉내 냄 (`strategy_entry_gate_authority.go` 의 withStrategyEntryGateAuthority). a124 tasks 2.6 (a) 가
// 「배달 실행자가 g.mu 를 기다리는 동안에도 손절 쪽 Notify 는 실행자를 기다리지 않는다」를 재는 데 씀.
// 돌려준 release 를 한 번 부르면 잠금을 놓음.
func HoldEntryGateLockForTest(g *EntryGate) (release func()) {
	g.mu.Lock()
	done := false
	return func() {
		if !done {
			done = true
			g.mu.Unlock()
		}
	}
}
