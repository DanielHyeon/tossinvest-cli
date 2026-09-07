package journal

// a065 task 6.4 의 독립 적대 리뷰가 찾은 결함들을 못 박는 시험이다.
//
// 모든 발견이 같은 뿌리였다: 같은 규칙이 두 자리에 서로 다른 방법으로 구현돼 있고,
// 시험은 자기가 보고 쓴 쪽을 통과시킨다. 그래서 여기서는 행동만 보지 않고
// **어디서 판정하는가** 도 AST 로 못 박는다.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positioncampaign"
)

// createCampaignFixtureAtGeneration 은 이미 position 행이 있는 scope 에 campaign 을
// 만든다. prospective generation CAS 가 실제 generation/version 과 맞아야 하므로
// 기본 fixture(0/0)를 쓸 수 없다.
func createCampaignFixtureAtGeneration(t *testing.T, j *Journal, generation, version int64) PositionCampaignRecord {
	t.Helper()
	insertCampaignDecision(t, j, "decision", "acct")
	insertCampaignStrategyLineage(t, j, "decision", "acct", "kr", "005930", "kr", "v1", "digest")
	got, err := j.CreatePositionCampaign(context.Background(), CreatePositionCampaignRequest{
		ID: "campaign", AccountRef: "acct", Market: "kr", Symbol: "005930",
		LaneID: "kr", LaneVersion: "v1", DecisionID: "decision", EvidenceDigest: "digest",
		ExpectedPositionGeneration: generation, ExpectedPositionVersion: version,
		ProspectiveToken: "token", CommandKey: "create",
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// ---------------------------------------------------------------------------
// I-1: 교체가 있는 campaign 은 offline 재구성된다
// ---------------------------------------------------------------------------

// review.md 가 "Accepted after round 2" 로 적은 대표 시나리오다. 원장은 만들었지만
// 재구성을 부르는 시험이 없어서, replay 가 leg 수준 attempt 불변을 지어낸 채로
// 통과했다 — 교체 주문은 같은 leg 위에 **새** attempt 를 만든다(design D5).
func TestCampaignWithReplacementReconstructsFromEvidence(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-old", "attempt-old", "old", "", "10")
	linked, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link-old", OrderID: "old", RequestedCap: "10",
		IntentID: "intent-old", AttemptID: "attempt-old",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('campaign-position','acct','kr','005930',1,NULL,'OPEN','4','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "old", AccountRef: "acct", Market: "kr",
		Symbol: "005930", Delta: "4", CumulativeQuantity: "4", CommittedAt: "2026-03-30T00:31:00Z"})
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent-old", "attempt-new", "new", "old", "6")
	if _, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: linked.CampaignVersion + 1,
		CommandKey: "link-new", OrderID: "new", PredecessorOrderID: "old", RequestedCap: "6",
		IntentID: "intent-old", AttemptID: "attempt-new",
	}); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "new", AccountRef: "acct", Market: "kr",
		Symbol: "005930", Delta: "2", CumulativeQuantity: "2", CommittedAt: "2026-03-30T00:32:00Z"})

	result, err := j.ReconstructPositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Fatalf("건강한 원장이 drift 로 신고됐다: reason=%s lastValid=%d", result.Reason, result.LastValidSequence)
	}
}

// 체결과 잔량 취소가 **관측 하나**로 올 때 원장과 재구성이 같은 답을 내는지 본다.
// 예전 판본은 원장이 CANCELLED 를 쓰고 표는 그것을 유도하지 못했다.
func TestResidualCancelInOneObservationAgreesWithReconstruction(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "10")
	if _, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "order", RequestedCap: "10",
		IntentID: "intent", AttemptID: "attempt",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('campaign-position','acct','kr','005930',1,NULL,'OPEN','3','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	// 3주 체결 + 잔량 취소가 한 번에 온다. leg 는 SUBMITTED 에서 곧장 닫힌다.
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr",
		Symbol: "005930", Delta: "3", CumulativeQuantity: "3", Terminal: true,
		CommittedAt: "2026-03-30T00:31:00Z"})

	legGot, err := j.CampaignLeg(ctx, campaign.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if legGot.State != positioncampaign.LegCancelled {
		t.Fatalf("leg=%+v, want CANCELLED (잔량 취소)", legGot)
	}
	result, err := j.ReconstructPositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Fatalf("원장과 재구성이 갈렸다: reason=%s", result.Reason)
	}
}

// ---------------------------------------------------------------------------
// I-4: 평범한 체결 하나가 손절 fail-closed 래치를 지우지 않는다
// ---------------------------------------------------------------------------

