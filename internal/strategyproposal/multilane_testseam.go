//go:build tossos_testseams

package strategyproposal

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"

// ProductionBatchAuthorityMultiLaneForTest 는 한 종목이 여러 가족 제안을 동시에
// 낸 상태를 만든다. 담는 열쇠는 생산 경로가 쓰는 바로 그 함수로 만든다 —
// 종목만으로 담으면 나중 제안이 앞의 것을 조용히 덮어써서, 여러 가족이 있는
// 상태를 만들려던 픽스처가 오히려 하나짜리 상태를 만든다.
func ProductionBatchAuthorityMultiLaneForTest(manifestDigest string, values map[string][]strategyflow.Result) ProductionBatchAuthority {
	sealed := make(map[string]ProductionAuthority, len(values))
	for symbol, proposals := range values {
		for _, proposal := range proposals {
			if !proposal.ValidProposal() {
				return ProductionBatchAuthority{}
			}
			sealed[batchKey(symbol, proposal.Lineage.LaneID)] = ProductionAuthority{proposal: proposal,
				snapshotID: "snapshot-" + symbol, snapshotDigest: proposal.Lineage.LaneEvidenceDigest}
		}
	}
	return ProductionBatchAuthority{values: sealed, manifestDigest: manifestDigest}
}

// ProductionBatchAuthorityWithFaultForTest 는 스코프 하나가 제안을 **잃은** 배치를
// 만든다. 생산 경로에서는 증거 재생 실패 같은 일이 그 상태를 만든다.
//
// 잃은 것을 시늉하려고 그냥 값을 빼면 안 된다 — 뺀 배치는 "원래 없던 종목"과
// 구별되지 않아서, 시험이 겨누는 바로 그 혼동을 시험 픽스처가 다시 저지른다.
func ProductionBatchAuthorityWithFaultForTest(manifestDigest string, values map[string][]strategyflow.Result,
	absence ProductionAbsence,
) ProductionBatchAuthority {
	batch := ProductionBatchAuthorityMultiLaneForTest(manifestDigest, values)
	batch.absence, batch.faulted = absence, true
	return batch
}
