package journal

// a192_order_terminal_release_test.go 는 종결 증명 해제의 장부 쪽 절반을 고정함(a192 design D5) —
// 증명 읽기, 같은 tx 안 재판정, 새 release cause. 픽스처는 합성 값만 씀.

import (
	"context"
	"strings"
	"testing"
)

func a192ConfirmOrder(t *testing.T, j *Journal, orderID, side string) {
	t.Helper()
	ctx := context.Background()
	attempt, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{
			ID: "intent-" + orderID, AccountRef: "acct-7", Market: "kr", TradingDay: "2026-03-30",
			Symbol: "SYN001", Side: side, OrderType: "LIMIT", Quantity: "1", Price: "100",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-" + orderID,
		},
		Kind: KindPlace, AttemptID: "attempt-" + orderID,
	})
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	if err := attempt.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := attempt.MarkAcked(ctx, orderID); err != nil {
		t.Fatal(err)
	}
	if err := attempt.Settle(ctx, StateConfirmed, ReasonBrokerAcknowledged, "acked"); err != nil {
		t.Fatal(err)
	}
}

func a192Scope(orderID, side string) FillSnapshotScope {
	return FillSnapshotScope{OrderID: orderID, AccountRef: "acct-7", Market: "kr", TradingDay: "2026-03-30",
		Symbol: "SYN001", Side: side}
}

func a192Record(t *testing.T, j *Journal, orderID, side, state, filled string) {
	t.Helper()
	if _, err := j.RecordFill(context.Background(), FillObservation{
		OrderID: orderID, Symbol: "SYN001", Market: "kr", TradingDay: "2026-03-30", Side: side,
		AccountRef: "acct-7", State: state, Terminal: true, Quantity: "1", FilledQuantity: filled,
		AveragePrice: "100", ObservedAt: "2026-03-30T00:30:00Z",
	}); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
}

func a192Row(t *testing.T, j *Journal) ReconcileState {
	t.Helper()
	state, entered, err := j.EnterReconcile(context.Background(), EnterReconcileRequest{
		AccountRef: "acct-7", Symbol: "SYN001", Cause: ReconcileCauseQuantityMismatch,
		Evidence: "order on SYN001 is not in the account's open list",
	})
	if err != nil || !entered {
		t.Fatalf("EnterReconcile: entered=%v err=%v", entered, err)
	}
	return state
}

// TestA192OrderTerminalReleaseCauseIsOneThisBuildWrites: 새 원인도 상수 집합(쓰기 제약)에 들어 있어야 함.
func TestA192OrderTerminalReleaseCauseIsOneThisBuildWrites(t *testing.T) {
	if !ValidReconcileReleaseCause(ReconcileReleaseOrderTerminalConfirmed) {
		t.Fatalf("%s must be a release cause this build writes", ReconcileReleaseOrderTerminalConfirmed)
	}
}

// TestA192TheReleaseRereadsTheProofInsideItsTransaction: 판정이 tx 안에서 읽은 증명을 받고, 참일 때만 그 행을
// 새 원인으로 해제함.
func TestA192TheReleaseRereadsTheProofInsideItsTransaction(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	a192ConfirmOrder(t, j, "o-a192-1", "SELL")
	watermark, err := j.MaxFillEventID(ctx)
	if err != nil {
		t.Fatalf("MaxFillEventID: %v", err)
	}
	row := a192Row(t, j)
	a192Record(t, j, "o-a192-1", "SELL", "CANCELLED", "0")

	req := OrderTerminalProofRequest{StateID: row.ID, Symbol: "SYN001", FillWatermark: watermark,
		Orders: []FillSnapshotScope{a192Scope("o-a192-1", "SELL")}}
	var seen OrderTerminalProof
	state, released, err := j.ReleaseReconcileOnOrderTerminalProof(ctx, req, "terminal proof (synthetic)",
		func(p OrderTerminalProof) bool { seen = p; return true })
	if err != nil || !released {
		t.Fatalf("release: released=%v err=%v", released, err)
	}
	if !seen.StateFound || seen.State.ID != row.ID || len(seen.Snapshots) != 1 || !seen.Snapshots[0].Found ||
		!seen.Snapshots[0].Record.Terminal || seen.FillAfterWatermark || seen.ReadErr != nil {
		t.Fatalf("proof read inside the transaction = %+v", seen)
	}
	if state.ReleaseCause != ReconcileReleaseOrderTerminalConfirmed {
		t.Fatalf("release cause = %q", state.ReleaseCause)
	}
	history, err := j.ReconcileStateHistory(ctx, "acct-7")
	if err != nil || len(history) != 1 || history[0].Active() ||
		history[0].ReleaseCause != ReconcileReleaseOrderTerminalConfirmed ||
		!strings.Contains(history[0].Evidence, "terminal proof (synthetic)") {
		t.Fatalf("history = %+v err=%v, want the row released with its evidence", history, err)
	}
	// 이미 풀린 행은 다시 풀리지 않음(id 고정 + released_at IS NULL).
	if _, again, err := j.ReleaseReconcileOnOrderTerminalProof(ctx, req, "again",
		func(p OrderTerminalProof) bool { return p.StateFound && p.State.Active() }); err != nil || again {
		t.Fatalf("second release: released=%v err=%v, want nothing", again, err)
	}
}

