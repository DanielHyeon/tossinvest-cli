package risk

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func TestAccountBaseGuardianPairedKRUS(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	policy := accountBaseTestPolicy()

	tests := []struct {
		name, symbol, quote, cash, rate, haircut, wantExposure string
		market                                                 costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, symbol: "005930", quote: "KRW", cash: "30", rate: "1", haircut: "1", wantExposure: "30"},
		{name: "US official", market: costs.MarketUS, symbol: "AAPL", quote: "USD", cash: "30", rate: "1.2", haircut: "1.05", wantExposure: "38"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, test.cash, policy, model, fx)
			if verdict := Evaluate(in); !verdict.Allowed {
				t.Fatalf("paired account-base entry refused: %s (%s)", verdict.Reason, verdict.Detail)
			}
			exposure, verdict := EntryExposureValue(in)
			if !verdict.Allowed {
				t.Fatalf("account-base exposure refused: %s (%s)", verdict.Reason, verdict.Detail)
			}
			if exposure.Currency != "KRW" || exposure.Amount != test.wantExposure {
				t.Fatalf("exposure = %+v, want %s KRW", exposure, test.wantExposure)
			}
			if fx.QuoteCurrency() != test.quote || fx.AccountCurrency() != "KRW" || fx.Digest() == "" || !fx.EvaluatedAt().Equal(now) {
				t.Fatalf("opaque FX binding lost scope: quote=%q base=%q digest=%q at=%s",
					fx.QuoteCurrency(), fx.AccountCurrency(), fx.Digest(), fx.EvaluatedAt())
			}
		})
	}
}

func TestAccountBaseCashRemainsQuoteCurrencyPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	policy := accountBaseTestPolicy()
	for _, test := range []struct {
		name, symbol, quote, rate, haircut string
		market                             costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, symbol: "005930", quote: "KRW", rate: "1", haircut: "1"},
		{name: "US official", market: costs.MarketUS, symbol: "AAPL", quote: "USD", rate: "1400", haircut: "1.01"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, "30", policy, model, fx)
			if verdict := Evaluate(in); !verdict.Allowed {
				t.Fatalf("quote cash at exact boundary refused: %s (%s)", verdict.Reason, verdict.Detail)
			}
			in.Account.CashAvailable.Amount = "29"
			verdict := Evaluate(in)
			if verdict.Allowed || verdict.Reason != ReasonInsufficientCash || verdict.Step != "cash" {
				t.Fatalf("quote cash shortfall = %+v, want cash refusal", verdict)
			}
		})
	}
}

func TestAccountBaseStrategyQuantityFloorsPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	policy := accountBaseTestPolicy()
	policy.MaxOrderNotional.Amount = "1000"
	policy.RiskBudget.Amount = "100"
	for _, test := range []struct {
		name, quote, rate, haircut, want string
		market                           costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, quote: "KRW", rate: "1", haircut: "1", want: "100"},
		{name: "US official", market: costs.MarketUS, quote: "USD", rate: "2", haircut: "1.1", want: "45"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			got, err := AccountBaseStrategyEntryQuantity(policy, test.market, "10", "9", fx)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("quantity = %s, want conservative floor %s", got, test.want)
			}
		})
	}
	if _, err := AccountBaseStrategyEntryQuantity(policy, costs.MarketKR, "10", "9", AccountBaseFX{}); !errors.Is(err, ErrAccountBaseFXUnavailable) {
		t.Fatalf("missing KR identity error = %v, want ErrAccountBaseFXUnavailable", err)
	}
}

func TestAccountBaseStrategyQuantityRequiresExplicitFXPaired(t *testing.T) {
	policy := accountBaseTestPolicy()
	for _, market := range []costs.Market{costs.MarketKR, costs.MarketUS} {
		if _, err := AccountBaseStrategyEntryQuantity(policy, market, "10", "9", AccountBaseFX{}); !errors.Is(err, ErrAccountBaseFXUnavailable) {
			t.Fatalf("market %s zero FX error = %v, want ErrAccountBaseFXUnavailable", market, err)
		}
	}
}

func TestAccountBaseOrderNotionalRoundsUpPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	for _, test := range []struct {
		name, symbol, quote, rate, haircut, below, exact string
		market                                           costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, symbol: "005930", quote: "KRW", rate: "1", haircut: "1", below: "29", exact: "30"},
		{name: "US official", market: costs.MarketUS, symbol: "AAPL", quote: "USD", rate: "1.2", haircut: "1.05", below: "37", exact: "38"},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := accountBaseTestPolicy()
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, "30", policy, model, fx)
			in.Policy.MaxOrderNotional.Amount = test.below
			verdict := Evaluate(in)
			if verdict.Allowed || verdict.Reason != ReasonMaxOrderExceeded || verdict.Step != "order_size" {
				t.Fatalf("rounded-up notional = %+v, want max order refusal", verdict)
			}
			in.Policy.MaxOrderNotional.Amount = test.exact
			if verdict = Evaluate(in); !verdict.Allowed {
				t.Fatalf("inclusive rounded ceiling refused: %+v", verdict)
			}
		})
	}
}

func TestAccountBaseExposureValueCeilsPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	policy := accountBaseTestPolicy()
	for _, test := range []struct {
		name, symbol, quote, rate, haircut, want string
		market                                   costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, symbol: "005930", quote: "KRW", rate: "1", haircut: "1", want: "30"},
		{name: "US official fractional", market: costs.MarketUS, symbol: "AAPL", quote: "USD", rate: "1.2", haircut: "1.05", want: "38"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, "30", policy, model, fx)
			got, verdict := EntryExposureValue(in)
			if !verdict.Allowed || got.Currency != "KRW" || got.Amount != test.want {
				t.Fatalf("exposure = %+v verdict=%+v, want %s KRW", got, verdict, test.want)
			}
		})
	}
}

func TestAccountBaseGatewayNotionalCeilsPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name, quote, rate, haircut, want string
		market                           costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, quote: "KRW", rate: "1", haircut: "1", want: "30"},
		{name: "US official", market: costs.MarketUS, quote: "USD", rate: "1.2", haircut: "1.05", want: "38"},
	} {
		t.Run(test.name, func(t *testing.T) {
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			got, err := AccountBaseOrderNotional(now, test.market, "30", fx)
			if err != nil {
				t.Fatal(err)
			}
			if got.Amount != test.want || got.Currency != "KRW" {
				t.Fatalf("account-base notional = %+v, want %s KRW", got, test.want)
			}
		})
	}
	if _, err := AccountBaseOrderNotional(now, costs.MarketUS, "30", AccountBaseFX{}); !errors.Is(err, ErrAccountBaseFXUnavailable) {
		t.Fatalf("missing Gateway FX error = %v, want ErrAccountBaseFXUnavailable", err)
	}
}

func TestAccountBaseOpenExposureBoundaryPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	for _, test := range []struct {
		name, symbol, quote, rate, haircut, existing, cap string
		market                                            costs.Market
	}{
		{name: "KR after US base usage", market: costs.MarketKR, symbol: "005930", quote: "KRW", rate: "1", haircut: "1", existing: "40", cap: "70"},
		{name: "US after KR base usage", market: costs.MarketUS, symbol: "AAPL", quote: "USD", rate: "1.2", haircut: "1.05", existing: "32", cap: "70"},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := accountBaseTestPolicy()
			policy.MaxOpenExposure.Amount = test.cap
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, "30", policy, model, fx)
			in.Account.OpenExposure = riskcalc.Money{Amount: test.existing, Currency: "KRW"}
			verdict := Evaluate(in)
			if verdict.Allowed || verdict.Reason != ReasonOpenExposureExceeded || verdict.Step != "open_exposure" {
				t.Fatalf("shared account-base boundary = %+v, want open exposure refusal", verdict)
			}
		})
	}
}

func TestAccountBaseDailyLossPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	for _, test := range []struct {
		name, symbol, quote, rate, haircut string
		market                             costs.Market
	}{
		{name: "KR identity", market: costs.MarketKR, symbol: "005930", quote: "KRW", rate: "1", haircut: "1"},
		{name: "US official", market: costs.MarketUS, symbol: "AAPL", quote: "USD", rate: "1400", haircut: "1.01"},
	} {
		t.Run(test.name, func(t *testing.T) {
			policy := accountBaseTestPolicy()
			policy.MaxDailyLoss.Amount = "100"
			fx := accountBaseFXForTest(now, test.quote, "KRW", test.rate, test.haircut)
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, "30", policy, model, fx)
			in.Account.DailyRealizedLoss = riskcalc.Money{Amount: "100", Currency: "KRW"}
			verdict := Evaluate(in)
			if verdict.Allowed || verdict.Reason != ReasonDailyLossLimitReached || verdict.Step != "daily_loss" {
				t.Fatalf("shared base loss boundary = %+v, want daily loss refusal", verdict)
			}
			in.Account.DailyRealizedLoss.Currency = test.quote
			if test.quote != "KRW" {
				verdict = Evaluate(in)
				if verdict.Allowed || verdict.Reason != ReasonInputUnavailable || verdict.Step != "daily_loss" {
					t.Fatalf("quote-currency account loss = %+v, want fail-closed base mismatch", verdict)
				}
			}
		})
	}
}

