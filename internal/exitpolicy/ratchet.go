package exitpolicy

// ratchet.go is StockOS `exit/baseline_ratchet.py`, ported, with the four
// corrections design D5 names.
//
// # The trigger table (baseline_ratchet.py:32-41, exit-policy Requirement 2)
//
//	reached R   level          baseline candidate
//	+0.4R       HALF_RISK      entry + risk × (−0.5)
//	+0.8R       BREAKEVEN      the cost-inclusive break-even
//	+1.0R       (unchanged)    propose a 40 % partial take-profit
//	+1.2R       PARTIAL_LOCK   entry + risk × 0.3
//	+2.0R       PROFIT_LOCK    entry + risk × 0.8
//
// `[미검증 — StockOS KOSPI 튜닝값]`. Every number is a field of RatchetConfig
// and none is written into the logic.
//
// # The four corrections
//
//  1. **The baseline exists at t0.** The original's `previous_stop` is optional
//     and a fresh position has none, so nothing protects it until +0.4R. Here
//     the baseline is required and its first value is the entry decision's stop
//     ([OpenRatchet]) — "No Stop = No Trade"'s stop becomes the post-fill
//     judgement's baseline rather than only a pre-trade check.
//  2. **The probe is a required watermark.** The original reads
//     `high_since_entry or current_price` (…:87) and its callers may pass
//     neither, which makes the level a function of the *last* price and
//     therefore not monotone. Here HighWater is a column, the evaluation
//     advances it, and the level is a function of it.
//  3. **The partial is proposed once per position.** The original re-suggests
//     40 % on every evaluation above +1.0R (…:113) and leaves de-duplication to
//     its caller. TossOS has no such caller — the observation loop would submit
//     one every five seconds — so `TakenRatioTotal > 0` suppresses it
//     (exit-policy: 포지션당 1회), as does an unresolved proposal.
//  4. **A breach of the baseline is an output.** The original module is
//     stop-*raising* only; the breach lives in its `exit_strategy` pipeline,
//     which TossOS does not port (it is signal-shaped). The rule is in the spec
//     as the t0 requirement's own scenario, so it is evaluated here, against the
//     price just observed and the baseline this evaluation just composed.
//
// # Ordering inside one evaluation
//
// Advance the watermark, compose the baseline, then judge the breach against the
// **new** baseline. The composed baseline is never below the previous one, so
// judging against it fires whenever judging against the old one would have, and
// sometimes sooner. §0.9 allows the change in that direction and only that one.
//
// A breach outranks a take-profit: it is the risk-reducing action, and the two
// cannot both be proposed by one evaluation.

import (
	"errors"
	"fmt"
	"math/big"
)

// ErrRefused is the class of every refusal to judge. exit-policy requires an
// alert on it, so it must be distinguishable from a judgement that decided
// nothing needed doing.
var ErrRefused = errors.New("exitpolicy: the evaluation inputs are not usable")

// Level is the ratchet's step. The strings are `exit_states.ratchet_level`'s
// CHECK values, which are the original's enum values (baseline_ratchet.py:23-28).
type Level string

const (
	// LevelNone is where a position opens. It means "no ratchet step has been
	// taken", never "no protection": the baseline is already the entry stop.
	LevelNone        Level = "NONE"
	LevelHalfRisk    Level = "HALF_RISK"
	LevelBreakeven   Level = "BREAKEVEN"
	LevelPartialLock Level = "PARTIAL_LOCK"
	LevelProfitLock  Level = "PROFIT_LOCK"
)

var levelRank = map[Level]int{
	LevelNone: 0, LevelHalfRisk: 1, LevelBreakeven: 2, LevelPartialLock: 3, LevelProfitLock: 4,
}

// Rank orders the levels. It is what "레벨 ≥ BREAKEVEN" is decided by, and what
// the monotone property asserts against.
func (l Level) Rank() int { return levelRank[l] }

// Valid reports whether l is one of the five.
func (l Level) Valid() bool { _, ok := levelRank[l]; return ok }

// Action is what an evaluation proposes. The ladder members keep the original's
// spelling (profit_ladder.py:35-39) so an operator comparing the two systems
// reads the same words; the ratchet members are TossOS's, because the original
// expresses them as a ratio and a boolean rather than as named actions.
type Action string