// spec: "stop evidence 가 missing 또는 invalid … 새 exposure-raising leg 는 fail
// closed 된다". 그 래치는 campaign 상태에 encode 되지 않으므로 상태에서 되계산하면
// 지워진다. 지워지면 그 다음 leg 가 손절 근거 없이 승인된다.
func TestCampaignFillDoesNotClearTheStopEntryLatch(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "10")
	linked, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "order", RequestedCap: "10",
		IntentID: "intent", AttemptID: "attempt",
	})
	if err != nil {
		t.Fatal(err)
	}
	// stop 증거가 없다 → 상태는 그대로 두고 진입만 fail closed 로 건다.
	latched, err := j.UpdateCampaignStop(ctx, UpdateCampaignStopRequest{
		CampaignID: campaign.ID, ExpectedVersion: linked.CampaignVersion, CommandKey: "stop",
		Candidate: positioncampaign.StopCandidate{Valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !latched.EntryBlocked {
		t.Fatalf("전제 실패: stop 증거 누락이 래치를 걸지 않았다: %+v", latched)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('campaign-position','acct','kr','005930',1,NULL,'OPEN','4','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr",
		Symbol: "005930", Delta: "4", CumulativeQuantity: "4", CommittedAt: "2026-03-30T00:31:00Z"})

	after, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !after.EntryBlocked {
		t.Fatalf("평범한 체결 하나가 손절 래치를 지웠다: %+v", after)
	}
	// 래치가 실제로 진입을 막는지까지 본다 — 열 값만 보면 소비자가 없을 수 있다.
	if _, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: after.Version, CommandKey: "plan-2",
		Sequence: 2, PlanID: "leg-2", RequestedQuantity: "5",
	}); err == nil {
		t.Fatal("래치가 걸린 campaign 이 다음 진입 leg 를 승인했다")
	}
	// 재구성도 같은 래치를 인증해야 한다. 한쪽만 고치면 다른 쪽이 옛 동작을 인증한다.
	result, err := j.ReconstructPositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Valid {
		t.Fatalf("재구성이 래치된 원장을 거부했다: reason=%s", result.Reason)
	}
	if !result.State.EntryBlocked {
		t.Fatal("재구성이 진입 차단을 잃었다")
	}
}

// ---------------------------------------------------------------------------
// I-5: 묶인 position generation 이 닫히면 새 노출을 올리지 않는다
// ---------------------------------------------------------------------------

func TestClosedBoundGenerationBlocksTheNextEntryLeg(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	leg, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	})
	if err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "10")
	if _, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: leg.CampaignVersion,
		CommandKey: "link", OrderID: "order", RequestedCap: "10",
		IntentID: "intent", AttemptID: "attempt",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('campaign-position','acct','kr','005930',1,NULL,'OPEN','4','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	applyCampaignFillForTest(t, j, AppliedFill{OrderID: "order", AccountRef: "acct", Market: "kr",
		Symbol: "005930", Delta: "4", CumulativeQuantity: "4", CommittedAt: "2026-03-30T00:31:00Z"})
	bound, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if bound.ActualPositionGeneration == 0 {
		t.Fatalf("전제 실패: campaign 이 generation 에 묶이지 않았다: %+v", bound)
	}
	// 청산 완료. instance_seq 는 그대로이므로 generation 검사로는 걸리지 않는다.
	if _, err := j.db.Exec(`UPDATE positions SET state='CLOSED',quantity='0',closed_at='2026-03-30T00:40:00Z'
		WHERE id='campaign-position'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: bound.Version, CommandKey: "plan-2",
		Sequence: 2, PlanID: "leg-2", RequestedQuantity: "5",
	}); err == nil {
		t.Fatal("죽은 generation 위에서 새 진입 leg 가 승인됐다")
	}
}

// 이 fail-closed 가 무엇을 거부하지 **않는지** 도 못 박는다. "최신 행이 CLOSED"
// 로 막으면 spec 의 "청산 후 재진입" 이 영구히 불가능해진다.
func TestReEntryAfterCloseIsAdmittedOnANewCampaign(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at,closed_at)
		VALUES ('old-position','acct','kr','005930',1,NULL,'CLOSED','0','70000',
		        '2026-03-29T00:31:00Z','2026-03-29T05:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	campaign := createCampaignFixtureAtGeneration(t, j, 1, 1)
	if _, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	}); err != nil {
		t.Fatalf("청산 후 새 campaign 의 재진입이 막혔다: %v", err)
	}
}

// ---------------------------------------------------------------------------
// I-8: 미해결 risk-reducing 판정에는 시간 경계가 있다
// ---------------------------------------------------------------------------

// 경계가 없으면 해소 경로가 없는 옛 intent 하나가 그 종목의 모든 진입을 영원히
// 막는다. 그것은 안전이 아니라 기능 장애다.
func TestStaleRiskReducingIntentFromAClosedGenerationDoesNotBlockForever(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	if _, err := j.db.Exec(`INSERT INTO intents
		(id,created_at,market,trading_day,account_ref,symbol,side,order_type,time_in_force,
		 quantity,price,currency,source,fingerprint,notes)
		VALUES ('stale-sell','2020-01-02T00:00:00Z','kr','2020-01-02','acct','005930',
		        'SELL','LIMIT','','1','70000','KRW','ancient','fp-stale','')`); err != nil {
		t.Fatal(err)
	}
	// 그 뒤에 새 generation 이 열렸다.
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('live-position','acct','kr','005930',9,NULL,'OPEN','4','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	}); err != nil {
		t.Fatalf("2020 년의 죽은 SELL intent 가 2026 년 진입을 막았다: %v", err)
	}
}

// 같은 경계가 **현재** generation 의 미해결 SELL 은 여전히 막는지 본다.
// 양성 대조군이 없으면 위 시험은 판정을 통째로 지워도 통과한다.
func TestUnresolvedRiskReducingIntentInTheLiveGenerationStillBlocks(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('live-position','acct','kr','005930',9,NULL,'OPEN','4','70000','2026-03-30T00:31:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO intents
		(id,created_at,market,trading_day,account_ref,symbol,side,order_type,time_in_force,
		 quantity,price,currency,source,fingerprint,notes)
		VALUES ('live-sell','2026-03-30T00:32:00Z','kr','2026-03-30','acct','005930',
		        'SELL','LIMIT','','1','70000','KRW','exit','fp-live','')`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.PlanCampaignLeg(ctx, PlanCampaignLegRequest{
		CampaignID: campaign.ID, ExpectedVersion: campaign.Version, CommandKey: "plan",
		Sequence: 1, PlanID: "leg-1", RequestedQuantity: "10",
	}); err == nil {
		t.Fatal("현재 generation 의 미해결 risk-reducing intent 가 진입을 막지 못했다")
	}
}

