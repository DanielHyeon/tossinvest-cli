package engine

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// a139 구현 리뷰 P1-1: 생산 Guardian **객체**가 1/3 예산으로 사이징하는지 행동으로 잼 — 정책 유도 헬퍼만 재면 생성자에서 예산을 되돌리는 편집이
// 스위트를 통과하고 audit 는 여전히 1/3 을 적음. 일일 손실 200,000 → 예산 66,666, 손절 폭 1,000 이면 위험 상한 66 주.
func TestA139TheProductionGuardianSizesWithAThirdOfTheDailyLoss(t *testing.T) {
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	gate := config.AutomationGate{Enabled: true, MaxOrderQuantity: 100, MaxOrderNotional: 1_000_000,
		MaxTotalExposure: 5_000_000, MaxDailyLossAmount: 200_000, MaxDailyLossRatio: 0.02, LimitCurrency: "KRW"}
	guardian, err := newProductionRiskGuardian(gate, j, clock.NewFake(time.Date(2026, 3, 30, 1, 30, 0, 0, time.UTC)), "acct-1", nil)
	if err != nil {
		t.Fatalf("newProductionRiskGuardian: %v", err)
	}
	account := risk.AccountState{
		Mode: risk.ModeNormal, AllowedSymbols: []string{"005930"},
		CashAvailable:     riskcalc.Money{Amount: "5000000", Currency: "KRW"},
		OpenExposure:      riskcalc.Money{Amount: "0", Currency: "KRW"},
		DailyRealizedLoss: riskcalc.Money{Amount: "0", Currency: "KRW"},
		AccountEquity:     riskcalc.Money{Amount: "100000000", Currency: "KRW"},
	}
	collected := errors.New("a139: the chain passed and the reservation collector was reached")
	issue := func(quantity string) error {
		_, err := guardian.IssueEntry(context.Background(), execgw.EntryIssuance{
			Intent: risk.Intent{Market: costs.MarketKR, Symbol: "005930", Side: risk.SideBuy, Quantity: quantity,
				LimitPrice: "10000", StopPrice: "9000", TargetPrice: "12500"},
			Account: account,
			Collect: func(context.Context, int) (execgw.ExposureSnapshot, error) {
				return execgw.ExposureSnapshot{}, collected
			},
		})
		return err
	}
	var refusal *execgw.IssueRefusal
	if err := issue("67"); !errors.As(err, &refusal) || refusal.Reason != string(risk.ReasonInvalidOrderSize) ||
		!strings.Contains(refusal.Detail, "66 units") || !strings.Contains(refusal.Detail, "66666") {
		t.Fatalf("67 shares over a 1,000 stop must be refused at the 66-unit cap of a 66,666 budget, got %v", err)
	}
	if err := issue("66"); !errors.Is(err, collected) {
		t.Fatalf("66 shares must clear the chain's risk cap and reach the reservation, got %v", err)
	}
}
