package verifylive

// reconcile.go 는 a121(reconcile-stale-verification-artifacts) 대사 경로다 — 판정 파이프라인과 그 타입.
//
// 공식 GET 읽기와 기록 추가 한 줄만 함. 브로커 변이 경로는 없음(봉인은 reconcile_seal_test.go 의 세 층이 핀).
// 설계 정본: openspec/changes/a121-reconcile-stale-verification-artifacts/design.md (G1·G2·G3, 로트 1 처분,
// freeze·codex 수리, RED 로트 처분). 판정 보조는 reconcile_check.go, 기록 읽기·쓰기는 reconcile_record.go.

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
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
	RefuseRecordLocked        ReconcileRefusalCode = "record-locked"
	// 후보 선택(G2)
	RefuseNoCandidate        ReconcileRefusalCode = "no-candidate"
	RefuseMultipleCandidates ReconcileRefusalCode = "multiple-candidates"
	RefuseM0Unresolved       ReconcileRefusalCode = "m0-unresolved"
	// 계좌(G3-1)
	RefuseAccountRefUnusable ReconcileRefusalCode = "account-ref-unusable"
	RefuseAccountMixed       ReconcileRefusalCode = "account-mixed"
	RefuseAccountMismatch    ReconcileRefusalCode = "account-mismatch"
	// 승인
	RefuseNotApproved     ReconcileRefusalCode = "not-approved"
	RefuseApprovalDisplay ReconcileRefusalCode = "approval-display"
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

