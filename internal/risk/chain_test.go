package risk

// chain_test.go is the port of StockOS `tests/test_guardian.py` (20 cases) onto
// the TossOS chain, plus the cases TossOS's own order and boundary rules need.
//
// # Case mapping (StockOS → TossOS)
//
// 이식 (7):
//
//	test_guardian_allows_dry_run_candidate_within_limits    → TestEntryWithinEveryLimitIsAllowed
//	test_guardian_blocks_oversized_order                    → TestOrderAboveThePerOrderCapIsRefused
//	test_guardian_requires_cash_for_order_plus_costs        → TestCashIsMeasuredWithCostsIncluded
//	test_guardian_blocks_entry_when_symbol_already_bought_today       → TestSameDayEntryLimitIsRefused
//	test_guardian_allows_same_day_reentry_when_dashboard_limit_allows → TestSameDayReentryWithinTheLimitIsAllowed
//	test_guardian_blocks_same_day_reentry_when_dashboard_limit_reached → TestSameDayEntryLimitIsRefused (동일 케이스, count=max)
//	test_guardian_blocks_same_day_reentry_until_dashboard_cooldown_elapses → TestSameDayCooldownIsRefused
//	test_guardian_blocks_entry_when_buy_order_is_pending    → TestPendingBuyBlocksAnotherEntry
//
// 구조 대체 (4) — risk-management의 "구조 대체" 열:
//
//	test_guardian_blocks_leveraged_inverse_symbol           → TestSymbolOutsideTheAllowlistIsRefused
//	    (레버리지/인버스 클래스 차단 → 심볼 allowlist. 분류 소스 `[미측정]`, P3 재설계)
//	test_guardian_blocks_buy_when_buy_pause_active          → TestOperatingModeBlocksAnEntry
//	test_guardian_blocks_buy_when_sell_only_mode_active     → TestOperatingModeBlocksAnEntry (HALT_ALL 행)
//	    (BUY_PAUSED/SELL_ONLY_MODE → 운영 모드. design D3: EXIT_ONLY 삭제)
//	test_guardian_allows_sell_when_buy_pause_active         → TestReductionPassesEveryEntryOnlyGate
//	test_guardian_sell_only_mode_takes_precedence_over_buy_pause → TestFirstFailureStopsTheChain
//	    ("두 차단이 겹치면 더 강한 코드가 표면화" → 고정 순서의 첫 실패 정지)
//
// 제외 (8) — risk-management의 "제외" 열, 각각의 사유:
//
//	test_guardian_blocks_live_when_live_disabled            LIVE_DISABLED — 인터록이 대체
//	test_guardian_accepts_auto_order_session_as_live_consent ARM/DASHBOARD 확인 — 게이트 flip 승인이 대체
//	test_guardian_blocks_new_symbol_when_max_positions_reached MAX_POSITIONS — P3
//	test_guardian_blocks_llm_veto                            LLM 게이트 — 미이식
//	test_guardian_uses_us_regular_session_for_us_entries     미국장 진입 시간창 — 미이식
//	test_guardian_us_time_window_is_symmetric_across_us_markets  같음
//	test_guardian_nasd_blocked_outside_us_session_window     같음
//	(test_guardian_blocks_llm_veto 외에 LLM confidence 케이스는 원본 파일에 없음)
//
// TossOS 신규 (원본에 카운터파트 없음): 체인 순서 고정, 총계 경계값 ≥, 자본 0
// 즉시 차단, 중복 주문, 위험 감소 경로의 비용 비게이트(§0.3), 입력 실패 fail-closed.

