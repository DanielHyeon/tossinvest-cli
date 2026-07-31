package exitpolicy

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
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
	Summary     string
	Recommended bool
	Ladder      LadderPolicy
}

var commonPolicies = buildCommonPolicies()

func buildCommonPolicies() []CommonPolicy {
	policies := []CommonPolicy{
		{
			ID: CommonLadderBalanced, Label: "균형형", Summary: "단계별 익절과 보호 승격을 균형 있게 적용합니다.",
			Ladder: LadderPolicy{
				PolicyID: CommonLadderBalanced, PolicyVersion: DefaultPolicyVersion,
				PolicyDigest: balancedPolicyDigest,
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
			ID: CommonLadderRunner, Label: "러너형", Summary: "초기 익절을 줄이고 상승 여력을 오래 보호합니다.",
			Ladder: LadderPolicy{
				PolicyID: CommonLadderRunner, PolicyVersion: DefaultPolicyVersion,
				PolicyDigest: runnerPolicyDigest,
				Rungs: []Rung{
					{TargetPct: "2.5", StopPct: "0", PartialRatio: "0"},
					{TargetPct: "4.5", StopPct: "2.0", PartialRatio: "0.15"},
					{TargetPct: "7.0", StopPct: "3.5", PartialRatio: "0"},
					{TargetPct: "999.0", StopPct: "5.0", PartialRatio: "0"},
				},
			},
		},
		{
			ID: CommonLadderHybrid50, Label: "하이브리드 50", Summary: "절반가량을 단계적으로 확보하고 나머지를 고점 추적으로 보호합니다.", Recommended: true,
			Ladder: LadderPolicy{
				PolicyID: CommonLadderHybrid50, PolicyVersion: DefaultPolicyVersion,
				PolicyDigest: hybrid50PolicyDigest,
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
	for index := range policies {
		identity, err := policies[index].Ladder.Identity()
		if err != nil {
			panic(err)
		}
		if identity.Version != policies[index].Ladder.PolicyVersion ||
			identity.Digest != policies[index].Ladder.PolicyDigest {
			panic(fmt.Sprintf("exitpolicy: pinned identity drift for %s", policies[index].ID))
		}
	}
	return policies
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
		policy.Ladder.PolicyVersion = adoptedRunnerPolicyVersion
		policy.Ladder.PolicyDigest = adoptedRunnerPolicyDigest
		identity, err := policy.Ladder.Identity()
		if err != nil {
			return LadderPolicy{}, err
		}
		if identity.Digest != adoptedRunnerPolicyDigest {
			return LadderPolicy{}, fmt.Errorf("%w: adopted RUNNER pinned identity drift",
				ErrPolicyIdentityConflict)
		}
	}
	return policy.Ladder, nil
}

type PolicyUnitDescriptor struct {
	Target  string
	Stop    string
	Partial string
	Runner  string
}

type RungDescriptor struct {
	ID           string
	TargetPct    string
	StopPct      string
	PartialRatio string
}

type OneShareDescriptor struct {
	Intermediate     string
	Final            string
	ProtectionBreach string
}

type CommonPolicyDescriptor struct {
	Identity       PolicyIdentity
	Label          string
	Summary        string
	Recommended    bool
	Rungs          []RungDescriptor
	RunnerTrailPct string
	FinalTakeFull  bool
	Unit           PolicyUnitDescriptor
	OneShare       OneShareDescriptor
}

func RegisteredCommonPolicyDescriptors() []CommonPolicyDescriptor {
	policies := RegisteredCommonPolicies()
	out := make([]CommonPolicyDescriptor, 0, len(policies))
	for _, policy := range policies {
		identity, err := policy.Ladder.Identity()
		if err != nil {
			panic(err)
		}
		descriptor := CommonPolicyDescriptor{
			Identity: identity, Label: policy.Label, Summary: policy.Summary,
			Recommended: policy.Recommended, RunnerTrailPct: policy.Ladder.RunnerTrailPct,
			FinalTakeFull: policy.Ladder.FinalTakeFull,
			Unit:          PolicyUnitDescriptor{Target: "%", Stop: "%", Partial: "ratio-of-remaining", Runner: "%"},
			OneShare: OneShareDescriptor{
				Intermediate: "매도 0주 · 보호선 승격", Final: "1주 전량", ProtectionBreach: "1주 전량",
			},
		}
		for index, rung := range policy.Ladder.Rungs {
			descriptor.Rungs = append(descriptor.Rungs, RungDescriptor{
				ID: fmt.Sprintf("T%d", index+1), TargetPct: rung.TargetPct,
				StopPct: rung.StopPct, PartialRatio: rung.PartialRatio,
			})
		}
		out = append(out, descriptor)
	}
	return out
}

func CommonPolicyFieldDescriptor() settingmeta.FieldDescriptor {
	policies := RegisteredCommonPolicyDescriptors()
	options := make([]settingmeta.Option, 0, len(policies))
	for _, policy := range policies {
		options = append(options, settingmeta.Option{
			ID: policy.Identity.ID, Label: policy.Label, Description: policy.Summary,
			Value: policy.Identity.ID, Recommended: policy.Recommended,
		})
	}
	return settingmeta.FieldDescriptor{
		Key: "exit.common-policy", Label: "공통 익절·보호 정책",
		Description: "새로 열리는 포지션에 적용할 immutable exit 정책입니다.",
		Type:        settingmeta.TypeEnum, Control: settingmeta.ControlRadioTile, Options: options,
		Default:     settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "기존 RATCHET 유지"},
		Effective:   settingmeta.State{Kind: settingmeta.StateUnapproved, Display: "기존 RATCHET 유지"},
		ApplyTiming: settingmeta.ApplyNewPositionOnly, SafetyDirection: settingmeta.SafetyContextual,
		Provenance: settingmeta.Provenance{OwnerChange: "a041-complete-exit-line-contract"},
	}
}

func cloneCommonPolicy(policy CommonPolicy) CommonPolicy {
	policy.Ladder.Rungs = append([]Rung(nil), policy.Ladder.Rungs...)
	return policy
}
