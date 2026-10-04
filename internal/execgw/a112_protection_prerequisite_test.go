//go:build tossos_testseams

package execgw_test

// a112 태스크 6.6 (h) · (i)(Manager 판정 2026-10-01 A66): a100 보호 증명이 **없음 · 불일치 · 만료** 인 세 모양 각각에서, 같은 시험 안에서
// (1) 노출을 올리는 매수는 브로커 요청 0 이고 (2) 축소 전용 경로(매도 · 취소 · 축소 정정)는 브로커에 닿는다. 앞 판의 짝은 갈라져 있었다 —
// 미배선(UNWIRED) 거절과 매도 수락이 서로 다른 시험(`protection_test.go:63` · `:94`)이었고, 만료는 판정 층(`protectionreadiness`)에서만 쟀다.
//
// 대조군(양성): 같은 어댑터 · 같은 계약에 **신선한 WIRED** 증명이면 같은 매수가 브로커에 닿는다 — 세 모양의 거절이 증명 때문이지 픽스처가
// 원래 매수를 못 내서가 아님을 보인다. 증명은 `protectionreadiness.ReadinessSnapshotForTest`(태그 seam — 생산 스냅숏은 서명 공급자만 만든다)로
// 봉인하고 그 뒤는 생산 어댑터 `protection.NewPairedReadinessAdapter` 와 게이트웨이 `Place`/`Cancel`/`Amend` 그대로다.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
)

func a112Verdict(market protectionreadiness.Market, mutate func(*protectionreadiness.Provenance)) protectionreadiness.Verdict {
	digest := func(value string) string { return strings.Repeat(value, 64) }
	p := protectionreadiness.Provenance{AccountID: "acct-7", ProfileID: "production", OrderType: "LIMIT",
		SessionScope: "regular", QuantityMin: 1, QuantityMax: 100, TriggerSource: "broker-native",
		ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b"),
		KeyID: "key", Serial: 7, BodyDigest: digest("c"), BuildDigest: digest("d"), EvidenceDigest: digest("e"),
		SupervisorDigest: digest("f"), IssuedAt: fixedNow.Add(-time.Minute), ExpiresAt: fixedNow.Add(time.Minute)}
	if mutate != nil {
		mutate(&p)
	}
	return protectionreadiness.Verdict{Market: market, State: protectionreadiness.Wired, Provenance: p}
}

