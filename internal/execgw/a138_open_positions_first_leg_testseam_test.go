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

// a138: 생산 전략 첫 다리 문도 최대 보유 포지션 수를 싣고, 넘으면 LIMIT_REACHED(첫 다리 fixture 가 testseam 태그 아래라 이 파일도 같음).
func TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum(t *testing.T) {
	for _, descriptor := range strategyflow.Descriptors() {
		if descriptor.LaneID != weeklyvaluelane.KRWeeklyLaneID {
			continue
		}
		decisionID := "a138-first-leg"
		rig := newGuardian(t, func(o *execgw.RiskGuardianOptions) {
			o.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			maxOpenPositions(1)(o)
		})
		a138HoldElsewhere(t, rig.journal)
		precheck, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(qFinalCampaignFirstLegIssuanceFixture(t, rig, descriptor, decisionID))
		if err != nil {
			t.Fatal(err)
		}
		_, err = rig.guardian.IssuePrecheckedQFinalCampaignFirstLeg(context.Background(), precheck)
		if !errors.Is(err, journal.ErrReservationLimitExceeded) || !strings.Contains(err.Error(), "MAX_OPEN_POSITIONS") {
			t.Fatalf("want the open position maximum refusal, got %v", err)
		}
		if decisionOnDisk(t, rig.journal, decisionID) {
			t.Fatal("a refused first leg must leave no decision behind")
		}
		return
	}
	t.Fatal("no KR weekly descriptor — the fixture this test drives is gone")
}
