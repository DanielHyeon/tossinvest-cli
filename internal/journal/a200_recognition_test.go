package journal

// a200_recognition_test.go 는 a200 의 새 계약(design D2 · §2 · §3)을 고정함 — 수렴 조정이 닫은 인스턴스에 늦게 닿은 엔진 청산 체결의
// 인정과, 인정하면 안 되는 체결이 오늘과 같은 귀속 실패로 남는 것. 하네스(a200Journal · a200ArmStop · a200Converge · readA200Ledger)는
// 1단계 재현 시험 파일(a200_convergence_close_exit_fill_repro_test.go)의 것을 그대로 씀. 픽스처는 합성 값만 씀.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// a200Marker 는 인정 표지 action 의 계약 이름(design D3). 생산 상수와 같은지는 구조 시험이 봄.
const a200Marker = "ADJUSTMENT_CLOSE_FILL_RECOGNISED"

// a200AssertRecognised 는 인정된 체결의 원장 모양: 차단 0 · 투영 CLOSED/0 · 표지 n · COMPLETED 0 · 성과 0.
func a200AssertRecognised(t *testing.T, got a200Ledger, markers int) {
	t.Helper()
	if len(got.Reconcile) != 0 {
		t.Fatalf("reconcile_states = %q, 인정된 체결은 차단을 만들지 않아야 함", got.Reconcile)
	}
	if got.State != PositionClosed || got.Quantity != "0" {
		t.Fatalf("투영 = %s/%s, 인정은 수량을 다시 빼지 않음(CLOSED/0 그대로)", got.State, got.Quantity)
	}
	if n := countOf(got.ExitActions, a200Marker); n != markers {
		t.Fatalf("표지 %d 개(exit_events=%q), 인정된 체결마다 하나여야 함(%d)", n, got.ExitActions, markers)
	}
	if countOf(got.ExitActions, ExitEventAdjustmentClosed) != 1 || countOf(got.ExitActions, ExitEventCompleted) != 0 {
		t.Fatalf("exit_events = %q, 인정돼도 ADJUSTMENT_CLOSED 1 · COMPLETED 0 그대로(성과 비목표)", got.ExitActions)
	}
	if got.Outcomes != 0 {
		t.Fatalf("trade_outcomes = %d, 인정은 성과를 만들지 않음", got.Outcomes)
	}
}

// a200AssertRefused 는 인정하면 안 되는 체결의 원장 모양: 오늘과 같은 ATTRIBUTION_FAILED(SELL_ON_CLOSED) 활성 1 · 표지 0.
func a200AssertRefused(t *testing.T, got a200Ledger) {
	t.Helper()
	if len(got.Reconcile) != 1 || !strings.HasPrefix(got.Reconcile[0], ReconcileCauseAttributionFailed+"|position projection refused transition X5") ||
		!strings.Contains(got.Reconcile[0], "(SELL_ON_CLOSED)") || !strings.HasSuffix(got.Reconcile[0], "|active") {
		t.Fatalf("reconcile_states = %q, 오늘과 같은 ATTRIBUTION_FAILED(SELL_ON_CLOSED) 활성 1", got.Reconcile)
	}
	if n := countOf(got.ExitActions, a200Marker); n != 0 {
		t.Fatalf("표지 %d 개, 인정 안 된 체결은 표지를 남기지 않음", n)
	}
}

// a200ManagedStop 은 엔진 진입 관리 포지션 31 + 손절 발의(주문 수량 qty)를 만들고, 시계를 1 초 넘김.
func a200ManagedStop(t *testing.T, qty string) (*Journal, *clock.Fake, Position, order) {
	t.Helper()
	j, clk := a200Journal(t)
	entry, _ := openedPosition(t, j, "31")
	p := currentPosition(t, j, entry)
	sell := a200ArmStop(t, j, p.ID, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: qty})
	clk.Advance(time.Second)
	return j, clk, p, sell
}

// --- 인정되어야 하는 정상 입력(design §2) ------------------------------------------------

