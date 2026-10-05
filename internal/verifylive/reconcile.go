package verifylive

// reconcile.go 는 a121(reconcile-stale-verification-artifacts) 대사 경로의 **골격**이다.
//
// RED 로트 산출물: 타입·상수·시그니처만 세운다. Reconcile 은 언제나 코드 없는 거절을 돌려주고(무동작),
// 기록·브로커 어느 쪽도 건드리지 않는다. 판정 구현은 GREEN 로트(tasks 3.1·3.2)의 몫이다.
//
// 설계 정본: openspec/changes/a121-reconcile-stale-verification-artifacts/design.md (G1·G2·G3, 로트 1 처분,
// freeze·codex 수리). 기존 파일은 편집하지 않는다 — Artifact.terminal() 의 셋째 종결은 GREEN 의 record.go 편집이다.

import (
	"context"
	"io"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

// KindReconcile 은 대사 줄의 종류임(design G2 「결정(골격)」).
const KindReconcile = "reconcile"

// StepReconcile 은 대사 줄의 StepID 임 — Steps() 카탈로그 ∪ {cleanup, abort} 밖의 고유 값이어야 함(design R1).
// StepID 만 비교하는 소비자 넷(LastEntry·heldAfter·m0ManualReconcileIDs·baselineSellable)이 이 줄을
// 단계 판정으로 오인하지 않게 하기 위함.
const StepReconcile StepID = "reconcile-absent"

// ReconcileBasisDomain 은 근거 지문의 버전·도메인 태그임(design G2 「근거 지문」).
const ReconcileBasisDomain = "a121/reconcile-basis/v1"

// ObservationReconcileBasis 는 근거 지문을 싣는 관측 키임 — 읽기 근거는 이 관측 하나에만 적음(Calls 비탑재).
const ObservationReconcileBasis = "reconcile.basis"

// 읽기 그룹 이름 — multiset (그룹, id, status, triggeredOrderId) 의 첫 성분.
const (
	ReconcileGroupConditionalOpen   = "conditional-open"
	ReconcileGroupPlainOpen         = "plain-open"
	ReconcileGroupConditionalClosed = "conditional-closed"
)

// ReconcileReader 는 대사 경로가 받는 유일한 브로커 의존임 — 공식 GET 읽기만 노출함(design G2 P1-1).
// 쓰기 메서드·Broker·*official.Client 를 노출하지 않음.
type ReconcileReader interface {
	ReconcileConditionalOrdersPage(ctx context.Context, status, symbol, cursor string, limit int) (official.ReconcileConditionalPage, error)
	ReconcileOpenOrdersPage(ctx context.Context, symbol, cursor string, limit int) (official.ReconcileOrderPage, error)
	ReconcileInstrument(ctx context.Context, symbol string) (string, error)
}

// ReconcileAccount 는 좁은 생성자가 자체 Accounts() 읽기로 확정한 계좌 사실임(design G3-2 P1-2·R2-2).
// Ref 는 원문 계좌 참조 — 마스킹 전에는 어떤 출력·기록에도 쓰지 않음.
type ReconcileAccount struct {
	Ref   string
	Seq   int
	Count int
}

// ReconcileApproval 은 목록 읽기 전 사람 승인에 보이는 내용임(design G3 F1 처분 ①, 승인 순서 freeze 재검 P2).
type ReconcileApproval struct {
	RecordAccountMask  string
	CurrentAccountMask string
	AccountCount       int
	Kind               string
	ID                 string
	Symbol             string
	Market             string
}

// WriteText 는 승인 내용을 운영자 화면에 씀. 골격: 아무것도 쓰지 않음.
func (a ReconcileApproval) WriteText(w io.Writer) {}

// ReconcileParams 는 대사 한 번의 입력임. 기록 경로는 cmd 가 --config-dir 에서 유도한 값이어야 함(G3-3).
type ReconcileParams struct {
	RecordPath string
	Market     string
	Account    ReconcileAccount
	Reader     ReconcileReader
	Out        io.Writer
	// Approve 는 승인 내용이 출력된 뒤, 주문·조건주문 목록 읽기 전에 불림. 오류면 거절.
	Approve func(context.Context, ReconcileApproval) error
	Now     func() time.Time
}

// ReconcileResult 는 추가된 대사 줄의 요약임.
type ReconcileResult struct {
	Artifact Artifact
	Basis    string
}

// ReconcileRefusalCode 는 거절한 가드의 이름임 — 가드마다 하나(시험이 공유 문구가 아닌 코드로 단언).
type ReconcileRefusalCode string

// 거절 코드. verifylive 판정 가드와 cmd 사전 검사 가드를 한 어휘로 둠.
const (
	// 기록(G2 F3·R2-3·동시성)
	RefuseRecordUnreadable    ReconcileRefusalCode = "record-unreadable"
	RefuseRecordUndecodable   ReconcileRefusalCode = "record-undecodable"
	RefuseRecordFormat        ReconcileRefusalCode = "record-format"
	RefuseRecordTailNoNewline ReconcileRefusalCode = "record-tail-no-newline"
	RefuseRecordChanged       ReconcileRefusalCode = "record-changed"
	// 후보 선택(G2)
	RefuseNoCandidate        ReconcileRefusalCode = "no-candidate"
	RefuseMultipleCandidates ReconcileRefusalCode = "multiple-candidates"
	RefuseM0Unresolved       ReconcileRefusalCode = "m0-unresolved"
	// 계좌(G3-1)
	RefuseAccountRefUnusable ReconcileRefusalCode = "account-ref-unusable"
	RefuseAccountMixed       ReconcileRefusalCode = "account-mixed"
	RefuseAccountMismatch    ReconcileRefusalCode = "account-mismatch"
	// 승인
	RefuseNotApproved ReconcileRefusalCode = "not-approved"
	// Q1(G1-4)·Q3(G1-6)
	RefuseRetentionUnmeasured ReconcileRefusalCode = "retention-unmeasured"
	RefuseRetentionEviction   ReconcileRefusalCode = "retention-eviction-model"
	RefuseCreatedAtZero       ReconcileRefusalCode = "created-at-zero"
	RefuseArtifactTooOld      ReconcileRefusalCode = "artifact-too-old"
	RefuseFreshnessUnfixed    ReconcileRefusalCode = "freshness-unfixed"
	RefuseFreshnessExceeded   ReconcileRefusalCode = "freshness-exceeded"
	// 양성 대조(P1-3)
	RefuseInstrumentControl ReconcileRefusalCode = "instrument-control"
	// 읽기·페이지(G1-6)
	RefuseReadError      ReconcileRefusalCode = "read-error"
	RefuseRepeatedCursor ReconcileRefusalCode = "repeated-cursor"
	RefuseEmptyCursor    ReconcileRefusalCode = "empty-cursor"
	RefusePageCap        ReconcileRefusalCode = "page-cap"
	// 행 검증(G1-5·F4·F6)
	RefuseRowIncomplete ReconcileRefusalCode = "row-incomplete"
	RefuseRowSymbol     ReconcileRefusalCode = "row-symbol"
	RefuseRowMarket     ReconcileRefusalCode = "row-market"
	RefuseOCO           ReconcileRefusalCode = "oco-row"
	RefuseDuplicateRow  ReconcileRefusalCode = "duplicate-row"
	RefuseReadsDiffer   ReconcileRefusalCode = "reads-differ"
	// 부재 판정(G1-1·2·3, Q6)
	RefuseOpenConditional     ReconcileRefusalCode = "open-conditional"
	RefuseOpenPlainOrder      ReconcileRefusalCode = "open-plain-order"
	RefuseClosedTargetExpired ReconcileRefusalCode = "closed-target-expired"
	RefuseClosedTargetPresent ReconcileRefusalCode = "closed-target-present"
	RefuseClosedFiredTrace    ReconcileRefusalCode = "closed-fired-trace"
	RefuseClosedStatus        ReconcileRefusalCode = "closed-status"
	// cmd 사전 검사(G3-2·G3-3·G3-4·동시성)
	RefuseNoConfigDir         ReconcileRefusalCode = "no-config-dir"
	RefuseEnvCredentials      ReconcileRefusalCode = "env-credentials"
	RefuseRecordOverride      ReconcileRefusalCode = "record-override"
	RefuseNoMarket            ReconcileRefusalCode = "no-market"
	RefuseNoCredentials       ReconcileRefusalCode = "no-credentials"
	RefuseExecutionLock       ReconcileRefusalCode = "execution-lock"
	RefuseRateBudget          ReconcileRefusalCode = "rate-budget"
	RefuseAccountsRead        ReconcileRefusalCode = "accounts-read"
	RefuseAccountCount        ReconcileRefusalCode = "account-count"
	RefuseAccountSeqZero      ReconcileRefusalCode = "account-seq-zero"
	RefuseAccountMalformedRow ReconcileRefusalCode = "account-malformed-row"
)

// ReconcileRefusal 은 대사 거절임. 거절은 기록에 아무것도 쓰지 않음.
type ReconcileRefusal struct {
	Code   ReconcileRefusalCode
	Detail string
}

func (r *ReconcileRefusal) Error() string {
	return "verify reconcile: refused (" + string(r.Code) + "): " + r.Detail
}

// Reconcile 은 기록이 소유한 stale 조건주문 하나를 공식 읽기로 대사함.
// 골격: 아무것도 읽지·쓰지 않고 코드 없는 거절을 돌려줌.
func Reconcile(ctx context.Context, p ReconcileParams) (ReconcileResult, error) {
	return ReconcileResult{}, &ReconcileRefusal{Detail: "a121: reconciliation is not implemented (RED skeleton)"}
}

// ReconcileBasisRow 는 근거 multiset 의 원소임 — (그룹, id, status, triggeredOrderId).
type ReconcileBasisRow struct {
	Group            string
	ID               string
	Status           string
	TriggeredOrderID string
}

// ReconcileBasisDigest 는 읽기 근거를 버전·도메인 태그와 함께 정규 순서로 지문화함. 골격: 빈 문자열.
func ReconcileBasisDigest(rows []ReconcileBasisRow) string { return "" }

// --- Q1(G1-4) 보존 한도 ----------------------------------------------------------

// evictionModel 은 Q1 측정이 밝힌 CLOSED 이력의 축출 모형임(codex F5).
type evictionModel string

const (
	evictionUnknown       evictionModel = ""
	evictionDuration      evictionModel = "duration"
	evictionCount         evictionModel = "count"
	evictionIndeterminate evictionModel = "indeterminate"
)

// retentionMeasurement 는 Q1 사람 실측의 결과임 — 기간 한도와 축출 모형.
type retentionMeasurement struct {
	Bound    time.Duration
	Eviction evictionModel
}

// reconcileRetention 은 Q1 보존 한도임. 생산 빌드에서는 **항상 nil**(측정 부재 = 거절, design G1-4 P0-1).
// 시험만 seam 으로 주입하고, 실값은 측정 뒤 별도 리뷰를 거친 상수 커밋으로만 들어옴. 설정·플래그·환경 변수 경로 금지.
var reconcileRetention *retentionMeasurement

// --- Q3(G1-6) 신선도 한도 --------------------------------------------------------

// reconcileFreshnessBoundValue 는 리뷰가 승인한 Q3 상수임(analysis/red-lot 메모 §1 — 정상 8 GET 은 1~2초,
// 429 한 번이면 넘도록 잡은 15초).
const reconcileFreshnessBoundValue = 15 * time.Second

// reconcileFreshnessBound 는 Q3 한도임 — 첫 읽기 시작부터 추가 승인 직전까지(codex F7). 영이면 미확정 = 거절.
// 생산 빌드에서는 위 상수로만 초기화되고 다른 대입이 없음(구조 시험이 핀) — 시험·seam 만 바꿈.
var reconcileFreshnessBound = reconcileFreshnessBoundValue

// --- F4 CLOSED status 허용 목록 --------------------------------------------------

// reconcileClosedStatusAllowlist 는 측정·골든으로 확정한 조건주문 종결 어휘임(codex F4).
// RED 로트 전사 결과: verify-execution-capability 영수증/골든에 조건주문 CLOSED status 측정이 없음(출처 부재) —
// 그래서 비워 둠(모든 CLOSED 행이 status 가드에서 거절되는 fail-closed). 지어내지 않음.
var reconcileClosedStatusAllowlist = map[string]bool{}
