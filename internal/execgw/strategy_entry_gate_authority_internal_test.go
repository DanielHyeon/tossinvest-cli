package execgw

import (
	"context"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestStrategyEntryGateAuthorityRejectsAllowedBlockedAllowedABAKRUS(t *testing.T) {
	for _, market := range []string{"kr", "us"} {
		t.Run(market, func(t *testing.T) {
			gate := NewEntryGate(clock.NewFake(time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)), map[RequiredQuery]time.Duration{})
			gateway := &Gateway{entry: gate, accountRef: "acct-aba"}
			first, err := gateway.ObserveStrategyEntryGate(context.Background(), market, "AAPL")
			if err != nil {
				t.Fatal(err)
			}
			gate.BlockSymbol(market, "AAPL", ReasonReconcileMismatch, "changed")
			if _, err := gateway.ObserveStrategyEntryGate(context.Background(), market, "AAPL"); err == nil {
				t.Fatal("blocked gate issued strategy authority")
			}
			gate.ClearSymbol(market, "AAPL", ReasonReconcileMismatch)
			second, err := gateway.ObserveStrategyEntryGate(context.Background(), market, "AAPL")
			if err != nil {
				t.Fatal(err)
			}
			if second.Generation() <= first.Generation() || second.Digest() == first.Digest() {
				t.Fatalf("ABA reused authority first=%d/%s second=%d/%s", first.Generation(), first.Digest(), second.Generation(), second.Digest())
			}
			calls := 0
			if err := gateway.withStrategyEntryGateAuthority(context.Background(), first, market, "AAPL", func() error {
				calls++
				return nil
			}); err == nil || calls != 0 {
				t.Fatalf("stale ABA authority crossed final interlock calls=%d err=%v", calls, err)
			}
			if err := gateway.withStrategyEntryGateAuthority(context.Background(), second, market, "AAPL", func() error {
				calls++
				return nil
			}); err != nil || calls != 1 {
				t.Fatalf("current authority final interlock calls=%d err=%v", calls, err)
			}
			gate.ClearSymbol(market, "AAPL", ReasonReconcileMismatch)
			third, err := gateway.ObserveStrategyEntryGate(context.Background(), market, "AAPL")
			if err != nil || third != second {
				t.Fatalf("no-op clear changed authority second=%+v third=%+v err=%v", second, third, err)
			}
		})
	}
}
