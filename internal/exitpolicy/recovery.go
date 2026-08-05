package exitpolicy

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrRecoveryAmbiguous = errors.New("exitpolicy: recovery candidates cross safety axes")
	ErrRecoveryIdentity  = errors.New("exitpolicy: recovery candidate identity mismatch")
)

// RecoverySelectorRevision is the revision of the recovery-selection semantics:
// what SelectRecoverySnapshot, compareRecoveryStage and validateRecoverySnapshot
// accept and what they refuse.
//
// # Why a quarantine carries it
//
// A quarantine is a judgement — "these two candidates cannot be reduced to one
// safe line". Only two things can change that judgement: the inputs, or the
// judge. A quarantined generation writes no new snapshot (that is what being
// quarantined means), so its inputs are frozen and its policy identity is fixed
// by definition. The judge is the only thing left, and the moment it changes is
// therefore the only moment a retry can produce new information rather than
// repeating the same arithmetic.
//
// Stamping the row means a build can tell its own judgement from its
// predecessor's and re-judge exactly once per revision. Without it, a fixed
// comparator leaves the rows its broken predecessor wrote sitting in the ledger
// forever — a078 fixed the first-rung refusal on 2026-08-03 and three positions
// stayed unjudged, stop included, because the quarantine cuts in ahead of the
// comparison and the fixed code is never reached (a078 issues.md I1).
//
// # When to bump it
//
// Bump when a change alters which pairs SelectRecoverySnapshot accepts or
// refuses. Do not bump for refactors, comments, or renames that leave the
// decision identical: an unnecessary bump re-judges every quarantined position,
// which is safe but noisy.
//
// # Both directions of getting it wrong are safe
//
//	forgot to bump   no automatic retry. That is today's behaviour — an operator
//	                 lifts it from the console (a079). Nothing new breaks.
//	bumped needlessly the retry runs the same selector, gets the same answer, and
//	                 re-quarantines under the current revision. One extra row.
//
// # Why 2 and not 1
//
// 1 is reserved and never written. NULL is "recorded before this column existed",
// and leaving a number between it and the first stamped revision means a reader
// of the ledger who asks "where is 1?" finds this paragraph.
const RecoverySelectorRevision int64 = 2

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
	Ladder  *LadderRecoveryDefinition  `json:"ladder,omitempty"`
}

type RatchetRecoveryDefinition struct {
	Evaluation RatchetSnapshotInput `json:"evaluation"`
}

type LadderRecoveryDefinition struct {
	Evaluation LadderSnapshotInput `json:"evaluation"`
}

func NewRatchetRecoveryPolicy(evaluation RatchetSnapshotInput) RecoveryPolicyDefinition {
	copy := evaluation
	if evaluation.Input.Config != nil {
		config := *evaluation.Input.Config
		copy.Input.Config = &config
	}
	return RecoveryPolicyDefinition{Ratchet: &RatchetRecoveryDefinition{Evaluation: copy}}
}

func NewLadderRecoveryPolicy(evaluation LadderSnapshotInput) RecoveryPolicyDefinition {
	copy := evaluation
	copy.Input.Policy.Rungs = append([]Rung(nil), evaluation.Input.Policy.Rungs...)
	return RecoveryPolicyDefinition{Ladder: &LadderRecoveryDefinition{Evaluation: copy}}
}

// ValidateRecoveryDerivation re-runs the exact immutable evaluator input and
// requires every output field to match. A digest is corruption evidence, not a
// substitute for this semantic check.
func ValidateRecoveryDerivation(line ExitLineSnapshot, definition RecoveryPolicyDefinition) error {
	if err := validateRecoverySnapshot(line); err != nil {
		return err
	}
	if (definition.Ratchet == nil) == (definition.Ladder == nil) {
		return fmt.Errorf("%w: recovery needs exactly one policy definition", ErrRecoveryIdentity)
	}
	var expected ExitLineSnapshot
	var err error
	if ratchet := definition.Ratchet; ratchet != nil {
		expected, err = EvaluateRatchetSnapshot(ratchet.Evaluation)
		if err == nil {
			input := ratchet.Evaluation.Input
			expected = expected.ChangedFromState(input.HighWater, input.Baseline, input.Level, NoRung)
		}
	} else {
		ladder := definition.Ladder
		expected, err = EvaluateLadderSnapshot(ladder.Evaluation)
		if err == nil {
			input := ladder.Evaluation.Input
			expected = expected.ChangedFromState(input.HighWater, input.Baseline,
				LevelNone, input.State.ActivatedRung)
		}
	}
	if err != nil {
		return fmt.Errorf("%w: cannot re-evaluate exact recovery input: %v", ErrRecoveryIdentity, err)
	}
	if !reflect.DeepEqual(expected, line) {
		return fmt.Errorf("%w: exact evaluator output does not match stored exit line", ErrRecoveryIdentity)
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

// compareRecoveryStage reports how far a is ahead of b along the policy's own
// stage axis: negative when b leads, positive when a leads, zero when they are
// level.
//
// It answers ordering only. Whether the two candidates are even the same thing
// was settled by the caller, which has already required an identical position,
// generation, policy id/version/digest, entry price and initial stop before
// reaching here. That guarantee is what makes a one-sided rung comparable: a
// ratchet policy never activates a rung and a ladder policy never uses a ratchet
// level, so within one policy identity "b has no rung and a has rung 0" can only
// mean the ladder has just activated its first one.
//
// This function used to refuse exactly that (change a078). NoRung is -1 and rung
// indices start at 0, so the ordering below already places "not activated yet"
// below every rung; the refusal was a duplicate of the caller's identity check
// that fired on the one input it was not meant to catch. It quarantined a live
// position at the moment it crossed its first take-profit line, which stopped
// the engine judging that position at all — stop included.
func compareRecoveryStage(a, b ExitLineSnapshot) (int, error) {
	if a.ActiveRung != NoRung || b.ActiveRung != NoRung {
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
