package engine

// a205: 청산 관측 주기 5s → 2s(사용자 지시 2026-11-27, KR · US 공통).
// 이 파일은 그 값과, 그 값에서 유도되는 하한가 갱신기 시한을 고정함(design D1 · D2).

import (
	"testing"
	"time"
)

// TestA205DefaultExitObservationIntervalIsTwoSeconds: 기본 상수와, 주기 옵션을 주지 않은 관측자의 Interval() 반환값이 둘 다 2초임을 고정함
// (생산 배선 cmd/tossctl/engine.go 의 ExitObserverOptions 에 Interval 없음). Run · Sleep 은 실행하지 않음 — Run 이 Interval() 로 자는 것은 AST 근거(claims C1).
func TestA205DefaultExitObservationIntervalIsTwoSeconds(t *testing.T) {
	if DefaultExitObservationInterval != 2*time.Second {
		t.Fatalf("DefaultExitObservationInterval = %s, want 2s (a205)", DefaultExitObservationInterval)
	}
	// 옵션 0 이면 상수로 떨어지는 갈래(Interval B1 거짓)를 잼 — 시험이 주기를 넘기는 갈래만 쓰면 이 갈래는 안 돎.
	if got := (&ExitObserver{}).Interval(); got != 2*time.Second {
		t.Fatalf("an observer given no interval runs every %s, want 2s", got)
	}
}

// TestA205LowerLimitRefreshBudgetAtTheDefaultInterval: 기본 주기에서 하한가 갱신 한 번의 예산 값을 고정함
// (읽기 1초 · 사전조회 1초, 합 = 주기 — 5초 때 2초 · 3초와 같은 구조, design D2 (가)).
// 이 시험은 예산 값만 잼 — 갱신기를 실제로 돌리지 않고, 시간상 비중첩을 증명하지 않음(시한은 토큰 수신 뒤 시작함 — 구현 리뷰 r1 #2).
func TestA205LowerLimitRefreshBudgetAtTheDefaultInterval(t *testing.T) {
	interval := DefaultExitObservationInterval
	read := lowerLimitReadTimeout(interval)
	if read != time.Second {
		t.Fatalf("lower-limit read deadline at the default interval = %s, want 1s (min(2s, interval/2))", read)
	}
	// refresh 의 사전조회 몫 = 주기 − 읽기 몫. 0 이하면 refresh B1 이 읽기 몫으로 대체해 예산이 2×읽기가 되어 주기를 넘음.
	if prefetch := interval - read; prefetch != time.Second {
		t.Fatalf("prefetch share at the default interval = %s, want 1s (interval − read)", prefetch)
	}
}