// TestA192TheReleaseKeepsTheRowWhenTheVerdictHolds: 판정이 거짓이면 오류 없이 해제하지 않음.
func TestA192TheReleaseKeepsTheRowWhenTheVerdictHolds(t *testing.T) {
	j, _ := openReservationJournal(t)
	row := a192Row(t, j)
	_, released, err := j.ReleaseReconcileOnOrderTerminalProof(context.Background(), OrderTerminalProofRequest{
		StateID: row.ID, Symbol: "SYN001"}, "evidence", func(OrderTerminalProof) bool { return false })
	if err != nil || released {
		t.Fatalf("held release: released=%v err=%v", released, err)
	}
	active, err := j.ActiveReconcileStates(context.Background())
	if err != nil || len(active) != 1 {
		t.Fatalf("active = %+v err=%v, want the row kept", active, err)
	}
}

// TestA192TheProofSeesFillsAfterTheWatermark: 워터마크 뒤 그 종목 체결만 셈.
func TestA192TheProofSeesFillsAfterTheWatermark(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	a192ConfirmOrder(t, j, "o-a192-buy", "BUY")
	before, err := j.MaxFillEventID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	row := a192Row(t, j)
	req := OrderTerminalProofRequest{StateID: row.ID, Symbol: "SYN001", FillWatermark: before}
	if p := j.ReadOrderTerminalProof(ctx, req); p.ReadErr != nil || p.FillAfterWatermark {
		t.Fatalf("proof before any fill = %+v", p)
	}
	a192Record(t, j, "o-a192-buy", "BUY", "FILLED", "1")
	if p := j.ReadOrderTerminalProof(ctx, req); !p.FillAfterWatermark {
		t.Fatalf("proof after a fill on the symbol = %+v, want the fill seen", p)
	}
	other := req
	other.Symbol = "SYNX"
	if p := j.ReadOrderTerminalProof(ctx, other); p.FillAfterWatermark {
		t.Fatalf("a fill on another symbol counted: %+v", p)
	}
	after, err := j.MaxFillEventID(ctx)
	if err != nil || after <= before {
		t.Fatalf("watermark %d → %d (err %v), want it to advance", before, after, err)
	}
}

// TestA192AnIncompleteScopeNeverReachesTheLegacyLookup: 범위가 불완전하면 스냅숏을 찾지 않음 — 같은 번호의
// legacy 빈 범위 행이 있어도.
func TestA192AnIncompleteScopeNeverReachesTheLegacyLookup(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	if _, err := j.db.ExecContext(ctx, `INSERT INTO fill_snapshots
		(order_id, symbol, market, state, terminal, fail_closed, quantity, filled_quantity, committed_at)
		VALUES ('o-a192-legacy', 'SYN001', 'kr', 'CANCELLED', 1, 0, '1', '0', '2026-03-30T00:00:00Z')`); err != nil {
		t.Fatalf("seeding a legacy blank-scope row: %v", err)
	}
	row := a192Row(t, j)
	incomplete := a192Scope("o-a192-legacy", "")
	p := j.ReadOrderTerminalProof(ctx, OrderTerminalProofRequest{StateID: row.ID, Symbol: "SYN001",
		Orders: []FillSnapshotScope{incomplete}})
	if p.ReadErr != nil || len(p.Snapshots) != 1 || p.Snapshots[0].Found {
		t.Fatalf("proof for an incomplete scope = %+v, want no snapshot", p)
	}
	// 대조: 같은 범위로 legacy 조회를 직접 하면 그 행이 잡힘 — 막는 것은 범위 검사임.
	if _, err := lookupFillSnapshotScoped(ctx, j.db, incomplete); err != nil {
		t.Fatalf("the legacy lookup itself should find the blank row: %v", err)
	}
}

// TestA192AReplacementTraceIsSeen: 그 번호를 겨눈 AMEND 시도가 있으면 Amended.
func TestA192AReplacementTraceIsSeen(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	a192ConfirmOrder(t, j, "o-a192-1", "SELL")
	row := a192Row(t, j)
	req := OrderTerminalProofRequest{StateID: row.ID, Symbol: "SYN001",
		Orders: []FillSnapshotScope{a192Scope("o-a192-1", "SELL")}}
	if p := j.ReadOrderTerminalProof(ctx, req); p.ReadErr != nil || p.Snapshots[0].Amended {
		t.Fatalf("proof before any amend = %+v", p)
	}
	if _, err := j.Prepare(ctx, PrepareRequest{
		Intent: Intent{ID: "intent-amend", AccountRef: "acct-7", Market: "kr", TradingDay: "2026-03-30",
			Symbol: "SYN001", Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "99",
			Currency: "KRW", Source: "engine", Fingerprint: "fp-amend"},
		Kind: KindAmend, AttemptID: "attempt-amend", TargetOrderID: "o-a192-1",
	}); err != nil {
		t.Fatalf("Prepare(AMEND): %v", err)
	}
	if p := j.ReadOrderTerminalProof(ctx, req); !p.Snapshots[0].Amended {
		t.Fatalf("proof after an amend = %+v, want the trace seen", p)
	}
}

