package journal

// reconcile_order_terminal_release.go 는 missing-order 종목 차단을 「그 주문들의 종결 증명」으로 푸는
// 장부 쪽 절반임(a192 design D2·D5).
//
// # 왜 장부 쪽에 따로 두나
//
// 해제를 판정한 읽기와 해제 쓰기 사이에 체결 감시가 추적기 잠금 없이 체결·거절·`fail_closed` 갱신을
// 커밋할 수 있음(`JournalLedger.Apply` 는 `RecordFill` 을 먼저 커밋하고 그 뒤에 `Refresh` 를 부름).
// 그래서 해제는 증명을 **같은 쓰기 트랜잭션 안에서 다시 읽고** 그 증명으로 판정을 다시 통과할 때만
// 커밋함. 판정 규칙 자체는 여기 두지 않음 — reconcile 패키지의 판정 함수 하나를 콜백으로 받음(규칙
// 복제 금지, I7).
//
// # tx 안에서 지키는 것
//
// 연결이 하나(journal.go `SetMaxOpenConns(1)`)라서 tx 를 쥔 채 `j.db` 를 쓰는 공개 메서드를 부르면 자기
// 자신을 기다림. 그래서 이 파일의 읽기는 전부 `queryRower` 하나로 함 — tx 밖(1차 판정)과 tx 안(재판정)이
// 같은 읽기 함수를 씀. lineage 는 공개 `ResolveCurrentOrderIDScoped` 를 쓰지 않음(충돌이면
// `EnterReconcile` 을 부르는 쓰기 경로) — 대신 「그 번호를 대체하는 흔적이 하나라도 있나」만 쓰기 없이 봄.

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ReconcileReleaseOrderTerminalConfirmed — missing-order 로 생긴 종목 차단을, 그 차단에 결속된 주문
// 전부가 장부에 깨끗한 종결로 기록되고 그 뒤의 비교들이 일치했다는 근거로 엔진이 푼 기록임(a192).
//
// 조정 크레딧 해제(ADJUSTMENT_APPLIED)와 다른 원인이라 따로 이름 붙임 — 이력이 「무엇이 차단을
// 닫았나」를 구분해서 말해야 함(spec: 구분되는 release cause).
const ReconcileReleaseOrderTerminalConfirmed = "ORDER_TERMINAL_CONFIRMED"

// OrderTerminalProofRequest 는 종결 증명을 읽을 대상임.
type OrderTerminalProofRequest struct {
	// StateID 는 이 프로세스가 삽입한 reconcile_states 행 id — 해제는 이 id 에 고정함.
	StateID string
	// Symbol 은 체결 워터마크 검사의 종목임.
	Symbol string
	// Orders 는 결속된 missing-order 의 정규 범위(OrderID = 감지기가 스냅숏을 쓴 원래 번호).
	Orders []FillSnapshotScope
	// FillWatermark 는 행 삽입 호출 **전**에 읽은 `max(fill_events.id)` 임.
	FillWatermark int64
}

// OrderTerminalSnapshot 은 결속 주문 하나의 장부 증명임.
type OrderTerminalSnapshot struct {
	Scope FillSnapshotScope
	// Found 는 완전한 범위로 스냅숏을 찾았는지임. 범위가 불완전하면 찾지 않음(legacy 빈 범위 행을
	// 읽지 않으려는 것 — `lookupFillSnapshotScoped` 는 불완전 범위면 legacy 행을 읽음).
	Found  bool
	Record FillSnapshotRecord
	// Amended 는 그 번호를 대체하는 정정 흔적(lineage 간선 또는 AMEND 시도)이 하나라도 있는지임.
	Amended bool
}

// OrderTerminalProof 는 판정에 넘기는 장부 증명 전부임.
type OrderTerminalProof struct {
	State      ReconcileState
	StateFound bool
	Snapshots  []OrderTerminalSnapshot
	// FillAfterWatermark 는 워터마크 뒤에 그 종목의 체결 이벤트가 하나라도 커밋됐는지임.
	FillAfterWatermark bool
	// ReadErr 는 읽기 실패임. 판정은 이것이 있으면 차단을 유지함(오류로 주기를 실패시키지 않음 — M5).
	ReadErr error
}

