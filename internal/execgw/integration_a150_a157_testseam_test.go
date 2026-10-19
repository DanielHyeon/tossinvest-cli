//go:build tossos_testseams

package execgw_test

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
)

func TestIntegrationA150AndA157FirstLegRefusalPrecedence(t *testing.T) {
	t.Run("nil guardian and policy", func(t *testing.T) {
		var guardian *execgw.RiskGuardian
		_, err := guardian.PrecheckQFinalCampaignFirstLeg(execgw.QFinalCampaignFirstLegIssuance{})
		a157Expect(t, "nil guardian", err, "q:RISK_CALCULATION_INVALID/guardian: Guardian is nil")
	})
	t.Run("missing policy before daily loss and binding", func(t *testing.T) {
		binding := a157Bindings()[0]
		rig, request, _ := a157Precheck(t, a157KRWeekly(t), a157Limit(), &binding)
		request.ExitPolicy = nil
		_, err := rig.guardian.PrecheckQFinalCampaignFirstLeg(request)
		a157Expect(t, "missing policy", err, "q:RISK_CALCULATION_INVALID/first_leg_exit_policy: admitted exit policy snapshot is required")
	})
	t.Run("valid policy retains daily loss precedence", func(t *testing.T) {
		binding := a157Bindings()[0]
		_, request, err := a157Precheck(t, a157KRWeekly(t), a157Limit(), &binding)
		if request.ExitPolicy == nil {
			t.Fatal("fixture lacks exit policy")
		}
		a157Expect(t, "daily loss and binding", err, a157DailyLimit)
	})
}