// TestA192AReadFailureIsReportedInTheProof: 읽기 실패는 증명의 ReadErr 로 옴(판정이 유지로 읽음).
func TestA192AReadFailureIsReportedInTheProof(t *testing.T) {
	j, _ := openReservationJournal(t)
	row := a192Row(t, j)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	if p := j.ReadOrderTerminalProof(context.Background(), OrderTerminalProofRequest{StateID: row.ID}); p.ReadErr == nil {
		t.Fatalf("proof from a closed journal = %+v, want a read error", p)
	}
}

// TestA192TheProofCountsUnattributableAndForeignFills: 워터마크 뒤의 종목 없는 체결(legacy 관측)과 다른 계좌의 같은
// 종목 체결도 「뒤의 체결」로 셈 — 귀속을 증명할 수 없거나 계좌를 가리지 않는 보수 방향(코드 리뷰 1 #1).
func TestA192TheProofCountsUnattributableAndForeignFills(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	before, err := j.MaxFillEventID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	row := a192Row(t, j)
	req := OrderTerminalProofRequest{StateID: row.ID, Symbol: "SYN001", FillWatermark: before}
	for name, insert := range map[string]string{
		"blank symbol": `INSERT INTO fill_events (order_id, symbol, market, delta_quantity, cumulative_quantity, committed_at)
			VALUES ('o-a192-legacy', '', '', '1', '1', '2026-03-30T00:31:00Z')`,
		"another account": `INSERT INTO fill_events (order_id, account_ref, symbol, market, trading_day, side, delta_quantity, cumulative_quantity, committed_at)
			VALUES ('o-a192-other', 'acct-other', 'SYN001', 'kr', '2026-03-30', 'BUY', '1', '1', '2026-03-30T00:31:00Z')`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := j.db.ExecContext(ctx, `DELETE FROM fill_events WHERE id > ?`, before); err != nil {
				t.Fatal(err)
			}
			if p := j.ReadOrderTerminalProof(ctx, req); p.FillAfterWatermark {
				t.Fatalf("control: proof before the insert = %+v", p)
			}
			if _, err := j.db.ExecContext(ctx, insert); err != nil {
				t.Fatalf("seeding: %v", err)
			}
			if p := j.ReadOrderTerminalProof(ctx, req); p.ReadErr != nil || !p.FillAfterWatermark {
				t.Fatalf("proof after the %s fill = %+v, want it counted", name, p)
			}
		})
	}
}

// TestA192TheReleaseNeverOverwritesARowAlreadyReleased 는 장부 쪽 고정을 혼자 잼(Manager 변이 g8·g9): 판정이 늘 참이어도
// 이미 다른 원인(OPERATOR)으로 풀린 행은 다시 쓰지 않고(`released_at IS NULL`), 갱신된 행이 하나가 아니면 오류임
// (`RowsAffected == 1`). 해제 원인·시각·근거는 그대로 남아야 함.
func TestA192TheReleaseNeverOverwritesARowAlreadyReleased(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	row := a192Row(t, j)
	if _, released, err := j.ReleaseReconcile(ctx, ReleaseReconcileRequest{
		AccountRef: "acct-7", Symbol: "SYN001", Cause: ReconcileReleaseOperator,
		Evidence: "operator synthetic: verified", ExpectCause: ReconcileCauseQuantityMismatch,
	}); err != nil || !released {
		t.Fatalf("operator release: released=%v err=%v", released, err)
	}
	before, err := j.ReconcileStateHistory(ctx, "acct-7")
	if err != nil || len(before) != 1 || before[0].ReleaseCause != ReconcileReleaseOperator {
		t.Fatalf("history before = %+v err=%v", before, err)
	}

	_, released, err := j.ReleaseReconcileOnOrderTerminalProof(ctx, OrderTerminalProofRequest{StateID: row.ID, Symbol: "SYN001"},
		"terminal proof (synthetic)", func(OrderTerminalProof) bool { return true })
	if err == nil || released {
		t.Fatalf("release over an already released row: released=%v err=%v, want an error and no release", released, err)
	}
	after, err := j.ReconcileStateHistory(ctx, "acct-7")
	if err != nil || len(after) != 1 {
		t.Fatalf("history after = %+v err=%v", after, err)
	}
	if after[0].ReleaseCause != before[0].ReleaseCause || !after[0].ReleasedAt.Equal(before[0].ReleasedAt) ||
		after[0].Evidence != before[0].Evidence {
		t.Fatalf("the released row changed: before %+v after %+v", before[0], after[0])
	}
}