// TestA200AStopFillAfterAConvergenceCloseIsRecognised 는 본 사고 모양(design §3-1): 같은 초, 전량.
func TestA200AStopFillAfterAConvergenceCloseIsRecognised(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "31")
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(stop): %v", err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "본 사고 모양")
	a200AssertRecognised(t, got, 1)
	want := []string{ExitEventOpened, "BASELINE_BREACH_EXIT", ExitEventAdjustmentClosed, a200Marker, ExitEventProposalFilled}
	if strings.Join(got.ExitActions, ",") != strings.Join(want, ",") {
		t.Fatalf("exit_events = %q, want %q(표지는 Project 훅, PROPOSAL_FILLED 는 Exit 훅)", got.ExitActions, want)
	}
	// 표지 행의 모양: 발의 intent · 결정 · 평가 증거 열이 전부 NULL(design D3 — 소비자가 무시하는 모양).
	var intentNull, decisionNull, projectedNull, recomputedNull bool
	if err := j.db.QueryRowContext(context.Background(), `SELECT proposed_intent_id IS NULL, decision_id IS NULL,
		projected_quantity IS NULL, recomputed_snapshot_json IS NULL FROM exit_events WHERE action = ?`, a200Marker).
		Scan(&intentNull, &decisionNull, &projectedNull, &recomputedNull); err != nil {
		t.Fatalf("표지 행 읽기: %v", err)
	}
	if !intentNull || !decisionNull || !projectedNull || !recomputedNull {
		t.Fatalf("표지 행 열 NULL = intent %v decision %v projected %v recomputed %v, 전부 NULL 이어야 함",
			intentNull, decisionNull, projectedNull, recomputedNull)
	}
}

// TestA200AnAdoptedStopFillIsRecognisedWithoutAnOutcome 은 편입 포지션의 같은 모양 — 인정되고 성과는 없음(비목표).
func TestA200AnAdoptedStopFillIsRecognisedWithoutAnOutcome(t *testing.T) {
	j, clk := a200Journal(t)
	ctx := context.Background()
	adoptable(t, j, "p-ext")
	if _, err := j.AdoptPosition(ctx, sampleAdoption("p-ext")); err != nil {
		t.Fatal(err)
	}
	if _, err := j.OpenAdoptedExitState(ctx, "p-ext"); err != nil {
		t.Fatal(err)
	}
	clk.Advance(time.Second)
	sell := a200ArmStop(t, j, "p-ext", order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop",
		side: "SELL", quantity: "10"})
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "10", "0", clk.Now())
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "67900")); err != nil {
		t.Fatalf("RecordFill(stop): %v", err)
	}
	got := readA200Ledger(t, j, "p-ext")
	got.log(t, "편입 인정")
	a200AssertRecognised(t, got, 1)
	if _, reason := probeFreeze(t, j, "p-ext"); reason != OutcomeRefusalPostAdoptionDisposal {
		t.Fatalf("편입 성과 탐침 사유 = %q, 성과 산식은 무변(post_adoption_disposal)", reason)
	}
}

// TestA200PartialStopFillsAfterAConvergenceCloseAreEachRecognised 는 조정 뒤 부분 체결 여러 번(design D4): 체결마다 인정 · 표지.
func TestA200PartialStopFillsAfterAConvergenceCloseAreEachRecognised(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "31")
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	if _, err := j.RecordFill(context.Background(), fillOf(sell, "20", "67900")); err != nil {
		t.Fatalf("RecordFill(working): %v", err)
	}
	first := readA200Ledger(t, j, p.ID)
	if first.PendingID != "i-stop" {
		t.Fatalf("Working 체결 뒤 pending=%q, 종결 전까지 살아 있어야 함", first.PendingID)
	}
	clk.Advance(time.Second)
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatalf("RecordFill(done): %v", err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "부분 체결 둘")
	a200AssertRecognised(t, got, 2)
}

// TestA200APartialTakeProfitBelowTheDecreaseIsRecognised 는 주문 수량 < 감소분(design §2 정상 5).
func TestA200APartialTakeProfitBelowTheDecreaseIsRecognised(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "10")
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "10", "72000")); err != nil {
		t.Fatalf("RecordFill: %v", err)
	}
	a200AssertRecognised(t, readA200Ledger(t, j, p.ID), 1)
}

