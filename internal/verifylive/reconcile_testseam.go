//go:build tossos_testseams

package verifylive

import "time"

// SetReconcilePoliciesForTest 는 저장소 시험 seam 이진에서만 Q1 보존 측정·Q3 신선도 한도를 주입함.
//
// 생산 빌드(태그 없음)에는 이 파일이 없으므로 reconcileRetention 은 nil(design G1-4 P0-1 — 측정 상수 커밋 전까지,
// 설정·플래그·환경 변수 경로 금지), reconcileFreshnessBound 는 리뷰 승인 상수 reconcileFreshnessBoundValue(15초)로
// 남음. cmd 패키지 시험이 대사 전 경로를 끝까지 몰 때만 씀. eviction 은 "duration" · "count" · "indeterminate" · "" 중 하나.
func SetReconcilePoliciesForTest(retentionBound time.Duration, eviction string, freshness time.Duration) (restore func()) {
	prevRetention, prevFreshness := reconcileRetention, reconcileFreshnessBound
	reconcileRetention = &retentionMeasurement{Bound: retentionBound, Eviction: evictionModel(eviction)}
	reconcileFreshnessBound = freshness
	return func() {
		reconcileRetention, reconcileFreshnessBound = prevRetention, prevFreshness
	}
}