const (
	ActionNone Action = ""
	// ActionBaselineBreach is the full liquidation the baseline being undercut
	// demands. Risk-reducing, and never withheld for a pending take-profit.
	ActionBaselineBreach Action = "BASELINE_BREACH_EXIT"
	// ActionRatchetPartial is the 40 % take-profit at +1.0R.
	ActionRatchetPartial Action = "RATCHET_PARTIAL"
	// The ladder's four (profit_ladder.py LadderActionKind).
	ActionLadderPartial          Action = "LADDER_PARTIAL"
	ActionLadderTakeProfit       Action = "LADDER_TAKE_PROFIT"
	ActionLadderHoldStopPromoted Action = "LADDER_HOLD_STOP_PROMOTED"
	ActionLadderStop             Action = "STOP_LOSS_LADDER"
)

// Orderable reports whether the action asks for an order, as opposed to
// recording a state change (profit_ladder-side `is_orderable_exit_action`).
// A state-only promotion must not reach the submission path, and it must not arm
// a pending proposal either — there is nothing for a fill to resolve.
func (a Action) Orderable() bool {
	switch a {
	case ActionBaselineBreach, ActionRatchetPartial, ActionLadderPartial,
		ActionLadderTakeProfit, ActionLadderStop:
		return true
	default:
		return false
	}
}

// RiskReducing reports whether the action only ever shrinks exposure. Every
// action here does; the method exists so the classification is read from the
// action rather than assumed at each call site, and so a future action that is
// not risk-reducing cannot be added without answering this.
func (a Action) RiskReducing() bool { return a.Orderable() }

// Reasons an evaluation withheld a proposal the trigger table would have made.
// They are recorded rather than inferred: "no proposal" and "a proposal was
// suppressed" are different facts about the same observation.
const (
	// SuppressedPending is StockOS's audit reason for the same condition
	// (test_profit_ladder.py:233 `ladder_blocked_by_pending_order`).
	SuppressedPending = "blocked_by_pending_proposal"
	// SuppressedAlreadyTaken is the once-per-position rule.
	SuppressedAlreadyTaken = "partial_already_taken"
)

// Proposal is what an evaluation asks the observation loop to submit.
type Proposal struct {
	Action Action
	// Level is the proposal's identity — the ratchet level, or the rung index in
	// decimal text under LADDER. It is what "once per level/rung" is keyed on and
	// what `exit_states.pending_level` holds.
	Level string
	// Ratio is the fraction of the **remaining** quantity to sell, a decimal
	// string in (0, 1]. The cumulative fraction is of the *initial* quantity;
	// [CumulativeAfter] converts between them (the original's denominator rule,
	// profit_ladder.py:57-66).
	Ratio string
}

// Zero reports that nothing was proposed.
func (p Proposal) Zero() bool { return p.Action == ActionNone }

// RatchetConfig is the trigger table. Decimal strings, so a configured value and
// a default are the same kind of thing and are compared the same way.
type RatchetConfig struct {
	HalfRiskTriggerR    string
	BreakevenTriggerR   string
	PartialTriggerR     string
	PartialLockTriggerR string
	ProfitLockTriggerR  string
	HalfRiskStopR       string
	PartialLockStopR    string
	ProfitLockStopR     string
	PartialRatio        string
}

// DefaultRatchetConfig is baseline_ratchet.py:32-41 verbatim.
// `[미검증 — StockOS KOSPI 튜닝값]`.
func DefaultRatchetConfig() RatchetConfig {
	return RatchetConfig{
		HalfRiskTriggerR:    "0.4",
		BreakevenTriggerR:   "0.8",
		PartialTriggerR:     "1.0",
		PartialLockTriggerR: "1.2",
		ProfitLockTriggerR:  "2.0",
		HalfRiskStopR:       "-0.5",
		PartialLockStopR:    "0.3",
		ProfitLockStopR:     "0.8",
		PartialRatio:        "0.4",
	}
}

