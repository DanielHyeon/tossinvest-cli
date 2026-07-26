package execgw_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// Fail-closed branch tests (harden-execution-base task 2.10).
//
// order-execution: "interactive auth challenge 요구, USD 주문의 통화 잔고 부족,
// 미지원 주문 유형은 제출 없이 거부하고 사유 코드와 함께 기록·통지" and
// "자동 환전·자동 승인은 금지된다".

// TestReasonCodeEnumIsStable holds the reason-code vocabulary to a golden file.
//
// These strings are written into the journal, into alerts and (Phase 2) into the
// ledger. Renaming one silently orphans records that already exist on disk, so a
// change here has to be a deliberate edit of the fixture, not a side effect.
func TestReasonCodeEnumIsStable(t *testing.T) {
	golden := filepath.Join("testdata", "reason_codes.golden")
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("reading %s: %v", golden, err)
	}

	codes := execgw.AllReasonCodes()
	lines := make([]string, 0, len(codes))
	for _, c := range codes {
		lines = append(lines, string(c))
	}
	got := strings.Join(lines, "\n") + "\n"

	if got != string(want) {
		t.Errorf("the reason-code enum changed.\n--- want (%s)\n%s\n--- got\n%s",
			golden, want, got)
	}

	// Uniqueness: two codes with the same string would make a journal record
	// ambiguous about which branch produced it.
	seen := map[execgw.ReasonCode]bool{}
	for _, c := range codes {
		if seen[c] {
			t.Errorf("duplicate reason code %q", c)
		}
		seen[c] = true
	}
}

// --- preflight --------------------------------------------------------------

func newPreflight(account execgw.AccountReader) *execgw.Preflight {
	return &execgw.Preflight{Account: account}
}

// TestPreflightRejectsUnsupportedOrderTypes: the shapes the official path cannot
// express are refused before anything is sent, with a code that says which one.
func TestPreflightRejectsUnsupportedOrderTypes(t *testing.T) {
	cases := []struct {
		name   string
		intent orderintent.PlaceIntent
		want   execgw.ReasonCode
	}{
		{
			name: "unknown market",
			intent: orderintent.PlaceIntent{
				Symbol: "AAPL", Market: "jp", Side: "buy", OrderType: "limit",
				Quantity: 1, Price: 10, CurrencyMode: "USD",
			},
			want: execgw.ReasonUnsupportedOrderType,
		},
		{
			name: "non-fractional market order",
			intent: orderintent.PlaceIntent{
				Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "market",
				Quantity: 1, CurrencyMode: "USD",
			},
			want: execgw.ReasonUnsupportedOrderType,
		},
		{
			name: "KR order priced in USD",
			intent: orderintent.PlaceIntent{
				Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
				Quantity: 1, Price: 70000, CurrencyMode: "USD",
			},
			want: execgw.ReasonUnsupportedOrderType,
		},
		{
			name: "limit order without a price",
			intent: orderintent.PlaceIntent{
				Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
				Quantity: 1, CurrencyMode: "KRW",
			},
			want: execgw.ReasonInvalidRequest,
		},
		{
			name: "non-positive quantity",
			intent: orderintent.PlaceIntent{
				Symbol: "005930", Market: "kr", Side: "buy", OrderType: "limit",
				Quantity: 0, Price: 70000, CurrencyMode: "KRW",
			},
			want: execgw.ReasonInvalidRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rejected := newPreflight(&fakeAccount{buyingPower: 1e9}).CheckPlace(context.Background(), tc.intent)
			if rejected == nil {
				t.Fatalf("want a refusal with %q, got none", tc.want)
			}
			if rejected.Reason != tc.want {
				t.Errorf("reason: got %q, want %q (%s)", rejected.Reason, tc.want, rejected.Detail)
			}
		})
	}
}

// TestPreflightRejectsKRWFundedUSBuy: a US buy priced in KRW can only settle by
// converting currency, and the spec forbids the engine triggering that
// automatically. Selling is unaffected — it produces currency, it does not spend it.
func TestPreflightRejectsKRWFundedUSBuy(t *testing.T) {
	buy := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 150000, CurrencyMode: "KRW",
	}
	rejected := newPreflight(&fakeAccount{buyingPower: 1e9}).CheckPlace(context.Background(), buy)
	if rejected == nil || rejected.Reason != execgw.ReasonFXConsentRequired {
		t.Fatalf("want fx_consent_required, got %v", rejected)
	}

	sell := buy
	sell.Side = "sell"
	if rejected := newPreflight(&fakeAccount{buyingPower: 1e9}).CheckPlace(context.Background(), sell); rejected != nil {
		t.Errorf("a US sell must not be blocked by the FX rule: %v", rejected)
	}
}

