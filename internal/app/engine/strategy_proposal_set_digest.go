package engine

import (
	"crypto/sha256"
	"encoding/hex"
)

// strategyProposalSetDigest 는 조립(strategyProposalAuthorityLoader.collectMarket)이 중재 결과에 적는 제안 집합 digest 의 유일한
// 식임(a112 5.2.2.2 — 조립도 이 함수를 부름): 조정자 순서대로 (승인 후보 종목, 계보 identity). A-lite 계약(dispatchHandoffs)이
// 같은 함수로 대조하므로 두 자리가 갈라질 수 없음 — a112_first_leg_owner_scope_seal_test.go 가 실제 조립의 digest 로도 잰다.
func strategyProposalSetDigest(entries []strategyProposalEntryAuthority) string {
	h := sha256.New()
	for _, entry := range entries {
		_, _ = h.Write([]byte(entry.route.approved.Symbol() + "\x00" + entry.authority.Proposal().Lineage.Identity + "\x00"))
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