// TestA200ASmallOrderStraddlingTheConvergenceIsRecognised 는 주문 수량 ≤ 감소분인 부분 청산 주문이 조정을 사이에 두고 체결되는
// 모양(구현 적대 리뷰 impl-r1 P1-1 의 탐침 TestRevA200ProbeStraddleSmallOrder 를 계약으로): 보유 31 · 손절 주문 10 · 조정 전 체결 3
// (31→28) → 조정 28→0 → 나머지 7. 주문 수량 10 ≤ 감소분 28 이므로 조정 뒤 체결은 인정됨 — 남은 7 이 체결돼야 계좌가 0 이 되므로 수량상 옳음.
func TestA200ASmallOrderStraddlingTheConvergenceIsRecognised(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "10")
	ctx := context.Background()
	if _, err := j.RecordFill(ctx, fillOf(sell, "3", "67900")); err != nil {
		t.Fatalf("RecordFill(조정 전 부분): %v", err)
	}
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "28", "0", clk.Now())
	if _, err := j.RecordFill(ctx, terminalFill(sell, "10", "67900")); err != nil {
		t.Fatalf("RecordFill(조정 뒤 나머지): %v", err)
	}
	got := readA200Ledger(t, j, p.ID)
	got.log(t, "작은 주문이 조정을 사이에 둠")
	a200AssertRecognised(t, got, 1)
}

// --- 계속 거절해야 하는 입력(design §2) — 편집 전에도 통과, 편집 뒤에도 통과해야 함 -------------

