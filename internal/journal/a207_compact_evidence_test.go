package journal

// a207 — 막힌 청산 재시도의 반복 판정은 압축 증거 행으로 기록된다(design D1~D4 · D8, 시험 T2 · T3 · T4 · T5 · T7 · T8).

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// a207ArmRow 는 무장 행 하나의 증거 모양임 — JSON 셋의 유무, 표지, decision id.
type a207ArmRow struct {
	id                         int64
	saved, recomputed, eff     bool
	source, decisionID, action string
}

func a207RowOf(t *testing.T, j *Journal, intentID string) a207ArmRow {
	t.Helper()
	var r a207ArmRow
	var source, decision, action sql.NullString
	err := j.db.QueryRow(`SELECT id, saved_snapshot_json IS NOT NULL, recomputed_snapshot_json IS NOT NULL,
		effective_snapshot_json IS NOT NULL, effective_source, decision_id, action
		FROM exit_events WHERE proposed_intent_id = ? AND action NOT IN ('PROPOSAL_REFUSED','PROPOSAL_CANCELLED','PROPOSAL_FILLED')`,
		intentID).Scan(&r.id, &r.saved, &r.recomputed, &r.eff, &source, &decision, &action)
	if err != nil {
		t.Fatalf("reading the arm row of %s: %v", intentID, err)
	}
	r.source, r.decisionID, r.action = source.String, decision.String, action.String
	return r
}

func (r a207ArmRow) compact() bool {
	return !r.saved && !r.recomputed && !r.eff && r.source == EffectiveSourceRecomputedRepeat && r.decisionID != ""
}

func (r a207ArmRow) full() bool {
	return r.recomputed && r.eff && r.source == EffectiveSourceRecomputed && r.decisionID != ""
}

// a207Arm 은 현재 상태에서 손절 이탈 판정 하나를 무장함(같은 기준선 · 같은 high water — 같은 action · 수준).
func a207Arm(t *testing.T, j *Journal, positionID, observation, price, intentID string) (exitpolicy.ExitLineSnapshot, ExitJudgement, error) {
	t.Helper()
	state := exitStateOf(t, j, positionID)
	snapshot, recovery := ratchetSnapshotForState(t, state, observation, price, "70000", "68000")
	if snapshot.Action != exitpolicy.ActionBaselineBreach || !snapshot.Orderable {
		t.Fatalf("fixture is not an orderable baseline breach: %+v", snapshot)
	}
	judgement := judgementForSnapshot(snapshot, recovery)
	judgement.Proposal = &ExitProposal{Action: string(snapshot.Action), Level: snapshot.Level,
		IntentID: intentID, Provenance: judgement.Provenance}
	return snapshot, judgement, j.RecordExitJudgement(context.Background(), judgement)
}

func a207Refuse(t *testing.T, j *Journal, positionID, intentID string) {
	t.Helper()
	_, released, err := j.ReleaseUnacceptedExitProposal(context.Background(), positionID, intentID, ProposalRefused)
	if err != nil || !released {
		t.Fatalf("releasing %s: released=%v err=%v", intentID, released, err)
	}
}

// T2 — 같은 키의 반복 무장은 압축, 구간 첫 행은 전체 증거.
func TestA207ARepeatedRefusedArmIsWrittenCompact(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
		t.Fatal(err)
	}
	a207Refuse(t, j, id, "exit-a207-1")
	if _, _, err := a207Arm(t, j, id, "obs-2", "67800", "exit-a207-2"); err != nil {
		t.Fatal(err)
	}
	a207Refuse(t, j, id, "exit-a207-2")
	if _, _, err := a207Arm(t, j, id, "obs-3", "67700", "exit-a207-3"); err != nil {
		t.Fatal(err)
	}

	if first := a207RowOf(t, j, "exit-a207-1"); !first.full() {
		t.Fatalf("the first arm of the run = %+v, want full evidence", first)
	}
	for _, intent := range []string{"exit-a207-2", "exit-a207-3"} {
		if row := a207RowOf(t, j, intent); !row.compact() {
			t.Fatalf("repeat arm %s = %+v, want compact (no JSON, marker, decision id kept)", intent, row)
		}
	}
	// 무장 셋 · 해제 훅 둘이 모두 남음(행 수 불변).
	var arms, hooks int
	if err := j.db.QueryRow(`SELECT sum(action=?), sum(action='PROPOSAL_REFUSED') FROM exit_events WHERE position_id=?`,
		string(exitpolicy.ActionBaselineBreach), id).
		Scan(&arms, &hooks); err != nil {
		t.Fatal(err)
	}
	if arms != 3 || hooks != 2 {
		t.Fatalf("arm rows %d · refusal hooks %d, want 3 · 2 — compaction must not drop rows", arms, hooks)
	}
}

