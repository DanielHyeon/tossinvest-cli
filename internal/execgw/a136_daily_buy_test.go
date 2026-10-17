package execgw_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a136: 진입 발급 문 셋이 모두 정책의 하루 매수 상한을 예약 tx 에 실어 보내고, 도달하면 LIMIT_REACHED 로 아무것도 남기지 않음.
// 매도(IssueReduction)는 상한에 도달한 뒤에도 그대로 발급됨(§0.3).

func smallDailyBuy(amount string) func(*execgw.RiskGuardianOptions) {
	return func(o *execgw.RiskGuardianOptions) {
		o.Policy.MaxDailyBuyNotional = riskcalc.Money{Amount: amount, Currency: "KRW"}
	}
}

func requireDailyBuyRefusal(t *testing.T, err error) {
	t.Helper()
	var refusal *execgw.IssueRefusal
	if !errors.As(err, &refusal) || refusal.Stage != execgw.StageIssuance || refusal.Reason != journal.IssueReasonLimitReached ||
		!strings.Contains(refusal.Detail, "DAILY_BUY_NOTIONAL") {
		t.Fatalf("want an issuance LIMIT_REACHED refusal naming DAILY_BUY_NOTIONAL, got %v", err)
	}
}

func TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling(t *testing.T) {
	// 진입 하나의 예약 금액은 700,700 KRW — 상한 1,000,000 이면 첫째는 통과, 둘째는 1,401,400 ≥ 1,000,000 로 거부.
	rig := newGuardian(t, smallDailyBuy("1000000"))
	ctx := context.Background()
	if _, err := rig.issue(ctx); err != nil {
		t.Fatalf("the first buy fits under the ceiling: %v", err)
	}
	_, err := rig.issue(ctx)
	requireDailyBuyRefusal(t, err)

	// 상한 도달 뒤에도 reduce-only 매도는 발급됨.
	sell := guardianIntent()
	sell.Side = "SELL"
	account := guardianAccount()
	account.HeldQuantity = "10"
	if _, err := rig.guardian.IssueReduction(ctx, execgw.ReductionIssuance{Intent: sell, Account: account, Reason: "stop"}); err != nil {
		t.Fatalf("a reduction must never be judged against the daily buy ceiling: %v", err)
	}
}

func TestA136QFinalEntryRefusesTheBuyThatReachesTheDailyCeiling(t *testing.T) {
	rig := newGuardian(t, smallDailyBuy("1"))
	precheck, err := rig.guardian.PrecheckQFinalEntry(qFinalKRRequest(t, rig, "a136", 20))
	if err != nil {
		t.Fatal(err)
	}
	_, err = rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
	requireDailyBuyRefusal(t, err)
}

func TestA136TheLimitsDigestBindsTheDailyBuyCeiling(t *testing.T) {
	a := newGuardian(t, smallDailyBuy("10000000")).guardian
	b := newGuardian(t, smallDailyBuy("20000000")).guardian
	if a.LimitsDigest() == b.LimitsDigest() {
		t.Fatal("raising only the daily buy ceiling must change the digest strategy bindings are sealed against")
	}
	if a.ExposureLimits() != b.ExposureLimits() {
		t.Fatal("the five-field snapshot a decision carries must not change with the daily buy ceiling")
	}
	if got := a.DailyBuyNotionalLimit(); got != (riskcalc.Money{Amount: "10000000", Currency: "KRW"}) {
		t.Fatalf("declared ceiling = %+v", got)
	}
}

func TestA136AManualBuyOutsideTheEngineCountsTowardTheDay(t *testing.T) {
	// 수동 매수 400,000 KRW + 진입 700,700 = 1,100,700 ≥ 1,000,000 → 거부. 수동 매수 자체는 막히지 않음(엔진 밖).
	rig := newGuardian(t, smallDailyBuy("1000000"))
	board := rig.guardian.ManualBuyBoard()
	day := a136ZeroManualBuy(rig.clock.Now()).Day
	board.Publish(execgw.ManualBuySnapshot{AsOf: rig.clock.Now(), Day: day, ByCurrency: map[string]string{"KRW": "400000"}})
	_, err := rig.issue(context.Background())
	requireDailyBuyRefusal(t, err)
}

func TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted(t *testing.T) {
	rig := newGuardian(t, nil)
	day := a136ZeroManualBuy(rig.clock.Now()).Day
	rig.guardian.ManualBuyBoard().Publish(execgw.ManualBuySnapshot{AsOf: rig.clock.Now(), Day: day, ByCurrency: map[string]string{"USD": "10"}})
	if _, err := rig.issue(context.Background()); !errors.Is(err, execgw.ErrManualBuyUnknown) {
		t.Fatalf("a door with no FX authority must refuse when USD manual buys exist, got %v", err)
	}
}
