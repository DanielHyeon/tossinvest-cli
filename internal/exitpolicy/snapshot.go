package exitpolicy

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// SnapshotContext binds one pure evaluation to one position generation and one
// explicit observation. Callers must reuse the resulting snapshot instead of
// independently re-evaluating the same quote for display and execution.
type SnapshotContext struct {
	PositionID         string
	PositionGeneration int64
	ObservationID      string
	RemainingQuantity  string
}

// ExitLineSnapshot is a value-only immutable decision record. It contains no
// slices, maps, pointers, clocks, or callbacks, so concurrent consumers can
// safely receive value copies of exactly the same evaluated fact.
type ExitLineSnapshot struct {
	SnapshotID  string
	DecisionID  string
	InputDigest string

	Policy             PolicyIdentity
	PositionID         string
	PositionGeneration int64
	ObservationID      string

	EntryPrice        string
	InitialStop       string
	ObservedPrice     string
	CurrentProtection string
	HighWater         string
	RatchetLevel      Level
	ActiveRung        int
	NextTarget        string
	NextProtection    string

	Action             Action
	Level              string
	Ratio              string
	ProjectedQuantity  string
	Orderable          bool
	StateOnly          bool
	Suppressed         string
	CancelPendingFirst bool
	Changed            bool
}

func (s ExitLineSnapshot) ExecutableProposal() Proposal {
	if !s.Orderable || s.ProjectedQuantity == "0" {
		return Proposal{}
	}
	return Proposal{Action: s.Action, Level: s.Level, Ratio: s.Ratio}
}

type LadderSnapshotInput struct {
	Context SnapshotContext
	Input   LadderInput
}

type RatchetSnapshotInput struct {
	Context SnapshotContext
	Input   RatchetInput
}

// ProjectWholeShares applies the existing whole-share contract exactly: ratio
// defaults to one, caps at one, multiplies as a rational, and rounds down.
func ProjectWholeShares(remaining, ratio string) (string, error) {
	held, ok := new(big.Rat).SetString(strings.TrimSpace(remaining))
	if !ok {
		return "", fmt.Errorf("%q is not a quantity", remaining)
	}
	share := big.NewRat(1, 1)
	if r := strings.TrimSpace(ratio); r != "" {
		share, ok = new(big.Rat).SetString(r)
		if !ok {
			return "", fmt.Errorf("%q is not a ratio", ratio)
		}
	}
	if share.Cmp(one) > 0 {
		share = new(big.Rat).Set(one)
	}
	product := new(big.Rat).Mul(held, share)
	units := new(big.Int).Quo(product.Num(), product.Denom())
	if units.Sign() < 0 {
		units.SetInt64(0)
	}
	return units.String(), nil
}

func EvaluateLadderSnapshot(in LadderSnapshotInput) (ExitLineSnapshot, error) {
	identity, err := in.Input.Policy.Identity()
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	ctxFields, quantity, err := canonicalSnapshotContext(in.Context)
	if err != nil {
		return ExitLineSnapshot{}, err
	}

	eval := in.Input
	// A caller may omit identity fields while older journal schemas are still in
	// use, but a supplied identity is evidence and must never be overwritten.
	// EvaluateLadder below compares it with the executable table and fails closed
	// on any mismatch.
	if strings.TrimSpace(eval.State.PolicyID) == "" {
		eval.State.PolicyID = identity.ID
	}
	if strings.TrimSpace(eval.State.PolicyVersion) == "" {
		eval.State.PolicyVersion = identity.Version
	}
	if strings.TrimSpace(eval.State.PolicyDigest) == "" {
		eval.State.PolicyDigest = identity.Digest
	}
	transition, err := EvaluateLadder(eval)
	if err != nil {
		return ExitLineSnapshot{}, err
	}

	fields, err := canonicalLadderInput(eval, identity)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	inputDigest := digestFields("tossos.exit-line.input.v1", append(ctxFields, fields...)...)

	action, level, ratio := transition.Action, "", ""
	if !transition.Proposal.Zero() {
		action, level, ratio = transition.Proposal.Action, transition.Proposal.Level, transition.Proposal.Ratio
	}
	projected := "0"
	orderable := false
	stateOnly := action == ActionLadderHoldStopPromoted
	if action.Orderable() {
		projected, err = ProjectWholeShares(quantity, ratio)
		if err != nil {
			return ExitLineSnapshot{}, err
		}
		orderable = projected != "0"
		stateOnly = !orderable
	}
	nextTarget, nextProtection, err := nextLadderLine(eval.EntryPrice, transition.NextState.ActivatedRung, eval.Policy)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	snapshot := ExitLineSnapshot{
		InputDigest: inputDigest, Policy: identity,
		PositionID: in.Context.PositionID, PositionGeneration: in.Context.PositionGeneration,
		ObservationID: in.Context.ObservationID,
		EntryPrice:    eval.EntryPrice, InitialStop: initialStopFrom(eval.InitialStop, eval.Baseline),
		ObservedPrice:     eval.ObservedPrice,
		CurrentProtection: transition.Baseline, HighWater: transition.HighWater,
		RatchetLevel: LevelNone, ActiveRung: transition.NextState.ActivatedRung,
		NextTarget: nextTarget, NextProtection: nextProtection,
		Action: action, Level: level, Ratio: ratio, ProjectedQuantity: projected,
		Orderable: orderable, StateOnly: stateOnly, Suppressed: transition.Suppressed,
		CancelPendingFirst: transition.CancelPendingFirst,
		Changed: transition.Raised || transition.HighWater != eval.HighWater ||
			transition.RungPromotedTo != NoRung || orderable,
	}
	snapshot.finishIDs()
	return snapshot, nil
}

