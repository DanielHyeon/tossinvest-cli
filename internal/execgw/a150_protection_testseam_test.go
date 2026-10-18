//go:build tossos_testseams

package execgw_test

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
)

func TestA150GuardianRefusesMissingProtectionSnapshot(t *testing.T) {
	rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
		options.Policy.RiskBudget = riskcalc.Money{Amount: "100000", Currency: "KRW"}
	})
	request := qFinalCampaignFirstLegIssuanceFixture(t, rig, strategyflow.Descriptors()[0], "a150-missing")
	request.ExitPolicy = nil
	if _, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request); err == nil {
		t.Fatal("missing exit policy accepted")
	}
}
