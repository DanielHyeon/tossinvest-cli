package exitpolicy

import (
	"fmt"
	"strings"
)

const (
	CommonLadderBalanced = "COMMON_LADDER_BALANCED"
	CommonLadderRunner   = "COMMON_LADDER_RUNNER"
	CommonLadderHybrid50 = "COMMON_LADDER_HYBRID_50"
)

// CommonPolicy is one server-authoritative profile exposed by Optimization.
type CommonPolicy struct {
	ID          string
	Label       string
	Recommended bool
	Ladder      LadderPolicy
}

var commonPolicies = []CommonPolicy{
	{
		ID: CommonLadderBalanced, Label: "균형형",
		Ladder: LadderPolicy{
			PolicyID: CommonLadderBalanced,
			Rungs: []Rung{
				{TargetPct: "1.5", StopPct: "0", PartialRatio: "0"},
				{TargetPct: "2.5", StopPct: "1.0", PartialRatio: "0.25"},
				{TargetPct: "4.0", StopPct: "2.0", PartialRatio: "0.25"},
				{TargetPct: "6.0", StopPct: "3.5", PartialRatio: "1.0"},
			},
			FinalTakeFull: true,
		},
	},
	{
		ID: CommonLadderRunner, Label: "러너형",
		Ladder: LadderPolicy{
			PolicyID: CommonLadderRunner,
			Rungs: []Rung{
				{TargetPct: "2.5", StopPct: "0", PartialRatio: "0"},
				{TargetPct: "4.5", StopPct: "2.0", PartialRatio: "0.15"},
				{TargetPct: "7.0", StopPct: "3.5", PartialRatio: "0"},
				{TargetPct: "999.0", StopPct: "5.0", PartialRatio: "0"},
			},
		},
	},
	{
		ID: CommonLadderHybrid50, Label: "하이브리드 50", Recommended: true,
		Ladder: LadderPolicy{
			PolicyID: CommonLadderHybrid50,
			Rungs: []Rung{
				{TargetPct: "1.8", StopPct: "0", PartialRatio: "0"},
				{TargetPct: "3.0", StopPct: "1.2", PartialRatio: "0.25"},
				{TargetPct: "4.8", StopPct: "2.5", PartialRatio: "1/3"},
				{TargetPct: "6.5", StopPct: "3.8", PartialRatio: "0"},
			},
			RunnerTrailPct: "6.5",
		},
	},
}

// RegisteredCommonPolicies returns deep copies so callers cannot mutate the
// process-wide registry.
func RegisteredCommonPolicies() []CommonPolicy {
	out := make([]CommonPolicy, 0, len(commonPolicies))
	for _, policy := range commonPolicies {
		out = append(out, cloneCommonPolicy(policy))
	}
	return out
}

func CommonPolicyByID(id string) (CommonPolicy, bool) {
	id = strings.TrimSpace(id)
	for _, policy := range commonPolicies {
		if policy.ID == id {
			return cloneCommonPolicy(policy), true
		}
	}
	return CommonPolicy{}, false
}

// CommonLadderForPosition resolves the immutable profile. StockOS A168 keeps
// adopted RUNNER positions floor-only, so their copy has no partial ratios.
func CommonLadderForPosition(id string, adopted bool) (LadderPolicy, error) {
	policy, ok := CommonPolicyByID(id)
	if !ok {
		return LadderPolicy{}, fmt.Errorf("exitpolicy: unknown common policy %q", strings.TrimSpace(id))
	}
	if adopted && policy.ID == CommonLadderRunner {
		for i := range policy.Ladder.Rungs {
			policy.Ladder.Rungs[i].PartialRatio = "0"
		}
	}
	return policy.Ladder, nil
}

func cloneCommonPolicy(policy CommonPolicy) CommonPolicy {
	policy.Ladder.Rungs = append([]Rung(nil), policy.Ladder.Rungs...)
	return policy
}
