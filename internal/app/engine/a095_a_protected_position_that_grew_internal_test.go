package engine

// a095 tasks 3.1~3.4 — 보호 중인 포지션의 수량 증가(engine-safety 델타 10판 「보호 중인 포지션의 수량 증가는 새 최대 수량마다
// 보고된다」, design D2 (i)(ii)).
//
// 엔진 개설 포지션의 비교 기준은 원장 조정의 순증 Σ(new_quantity − prev_quantity)임 — 대사 수렴이 계좌의 체결로 설명되지 않는
// 수량을 조정 행으로 적고 투영을 옮김(converge.go:209-227 · position_adjustments.go:304-350). 등급은 normal(Q4): 손절은 발동
// 시점의 투영 수량 전량을 청산하므로 증가분도 보호받음(review §4.4).

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// engineOpened 는 진입 결정으로 열린 포지션을 원장에 직접 세움 — 체결 경로 전체를 태우지 않고 「결정이 정당화하는 인스턴스,
// 조정 이력 없음」이라는 사실만 만듦. 결정 행이 없으므로 이 연결에서만 외래 키 검사를 끔.
func (f *a095Fixture) engineOpened(symbol, quantity string) string {
	f.t.Helper()
	ctx := context.Background()
	conn, err := f.sideDB().Conn(ctx)
	if err != nil {
		f.t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		f.t.Fatal(err)
	}
	id := journal.PositionID(a095Account, "kr", symbol, 1)
	if _, err := conn.ExecContext(ctx, `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state, quantity, avg_price, opened_at)
		VALUES (?, ?, 'kr', ?, 1, 'decision-a095', ?, ?, '70000', '2026-09-30T00:00:00Z')`,
		id, a095Account, symbol, journal.PositionOpen, quantity); err != nil {
		f.t.Fatalf("seeding the engine-opened position: %v", err)
	}
	f.holds(symbol, quantity, 70000)
	return id
}

func (f *a095Fixture) grownReports(positionID string) []obs.Event {
	var out []obs.Event
	for _, e := range f.capture.events {
		if strings.HasSuffix(e.Key, "|grown|"+positionID) {
			out = append(out, e)
		}
	}
	return out
}

func TestA095AnEngineOpenedPositionThatGrewIsReported(t *testing.T) {
	f := newA095(t)
	id := f.engineOpened("005930", "10")

	f.cycle()
	if got := f.grownReports(id); len(got) != 0 {
		t.Fatalf("reports = %+v with no adjustment; the fills explain every share", got)
	}

	f.holds("005930", "15", 70000) // 손으로 5주 더 삼 — 수렴이 +5 조정을 적음
	f.cycle()
	got := f.grownReports(id)
	if len(got) != 1 {
		t.Fatalf("reports = %+v, want one: an engine-opened position that grew is checked too (결정 (3)(i))", got)
	}
	if got[0].Type != obs.EventExitPositionUnmanaged || obs.SeverityOf(got[0].Type) != obs.SeverityNormal {
		t.Errorf("report kind = %s, want the normal kind (Q4)", got[0].Type)
	}
	if q := got[0].Fields[obs.FieldQuantity]; q != "15" {
		t.Errorf("reported quantity = %v, want 15", q)
	}

	f.cycle() // 같은 수량의 반복
	f.holds("005930", "12", 70000)
	f.cycle() // 감소
	if n := len(f.grownReports(id)); n != 1 {
		t.Errorf("reports = %d after a repeat and a decrease, want still 1", n)
	}
	f.holds("005930", "20", 70000)
	f.cycle() // 새 최대
	if n := len(f.grownReports(id)); n != 2 {
		t.Errorf("reports = %d after a new maximum, want 2", n)
	}
	f.assertNoCriticalConsequence(t)
}

// 순증이 0 인 조정 이력(+5 뒤 −5) — 「0 보다 크다」의 경계. off-by-one(≥ 0) 변이가 여기서 죽음.
func TestA095AnAdjustmentHistoryThatNetsToZeroIsNotAnIncrease(t *testing.T) {
	f := newA095(t)
	id := f.engineOpened("005930", "10")
	f.exec(`INSERT INTO position_adjustments (id, position_id, kind, expected_prev_quantity, prev_quantity,
		  new_quantity, broker_as_of, created_at) VALUES
		('a095-up', '` + id + `', 'EXTERNAL', '10', '10', '15', '2026-09-30T00:10:00Z', '2026-09-30T00:10:00Z'),
		('a095-down', '` + id + `', 'EXTERNAL', '15', '15', '10', '2026-09-30T00:20:00Z', '2026-09-30T00:20:00Z')`)
	f.cycle()
	if got := f.grownReports(id); len(got) != 0 {
		t.Errorf("reports = %+v; an increase that was sold back nets to zero", got)
	}
}

// 3.4 — 475150 원장 순서: 편입 2 → 3 · 4 · 5 · 8 · 26 · 32. 여섯 수량이 각각 한 번 보고되고 32 가 전해짐.
func TestA095TheAdoptedGrowthReplayReportsEveryNewMaximum(t *testing.T) {
	f := newA095(t)
	f.holds("475150", "2", 50000)
	if cycle := f.cycle(); cycle.Adopted != 1 {
		t.Fatalf("adopted = %d (%v)", cycle.Adopted, cycle.Err)
	}
	id := f.position("475150").ID
	for _, q := range []string{"3", "4", "5", "8", "26", "32"} {
		f.holds("475150", q, 50000)
		f.cycle()
	}
	f.cycle() // 32 의 반복 — 새 최대가 아님
	got := f.grownReports(id)
	if len(got) != 6 {
		t.Fatalf("reports = %d, want 6 — one per new maximum", len(got))
	}
	if last := got[len(got)-1].Fields[obs.FieldQuantity]; last != "32" {
		t.Errorf("the last report says %v, want 32", last)
	}
	for i, r := range got { // 고정은 마지막 보고까지 — 첫 보고만 보면 기준이 움직여도 통과함(독립 리뷰 P2-2)
		if adopted := r.Fields["adopted_quantity"]; adopted != "2" {
			t.Errorf("report %d: adopted_quantity = %v, want the frozen 2", i, adopted)
		}
	}
	f.assertNoCriticalConsequence(t)
}

