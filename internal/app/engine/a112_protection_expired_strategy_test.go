//go:build tossos_testseams

package engine

// a112 태스크 6.6 (h) — 엔진 수준 만료 1(Manager 판정 2026-10-01 A66). 엔진 시험의 Gateway 스파이는 보호 관측을 일반 오류(`failProtection`)로만
// 주입해 없음 · 불일치 · 만료를 가르지 못한다. 여기서는 dispatch 주기의 보호 관측을 **실제 게이트웨이**(`execgw.Gateway.ObserveStrategyProtection`
// — 생산 어댑터 `protection.NewPairedReadinessAdapter` 위)로 돌리고, 주문 송신(`PlaceClaimedStrategy`)만 스파이로 센다. 증명은 태그 seam
// `protectionreadiness.ReadinessSnapshotForTest` 로 봉인한다(생산 스냅숏은 서명 공급자만 만든다). 세 모양의 게이트웨이 층 구분은
// `execgw` `TestEachProtectionAttestationFailureStopsBuysAndKeepsReductionsFlowing`.
//
// 대조군(양성): 같은 주기 · 같은 계약에 신선한 증명이면 첫 레그가 송신된다 — 만료 거절이 증명 때문임을 보인다.

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

type a112ReadinessProvider struct {
	snapshot protectionreadiness.ReadinessSnapshot
}

func (provider a112ReadinessProvider) Current(context.Context) (protectionreadiness.ReadinessSnapshot, error) {
	return provider.snapshot, nil
}

// a112RealProtectionGateway 는 보호 관측만 실제 게이트웨이에 맡기고 나머지(진입 관문 · 송신)는 스파이 그대로 두는 dispatch Gateway 다.
type a112RealProtectionGateway struct {
	*strategyDispatchGatewaySpy
	real *execgw.Gateway
}

func (gateway a112RealProtectionGateway) ObserveStrategyProtection(ctx context.Context, market string, quantity uint64) (execgw.StrategyProtectionAuthority, error) {
	return gateway.real.ObserveStrategyProtection(ctx, market, quantity)
}

func a112AttestationAt(now time.Time, expiresAt time.Time) protectionreadiness.ReadinessSnapshot {
	digest := func(value string) string { return strings.Repeat(value, 64) }
	verdict := func(market protectionreadiness.Market) protectionreadiness.Verdict {
		return protectionreadiness.Verdict{Market: market, State: protectionreadiness.Wired, Provenance: protectionreadiness.Provenance{
			AccountID: "acct-risk-loader", ProfileID: "production", OrderType: "LIMIT", SessionScope: "regular", QuantityMin: 1, QuantityMax: 1_000_000,
			TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b"),
			KeyID: "key", Serial: 9, BodyDigest: digest("c"), BuildDigest: digest("d"), EvidenceDigest: digest("e"), SupervisorDigest: digest("f"),
			IssuedAt: now.Add(-time.Hour), ExpiresAt: expiresAt}}
	}
	return protectionreadiness.ReadinessSnapshotForTest(verdict(protectionreadiness.MarketKR), verdict(protectionreadiness.MarketUS))
}

func TestAnExpiredProtectionAttestationStopsTheStrategyFirstLegBeforeTheBroker(t *testing.T) {
	for _, tc := range []struct {
		name      string
		expiresAt func(now time.Time) time.Time
		placed    int
	}{
		{"control: fresh attestation", func(now time.Time) time.Time { return now.Add(time.Hour) }, 1},
		{"expired exactly at the wave", func(now time.Time) time.Time { return now }, 0},
		{"expired a minute before the wave", func(now time.Time) time.Time { return now.Add(-time.Minute) }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newA112TradingFixture(t, a112TradingOptions{})
			digest := func(value string) string { return strings.Repeat(value, 64) }
			contracts := []protectionreadiness.RuntimeContract{
				{Market: protectionreadiness.MarketKR, SessionScope: "regular", TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b")},
				{Market: protectionreadiness.MarketUS, SessionScope: "regular", TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b")},
			}
			adapter, err := protection.NewPairedReadinessAdapter(a112ReadinessProvider{a112AttestationAt(fixture.now, tc.expiresAt(fixture.now))},
				"acct-risk-loader", "production", contracts)
			if err != nil {
				t.Fatal(err)
			}
			real, err := execgw.New(execgw.Options{Journal: fixture.journal, Trading: trading.NewService(config.Trading{}, nil),
				Clock: clock.NewFake(fixture.now), AccountRef: "acct-risk-loader", Source: "a112-6.6", ProtectionReadiness: adapter})
			if err != nil {
				t.Fatal(err)
			}
			fixture.cycle.gateway = a112RealProtectionGateway{strategyDispatchGatewaySpy: fixture.spy, real: real}
			err = fixture.deliverKR(t)
			if got := len(fixture.placedSymbols()); got != tc.placed {
				t.Fatalf("strategy first legs sent=%d err=%v, want %d", got, err, tc.placed)
			}
			if tc.placed == 0 && (err == nil || !strings.Contains(err.Error(), "attestation_expired")) {
				t.Fatalf("err=%v, want the cycle refused by the expired attestation (attestation_expired)", err)
			}
		})
	}
}
