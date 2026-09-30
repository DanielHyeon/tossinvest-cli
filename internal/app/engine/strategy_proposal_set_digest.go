package engine

import (
	"crypto/sha256"
	"encoding/hex"
)

// strategyProposalSetDigest 는 조립(strategyProposalAuthorityLoader.collectMarket)이 중재 결과에 적는 제안 집합 digest 와 같은
// 식임: 조정자 순서대로 (승인 후보 종목, 계보 identity). 두 자리가 갈라지면 활성화 시장이 전부 닫힘(fail-closed) —
// a112_first_leg_owner_scope_seal_test.go 가 실제 조립의 digest 와 이 식이 같음을 잰다.
func strategyProposalSetDigest(entries []strategyProposalEntryAuthority) string {
	h := sha256.New()
	for _, entry := range entries {
		_, _ = h.Write([]byte(entry.route.approved.Symbol() + "\x00" + entry.authority.Proposal().Lineage.Identity + "\x00"))
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
