package engine

// a208 — 편입 전용 보호선 정책(engine.adoption.floor_only_exit, 기본 거짓).
//
// 켜져 있으면 새 자동 편입은 공통 정책과 무관하게 RUNNER 편입 사본을 기록함 → exitpolicy.CommonLadderForPosition 의
// 편입 RUNNER 갈래(부분익절 0, 정체성 1.0.0-adopted.1)로 판정됨. 손절(합성 손절가)과 판정 경로는 그대로이고,
// 정책은 편입 기록에 스냅숏되므로 토글을 뒤집어도 이미 편입된 포지션은 바뀌지 않음. 갈래를 adoptOne 밖 이 파일에 두어
// adoptOne 에 새 분기를 만들지 않음(design 3.1).

import "github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"

// adoptionExitPolicyID 는 편입 기록에 적을 정책 ID 를 고름. 거짓이면 a208 이전과 같은 공통 정책(§0.2).
func adoptionExitPolicyID(opts ReconcileDriverOptions) string {
	if opts.Adoption.FloorOnlyExit {
		return exitpolicy.CommonLadderRunner
	}
	return opts.CommonPolicy
}

// adoptionAlertBody 는 편입 알림 본문을 고름. 거짓이면 a208 이전 문구 바이트 그대로(「보호선 승격,」 뒤 공백 없음까지 원문).
func adoptionAlertBody(floorOnly bool) string {
	applied := "지금부터 보호선 승격,중간 익절, 손절이 엔진이 직접 연 포지션과 동일하게 적용된다.\n"
	if floorOnly {
		applied = "지금부터 보호선 승격과 손절만 적용되고 중간 익절은 하지 않는다(편입 전용 보호선 정책).\n"
	}
	return "계좌가 보유 중인데 이를 설명하는 진입 결정이 없어 편입 기록을 남겼다. " + applied +
		"기준선은 매입 원가가 아니라 **편입 시점에 관측된 가격**으로 잡는다 — 오래 들고 있던 " +
		"수익 종목은 지금부터 보호되고, 과거 수익은 R 척도에 포함되지 않는다."
}
