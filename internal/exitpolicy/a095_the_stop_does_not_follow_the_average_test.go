package exitpolicy_test

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// a095 tasks 5.1 · 5.2 — 손절 평가기는 평단을 읽지 않고, 사다리의 선은 진입가 · 관측 워터마크 · 이전 기준선에서만 나옴.
//
// 5.1 의 구조 절반: 자동 경로(판정 · 관측 갱신 · 복구)는 전부 이 두 평가기의 입력으로 손절을 계산함. 입력 타입에 평단 · 원가
// 필드가 없으면 평단 하락이 손절을 끌어내릴 길이 평가기 안에는 없음. 필드가 새로 생기면 이 시험이 먼저 깨져 그 편집이
// exit-policy 델타(평단 하락 비하향)와 대조되게 함.
var averageLike = regexp.MustCompile(`(?i)avg|average|cost|basis`)

func TestA095TheEvaluatorsTakeNoAveragePrice(t *testing.T) {
	for _, typ := range []reflect.Type{
		reflect.TypeOf(exitpolicy.RatchetInput{}),
		reflect.TypeOf(exitpolicy.LadderInput{}),
	} {
		// 평가기 입력의 필드를 전수로 셈 — 이름 하나의 존재가 아니라 모집단 전체를 봄.
		if typ.NumField() == 0 {
			t.Fatalf("%s has no fields; the census would pass on an empty sample", typ)
		}
		for i := 0; i < typ.NumField(); i++ {
			if name := typ.Field(i).Name; averageLike.MatchString(name) {
				t.Errorf("%s.%s: an evaluator input carries an average/cost field — a stop computed from it "+
					"would follow an average-down (exit-policy: 평균단가 하락이 유효 손절을 낮추지 않는다)", typ, name)
			}
		}
	}
}

// 5.2 — EvaluateLadder 의 산출 무변화: rung 잠금가는 진입가에서, 이전 기준선은 최댓값 합성에 그대로.
// DefaultLadderPolicy 의 둘째 rung 은 목표 2.5% · 잠금 1.0% 임(ladder.go DefaultLadderPolicy).
func TestA095TheLadderLinesComeFromTheEntry(t *testing.T) {
	base := func(entry, observed, baseline string) exitpolicy.LadderInput {
		return exitpolicy.LadderInput{
			EntryPrice: entry, ObservedPrice: observed, HighWater: observed, Baseline: baseline,
			Policy: exitpolicy.DefaultLadderPolicy(),
			State: exitpolicy.LadderState{PolicyID: "default_v1",
				ActivatedRung: exitpolicy.NoRung, PendingRung: exitpolicy.NoRung},
		}
	}
	eval := func(in exitpolicy.LadderInput) exitpolicy.LadderTransition {
		t.Helper()
		got, err := exitpolicy.EvaluateLadder(in)
		if err != nil {
			t.Fatalf("EvaluateLadder: %v", err)
		}
		return got
	}

	// 진입 10000, 관측 10300(+3%) → 둘째 rung(2.5%) 도달 → 잠금가 = 진입 × 1.01 = 10100.
	got := eval(base("10000", "10300", "9800"))
	if got.NextState.ActivatedRung != 1 || got.Baseline != "10100" {
		t.Errorf("rung/baseline = %d/%s, want 1/10100 (lock = entry × 1.01)", got.NextState.ActivatedRung, got.Baseline)
	}
	// 같은 수익률에서 진입가만 바꾸면 잠금가가 진입가를 따라 움직임 — 잠금가의 분모가 진입가라는 증거.
	got = eval(base("20000", "20600", "19600"))
	if got.Baseline != "20200" {
		t.Errorf("baseline = %s, want 20200 (lock = entry × 1.01)", got.Baseline)
	}
	// 이전 기준선이 잠금가보다 높으면 그대로 — 최댓값 합성이 기준선을 내리지 않음.
	got = eval(base("10000", "10300", "10250"))
	if got.Baseline != "10250" {
		t.Errorf("baseline = %s, want the previous 10250 kept (max composition)", got.Baseline)
	}
}