// Validate refuses a table that is not a ratchet.
//
// The original validates none of this — its config is a frozen dataclass built
// in code. TossOS's is configuration (design D7: ladder rung 정책·… 는 config),
// and a configured table whose triggers descend or whose locks fall would make
// the level and the baseline non-monotone, which is the property everything else
// here is built on.
func (c RatchetConfig) Validate() error {
	triggers := []struct{ field, value string }{
		{"half risk trigger", c.HalfRiskTriggerR},
		{"breakeven trigger", c.BreakevenTriggerR},
		{"partial trigger", c.PartialTriggerR},
		{"partial lock trigger", c.PartialLockTriggerR},
		{"profit lock trigger", c.ProfitLockTriggerR},
	}
	var prev *big.Rat
	for _, t := range triggers {
		r, err := parseRat(t.field, t.value)
		if err != nil {
			return refusal(t.field, err.Error())
		}
		if prev != nil && r.Cmp(prev) <= 0 {
			return refusal(t.field, fmt.Sprintf(
				"trigger %s is not above the trigger before it; the levels would not be reachable in order",
				t.value))
		}
		prev = r
	}

	stops := []struct{ field, value string }{
		{"half risk stop", c.HalfRiskStopR},
		{"partial lock stop", c.PartialLockStopR},
		{"profit lock stop", c.ProfitLockStopR},
	}
	prev = nil
	for _, s := range stops {
		r, err := parseRat(s.field, s.value)
		if err != nil {
			return refusal(s.field, err.Error())
		}
		if prev != nil && r.Cmp(prev) < 0 {
			return refusal(s.field, fmt.Sprintf(
				"stop %s is below the stop of the level before it; a ratchet's baseline does not descend", s.value))
		}
		prev = r
	}

	ratio, err := parseRat("partial ratio", c.PartialRatio)
	if err != nil {
		return refusal("partial ratio", err.Error())
	}
	if ratio.Sign() <= 0 || ratio.Cmp(one) > 0 {
		return refusal("partial ratio", fmt.Sprintf(
			"%s is not a fraction of the remaining quantity in (0, 1]", c.PartialRatio))
	}
	return nil
}

// RatchetInput is one observation and the state it is judged against.
type RatchetInput struct {
	// Entry and InitialStop are frozen at t0. Their difference is the R
	// denominator and it does not move for a partial take-profit or an
	// adjustment (exit-policy SHALL).
	Entry       string
	InitialStop string
	// ObservedPrice is the single latest price. Breach is judged against it.
	ObservedPrice string
	// HighWater is the watermark as it stood before this observation.
	HighWater string
	// Baseline is the protected baseline as it stood before this observation.
	// Required: there is no such thing as an unprotected open position here.
	Baseline string
	// RealBreakeven is the cost-inclusive break-even sell price
	// (costs.Model.BreakEvenSellPrice).
	RealBreakeven string
	// TakenRatioTotal is the cumulative taken fraction of the initial quantity.
	// "" reads as "0".
	TakenRatioTotal string
	// PendingAction is the outstanding proposal's action, empty when none.
	PendingAction Action
	// Config is nil for the default table.
	Config *RatchetConfig
}

// RatchetDecision is what one observation implies.
type RatchetDecision struct {
	// HighWater is the watermark after this observation: max(previous, observed).
	HighWater string
	// RMultiple is the watermark's R, rendered for the audit trail.
	RMultiple string
	// Level is the step the watermark has reached.
	Level Level
	// Baseline is the composed baseline. Never below the input's.
	Baseline string
	// Raised is the strict `>` exit-policy records a rise on. A baseline that
	// merely re-composed to the same number is not a rise and must not produce an
	// event that says one happened.
	Raised bool
	// WinningReason names the candidate the baseline came from.
	WinningReason string
	// Proposal is what to submit, or the zero value.
	Proposal Proposal
	// Suppressed names why a proposal the table would have made was withheld.
	Suppressed string
	// CancelPendingFirst is set when a risk-reducing proposal has to displace an
	// outstanding take-profit. Two working sell orders on one position can
	// oversell it, so the observation loop cancels before it submits (task 7.4).
	CancelPendingFirst bool
}