// TestA200AFillFromAnotherIntentStaysAnAttributionFailure: 체결 주문이 그 인스턴스의 청산 발의가 아님((다)).
func TestA200AFillFromAnotherIntentStaysAnAttributionFailure(t *testing.T) {
	j, clk, p, _ := a200ManagedStop(t, "31")
	other := place(t, j, order{intentID: "i-other", attemptID: "a-other", orderID: "o-other", side: "SELL", quantity: "31"})
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	if _, err := j.RecordFill(context.Background(), terminalFill(other, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	a200AssertRefused(t, readA200Ledger(t, j, p.ID))
}

// TestA200AFillAfterThePendingWasResolvedStaysAnAttributionFailure: 발의가 이미 해소(종결 관측 · 해제)된 뒤의 체결((다), codex r2 P2).
func TestA200AFillAfterThePendingWasResolvedStaysAnAttributionFailure(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "31")
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	// 해소를 원장에서 직접 만듦 — 종결 관측의 Exit 훅(ApplyExitFill B11)이 하는 것과 같은 세 열 비우기.
	if _, err := j.db.ExecContext(context.Background(), `UPDATE exit_states
		SET pending_action = NULL, pending_level = NULL, pending_intent_id = NULL WHERE position_id = ?`, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	a200AssertRefused(t, readA200Ledger(t, j, p.ID))
}

// TestA200AnAmendedStopStaysAnAttributionFailure: 정정 뒤 체결(거절하는 정상 입력 (나)).
func TestA200AnAmendedStopStaysAnAttributionFailure(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "31")
	ctx := context.Background()
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	amend, err := j.Prepare(ctx, PrepareRequest{
		Intent: testIntentFor(sell), Kind: KindAmend, AttemptID: "a-stop-amend", TargetOrderID: "o-stop",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkDispatchStarted(ctx); err != nil {
		t.Fatal(err)
	}
	if err := amend.MarkAcked(ctx, "o-stop-b"); err != nil {
		t.Fatal(err)
	}
	if err := amend.ResolveConfirmedWithLineage(ctx, LineageEdge{
		ParentOrderID: "o-stop", ChildOrderID: "o-stop-b", Relation: RelationReplaces,
		ParentFilledQuantity: "0", RequestedQuantity: "31",
	}, "", "amended"); err != nil {
		t.Fatal(err)
	}
	clk.Advance(time.Second)
	child := order{orderID: "o-stop-b", symbol: sell.symbol, market: sell.market, account: sell.account,
		tradingDay: sell.tradingDay, quantity: "31", side: "SELL"}
	if _, err := j.RecordFill(ctx, terminalFill(child, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	a200AssertRefused(t, readA200Ledger(t, j, p.ID))
}

// TestA200ASecondPlaceAttemptStaysAnAttributionFailure: 한 intent 에 PLACE attempt 가 둘(재시도 · 재제출 — 전제 (iii)).
func TestA200ASecondPlaceAttemptStaysAnAttributionFailure(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "31")
	ctx := context.Background()
	if _, err := j.Prepare(ctx, PrepareRequest{Intent: testIntentFor(sell), Kind: KindPlace, AttemptID: "a-stop-2"}); err != nil {
		t.Fatalf("둘째 PLACE attempt: %v", err)
	}
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	if _, err := j.RecordFill(ctx, terminalFill(sell, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	a200AssertRefused(t, readA200Ledger(t, j, p.ID))
}

// TestA200AManualKindCloseStaysAnAttributionFailure: 닫은 조정의 kind 가 수렴이 쓰는 값이 아님(지어낸 MANUAL — 생산 쓰기 0).
func TestA200AManualKindCloseStaysAnAttributionFailure(t *testing.T) {
	j, clk, p, sell := a200ManagedStop(t, "31")
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, "005930")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, AdjustmentRequest{
		AccountRef: "acct-1", Market: "kr", Symbol: "005930", Kind: AdjustmentManual,
		ExpectedPrevQuantity: "31", ExpectedFillWatermark: watermark, NewQuantity: "0",
		BrokerAsOf: clk.Now().UTC().Format(time.RFC3339), Evidence: "an operator declared it (synthetic)",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.RecordFill(ctx, terminalFill(sell, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	a200AssertRefused(t, readA200Ledger(t, j, p.ID))
}

// TestA200AnUnmanagedInstanceStaysAnAttributionFailure: exit 상태가 없는 인스턴스(발의 없음 — (다)).
func TestA200AnUnmanagedInstanceStaysAnAttributionFailure(t *testing.T) {
	j, clk := a200Journal(t)
	ctx := context.Background()
	entry := place(t, j, order{intentID: "i-entry", attemptID: "a-entry", orderID: "o-entry", side: "BUY", quantity: "31"})
	if _, err := j.RecordFill(ctx, terminalFill(entry, "31", "70000")); err != nil {
		t.Fatal(err)
	}
	p := currentPosition(t, j, entry)
	sell := place(t, j, order{intentID: "i-stop", attemptID: "a-stop", orderID: "o-stop", side: "SELL", quantity: "31"})
	clk.Advance(time.Second)
	a200Converge(t, j, "005930", "31", "0", clk.Now())
	if _, err := j.RecordFill(ctx, terminalFill(sell, "31", "67900")); err != nil {
		t.Fatal(err)
	}
	// 무관리라 exit 상태가 없으므로 readA200Ledger 대신 차단 행만 직접 읽음.
	var cause string
	var active bool
	if err := j.db.QueryRowContext(ctx, `SELECT cause, released_at IS NULL FROM reconcile_states
		WHERE account_ref = 'acct-1' AND symbol = '005930'`).Scan(&cause, &active); err != nil {
		t.Fatalf("차단 행 읽기(%s): %v", p.ID, err)
	}
	if cause != ReconcileCauseAttributionFailed || !active {
		t.Fatalf("차단 = %s/%v, 무관리 인스턴스는 오늘처럼 ATTRIBUTION_FAILED 활성", cause, active)
	}
}

// --- 구조(design §3-5) -------------------------------------------------------------------

// a200RecognitionFile 은 판정 · 표지를 담는 새 파일(design §5).
const a200RecognitionFile = "a200_adjustment_close_fill_recognition.go"

// TestA200TheRecognitionIsCalledOnceAndReadsNoFillLedger: 판정 함수의 생산 호출은 ProjectPosition 한 자리, 새 파일은
// 보호 열 철자 · 체결 누적 표 · resolveFillOrigin 재호출이 없음, 표지 action 상수가 계약 이름과 같음.
func TestA200TheRecognitionIsCalledOnceAndReadsNoFillLedger(t *testing.T) {
	source, err := os.ReadFile(a200RecognitionFile)
	if err != nil {
		t.Fatalf("판정 파일 %s: %v", a200RecognitionFile, err)
	}
	text := string(source)
	for _, banned := range []string{"pending_intent_id", "fill_events", "fill_snapshots", "resolveFillOrigin("} {
		if strings.Contains(text, banned) {
			t.Errorf("%s 에 %q — 판정은 PendingState · origin 을 받고 체결 누적 표를 읽지 않아야 함", a200RecognitionFile, banned)
		}
	}
	if !strings.Contains(text, `"`+a200Marker+`"`) {
		t.Errorf("%s 에 표지 action 리터럴 %q 가 없음", a200RecognitionFile, a200Marker)
	}

	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	calls := map[string]int{}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "recogniseAdjustmentCloseFill" {
						calls[name+":"+fn.Name.Name]++
					}
				}
				return true
			})
		}
	}
	if len(calls) != 1 || calls["position_projection.go:ProjectPosition"] != 1 {
		t.Fatalf("recogniseAdjustmentCloseFill 생산 호출 = %v, ProjectPosition 한 자리여야 함", calls)
	}
}

// TestA200OtherRefusalsReadNothingBeforeTheRefusalCheck 는 구현 적대 리뷰 impl-r1 P2-1: SELL_ON_CLOSED 가 아닌 거절(OVERSELL 등)은
// 판정의 원장 읽기를 하지 않아야 함 — 그 읽기가 실패하면 무관한 거절의 체결 tx 까지 롤백됨. ApplyTx 에 읽기 계수기가 없어 구조로 잼:
// recogniseAdjustmentCloseFill 의 첫 문장이 recognisableRefusal 을 부르는 if 이고 그 갈래가 바로 반환하며, readRecognitionFacts 호출은
// 그 뒤에만 있음. (가)의 정의는 recognisableRefusal 한 곳(recognitionVerdict 도 그것을 부름).
func TestA200OtherRefusalsReadNothingBeforeTheRefusalCheck(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, a200RecognitionFile, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := func(n ast.Node, name string) []token.Pos {
		var at []token.Pos
		ast.Inspect(n, func(x ast.Node) bool {
			if call, ok := x.(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == name {
					at = append(at, call.Pos())
				}
			}
			return true
		})
		return at
	}
	var recognise, verdict *ast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			switch fn.Name.Name {
			case "recogniseAdjustmentCloseFill":
				recognise = fn
			case "recognitionVerdict":
				verdict = fn
			}
		}
	}
	if recognise == nil || verdict == nil {
		t.Fatal("recogniseAdjustmentCloseFill · recognitionVerdict 가 없음")
	}
	first, ok := recognise.Body.List[0].(*ast.IfStmt)
	if !ok || len(calls(first.Cond, "recognisableRefusal")) != 1 {
		t.Fatal("recogniseAdjustmentCloseFill 의 첫 문장이 recognisableRefusal 을 부르는 if 가 아님 — 다른 거절에서 원장을 읽게 됨")
	}
	if _, ok := first.Body.List[len(first.Body.List)-1].(*ast.ReturnStmt); !ok {
		t.Fatal("거절 유형 검사 갈래가 반환하지 않음")
	}
	reads := calls(recognise.Body, "readRecognitionFacts")
	if len(reads) != 1 || reads[0] < first.End() {
		t.Fatalf("readRecognitionFacts 호출 %v — 거절 유형 검사(끝 %v) 뒤에 한 번이어야 함", reads, first.End())
	}
	if len(calls(verdict.Body, "recognisableRefusal")) != 1 {
		t.Fatal("recognitionVerdict 가 (가)를 recognisableRefusal 로 판정하지 않음 — 정의가 둘이 됨")
	}
	if !recognisableRefusal(position.RefusalSellOnClosed) || recognisableRefusal(position.RefusalOversell) ||
		recognisableRefusal(position.RefusalEntryWhileClosing) || recognisableRefusal(position.RefusalUnattributedSell) {
		t.Fatal("recognisableRefusal 은 SELL_ON_CLOSED 만 참이어야 함")
	}
}
