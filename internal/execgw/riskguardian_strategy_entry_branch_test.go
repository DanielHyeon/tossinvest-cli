package execgw_test

import (
	"context"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"
)

// The dormant build has no authentic frozen-source manifest and deliberately
// exposes no test-only Decision minter. Only the invalid opaque-value branch is
// therefore reachable here without weakening the production authority boundary.
func TestRiskGuardianIssueStrategyEntryRejectsInvalidOpaqueDecisionBeforeCollection(t *testing.T) {
	rig := newGuardian(t, nil)
	issued, err := rig.guardian.IssueStrategyEntry(context.Background(), execgw.StrategyEntryIssuance{
		Decision:              strategyengine.Decision{},
		Collect:               rig.collect,
		ExpectedPolicyVersion: rig.guardian.PolicyVersion(),
		ExpectedLimitsDigest:  rig.guardian.LimitsDigest(),
	})
	if err == nil || !strings.Contains(err.Error(), "strategy decision is invalid") ||
		!issued.Decision.IsZero() || rig.collections != 0 {
		t.Fatalf("issued=%+v collections=%d err=%v", issued, rig.collections, err)
	}
}

func TestRiskGuardianIssueEntryRejectsMissingCollectorBeforeIssuingAuthority(t *testing.T) {
	rig := newGuardian(t, nil)
	issued, err := rig.guardian.IssueEntry(context.Background(), execgw.EntryIssuance{
		Intent:  guardianIntent(),
		Account: guardianAccount(),
	})
	if err == nil || !strings.Contains(err.Error(), "needs a snapshot collector") || !issued.Decision.IsZero() {
		t.Fatalf("issued=%+v err=%v", issued, err)
	}
}