func a112AttestedGateway(t *testing.T, snapshot protectionreadiness.ReadinessSnapshot) (*execgw.Gateway, *journal.Journal, *clock.Fake, *fakeBroker, *boundaryProvider) {
	t.Helper()
	digest := func(value string) string { return strings.Repeat(value, 64) }
	provider := &boundaryProvider{snapshot: snapshot}
	contracts := []protectionreadiness.RuntimeContract{
		{Market: protectionreadiness.MarketKR, SessionScope: "regular", TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b")},
		{Market: protectionreadiness.MarketUS, SessionScope: "regular", TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b")},
	}
	adapter, err := protection.NewPairedReadinessAdapter(provider, "acct-7", "production", contracts)
	if err != nil {
		t.Fatal(err)
	}
	broker := &fakeBroker{result: domain.MutationResult{Status: "accepted", OrderID: "O-1", CurrentOrderID: "O-1"}}
	gw, j, clk := newGatewayWithReadiness(t, broker, adapter)
	return gw, j, clk, broker, provider
}

func TestEachProtectionAttestationFailureStopsBuysAndKeepsReductionsFlowing(t *testing.T) {
	for _, tc := range []struct {
		name     string
		snapshot protectionreadiness.ReadinessSnapshot
		buys     int    // 매수의 브로커 요청 수
		cause    string // 거절 사유 코드(모양이 원인임을 가름)
	}{
		{"control: fresh WIRED attestation", protectionreadiness.ReadinessSnapshotForTest(a112Verdict(protectionreadiness.MarketKR, nil), a112Verdict(protectionreadiness.MarketUS, nil)), 1, ""},
		{"missing: default (UNWIRED) snapshot", protectionreadiness.DefaultSnapshot(), 0, "missing_evidence"},
		{"mismatched: attestation for another account", protectionreadiness.ReadinessSnapshotForTest(
			a112Verdict(protectionreadiness.MarketKR, func(p *protectionreadiness.Provenance) { p.AccountID = "acct-other" }), a112Verdict(protectionreadiness.MarketUS, nil)), 0, "attestation_scope_mismatch"},
		{"mismatched: attestation for another tool digest", protectionreadiness.ReadinessSnapshotForTest(
			a112Verdict(protectionreadiness.MarketKR, func(p *protectionreadiness.Provenance) { p.ToolDigest = strings.Repeat("9", 64) }), a112Verdict(protectionreadiness.MarketUS, nil)), 0, "attestation_scope_mismatch"},
		{"expired: attestation expired exactly now", protectionreadiness.ReadinessSnapshotForTest(
			a112Verdict(protectionreadiness.MarketKR, func(p *protectionreadiness.Provenance) { p.ExpiresAt = fixedNow }), a112Verdict(protectionreadiness.MarketUS, nil)), 0, "attestation_expired"},
		{"expired: attestation expired a minute ago", protectionreadiness.ReadinessSnapshotForTest(
			a112Verdict(protectionreadiness.MarketKR, func(p *protectionreadiness.Provenance) {
				p.IssuedAt, p.ExpiresAt = fixedNow.Add(-2*time.Minute), fixedNow.Add(-time.Minute)
			}), a112Verdict(protectionreadiness.MarketUS, nil)), 0, "attestation_expired"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gw, j, clk, broker, provider := a112AttestedGateway(t, tc.snapshot)
			// (1) 노출 상승: KR 매수.
			buy := placeIntent()
			out, err := gw.Place(context.Background(), execgw.PlaceRequest{Intent: buy, Decision: entryDecision(t, j, clk, buy, testLimits())})
			places, _, _ := broker.totals()
			if places != tc.buys {
				t.Fatalf("buy reached the broker %d times (out=%+v err=%v), want %d", places, out, err, tc.buys)
			}
			if tc.buys == 0 && (provider.calls == 0 || out.State == journal.StateConfirmed || !strings.Contains(out.Detail, ": "+tc.cause+")")) {
				t.Fatalf("refused buy: provider calls=%d state=%s detail=%q — the refusal must come from this attestation shape (%s), read before the broker", provider.calls, out.State, out.Detail, tc.cause)
			}
			// (2) 축소 전용: 매도 · 취소 · 축소 정정은 같은 게이트웨이에서 브로커에 닿는다.
			sell := orderintent.PlaceIntent{Symbol: "005930", Market: "kr", Side: "sell", OrderType: "limit", Quantity: 1, Price: 70000, CurrencyMode: "KRW"}
			if out, err := gw.Place(context.Background(), execgw.PlaceRequest{Intent: sell,
				Decision: exitDecision(t, j, clk, journal.KindPlace, sell.Market, sell.Symbol, sell.Side, sell.Quantity)}); err != nil || out.State != journal.StateConfirmed {
				t.Fatalf("reduce-only sell blocked: out=%+v err=%v", out, err)
			}
			if out, err := gw.Cancel(context.Background(), execgw.CancelRequest{Intent: orderintent.CancelIntent{OrderID: "O-9", Symbol: "005930"},
				Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
				Decision: exitDecision(t, j, clk, journal.KindCancel, "kr", "005930", "BUY", 2)}); err != nil || out.State != journal.StateConfirmed {
				t.Fatalf("reduce-only cancel blocked: out=%+v err=%v", out, err)
			}
			quantity := 1.0
			if out, err := gw.Amend(context.Background(), execgw.AmendRequest{Intent: orderintent.AmendIntent{OrderID: "O-8", Quantity: &quantity}, Symbol: "005930",
				Order:    execgw.OrderRef{Market: "kr", Side: "BUY", Quantity: 2, Price: 70000, Currency: "KRW"},
				Decision: exitDecision(t, j, clk, journal.KindAmend, "kr", "005930", "BUY", 2)}); err != nil || out.State != journal.StateConfirmed {
				t.Fatalf("reduce-only amend blocked: out=%+v err=%v", out, err)
			}
			places, cancels, amends := broker.totals()
			if places != tc.buys+1 || cancels != 1 || amends != 1 {
				t.Fatalf("broker calls place=%d cancel=%d amend=%d, want buys %d + sell 1, cancel 1, amend 1", places, cancels, amends, tc.buys)
			}
		})
	}
}
