package journal

import (
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// StrategyExitPolicySnapshot freezes the policy identity at admission.
type StrategyExitPolicySnapshot struct {
	Kind     string                    `json:"kind"`
	PolicyID string                    `json:"policy_id"`
	Identity exitpolicy.PolicyIdentity `json:"identity"`
}

// NewStrategyExitPolicySnapshot pins the same policy semantics used by exit seeds.
func NewStrategyExitPolicySnapshot(commonPolicy string) (StrategyExitPolicySnapshot, error) {
	if commonPolicy != strings.TrimSpace(commonPolicy) {
		return StrategyExitPolicySnapshot{}, fmt.Errorf("strategy protection: noncanonical policy id")
	}
	kind := ExitPolicyRatchet
	if commonPolicy != "" {
		kind = ExitPolicyLadder
	}
	identity, err := seedPolicyIdentity(kind, commonPolicy, false, exitpolicy.PolicyIdentity{})
	if err != nil {
		return StrategyExitPolicySnapshot{}, err
	}
	return StrategyExitPolicySnapshot{Kind: kind, PolicyID: commonPolicy, Identity: identity}, nil
}

func validateStrategyExitPolicySnapshot(snapshot StrategyExitPolicySnapshot) error {
	if snapshot.Kind != ExitPolicyRatchet && snapshot.Kind != ExitPolicyLadder {
		return fmt.Errorf("strategy protection: unsupported policy kind")
	}
	if (snapshot.Kind == ExitPolicyRatchet && snapshot.PolicyID != "") || (snapshot.Kind == ExitPolicyLadder && snapshot.PolicyID == "") || snapshot.PolicyID != strings.TrimSpace(snapshot.PolicyID) {
		return fmt.Errorf("strategy protection: invalid policy id")
	}
	expected, err := seedPolicyIdentity(snapshot.Kind, snapshot.PolicyID, false, snapshot.Identity)
	if err != nil {
		return err
	}
	if snapshot.Identity != expected {
		return fmt.Errorf("strategy protection: missing or noncanonical policy identity")
	}
	return nil
}
