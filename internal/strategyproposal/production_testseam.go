//go:build tossos_testseams

package strategyproposal

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"

func ProductionBatchAuthorityForTest(manifestDigest string, values map[string]strategyflow.Result) ProductionBatchAuthority {
	sealed := make(map[string]ProductionAuthority, len(values))
	for symbol, proposal := range values {
		if !proposal.ValidProposal() {
			return ProductionBatchAuthority{}
		}
		// 담는 열쇠는 생산 경로와 같은 함수로 만든다. 종목만으로 담으면
		// 한 종목의 여러 가족 제안이 서로를 덮어쓰고, For 가 못 찾는다.
		sealed[batchKey(symbol, proposal.Lineage.LaneID)] = ProductionAuthority{proposal: proposal, snapshotID: "snapshot-" + symbol, snapshotDigest: proposal.Lineage.LaneEvidenceDigest}
	}
	return ProductionBatchAuthority{values: sealed, manifestDigest: manifestDigest}
}
