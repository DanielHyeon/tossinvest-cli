package execgw

// a137 Guardian 일일 손실 측정 · 바닥 규칙 · 계좌 진입 차단 분류 시험(내부 — 원장 읽기 자리를 대체해 행을 넣음).

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a137Now = KST 2026-10-14 10:30 — 일 키 2026-10-14.
var a137Now = time.Date(2026, 10, 14, 1, 30, 0, 0, time.UTC)

func a137Guardian(t *testing.T) (*RiskGuardian, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(a137Now)
	j, err := journal.Open(context.Background(), journal.Options{Path: filepath.Join(t.TempDir(), "journal.db"), Clock: clk,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = j.Close() })
	board := NewManualBuyBoard()
	board.Publish(ManualBuySnapshot{AsOf: a137Now, Day: "2026-10-14", ByCurrency: map[string]string{}})
	g, err := NewRiskGuardian(RiskGuardianOptions{Journal: j, Clock: clk, AccountRef: "acct-7", Policy: risk.DefaultPolicy(),
		Costs: costs.DefaultModel(), PolicyVersion: "a137/test", ManualBuys: board})
	if err != nil {
		t.Fatal(err)
	}
	return g, j, clk
}

func a137Loss(pnl string) journal.DailyLossRow {
	return journal.DailyLossRow{PositionID: "p-" + pnl, Market: "kr", Symbol: "005930", OpenedAt: "2026-10-14T00:10:00Z",
		ClosedAt: "2026-10-14T01:00:00Z", Outcome: &pnl}
}

func a137Rows(rows []journal.DailyLossRow, err error, calls *int) func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) {
	return func(context.Context, string, time.Time, time.Time) ([]journal.DailyLossRow, error) {
		if calls != nil {
			*calls++
		}
		return rows, err
	}
}

func a137Account(callerLoss string) risk.AccountState {
	return risk.AccountState{Mode: risk.ModeNormal, AllowedSymbols: []string{"005930"},
		CashAvailable:     riskcalc.Money{Amount: "5000000", Currency: "KRW"},
		OpenExposure:      riskcalc.Money{Amount: "0", Currency: "KRW"},
		DailyRealizedLoss: riskcalc.Money{Amount: callerLoss, Currency: "KRW"},
		AccountEquity:     riskcalc.Money{Amount: "10000000", Currency: "KRW"}}
}

func a137Intent() risk.Intent {
	return risk.Intent{Market: costs.MarketKR, Symbol: "005930", Side: risk.SideBuy, Quantity: "10",
		LimitPrice: "70000", StopPrice: "69000", TargetPrice: "72000"}
}

func a137Issue(t *testing.T, g *RiskGuardian, account risk.AccountState) error {
	t.Helper()
	_, err := g.IssueEntry(context.Background(), EntryIssuance{Intent: a137Intent(), Account: account,
		Collect: func(context.Context, int) (ExposureSnapshot, error) {
			return ExposureSnapshot{}, errors.New("collector must not be reached")
		}})
	return err
}

func a137Refusal(t *testing.T, err error) *IssueRefusal {
	t.Helper()
	var refusal *IssueRefusal
	if !errors.As(err, &refusal) {
		t.Fatalf("err = %v, want an IssueRefusal", err)
	}
	return refusal
}

// 서명 매니페스트(호출자)가 0 이어도 원장 손실이 한도 이상이면 거절 + ENTRY_BLOCKED 강화.
func TestA137TheLedgerLossRefusesAZeroCallerAndEscalates(t *testing.T) {
	g, j, _ := a137Guardian(t)
	g.dailyLossRows = a137Rows([]journal.DailyLossRow{a137Loss("-60000"), a137Loss("-50000")}, nil, nil)
	refusal := a137Refusal(t, a137Issue(t, g, a137Account("0")))
	if refusal.Reason != string(risk.ReasonDailyLossLimitReached) {
		t.Fatalf("refusal = %+v, want DAILY_LOSS_LIMIT_REACHED from the measured 110000", refusal)
	}
	mode, err := j.CurrentOperatingMode(context.Background(), "acct-7")
	if err != nil || mode.Mode != journal.ModeEntryBlocked || mode.Cause != journal.ModeTriggerDailyLossLimit {
		t.Fatalf("mode = %+v err = %v, want ENTRY_BLOCKED by the daily-loss trigger", mode, err)
	}
}