// TestPreflightRejectsInsufficientBalance drives the check off a recorded broker
// response fixture rather than a hand-made number.
func TestPreflightRejectsInsufficientBalance(t *testing.T) {
	fixture, err := os.ReadFile(filepath.Join("testdata", "insufficient_usd_balance.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	client := official.New(official.Credentials{APIKey: "k", SecretKey: "s"},
		filepath.Join(t.TempDir(), "token.json"),
		official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()),
		official.WithAccountSeq(7))
	pf := newPreflight(execgw.OfficialAccount{Client: client})

	// 1 share at $150 against $12.50 of buying power.
	tooBig := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 150, CurrencyMode: "USD",
	}
	rejected := pf.CheckPlace(context.Background(), tooBig)
	if rejected == nil || rejected.Reason != execgw.ReasonBalanceInsufficient {
		t.Fatalf("want balance_insufficient, got %v", rejected)
	}
	if !strings.Contains(rejected.Detail, "12.5") {
		t.Errorf("the refusal must name what was available: %s", rejected.Detail)
	}

	// A fractional buy for less than the available cash passes.
	affordable := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "market",
		Amount: 10, CurrencyMode: "USD", Fractional: true,
	}
	if rejected := pf.CheckPlace(context.Background(), affordable); rejected != nil {
		t.Errorf("an affordable order must pass: %v", rejected)
	}
}

// TestPreflightFailsClosedWhenTheBalanceIsUnreadable: not knowing the balance is
// not permission to trade.
func TestPreflightFailsClosedWhenTheBalanceIsUnreadable(t *testing.T) {
	account := &fakeAccount{err: official.ErrServer}
	intent := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 150, CurrencyMode: "USD",
	}
	rejected := newPreflight(account).CheckPlace(context.Background(), intent)
	if rejected == nil || rejected.Reason != execgw.ReasonBalanceUnavailable {
		t.Fatalf("want balance_unavailable, got %v", rejected)
	}

	// With no account reader at all the gateway cannot check, and a buy is
	// refused rather than assumed funded.
	if rejected := (&execgw.Preflight{}).CheckPlace(context.Background(), intent); rejected == nil ||
		rejected.Reason != execgw.ReasonBalanceUnavailable {
		t.Errorf("want balance_unavailable without an account reader, got %v", rejected)
	}
}

// TestPreflightLetsSellsThrough: an exit does not need buying power, and gating it
// on one would close the exits (§0.3).
func TestPreflightLetsSellsThrough(t *testing.T) {
	sell := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "sell", OrderType: "limit",
		Quantity: 1, Price: 150, CurrencyMode: "USD",
	}
	if rejected := (&execgw.Preflight{}).CheckPlace(context.Background(), sell); rejected != nil {
		t.Errorf("a sell must not need a balance check: %v", rejected)
	}
}

// --- broker branch classification -------------------------------------------

