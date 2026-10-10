package main

// order_test.go 는 하드 가드와 confirm token 을 잼 — 네트워크 없음.

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

const testKey = "a087ms-20261012T101500-deadbeef"

func goodSpec() orderSpec {
	return orderSpec{Market: "KR", Side: "sell", OrderType: "market", Symbol: "005930", Quantity: 1, ClientOrderID: testKey}
}

// TestAssembleRefusesEachGuard 는 가드마다 그 가드의 문구로 거절되는지 봄 — 다른 가드가 대신 막는
// 경우를 통과로 세지 않으려고 문구를 가드별로 단언함.
func TestAssembleRefusesEachGuard(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		mutate func(*orderSpec)
		want   string
	}{
		{"buy", func(s *orderSpec) { s.Side = "buy" }, "sell orders only"},
		{"uppercase SELL is not the fixed value", func(s *orderSpec) { s.Side = "SELL" }, "sell orders only"},
		{"limit", func(s *orderSpec) { s.OrderType = "limit" }, "market orders only"},
		{"US market with a KR-shaped symbol", func(s *orderSpec) { s.Market = "US" }, "KR orders only"},
		{"US ticker", func(s *orderSpec) { s.Symbol = "AAPL" }, "six digits"},
		{"five digits", func(s *orderSpec) { s.Symbol = "05930" }, "six digits"},
		{"seven digits", func(s *orderSpec) { s.Symbol = "0059300" }, "six digits"},
		{"quantity over the cap", func(s *orderSpec) { s.Quantity = 3 }, "1..2 shares"},
		{"zero quantity", func(s *orderSpec) { s.Quantity = 0 }, "1..2 shares"},
		{"negative quantity", func(s *orderSpec) { s.Quantity = -1 }, "1..2 shares"},
		{"foreign idempotency key", func(s *orderSpec) { s.ClientOrderID = "my-order-001" }, "not a key this probe issued"},
		{"no idempotency key", func(s *orderSpec) { s.ClientOrderID = "" }, "not a key this probe issued"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := goodSpec()
			tc.mutate(&spec)
			order, err := assembleOrder(spec)
			if err == nil {
				t.Fatalf("assembleOrder(%+v) = %q, want a refusal", spec, order.WireBody)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("refusal %q does not name its own guard (%q)", err, tc.want)
			}
			if order.WireBody != "" {
				t.Fatalf("a refused spec still produced a body %q", order.WireBody)
			}
		})
	}
}

// TestAssembleProducesTheMarketSellBody 는 허용되는 두 모양(1주·2주)의 바이트를 원문으로 고정함.
func TestAssembleProducesTheMarketSellBody(t *testing.T) {
	t.Parallel()
	for qty, want := range map[int]string{
		1: `{"symbol":"005930","side":"SELL","orderType":"MARKET","quantity":"1","clientOrderId":"` + testKey + `","confirmHighValueOrder":false}`,
		2: `{"symbol":"005930","side":"SELL","orderType":"MARKET","quantity":"2","clientOrderId":"` + testKey + `","confirmHighValueOrder":false}`,
	} {
		spec := goodSpec()
		spec.Quantity = qty
		order, err := assembleOrder(spec)
		if err != nil {
			t.Fatalf("qty %d: %v", qty, err)
		}
		if order.WireBody != want {
			t.Fatalf("qty %d body\n got %s\nwant %s", qty, order.WireBody, want)
		}
		if !order.IssuedAt.Equal(time.Date(2026, 10, 12, 10, 15, 0, 0, kst)) {
			t.Fatalf("issued at %s, want the time inside the key", order.IssuedAt)
		}
	}
}

