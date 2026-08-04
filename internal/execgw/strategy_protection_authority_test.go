//go:build tossos_testseams

package execgw_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/protectionreadiness"
)

func TestObserveStrategyProtectionPairsKRUSWithoutBroker(t *testing.T) {
	digest := func(value string) string { return strings.Repeat(value, 64) }
	provenance := func(market protectionreadiness.Market, serial uint64) protectionreadiness.Verdict {
		return protectionreadiness.Verdict{Market: market, State: protectionreadiness.Wired,
			Provenance: protectionreadiness.Provenance{AccountID: "acct-7", ProfileID: "production", OrderType: "LIMIT",
				SessionScope: "regular", QuantityMin: 1, QuantityMax: 100, TriggerSource: "broker-native",
				ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b"),
				KeyID: "key", Serial: serial, BodyDigest: digest("c"), BuildDigest: digest("d"), EvidenceDigest: digest("e"),
				SupervisorDigest: digest("f"), IssuedAt: fixedNow.Add(-time.Minute), ExpiresAt: fixedNow.Add(time.Minute)}}
	}
	provider := &boundaryProvider{snapshot: protectionreadiness.ReadinessSnapshotForTest(
		provenance(protectionreadiness.MarketKR, 11), provenance(protectionreadiness.MarketUS, 22))}
	contracts := []protectionreadiness.RuntimeContract{
		{Market: protectionreadiness.MarketKR, SessionScope: "regular", TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b")},
		{Market: protectionreadiness.MarketUS, SessionScope: "regular", TriggerSource: "broker-native", ReplaceSemantics: "CONTINUOUS_COVERAGE", BrokerCapabilityDigest: digest("a"), ToolDigest: digest("b")},
	}
	adapter, err := protection.NewPairedReadinessAdapter(provider, "acct-7", "production", contracts)
	if err != nil {
		t.Fatal(err)
	}
	broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "unexpected"}}
	gw, _, _ := newGatewayWithReadiness(t, broker, adapter)
	for _, tc := range []struct {
		market string
		serial uint64
	}{{"kr", 11}, {"us", 22}} {
		authority, err := gw.ObserveStrategyProtection(context.Background(), tc.market, 3)
		if err != nil || authority.Market() != strings.ToUpper(tc.market) || authority.Generation() != tc.serial ||
			!strings.HasPrefix(authority.Digest(), "sha256:") {
			t.Fatalf("%s authority=%+v err=%v", tc.market, authority, err)
		}
	}
	if places, _, _ := broker.totals(); places != 0 {
		t.Fatalf("protection observation reached broker calls=%d", places)
	}
}

func TestObserveStrategyProtectionFailsClosedForDefaultKRUS(t *testing.T) {
	provider := &boundaryProvider{snapshot: protectionreadiness.DefaultSnapshot()}
	adapter, err := protection.NewReadinessAdapter(provider, "acct-7", "production")
	if err != nil {
		t.Fatal(err)
	}
	gw, _, _ := newGatewayWithReadiness(t, &fakeBroker{}, adapter)
	for _, market := range []string{"kr", "us"} {
		if _, err := gw.ObserveStrategyProtection(context.Background(), market, 1); err == nil {
			t.Fatalf("%s default protection observed as authority", market)
		}
	}
}

var _ = execgw.StrategyProtectionAuthority{}
