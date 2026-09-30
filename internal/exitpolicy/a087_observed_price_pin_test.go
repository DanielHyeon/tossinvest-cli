package exitpolicy_test

// a087_observed_price_pin_test.go 는 청산 주문 가격 사다리의 폴백 두 단이 **왜 닫혀 있는지**를 고정함
// (a087 issues.md I-P1, design D2a 「반증」).
//
// engine 의 ExitObserver.sellIntent 는 관측가가 비면 기준선으로 폴백하고(B1), 둘 다 비면 제출을 거부함(B2).
// 그 관측가의 유일한 생산 출처는 EvaluateRatchet · EvaluateLadder 가 받아 들인 값이고, 두 평가기는 성공 반환 전에
// 관측가를 positive() 로 검사함. 그래서 B1 · B2 는 생산에서 도달 불가이며, 그 등식은 이 가드 하나에 기대고 있음.
// 이 시험은 그 가드를 못 박음 — 가드가 빠지면 B1 · B2 가 다시 열린 문이 된다는 사실이 여기서 드러남.
//
// 빈 문자열 · 공백 · "0" · 파싱 불가 값을 두 평가기 모두에 넣어 "observed price" 필드 거부를 단언함.
// ratchet 쪽 "-1" 과 "1e4" 는 TestUnusableInputsAreRefused 가 이미 단언하므로 여기서 반복하지 않음.

import (
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// 평가기가 받아 들이면 sellIntent 의 B1(공백 → 기준선 폴백) · B2(가격 없음 거부)로 이어질 수 있는 관측가 모양들.
var a087UnpricedObservations = []struct {
	name  string
	value string
}{
	{"empty", ""},
	{"whitespace only", "   "},
	{"zero", "0"},
	{"unparseable", "not-a-price"},
}

// assertObservedPriceRefusal 은 오류가 관측가 필드를 이름으로 대는 거부인지 확인함.
func assertObservedPriceRefusal(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("an unpriced observation was evaluated; sellIntent's fallback and no-price refusal would become reachable")
	}
	if !errors.Is(err, exitpolicy.ErrRefused) {
		t.Errorf("err = %v, want it to match ErrRefused", err)
	}
	var refusal *exitpolicy.RefusalError
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %v, want a *RefusalError naming the field", err)
	}
	if refusal.Field != "observed price" {
		t.Errorf("field = %q, want %q — the refusal must come from the observed-price guard", refusal.Field, "observed price")
	}
}

// TestA087RatchetRefusesAnUnpricedObservation 은 ratchet 평가기의 관측가 가드를 고정함.
func TestA087RatchetRefusesAnUnpricedObservation(t *testing.T) {
	t.Parallel()
	for _, tc := range a087UnpricedObservations {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := input("10100")
			in.ObservedPrice = tc.value
			_, err := exitpolicy.EvaluateRatchet(in)
			assertObservedPriceRefusal(t, err)
		})
	}
}

// TestA087LadderRefusesAnUnpricedObservation 은 ladder 평가기의 관측가 가드를 고정함.
func TestA087LadderRefusesAnUnpricedObservation(t *testing.T) {
	t.Parallel()
	for _, tc := range a087UnpricedObservations {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			in := ladderInput("10100")
			in.ObservedPrice = tc.value
			_, err := exitpolicy.EvaluateLadder(in)
			assertObservedPriceRefusal(t, err)
		})
	}
}