// Changed reports whether this evaluation moved anything worth recording — the
// condition `exit_events` is appended on (task 7.4), stated here so the loop does
// not re-derive it.
func (d RatchetDecision) Changed(previousHighWater string, previousLevel Level) bool {
	return d.Raised || d.Level != previousLevel || d.HighWater != previousHighWater ||
		!d.Proposal.Zero()
}

// OpenRatchet is the t0 state of a position under the ratchet policy.
type OpenRatchet struct {
	Baseline    string
	HighWater   string
	InitialRisk string
	Level       Level
}

// OpenRatchet computes the state a position opens with.
//
// The baseline is the entry decision's stop, which is D5's first correction and
// the reason `exit_states.baseline_price` is NOT NULL with no default. The
// watermark starts at the entry price: the position was acquired there, so that
// is the highest price it is known to have traded at since it existed, and
// seeding it lower would make the first observation look like a new high.
func OpenRatchetState(entry, initialStop string) (OpenRatchet, error) {
	e, s, risk, err := riskOf(entry, initialStop)
	if err != nil {
		return OpenRatchet{}, err
	}
	return OpenRatchet{
		Baseline:    formatPrice(s),
		HighWater:   formatPrice(e),
		InitialRisk: formatPrice(risk),
		Level:       LevelNone,
	}, nil
}

// EvaluateRatchet judges one observation.
func EvaluateRatchet(in RatchetInput) (RatchetDecision, error) {
	cfg := DefaultRatchetConfig()
	if in.Config != nil {
		cfg = *in.Config
	}
	if err := cfg.Validate(); err != nil {
		return RatchetDecision{}, err
	}
	entry, _, risk, err := riskOf(in.Entry, in.InitialStop)
	if err != nil {
		return RatchetDecision{}, err
	}
	observed, err := positive("observed price", in.ObservedPrice)
	if err != nil {
		return RatchetDecision{}, err
	}
	previousHigh, err := positive("high water", in.HighWater)
	if err != nil {
		return RatchetDecision{}, err
	}
	previousBaseline, err := positive("baseline", in.Baseline)
	if err != nil {
		return RatchetDecision{}, err
	}
	realBreakeven, err := positive("real breakeven", in.RealBreakeven)
	if err != nil {
		return RatchetDecision{}, err
	}
	taken, err := fraction("taken ratio total", in.TakenRatioTotal)
	if err != nil {
		return RatchetDecision{}, err
	}

	// 1. The watermark. Monotone by construction, which is what makes the level
	//    monotone and therefore what makes the baseline monotone.
	probe := previousHigh
	if observed.Cmp(probe) > 0 {
		probe = observed
	}
	rMultiple := new(big.Rat).Quo(new(big.Rat).Sub(probe, entry), risk)

	// 2. The level and its own candidate (_ratchet_candidate).
	level, levelCandidate, wantsPartial, err := ratchetCandidate(cfg, entry, risk, rMultiple, realBreakeven)
	if err != nil {
		return RatchetDecision{}, err
	}

	decision := RatchetDecision{
		HighWater:     formatPrice(probe),
		RMultiple:     formatRMultiple(rMultiple),
		Level:         level,
		Baseline:      formatPrice(previousBaseline),
		WinningReason: CandidatePrevious,
	}

	// 3. The composition. Below the first trigger there is no candidate at all
	//    and the baseline stays where it is — the original returns early here
	//    (baseline_ratchet.py:92-93) and so does this, rather than composing a
	//    max whose only term is the anchor.
	if levelCandidate != "" {
		// The real break-even is excluded below BREAKEVEN
		// (baseline_ratchet.py:96): at HALF_RISK the policy is deliberately still
		// below the entry, and admitting the break-even there would jump the
		// baseline three steps on the first trigger.
		breakevenTerm := ""
		if level.Rank() >= LevelBreakeven.Rank() {
			breakevenTerm = in.RealBreakeven
		}
		out, err := ComputeProtectedStop(
			BuildCandidates(in.Baseline, breakevenTerm, levelCandidate), in.Baseline)
		if err != nil {
			return RatchetDecision{}, err
		}
		composed, err := parseRat("composed baseline", out.Price)
		if err != nil {
			return RatchetDecision{}, refusal("composed baseline", err.Error())
		}
		decision.Baseline = out.Price
		decision.WinningReason = out.WinningReason
		decision.Raised = composed.Cmp(previousBaseline) > 0
	}

	// 4. The breach, against the baseline this evaluation just composed.
	baseline, err := parseRat("baseline", decision.Baseline)
	if err != nil {
		return RatchetDecision{}, refusal("baseline", err.Error())
	}
	if observed.Cmp(baseline) < 0 {
		if in.PendingAction == ActionBaselineBreach {
			decision.Suppressed = SuppressedPending
			return decision, nil
		}
		decision.Proposal = Proposal{
			Action: ActionBaselineBreach,
			Level:  string(level),
			Ratio:  "1",
		}
		decision.CancelPendingFirst = in.PendingAction != ActionNone
		return decision, nil
	}

	// 5. The take-profit. Withheld while a proposal is unresolved, and withheld
	//    for good once any of the position has been taken.
	if wantsPartial {
		switch {
		case taken.Sign() > 0:
			decision.Suppressed = SuppressedAlreadyTaken
		case in.PendingAction != ActionNone:
			decision.Suppressed = SuppressedPending
		default:
			decision.Proposal = Proposal{
				Action: ActionRatchetPartial,
				Level:  string(level),
				Ratio:  cfg.PartialRatio,
			}
		}
	}
	return decision, nil
}