import (
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

var testNow = time.Date(2026, 7, 26, 1, 20, 0, 0, time.UTC)

func krw(amount string) riskcalc.Money { return riskcalc.Money{Amount: amount, Currency: "KRW"} }
func usd(amount string) riskcalc.Money { return riskcalc.Money{Amount: amount, Currency: "USD"} }

// usdPolicy is the default policy restated in USD, for the cases that check a
// US intent against limits in its own currency. The amounts are scaled, not
// converted: no FX rate belongs in a test fixture.
func usdPolicy() Policy {
	p := DefaultPolicy()
	p.MaxOrderNotional = usd("10000")
	p.MaxOpenExposure = usd("100000")
	p.MaxDailyLoss = usd("1000")
	p.RiskBudget = usd("1000")
	return p
}

// entryInput is an entry that passes every check, so each case can break exactly
// one thing. Its numbers: 9 shares at 10,000 (notional 90,000), stop 9,800
// (200 wide, R = 200), target 10,500 (RR 2.5), against the conservative default
// policy.
func entryInput() Input {
	return Input{
		Now: testNow,
		Intent: Intent{
			AccountRef:  "acct-1",
			Market:      costs.MarketKR,
			Symbol:      "005930",
			Side:        SideBuy,
			Quantity:    "9",
			LimitPrice:  "10000",
			StopPrice:   "9800",
			TargetPrice: "10500",
		},
		Account: AccountState{
			Mode:              ModeNormal,
			AllowedSymbols:    []string{"005930", "000660"},
			CashAvailable:     krw("1000000"),
			OpenExposure:      krw("0"),
			DailyRealizedLoss: krw("0"),
			AccountEquity:     krw("1000000"),
		},
		Policy: DefaultPolicy(),
		Costs:  costs.DefaultModel(),
	}
}

func requireAllowed(t *testing.T, in Input) Decision {
	t.Helper()
	got := Evaluate(in)
	if !got.Allowed {
		t.Fatalf("refused with %s: %s", got.Reason, got.Detail)
	}
	if got.Reason != ReasonAllowed {
		t.Fatalf("allowed decision carries reason %s, want %s", got.Reason, ReasonAllowed)
	}
	return got
}

func requireRefused(t *testing.T, in Input, want ReasonCode) Decision {
	t.Helper()
	got := Evaluate(in)
	if got.Allowed {
		t.Fatalf("allowed, want refusal %s", want)
	}
	if got.Reason != want {
		t.Fatalf("refused with %s (%s), want %s", got.Reason, got.Detail, want)
	}
	if strings.TrimSpace(got.Detail) == "" {
		t.Fatalf("%s carries no detail; an operator cannot act on a bare code", want)
	}
	return got
}

// --- the chain itself ---------------------------------------------------------

func TestEntryWithinEveryLimitIsAllowed(t *testing.T) {
	requireAllowed(t, entryInput())
}

func TestChainOrderIsFixed(t *testing.T) {
	// risk-management: 체인은 고정 순서로 평가되고 첫 실패에서 정지한다(SHALL),
	// and 순서의 권위는 이 표다. The order is an interface, not an implementation
	// detail: it decides which reason an operator sees, and docs/guardian-chain.md
	// documents this exact list.
	want := []string{
		"kill_switch",
		"operating_mode",
		"entry_gate_latch",
		"symbol_allowlist",
		"stop_contract",
		"order_size",
		"min_reward_risk",
		"cash",
		"same_day_reentry",
		"open_exposure",
		"daily_loss",
		"duplicate_order",
	}
	got := EntryChainSteps()
	if len(got) != len(want) {
		t.Fatalf("chain has %d steps %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("step %d is %q, want %q (full chain %v)", i, got[i], want[i], got)
		}
	}
}

func TestFirstFailureStopsTheChain(t *testing.T) {
	// Everything is wrong at once. The kill switch is first, so that is the code
	// recorded — the earlier the check, the more fundamental the violation, and a
	// post-mortem that names the deepest cause is the one worth having.
	in := entryInput()
	in.Account.KillSwitchActive = true
	in.Account.Mode = ModeHaltAll
	in.Account.EntryBlockedLatch = true
	in.Intent.Symbol = "NOTLISTED"
	in.Intent.StopPrice = ""
	in.Account.DuplicateOrder = true
	requireRefused(t, in, ReasonKillSwitchActive)

	// Remove the kill switch and the next one surfaces, and so on down.
	in.Account.KillSwitchActive = false
	requireRefused(t, in, ReasonOperatingModeBlocked)
	in.Account.Mode = ModeNormal
	requireRefused(t, in, ReasonEntryGateBlocked)
	in.Account.EntryBlockedLatch = false
	requireRefused(t, in, ReasonSymbolNotAllowed)
	in.Intent.Symbol = "005930"
	requireRefused(t, in, ReasonStopMissing)
}

func TestOperatingModeBlocksAnEntry(t *testing.T) {
	// D3's 3-mode table. ENTRY_BLOCKED and HALT_ALL both refuse an entry; the
	// difference between them is what they do to a *reduction* (below) and who
	// may enter them, not what they do here.
	for _, mode := range []OperatingMode{ModeEntryBlocked, ModeHaltAll} {
		in := entryInput()
		in.Account.Mode = mode
		got := requireRefused(t, in, ReasonOperatingModeBlocked)
		if !strings.Contains(got.Detail, string(mode)) {
			t.Fatalf("detail %q does not name the mode %s", got.Detail, mode)
		}
	}
}

func TestUnknownOperatingModeFailsClosed(t *testing.T) {
	// A mode nobody recognises is not NORMAL. This is the one place a typo in a
	// persisted enum could otherwise open the gate.
	in := entryInput()
	in.Account.Mode = OperatingMode("normal")
	requireRefused(t, in, ReasonOperatingModeBlocked)

	in.Account.Mode = ""
	requireRefused(t, in, ReasonOperatingModeBlocked)
}

func TestEntryGateLatchBlocksAnEntry(t *testing.T) {
	// The latch itself (401/403 · SLO · RECONCILE · recovery) lands in task 2.4;
	// what is fixed here is that the chain consults it, in this position, and
	// carries the latch's own reason into the detail so 2.4's mapping is visible.
	in := entryInput()
	in.Account.EntryBlockedLatch = true
	in.Account.EntryBlockedReason = "broker_auth_rejected"
	got := requireRefused(t, in, ReasonEntryGateBlocked)
	if !strings.Contains(got.Detail, "broker_auth_rejected") {
		t.Fatalf("detail %q drops the latch's own reason", got.Detail)
	}
}

func TestSymbolOutsideTheAllowlistIsRefused(t *testing.T) {
	in := entryInput()
	in.Intent.Symbol = "SOXS"
	requireRefused(t, in, ReasonSymbolNotAllowed)

	// An empty allowlist admits nothing. It is not "unrestricted": the allowlist
	// is what replaces StockOS's instrument-class blocks, so an empty one that
	// meant "everything" would silently re-admit every class it excludes.
	in = entryInput()
	in.Account.AllowedSymbols = nil
	requireRefused(t, in, ReasonSymbolNotAllowed)
}

func TestOrderAboveThePerOrderCapIsRefused(t *testing.T) {
	// 주문 단위 한도는 포함 상한(초과 시 차단): reaching the cap exactly is allowed,
	// one unit past it is not. This is the opposite tie rule from the aggregates
	// below, and the difference is deliberate (design D2).
	atQuantityCap := entryInput()
	atQuantityCap.Intent.Quantity = "100"     // exactly the quantity cap
	atQuantityCap.Intent.LimitPrice = "10000" // notional 1,000,000 = the notional cap
	atQuantityCap.Account.CashAvailable = krw("2000000")
	requireAllowed(t, atQuantityCap)

	overQuantity := entryInput()
	overQuantity.Intent.Quantity = "101"
	requireRefused(t, overQuantity, ReasonMaxOrderExceeded)

	// Notional over the cap while the quantity is still inside it.
	overNotional := entryInput()
	overNotional.Intent.Quantity = "100"
	overNotional.Intent.LimitPrice = "20000"
	overNotional.Intent.StopPrice = "19500"
	overNotional.Intent.TargetPrice = "21500"
	overNotional.Account.CashAvailable = krw("100000000")
	requireRefused(t, overNotional, ReasonMaxOrderExceeded)
}

func TestCashIsMeasuredWithCostsIncluded(t *testing.T) {
	// StockOS's case: cash exactly equals the notional, so only the fee makes it
	// unaffordable. risk-management merges its two codes into one
	// (INSUFFICIENT_CASH · 비용 포함) — see issues.md I4.
	in := entryInput()
	in.Account.CashAvailable = krw("90000")
	got := requireRefused(t, in, ReasonInsufficientCash)
	if !strings.Contains(got.Detail, "90000") {
		t.Fatalf("detail %q does not show the numbers compared", got.Detail)
	}

	// Cash that covers notional *and* cost passes: the boundary is inclusive.
	in.Account.CashAvailable = krw("90090")
	requireAllowed(t, in)
}

func TestPendingBuyBlocksAnotherEntry(t *testing.T) {
	in := entryInput()
	in.Account.PendingBuy = true
	requireRefused(t, in, ReasonPendingBuyOrderBlocked)
}

func TestSameDayEntryLimitIsRefused(t *testing.T) {
	// Default: 심볼당 당일 최대 2회 `[미검증]`. At the limit, another entry is
	// refused regardless of the cooldown.
	in := entryInput()
	in.Account.SameDayEntryCount = 2
	in.Account.LastEntryAt = testNow.Add(-4 * time.Hour)
	got := requireRefused(t, in, ReasonSameDayReentryBlocked)
	if !strings.Contains(got.Detail, "2") {
		t.Fatalf("detail %q does not show the count against the limit", got.Detail)
	}
}

func TestSameDayCooldownIsRefused(t *testing.T) {
	// Default cooldown 30분 `[미검증]`; the last entry was 10 minutes ago.
	in := entryInput()
	in.Account.SameDayEntryCount = 1
	in.Account.LastEntryAt = testNow.Add(-10 * time.Minute)
	requireRefused(t, in, ReasonSameDayReentryCooldownActive)
}

func TestSameDayReentryWithinTheLimitIsAllowed(t *testing.T) {
	in := entryInput()
	in.Account.SameDayEntryCount = 1
	in.Account.LastEntryAt = testNow.Add(-31 * time.Minute)
	requireAllowed(t, in)

	// The cooldown boundary is the elapsed instant itself: at exactly the
	// cooldown the wait is over. (Refusing here would make the cooldown
	// unbounded in the presence of a coarse clock.)
	in.Account.LastEntryAt = testNow.Add(-30 * time.Minute)
	requireAllowed(t, in)
}

func TestSameDayCountWithoutATimestampStillCounts(t *testing.T) {
	// A count with no last-entry instant must not skip the cooldown silently: the
	// count is the durable fact, the timestamp is the refinement.
	in := entryInput()
	in.Account.SameDayEntryCount = 1
	in.Account.LastEntryAt = time.Time{}
	requireRefused(t, in, ReasonSameDayReentryCooldownActive)
}

func TestOpenExposureBoundaryBlocksAtTheLimit(t *testing.T) {
	// 총계 한도의 경계값은 도달 시 차단(≥) — 예약 트랜잭션의 실장과 일치.
	// Projected exposure is 90,090 (notional + the over-estimated cost).
	in := entryInput()
	in.Policy.MaxOpenExposure = krw("90090")
	requireRefused(t, in, ReasonOpenExposureExceeded)

	in.Policy.MaxOpenExposure = krw("90091")
	requireAllowed(t, in)

	// Existing exposure counts toward the same total.
	in.Policy.MaxOpenExposure = krw("1000000")
	in.Account.OpenExposure = krw("909911")
	requireRefused(t, in, ReasonOpenExposureExceeded)
}

func TestDailyLossBoundaryBlocksAtTheLimit(t *testing.T) {
	// Equity is raised so the 1% ratio (1,000,000) is not the binding ceiling —
	// this case is about the absolute one.
	in := entryInput()
	in.Account.AccountEquity = krw("100000000")
	in.Account.DailyRealizedLoss = krw("100000") // exactly the default limit
	requireRefused(t, in, ReasonDailyLossLimitReached)

	in.Account.DailyRealizedLoss = krw("99999")
	requireAllowed(t, in)
}

func TestDailyLossRatioBlocksBeforeTheAbsoluteLimit(t *testing.T) {
	// 1% of a 1,000,000 account is 10,000 — well under the 100,000 absolute
	// limit, so the ratio is the binding one on a small account.
	in := entryInput()
	in.Account.DailyRealizedLoss = krw("10000")
	requireRefused(t, in, ReasonDailyLossLimitReached)

	in.Account.DailyRealizedLoss = krw("9999")
	requireAllowed(t, in)
}

func TestZeroOrNegativeEquityBlocksImmediately(t *testing.T) {
	// risk-management: (계좌자본 0 이하이면 즉시 차단). A ratio over zero equity is
	// not a large number, it is an undefined one.
	for _, equity := range []string{"0", "-1"} {
		in := entryInput()
		in.Account.AccountEquity = krw(equity)
		requireRefused(t, in, ReasonDailyLossLimitReached)
	}
}

func TestNegativeAggregatesAreRefusedRatherThanCompared(t *testing.T) {
	// The aggregates are magnitudes by their producer's contract
	// (riskcalc.DailyLoss returns max(0, −net); GrossLongExposure sums
	// non-negative terms). A caller that passed signed P&L instead would hand a
	// day that lost 100,000 over as "−100,000" — and every comparison in the
	// chain would read that as a limit with room left. It is the one input error
	// in this file that opens the gate, so it fails closed by name.
	loss := entryInput()
	loss.Account.AccountEquity = krw("100000000")
	loss.Account.DailyRealizedLoss = krw("-100000")
	got := requireRefused(t, loss, ReasonInputUnavailable)
	if !strings.Contains(got.Detail, "-100000") {
		t.Fatalf("detail %q does not show the value it refused", got.Detail)
	}

	exposure := entryInput()
	exposure.Account.OpenExposure = krw("-5000000")
	requireRefused(t, exposure, ReasonInputUnavailable)

	// Zero is not negative: an account with nothing open and a day with no loss
	// are the ordinary case, and they still pass.
	clean := entryInput()
	clean.Account.OpenExposure = krw("0")
	clean.Account.DailyRealizedLoss = krw("0")
	requireAllowed(t, clean)
}

func TestDuplicateOrderIsRefused(t *testing.T) {
	in := entryInput()
	in.Account.DuplicateOrder = true
	requireRefused(t, in, ReasonDuplicateOrder)
}

// --- the risk-reducing path ---------------------------------------------------

func reductionInput() Input {
	in := entryInput()
	in.Intent.Side = SideSell
	in.Intent.Quantity = "9"
	in.Intent.StopPrice = ""
	in.Intent.TargetPrice = ""
	in.Account.HeldQuantity = "9"
	return in
}

func TestReductionPassesEveryEntryOnlyGate(t *testing.T) {
	// The mode table (risk-management): RISK_REDUCING is 허용 in NORMAL,
	// ENTRY_BLOCKED *and* HALT_ALL. The kill switch is BLOCK-ONLY — 어떤 소비자도
	// 강제청산을 유발하지 않는다, and equally it must not prevent one. The
	// allowlist is an entry control: a symbol removed from it must still be
	// closeable, or the removal would strand the position.
	in := reductionInput()
	in.Account.KillSwitchActive = true
	in.Account.Mode = ModeHaltAll
	in.Account.EntryBlockedLatch = true
	in.Account.AllowedSymbols = nil
	in.Account.CashAvailable = krw("0")
	in.Account.OpenExposure = krw("999999999")
	in.Account.DailyRealizedLoss = krw("999999999")
	in.Account.SameDayEntryCount = 9
	in.Account.PendingBuy = true
	requireAllowed(t, in)
}

func TestReductionIsNeverCostGated(t *testing.T) {
	// §0.3 / exit-policy: 청산·부분익절 발의는 예상 비용을 이유로 차단되지 않는다
	// (SHALL NOT — StockOS SELL_COST_BUFFER 미이식). The behavioural half of the
	// structural test in internal/costs: with every rate at the cap — the most
	// expensive configuration that loads at all — the reduction still passes.
	atCap := "0.05"
	model, err := costs.NewModel(map[string]string{
		costs.KeyKRSellCommissionRate: atCap,
		costs.KeyKRSellTaxRate:        atCap,
		costs.KeyKRBuyCommissionRate:  atCap,
	})
	if err != nil {
		t.Fatalf("NewModel: %v", err)
	}
	in := reductionInput()
	in.Costs = model
	in.Account.CashAvailable = krw("0")
	requireAllowed(t, in)

	// Even an absent cost model does not stop a liquidation: the entry path
	// refuses one (a trade cannot be priced as free), the reduction path never
	// asks, because "we could not price the exit" must not become "so hold it".
	in.Costs = costs.Model{}
	requireAllowed(t, in)
}

func TestReductionAboveHoldingsIsRefused(t *testing.T) {
	// risk-management: 매도는 보유수량 이하 reduce-only, short 노출은 구조적으로
	// 금지된다(SHALL NOT).
	in := reductionInput()
	in.Intent.Quantity = "10"
	in.Account.HeldQuantity = "9"
	requireRefused(t, in, ReasonSellExceedsHoldings)

	// Exactly the held quantity is a flatten, not a short.
	in.Intent.Quantity = "9"
	requireAllowed(t, in)
}

func TestReductionWithoutAKnownHoldingFailsClosed(t *testing.T) {
	// An unknown holding is not "sell as much as you like": without the number
	// there is no way to tell a flatten from a short.
	in := reductionInput()
	in.Account.HeldQuantity = ""
	requireRefused(t, in, ReasonInputUnavailable)
}

func TestReductionSizeMustBePositive(t *testing.T) {
	in := reductionInput()
	in.Intent.Quantity = "0"
	requireRefused(t, in, ReasonInvalidOrderSize)
}

// --- fail-closed inputs --------------------------------------------------------

func TestUnusableInputsRefuseRatherThanDefault(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*Input)
		want ReasonCode
	}{
		{"no side", func(in *Input) { in.Intent.Side = "" }, ReasonInputUnavailable},
		{"no symbol", func(in *Input) { in.Intent.Symbol = "  " }, ReasonInputUnavailable},
		{"no account", func(in *Input) { in.Intent.AccountRef = "" }, ReasonInputUnavailable},
		{"no evaluation instant", func(in *Input) { in.Now = time.Time{} }, ReasonInputUnavailable},
		{"unknown market", func(in *Input) { in.Intent.Market = costs.Market("KR") }, ReasonInputUnavailable},
		{"unparsable cash", func(in *Input) { in.Account.CashAvailable = krw("plenty") }, ReasonInputUnavailable},
		{"cash without currency", func(in *Input) { in.Account.CashAvailable = riskcalc.Money{Amount: "1000000"} }, ReasonInputUnavailable},
		{"foreign-currency cash", func(in *Input) {
			in.Account.CashAvailable = riskcalc.Money{Amount: "1000000", Currency: "USD"}
		}, ReasonInputUnavailable},
		{"unset notional cap", func(in *Input) { in.Policy.MaxOrderNotional = riskcalc.Money{} }, ReasonInputUnavailable},
		{"unset exposure cap", func(in *Input) { in.Policy.MaxOpenExposure = riskcalc.Money{} }, ReasonInputUnavailable},
		{"unset daily loss cap", func(in *Input) { in.Policy.MaxDailyLoss = riskcalc.Money{} }, ReasonInputUnavailable},
		{"unset risk budget", func(in *Input) { in.Policy.RiskBudget = riskcalc.Money{} }, ReasonInputUnavailable},
		{"unset minimum RR", func(in *Input) { in.Policy.MinRewardRisk = "" }, ReasonInputUnavailable},
		{"absent cost model", func(in *Input) { in.Costs = costs.Model{} }, ReasonInputUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := entryInput()
			tc.mut(&in)
			requireRefused(t, in, tc.want)
		})
	}
}