// T2 — 술어의 각 항(직접 호출): 같은 트랜잭션에서 action · 수준 · 세대 · 해제 · decision id 가 하나라도 다르면 압축 아님.
func TestA207TheRepeatPredicateNeedsEveryTerm(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	snapshot, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1")
	if err != nil {
		t.Fatal(err)
	}
	level := a207LevelAfterOf(t, j, "exit-a207-1")
	action := string(snapshot.Action)
	ask := func(gen int64, act, lvl string) bool {
		t.Helper()
		tx, err := j.db.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		got, err := repeatsARefusedArmTx(context.Background(), tx, id, gen, act, lvl)
		if err != nil {
			t.Fatalf("predicate error: %v", err)
		}
		return got
	}
	gen := a207GenerationOf(t, j, id)
	if ask(gen, action, level) {
		t.Fatal("an unreleased previous arm made the next one compact")
	}
	a207Refuse(t, j, id, "exit-a207-1")
	if !ask(gen, action, level) {
		t.Fatal("same generation · action · level after a refusal must be compact")
	}
	if ask(gen, string(exitpolicy.ActionLadderStop), level) {
		t.Fatal("a different action must be full")
	}
	if ask(gen, action, level+"-other") {
		t.Fatal("a different level must be full")
	}
	if ask(gen+1, action, level) {
		t.Fatal("a different lifecycle generation must be full")
	}
	// 직전 무장이 decision id 를 갖지 않으면(legacy 판정) 앵커가 없으므로 전체 증거.
	if _, err := j.db.Exec(`UPDATE exit_events SET decision_id=NULL WHERE proposed_intent_id='exit-a207-1' AND action=?`, action); err != nil {
		t.Fatal(err)
	}
	if ask(gen, action, level) {
		t.Fatal("a previous arm without a decision id must not anchor a compact row")
	}
}

// T8 — 술어 조회가 실패하면 전체 증거로 쓰고 판정은 무장으로 끝남.
func TestA207ALookupFaultWritesTheFullRowAndStillArms(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
		t.Fatal(err)
	}
	a207Refuse(t, j, id, "exit-a207-1")
	orig := repeatsARefusedArmTx
	repeatsARefusedArmTx = func(context.Context, *sql.Tx, string, int64, string, string) (bool, error) {
		return false, errors.New("injected lookup fault")
	}
	t.Cleanup(func() { repeatsARefusedArmTx = orig })
	_, judgement, err := a207Arm(t, j, id, "obs-2", "67800", "exit-a207-2")
	if err != nil {
		t.Fatalf("a lookup fault failed the judgement: %v", err)
	}
	if row := a207RowOf(t, j, "exit-a207-2"); !row.full() {
		t.Fatalf("row after a lookup fault = %+v, want full evidence", row)
	}
	if state := exitStateOf(t, j, id); state.PendingIntentID != judgement.Proposal.IntentID {
		t.Fatalf("pending intent = %q, want the proposal armed despite the lookup fault", state.PendingIntentID)
	}
}

// T4 — 압축 행과 같은 decision 의 판정은 지금처럼 제출되지 않음(유일 인덱스 · :494 가드가 압축 행에도 섬).
func TestA207ACompactRowStillBlocksItsDuplicateDecision(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
		t.Fatal(err)
	}
	a207Refuse(t, j, id, "exit-a207-1")
	_, judgement, err := a207Arm(t, j, id, "obs-2", "67800", "exit-a207-2")
	if err != nil {
		t.Fatal(err)
	}
	if row := a207RowOf(t, j, "exit-a207-2"); !row.compact() {
		t.Fatalf("fixture row = %+v, want compact", row)
	}
	a207Refuse(t, j, id, "exit-a207-2")
	// 같은 견적 · 같은 상태 → 같은 decision. 발의가 풀린 뒤라 pending 가드가 아니라 decision 가드만 이것을 막음.
	if err := j.RecordExitJudgement(context.Background(), judgement); !errors.Is(err, ErrProposalPending) {
		t.Fatalf("duplicate decision on a compact row = %v, want ErrProposalPending", err)
	}
	if state := exitStateOf(t, j, id); state.Pending() {
		t.Fatal("the duplicate decision armed a proposal")
	}
}

