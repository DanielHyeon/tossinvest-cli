package console

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

func ratchetViewSnapshot(t *testing.T, positionID string, generation int64, quantity,
	observationID, observed, high, baseline, taken string, level exitpolicy.Level,
) (exitpolicy.ExitLineSnapshot, exitpolicy.RecoveryPolicyDefinition) {
	t.Helper()
	in := exitpolicy.RatchetSnapshotInput{
		Context: exitpolicy.SnapshotContext{PositionID: positionID, PositionGeneration: generation,
			ObservationID: observationID, RemainingQuantity: quantity},
		Input: exitpolicy.RatchetInput{
			Entry: "70000", InitialStop: "68000", ObservedPrice: observed, HighWater: high,
			Baseline: baseline, RealBreakeven: "70010", TakenRatioTotal: taken, Level: level,
		},
	}
	line, err := exitpolicy.EvaluateRatchetSnapshot(in)
	if err != nil {
		t.Fatal(err)
	}
	return line, exitpolicy.NewRatchetRecoveryPolicy(in)
}

func storedViewSnapshotJSON(t *testing.T, line exitpolicy.ExitLineSnapshot,
	recovery exitpolicy.RecoveryPolicyDefinition, observedAt string,
) string {
	t.Helper()
	digestBytes, err := json.Marshal(struct {
		Line   exitpolicy.ExitLineSnapshot         `json:"line"`
		Policy exitpolicy.RecoveryPolicyDefinition `json:"policy"`
	}{Line: line, Policy: recovery})
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(digestBytes)
	stored := journal.StoredExitSnapshot{
		Line: line, RecoveryPolicy: recovery, ObservationSource: "official.quote",
		ObservedAt: observedAt, OutputDigest: fmt.Sprintf("sha256:%x", digest[:]),
	}
	raw, err := json.Marshal(stored)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func nullableViewValue(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func writeViewSnapshot(t *testing.T, path string, line exitpolicy.ExitLineSnapshot,
	recovery exitpolicy.RecoveryPolicyDefinition, observedAt string,
) {
	t.Helper()
	rung := any(nil)
	if line.ActiveRung >= 0 {
		rung = line.ActiveRung
	}
	execRawArgs(t, path, `
		UPDATE exit_states SET
		 snapshot_status=?, policy_id=?, policy_version=?, policy_digest=?, snapshot_id=?,
		 decision_id=?, observation_id=?, position_generation=?, next_target=?, next_protection=?,
		 last_observation_source=?, last_observed_at=?, snapshot_action=?, snapshot_ratio=?,
		 projected_quantity=?, state_only=?, suppressed_reason=?, effective_snapshot_json=?,
		 baseline_price=?, high_water=?, ratchet_level=?, active_rung=?, updated_at=?
		 WHERE position_id=?`,
		journal.SnapshotStatusEvaluated, line.Policy.ID, line.Policy.Version, line.Policy.Digest,
		line.SnapshotID, line.DecisionID, line.ObservationID, line.PositionGeneration,
		line.NextTarget, line.NextProtection, "official.quote", observedAt, string(line.Action),
		line.Ratio, line.ProjectedQuantity, boolNumber(line.StateOnly), line.Suppressed,
		storedViewSnapshotJSON(t, line, recovery, observedAt), line.CurrentProtection,
		line.HighWater, string(line.RatchetLevel), rung, observedAt, line.PositionID)
}

func boolNumber(value bool) int {
	if value {
		return 1
	}
	return 0
}

func execRawArgs(t *testing.T, path, statement string, args ...any) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(context.Background(), statement, args...); err != nil {
		t.Fatal(err)
	}
}

func TestPositionsRenderCanonicalExitLineFixtures(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('pos-stale','123-45-678901','kr','111111',1,'d-1','OPEN','4','70000','2026-07-27T00:30:00Z'),
		       ('pos-unknown','123-45-678901','kr','222222',1,'d-1','OPEN','3','70000','2026-07-27T00:30:00Z'),
		       ('pos-one','123-45-678901','kr','333333',1,'d-1','OPEN','1','70000','2026-07-27T00:30:00Z');
		INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
		 high_water,ratchet_level,taken_ratio_total,completed,updated_at)
		VALUES ('pos-stale','RATCHET','70000','68000','2000','68000','70000','NONE','0',0,'2026-07-27T00:59:00Z'),
		       ('pos-unknown','RATCHET','70000','68000','2000','68000','70000','NONE','0',0,'2026-07-27T00:59:00Z'),
		       ('pos-one','RATCHET','70000','68000','2000','68000','70000','NONE','0',0,'2026-07-27T00:59:00Z');`)

	complete, completeRecovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-complete",
		"70500", "70000", "68000", "0", exitpolicy.LevelNone)
	writeViewSnapshot(t, h.journal, complete, completeRecovery, "2026-07-27T00:59:45Z")
	stale, staleRecovery := ratchetViewSnapshot(t, "pos-stale", 1, "4", "obs-stale",
		"70500", "70000", "68000", "0", exitpolicy.LevelNone)
	writeViewSnapshot(t, h.journal, stale, staleRecovery, "2026-07-27T00:58:00Z")
	one, oneRecovery := ratchetViewSnapshot(t, "pos-one", 1, "1", "obs-one",
		"72000", "72000", "68000", "0", exitpolicy.LevelNone)
	if !one.StateOnly || one.ProjectedQuantity != "0" {
		t.Fatalf("one-share fixture is not state-only: %+v", one)
	}
	writeViewSnapshot(t, h.journal, one, oneRecovery, "2026-07-27T00:59:50Z")

	h.authenticate(t)
	page := h.page(t, "/positions")
	for _, want := range []string{
		"평가 완료", "현재 보호선", "다음 익절", "다음 보호선", "예상 수량",
		"오래된 평가", "평가 시각이 표시 허용 범위를 지났다", "근거 없음",
		"이전 원장에는 exit snapshot 근거가 없다", "중간 매도 없음 · 보호선 승격",
		"최종 익절·손절 시 1주 전량", "official.quote", complete.DecisionID,
		`href="/optimization?category=position-management"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("positions page lacks %q", want)
		}
	}
	staleRow := positionHTMLRow(t, page, "111111")
	if strings.Contains(staleRow, stale.CurrentProtection) || strings.Contains(staleRow, stale.NextTarget) {
		t.Error("stale row exposes actionable exit values instead of em dashes")
	}
}