func TestForeignCurrencyIntentIsNotSizedAgainstADomesticBudget(t *testing.T) {
	// Ported from a090 test_us_without_fx_fails_closed_without_mixed_currency_qty:
	// KRW 예산 ÷ USD 폭 = 통화 혼합 수량 차단. Normalisation belongs to
	// riskcalc.convert with its staleness bound, so until 4.x supplies converted
	// inputs a US intent against a KRW policy refuses.
	in := entryInput()
	in.Intent.Market = costs.MarketUS
	in.Intent.LimitPrice = "8"
	in.Intent.StopPrice = "7.88"
	in.Intent.TargetPrice = "8.30"
	in.Intent.Quantity = "10"
	requireRefused(t, in, ReasonInputUnavailable)

	// …and the currency-free rungs before it still fire first (a090
	// test_us_market_stop_rr_rungs_still_enforced: stop·RR rung은 통화를
	// 건너지 않으므로 전 시장 공통).
	in.Intent.StopPrice = ""
	requireRefused(t, in, ReasonStopMissing)

	// With a policy configured in the intent's own currency the size rung passes
	// and the RR rung is reached — the rungs before the money are the same for
	// every market. 8.20 clears the US break-even (≈8.1699 at the placeholder
	// rates) but is only 1.67R, so the refusal is the RR rung.
	in = entryInput()
	in.Intent.Market = costs.MarketUS
	in.Intent.LimitPrice = "8"
	in.Intent.StopPrice = "7.88"
	in.Intent.TargetPrice = "8.20"
	in.Intent.Quantity = "10"
	in.Policy = usdPolicy()
	in.Account.CashAvailable = usd("10000")
	in.Account.OpenExposure = usd("0")
	in.Account.DailyRealizedLoss = usd("0")
	in.Account.AccountEquity = usd("10000")
	requireRefused(t, in, ReasonMinRRNotMet)

	// …and the same intent at 2.5R passes every rung.
	in.Intent.TargetPrice = "8.30"
	requireAllowed(t, in)
}