// TestBrokerBranchesMapToStableReasonCodes: the broker asking a human to do
// something is a refusal with a name, never something to answer automatically.
func TestBrokerBranchesMapToStableReasonCodes(t *testing.T) {
	fxBody, err := os.ReadFile(filepath.Join("testdata", "fx_consent_required.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	authBody, err := os.ReadFile(filepath.Join("testdata", "interactive_auth_challenge.json"))
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	cases := []struct {
		name string
		err  error
		want execgw.ReasonCode
	}{
		{
			name: "interactive trade auth",
			err:  trading.ErrInteractiveAuthRequired,
			want: execgw.ReasonInteractiveAuthRequired,
		},
		{
			name: "interactive challenge in the body",
			err:  &official.APIError{Code: 403, Body: string(authBody)},
			want: execgw.ReasonInteractiveAuthRequired,
		},
		{
			name: "fx consent branch",
			err: &trading.BranchRequiredError{
				Branch: trading.BranchFXConsentRequired,
				Source: trading.BranchSourcePrepareRejection,
				FX:     &trading.FXConfirmationContext{NeedExchangeUSD: 180},
			},
			want: execgw.ReasonFXConsentRequired,
		},
		{
			name: "fx consent in the body",
			err:  &official.APIError{Code: 400, Body: string(fxBody)},
			want: execgw.ReasonFXConsentRequired,
		},
		{
			name: "funding branch",
			err: &trading.BranchRequiredError{
				Branch: trading.BranchFundingRequired,
				Source: trading.BranchSourcePrepareRejection,
			},
			want: execgw.ReasonFundingRequired,
		},
		{
			name: "plain auth rejection",
			err:  official.ErrAuth,
			want: execgw.ReasonBrokerAuthRejected,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := execgw.ClassifyBrokerRefusal(tc.err)
			if !ok {
				t.Fatalf("want %q, got no classification", tc.want)
			}
			if got != tc.want {
				t.Errorf("reason: got %q, want %q", got, tc.want)
			}
		})
	}

	if _, ok := execgw.ClassifyBrokerRefusal(errors.New("something else")); ok {
		t.Error("an unrelated error must not be classified as an operator branch")
	}
	if _, ok := execgw.ClassifyBrokerRefusal(nil); ok {
		t.Error("nil must not be classified")
	}
}

// --- gateway integration ----------------------------------------------------

// TestGatewayRefusesFailClosedBranchesBeforeDispatch: the refusal is journalled
// with its reason code and the broker is never called.
func TestGatewayRefusesFailClosedBranchesBeforeDispatch(t *testing.T) {
	broker := &fakeBroker{result: acceptedResult("O-1")}
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
		Preflight: &execgw.Preflight{Account: &fakeAccount{buyingPower: 5}},
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	intent := orderintent.PlaceIntent{
		Symbol: "AAPL", Market: "us", Side: "buy", OrderType: "limit",
		Quantity: 1, Price: 150, CurrencyMode: "USD",
	}
	out, err := gw.Place(context.Background(), execgw.PlaceRequest{
		Intent: intent,
		Decision: entryDecision(t, j, clk, intent, execgw.Limits{
			MaxQuantity:        execgw.Bound(10),
			MaxNotional:        execgw.Bound(100_000),
			MaxTotalExposure:   execgw.Bound(500_000),
			MaxDailyLossAmount: execgw.Bound(20_000),
			MaxDailyLossRatio:  execgw.Bound(0.02),
			Currency:           "USD",
		}),
	})

	var rejected *execgw.RejectedError
	if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonBalanceInsufficient {
		t.Fatalf("want balance_insufficient, got %v", err)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Errorf("broker place calls: got %d, want 0", places)
	}
	if out.State != journal.StateNotDispatched {
		t.Errorf("state: got %s, want NOT_DISPATCHED", out.State)
	}

	rec, err := j.LookupAttempt(context.Background(), out.AttemptID)
	if err != nil {
		t.Fatalf("LookupAttempt: %v", err)
	}
	if rec.ReasonCode != string(execgw.ReasonBalanceInsufficient) {
		t.Errorf("journalled reason code: got %q, want %q",
			rec.ReasonCode, execgw.ReasonBalanceInsufficient)
	}
}

// TestGatewayNeverAutoApprovesABranch: when the broker answers a submission with
// "a human must approve this", the attempt is closed with the branch's reason code
// and nothing is retried or confirmed.
func TestGatewayNeverAutoApprovesABranch(t *testing.T) {
	broker := &fakeBroker{err: trading.ErrInteractiveAuthRequired}
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test",
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}

	out, _ := gw.Place(context.Background(), placeRequest(t, j, clk))
	if out.Reason != execgw.ReasonInteractiveAuthRequired {
		t.Errorf("reason: got %q, want %q (%s)", out.Reason, execgw.ReasonInteractiveAuthRequired, out.Detail)
	}
	if out.State != journal.StateFailedConfirmed {
		t.Errorf("state: got %s, want FAILED_CONFIRMED — a challenge is a definitive refusal", out.State)
	}
	if places, _, _ := broker.totals(); places != 1 {
		t.Errorf("broker place calls: got %d, want exactly 1 (no retry, no auto-approval)", places)
	}
}
