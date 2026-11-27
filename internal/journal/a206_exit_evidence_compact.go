package journal

// a206_exit_evidence_compact.go 는 닫힌 포지션의 exit_events 판정 증거를 비우는 운영자 정비임(a206 design D1 · D2 · D4).
//
// 거절된 청산의 재무장마다 스냅숏 JSON 셋이 새 행에 쌓여 운영 저널의 79% 가 exit_events 였음. 닫힌(CLOSED · completed) 포지션의
// 스냅숏은 엔진이 다시 읽지 않으므로(design F4) 브로커 주문에 연결되지 않은 행의 증거 20 열을 통째로 NULL 로 바꾸고 VACUUM 함.
// 일부 열만 비우면 리더가 그 행을 손상(ErrExitSnapshotCorrupt)으로 읽으므로 20 열은 언제나 함께 비움(design S1 기각 · F5).
// 행 자체와 action · proposed_intent_id · 가격 · 시각 · lifecycle_generation 은 남김.
//
// 자동 경로는 이 파일을 부르지 않음 — 호출자는 엔진 잠금을 쥔 `tossctl engine exit-evidence-compact --apply` 하나뿐임.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
)

// a206EvidenceColumns 는 schemaV10 이 exit_events 에 더한 판정 증거 열 20 개임(lifecycle_generation 제외 — 행의 수명주기 결속이라 남김).
var a206EvidenceColumns = []string{
	"position_generation", "policy_id", "policy_version", "policy_digest", "snapshot_id", "decision_id",
	"observation_id", "next_target", "next_protection", "observation_source", "observed_at",
	"projected_quantity", "proposal_ratio", "state_only", "suppressed_reason", "saved_snapshot_json",
	"recomputed_snapshot_json", "effective_snapshot_json", "effective_source", "arm_suppressed_reason",
}

// a206KeepSQL 은 a207 압축 행 자신과 그 앵커의 id 임 — a207 design D8 이 정한 술어를 그대로 씀(사람 결정 6).
// 압축 행을 비우면 표지가 사라져 다음 실행이 앵커를 못 찾고 비우므로(멱등 깨짐) 둘 다 남김.
const a206KeepSQL = `
	SELECT c.id FROM exit_events c WHERE c.effective_source = 'recomputed_repeat'
	UNION
	SELECT (SELECT max(a.id) FROM exit_events a
	         WHERE a.position_id = c.position_id AND a.id < c.id
	           AND coalesce(a.proposed_intent_id, '') <> ''
	           AND coalesce(a.action, '') NOT IN ('OPENED','READOPTED','PROPOSAL_REFUSED','PROPOSAL_CANCELLED',
	                                              'PROPOSAL_FILLED','COMPLETED','ADJUSTMENT_CLOSED')
	           AND a.recomputed_snapshot_json IS NOT NULL
	           AND coalesce(a.effective_source, '') <> 'recomputed_repeat')
	  FROM exit_events c WHERE c.effective_source = 'recomputed_repeat'`

// a206TargetSQL 은 비울 행의 id 임(design D1 · D2).
//   - CLOSED 이고 exit_states.completed=1 인 포지션의 행만 — 열린 포지션은 엔진이 지금 판정 중임.
//   - 그 intent 의 attempt 중 broker_order_id 를 가진 것이 하나라도 있으면 제외 — 콘솔 orders 가 연결할 수 있는 행의 상위 집합임.
//   - a207 압축 행 · 앵커 제외. 앵커가 없는 압축 행의 max() 는 NULL 이고, NOT IN 목록의 NULL 은 모든 비교를 NULL 로 만들어
//     대상 전체를 지우므로 NULL 을 먼저 걸러냄.
//   - 이미 20 열이 모두 NULL 인 행(훅 행 · legacy 행 · 앞선 정비 행)은 세지 않음 — 두 번째 실행이 0 행이 되는 근거임.
var a206TargetSQL = `
	SELECT e.id FROM exit_events e
	  JOIN positions p ON p.id = e.position_id AND p.state = 'CLOSED'
	  JOIN exit_states s ON s.position_id = e.position_id AND s.completed = 1
	 WHERE NOT EXISTS (SELECT 1 FROM mutation_attempts m
	                    WHERE m.intent_id = e.proposed_intent_id AND coalesce(m.broker_order_id, '') <> '')
	   AND e.id NOT IN (SELECT id FROM (` + a206KeepSQL + `) WHERE id IS NOT NULL)
	   AND NOT (e.` + strings.Join(a206EvidenceColumns, " IS NULL AND e.") + ` IS NULL)`