// T3 — 리더: 판정 이력 · 계좌 이력 · orders 단일 연결은 오류 없이 새 사유, 표지 없는 부분 모양은 여전히 손상.
func TestA207ReadersReadACompactRowAsARepeat(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
		t.Fatal(err)
	}
	a207Refuse(t, j, id, "exit-a207-1")
	if _, _, err := a207Arm(t, j, id, "obs-2", "67800", "exit-a207-2"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	events, err := j.ExitEvents(ctx, id)
	if err != nil {
		t.Fatalf("ExitEvents over a compact row: %v", err)
	}
	var seen bool
	for _, e := range events {
		if e.ProposedIntentID != "exit-a207-2" || e.Action != string(exitpolicy.ActionBaselineBreach) {
			continue
		}
		seen = true
		ev := e.Evaluation
		if ev.EffectiveSource != EffectiveSourceRecomputedRepeat || ev.Effective.Snapshot != nil ||
			ev.Effective.UnknownReason != UnknownReasonRepeatOfRefusedProposal ||
			ev.Recomputed.UnknownReason != UnknownReasonRepeatOfRefusedProposal ||
			ev.Saved.UnknownReason != UnknownReasonRepeatOfRefusedProposal {
			t.Fatalf("compact evaluation = %+v, want the repeat reason on all three views", ev)
		}
	}
	if !seen {
		t.Fatal("the compact arm row is missing from the history")
	}

	var account string
	if err := j.db.QueryRow(`SELECT account_ref FROM positions WHERE id=?`, id).Scan(&account); err != nil {
		t.Fatal(err)
	}
	ro := openTestReadOnly(t, j.Path())
	history, err := ro.AccountExitEvents(ctx, account, 100)
	if err != nil {
		t.Fatalf("AccountExitEvents: %v", err)
	}
	var historySeen bool
	for _, e := range history {
		if e.Event.ProposedIntentID == "exit-a207-2" && e.Event.Action == string(exitpolicy.ActionBaselineBreach) {
			historySeen = true
			if e.Event.Evaluation.Effective.UnknownReason != UnknownReasonRepeatOfRefusedProposal {
				t.Fatalf("account history reason = %q, want %q", e.Event.Evaluation.Effective.UnknownReason,
					UnknownReasonRepeatOfRefusedProposal)
			}
		}
	}
	if !historySeen {
		t.Fatal("the compact arm row is missing from the account history")
	}

	insertIntent(t, j, "exit-a207-2")
	insertAttemptWithBrokerOrder(t, j, "attempt-a207-2", "exit-a207-2", "broker-a207-2")
	links, err := ro.BrokerOrderExitLinks(ctx, []BrokerOrderScope{{BrokerOrderID: "broker-a207-2", AccountRef: "acct-1",
		Market: "us", TradingDay: "2026-03-30"}})
	if err != nil {
		t.Fatalf("BrokerOrderExitLinks: %v", err)
	}
	if len(links) != 1 || links[0].Ambiguous || links[0].Event.Evaluation.Effective.UnknownReason != UnknownReasonRepeatOfRefusedProposal {
		t.Fatalf("orders link = %+v, want one unambiguous link with the repeat reason", links)
	}

	// 표지 없이 JSON 만 빠진 행은 여전히 손상.
	if _, err := j.db.Exec(`UPDATE exit_events SET effective_source=? WHERE proposed_intent_id='exit-a207-2' AND action=?`,
		EffectiveSourceRecomputed, string(exitpolicy.ActionBaselineBreach)); err != nil {
		t.Fatal(err)
	}
	if _, err := j.ExitEvents(ctx, id); !errors.Is(err, ErrExitSnapshotCorrupt) {
		t.Fatalf("a partial row without the marker = %v, want ErrExitSnapshotCorrupt", err)
	}
}

// T3 — 표지가 있어도 모양이 어긋나면(JSON 이 남음 · 평탄화 필수 열 없음 · arm suppression 있음) 손상.
func TestA207AMarkedRowWithTheWrongShapeIsCorrupt(t *testing.T) {
	for name, mutate := range map[string]string{
		"json kept":                `UPDATE exit_events SET recomputed_snapshot_json='{}' WHERE proposed_intent_id='exit-a207-2' AND action=?`,
		"decision id missing":      `UPDATE exit_events SET decision_id=NULL WHERE proposed_intent_id='exit-a207-2' AND action=?`,
		"arm suppression added":    `UPDATE exit_events SET arm_suppressed_reason='working_order_not_cleared' WHERE proposed_intent_id='exit-a207-2' AND action=?`,
		"state only not canonical": `UPDATE exit_events SET state_only=7 WHERE proposed_intent_id='exit-a207-2' AND action=?`,
	} {
		t.Run(name, func(t *testing.T) {
			j := exitFixture(t)
			_, seed := openedPosition(t, j, "10")
			id := seed.PositionID
			if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
				t.Fatal(err)
			}
			a207Refuse(t, j, id, "exit-a207-1")
			if _, _, err := a207Arm(t, j, id, "obs-2", "67800", "exit-a207-2"); err != nil {
				t.Fatal(err)
			}
			if row := a207RowOf(t, j, "exit-a207-2"); !row.compact() {
				t.Fatalf("fixture row = %+v, want compact before the shape is broken", row)
			}
			if _, err := j.db.Exec(mutate, string(exitpolicy.ActionBaselineBreach)); err != nil {
				t.Fatal(err)
			}
			if _, err := j.ExitEvents(context.Background(), id); !errors.Is(err, ErrExitSnapshotCorrupt) {
				t.Fatalf("err = %v, want ErrExitSnapshotCorrupt", err)
			}
		})
	}
}