// ---------------------------------------------------------------------------
// I-7: EXIT FIRST 거절이 증거를 남긴다
// ---------------------------------------------------------------------------

// LinkCampaignOrder 는 attempt 가 이미 CONFIRMED 일 것을 요구한다. 그래서 이 시점의
// 거절은 주문을 되돌리지 못한다 — 아무것도 안 남기면 그 주문의 실제 체결이 원장에서
// 영구히 사라진다.
func TestExposureRefusalAtLinkLeavesRefusalEvidence(t *testing.T) {
	j := openTestJournal(t)
	ctx := context.Background()
	campaign := createCampaignFixture(t, j)
	if _, err := j.db.Exec(`INSERT INTO positions
		(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('closing-position','acct','kr','005930',1,NULL,'CLOSING','4','70000','2026-03-30T00:20:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO campaign_legs
		(campaign_id,sequence,plan_id,requested_quantity,filled_quantity,residual_quantity,state,version,created_at,updated_at)
		VALUES (?,1,'leg-1','10','0','10','PLANNED',1,'2026-03-30T00:21:00Z','2026-03-30T00:21:00Z')`,
		campaign.ID); err != nil {
		t.Fatal(err)
	}
	insertCampaignExecutionLineage(t, j, campaign.DecisionID, "intent", "attempt", "order", "", "10")
	_, err := j.LinkCampaignOrder(ctx, LinkCampaignOrderRequest{
		CampaignID: campaign.ID, LegSequence: 1, ExpectedVersion: campaign.Version,
		CommandKey: "link", OrderID: "order", RequestedCap: "10",
		IntentID: "intent", AttemptID: "attempt",
	})
	if err == nil {
		t.Fatal("CLOSING position 위에서 link 가 승인됐다")
	}
	var refusals int
	if err := j.db.QueryRow(`SELECT count(*) FROM campaign_events
		WHERE campaign_id=? AND event_kind='ORDER_LINK_REFUSED'`, campaign.ID).Scan(&refusals); err != nil {
		t.Fatal(err)
	}
	if refusals != 1 {
		t.Fatalf("EXIT FIRST 거절이 증거를 남기지 않았다: ORDER_LINK_REFUSED %d 건", refusals)
	}
	after, err := j.PositionCampaign(ctx, campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.State != positioncampaign.CampaignReconcile || !after.EntryBlocked {
		t.Fatalf("거절이 campaign 을 격리하지 않았다: %+v", after)
	}
}

// ---------------------------------------------------------------------------
// 구조: 판정이 하나인지 AST 로 못 박는다
// ---------------------------------------------------------------------------

func campaignSourceFile(t *testing.T, path string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func campaignFuncDecl(t *testing.T, file *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == name && function.Body != nil {
			return function
		}
	}
	t.Fatalf("%s 선언을 찾지 못했다", name)
	return nil
}

// 행동 시험만으로는 부족하다. 적용 쪽이 함수 안에서 직접 계산하더라도 우연히 같은
// 답을 내면 행동 시험은 통과하고, 다음 편집에서 다시 조용히 갈린다 — 실제로 그렇게
// 갈려 있었다. 그래서 "legState 에 값을 넣는 자리" 를 열거한다.
func TestCampaignFillHasExactlyOneLegStateJudgement(t *testing.T) {
	function := campaignFuncDecl(t, campaignSourceFile(t, "position_campaign.go"), "ApplyPositionCampaignFill")
	shared := 0
	var offending []string
	ast.Inspect(function.Body, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for index, target := range assign.Lhs {
			name, isIdent := target.(*ast.Ident)
			if !isIdent || name.Name != "legState" || index >= len(assign.Rhs) {
				continue
			}
			call, isCall := assign.Rhs[index].(*ast.CallExpr)
			if !isCall {
				offending = append(offending, "legState 에 호출이 아닌 값을 넣는다")
				continue
			}
			selector, isSelector := call.Fun.(*ast.SelectorExpr)
			if !isSelector {
				offending = append(offending, "legState 에 알 수 없는 호출 결과를 넣는다")
				continue
			}
			switch selector.Sel.Name {
			case "LegStateAfterFill":
				shared++
			case "LegState":
				// 표가 모르는 사실일 때 이전 상태를 그대로 두는 변환.
			default:
				offending = append(offending, "legState 를 "+selector.Sel.Name+" 로 정한다")
			}
		}
		return true
	})
	if shared != 1 {
		t.Fatalf("공용 판정 LegStateAfterFill 호출이 %d 번이다, 정확히 1 번이어야 한다", shared)
	}
	if len(offending) != 0 {
		t.Fatalf("두 번째 leg 상태 판정이 있다: %v", offending)
	}
}

