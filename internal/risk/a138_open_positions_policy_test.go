package risk

import (
	"strings"
	"testing"
)

// a138: 최대 보유 포지션 수는 정책의 일부 — 기본값 6(사용자 결정 2026-10-15), 1 미만은 정책 검증이 거부(fail-closed).
func TestA138DefaultPolicyCarriesTheOpenPositionMaximum(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxOpenPositions != 6 {
		t.Fatalf("default open position maximum = %d, want 6", p.MaxOpenPositions)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("the default policy must validate: %v", err)
	}
}

func TestA138PolicyValidationRefusesAnUnusableOpenPositionMaximum(t *testing.T) {
	for _, bad := range []int{0, -1} {
		p := DefaultPolicy()
		p.MaxOpenPositions = bad
		err := p.Validate()
		if err == nil {
			t.Errorf("an open position maximum of %d must be refused", bad)
			continue
		}
		if !strings.Contains(err.Error(), "open positions") {
			t.Errorf("the refusal must name the open position maximum, got %v", err)
		}
	}
}