// T5 — 막힘 없는 경로의 행은 지금과 같음: 첫 무장 · 판정 전용 행은 전체 증거, 같은 키라도 판정 전용 행은 압축 안 함.
func TestA207RowsOutsideARefusalRunAreUnchanged(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	ctx := context.Background()
	// 판정 전용(발의 없음) 행.
	state := exitStateOf(t, j, id)
	snap, rec := ratchetSnapshotForState(t, state, "obs-flat", "69900", "70000", "68000")
	if err := j.RecordExitJudgement(ctx, judgementForSnapshot(snap, rec)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
		t.Fatal(err)
	}
	if row := a207RowOf(t, j, "exit-a207-1"); !row.full() {
		t.Fatalf("first arm = %+v, want full", row)
	}
	a207Refuse(t, j, id, "exit-a207-1")
	// 해제 뒤 같은 상태의 판정 전용 행(발의 없음)도 압축하지 않음 — 압축은 무장 행만.
	state = exitStateOf(t, j, id)
	flat, flatRec := ratchetSnapshotForState(t, state, "obs-flat-2", "67850", "70000", "68000")
	judgement := judgementForSnapshot(flat, flatRec)
	judgement.ArmSuppressedReason = ArmSuppressedWorkingOrder
	if err := j.RecordExitJudgement(ctx, judgement); err != nil {
		t.Fatal(err)
	}
	var compact int
	if err := j.db.QueryRow(`SELECT count(*) FROM exit_events WHERE position_id=? AND effective_source=?`, id,
		EffectiveSourceRecomputedRepeat).Scan(&compact); err != nil {
		t.Fatal(err)
	}
	if compact != 0 {
		t.Fatalf("compact rows = %d outside a repeated arm, want 0", compact)
	}
}

// T7 — 압축 무장 행의 intent 도 재생 출처를 입증함(action 은 남으므로 손절은 그대로 Critical 로 분류됨).
func TestA207ACompactArmStillProvesTheReplayOrigin(t *testing.T) {
	j := exitFixture(t)
	_, seed := openedPosition(t, j, "10")
	id := seed.PositionID
	if _, _, err := a207Arm(t, j, id, "obs-1", "67900", "exit-a207-1"); err != nil {
		t.Fatal(err)
	}
	a207Refuse(t, j, id, "exit-a207-1")
	if _, _, err := a207Arm(t, j, id, "obs-2", "67800", "exit-a207-2"); err != nil {
		t.Fatal(err)
	}
	if row := a207RowOf(t, j, "exit-a207-2"); !row.compact() {
		t.Fatalf("fixture row = %+v, want compact", row)
	}
	var account, market, symbol string
	if err := j.db.QueryRow(`SELECT account_ref, market, symbol FROM positions WHERE id=?`, id).Scan(&account, &market, &symbol); err != nil {
		t.Fatal(err)
	}
	action, ok, err := j.ExitProposalForIntent(context.Background(), Intent{ID: "exit-a207-2", AccountRef: account, Market: market, Symbol: symbol})
	if err != nil || !ok || action != string(exitpolicy.ActionBaselineBreach) {
		t.Fatalf("exit origin of a compact arm = %q ok=%v err=%v, want the stop action proven", action, ok, err)
	}
}

func a207LevelAfterOf(t *testing.T, j *Journal, intentID string) string {
	t.Helper()
	var level sql.NullString
	if err := j.db.QueryRow(`SELECT level_after FROM exit_events WHERE proposed_intent_id=? AND action NOT IN ('PROPOSAL_REFUSED','PROPOSAL_CANCELLED','PROPOSAL_FILLED')`,
		intentID).Scan(&level); err != nil {
		t.Fatal(err)
	}
	return level.String
}

func a207GenerationOf(t *testing.T, j *Journal, positionID string) int64 {
	t.Helper()
	var gen int64
	if err := j.db.QueryRow(`SELECT coalesce(lifecycle_generation,1) FROM exit_states WHERE position_id=?`, positionID).Scan(&gen); err != nil {
		t.Fatal(err)
	}
	return gen
}