// 재구성 쪽에도 두 번째 판정이 남아 있지 않은지 본다.
func TestReplayDerivesLegStateThroughTheSharedJudgement(t *testing.T) {
	file := campaignSourceFile(t, "../positioncampaign/replay.go")
	calls := 0
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name, isIdent := call.Fun.(*ast.Ident)
		if !isIdent {
			return true
		}
		switch name.Name {
		case "LegStateAfterFill":
			calls++
		case "replayLegFillTransition":
			t.Error("옛 판정 replayLegFillTransition 이 아직 불린다")
		}
		return true
	})
	if calls != 1 {
		t.Fatalf("replay 가 공용 판정을 %d 번 부른다, 정확히 1 번이어야 한다", calls)
	}
}

// entry_blocked 를 쓰는 **모든** 자리가 같은 래치 함수를 거치는지 본다.
// 한 자리라도 상태에서 되계산하면 손절 래치가 거기서 지워진다.
func TestEveryEntryBlockedWriterGoesThroughTheLatch(t *testing.T) {
	sites := map[string]string{
		"position_campaign.go":         "ApplyPositionCampaignFill",
		"strategy_dispatch_runtime.go": "linkConfirmedStrategyCampaignTx",
	}
	for path, name := range sites {
		function := campaignFuncDecl(t, campaignSourceFile(t, path), name)
		latched := 0
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if selector, isSelector := call.Fun.(*ast.SelectorExpr); isSelector &&
				selector.Sel.Name == "LatchEntryBlocked" {
				latched++
			}
			return true
		})
		if latched != 1 {
			t.Errorf("%s: LatchEntryBlocked 호출이 %d 번이다, 정확히 1 번이어야 한다", name, latched)
		}
	}
	// LinkCampaignOrder 는 별도 자리다 — 위 두 곳과 같은 규칙을 쓴다.
	function := campaignFuncDecl(t, campaignSourceFile(t, "position_campaign.go"), "LinkCampaignOrder")
	latched := 0
	ast.Inspect(function.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if selector, isSelector := call.Fun.(*ast.SelectorExpr); isSelector &&
			selector.Sel.Name == "LatchEntryBlocked" {
			latched++
		}
		return true
	})
	if latched != 1 {
		t.Errorf("LinkCampaignOrder: LatchEntryBlocked 호출이 %d 번이다, 정확히 1 번이어야 한다", latched)
	}
}
