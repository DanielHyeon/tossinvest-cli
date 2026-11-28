package main

// a193 — 프로덕션 조립이 세션 중 IN_DOUBT 해소 실행자를 **보조 실행자로** 실제로 등록한다.
// 감독 루프로 들어가면 그 정지가 엔진(손절 루프)을 내리므로, 보조 집합에 있고 감독 집합에 없음을 같이 봄.

import (
	"slices"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestA193TheProductionRuntimeStartsTheSettlerAsAuxiliary(t *testing.T) {
	runtime, err := engineRuntime(t.Context(), runtimeBranchContext(), clock.NewFake(runtimeBranchNow), nil, nil)
	if err != nil {
		t.Fatalf("engineRuntime: %v", err)
	}
	if !slices.Contains(runtime.AuxiliaryNames(), "in-doubt-settle") {
		t.Fatalf("auxiliary executors %v do not include the in-doubt settler", runtime.AuxiliaryNames())
	}
	if slices.Contains(runtime.LoopNames(), "in-doubt-settle") {
		t.Fatalf("the in-doubt settler is a supervised loop %v — its stop would bring the stop-loss loop down", runtime.LoopNames())
	}
}