// ExitEvidenceCompaction 은 정비 한 번의 집계임 — 행 id · 종목 · 가격은 담지 않음(audit · 표준 출력에 그대로 나감).
type ExitEvidenceCompaction struct {
	Rows          int64 // exit_events 전체 행 수(정비 전)
	Kept          int64 // a207 압축 행 · 앵커로 남긴 행 수(전체 포지션)
	Targets       int64 // 증거를 비울(비운) 행 수
	SnapshotBytes int64 // 대상 행의 스냅숏 JSON 셋 바이트 합(정비 전)
	FileBytes     int64 // journal.db + -wal 크기(정비 전)

	// 아래는 적용(CompactClosedExitEvidence)에서만 채움.
	Backup         string // VACUUM INTO 백업 경로 — 사람이 검증 뒤 지움
	Emptied        int64  // UPDATE 가 실제로 바꾼 행 수
	RowsAfter      int64  // exit_events 전체 행 수(정비 뒤) — Rows 와 같아야 함(행 삭제 없음)
	Remaining      int64  // 정비 뒤 다시 센 대상 수 — 0 이어야 함
	FileBytesAfter int64  // VACUUM · checkpoint 뒤 journal.db + -wal 크기
}

// a206RowQuerier 는 *sql.DB 와 *sql.Tx 가 같이 갖는 한 행 질의임 — 계획 집계를 읽기 전용 연결과 쓰기 연결에서 같은 SQL 로 셈.
type a206RowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// exitEvidenceCompactionPlanQ 는 쓰지 않고 대상 집계만 셈.
func exitEvidenceCompactionPlanQ(ctx context.Context, q a206RowQuerier) (ExitEvidenceCompaction, error) {
	var plan ExitEvidenceCompaction
	err := q.QueryRowContext(ctx, `
		SELECT (SELECT count(*) FROM exit_events),
		       (SELECT count(*) FROM (`+a206KeepSQL+`) WHERE id IS NOT NULL),
		       count(*),
		       coalesce(sum(length(coalesce(saved_snapshot_json,'')) + length(coalesce(recomputed_snapshot_json,'')) +
		                    length(coalesce(effective_snapshot_json,''))), 0)
		  FROM exit_events WHERE id IN (`+a206TargetSQL+`)`).
		Scan(&plan.Rows, &plan.Kept, &plan.Targets, &plan.SnapshotBytes)
	if err != nil {
		return ExitEvidenceCompaction{}, fmt.Errorf("journal: counting the exit evidence to compact: %w", err)
	}
	return plan, nil
}

// journalFileBytes 는 본 파일과 WAL 크기의 합임 — 커밋이 WAL 에만 있을 수 있으므로 둘을 같이 셈.
func journalFileBytes(path string) (int64, error) {
	var total int64
	for _, p := range []string{path, path + "-wal"} {
		info, err := os.Stat(p)
		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return 0, fmt.Errorf("journal: measuring %s: %w", p, err)
		default:
			total += info.Size()
		}
	}
	return total, nil
}

// ExitEvidenceCompactionPlan 은 dry-run 집계임 — 읽기 전용 연결이라 아무것도 쓰지 않음(design D3).
func (r *ReadOnly) ExitEvidenceCompactionPlan(ctx context.Context) (ExitEvidenceCompaction, error) {
	plan, err := exitEvidenceCompactionPlanQ(ctx, r.db)
	if err != nil {
		return ExitEvidenceCompaction{}, err
	}
	if plan.FileBytes, err = journalFileBytes(r.path); err != nil {
		return ExitEvidenceCompaction{}, err
	}
	return plan, nil
}

