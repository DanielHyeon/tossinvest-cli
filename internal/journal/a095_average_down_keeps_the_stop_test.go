package journal

import (
	"context"
	"testing"
)

// a095 tasks 5.1 — 평단이 내려간 포지션에서 자동 경로의 유효 손절가가 내려가지 않음을 고정함(exit-policy 델타
// 「손절가는 평균단가 변화를 따라 내려가서는 안 된다」).
//
// 물타기(추가 매수)가 원장에 닿는 길은 대사 수렴의 조정 하나임(ApplyPositionAdjustment — 투영 수량 · 평단만 옮김).
// 이 시험은 그 조정이 exit state 의 네 기준 열(진입가 · 최초 손절 · 최초 위험 · 기준선)을 건드리지 않음을 행동으로
// 확인함. 평가기 입력 타입에 평단 필드가 없다는 것은 exitpolicy 의 TestA095TheEvaluatorsTakeNoAveragePrice(필드 전수 · 존재 검사),
// exit 관측이 진입가 자리에 exit state 의 값을 넣는다는 배선은 engine 의 기존 시험(TestTheCostBasisDoesNotChangeTheFirstJudgement 등)이 막음.
//
// 운영자 재편입 reset(resetExitStateForReadoptTx — positionpolicy.ActionReadopt)은 이 요구 밖임 — 이전 기준선과
// 비교하지 않고 새 관측의 합성 손절로 다시 세우는 사람 행위이며, 그 하향의 승인 · audit 여부는 issues.md I6 소관.
func TestA095AnAverageDownLeavesTheStopWhereItWas(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	insertDecision(t, j, "decision-a095", "nonce-a095")
	insertPosition(t, j, "p-avgdown", "decision-a095") // 10주 · 평단 70000
	insertExitState(t, j, "p-avgdown")                 // 진입 70000 · 손절 68000 · 위험 2000 · 기준선 68000

	before, err := j.ExitState(ctx, "p-avgdown")
	if err != nil {
		t.Fatal(err)
	}

	// 60000 에 10주를 더 사서 평단이 65000 으로 내려간 계좌를 대사가 수렴시킴. 평단 기준 손절을 다시 계산하면
	// 65000 − 2000 = 63000 으로 내려가는 자리임.
	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	result, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentExternal,
		ExpectedPrevQuantity: "10", ExpectedFillWatermark: watermark,
		NewQuantity: "20", NewAvgPrice: "65000",
		BrokerAsOf: "2026-09-30T01:00:00Z", Evidence: "the account holds 20 at an average of 65000",
	})
	if err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
	// 시나리오가 실제로 일어났는지 먼저 확인함 — 평단이 안 움직였다면 아래 단언은 아무것도 재지 않음.
	if result.Position.AvgPrice != "65000" || result.Position.Quantity != "20" {
		t.Fatalf("projection = (%s @ %s), want (20 @ 65000): the average-down did not reach the ledger",
			result.Position.Quantity, result.Position.AvgPrice)
	}

	after, err := j.ExitState(ctx, "p-avgdown")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ name, before, after string }{
		{"entry_price", before.EntryPrice, after.EntryPrice},
		{"initial_stop", before.InitialStop, after.InitialStop},
		{"initial_risk", before.InitialRisk, after.InitialRisk},
		{"baseline_price", before.Baseline, after.Baseline},
	} {
		if c.before != c.after {
			t.Errorf("%s moved %s → %s on an average-down; the stop must not follow the average down "+
				"(§0.6 — 보수 방향만)", c.name, c.before, c.after)
		}
	}
}
