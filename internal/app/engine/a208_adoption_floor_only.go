package engine

// a208 — 편입 전용 보호선 정책(engine.adoption.floor_only_exit, 기본 거짓).
//
// 켜져 있으면 새 자동 편입은 설정된 공통 정책의 편입 보호선 사본을 기록함(F3 사용자 결정 「F3 2」) — 트리거 · 보호선 ·
// 러너 트레일은 공통 정책과 같고 부분익절만 0. 사본이 없는 공통 정책(BALANCED · 없음)이면 토글을 무시함(= OFF, fail-closed).
// 손절(합성 손절가)과 판정 경로는 그대로이고, 정책은 편입 기록에 스냅숏되므로 토글을 뒤집어도 이미 편입된 포지션은 바뀌지 않음.
// 갈래를 adoptOne 밖 이 파일에 두어 adoptOne 에 새 분기를 만들지 않음(design 3.1).

import (
	"strconv"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// adoptionFloorOnlyPolicy 는 토글이 참이고 공통 정책에 사본이 있을 때만 사본 ID 와 참을 돌려줌.
func adoptionFloorOnlyPolicy(opts ReconcileDriverOptions) (string, bool) {
	if !opts.Adoption.FloorOnlyExit {
		return "", false
	}
	return exitpolicy.AdoptedFloorOnlyPolicyID(opts.CommonPolicy)
}

// adoptionExitPolicyID 는 편입 기록에 적을 정책 ID 를 고름. 사본이 적용되지 않으면 a208 이전과 같은 공통 정책(§0.2).
func adoptionExitPolicyID(opts ReconcileDriverOptions) string {
	if id, ok := adoptionFloorOnlyPolicy(opts); ok {
		return id
	}
	return opts.CommonPolicy
}

// adoptionAlertBody 는 편입 알림 본문을 고름. 사본이 적용되지 않으면 a208 이전 문구 바이트 그대로(「보호선 승격,」 뒤 공백 없음까지 원문).
func adoptionAlertBody(opts ReconcileDriverOptions) string {
	applied := "지금부터 보호선 승격,중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다.\n"
	if _, ok := adoptionFloorOnlyPolicy(opts); ok {
		applied = "지금부터 보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다(편입 전용 보호선 정책).\n"
	}
	return "계좌가 보유 중인데 이를 설명하는 진입 결정이 없어 편입 기록을 남겼다. " + applied +
		"기준선은 매입 원가가 아니라 **편입 시점에 관측된 가격**으로 잡는다 — 오래 들고 있던 " +
		"수익 종목은 지금부터 보호되고, 과거 수익은 R 척도에 포함되지 않는다."
}

// floorOnlyAuditValue 는 기동 audit 행 값임. 무시되면 「ignored」 — 값이 달라야 공통 정책만 바뀐 기동도
// RecordChange 의 같은 값 중복 제거에 묻히지 않음(사유는 detail).
func floorOnlyAuditValue(adoption config.Adoption) string {
	if adoption.FloorOnlyIgnored != "" {
		return "ignored"
	}
	return strconv.FormatBool(adoption.FloorOnlyExit)
}