func TestMintedKeyFitsTheContractAndRoundTrips(t *testing.T) {
	t.Parallel()
	id, err := mintClientOrderID(fixedNow, bytes.NewReader([]byte{1, 2, 3, 4}))
	if err != nil {
		t.Fatal(err)
	}
	if id != "a087ms-20261012T101500-01020304" || len(id) > 36 {
		t.Fatalf("minted %q (len %d)", id, len(id))
	}
	issued, err := parseClientOrderID(id)
	if err != nil || !issued.Equal(fixedNow) {
		t.Fatalf("parse(%q) = %s, %v", id, issued, err)
	}
	if _, err := mintClientOrderID(fixedNow, bytes.NewReader(nil)); err == nil {
		t.Fatal("an exhausted random source must not mint a key")
	}
}

// TestConfirmTokenBindsTheWholeBody 는 칸 하나만 바꿔도 토큰이 무효가 되는지 봄.
func TestConfirmTokenBindsTheWholeBody(t *testing.T) {
	t.Parallel()
	base, err := assembleOrder(goodSpec())
	if err != nil {
		t.Fatal(err)
	}
	token := confirmToken(base)
	if !strings.HasPrefix(token, testKey+".") || len(token) != len(testKey)+1+16 {
		t.Fatalf("token %q has the wrong shape", token)
	}
	if err := verifyConfirm(base, token, fixedNow); err != nil {
		t.Fatalf("the matching token was refused: %v", err)
	}
	variants := map[string]func(*orderSpec){
		"qty":    func(s *orderSpec) { s.Quantity = 2 },
		"symbol": func(s *orderSpec) { s.Symbol = "000660" },
		"key":    func(s *orderSpec) { s.ClientOrderID = "a087ms-20261012T101500-deadbeee" },
	}
	for name, mutate := range variants {
		spec := goodSpec()
		mutate(&spec)
		other, err := assembleOrder(spec)
		if err != nil {
			t.Fatal(err)
		}
		if err := verifyConfirm(other, token, fixedNow); err == nil || !strings.Contains(err.Error(), "does not match") {
			t.Fatalf("%s changed but the old token was accepted (err=%v)", name, err)
		}
	}
	tampered := token[:len(token)-1] + "0"
	if tampered == token {
		tampered = token[:len(token)-1] + "1"
	}
	if err := verifyConfirm(base, tampered, fixedNow); err == nil {
		t.Fatal("a tampered digest was accepted")
	}
}

func TestConfirmWindow(t *testing.T) {
	t.Parallel()
	order, err := assembleOrder(goodSpec())
	if err != nil {
		t.Fatal(err)
	}
	token := confirmToken(order)
	if err := verifyConfirm(order, token, fixedNow.Add(confirmWindow)); err != nil {
		t.Fatalf("the window edge was refused: %v", err)
	}
	if err := verifyConfirm(order, token, fixedNow.Add(confirmWindow+time.Second)); err == nil ||
		!strings.Contains(err.Error(), "the window is") {
		t.Fatalf("an expired token was accepted (err=%v)", err)
	}
	if err := verifyConfirm(order, token, fixedNow.Add(-futureSkew-time.Second)); err == nil ||
		!strings.Contains(err.Error(), "in the future") {
		t.Fatalf("a key from the future was accepted (err=%v)", err)
	}
	if confirmWindow >= 10*time.Minute {
		t.Fatalf("confirmWindow %s must stay inside the broker's 10-minute idempotency window", confirmWindow)
	}
}

func TestSplitConfirmToken(t *testing.T) {
	t.Parallel()
	for _, bad := range []string{"", "abc", testKey, testKey + ".", testKey + ".0123", ".0123456789abcdef"} {
		if _, err := splitConfirmToken(bad); err == nil {
			t.Errorf("splitConfirmToken(%q) accepted", bad)
		}
	}
}

// TestCLIOffersNoSideTypeOrMarketFlag 는 고정값을 바꿀 플래그 자체가 없음을 봄.
func TestCLIOffersNoSideTypeOrMarketFlag(t *testing.T) {
	t.Parallel()
	for _, flagName := range []string{"--side", "--type", "--order-type", "--market", "--price"} {
		if _, err := parseOptions([]string{"--symbol", "005930", "--qty", "1", flagName, "x"}); err == nil {
			t.Errorf("%s was accepted as a flag", flagName)
		}
	}
}