// 둘 다 유효하면 큰 쪽 — 호출자 값이 더 크면 호출자 값으로 판정.
func TestA137ALargerValidCallerValueWins(t *testing.T) {
	g, _, _ := a137Guardian(t)
	g.dailyLossRows = a137Rows([]journal.DailyLossRow{a137Loss("-110000")}, nil, nil)
	refusal := a137Refusal(t, a137Issue(t, g, a137Account("120000")))
	if !strings.Contains(refusal.Detail, "120000") {
		t.Fatalf("detail = %q, want the larger caller value 120000", refusal.Detail)
	}
	g.dailyLossRows = a137Rows([]journal.DailyLossRow{a137Loss("-130000")}, nil, nil)
	refusal = a137Refusal(t, a137Issue(t, g, a137Account("120000")))
	if !strings.Contains(refusal.Detail, "130000") {
		t.Fatalf("detail = %q, want the larger measured value 130000", refusal.Detail)
	}
}

// 무효한 호출자 값은 측정값으로 대체하지 않음 — 지금처럼 거절.
func TestA137AnInvalidCallerValueIsNotReplaced(t *testing.T) {
	g, _, _ := a137Guardian(t)
	g.dailyLossRows = a137Rows(nil, nil, nil)
	for name, account := range map[string]risk.AccountState{
		"negative":    a137Account("-1"),
		"unparseable": a137Account("lots"),
		"currency mismatch": func() risk.AccountState {
			a := a137Account("0")
			a.DailyRealizedLoss.Currency = "USD"
			return a
		}(),
	} {
		refusal := a137Refusal(t, a137Issue(t, g, account))
		if refusal.Reason != string(risk.ReasonInputUnavailable) {
			t.Errorf("%s: refusal = %+v, want INPUT_UNAVAILABLE (the invalid caller value is not replaced)", name, refusal)
		}
	}
}

// 측정 실패는 호출자 값(0)으로 떨어지지 않음 — daily_loss 단 INPUT_UNAVAILABLE, 사유는 측정 실패.
func TestA137AMeasurementFailureNeverFallsBackToTheCaller(t *testing.T) {
	g, _, _ := a137Guardian(t)
	g.dailyLossRows = a137Rows(nil, errors.New("database is locked"), nil)
	refusal := a137Refusal(t, a137Issue(t, g, a137Account("0")))
	if refusal.Reason != string(risk.ReasonInputUnavailable) || !strings.Contains(refusal.Detail, "cannot be measured") {
		t.Fatalf("refusal = %+v, want INPUT_UNAVAILABLE naming the measurement failure", refusal)
	}
	if strings.Contains(refusal.Detail, "database is locked") {
		t.Fatalf("detail %q carries the raw error text", refusal.Detail)
	}
	// 성과 행 없는 오늘의 청산도 같은 거절.
	row := a137Loss("-1")
	row.Outcome = nil
	g.dailyLossRows = a137Rows([]journal.DailyLossRow{row}, nil, nil)
	if refusal := a137Refusal(t, a137Issue(t, g, a137Account("0"))); refusal.Reason != string(risk.ReasonInputUnavailable) {
		t.Fatalf("refusal = %+v, want INPUT_UNAVAILABLE for a close without an outcome", refusal)
	}
}

// 앞 단(kill switch)의 사유는 측정 실패에 가려지지 않음.
func TestA137TheKillSwitchReasonSurvivesAMeasurementFailure(t *testing.T) {
	g, _, _ := a137Guardian(t)
	g.dailyLossRows = a137Rows(nil, errors.New("read failed"), nil)
	account := a137Account("0")
	account.KillSwitchActive = true
	if refusal := a137Refusal(t, a137Issue(t, g, account)); refusal.Reason != string(risk.ReasonKillSwitchActive) {
		t.Fatalf("refusal = %+v, want KILL_SWITCH_ACTIVE", refusal)
	}
}

// 위험 축소(매도 · 손절)는 측정기를 부르지 않음.
func TestA137RiskReductionNeverMeasures(t *testing.T) {
	g, _, _ := a137Guardian(t)
	calls := 0
	g.dailyLossRows = a137Rows(nil, errors.New("read failed"), &calls)
	_, err := g.IssueReduction(context.Background(), ReductionIssuance{
		Intent:  risk.Intent{Market: costs.MarketKR, Symbol: "005930", Side: risk.SideSell, Quantity: "10", LimitPrice: "69000"},
		Account: risk.AccountState{HeldQuantity: "10"}, Reason: "stop"})
	if err != nil {
		t.Fatalf("reduction refused: %v", err)
	}
	if calls != 0 {
		t.Fatalf("the reduction path measured the daily loss %d time(s)", calls)
	}
}