// WithInitialStop returns a value copy with the frozen t0 stop included. The
// field is descriptive and does not participate in a decision already made.
func (s ExitLineSnapshot) WithInitialStop(stop string) ExitLineSnapshot {
	s.InitialStop = strings.TrimSpace(stop)
	return s
}

func EvaluateRatchetSnapshot(in RatchetSnapshotInput) (ExitLineSnapshot, error) {
	cfg := DefaultRatchetConfig()
	if in.Input.Config != nil {
		cfg = *in.Input.Config
	}
	identity, err := RatchetPolicyIdentity(cfg)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	ctxFields, quantity, err := canonicalSnapshotContext(in.Context)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	decision, err := EvaluateRatchet(in.Input)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	fields, err := canonicalRatchetInput(in.Input, cfg, identity)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	inputDigest := digestFields("tossos.exit-line.input.v1", append(ctxFields, fields...)...)

	action, level, ratio := decision.Proposal.Action, decision.Proposal.Level, decision.Proposal.Ratio
	projected := "0"
	orderable := false
	stateOnly := false
	if action.Orderable() {
		projected, err = ProjectWholeShares(quantity, ratio)
		if err != nil {
			return ExitLineSnapshot{}, err
		}
		orderable = projected != "0"
		stateOnly = !orderable
	}
	nextTarget, nextProtection, err := nextRatchetLine(in.Input, cfg, decision)
	if err != nil {
		return ExitLineSnapshot{}, err
	}
	previousLevel := in.Input.Level
	if !previousLevel.Valid() {
		previousLevel, err = ratchetLevelAtPreviousHigh(in.Input, cfg)
		if err != nil {
			return ExitLineSnapshot{}, err
		}
	}
	snapshot := ExitLineSnapshot{
		InputDigest: inputDigest, Policy: identity,
		PositionID: in.Context.PositionID, PositionGeneration: in.Context.PositionGeneration,
		ObservationID: in.Context.ObservationID,
		EntryPrice:    in.Input.Entry, InitialStop: in.Input.InitialStop,
		ObservedPrice:     in.Input.ObservedPrice,
		CurrentProtection: decision.Baseline, HighWater: decision.HighWater,
		RatchetLevel: decision.Level, ActiveRung: NoRung,
		NextTarget: nextTarget, NextProtection: nextProtection,
		Action: action, Level: level, Ratio: ratio, ProjectedQuantity: projected,
		Orderable: orderable, StateOnly: stateOnly, Suppressed: decision.Suppressed,
		CancelPendingFirst: decision.CancelPendingFirst,
		Changed: decision.Raised || decision.Level != previousLevel ||
			decision.HighWater != in.Input.HighWater || orderable,
	}
	snapshot.finishIDs()
	return snapshot, nil
}

func (s ExitLineSnapshot) ChangedFromState(previousHighWater, previousProtection string,
	previousLevel Level, previousRung int) ExitLineSnapshot {
	if s.ActiveRung == NoRung {
		s.Changed = s.HighWater != previousHighWater || s.CurrentProtection != previousProtection ||
			s.RatchetLevel != previousLevel || s.Orderable
	} else {
		s.Changed = s.HighWater != previousHighWater || s.CurrentProtection != previousProtection ||
			s.ActiveRung != previousRung || s.Orderable
	}
	return s
}

func (s *ExitLineSnapshot) finishIDs() {
	s.SnapshotID = "els_" + strings.TrimPrefix(digestFields("snapshot", s.InputDigest, s.Policy.Digest), "sha256:")
	s.DecisionID = "eld_" + strings.TrimPrefix(digestFields("decision", s.SnapshotID,
		string(s.Action), s.Level, s.Ratio, s.ProjectedQuantity, strconv.FormatBool(s.Orderable)), "sha256:")
}

func canonicalSnapshotContext(ctx SnapshotContext) ([]string, string, error) {
	if strings.TrimSpace(ctx.PositionID) == "" || ctx.PositionGeneration < 0 || strings.TrimSpace(ctx.ObservationID) == "" {
		return nil, "", refusal("snapshot identity", "position, non-negative generation, and observation identity are required")
	}
	quantity, err := positive("remaining quantity", ctx.RemainingQuantity)
	if err != nil {
		return nil, "", err
	}
	canonicalQuantity := quantity.RatString()
	return []string{strings.TrimSpace(ctx.PositionID), strconv.FormatInt(ctx.PositionGeneration, 10),
		strings.TrimSpace(ctx.ObservationID), canonicalQuantity}, canonicalQuantity, nil
}