func TestOrdersJoinExitEvidenceOnlyByAttemptIntentLineage(t *testing.T) {
	path := t.TempDir() + "/journal.db"
	seedEngineJournal(t, path, journalFixture+`
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('pos-order-a043','123-45-678901','kr','005930',2,'d-1','OPEN','1','70000','2026-07-27T00:30:00Z');
		INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,time_in_force,
		 quantity,price,currency,source,fingerprint,notes)
		VALUES ('exit-intent-a043','2026-07-27T00:59:40Z','kr','2026-07-27','123-45-678901','005930',
		 'SELL','MARKET','DAY','1',NULL,'KRW','exit-policy','fp-exit-a043','');`)

	j, err := journal.Open(context.Background(), journal.Options{Path: path,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	state, err := j.OpenExitState(context.Background(), journal.ExitStateSeed{
		PositionID: "pos-order-a043", EntryPrice: "70000", InitialStop: "68000",
	})
	if err != nil {
		t.Fatal(err)
	}
	line, recovery := ratchetViewSnapshot(t, state.PositionID, state.PositionGeneration, "1",
		"obs-order-a043", "67900", "70000", "68000", "0", exitpolicy.LevelNone)
	judgement := journal.ExitJudgement{
		PositionID: line.PositionID, Snapshot: line, RecoveryPolicy: recovery,
		ObservationSource: "official.quote", ObservedAt: time.Date(2026, 7, 27, 0, 59, 40, 0, time.UTC),
		Provenance: journal.ExitDecisionProvenance{ObservationID: line.ObservationID,
			SnapshotID: line.SnapshotID, DecisionID: line.DecisionID, Policy: line.Policy},
		ObservedPrice: line.ObservedPrice, HighWater: line.HighWater,
		Baseline: line.CurrentProtection, RatchetLevel: string(line.RatchetLevel), ActiveRung: line.ActiveRung,
		Proposal: &journal.ExitProposal{Action: string(line.Action), Level: line.Level,
			IntentID: "exit-intent-a043", Provenance: journal.ExitDecisionProvenance{
				ObservationID: line.ObservationID, SnapshotID: line.SnapshotID,
				DecisionID: line.DecisionID, Policy: line.Policy}},
	}
	if err := j.RecordExitJudgement(context.Background(), judgement); err != nil {
		t.Fatal(err)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	execRaw(t, path, `
		INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,broker_order_id,fingerprint,recorded_at)
		VALUES ('exit-attempt-a043','exit-intent-a043','PLACE','RECORDED',1,'broker-exit-a043','fp','2026-07-27T00:59:41Z');`)

	linked := livePlainOrder("broker-exit-a043", "005930")
	linked.OrderedAt = "2026-07-27T00:59:41Z"
	unlinked := livePlainOrder("same-symbol-unlinked", "005930")
	unlinked.OrderedAt = linked.OrderedAt
	reader := &countingOrders{lists: OrdersReading{AccountRef: "123-45-678901", Open: []OrderRecord{linked, unlinked}}}
	h := newOrdersHarness(t, reader, func(o *Options) { o.JournalPath = path })
	h.authenticate(t)
	page := body(t, h.get(t, "/orders"))

	linkedRow := orderHTMLRow(t, page, "broker-exit-a043")
	for _, want := range []string{"보호선 이탈 · 전량 청산", "관측가", "현재 보호선", line.DecisionID,
		"exit-attempt-a043", `href="/optimization?category=exit-protection"`} {
		if !strings.Contains(linkedRow+page, want) {
			t.Errorf("linked order evidence lacks %q", want)
		}
	}
	if row := orderHTMLRow(t, page, "same-symbol-unlinked"); !strings.Contains(row, "근거 미연결") ||
		strings.Contains(row, line.DecisionID) {
		t.Errorf("same-symbol/time order was fuzzy-linked: %s", row)
	}
}

func positionHTMLRow(t *testing.T, page, symbol string) string {
	t.Helper()
	start := strings.Index(page, `class="position-row" data-symbol="`+symbol+`"`)
	if start < 0 {
		t.Fatalf("position row %s not found", symbol)
	}
	end := strings.Index(page[start:], "</tr>")
	if end < 0 {
		t.Fatalf("position row %s does not close", symbol)
	}
	return page[start : start+end]
}

func orderHTMLRow(t *testing.T, page, orderID string) string {
	t.Helper()
	needle := "<code>" + orderID + "</code>"
	at := strings.Index(page, needle)
	if at < 0 {
		t.Fatalf("order row %s not found", orderID)
	}
	start := strings.LastIndex(page[:at], `<tr class="order-row"`)
	end := strings.Index(page[at:], "</tr>")
	if start < 0 || end < 0 {
		t.Fatalf("order row %s boundaries not found", orderID)
	}
	return page[start : at+end]
}