// CompactClosedExitEvidence 는 design D4 순서로 정비함: 백업 → UPDATE 한 트랜잭션(무결성 검사 포함) → VACUUM → 재집계.
// 엔진 잠금은 호출자가 쥠(이 패키지는 잠금을 모름).
//
// 실패 계약:
//   - 백업 실패: 아무것도 바뀌지 않음.
//   - UPDATE · 무결성 검사 실패: 트랜잭션을 되돌림. 백업은 남김(report.Backup).
//   - VACUUM 이후 실패: 비운 증거는 이미 커밋됐고 유효함 — 파일만 아직 안 줄었음. 다시 실행하면(대상 0 행) VACUUM 을 다시 함.
func (j *Journal) CompactClosedExitEvidence(ctx context.Context) (ExitEvidenceCompaction, error) {
	report, err := exitEvidenceCompactionPlanQ(ctx, j.db)
	if err != nil {
		return ExitEvidenceCompaction{}, err
	}
	if report.FileBytes, err = journalFileBytes(j.path); err != nil {
		return ExitEvidenceCompaction{}, err
	}

	// 백업 이름은 마이그레이션 백업과 구분되게 a206 표지 + UTC 시각임(사람 결정 3 — 저널 옆).
	// VACUUM INTO 는 이미 있는 비어 있지 않은 파일을 덮어쓰지 않으므로 같은 이름이면 실패하고 아무것도 안 바뀜.
	backup := fmt.Sprintf("%s.pre-a206-%s.bak", j.path, j.clk.Now().UTC().Format(backupTimeLayout))
	if _, err := j.db.ExecContext(ctx, "VACUUM INTO ?", backup); err != nil {
		return report, fmt.Errorf("journal: backing up %s to %s before the exit evidence compaction: %w; nothing was changed",
			j.path, backup, err)
	}
	// 백업은 원장 전체를 담으므로 원장과 같은 권한임(backup.go 와 같은 이유).
	if err := os.Chmod(backup, 0o600); err != nil {
		return report, fmt.Errorf("journal: securing the backup %s: %w; nothing was changed", backup, err)
	}
	report.Backup = backup

	if report.Emptied, err = j.emptyClosedExitEvidenceTx(ctx); err != nil {
		return report, fmt.Errorf("%w; the journal is unchanged and the backup is at %s", err, backup)
	}

	// VACUUM 은 트랜잭션 밖에서만 돎. WAL 모드라 새 페이지는 WAL 에 먼저 쓰이므로 TRUNCATE checkpoint 로 본 파일을 줄임.
	if _, err := j.db.ExecContext(ctx, "VACUUM"); err != nil {
		return report, fmt.Errorf("journal: VACUUM after the exit evidence compaction: %w; the emptied evidence is committed, "+
			"the file is not smaller yet — run the command again to retry the VACUUM (backup %s)", err, backup)
	}
	var busy, logFrames, checkpointed int64
	if err := j.db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &logFrames, &checkpointed); err != nil {
		return report, fmt.Errorf("journal: checkpointing after the exit evidence compaction: %w (backup %s)", err, backup)
	}
	// 열린 읽기 연결(콘솔 등)이 WAL 을 잡으면 checkpoint 는 SQL 오류 없이 busy=1 을 돌려줌 — 본 파일이 아직 안 줄었으므로 성공으로 보고하지 않음.
	if busy != 0 {
		return report, fmt.Errorf("journal: the checkpoint after the exit evidence compaction was blocked by another open reader; "+
			"the emptied evidence is committed and the file shrinks at the next checkpoint — stop the console and run the command again "+
			"(backup %s)", backup)
	}

	after, err := exitEvidenceCompactionPlanQ(ctx, j.db)
	if err != nil {
		return report, err
	}
	report.RowsAfter, report.Remaining = after.Rows, after.Targets
	if report.FileBytesAfter, err = journalFileBytes(j.path); err != nil {
		return report, err
	}
	return report, nil
}

// emptyClosedExitEvidenceTx 는 대상 계산과 UPDATE 를 한 트랜잭션(BEGIN IMMEDIATE)에서 함.
// 대상 id 를 임시 표에 먼저 담는 이유는 제외 집합(a207 앵커)을 UPDATE 전에 확정하기 위함임(design D1) — UPDATE 가 바꾸는 열
// (recomputed_snapshot_json · effective_source)을 제외 술어가 읽으므로 계산과 쓰기를 한 문장에 섞지 않음.
// 커밋 전에 같은 트랜잭션에서 integrity_check 를 돌려 실패면 되돌림(D4).
func (j *Journal) emptyClosedExitEvidenceTx(ctx context.Context) (int64, error) {
	tx, err := j.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("journal: beginning the exit evidence compaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, stmt := range []string{
		"CREATE TEMP TABLE a206_compact_target(id INTEGER PRIMARY KEY)",
		"INSERT INTO temp.a206_compact_target(id) " + a206TargetSQL,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return 0, fmt.Errorf("journal: selecting the exit evidence to compact: %w", err)
		}
	}
	result, err := tx.ExecContext(ctx, "UPDATE exit_events SET "+strings.Join(a206EvidenceColumns, " = NULL, ")+
		" = NULL WHERE id IN (SELECT id FROM temp.a206_compact_target)")
	if err != nil {
		return 0, fmt.Errorf("journal: emptying the closed exit evidence: %w", err)
	}
	emptied, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("journal: counting the emptied exit evidence: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DROP TABLE temp.a206_compact_target"); err != nil {
		return 0, fmt.Errorf("journal: dropping the compaction target list: %w", err)
	}
	var verdict string
	if err := tx.QueryRowContext(ctx, "PRAGMA integrity_check(1)").Scan(&verdict); err != nil {
		return 0, fmt.Errorf("journal: integrity check before committing the compaction: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(verdict), "ok") {
		return 0, fmt.Errorf("journal: integrity check before committing the compaction failed: %s", verdict)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("journal: committing the exit evidence compaction: %w", err)
	}
	return emptied, nil
}
