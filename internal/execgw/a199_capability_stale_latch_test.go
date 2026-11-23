package execgw_test

// a199_capability_stale_latch_test.go — a199 D4: 낡은 capability 증명의 진입 차단은 기존 집행 지점 하나(게이트웨이의
// 진입 검사 → EntryGate 래치)에 새 사유 하나로 건다. 여기서 고정하는 것:
//   - 사유가 닫힌 열거(AllReasonCodes)에 있고 철자가 계약값이다(원장 · 알림에 쓰이는 이름).
//   - 보고 순서: 기존 쌍의 상대 순서를 깨지 않고, 운영 모드보다 앞(원인이 결과보다 먼저) · 전략 중앙 무결성보다 뒤(붙임).
//   - 보호가 WIRED 인 게이트웨이에서 매수는 이 사유로 거절되고 브로커 호출 0, 매도는 통과한다(청산은 막지 않음).
//   - 보호가 UNWIRED 인 게이트웨이(지금 생산)에서는 매수 거절 사유가 지금과 같다(보호 검사가 진입 검사보다 먼저).

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestA199TheStaleAttestationReasonIsRegistered(t *testing.T) {
	if execgw.ReasonCapabilityAttestationStale != "capability_attestation_stale" {
		t.Fatalf("spelling = %q — the code is written into ledger rows and alerts", execgw.ReasonCapabilityAttestationStale)
	}
	found := 0
	for _, c := range execgw.AllReasonCodes() {
		if c == execgw.ReasonCapabilityAttestationStale {
			found++
		}
	}
	if found != 1 {
		t.Fatalf("AllReasonCodes() lists the stale-attestation reason %d time(s), want 1", found)
	}
}

func TestA199TheStaleAttestationLatchReadsBeforeTheModeAndAfterStrategyCentral(t *testing.T) {
	gate := func() *execgw.EntryGate {
		return execgw.NewEntryGate(clock.NewFake(time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)),
			map[execgw.RequiredQuery]time.Duration{})
	}
	g := gate()
	g.Block(execgw.ReasonOperatingModeBlocked, "ENTRY_BLOCKED")
	g.Block(execgw.ReasonCapabilityAttestationStale, "stale")
	if r := g.CheckEntry(); r == nil || r.Reason != execgw.ReasonCapabilityAttestationStale {
		t.Errorf("with the mode also latched CheckEntry = %v, want the stale-attestation cause first", r)
	}
	g = gate()
	g.Block(execgw.ReasonCapabilityAttestationStale, "stale")
	g.Block(execgw.ReasonStrategyCentralIntegrity, "strategy central")
	if r := g.CheckEntry(); r == nil || r.Reason != execgw.ReasonStrategyCentralIntegrity {
		t.Errorf("with strategy central also latched CheckEntry = %v, want strategy central first (appended after it)", r)
	}
}

// a199Gateway 는 진입 게이트를 단 게이트웨이. wired=false 면 이 빌드의 실제 보호 판정(UNWIRED)을 씀.
func a199Gateway(t *testing.T, broker trading.Broker, gate *execgw.EntryGate, wired bool) (*execgw.Gateway, *journal.Journal, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(fixedNow)
	j := openJournal(t, clk)
	opts := execgw.Options{
		Journal: j, Trading: trading.NewService(openPolicy(), broker), Clock: clk,
		AccountRef: "acct-7", Source: "test", Entry: gate,
	}
	if !wired {
		opts.UseReadinessAdapterForTest()
	}
	gw, err := execgw.New(opts)
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	return gw, j, clk
}

func a199StaleGate() *execgw.EntryGate {
	g := execgw.NewEntryGate(clock.NewFake(fixedNow), map[execgw.RequiredQuery]time.Duration{})
	g.Block(execgw.ReasonCapabilityAttestationStale, "capability attestation is stale")
	return g
}

func TestA199AWiredGatewayRefusesABuyForTheStaleAttestationAndAdmitsASell(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := a199Gateway(t, broker, a199StaleGate(), true)

	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err == nil {
		t.Fatalf("a buy passed a gateway whose entry gate carries the stale-attestation latch: %+v", out)
	}
	if out.Reason != execgw.ReasonCapabilityAttestationStale {
		t.Fatalf("reason = %q (%s), want %q", out.Reason, out.Detail, execgw.ReasonCapabilityAttestationStale)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Fatalf("broker place calls = %d, want 0 — the refusal must precede the broker", places)
	}

	sell := orderintent.PlaceIntent{
		Symbol: "000660", Market: "kr", Side: "sell", OrderType: "limit",
		Quantity: 2, Price: 70000, CurrencyMode: "KRW",
	}
	out, err = gw.Place(context.Background(), execgw.PlaceRequest{
		Intent:   sell,
		Decision: exitDecision(t, j, clk, journal.KindPlace, sell.Market, sell.Symbol, sell.Side, sell.Quantity),
	})
	if err != nil || out.State != journal.StateConfirmed {
		t.Fatalf("a reduce-only sell must pass the stale-attestation latch: state=%s err=%v (%s)", out.State, err, out.Detail)
	}
}

func TestA199AnUnwiredGatewayStillRefusesABuyForProtectionFirst(t *testing.T) {
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "O-1"}}
	gw, j, clk := a199Gateway(t, broker, a199StaleGate(), false)

	out, err := gw.Place(context.Background(), placeRequest(t, j, clk))
	if err == nil {
		t.Fatalf("a buy passed an unwired gateway: %+v", out)
	}
	if out.Reason != execgw.ReasonProtectionNotWired {
		t.Fatalf("reason = %q, want %q — the stale latch must not change the reason today's build reports",
			out.Reason, execgw.ReasonProtectionNotWired)
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Fatalf("broker place calls = %d, want 0", places)
	}
}