// ratchetCandidate is `_ratchet_candidate` (baseline_ratchet.py:104-135): the
// level the R multiple has reached, that level's own baseline candidate, and
// whether the partial trigger is met.
//
// The partial flag is computed before the level cascade and independently of it,
// exactly as the original does, because +1.0R sits between two levels and
// promotes neither.
func ratchetCandidate(cfg RatchetConfig, entry, risk, rMultiple, realBreakeven *big.Rat) (
	Level, string, bool, error) {
	reached := func(field, trigger string) (bool, error) {
		t, err := parseRat(field, trigger)
		if err != nil {
			return false, refusal(field, err.Error())
		}
		return rMultiple.Cmp(t) >= 0, nil
	}
	at := func(field, r string) (string, error) {
		m, err := parseRat(field, r)
		if err != nil {
			return "", refusal(field, err.Error())
		}
		return formatPrice(new(big.Rat).Add(entry, new(big.Rat).Mul(risk, m))), nil
	}

	partial, err := reached("partial trigger", cfg.PartialTriggerR)
	if err != nil {
		return LevelNone, "", false, err
	}

	if hit, err := reached("profit lock trigger", cfg.ProfitLockTriggerR); err != nil {
		return LevelNone, "", false, err
	} else if hit {
		price, err := at("profit lock stop", cfg.ProfitLockStopR)
		return LevelProfitLock, price, partial, err
	}
	if hit, err := reached("partial lock trigger", cfg.PartialLockTriggerR); err != nil {
		return LevelNone, "", false, err
	} else if hit {
		price, err := at("partial lock stop", cfg.PartialLockStopR)
		return LevelPartialLock, price, partial, err
	}
	if hit, err := reached("breakeven trigger", cfg.BreakevenTriggerR); err != nil {
		return LevelNone, "", false, err
	} else if hit {
		return LevelBreakeven, formatPrice(realBreakeven), partial, nil
	}
	if hit, err := reached("half risk trigger", cfg.HalfRiskTriggerR); err != nil {
		return LevelNone, "", false, err
	} else if hit {
		price, err := at("half risk stop", cfg.HalfRiskStopR)
		return LevelHalfRisk, price, partial, err
	}
	// Below the first trigger nothing is proposed at all — the original zeroes the
	// partial here too (…:135), which matters only for a configured table whose
	// partial trigger sits below its first level trigger.
	return LevelNone, "", false, nil
}

// CumulativeAfter converts a proposal's ratio of the *remaining* quantity into
// the cumulative ratio of the *initial* quantity it will produce when it fills.
//
//	after = taken + (1 − taken) × ratio
//
// It is the denominator rule (profit_ladder.py:57-66, exit-policy: 누적 비율은
// 초기 수량 기준, 각 발의 비율은 잔여 수량 기준) written once, so the two bases
// cannot drift apart in two call sites.
func CumulativeAfter(takenTotal, ratioOfRemaining string) (string, error) {
	taken, err := fraction("taken ratio total", takenTotal)
	if err != nil {
		return "", err
	}
	ratio, err := fraction("ratio of remaining", ratioOfRemaining)
	if err != nil {
		return "", err
	}
	remaining := new(big.Rat).Sub(one, taken)
	return formatRatio(new(big.Rat).Add(taken, new(big.Rat).Mul(remaining, ratio))), nil
}