func canonicalLadderInput(in LadderInput, identity PolicyIdentity) ([]string, error) {
	values := []string{in.EntryPrice, initialStopFrom(in.InitialStop, in.Baseline), in.ObservedPrice,
		in.HighWater, in.Baseline, in.State.TakenRatioTotal}
	fields := []string{identity.ID, identity.Version, identity.Digest}
	for index, value := range values {
		if index == 5 && strings.TrimSpace(value) == "" {
			value = "0"
		}
		canonical, err := canonicalNumber("ladder snapshot input", value)
		if err != nil {
			return nil, err
		}
		fields = append(fields, canonical)
	}
	fields = append(fields, strconv.Itoa(in.State.ActivatedRung), strconv.FormatBool(in.State.Completed),
		string(in.State.PendingAction), strconv.Itoa(in.State.PendingRung))
	return fields, nil
}

func canonicalRatchetInput(in RatchetInput, cfg RatchetConfig, identity PolicyIdentity) ([]string, error) {
	values := []string{in.Entry, in.InitialStop, in.ObservedPrice, in.HighWater, in.Baseline,
		in.RealBreakeven, in.TakenRatioTotal}
	fields := []string{identity.ID, identity.Version, identity.Digest}
	for index, value := range values {
		if index == 6 && strings.TrimSpace(value) == "" {
			value = "0"
		}
		canonical, err := canonicalNumber("ratchet snapshot input", value)
		if err != nil {
			return nil, err
		}
		fields = append(fields, canonical)
	}
	fields = append(fields, string(in.Level), string(in.PendingAction))
	return fields, nil
}

func ratchetLevelAtPreviousHigh(in RatchetInput, cfg RatchetConfig) (Level, error) {
	entry, _, risk, err := riskOf(in.Entry, in.InitialStop)
	if err != nil {
		return LevelNone, err
	}
	high, err := positive("high water", in.HighWater)
	if err != nil {
		return LevelNone, err
	}
	breakeven, err := positive("real breakeven", in.RealBreakeven)
	if err != nil {
		return LevelNone, err
	}
	rMultiple := new(big.Rat).Quo(new(big.Rat).Sub(high, entry), risk)
	level, _, _, err := ratchetCandidate(cfg, entry, risk, rMultiple, breakeven)
	return level, err
}

func nextLadderLine(entry string, active int, policy LadderPolicy) (string, string, error) {
	next := active + 1
	if next < 0 {
		next = 0
	}
	if next >= len(policy.Rungs) {
		return "", "", nil
	}
	rung := policy.Rungs[next]
	target := ""
	if !(policy.PolicyID == CommonLadderRunner && strings.TrimSpace(rung.TargetPct) == "999.0") {
		entryRat, err := positive("entry price", entry)
		if err != nil {
			return "", "", err
		}
		pct, err := parseRat("rung target percent", rung.TargetPct)
		if err != nil {
			return "", "", err
		}
		target = formatPrice(new(big.Rat).Mul(entryRat,
			new(big.Rat).Add(one, new(big.Rat).Quo(pct, hundred))))
	}
	protection, err := LockPrice(entry, rung.StopPct)
	return target, protection, err
}

func nextRatchetLine(in RatchetInput, cfg RatchetConfig, decision RatchetDecision) (string, string, error) {
	entry, _, risk, err := riskOf(in.Entry, in.InitialStop)
	if err != nil {
		return "", "", err
	}
	rMultiple, err := parseRat("r multiple", decision.RMultiple)
	if err != nil {
		return "", "", err
	}
	taken, err := fraction("taken ratio total", in.TakenRatioTotal)
	if err != nil {
		return "", "", err
	}
	type next struct {
		trigger, protection string
		breakeven           bool
	}
	candidates := []next{
		{cfg.HalfRiskTriggerR, cfg.HalfRiskStopR, false},
		{cfg.BreakevenTriggerR, "", true},
		{cfg.PartialTriggerR, "", false},
		{cfg.PartialLockTriggerR, cfg.PartialLockStopR, false},
		{cfg.ProfitLockTriggerR, cfg.ProfitLockStopR, false},
	}
	for index, candidate := range candidates {
		trigger, parseErr := parseRat("ratchet trigger", candidate.trigger)
		if parseErr != nil {
			return "", "", parseErr
		}
		if index == 2 && taken.Sign() > 0 {
			continue
		}
		if rMultiple.Cmp(trigger) >= 0 {
			continue
		}
		target := formatPrice(new(big.Rat).Add(entry, new(big.Rat).Mul(risk, trigger)))
		protection := decision.Baseline
		switch {
		case candidate.breakeven:
			protection = in.RealBreakeven
		case candidate.protection != "":
			stopR, parseErr := parseRat("ratchet protection", candidate.protection)
			if parseErr != nil {
				return "", "", parseErr
			}
			protection = formatPrice(new(big.Rat).Add(entry, new(big.Rat).Mul(risk, stopR)))
		}
		return target, protection, nil
	}
	return "", "", nil
}

func initialStopFrom(initialStop, baseline string) string {
	if strings.TrimSpace(initialStop) != "" {
		return strings.TrimSpace(initialStop)
	}
	return baseline
}
