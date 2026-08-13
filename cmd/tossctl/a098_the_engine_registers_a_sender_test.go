package main

// a098 — 프로덕션 조립이 배달 실행자를 **실제로** 등록한다.
//
// 이 파일이 없으면 a098 은 자기가 고치려는 결함을 한 단계 위에서 반복한다:
// `Notifier.Flush` 가 프로덕션 호출자 0 이라 안 도는 것처럼, `alertDeliverer` 도
// 조립이 안 넘기면 안 돈다. **경로의 존재는 배달이 아니다.**

import (
	"reflect"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// TestProductionRuntimeStartsExactlyOneAlertDeliverer pins the auxiliary set.
//
// 집합 동일성으로 본다 — 부분 일치는 둘째가 늘어도 초록이다.
func TestProductionRuntimeStartsExactlyOneAlertDeliverer(t *testing.T) {
	runtime, err := engineRuntime(t.Context(), runtimeBranchContext(), clock.NewFake(runtimeBranchNow), nil, nil)
	if err != nil {
		t.Fatalf("engineRuntime: %v", err)
	}
	want := []string{"alert-delivery"}
	if got := runtime.AuxiliaryNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("production auxiliary names=%v want=%v — 밀린 알림을 보내는 주체가 없다", got, want)
	}
}

// TestTheAlertDelivererIsNotASupervisedLoop is the other half, and it is a
// separate assertion because the two can fail independently: registering it in
// both places would pass the test above while making an alert fault stop the
// engine (engine-safety Scenario 「프로덕션 감독 루프 집합은 넷이다」의 마지막 조항).
func TestTheAlertDelivererIsNotASupervisedLoop(t *testing.T) {
	runtime, err := engineRuntime(t.Context(), runtimeBranchContext(), clock.NewFake(runtimeBranchNow), nil, nil)
	if err != nil {
		t.Fatalf("engineRuntime: %v", err)
	}
	want := []string{"reconcile", "exit", "filldetect", engine.StrategyEntryLoopName}
	if got := runtime.LoopNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("production loop names=%v want=%v", got, want)
	}
	for _, name := range runtime.LoopNames() {
		if name == "alert-delivery" {
			t.Fatal("the delivery executor is a supervised loop; 알림 결함이 엔진을 내린다")
		}
	}
}