// Reconcile 은 기록이 소유한 stale 조건주문 하나를 공식 읽기로 대사함 — 조건이 전부 참일 때만 대사 줄 하나를 추가함.
//
// 가드 순서는 design 「RED 로트 처분」 의 계약 그대로임(Q6 도달이 이 순서에 의존 — load-bearing):
// 사전 검사(기록 → 후보 → 계좌, 목록 읽기 전) → 정적 거절(Q1/Q3, A-GREEN P2-a 로 승인 앞) → 승인 → 종목 조회 → 두 번 읽기(페이지·중복 포함,
// 조건주문 OPEN → 일반 OPEN → CLOSED) → 종목 조회 → 신선도·Q1 재검 → multiset 비교 → 행 검사(OCO → 필드 결측 →
// 심볼 → 시장) → 부재 검사(OPEN 조건 → 일반 OPEN → CLOSED target(EXPIRED → Q6) → 발동 흔적) → status 허용 목록 →
// 추가 준비(기록 파일 flock → 그 fd 로 엄격 해독 → 지문 → 개행 → 줄 직렬화) → 쓰기 경계 직전 신선도·Q1 재검 → 한 번의 쓰기·동기화·read-back. 거절은 기록에 아무것도 쓰지 않음.
func Reconcile(ctx context.Context, p ReconcileParams) (ReconcileResult, error) {
	now := p.Now
	if now == nil {
		now = time.Now
	}
	if p.Reader == nil {
		return ReconcileResult{}, refuse(RefuseReadError, "reconcile: no read-only official reader was supplied")
	}
	if p.Approve == nil {
		return ReconcileResult{}, refuse(RefuseNotApproved, "reconcile: no human approval step was supplied")
	}
	out := p.Out
	if out == nil {
		out = io.Discard
	}

	// 1. 기록 사전 검사 — 엄격 해독·개행 꼬리(F3·R2-3). 지문은 판정한 바이트 그대로의 sha256(P2-7).
	raw, entries, err := readRecordStrict(p.RecordPath)
	if err != nil {
		return ReconcileResult{}, err
	}
	fingerprint := sha256.Sum256(raw)

	// 2. 후보 선택 — PendingCleanup 의 조건주문 중 M0 가 가리키지 않는 것, 정확히 하나(G2 「선택」).
	target, err := selectReconcileCandidate(entries)
	if err != nil {
		return ReconcileResult{}, err
	}

	// 3. 계좌 결속(G3-1) — mixed 가 mismatch 보다 먼저.
	recordMask, currentMask, err := bindReconcileAccount(entries, target, p.Account.Ref)
	if err != nil {
		return ReconcileResult{}, err
	}

	// 4. 정적 거절 — Q1·Q3 는 로컬 판정이라 사람 승인 **앞**에서 거절함(A-GREEN P2-a: y 를 누른 뒤에 측정 부재로
	// 거절당하는 모양 제거). 읽기 뒤의 재검(F7)은 그대로임.
	if err := checkReconcileRetention(target, now()); err != nil {
		return ReconcileResult{}, err
	}
	if reconcileFreshnessBound <= 0 {
		return ReconcileResult{}, refuse(RefuseFreshnessUnfixed, "the Q3 freshness bound is not fixed")
	}

	// 5. 사람 승인 — 두 마스크·계좌 수를 보인 뒤, 목록 읽기 전(F1 처분 ①, 승인 순서 freeze 재검 P2).
	approval := ReconcileApproval{
		RecordAccountMask: recordMask, CurrentAccountMask: currentMask, AccountCount: p.Account.Count,
		Kind: target.Kind, ID: target.ID, Symbol: target.Symbol, Market: p.Market,
	}
	if err := approval.WriteText(out); err != nil {
		return ReconcileResult{}, refuse(RefuseApprovalDisplay, "the approval could not be shown to the operator: "+err.Error())
	}
	if err := p.Approve(ctx, approval); err != nil {
		return ReconcileResult{}, refuse(RefuseNotApproved, "operator approval was not given: "+err.Error())
	}

	// 6. 양성 대조 ① — 조회 심볼은 artifact 줄의 바이트 그대로(호출자 입력 아님).
	symbol := target.Symbol
	if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {
		return ReconcileResult{}, err
	}

	// 7. 두 번 읽기 — 창의 시작은 첫 목록 읽기(승인 대기는 창 밖).
	start := now()
	first, err := readReconcileSet(ctx, p.Reader, symbol)
	if err != nil {
		return ReconcileResult{}, err
	}
	second, err := readReconcileSet(ctx, p.Reader, symbol)
	if err != nil {
		return ReconcileResult{}, err
	}

	// 8. 양성 대조 ② · 신선도·Q1 재검(F7).
	if err := checkReconcileInstrument(ctx, p.Reader, symbol); err != nil {
		return ReconcileResult{}, err
	}
	if err := checkReconcileWindow(target, start, now()); err != nil {
		return ReconcileResult{}, err
	}

	// 9. multiset 비교(P1-4).
	basisRows := first.basis()
	if !sameReconcileMultiset(basisRows, second.basis()) {
		return ReconcileResult{}, refuse(RefuseReadsDiffer, "the two reads disagree as a multiset of (group, id, status, triggeredOrderId)")
	}

	// 10. 행 검사 — 두 스냅숏 **각각**(codex CG-1: 비교 tuple 에 없는 second·심볼·시장이 둘째 읽기에서만 나타나도
	// 잡음) · 부재 검사 · status 허용 목록(마지막).
	if err := checkReconcileRows(first, symbol, p.Market); err != nil {
		return ReconcileResult{}, err
	}
	if err := checkReconcileRows(second, symbol, p.Market); err != nil {
		return ReconcileResult{}, err
	}
	if err := checkReconcileAbsence(first, target); err != nil {
		return ReconcileResult{}, err
	}
	if err := checkReconcileClosedStatus(first); err != nil {
		return ReconcileResult{}, err
	}

	// 11. 추가 준비 — 기록 파일 자체를 잠그고(codex CG-2: 같은 inode 의 다른 대사와 직렬화) 그 fd 로 다시 읽어
	// 엄격 해독 → 지문 → 개행. 판정한 바이트가 곧 쓸 파일임.
	lock, err := lockRecordForAppend(p.RecordPath)
	if err != nil {
		if errors.Is(err, errRecordLocked) {
			return ReconcileResult{}, refuse(RefuseRecordLocked, "another reconciliation holds this record file (the same file, possibly under another path)")
		}
		return ReconcileResult{}, refuse(RefuseRecordUnreadable, "the record cannot be opened for the append: "+err.Error())
	}
	defer lock.release()
	again, err := lock.contents()
	if err != nil {
		return ReconcileResult{}, refuse(RefuseRecordUnreadable, err.Error())
	}
	if _, err := decodeRecordStrict(again); err != nil {
		return ReconcileResult{}, err
	}
	if sha256.Sum256(again) != fingerprint {
		return ReconcileResult{}, refuse(RefuseRecordChanged, "the record changed between candidate selection and the append")
	}
	if len(again) > 0 && again[len(again)-1] != '\n' {
		return ReconcileResult{}, refuse(RefuseRecordTailNoNewline, "the record does not end with a newline")
	}

	// 12. 줄을 미리 직렬화하고(codex CG-5) 쓰기 경계 바로 앞에서 신선도·Q1 을 다시 잼 — 그 뒤에는 한 번의 쓰기뿐.
	basis := ReconcileBasisDigest(basisRows)
	at := now()
	reconciled := target
	reconciled.ReconciledAbsent = true
	reconciled.ReconciledAt = at.UTC()
	line, err := encodeReconcileLine(recordMask, reconciled, basis, start, at)
	if err != nil {
		return ReconcileResult{}, err
	}
	if err := checkReconcileWindow(target, start, at); err != nil {
		return ReconcileResult{}, err
	}

	// 13. 추가 — 한 번의 Write(개행 포함) → fsync → 다시 읽어 대조(codex CG-3). 실패는 침묵하지 않음.
	if err := lock.appendLine(line, int64(len(again))); err != nil {
		if errors.Is(err, errRecordSizeMoved) {
			return ReconcileResult{}, refuse(RefuseRecordChanged, "the record grew between the locked read and the append — another writer interleaved")
		}
		return ReconcileResult{}, fmt.Errorf("verify reconcile: %w", err)
	}
	return ReconcileResult{Artifact: reconciled, Basis: basis}, nil
}

// ReconcileBasisRow 는 근거 multiset 의 원소임 — (그룹, id, status, triggeredOrderId).
type ReconcileBasisRow struct {
	Group            string
	ID               string
	Status           string
	TriggeredOrderID string
}

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
