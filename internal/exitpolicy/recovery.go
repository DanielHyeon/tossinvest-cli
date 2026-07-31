package exitpolicy

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	ErrRecoveryAmbiguous = errors.New("exitpolicy: recovery candidates cross safety axes")
	ErrRecoveryIdentity  = errors.New("exitpolicy: recovery candidate identity mismatch")
)

type RecoverySource string

const (
	RecoveryRecomputed    RecoverySource = "recomputed"
	RecoverySavedMonotone RecoverySource = "saved_monotone"
)

// RecoveryPolicyDefinition is the exact immutable policy meaning and the
// minimal evaluation input needed to re-derive the next line after a restart.
// Exactly one arm must be present. It is evidence only and grants no execution
// authority.
type RecoveryPolicyDefinition struct {
	Ratchet *RatchetRecoveryDefinition `json:"ratchet,omitempty"`
	Ladder  *LadderPolicy              `json:"ladder,omitempty"`
}

type RatchetRecoveryDefinition struct {
	Config          RatchetConfig `json:"config"`
	RealBreakeven   string        `json:"real_breakeven"`
	TakenRatioTotal string        `json:"taken_ratio_total"`
}

func NewRatchetRecoveryPolicy(config RatchetConfig, realBreakeven, takenRatioTotal string) RecoveryPolicyDefinition {
	return RecoveryPolicyDefinition{Ratchet: &RatchetRecoveryDefinition{
		Config: config, RealBreakeven: strings.TrimSpace(realBreakeven),
		TakenRatioTotal: strings.TrimSpace(takenRatioTotal),
	}}
}

func NewLadderRecoveryPolicy(policy LadderPolicy) RecoveryPolicyDefinition {
	copy := policy
	copy.Rungs = append([]Rung(nil), policy.Rungs...)
	return RecoveryPolicyDefinition{Ladder: &copy}
}

// ValidateRecoveryDerivation resolves the selected immutable policy definition
// by its exact identity and re-derives NextTarget/NextProtection. A digest is
// not accepted as a substitute for this semantic check.
func ValidateRecoveryDerivation(line ExitLineSnapshot, definition RecoveryPolicyDefinition) error {
	if (definition.Ratchet == nil) == (definition.Ladder == nil) {
		return fmt.Errorf("%w: recovery needs exactly one policy definition", ErrRecoveryIdentity)
	}
	var target, protection string
	if ratchet := definition.Ratchet; ratchet != nil {
		identity, err := RatchetPolicyIdentity(ratchet.Config)
		if err != nil || !sameRecoveryPolicy(identity, line.Policy) || line.ActiveRung != NoRung {
			return fmt.Errorf("%w: ratchet definition does not resolve stored policy identity", ErrRecoveryIdentity)
		}
		entry, _, risk, err := riskOf(line.EntryPrice, line.InitialStop)
		if err != nil {
			return err
		}
		high, err := positive("recovery high water", line.HighWater)
		if err != nil {
			return err
		}
		if _, err := positive("recovery real breakeven", ratchet.RealBreakeven); err != nil {
			return err
		}
		rMultiple := formatRMultiple(new(big.Rat).Quo(new(big.Rat).Sub(high, entry), risk))
		target, protection, err = nextRatchetLine(RatchetInput{
			Entry: line.EntryPrice, InitialStop: line.InitialStop,
			RealBreakeven: ratchet.RealBreakeven, TakenRatioTotal: ratchet.TakenRatioTotal,
		}, ratchet.Config, RatchetDecision{RMultiple: rMultiple, Baseline: line.CurrentProtection})
		if err != nil {
			return err
		}
	} else {
		identity, err := definition.Ladder.Identity()
		if err != nil || !sameRecoveryPolicy(identity, line.Policy) || line.ActiveRung == NoRung {
			return fmt.Errorf("%w: ladder definition does not resolve stored policy identity", ErrRecoveryIdentity)
		}
		target, protection, err = nextLadderLine(line.EntryPrice, line.ActiveRung, *definition.Ladder)
		if err != nil {
			return err
		}
	}
	if target != line.NextTarget || protection != line.NextProtection {
		return fmt.Errorf("%w: derived next line is %q/%q, stored is %q/%q",
			ErrRecoveryIdentity, target, protection, line.NextTarget, line.NextProtection)
	}
	return nil
}

// SelectRecoverySnapshot chooses one complete candidate. It never constructs a
// tuple by taking maxima from different candidates.
func SelectRecoverySnapshot(saved *ExitLineSnapshot, recomputed ExitLineSnapshot) (ExitLineSnapshot, RecoverySource, error) {
	if err := validateRecoverySnapshot(recomputed); err != nil {
		return ExitLineSnapshot{}, "", err
	}
	if saved == nil {
		return recomputed, RecoveryRecomputed, nil
	}
	if err := validateRecoverySnapshot(*saved); err != nil {
		return ExitLineSnapshot{}, "", err
	}
	if saved.PositionID != recomputed.PositionID || saved.PositionGeneration != recomputed.PositionGeneration ||
		!sameRecoveryPolicy(saved.Policy, recomputed.Policy) || saved.EntryPrice != recomputed.EntryPrice ||
		saved.InitialStop != recomputed.InitialStop || strings.TrimSpace(saved.InputDigest) == "" ||
		strings.TrimSpace(recomputed.InputDigest) == "" {
		return ExitLineSnapshot{}, "", ErrRecoveryIdentity
	}
	protection, err := compareRecoveryDecimal(recomputed.CurrentProtection, saved.CurrentProtection)
	if err != nil {
		return ExitLineSnapshot{}, "", err
	}
	high, err := compareRecoveryDecimal(recomputed.HighWater, saved.HighWater)
	if err != nil {
		return ExitLineSnapshot{}, "", err
	}
	stage, err := compareRecoveryStage(recomputed, *saved)
	if err != nil {
		return ExitLineSnapshot{}, "", err
	}
	if stage == 0 && (saved.NextTarget != recomputed.NextTarget || saved.NextProtection != recomputed.NextProtection) {
		return ExitLineSnapshot{}, "", fmt.Errorf("%w: equal policy stage has different derived next lines", ErrRecoveryIdentity)
	}
	if protection >= 0 && high >= 0 && stage >= 0 {
		return recomputed, RecoveryRecomputed, nil
	}
	if protection <= 0 && high <= 0 && stage <= 0 {
		return *saved, RecoverySavedMonotone, nil
	}
	return ExitLineSnapshot{}, "", ErrRecoveryAmbiguous
}

