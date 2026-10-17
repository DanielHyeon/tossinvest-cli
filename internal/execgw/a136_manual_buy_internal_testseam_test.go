//go:build tossos_testseams

package execgw

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// a136 D9=b · FX 조항: 수동 매수의 base 환산 — base 통화는 그대로, 다른 통화는 그 시장의 봉인 authority 로만 올림, 없으면 거부.

var a136Now = time.Date(2026, 10, 14, 2, 0, 0, 0, time.UTC)

func a136Guardian(board *ManualBuyBoard) *RiskGuardian {
	return &RiskGuardian{policy: risk.DefaultPolicy(), manualBuys: board}
}

func a136Board(byCurrency map[string]string, err error) *ManualBuyBoard {
	b := NewManualBuyBoard()
	b.Publish(ManualBuySnapshot{AsOf: a136Now, Day: "2026-10-14", ByCurrency: byCurrency, Err: err})
	return b
}

func TestA136ManualBuysConvertOnlyThroughTheMarketsSealedAuthority(t *testing.T) {
	usd, err := officialfx.EvidenceForAuthorityTest("USD", "KRW", "1400", "1.01", a136Now.Add(-time.Minute), a136Now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	g := a136Guardian(a136Board(map[string]string{"KRW": "1000", "USD": "100"}, nil))
	total, err := g.manualBuyTotal(a136Now, manualBuyFX{costs.MarketUS: usd})
	if err != nil {
		t.Fatal(err)
	}
	// 1000 + ceil(100 × 1400 × 1.01 = 141400) = 142400
	if total.Amount != "142400" || total.Day != "2026-10-14" || !total.AsOf.Equal(a136Now) {
		t.Fatalf("total = %+v, want 142400 KRW for 2026-10-14", total)
	}

	if _, err := g.manualBuyTotal(a136Now, nil); !errors.Is(err, ErrManualBuyUnknown) {
		t.Fatalf("USD manual buys without the US authority must refuse the buy, got %v", err)
	}
	stale := a136Now.Add(10 * time.Minute)
	if _, err := g.manualBuyTotal(stale, manualBuyFX{costs.MarketUS: usd}); !errors.Is(err, ErrManualBuyUnknown) {
		t.Fatalf("an expired US authority must refuse the buy, got %v", err)
	}
}

func TestA136ABaseOnlyOrZeroForeignManualBuyNeedsNoAuthority(t *testing.T) {
	g := a136Guardian(a136Board(map[string]string{"KRW": "1000", "USD": "0"}, nil))
	total, err := g.manualBuyTotal(a136Now, nil)
	if err != nil || total.Amount != "1000" {
		t.Fatalf("total = %+v err = %v, want 1000 with no authority", total, err)
	}
}

func TestA136AnUnknownManualBuySnapshotRefusesTheBuy(t *testing.T) {
	for name, g := range map[string]*RiskGuardian{
		"no board":         a136Guardian(nil),
		"never published":  a136Guardian(NewManualBuyBoard()),
		"failed read":      a136Guardian(a136Board(nil, errors.New("orders read failed"))),
		"unknown currency": a136Guardian(a136Board(map[string]string{"JPY": "5"}, nil)),
		"negative amount":  a136Guardian(a136Board(map[string]string{"KRW": "-1"}, nil)),
	} {
		if _, err := g.manualBuyTotal(a136Now, nil); !errors.Is(err, ErrManualBuyUnknown) {
			t.Errorf("%s: want ErrManualBuyUnknown, got %v", name, err)
		}
	}
}

func TestA136QFinalCarriesTheOtherMarketOfThePairOnly(t *testing.T) {
	kr, err := officialfx.EvidenceForAuthorityTest("KRW", "KRW", "1", "1", a136Now.Add(-time.Minute), a136Now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	us, err := officialfx.EvidenceForAuthorityTest("USD", "KRW", "1400", "1.01", a136Now.Add(-time.Minute), a136Now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	fx := qFinalManualBuyFX(QFinalEntryIssuance{FXAuthority: kr, PairedFXAuthority: us}, costs.MarketKR)
	if fx[costs.MarketKR].Digest() != kr.Digest() || fx[costs.MarketUS].Digest() != us.Digest() || len(fx) != 2 {
		t.Fatalf("a KR request must carry its own and the paired US authority, got %v", fx)
	}
	if fx := qFinalManualBuyFX(QFinalEntryIssuance{FXAuthority: us}, costs.MarketUS); len(fx) != 1 || fx[costs.MarketUS].Digest() != us.Digest() {
		t.Fatalf("without a paired authority only the request's own market is available, got %v", fx)
	}
}
