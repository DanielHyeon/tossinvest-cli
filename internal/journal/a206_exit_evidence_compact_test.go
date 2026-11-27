package journal

// a206 — 닫힌 포지션의 미연결 행만 증거 20 열을 통째로 비운다(design D1 · D2 · D4, 시험 D6 (1) · (2) · (3)).

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// a206Position 은 openedPosition 과 같은 순서(진입 → 체결 → exit 상태)로 포지션 하나를 만듦 — 한 저널에 여럿을 두려고 id 를 tag 로 가름.
func a206Position(t *testing.T, j *Journal, tag, symbol string) string {
	t.Helper()
	ctx := context.Background()
	o := place(t, j, order{
		intentID: "i-entry-" + tag, attemptID: "a-entry-" + tag, orderID: "o-entry-" + tag,
		decisionID: "d-entry-" + tag, side: "BUY", quantity: "10", symbol: symbol,
	})
	if _, err := j.RecordFill(ctx, terminalFill(o, "10", "70000")); err != nil {
		t.Fatalf("RecordFill(%s): %v", tag, err)
	}
	p := currentPosition(t, j, o)
	if _, err := j.OpenExitState(ctx, ExitStateSeed{PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000"}); err != nil {
		t.Fatalf("OpenExitState(%s): %v", tag, err)
	}
	return p.ID
}

// a206Flat 은 발의 없는 판정 행 하나를 씀(기준선 위 · high water 아래 — 상태가 안 바뀌는 관측).
func a206Flat(t *testing.T, j *Journal, positionID, observation string) {
	t.Helper()
	snap, rec := ratchetSnapshotForState(t, exitStateOf(t, j, positionID), observation, "69000", "70000", "68000")
	if err := j.RecordExitJudgement(context.Background(), judgementForSnapshot(snap, rec)); err != nil {
		t.Fatalf("flat judgement %s: %v", observation, err)
	}
}

func a206LastEventID(t *testing.T, j *Journal, positionID string) int64 {
	t.Helper()
	var id int64
	if err := j.db.QueryRow(`SELECT max(id) FROM exit_events WHERE position_id=?`, positionID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func a206Close(t *testing.T, j *Journal, positionID string) {
	t.Helper()
	for _, stmt := range []string{
		`UPDATE positions SET state='CLOSED', quantity='0', closed_at='2026-03-31T00:00:00Z' WHERE id=?`,
		`UPDATE exit_states SET completed=1 WHERE position_id=?`,
	} {
		if _, err := j.db.Exec(stmt, positionID); err != nil {
			t.Fatalf("closing %s: %v", positionID, err)
		}
	}
}

// a206Fixture 는 대상 술어의 모든 갈래를 한 저널에 둠. targets 는 비워야 할 행 id 임(손으로 고른 정답 — 술어 SQL 을 쓰지 않음).
type a206Fixture struct {
	j         *Journal
	targets   map[int64]bool
	positions []string
	linked    string // 브로커 주문 번호 — orders 연결 결과 비교용
	compacts  int    // a207 압축 행 수(앵커 제외)
}

func newA206Fixture(t *testing.T) a206Fixture {
	t.Helper()
	j := exitFixture(t)
	f := a206Fixture{j: j, targets: map[int64]bool{}, linked: "broker-a206-l1"}

	// C: 닫힘. 판정 전용 행(대상) → 무장 c0(전체, 거절) → 무장 c1(전체, 거절) → c2 · c3 압축(a207).
	// 앵커 후보가 c0 · c1 둘이고 D8 은 압축 행 앞의 가장 큰 id(c1)만 앵커로 침 — c0 은 대상, c1 · c2 · c3 은 남김.
	// c1 을 전체 증거로 쓰게 하려고 그 한 번만 a207 반복 술어가 거짓을 돌려주게 함(조회 실패 → 전체 증거, a207 D1b 와 같은 주입).
	c := a206Position(t, j, "c", "005930")
	a206Flat(t, j, c, "obs-c")
	f.targets[a206LastEventID(t, j, c)] = true
	for i, intent := range []string{"exit-a206-c0", "exit-a206-c1", "exit-a206-c2", "exit-a206-c3"} {
		if intent == "exit-a206-c1" {
			orig := repeatsARefusedArmTx
			repeatsARefusedArmTx = func(context.Context, *sql.Tx, string, int64, string, string) (bool, error) { return false, nil }
			_, _, err := a207Arm(t, j, c, "obs-c1", "67900", intent)
			repeatsARefusedArmTx = orig
			if err != nil {
				t.Fatal(err)
			}
		} else if _, _, err := a207Arm(t, j, c, "obs-c"+string(rune('0'+i)), "67900", intent); err != nil {
			t.Fatal(err)
		}
		if i < 3 {
			a207Refuse(t, j, c, intent)
		}
	}
	for _, intent := range []string{"exit-a206-c2", "exit-a206-c3"} {
		if row := a207RowOf(t, j, intent); !row.compact() {
			t.Fatalf("fixture: %s = %+v, want an a207 compact row", intent, row)
		}
		f.compacts++
	}
	for _, intent := range []string{"exit-a206-c0", "exit-a206-c1"} {
		if row := a207RowOf(t, j, intent); !row.full() {
			t.Fatalf("fixture: %s = %+v, want full evidence", intent, row)
		}
	}
	f.targets[a207RowOf(t, j, "exit-a206-c0").id] = true

	// L: 닫힘. 판정 전용 행(대상) → 무장 l1 은 브로커 주문 번호를 가진 attempt 에 연결 — 남김.
	l := a206Position(t, j, "l", "000660")
	a206Flat(t, j, l, "obs-l0")
	f.targets[a206LastEventID(t, j, l)] = true
	if _, _, err := a207Arm(t, j, l, "obs-l1", "67900", "exit-a206-l1"); err != nil {
		t.Fatal(err)
	}
	insertIntent(t, j, "exit-a206-l1")
	insertAttemptWithBrokerOrder(t, j, "attempt-a206-l1", "exit-a206-l1", f.linked)
	// 같은 intent 에 브로커 번호 없는 attempt 가 하나 더 있어도 「하나라도 연결이면 남김」 임.
	if _, err := j.db.Exec(`INSERT INTO mutation_attempts (id, intent_id, kind, state, attempt_no, fingerprint, recorded_at)
		VALUES ('attempt-a206-l1b', 'exit-a206-l1', 'PLACE', 'RECORDED', 2, 'fp', '2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}

	// U: 닫힘. 무장 u1 의 attempt 에 브로커 주문 번호가 없음 — 연결이 아니므로 대상.
	// 무장 전에 「작업 주문 때문에 무장 보류」 판정 행도 하나 둠 — arm_suppressed_reason 만 남기면 리더가 손상으로 읽음.
	u := a206Position(t, j, "u", "035420")
	held, heldRec := ratchetSnapshotForState(t, exitStateOf(t, j, u), "obs-u0", "67850", "70000", "68000")
	heldJudgement := judgementForSnapshot(held, heldRec)
	heldJudgement.ArmSuppressedReason = ArmSuppressedWorkingOrder
	if err := j.RecordExitJudgement(context.Background(), heldJudgement); err != nil {
		t.Fatal(err)
	}
	f.targets[a206LastEventID(t, j, u)] = true
	if _, _, err := a207Arm(t, j, u, "obs-u1", "67900", "exit-a206-u1"); err != nil {
		t.Fatal(err)
	}
	f.targets[a207RowOf(t, j, "exit-a206-u1").id] = true
	insertIntent(t, j, "exit-a206-u1")
	insertAttempt(t, j, "attempt-a206-u1", "exit-a206-u1")

	// O: 열림. 판정 전용 행 · 무장 행 모두 남김.
	o := a206Position(t, j, "o", "051910")
	a206Flat(t, j, o, "obs-o0")
	if _, _, err := a207Arm(t, j, o, "obs-o1", "67900", "exit-a206-o1"); err != nil {
		t.Fatal(err)
	}

	// K: CLOSED 이지만 exit 상태가 completed 0, Q: completed 1 이지만 OPEN — 둘 다 「닫힘」 조건의 반쪽뿐이라 남김.
	k := a206Position(t, j, "k", "066570")
	a206Flat(t, j, k, "obs-k0")
	q := a206Position(t, j, "q", "207940")
	a206Flat(t, j, q, "obs-q0")

	for _, id := range []string{c, l, u} {
		a206Close(t, j, id)
	}
	if _, err := j.db.Exec(`UPDATE positions SET state='CLOSED', quantity='0', closed_at='2026-03-31T00:00:00Z' WHERE id=?`, k); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE exit_states SET completed=1 WHERE position_id=?`, q); err != nil {
		t.Fatal(err)
	}
	f.positions = []string{c, l, u, o, k, q}
	return f
}

// a206Evidence 는 행 id → 증거 20 열의 값(문자열화, NULL 은 표지)임.
func a206Evidence(t *testing.T, db *sql.DB) map[int64][]string {
	t.Helper()
	rows, err := db.Query(`SELECT id, ` + strings.Join(a206EvidenceColumns, ", ") + ` FROM exit_events`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[int64][]string{}
	for rows.Next() {
		var id int64
		vals := make([]sql.NullString, len(a206EvidenceColumns))
		dest := []any{&id}
		for i := range vals {
			dest = append(dest, &vals[i])
		}
		if err := rows.Scan(dest...); err != nil {
			t.Fatal(err)
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = "<null>"
			if v.Valid {
				row[i] = "=" + v.String
			}
		}
		out[id] = row
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func a206AllNull(row []string) bool {
	for _, v := range row {
		if v != "<null>" {
			return false
		}
	}
	return true
}

// 행 자체 열(id · action · intent · 가격 · 시각 · 세대)은 정비 전후 같아야 함.
func a206Skeleton(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(`SELECT id || '|' || position_id || '|' || coalesce(action,'') || '|' || coalesce(proposed_intent_id,'') ||
		'|' || coalesce(observed_price,'') || '|' || coalesce(high_water,'') || '|' || coalesce(baseline_after,'') || '|' ||
		coalesce(level_after,'') || '|' || created_at || '|' || coalesce(lifecycle_generation,'') FROM exit_events ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

// D6 (1) · (3) — 대상 술어: 닫힌 · 미연결 행만 비고, 열린 포지션 · 연결 행 · 훅 행 · a207 압축 행 · 앵커는 바이트 그대로.
// 두 번째 실행은 0 행이고 압축 행 · 앵커는 여전히 남음.
func TestA206CompactionEmptiesOnlyClosedUnlinkedEvidence(t *testing.T) {
	f := newA206Fixture(t)
	j, ctx := f.j, context.Background()
	before := a206Evidence(t, j.db)
	skeleton := a206Skeleton(t, j.db)

	ro, err := OpenReadOnly(ctx, ReadOnlyOptions{Path: j.Path()})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := ro.ExitEvidenceCompactionPlan(ctx)
	_ = ro.Close()
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if plan.Targets != int64(len(f.targets)) || plan.Kept != int64(f.compacts+1) || plan.SnapshotBytes <= 0 {
		t.Fatalf("plan = %+v, want %d targets, %d kept (compact rows + anchor), snapshot bytes > 0", plan, len(f.targets), f.compacts+1)
	}

	report, err := j.CompactClosedExitEvidence(ctx)
	if err != nil {
		t.Fatalf("CompactClosedExitEvidence: %v", err)
	}
	if report.Emptied != int64(len(f.targets)) || report.Remaining != 0 || report.RowsAfter != report.Rows {
		t.Fatalf("report = %+v, want %d emptied, 0 remaining, rows unchanged", report, len(f.targets))
	}
	after := a206Evidence(t, j.db)
	for id, row := range before {
		switch {
		case f.targets[id]:
			if a206AllNull(row) || !a206AllNull(after[id]) {
				t.Errorf("target row %d: before all-null=%v, after %v — want evidence before and all 20 NULL after",
					id, a206AllNull(row), after[id])
			}
		case !reflect.DeepEqual(row, after[id]):
			t.Errorf("row %d is not a target but its evidence changed", id)
		}
	}
	if got := a206Skeleton(t, j.db); !reflect.DeepEqual(got, skeleton) {
		t.Fatal("row identity, action, intent, prices or times changed — only the evidence columns may change")
	}

	// 백업은 쓰기 전 상태를 들고 있음(spec 「적용 전 백업」).
	if info, err := os.Stat(report.Backup); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("backup %s: %v (mode %v), want a 0600 file", report.Backup, err, info)
	}
	backup, err := sql.Open("sqlite", "file:"+report.Backup+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer backup.Close()
	if got := a206Evidence(t, backup); !reflect.DeepEqual(got, before) {
		t.Fatal("the backup does not hold the pre-compaction evidence")
	}

	// D6 (3) 멱등: 두 번째 실행은 0 행, 압축 행 · 앵커 그대로.
	// 백업 이름이 시각이라 같은 순간의 두 번째 백업은 VACUUM INTO 가 거절함 — 시계를 1 초 옮김.
	j.clk.(*clock.Fake).Advance(time.Second)
	second, err := j.CompactClosedExitEvidence(ctx)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if second.Emptied != 0 || second.Targets != 0 || second.Kept != int64(f.compacts+1) {
		t.Fatalf("second run = %+v, want 0 emptied, 0 targets, the compact rows and anchor still kept", second)
	}
	if got := a206Evidence(t, j.db); !reflect.DeepEqual(got, after) {
		t.Fatal("the second run changed evidence")
	}
}

// D6 (2) — 정비 뒤 세 리더가 오류 없이 같은 연결 결과를 냄. 비운 행은 legacy_event 로 읽힘.
func TestA206ReadersAgreeAfterCompaction(t *testing.T) {
	f := newA206Fixture(t)
	j, ctx := f.j, context.Background()
	scopes := []BrokerOrderScope{{BrokerOrderID: f.linked, AccountRef: "acct-1", Market: "us", TradingDay: "2026-03-30"}}

	read := func() (map[string][]ExitEvent, []AccountExitEvent, []BrokerOrderExitLink) {
		t.Helper()
		history := map[string][]ExitEvent{}
		for _, id := range f.positions {
			events, err := j.ExitEvents(ctx, id)
			if err != nil {
				t.Fatalf("ExitEvents(%s): %v", id, err)
			}
			history[id] = events
		}
		ro, err := OpenReadOnly(ctx, ReadOnlyOptions{Path: j.Path()})
		if err != nil {
			t.Fatal(err)
		}
		defer ro.Close()
		account, err := ro.AccountExitEvents(ctx, "acct-1", 1000)
		if err != nil {
			t.Fatalf("AccountExitEvents: %v", err)
		}
		links, err := ro.BrokerOrderExitLinks(ctx, scopes)
		if err != nil {
			t.Fatalf("BrokerOrderExitLinks: %v", err)
		}
		return history, account, links
	}

	historyBefore, accountBefore, linksBefore := read()
	if len(linksBefore) != 1 || linksBefore[0].Ambiguous || linksBefore[0].Event.Evaluation.Effective.Snapshot == nil {
		t.Fatalf("fixture: links = %+v, want one unambiguous link carrying a snapshot", linksBefore)
	}
	if _, err := j.CompactClosedExitEvidence(ctx); err != nil {
		t.Fatal(err)
	}
	historyAfter, accountAfter, linksAfter := read()

	if !reflect.DeepEqual(linksAfter, linksBefore) {
		t.Fatalf("orders links changed:\nbefore %+v\nafter  %+v", linksBefore, linksAfter)
	}
	legacy := func(e ExitEvent) bool {
		ev := e.Evaluation
		return ev.Effective.Snapshot == nil && ev.Recomputed.Snapshot == nil && ev.Saved.Snapshot == nil &&
			ev.Effective.UnknownReason == "legacy_event" && ev.Recomputed.UnknownReason == "legacy_event"
	}
	for id, before := range historyBefore {
		after := historyAfter[id]
		if len(after) != len(before) {
			t.Fatalf("ExitEvents(%s): %d rows after, %d before", id, len(after), len(before))
		}
		for i := range before {
			switch {
			case f.targets[before[i].ID]:
				if !legacy(after[i]) {
					t.Errorf("emptied row %d reads as %+v, want legacy_event", before[i].ID, after[i].Evaluation)
				}
			case !reflect.DeepEqual(after[i], before[i]):
				t.Errorf("kept row %d reads differently after the compaction", before[i].ID)
			}
		}
	}
	if len(accountAfter) != len(accountBefore) {
		t.Fatalf("AccountExitEvents: %d rows after, %d before", len(accountAfter), len(accountBefore))
	}
	for i := range accountBefore {
		b, a := accountBefore[i], accountAfter[i]
		if a.Event.Evaluation.Effective.UnknownReason == "invalid_event_evidence" {
			t.Errorf("account history row %d reads as invalid evidence after the compaction", a.Event.ID)
		}
		switch {
		case f.targets[b.Event.ID]:
			if !legacy(a.Event) {
				t.Errorf("account history row %d = %+v, want legacy_event", b.Event.ID, a.Event.Evaluation)
			}
		case !reflect.DeepEqual(a, b):
			t.Errorf("account history row %d changed", b.Event.ID)
		}
	}
}

// 앵커가 없는 압축 행이 있으면 제외 집합에 NULL 이 생김. 그 NULL 이 NOT IN 을 NULL 로 만들어 대상 전체를 삼키면 안 됨.
// 앵커 후보 둘(c0 · c1)의 recomputed JSON 만 지워 앵커 없음을 만듦 — c1 은 부분 증거가 되어 대상에 더해짐(c0 은 원래 대상).
func TestA206AnAnchorlessCompactRowDoesNotSwallowTheTargets(t *testing.T) {
	f := newA206Fixture(t)
	j, ctx := f.j, context.Background()
	for _, intent := range []string{"exit-a206-c0", "exit-a206-c1"} {
		if _, err := j.db.Exec(`UPDATE exit_events SET recomputed_snapshot_json=NULL WHERE id=?`, a207RowOf(t, j, intent).id); err != nil {
			t.Fatal(err)
		}
	}
	report, err := j.CompactClosedExitEvidence(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if report.Kept != int64(f.compacts) || report.Emptied != int64(len(f.targets)+1) {
		t.Fatalf("report = %+v, want the %d compact rows kept and %d rows emptied (the targets and the ex-anchor c1)",
			report, f.compacts, len(f.targets)+1)
	}
}
