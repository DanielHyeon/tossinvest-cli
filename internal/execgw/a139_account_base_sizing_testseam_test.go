//go:build tossos_testseams

package execgw_test

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// a139 구현 리뷰 P2-1: account-base(FX) 경로의 1/3 사이징을 직접 잼 — 짝 계약 시험은 예산을 옛 값으로 고정했으므로 여기서 기본 정책
// (예산 33,333 KRW)의 q_final 을 박음. US: 33,333 / (손절 폭 5 USD × 1,400) = 4.76 → 4. KR: 다른 상한(10)이 묶음.
func TestA139TheAccountBasePathSizesWithAThirdOfTheDailyLoss(t *testing.T) {
	for _, c := range []struct {
		market, rate string
		costMarket   costs.Market
		want         uint64
	}{
		{"KR", "1", costs.MarketKR, 10},
		{"US", "1400", costs.MarketUS, 4},
	} {
		t.Run(c.market, func(t *testing.T) {
			decisionID := "a139-account-base-" + strings.ToLower(c.market)
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			})
			request := qFinalAccountBaseRequest(t, rig, strings.ToLower(c.market), c.market)
			fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, c.costMarket, guardianPolicy(), c.rate, "1")
			if err != nil {
				t.Fatal(err)
			}
			bindPairedAccountBaseFXForTest(&request, fx)
			precheck, err := rig.guardian.PrecheckQFinalEntry(request)
			if err != nil {
				t.Fatalf("%s precheck: %v", c.market, err)
			}
			if precheck.QCandidate() != 20 || precheck.QFinal() != c.want {
				t.Fatalf("%s candidate=%d final=%d, want 20/%d", c.market, precheck.QCandidate(), precheck.QFinal(), c.want)
			}
		})
	}
}
