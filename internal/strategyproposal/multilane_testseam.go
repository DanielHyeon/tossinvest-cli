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