func TestEvaluateDoesNotMutateItsInput(t *testing.T) {
	// Ported from a090 test_build_trade_plan_does_not_mutate_inputs. The chain is
	// a pure function of the snapshot it is handed; a chain that edited its input
	// would make two evaluations of the same intent disagree.
	in := entryInput()
	before := entryInput()
	Evaluate(in)

	if in.Intent != before.Intent {
		t.Fatalf("intent mutated: %+v → %+v", before.Intent, in.Intent)
	}
	if len(in.Account.AllowedSymbols) != len(before.Account.AllowedSymbols) {
		t.Fatalf("allowlist mutated: %v", in.Account.AllowedSymbols)
	}
	for i := range before.Account.AllowedSymbols {
		if in.Account.AllowedSymbols[i] != before.Account.AllowedSymbols[i] {
			t.Fatalf("allowlist mutated: %v → %v", before.Account.AllowedSymbols, in.Account.AllowedSymbols)
		}
	}
	if in.Policy != before.Policy || in.Costs != before.Costs {
		t.Fatal("policy or cost model mutated")
	}
}

func TestEvaluateIsDeterministic(t *testing.T) {
	in := entryInput()
	in.Account.DailyRealizedLoss = krw("10000")
	first := Evaluate(in)
	for i := 0; i < 8; i++ {
		if got := Evaluate(in); got != first {
			t.Fatalf("evaluation %d returned %+v, first returned %+v", i, got, first)
		}
	}
}