// MaxFillEventID 는 지금까지 커밋된 체결 이벤트의 최대 id 임(없으면 0).
//
// 차단 행을 삽입하기 전에 읽어 두면, 그 뒤 그 종목에 커밋된 체결(거절된 투영 포함 — 거절돼도 delta 는
// `fill_events` 에 남음)을 id 비교로 셀 수 있음. 시각이 아니라 id 를 쓰는 이유: `RecordFill` 의
// `committed_at` 은 tx 시작 전에 잡은 시각이라 커밋 순서가 아님. SQLite 쓰기는 직렬이라 AUTOINCREMENT
// id 순서가 커밋 순서임.
func (j *Journal) MaxFillEventID(ctx context.Context) (int64, error) {
	var id int64
	if err := j.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 0) FROM fill_events`).Scan(&id); err != nil {
		return 0, fmt.Errorf("journal: reading the fill event watermark: %w", err)
	}
	return id, nil
}

// ReadOrderTerminalProof 는 tx 밖에서 종결 증명을 읽음(1차 판정용).
//
// 실패는 오류가 아니라 증명의 `ReadErr` 로 돌려줌 — 판정이 그것을 「유지」로 읽음.
func (j *Journal) ReadOrderTerminalProof(ctx context.Context, req OrderTerminalProofRequest) OrderTerminalProof {
	return readOrderTerminalProof(ctx, j.db, req)
}

// ReleaseReconcileOnOrderTerminalProof 는 증명을 같은 쓰기 트랜잭션 안에서 다시 읽어 decide 가 참일 때만
// 그 행(id 고정, 아직 해제 안 됨)을 ORDER_TERMINAL_CONFIRMED 로 해제함.
//
// decide 가 거짓이면 오류가 아니라 「해제 안 함」(false, nil) — 차단은 그대로 남음. decide 는 순수 함수여야
// 함: tx 를 쥔 동안 공개 Journal 메서드나 추적기 잠금을 부르면 자기 자신을 기다림.
func (j *Journal) ReleaseReconcileOnOrderTerminalProof(ctx context.Context, req OrderTerminalProofRequest,
	evidence string, decide func(OrderTerminalProof) bool,
) (ReconcileState, bool, error) {
	evidence = strings.TrimSpace(evidence)
	switch {
	case strings.TrimSpace(req.StateID) == "":
		return ReconcileState{}, false, fmt.Errorf("%w: an order-terminal release names the row it releases", ErrInvalidRequest)
	case evidence == "":
		return ReconcileState{}, false, fmt.Errorf("%w: releasing a RECONCILE state requires the evidence that it is over", ErrInvalidRequest)
	case decide == nil:
		return ReconcileState{}, false, fmt.Errorf("%w: an order-terminal release needs its verdict", ErrInvalidRequest)
	case !ValidReconcileReleaseCause(ReconcileReleaseOrderTerminalConfirmed):
		// 상수 집합이 제약임 — 이 판본이 쓰는 원인으로 등록되지 않았으면 쓰지 않음.
		return ReconcileState{}, false, fmt.Errorf(
			"%w: RECONCILE release cause %q is not one this build writes", ErrInvalidRequest, ReconcileReleaseOrderTerminalConfirmed)
	}

	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: starting the order-terminal release: %w", err)
	}
	defer tx.Rollback()

	proof := readOrderTerminalProof(ctx, tx, req)
	if !decide(proof) {
		return proof.State, false, nil
	}

	now := j.clk.Now().UTC()
	nowText := formatJournalTime(now)
	result, err := tx.ExecContext(ctx,
		`UPDATE reconcile_states SET released_at = ?, release_cause = ?, evidence = ?
		 WHERE id = ? AND released_at IS NULL`,
		nowText, ReconcileReleaseOrderTerminalConfirmed, proof.State.Evidence+" | released: "+evidence, req.StateID)
	if err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: releasing RECONCILE state %s on its order-terminal proof: %w", req.StateID, err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: verifying the order-terminal release of %s: %w", req.StateID, err)
	}
	if changed != 1 {
		// 같은 tx 에서 활성으로 읽은 행이 갱신되지 않았다는 것은 판정 근거가 무너졌다는 뜻 — 오류로 알림.
		return ReconcileState{}, false, fmt.Errorf("journal: order-terminal release of %s updated %d rows, want one", req.StateID, changed)
	}
	if err := tx.Commit(); err != nil {
		return ReconcileState{}, false, fmt.Errorf("journal: committing the order-terminal release of %s: %w", req.StateID, err)
	}
	released := proof.State
	released.ReleasedAt = now
	released.ReleaseCause = ReconcileReleaseOrderTerminalConfirmed
	released.Evidence = proof.State.Evidence + " | released: " + evidence
	return released, true, nil
}

// readOrderTerminalProof 는 tx 안팎이 같이 쓰는 읽기 하나임. 첫 실패에서 멈추고 `ReadErr` 에 담음.
func readOrderTerminalProof(ctx context.Context, q queryRower, req OrderTerminalProofRequest) OrderTerminalProof {
	var proof OrderTerminalProof
	state, err := scanReconcileState(q.QueryRowContext(ctx, reconcileSelect+" WHERE id = ?", req.StateID))
	switch {
	case err == nil:
		proof.State, proof.StateFound = state, true
	case errors.Is(err, ErrReconcileStateNotFound):
		// 행이 없으면 판정이 「행이 바뀜」으로 읽음 — 읽기 실패가 아님.
	default:
		proof.ReadErr = err
		return proof
	}

	proof.Snapshots = make([]OrderTerminalSnapshot, 0, len(req.Orders))
	for _, scope := range req.Orders {
		snapshot := OrderTerminalSnapshot{Scope: scope}
		canonical := canonicalFillSnapshotScope(scope)
		if canonical.complete() {
			record, err := lookupFillSnapshotScoped(ctx, q, canonical)
			switch {
			case err == nil:
				snapshot.Found, snapshot.Record = true, record
			case errors.Is(err, ErrFillNotFound):
			default:
				proof.ReadErr = fmt.Errorf("journal: reading the terminal proof of order %s: %w", scope.OrderID, err)
				return proof
			}
		}
		amended, err := orderHasReplacementTrace(ctx, q, scope.OrderID)
		if err != nil {
			proof.ReadErr = err
			return proof
		}
		snapshot.Amended = amended
		proof.Snapshots = append(proof.Snapshots, snapshot)
	}

	// 그 종목의 체결 **또는 종목이 비어 귀속할 수 없는 체결**(schema-v15 legacy 관측은 종목 없이 들어올 수 있음 —
	// `legacyUnscoped`)이 워터마크 뒤에 있으면 「뒤의 체결」로 셈. 계좌는 보지 않음 — 다른 계좌의 같은 종목 체결도
	// 유지 쪽으로 세는 보수 방향(코드 리뷰 1 #1).
	var after int
	if err := q.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM fill_events WHERE id > ? AND (UPPER(TRIM(symbol)) = ? OR TRIM(symbol) = ''))`,
		req.FillWatermark, normaliseSymbol(req.Symbol)).Scan(&after); err != nil {
		proof.ReadErr = fmt.Errorf("journal: reading fills after the order-terminal watermark: %w", err)
		return proof
	}
	proof.FillAfterWatermark = after != 0
	return proof
}