func validateRecoverySnapshot(s ExitLineSnapshot) error {
	if strings.TrimSpace(s.SnapshotID) == "" || strings.TrimSpace(s.DecisionID) == "" ||
		strings.TrimSpace(s.ObservationID) == "" || strings.TrimSpace(s.PositionID) == "" || s.PositionGeneration < 0 {
		return fmt.Errorf("%w: incomplete snapshot identity", ErrRecoveryIdentity)
	}
	if err := s.Policy.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrRecoveryIdentity, err)
	}
	if !validSHA256Identity(s.InputDigest) {
		return fmt.Errorf("%w: invalid input digest", ErrRecoveryIdentity)
	}
	expected := s
	expected.finishIDs()
	if expected.SnapshotID != s.SnapshotID || expected.DecisionID != s.DecisionID {
		return fmt.Errorf("%w: snapshot/decision id does not bind the stored output", ErrRecoveryIdentity)
	}
	entry, err := positive("recovery entry", s.EntryPrice)
	if err != nil {
		return err
	}
	stop, err := positive("recovery initial stop", s.InitialStop)
	if err != nil {
		return err
	}
	if stop.Cmp(entry) >= 0 {
		return fmt.Errorf("%w: initial stop is not below entry", ErrRecoveryIdentity)
	}
	if _, err := positive("recovery observed price", s.ObservedPrice); err != nil {
		return err
	}
	if _, err := positive("recovery protection", s.CurrentProtection); err != nil {
		return err
	}
	if _, err := positive("recovery high water", s.HighWater); err != nil {
		return err
	}
	for _, next := range []struct{ name, value string }{{"next target", s.NextTarget}, {"next protection", s.NextProtection}} {
		if strings.TrimSpace(next.value) != "" {
			if _, err := positive(next.name, next.value); err != nil {
				return err
			}
		}
	}
	if s.ActiveRung < NoRung {
		return fmt.Errorf("%w: invalid active rung %d", ErrRecoveryIdentity, s.ActiveRung)
	}
	if s.ActiveRung == NoRung && !s.RatchetLevel.Valid() {
		return fmt.Errorf("%w: invalid ratchet level %q", ErrRecoveryIdentity, s.RatchetLevel)
	}
	validAction := s.Action == ActionNone || s.Action == ActionLadderHoldStopPromoted || s.Action.Orderable()
	if !validAction {
		return fmt.Errorf("%w: unknown action %q", ErrRecoveryIdentity, s.Action)
	}
	if s.Action.Orderable() {
		if strings.TrimSpace(s.Ratio) == "" || strings.TrimSpace(s.ProjectedQuantity) == "" {
			return fmt.Errorf("%w: orderable action lacks projection", ErrRecoveryIdentity)
		}
		projectedZero := strings.TrimSpace(s.ProjectedQuantity) == "0"
		if s.Orderable == projectedZero || s.StateOnly == s.Orderable {
			return fmt.Errorf("%w: action/projection flags disagree", ErrRecoveryIdentity)
		}
	} else if s.Orderable || strings.TrimSpace(s.ProjectedQuantity) != "0" {
		return fmt.Errorf("%w: non-orderable action carries an order projection", ErrRecoveryIdentity)
	}
	return nil
}

func validSHA256Identity(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	for _, r := range strings.TrimPrefix(value, "sha256:") {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func compareRecoveryDecimal(a, b string) (int, error) {
	ar, err := positive("recovery value", a)
	if err != nil {
		return 0, err
	}
	br, err := positive("recovery value", b)
	if err != nil {
		return 0, err
	}
	return ar.Cmp(br), nil
}

func compareRecoveryStage(a, b ExitLineSnapshot) (int, error) {
	if a.ActiveRung != NoRung || b.ActiveRung != NoRung {
		if a.ActiveRung == NoRung || b.ActiveRung == NoRung {
			return 0, ErrRecoveryIdentity
		}
		switch {
		case a.ActiveRung < b.ActiveRung:
			return -1, nil
		case a.ActiveRung > b.ActiveRung:
			return 1, nil
		default:
			return 0, nil
		}
	}
	rank := func(level Level) (int, bool) {
		for i, candidate := range []Level{LevelNone, LevelHalfRisk, LevelBreakeven, LevelPartialLock, LevelProfitLock} {
			if level == candidate {
				return i, true
			}
		}
		return 0, false
	}
	ar, aok := rank(a.RatchetLevel)
	br, bok := rank(b.RatchetLevel)
	if !aok || !bok {
		return 0, ErrRecoveryIdentity
	}
	switch {
	case ar < br:
		return -1, nil
	case ar > br:
		return 1, nil
	default:
		return 0, nil
	}
}

func sameRecoveryPolicy(a, b PolicyIdentity) bool {
	return strings.TrimSpace(a.ID) == strings.TrimSpace(b.ID) &&
		strings.TrimSpace(a.Version) == strings.TrimSpace(b.Version) &&
		strings.TrimSpace(a.Digest) == strings.TrimSpace(b.Digest)
}
