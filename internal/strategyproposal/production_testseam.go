//go:build tossos_testseams

package strategyproposal

import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"

func ProductionBatchAuthorityForTest(manifestDigest string, values map[string]strategyflow.Result) ProductionBatchAuthority {
	sealed := make(map[string]ProductionAuthority, len(values))
	for symbol, proposal := range values {
		if !proposal.ValidProposal() {
			return ProductionBatchAuthority{}
		}
		sealed[symbol] = ProductionAuthority{proposal: proposal, snapshotID: "snapshot-" + symbol, snapshotDigest: proposal.Lineage.LaneEvidenceDigest}
	}
	return ProductionBatchAuthority{values: sealed, manifestDigest: manifestDigest}
}