func TestAccountBaseFXFailsClosedPaired(t *testing.T) {
	now := time.Date(2026, 8, 4, 2, 0, 0, 0, time.UTC)
	model := zeroCostModel(t)
	policy := accountBaseTestPolicy()
	validKR := accountBaseFXForTest(now, "KRW", "KRW", "1", "1")
	validUS := accountBaseFXForTest(now, "USD", "KRW", "1400", "1.01")

	tests := []struct {
		name   string
		market costs.Market
		symbol string
		quote  string
		fx     AccountBaseFX
	}{
		{name: "KR wrong US scope", market: costs.MarketKR, symbol: "005930", quote: "KRW", fx: validUS},
		{name: "US missing authority", market: costs.MarketUS, symbol: "AAPL", quote: "USD"},
		{name: "US stale evaluation", market: costs.MarketUS, symbol: "AAPL", quote: "USD", fx: accountBaseFXForTest(now.Add(-time.Second), "USD", "KRW", "1400", "1.01")},
		{name: "US wrong base", market: costs.MarketUS, symbol: "AAPL", quote: "USD", fx: accountBaseFXForTest(now, "USD", "USD", "1", "1")},
		{name: "US tampered rate", market: costs.MarketUS, symbol: "AAPL", quote: "USD", fx: func() AccountBaseFX { fx := validUS; fx.rate = "1"; return fx }()},
		{name: "KR tampered identity", market: costs.MarketKR, symbol: "005930", quote: "KRW", fx: func() AccountBaseFX { fx := validKR; fx.haircut = "1.1"; return fx }()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := accountBaseEntryInput(now, test.market, test.symbol, test.quote, "100000", policy, model, test.fx)
			verdict := Evaluate(in)
			if verdict.Allowed || verdict.Reason != ReasonInputUnavailable || verdict.Step != "order_size" {
				t.Fatalf("unsafe FX verdict = %+v, want order-size input refusal", verdict)
			}
		})
	}

	if _, err := BindAccountBaseFX(now, costs.MarketUS, policy, officialfx.Evidence{}); !errors.Is(err, ErrAccountBaseFXUnavailable) {
		t.Fatalf("zero caller authority error = %v, want ErrAccountBaseFXUnavailable", err)
	}
}

func accountBaseTestPolicy() Policy {
	policy := DefaultPolicy()
	policy.MaxOrderQuantity = "100"
	policy.MaxOrderNotional = riskcalc.Money{Amount: "1000000", Currency: "KRW"}
	policy.MaxOpenExposure = riskcalc.Money{Amount: "1000000", Currency: "KRW"}
	policy.MaxDailyLoss = riskcalc.Money{Amount: "100000", Currency: "KRW"}
	policy.RiskBudget = riskcalc.Money{Amount: "100000", Currency: "KRW"}
	return policy
}

func accountBaseEntryInput(now time.Time, market costs.Market, symbol, quote, cash string, policy Policy, model costs.Model, fx AccountBaseFX) Input {
	return Input{
		Now: now,
		Intent: Intent{AccountRef: "acct-base", Market: market, Symbol: symbol, Side: SideBuy,
			Quantity: "3", LimitPrice: "10", StopPrice: "9", TargetPrice: "12"},
		Account: AccountState{
			Mode: ModeNormal, AllowedSymbols: []string{symbol},
			CashAvailable:     riskcalc.Money{Amount: cash, Currency: quote},
			OpenExposure:      riskcalc.Money{Amount: "0", Currency: "KRW"},
			DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"},
			AccountEquity:     riskcalc.Money{Amount: "1000000", Currency: "KRW"},
		},
		Policy: policy, Costs: model, AccountBaseFX: fx,
	}
}

func zeroCostModel(t *testing.T) costs.Model {
	t.Helper()
	overrides := make(map[string]string)
	for _, key := range costs.OverrideKeys() {
		overrides[key] = "0"
	}
	model, err := costs.NewModel(overrides)
	if err != nil {
		t.Fatal(err)
	}
	return model
}

func accountBaseFXForTest(at time.Time, quote, base, rate, haircut string) AccountBaseFX {
	source, version := officialfx.OfficialSource, officialfx.OfficialVersion
	if quote == base {
		source, version = officialfx.IdentitySource, officialfx.IdentityVersion
	}
	fx := AccountBaseFX{
		quoteCurrency: quote, accountCurrency: base, rate: rate, haircut: haircut,
		source: source, version: version, digest: "sha256:" + strings.Repeat("a", 64),
		observedAt: at.Add(-time.Minute), freshUntil: at.Add(time.Minute), evaluatedAt: at,
	}
	fx.seal = sealAccountBaseFX(fx)
	return fx
}