// TakenAfterFill is the cumulative taken fraction after a sell of `sold` units
// left `remainingAfter` units held.
//
// It reconstructs the initial quantity rather than storing it:
//
//	remaining_before = remaining_after + sold
//	initial          = remaining_before ÷ (1 − taken_before)
//	after            = taken_before + sold × (1 − taken_before) ÷ remaining_before
//
// The identity is exact and needs no column, which is why `exit_states` has none
// — D7's table carries `initial_quantity` on `trade_outcomes` and not here, and
// adding one would have been a second authority on the same number.
func TakenAfterFill(takenBefore, sold, remainingAfter string) (string, error) {
	taken, err := fraction("taken ratio total", takenBefore)
	if err != nil {
		return "", err
	}
	soldQty, err := nonNegative("sold quantity", sold)
	if err != nil {
		return "", err
	}
	remaining, err := nonNegative("remaining quantity", remainingAfter)
	if err != nil {
		return "", err
	}
	before := new(big.Rat).Add(remaining, soldQty)
	if before.Sign() == 0 {
		// Nothing was held and nothing moved. There is no fraction of nothing.
		return formatRatio(taken), nil
	}
	share := new(big.Rat).Quo(new(big.Rat).Mul(soldQty, new(big.Rat).Sub(one, taken)), before)
	return formatRatio(new(big.Rat).Add(taken, share)), nil
}

// --- input validation ---------------------------------------------------------
//
// `_validate_inputs` (baseline_ratchet.py:155-170), ported field for field, plus
// the two fields TossOS added. A violation is a refusal to judge: exit-policy
// requires the observation loop to alert and to leave the state untouched, and
// returning "no change" would be indistinguishable from a quiet judgement.

func riskOf(entry, initialStop string) (*big.Rat, *big.Rat, *big.Rat, error) {
	e, err := positive("entry price", entry)
	if err != nil {
		return nil, nil, nil, err
	}
	s, err := positive("initial stop", initialStop)
	if err != nil {
		return nil, nil, nil, err
	}
	if s.Cmp(e) >= 0 {
		return nil, nil, nil, refusal("initial stop", fmt.Sprintf(
			"%s is not below the entry price %s; TossOS is long-only, so the risk per unit does not exist",
			initialStop, entry))
	}
	return e, s, new(big.Rat).Sub(e, s), nil
}

func positive(field, value string) (*big.Rat, error) {
	r, err := parseRat(field, value)
	if err != nil {
		return nil, refusal(field, err.Error())
	}
	if r.Sign() <= 0 {
		return nil, refusal(field, fmt.Sprintf("%s is not a positive price", value))
	}
	return r, nil
}

func nonNegative(field, value string) (*big.Rat, error) {
	r, err := parseRat(field, value)
	if err != nil {
		return nil, refusal(field, err.Error())
	}
	if r.Sign() < 0 {
		return nil, refusal(field, fmt.Sprintf("%s is negative", value))
	}
	return r, nil
}

// fraction reads a ratio in [0, 1]. The empty string is 0: `taken_ratio_total`
// defaults to '0' in the schema and a caller that read the column before
// anything was taken holds the same fact.
func fraction(field, value string) (*big.Rat, error) {
	r, err := parseRatOr(field, value, "0")
	if err != nil {
		return nil, refusal(field, err.Error())
	}
	if r.Sign() < 0 || r.Cmp(one) > 0 {
		return nil, refusal(field, fmt.Sprintf("%s is not a fraction in [0, 1]", value))
	}
	return r, nil
}

// mustRat is for values this package produced and has already validated.
func mustRat(s string) *big.Rat {
	r, _ := new(big.Rat).SetString(s)
	if r == nil {
		return new(big.Rat)
	}
	return r
}