// orderHasReplacementTrace 는 그 번호를 부모로 하는 대체 간선이나 그 번호를 겨눈 AMEND 시도가 하나라도
// 있는지임.
//
// 범위·확정 여부를 따지지 않는 넓은 존재 검사 — 정정 lineage 해석 규칙(`scopedLineageChildren`)을 여기
// 다시 쓰지 않으려는 것(규칙 복제 금지). 넓게 보는 쪽은 차단을 유지하는 쪽이라 보수 방향임.
func orderHasReplacementTrace(ctx context.Context, q queryRower, orderID string) (bool, error) {
	var found int
	if err := q.QueryRowContext(ctx, `SELECT
		EXISTS(SELECT 1 FROM scoped_lineage_edges WHERE parent_order_id = ? AND relation = ?)
		OR EXISTS(SELECT 1 FROM lineage_edges WHERE parent_order_id = ? AND relation = ?)
		OR EXISTS(SELECT 1 FROM mutation_attempts WHERE kind = ? AND target_order_id = ?)`,
		orderID, RelationReplaces, orderID, RelationReplaces, string(KindAmend), orderID).Scan(&found); err != nil {
		return false, fmt.Errorf("journal: reading the replacement trace of order %s: %w", orderID, err)
	}
	return found != 0, nil
}
