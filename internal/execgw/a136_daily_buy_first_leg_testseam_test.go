//go:build tossos_testseams

package execgw_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
)

// a136: 생산 전략 첫 다리 문도 하루 매수 상한을 싣고, 도달하면 LIMIT_REACHED(첫 다리 fixture 가 testseam 태그 아래라 이 파일도 같음).
func TestA136QFinalCampaignFirstLegRefusesTheBuyThatReachesTheDailyCeiling(t *testing.T) {
	for _, descriptor := range strategyflow.Descriptors() {
		if descriptor.LaneID != weeklyvaluelane.KRWeeklyLaneID {
			continue
		}
		decisionID := "a136-first-leg"
		rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
			o.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			smallDailyBuy("1")(o)
		})
		precheck, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, decisionID))
		if err != nil {
			t.Fatal(err)
		}
		_, err = rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), precheck)
		// 이 문은 기존부터 예약 거부를 IssueRefusal 로 분류하지 않고 감싸서 돌려줌(노출 한도와 같음) — 원인만 대조.
		if !errors.Is(err, journal.ErrReservationLimitExceeded) || !strings.Contains(err.Error(), "DAILY_BUY_NOTIONAL") {
			t.Fatalf("want the daily buy ceiling refusal, got %v", err)
		}
		if decisionOnDisk(t, rig.journal, decisionID) {
			t.Fatal("a refused first leg must leave no decision behind")
		}
		return
	}
	t.Fatal("no KR weekly descriptor — the fixture this test drives is gone")
}
