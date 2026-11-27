package journal

// a209 — 주문 화면의 청산 근거 연결(ReadOnly.BrokerOrderExitLinks)은 무장 행만 셈.
//
// 엔진의 청산 발의 하나는 설계상 exit_events 행 둘을 가짐: 무장 행(증거 20 열)과 종결 훅 행(PROPOSAL_FILLED ·
// PROPOSAL_REFUSED · PROPOSAL_CANCELLED, 증거 NULL, 같은 proposed_intent_id). 기존 픽스처는 무장 행만 심어서
// 훅 행까지 세던 결함을 못 잡았음 — 그래서 여기 시험은 실제 수명주기(무장 → 주문 → 체결 · 해제)로 두 행을 만듦.

import (
	"context"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// a209ArmAndPlace 는 손절 하나를 무장하고 그 intent 로 매도 주문을 접수 확정까지 기록함(체결 · 해제 전 상태).
func a209ArmAndPlace(t *testing.T, j *Journal, intentID string) (positionID string, snapshot exitpolicy.ExitLineSnapshot, sell order) {
	t.Helper()
	_, seed := openedPosition(t, j, "10")
	snapshot, _, err := a207Arm(t, j, seed.PositionID, "obs-a209", "67900", intentID)
	if err != nil {
		t.Fatal(err)
	}
	sell = place(t, j, order{intentID: intentID, attemptID: "attempt-" + intentID, orderID: "broker-" + intentID,
		side: "SELL", quantity: "10"})
	return seed.PositionID, snapshot, sell
}

// a209ActionsOf 는 intent 에 묶인 exit_events 행의 action 을 id 순으로 돌려줌 — 픽스처가 실제 두 행 모양인지 단언하는 데 씀.
func a209ActionsOf(t *testing.T, j *Journal, intentID string) []string {
	t.Helper()
	rows, err := j.db.Query(`SELECT action FROM exit_events WHERE proposed_intent_id=? ORDER BY id`, intentID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			t.Fatal(err)
		}
		out = append(out, action)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func a209LinkOf(t *testing.T, j *Journal, sell order) BrokerOrderExitLink {
	t.Helper()
	links, err := openTestReadOnly(t, j.Path()).BrokerOrderExitLinks(context.Background(), []BrokerOrderScope{{
		BrokerOrderID: sell.orderID, AccountRef: sell.account, Market: sell.market, TradingDay: sell.tradingDay}})
	if err != nil {
		t.Fatalf("BrokerOrderExitLinks: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("links = %+v, want one answer", links)
	}
	return links[0]
}

// 무장 행 + PROPOSAL_FILLED 의 체결 주문은 모호가 아니고 무장 행의 근거(decision)를 보여 줌.
func TestA209AFilledExitOrderLinksItsArmingRowDespiteTheFillHook(t *testing.T) {
	j := exitFixture(t)
	_, snapshot, sell := a209ArmAndPlace(t, j, "exit-a209-filled")
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "10", "67900")); err != nil {
		t.Fatalf("RecordFill(sell): %v", err)
	}
	if got := a209ActionsOf(t, j, "exit-a209-filled"); len(got) != 2 ||
		got[0] != string(exitpolicy.ActionBaselineBreach) || got[1] != ExitEventProposalFilled {
		t.Fatalf("fixture rows = %v, want the real lifecycle shape [arm, %s]", got, ExitEventProposalFilled)
	}

	link := a209LinkOf(t, j, sell)
	if link.Ambiguous || link.UnknownReason != "" {
		t.Fatalf("filled exit link = %+v, want unambiguous with no unknown reason", link)
	}
	if link.Event.Action != string(exitpolicy.ActionBaselineBreach) || link.IntentID != "exit-a209-filled" ||
		link.AttemptID != "attempt-exit-a209-filled" || link.Event.Evaluation.Effective.Snapshot == nil ||
		link.Event.Evaluation.Effective.Snapshot.Line.DecisionID != snapshot.DecisionID {
		t.Fatalf("filled exit link = %+v, want the arming row's decision %s", link, snapshot.DecisionID)
	}
}

// 무장 행 + PROPOSAL_REFUSED / PROPOSAL_CANCELLED 의 주문도 모호가 아님.
func TestA209AReleasedExitOrderLinksItsArmingRowDespiteTheReleaseHook(t *testing.T) {
	for _, tc := range []struct {
		resolution ProposalResolution
		hook       string
	}{
		{ProposalRefused, ExitEventProposalRefused},
		{ProposalCancelled, ExitEventProposalCancelled},
	} {
		t.Run(tc.hook, func(t *testing.T) {
			j := exitFixture(t)
			intentID := "exit-a209-" + tc.hook
			positionID, snapshot, sell := a209ArmAndPlace(t, j, intentID)
			if err := j.ResolveExitProposal(context.Background(), positionID, intentID, tc.resolution); err != nil {
				t.Fatalf("ResolveExitProposal: %v", err)
			}
			if got := a209ActionsOf(t, j, intentID); len(got) != 2 || got[1] != tc.hook {
				t.Fatalf("fixture rows = %v, want [arm, %s]", got, tc.hook)
			}

			link := a209LinkOf(t, j, sell)
			if link.Ambiguous || link.Event.Action != string(exitpolicy.ActionBaselineBreach) ||
				link.Event.Evaluation.Effective.Snapshot == nil ||
				link.Event.Evaluation.Effective.Snapshot.Line.DecisionID != snapshot.DecisionID {
				t.Fatalf("%s exit link = %+v, want the arming row's decision %s", tc.hook, link, snapshot.DecisionID)
			}
		})
	}
}

// 무장 행이 둘이면 훅 행이 있어도 지금처럼 모호임(fail-closed 유지).
func TestA209TwoArmingRowsStayAmbiguousWithAHookRow(t *testing.T) {
	j := exitFixture(t)
	positionID, _, sell := a209ArmAndPlace(t, j, "exit-a209-two")
	if _, err := j.RecordFill(context.Background(), terminalFill(sell, "10", "67900")); err != nil {
		t.Fatalf("RecordFill(sell): %v", err)
	}
	// 같은 intent 를 가리키는 둘째 무장 행 — 원장이 만들지 않는 모양이라 직접 심음.
	if _, err := j.db.Exec(`INSERT INTO exit_events(position_id, action, proposed_intent_id, created_at)
		VALUES (?, ?, 'exit-a209-two', '2026-03-30T01:00:00Z')`, positionID, string(exitpolicy.ActionBaselineBreach)); err != nil {
		t.Fatal(err)
	}
	if got := a209ActionsOf(t, j, "exit-a209-two"); len(got) != 3 {
		t.Fatalf("fixture rows = %v, want two arms and the fill hook", got)
	}

	link := a209LinkOf(t, j, sell)
	if !link.Ambiguous || link.UnknownReason != "ambiguous_exit_evidence" || link.Event.ID != 0 {
		t.Fatalf("two-arm link = %+v, want the fail-closed ambiguous marker", link)
	}
}
