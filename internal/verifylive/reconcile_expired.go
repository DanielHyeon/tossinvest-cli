package verifylive

// reconcile_expired.go 는 a164(verify-reconcile-closes-broker-expired-conditional) 만료 대사 판정이다.
//
// 부재 대사(a121)와 같은 사전 검사·계좌 결속·승인·두 번 읽기·행 검사·잠금·추가 경로를 쓰고(Reconcile 안의 p.Expired 분기),
// 판정 하나만 바꿈: 대상 id 의 종결이 이번 실행의 by-id 상세 GET 두 번으로 EXPIRED·미발동이라고 **양성** 확인될 때만 닫음.
// 설계 정본: openspec/changes/a164-verify-reconcile-closes-broker-expired-conditional/design.md (D1~D10).

// ReconcileModeExpired 는 승인 화면·질문이 이름하는 만료 대사 모드임(design D1).
const ReconcileModeExpired = "expired"

// StepReconcileExpired 는 만료 대사 줄의 StepID 임 — Steps() 카탈로그 ∪ {cleanup, abort, reconcile-absent} 밖(design D5).
const StepReconcileExpired StepID = "reconcile-expired"

// ReconcileExpiredBasisDomain 은 만료 대사 근거 지문의 버전·도메인 태그임(design D5).
const ReconcileExpiredBasisDomain = "a164/reconcile-expired-basis/v1"

// 만료 대사 줄의 관측 키(design D5) — 근거 지문은 부재 대사와 같은 키 ObservationReconcileBasis 에 실음.
const (
	ObservationReconcileExpiredByID       = "reconcile.expired.by_id"
	ObservationReconcileExpiredClosedRows = "reconcile.expired.closed_rows"
)

// ReconcileGroupConditionalByID 는 근거 multiset 에서 by-id statement 행의 그룹 이름임(design D5).
const ReconcileGroupConditionalByID = "conditional-by-id"

// reconcileExpiryLegVocabulary 는 만료 leg(first.status) 어휘임 — 골격: 비어 있음(모든 진술 거절).
var reconcileExpiryLegVocabulary = map[string]bool{}

// reconcileExpiryClosedVocabulary 는 만료 모드 CLOSED 행 status 허용 목록임(A6) — 골격: 비어 있음.
var reconcileExpiryClosedVocabulary = map[string]bool{}
