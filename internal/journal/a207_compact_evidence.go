package journal

// a207_compact_evidence.go 는 막힌 청산 재시도의 압축 증거 행을 만들고 읽는 자리임(a207 design D1~D4 · D8).
//
// 거절된 손절은 다음 관측에서 다시 무장됨(§0.3 — 대기 · 상한 없음). 그 재무장마다 스냅숏 JSON 셋(평균 6.5 KB)을 새로 쓰면 2 s 주기에서
// 포지션 하나가 하루 ≈309 MB 를 씀. 그래서 같은 포지션의 직전 무장이 같은 (세대, action, 수준)이고 이미 해제됐으면 JSON 셋을 빼고
// 평탄화 증거(decision id 포함)만 씀. 행 수 · 무장 · 제출 · 해제 · 중복 판정 가드는 그대로임.

import (
	"context"
	"database/sql"
	"fmt"
)

// a207HookActions 는 무장 행이 아닌 행(훅)의 action 임 — 「직전 무장」을 찾을 때 건너뜀.
const a207HookActions = `'` + ExitEventOpened + `','` + ExitEventReadopted + `','` + ExitEventProposalRefused + `','` +
	ExitEventProposalCancelled + `','` + ExitEventProposalFilled + `','` + ExitEventCompleted + `','` + ExitEventAdjustmentClosed + `'`

// repeatsARefusedArmTx 는 이 무장이 「직전 무장과 같은 조건의 반복」인지 판정 트랜잭션 안에서 답함(design D1 · D2).
//
// 참이려면 셋 다임: 이 포지션의 직전 무장 행(훅 아님 · intent 있음)이 decision id 를 가짐(전체 또는 압축 — 구간의 앵커가 늘 있게 함, D8),
// 그 행의 세대 · action · 수준이 이 무장과 같음, 그 intent 에 해제 훅(PROPOSAL_REFUSED · PROPOSAL_CANCELLED)이 있음.
// 같은 트랜잭션(BEGIN IMMEDIATE)의 읽기라 판정과 쓰기 사이에 다른 쓰기가 끼지 않음.
//
// 패키지 변수인 것은 조회 실패 시 전체 증거로 떨어지는지(D1b)를 시험이 주입해 재기 위함 — 생산 값은 아래 SQL 하나뿐임
// (선례 alert_claim.go readSettledAttemptsTx).
var repeatsARefusedArmTx = func(ctx context.Context, tx *sql.Tx, positionID string, lifecycle int64,
	action, levelAfter string) (bool, error) {
	var repeat int
	err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
		  SELECT 1 FROM (
		    SELECT e.proposed_intent_id AS intent, coalesce(e.lifecycle_generation, 1) AS generation,
		           e.action AS action, coalesce(e.level_after, '') AS level, e.decision_id AS decision
		      FROM exit_events e
		     WHERE e.position_id = ? AND coalesce(e.proposed_intent_id, '') <> ''
		       AND coalesce(e.action, '') NOT IN (`+a207HookActions+`)
		     ORDER BY e.id DESC LIMIT 1) prev
		   WHERE prev.generation = ? AND prev.action = ? AND prev.level = ? AND prev.decision IS NOT NULL
		     AND EXISTS (SELECT 1 FROM exit_events h WHERE h.proposed_intent_id = prev.intent
		                    AND h.action IN (?, ?)))`,
		positionID, lifecycle, action, levelAfter, ExitEventProposalRefused, ExitEventProposalCancelled).Scan(&repeat)
	if err != nil {
		return false, fmt.Errorf("journal: reading the previous arm of %s: %w", positionID, err)
	}
	return repeat != 0, nil
}

// hydrateCompactExitEvent 는 압축 표지를 가진 행을 읽음(design D4). 모양이 맞으면 세 뷰를 반복 재시도 사유의 unknown 으로 채우고,
// 하나라도 어긋나면 손상임 — 표지만 붙은 부분 증거가 정상으로 읽히면 안 됨.
//
// 모양: 필수 평탄화 열이 전부 있고 state_only 가 0/1, JSON 셋은 NULL, arm suppression 은 NULL(무장 행이라 원래 없음).
// 스냅숏이 없으므로 평탄화 ↔ 스냅숏 교차 검사는 하지 않음(design §3.2 R3).
func hydrateCompactExitEvent(event *ExitEvent, evidence exitEventEvidence) error {
	shaped := evidence.Generation.Valid && evidence.PolicyID.Valid && evidence.PolicyVersion.Valid &&
		evidence.PolicyDigest.Valid && evidence.SnapshotID.Valid && evidence.DecisionID.Valid &&
		evidence.ObservationID.Valid && evidence.ObservationSource.Valid && evidence.ObservedAt.Valid &&
		evidence.ProjectedQuantity.Valid && evidence.StateOnly.Valid &&
		(evidence.StateOnly.Int64 == 0 || evidence.StateOnly.Int64 == 1) &&
		!evidence.SavedJSON.Valid && !evidence.RecomputedJSON.Valid && !evidence.EffectiveJSON.Valid &&
		!evidence.ArmSuppressedReason.Valid
	if !shaped {
		return fmt.Errorf("%w: compact event evidence is not in its shape", ErrExitSnapshotCorrupt)
	}
	event.Evaluation.EffectiveSource = EffectiveSourceRecomputedRepeat
	event.Evaluation.Saved.UnknownReason = UnknownReasonRepeatOfRefusedProposal
	event.Evaluation.Recomputed.UnknownReason = UnknownReasonRepeatOfRefusedProposal
	event.Evaluation.Effective.UnknownReason = UnknownReasonRepeatOfRefusedProposal
	return nil
}
