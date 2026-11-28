package config

// a208 F3 — 토글 무시 사유(기동 audit 행 detail).

import (
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// floorOnlyIgnoredReason 은 토글이 참인데 공통 정책에 편입 보호선 사본이 없을 때의 사유를 돌려줌(그 밖에는 빈 값).
// 판정 함수는 엔진과 같은 exitpolicy.AdoptedFloorOnlyPolicyID 하나 — 정의를 둘로 두지 않음.
func (a Adoption) floorOnlyIgnoredReason(commonPolicy string) string {
	if !a.FloorOnlyExit {
		return ""
	}
	if _, ok := exitpolicy.AdoptedFloorOnlyPolicyID(commonPolicy); ok {
		return ""
	}
	name := strings.TrimSpace(commonPolicy)
	if name == "" {
		name = "없음(RATCHET)"
	}
	return "공통 정책 " + name + " 에는 편입 보호선 사본이 없어 토글을 무시함 — 새 편입은 꺼져 있을 때와 같음"
}
