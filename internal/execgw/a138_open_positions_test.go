package execgw_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a138: 진입 발급 문 셋이 모두 정책의 최대 보유 포지션 수를 예약 tx 에 실어 보내고, 넘으면 LIMIT_REACHED 로 아무것도 남기지 않음.
// 매도(IssueReduction)는 상한을 넘은 뒤에도 그대로 발급됨(§0.3).

func maxOpenPositions(n int) func(*execgw.RiskGuardianOptions) {
	return func(o *execgw.RiskGuardianOptions) { o.Policy.MaxOpenPositions = n }
}

// a138HoldElsewhere 는 rig 원장에 대사가 접어 넣은 외부 보유 한 종목(999999)을 둠 — 문이 사는 종목과 다르므로 한 칸을 차지함.
func a138HoldElsewhere(t *testing.T, j *journal.Journal) {
	t.Helper()
	ctx := context.Background()
	watermark, err := j.FillWatermark(ctx, "999999")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := j.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
		AccountRef: "acct-7", Market: "kr", Symbol: "999999", Kind: journal.AdjustmentExternal,
		ExpectedPrevQuantity: "0", ExpectedFillWatermark: watermark, NewQuantity: "3",
		BrokerAsOf: "2026-03-30T00:29:00Z", Evidence: "holdings snapshot reported 3",
	}); err != nil {
		t.Fatalf("ApplyPositionAdjustment: %v", err)
	}
}

func requireOpenPositionRefusal(t *testing.T, err error) {
	t.Helper()
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) || refusal.Stage != execgw.StageIssuance || refusal.Reason != journal.IssueReasonLimitReached ||
		!strings.Contains(refusal.Detail, "MAX_OPEN_POSITIONS") {
		t.Fatalf("want an issuance LIMIT_REACHED refusal naming MAX_OPEN_POSITIONS, got %v", err)
	}
}

func TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum(t *testing.T) {
	rig := newGuardian(t, maxOpenPositions(1))
	a138HoldElsewhere(t, rig.journal)
	ctx := context.Background()
	_, err := rig.issue(ctx)
	requireOpenPositionRefusal(t, err)

	// 상한을 넘은 뒤에도 reduce-only 매도는 발급됨.
	sell := guardianIntent()
	sell.Side = "SELL"
	account := guardianAccount()
	account.HeldQuantity = "10"
	if _, err := rig.guardian.IssueReduction(ctx, execgw.ReductionIssuance{Intent: sell, Account: account, Reason: "stop"}); err != nil {
		t.Fatalf("a reduction must never be judged against the open position maximum: %v", err)
	}

	// 상한 2 면 같은 진입이 통과(외부 1 + 이번 1).
	fits := newGuardian(t, maxOpenPositions(2))
	a138HoldElsewhere(t, fits.journal)
	if _, err := fits.issue(ctx); err != nil {
		t.Fatalf("with a maximum of 2 the entry fits: %v", err)
	}
}

func TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum(t *testing.T) {
	rig := newGuardian(t, maxOpenPositions(1))
	a138HoldElsewhere(t, rig.journal)
	precheck, err := rig.guardian.PrecheckQFinalEntry(qFinalKRRequest(t, rig, "a138", 20))
	if err != nil {
		t.Fatal(err)
	}
	_, err = rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
	requireOpenPositionRefusal(t, err)
}

func TestA138TheLimitsDigestBindsTheOpenPositionMaximum(t *testing.T) {
	a := newGuardian(t, maxOpenPositions(6)).guardian
	b := newGuardian(t, maxOpenPositions(7)).guardian
	if a.LimitsDigest() == b.LimitsDigest() {
		t.Fatal("changing only the open position maximum must change the digest strategy bindings are sealed against")
	}
	if a.ExposureLimits() != b.ExposureLimits() {
		t.Fatal("the five-field snapshot a decision carries must not change with the open position maximum")
	}
	if got := a.MaxOpenPositionsLimit(); got != 6 {
		t.Fatalf("declared maximum = %d, want 6", got)
	}
	// 구현 리뷰 P3-2: 기본값이 아닌 정책도 그대로 진술해야 함(「6 을 돌려줌」 변이를 잡음).
	if got := b.MaxOpenPositionsLimit(); got != 7 {
		t.Fatalf("declared maximum = %d, want 7", got)
	}
}
