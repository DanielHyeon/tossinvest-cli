package main

// label.go 는 「냈을 제안」의 결과를 같은 evidence 의 뒤 봉으로 잼 — 추가 GET 0(design §4 · D10).

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyevidence"

// labelResult 는 라벨 한 줄의 값.
type labelResult struct {
	Outcome    string // stop · target · close
	ExitMinor  uint64
	ExitOpenMS uint64 // 결과를 정한 봉의 여는 시각
	RPPM       int64
}

// labelProposal 은 fromIndex(제안 다음 봉)부터 순서대로 보며 손절과 목표 중 먼저 닿은 쪽을 고름. 한 봉에서 둘 다면 손절(보수).
// 끝까지 둘 다 없으면 final 일 때만 마지막 정규장 봉의 종가로 「close」 라벨을 냄(§2.1-12) — 아니면 아직 라벨 없음.
// 체결 가능성(호가 잔량)은 재지 않음 — 보고서의 한계.
func labelProposal(series strategyevidence.BarSeries, fromIndex int, plan sizingPlan, final bool) (labelResult, bool) {
	for i := fromIndex; i < len(series.Bars); i++ {
		p := series.Bars[i].Payload
		if p.LowMinor <= plan.Stop {
			return labelResult{Outcome: "stop", ExitMinor: plan.Stop, ExitOpenMS: p.OpenAtMS, RPPM: rMultiplePPM(plan.Stop, plan)}, true
		}
		if p.HighMinor >= plan.Target {
			return labelResult{Outcome: "target", ExitMinor: plan.Target, ExitOpenMS: p.OpenAtMS, RPPM: rMultiplePPM(plan.Target, plan)}, true
		}
	}
	if !final || len(series.Bars) == 0 {
		return labelResult{}, false
	}
	last := series.Bars[len(series.Bars)-1].Payload
	return labelResult{Outcome: "close", ExitMinor: last.CloseMinor, ExitOpenMS: last.OpenAtMS, RPPM: rMultiplePPM(last.CloseMinor, plan)}, true
}