// 측정 토큰: 이 Guardian 의 것 · 10초 안 · 미래 아님 · 일 키 같음 · 봉인 오류 없음 — 하나라도 어긋나면 측정 실패.
func TestA137APrecheckMeasurementMustBeThisGuardiansFreshTodaysOne(t *testing.T) {
	g, _, clk := a137Guardian(t)
	g.dailyLossRows = a137Rows(nil, nil, nil)
	fresh := g.MeasureDailyLoss(context.Background())
	if _, err := g.dailyLossFloor(a137Account("0"), fresh, clk.Now(), nil); err != nil {
		t.Fatalf("a fresh measurement failed: %v", err)
	}
	other, _, _ := a137Guardian(t)
	other.dailyLossRows = a137Rows(nil, nil, nil)
	cases := map[string]struct {
		m   DailyLossMeasurement
		now time.Time
	}{
		"zero value":       {DailyLossMeasurement{}, clk.Now()},
		"another guardian": {other.MeasureDailyLoss(context.Background()), clk.Now()},
		"stale":            {fresh, clk.Now().Add(11 * time.Second)},
		"future":           {fresh, clk.Now().Add(-time.Second)},
	}
	for name, c := range cases {
		if _, err := g.dailyLossFloor(a137Account("0"), c.m, c.now, nil); err == nil {
			t.Errorf("%s: want a measurement failure", name)
		}
	}
	// KST 08:00 경계를 걸친 측정: 측정은 07:59:58 KST, 판정은 08:00:01 KST.
	clk.Set(time.Date(2026, 10, 13, 22, 59, 58, 0, time.UTC))
	before := g.MeasureDailyLoss(context.Background())
	if _, err := g.dailyLossFloor(a137Account("0"), before, before.at.Add(3*time.Second), nil); err == nil {
		t.Fatal("a measurement of the previous account day must not be used")
	}
	failed := g.MeasureDailyLoss(context.Background())
	failed.err = errors.New("sealed failure")
	if _, err := g.dailyLossFloor(a137Account("0"), failed, failed.at, nil); err == nil {
		t.Fatal("a sealed measurement failure must fail the floor")
	}
}

// 계좌 진입 차단 분류: AUTO · 일일 손실 trigger 의 ENTRY_BLOCKED 투영만 조용한 차단. 나머지는 오늘과 같음.
func a137Gate(t *testing.T) (*Gateway, *EntryGate) {
	t.Helper()
	gate := NewEntryGate(clock.NewFake(a137Now), map[RequiredQuery]time.Duration{})
	return &Gateway{entry: gate, accountRef: "acct-7"}, gate
}

func a137Project(gate *EntryGate, seq int64, mode, actor, cause string) {
	gate.ProjectOperatingMode(journal.OperatingModeRecord{AccountRef: "acct-7", Mode: mode, Actor: actor, Cause: cause, Seq: seq})
}

func a137IsAccountBlock(err error) bool {
	var block *DailyLossEntryBlock
	return errors.As(err, &block)
}

func TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock(t *testing.T) {
	ctx := context.Background()
	gw, gate := a137Gate(t)
	a137Project(gate, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
	_, err := gw.ObserveStrategyEntryGate(ctx, "kr", "005930")
	var rejected *RejectedError
	if !a137IsAccountBlock(err) || !errors.As(err, &rejected) || rejected.Reason != ReasonOperatingModeBlocked {
		t.Fatalf("err = %v, want the mode refusal marked as a daily-loss account block", err)
	}

	for name, setup := range map[string]func(*EntryGate){
		"operator with the trigger spelling": func(g *EntryGate) {
			a137Project(g, 1, journal.ModeEntryBlocked, journal.ModeActorOperator, journal.ModeTriggerDailyLossLimit)
		},
		"automatic other trigger": func(g *EntryGate) {
			a137Project(g, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, "CREDENTIAL_REJECTED")
		},
		"halt all": func(g *EntryGate) {
			a137Project(g, 1, journal.ModeHaltAll, journal.ModeActorOperator, "operator halt")
		},
		"boot restore failure": func(g *EntryGate) {
			g.Block(ReasonOperatingModeBlocked, "operating mode could not be restored at startup")
		},
		"daily loss plus a reconcile latch": func(g *EntryGate) {
			a137Project(g, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
			g.Block(ReasonReconcileMismatch, "reconcile")
		},
		// latchOrder 밖의 사유는 모드보다 **뒤**에 읽히므로 관측 거절은 모드 문구가 됨 — 그래도 latch 가 둘이면 조용한 차단이 아님.
		"daily loss plus a latch outside latchOrder": func(g *EntryGate) {
			a137Project(g, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
			g.Block(ReasonCode("A137_UNLISTED_LATCH"), "unlisted")
		},
		"daily loss plus a symbol latch": func(g *EntryGate) {
			a137Project(g, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
			g.BlockSymbol("kr", "005930", ReasonReconcileMismatch, "symbol")
		},
	} {
		gw, gate := a137Gate(t)
		setup(gate)
		_, err := gw.ObserveStrategyEntryGate(ctx, "kr", "005930")
		if err == nil || a137IsAccountBlock(err) {
			t.Errorf("%s: err = %v, want an ordinary refusal (today's behaviour)", name, err)
		}
	}
}

// 제출 직전 재확인(펜스): 세대 변화의 원인이 AUTO 일일 손실 모드면 계좌 차단, 아니면 지금처럼.
func TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange(t *testing.T) {
	ctx := context.Background()
	gw, gate := a137Gate(t)
	expected, err := gw.ObserveStrategyEntryGate(ctx, "kr", "005930")
	if err != nil {
		t.Fatal(err)
	}
	a137Project(gate, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
	err = gw.withStrategyEntryGateAuthority(ctx, expected, "kr", "005930", func() error { return nil })
	if !a137IsAccountBlock(err) {
		t.Fatalf("err = %v, want a daily-loss account block", err)
	}
	fenced := strategyFenceRejection(err)
	if fenced.Reason != ReasonStrategyDispatchFenced || !a137IsAccountBlock(fenced) {
		t.Fatalf("fenced = %+v, want a fenced rejection that keeps the account-block cause", fenced)
	}

	// 일일 손실 모드와 다른 latch 가 함께 생기면 펜스는 오늘처럼(조용한 차단 아님).
	gw3, gate3 := a137Gate(t)
	expected3, err := gw3.ObserveStrategyEntryGate(ctx, "kr", "005930")
	if err != nil {
		t.Fatal(err)
	}
	a137Project(gate3, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
	gate3.Block(ReasonReconcileMismatch, "reconcile")
	if err := gw3.withStrategyEntryGateAuthority(ctx, expected3, "kr", "005930", func() error { return nil }); err == nil || a137IsAccountBlock(err) {
		t.Fatalf("err = %v, want an ordinary fence when another latch joins the daily-loss mode", err)
	}

	// P2-1: 순서를 바꿔도(reconcile latch 가 먼저, 일일 손실 투영이 뒤) 표지 없음 — 오늘처럼 시끄러운 펜스.
	gw4, gate4 := a137Gate(t)
	expected4, err := gw4.ObserveStrategyEntryGate(ctx, "kr", "005930")
	if err != nil {
		t.Fatal(err)
	}
	gate4.Block(ReasonReconcileMismatch, "reconcile")
	a137Project(gate4, 1, journal.ModeEntryBlocked, journal.ModeActorAuto, journal.ModeTriggerDailyLossLimit)
	if err := gw4.withStrategyEntryGateAuthority(ctx, expected4, "kr", "005930", func() error { return nil }); err == nil || a137IsAccountBlock(err) {
		t.Fatalf("err = %v, want an ordinary fence (reconcile latch before the daily-loss projection)", err)
	}

	gw2, gate2 := a137Gate(t)
	expected2, err := gw2.ObserveStrategyEntryGate(ctx, "kr", "005930")
	if err != nil {
		t.Fatal(err)
	}
	gate2.Block(ReasonReconcileMismatch, "reconcile")
	err = gw2.withStrategyEntryGateAuthority(ctx, expected2, "kr", "005930", func() error { return nil })
	if err == nil || a137IsAccountBlock(err) {
		t.Fatalf("err = %v, want today's ordinary fence", err)
	}
	if fenced := strategyFenceRejection(err); a137IsAccountBlock(fenced) || fenced.Reason != ReasonStrategyDispatchFenced {
		t.Fatalf("fenced = %+v, want a plain fenced rejection", fenced)
	}
}

// 강화 입구: 성공하면 계좌 차단, 실패하면 차단이 아닌 오류(조용히 삼켜지지 않음).
func TestA137EscalationSucceedsAsABlockAndFailsLoudly(t *testing.T) {
	g, j, _ := a137Guardian(t)
	refusal := chainRefusal(risk.Decision{Reason: risk.ReasonDailyLossLimitReached, Detail: "limit"})
	if err := g.EscalateDailyLossLimit(context.Background(), refusal); !a137IsAccountBlock(err) {
		t.Fatalf("err = %v, want an account block", err)
	}
	if mode, err := j.CurrentOperatingMode(context.Background(), "acct-7"); err != nil || mode.Mode != journal.ModeEntryBlocked {
		t.Fatalf("mode = %+v err = %v", mode, err)
	}
	_ = j.Close()
	err := g.EscalateDailyLossLimit(context.Background(), refusal)
	if err == nil || a137IsAccountBlock(err) {
		t.Fatalf("err = %v, want a loud escalation failure that is not a quiet block", err)
	}
	var issue *IssueRefusal
	if !errors.As(err, &issue) {
		t.Fatalf("the refusal must stay in the joined error: %v", err)
	}
}
