package engine

import "errors"

// strategyScheduleStillMatchesAdmission 은 transport 직전 최종 검사가 쓰는 drift 판정이다(a112 6.3 — 「transport 전 digest/version drift 거절」).
//
// 방금 다시 수집한 일정 권한(fresh)이 dispatch admission 때 쓴 것(expected)과 **모든 축에서 같아야** 통과한다: 준비 · 서명 활성화 존재(양쪽) ·
// desired revision · 달력 버전 · 활성화 매니페스트 digest · 활성화 세대 · 활성화 만료. 하나라도 다르면 사람이 서명한 일정이 바뀐 것이므로
// 브로커 바이트 전에 멈춘다(게이트웨이가 이 오류에서 SUBMITTING 을 넘기지 않는다).
//
// 왜 순수 함수인가: 이 비교는 생산 조립(`NewPairedStrategyEntryProductionAssembly`)의 재검증 클로저 안에 있었고, 그 클로저는 설정 디렉터리 ·
// 환경 · 공식 클라이언트를 읽으므로 축마다 시험할 수 없었다. 수집(I/O)과 판정을 갈라 판정만 여기로 **의미 무변경 이동**했다 — 조건식 철자는
// 이동 전과 같다(영수증 `analysis/measurements/lot-6.3/move-receipt.txt`, 양쪽을 못 박는 시험 `TestTheScheduleDriftJudgementMovedVerbatim`).
func strategyScheduleStillMatchesAdmission(fresh, expected strategyScheduleMarketAuthority) error {
	if !fresh.snapshot.Ready || fresh.restore.Activation == nil || expected.restore.Activation == nil ||
		fresh.desired.Revision != expected.desired.Revision || fresh.calendar.Version != expected.calendar.Version ||
		fresh.snapshot.ActivationManifestDigest != expected.snapshot.ActivationManifestDigest ||
		fresh.restore.Activation.Generation() != expected.restore.Activation.Generation() ||
		!fresh.restore.Activation.ExpiresAt().Equal(expected.restore.Activation.ExpiresAt()) {
		return errors.New("engine: signed scheduler activation no longer matches dispatch admission")
	}
	return nil
}