// 3.1 — checkExternalIncrease B2 무변화: 편입 기록 조회 오류는 조용히 반환하고 무관리 보고로 보내지 않음(R2-B2 삭제).
func TestA095AnAdoptionLookupFailureStaysSilent(t *testing.T) {
	f := newA095(t)
	f.holds("005930", "10", 70000)
	if cycle := f.cycle(); cycle.Adopted != 1 {
		t.Fatalf("adopted = %d", cycle.Adopted)
	}
	id := f.position("005930").ID
	before := len(f.capture.events)
	f.exec(`ALTER TABLE position_adoptions RENAME TO a095_hidden_adoptions`)
	f.holds("005930", "15", 70000)
	f.cycle()
	for _, e := range f.capture.events[before:] {
		if strings.HasSuffix(e.Key, id) {
			t.Errorf("event %s %q after a lookup failure; B2 returns silently", e.Type, e.Key)
		}
	}
}

// 음수 순증 — 손으로 일부를 판 엔진 개설 포지션은 증가가 아님(codex 교차 리뷰 보강).
func TestA095APartialSaleByHandIsNotAnIncrease(t *testing.T) {
	f := newA095(t)
	id := f.engineOpened("005930", "10")
	f.holds("005930", "6", 70000) // 수렴이 −4 조정을 적음
	f.cycle()
	if got := f.grownReports(id); len(got) != 0 {
		t.Errorf("reports = %+v; a sale by hand nets negative", got)
	}
}

// 교차 인스턴스 — 앞 인스턴스의 조정은 뒤 인스턴스의 순증에 들지 않음(PositionAdjustments 는 인스턴스 id 로 거름).
func TestA095AnEarlierInstancesAdjustmentsDoNotCount(t *testing.T) {
	f := newA095(t)
	ctx := context.Background()
	old := journal.PositionID(a095Account, "kr", "005930", 1)
	conn, err := f.sideDB().Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	for _, q := range []string{
		`PRAGMA foreign_keys = OFF`,
		`INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state, quantity, avg_price,
		   opened_at, closed_at) VALUES ('` + old + `', '` + a095Account + `', 'kr', '005930', 1, 'decision-old', '` +
			journal.PositionClosed + `', '0', '70000', '2026-09-29T00:00:00Z', '2026-09-29T06:00:00Z')`,
		`INSERT INTO position_adjustments (id, position_id, kind, expected_prev_quantity, prev_quantity, new_quantity,
		   broker_as_of, created_at) VALUES ('a095-old-up', '` + old + `', 'EXTERNAL', '10', '10', '15',
		   '2026-09-29T01:00:00Z', '2026-09-29T01:00:00Z')`,
	} {
		if _, err := conn.ExecContext(ctx, q); err != nil {
			t.Fatalf("seeding: %v", err)
		}
	}
	id := journal.PositionID(a095Account, "kr", "005930", 2)
	if _, err := conn.ExecContext(ctx, `INSERT INTO positions
		  (id, account_ref, market, symbol, instance_seq, entry_decision_id, state, quantity, avg_price, opened_at)
		VALUES (?, ?, 'kr', '005930', 2, 'decision-new', ?, '10', '70000', '2026-09-30T00:00:00Z')`,
		id, a095Account, journal.PositionOpen); err != nil {
		t.Fatalf("seeding the new instance: %v", err)
	}
	f.holds("005930", "10", 70000)
	f.cycle()
	if got := f.grownReports(id); len(got) != 0 {
		t.Errorf("reports = %+v; the earlier instance's increase belongs to that instance", got)
	}
	if p := f.position("005930"); p.ID != id {
		t.Fatalf("arrangement: the current instance is %s, want %s", p.ID, id)
	}
}

// 새 요구의 SHALL NOT — 수량 증가 보고는 진입가 · 최초 손절 · 최초 위험 · 기준선을 바꾸지 않음(독립 리뷰 P2-2). 엔진 개설 포지션에
// exit state 를 두고 증가 전후 네 열을 비교함.
func TestA095AGrowthReportLeavesTheProtectionColumnsAlone(t *testing.T) {
	f := newA095(t)
	id := f.engineOpened("005930", "10")
	f.exec(`INSERT INTO exit_states (position_id, policy_kind, entry_price, initial_stop, initial_risk, baseline_price,
		  high_water, ratchet_level, updated_at)
		VALUES ('` + id + `', 'RATCHET', '70000', '68000', '2000', '68000', '70000', 'NONE', '2026-09-30T00:00:00Z')`)
	before, err := f.j.ExitState(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	f.holds("005930", "15", 70000)
	f.cycle()
	if len(f.grownReports(id)) != 1 {
		t.Fatalf("arrangement: no growth report (%+v)", f.capture.events)
	}
	after, err := f.j.ExitState(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if before.EntryPrice != after.EntryPrice || before.InitialStop != after.InitialStop ||
		before.InitialRisk != after.InitialRisk || before.Baseline != after.Baseline {
		t.Errorf("exit state moved %+v → %+v on a growth report", before, after)
	}
}
