package exitpolicy

// ladder.go is StockOS `packages/trading/stockos_trading/profit_ladder.py`,
// ported: the user-requested ratchet of "buy → target up → basis adjust → take
// profit → next target up".
//
// # What carried over unchanged
//
// The rung table and its two validation rules (targets strictly increasing,
// locks monotonically non-decreasing — profit_ladder.py:88-102), the default
// four-rung set (…:165-176), the highest-reached-rung scan (…:250-259), the
// three promotion outcomes (final take-full, partial, state-only — …:262-292),
// the denominator rule (cumulative against the initial quantity, each rung
// against the remaining one — …:57-66), and the decision-time / execution-time
// split (…:26-34).
//
// # What changed, and why
//
//  1. **float → decimal.** Every number is a decimal string over a `math/big`
//     rational. `10_000 * (1 + 1.5/100)` is not 10150 in binary floating point,
//     and a rung boundary decided by that error is a boundary nobody specified.
//     The original's own tests compare `protected_stop_pct == 1.0` and get away
//     with it because the values are small dyadic-ish constants; a KRW price
//     times a percent is neither.
//
//  2. **The probe is the watermark, not the bar's high.** The original takes
//     `high`/`low` from an OHLC bar and consults `FillModel.STOP_FIRST` to break
//     a same-bar target-and-stop tie (…:194-208). TossOS observes one price at a
//     time and has no bar, so exit-policy explicitly demotes that model
//     ("OHLC 입력이 있는 백테스트 전용이므로 이 change의 SHALL이 아니다"). Rung
//     attainment reads the watermark and breach reads the observation, which is
//     what the original's live tick path does by calling the evaluator twice.
//
//  3. **The protected stop is a price, composed, and it exists at t0.** The
//     original stores a percent and only checks it once a rung has activated
//     (…:232-233: `if not same_bar_decision and state.activated_rung_index >=
//     0`), so a ladder position is unprotected until its first rung. Here the
//     baseline is `exit_states.baseline_price`, it starts at the entry stop like
//     every other position (D5's first correction), and each rung's stop enters
//     the same [ComputeProtectedStop] composition the ratchet uses — which is
//     what that module was extracted for in the original (the common ladder
//     runner is one of its two callers).
//
//     A consequence: there is no `protected_stop_pct` state field. D7's
//     `exit_states` carries `active_rung` and `baseline_price` and no percent
//     column, so the percent is derived from the rung table when it is needed. A
//     stored percent that disagreed with its stored rung would be a state nobody
//     could explain.
//
//  4. **Promotion is judged before breach.** The original checks the stop first
//     (…:230-236). With a watermark probe, promoting first is the strictly more
//     protective order: a promoted rung's lock is never below the previous one,
//     so the breach fires whenever it would have fired against the older lock
//     and sometimes sooner. §0.9 permits movement in that direction only.
//
//  5. **The pending lifecycle is inside the evaluator.** The original blocks a
//     duplicate order one layer up, in `exit_strategy` (test_profit_ladder.py:208
//     — audit reason `ladder_blocked_by_pending_order`), which TossOS does not
//     port because that pipeline is signal-shaped. The rule is the same and the
//     audit reason is kept.
//
//  6. **A breach is never withheld for a pending take-profit.** §0 invariant 3:
//     nothing may delay a liquidation. The evaluator instead reports that the
//     working order has to be cancelled first (task 7.4), because two working
//     sells on one position can oversell it.
//
// # What was not ported
//
// `FillModel` and the same-bar tie-break (backtest-only, see 2 above),
// `ladder_order_idempotency_key` (TossOS's idempotency key is the journal's
// `f(decision_id, generation)` from 2a, and a second key scheme would be a
// second authority on the same question), and `LimitUpHoldState` (the limit-up
// close-defer domain, which is not in this change).

import (
	"fmt"
	"math/big"
	"strconv"
)

// The audit reasons, kept as the original spells them (profit_ladder.py) so an
// operator comparing the two systems reads the same words.
const (
	ReasonNoRungPromotion  = "no_rung_promotion"
	ReasonFinalRungTakeAll = "ladder_final_rung_take_full"
	ReasonRungPartialTake  = "ladder_rung_partial_take"
	ReasonStateOnly        = "ladder_rung_state_only_promotion"
	ReasonStopBreached     = "protected_stop_breached"
	ReasonBlockedByPending = "ladder_blocked_by_pending_order"
	ReasonCompleted        = "ladder_completed"
)

// NoRung is the rung index of a position that has activated none — the
// original's `activated_rung_index = -1` (profit_ladder.py:110).
const NoRung = -1

// Rung is one step of the ratchet.
//
// StopPct is the protected stop relative to the entry price, in percent
// (positive = above entry, 0 = breakeven). PartialRatio is taken from the
// **remaining** quantity at that rung, not from the entry quantity — the
// cumulative total is initial-quantity-based and each rung's ratio is
// remaining-quantity-based (profit_ladder.py:57-66).
type Rung struct {
	TargetPct    string
	StopPct      string
	PartialRatio string
}

// Validate is `LadderRung.__post_init__` (profit_ladder.py:72-76).
func (r Rung) Validate() error {
	if _, err := parseRat("rung target percent", r.TargetPct); err != nil {
		return refusal("rung target percent", err.Error())
	}
	if _, err := parseRat("rung stop percent", r.StopPct); err != nil {
		return refusal("rung stop percent", err.Error())
	}
	ratio, err := parseRatio("rung partial ratio", r.PartialRatio)
	if err != nil {
		return refusal("rung partial ratio", err.Error())
	}
	if ratio.Sign() < 0 || ratio.Cmp(one) > 0 {
		return refusal("rung partial ratio", fmt.Sprintf(
			"partial_ratio must be in [0, 1], got %s", r.PartialRatio))
	}
	return nil
}

// LadderPolicy is an immutable rung set.
type LadderPolicy struct {
	PolicyID string
	Rungs    []Rung
	// FinalTakeFull makes the last rung liquidate the remainder rather than take
	// its own partial (profit_ladder.py:159).
	FinalTakeFull bool
	// RunnerTrailPct protects the remainder from the high-water mark after the
	// final rung. Empty disables the runner.
	RunnerTrailPct string
}

// DefaultLadderPolicy is `default_policy()` (profit_ladder.py:163-176) verbatim:
// 1.5/2.5/4.0/6.0 % targets, 0/1.0/2.0/3.5 % locks, 0/0.25/0.25/1.0 partials.
//
// `[미검증 — StockOS KOSPI 튜닝값]`. The set was tuned against a different broker
// on a different venue and TossOS has measured none of it; §0.9 permits changing
// these only in the conservative direction until a tracer produces evidence.
func DefaultLadderPolicy() LadderPolicy {
	return LadderPolicy{
		PolicyID: "default_v1",
		Rungs: []Rung{
			{TargetPct: "1.5", StopPct: "0", PartialRatio: "0"},
			{TargetPct: "2.5", StopPct: "1.0", PartialRatio: "0.25"},
			{TargetPct: "4.0", StopPct: "2.0", PartialRatio: "0.25"},
			{TargetPct: "6.0", StopPct: "3.5", PartialRatio: "1.0"},
		},
		FinalTakeFull: true,
	}
}

// Validate is `ProfitLadderPolicy.__post_init__` (profit_ladder.py:88-102).
//
// The two ordering rules are the ratchet contract itself: a target that does not
// rise makes a rung unreachable, and a lock that falls makes the protected stop
// descend, which is the one thing every part of this package promises cannot
// happen.
func (p LadderPolicy) Validate() error {
	if p.PolicyID == "" {
		return refusal("ladder policy id", "policy_id must be non-empty")
	}
	if len(p.Rungs) == 0 {
		return refusal("ladder policy", "a ladder policy requires at least one rung")
	}
	for i, rung := range p.Rungs {
		if err := rung.Validate(); err != nil {
			return err
		}
		if i == 0 {
			continue
		}
		prev := p.Rungs[i-1]
		target, _ := parseRat("target", rung.TargetPct)
		prevTarget, _ := parseRat("target", prev.TargetPct)
		if target.Cmp(prevTarget) <= 0 {
			return refusal("rung target percent", fmt.Sprintf(
				"rung target_pct must be strictly increasing; %s <= %s", rung.TargetPct, prev.TargetPct))
		}
		stop, _ := parseRat("stop", rung.StopPct)
		prevStop, _ := parseRat("stop", prev.StopPct)
		if stop.Cmp(prevStop) < 0 {
			return refusal("rung stop percent", fmt.Sprintf(
				"rung stop_pct must be monotonically non-decreasing; %s < %s", rung.StopPct, prev.StopPct))
		}
	}
	if p.RunnerTrailPct != "" {
		trail, err := parseRat("runner trail percent", p.RunnerTrailPct)
		if err != nil {
			return refusal("runner trail percent", err.Error())
		}
		if trail.Sign() <= 0 || trail.Cmp(big.NewRat(100, 1)) >= 0 {
			return refusal("runner trail percent", "runner trail percent must be in (0, 100)")
		}
	}
	return nil
}

// LadderState is the per-position state, the columns of `exit_states` that the
// ladder reads.
//
// The split the original documents (profit_ladder.py:26-34) is preserved and is
// the reason [EvaluateLadder] is careful about which fields it touches:
// ActivatedRung moves at **decision** time; TakenRatioTotal and Completed move
// at **execution** time, in the fill transaction's apply hook, and an evaluation
// copies them through untouched.
type LadderState struct {
	PolicyID string
	// ActivatedRung is the highest rung reached, or NoRung.
	ActivatedRung int
	// TakenRatioTotal is the cumulative taken fraction of the initial quantity.
	TakenRatioTotal string
	// Completed is set when the position's ladder is finished.
	Completed bool
	// PendingAction and PendingRung are the outstanding proposal, or ActionNone
	// and NoRung.
	PendingAction Action
	PendingRung   int
}

// LadderInput is one observation and the state it is judged against.
type LadderInput struct {
	EntryPrice    string
	ObservedPrice string
	// HighWater is the watermark as it stood before this observation.
	HighWater string
	// Baseline is the protected baseline as it stood before this observation.
	// Required for the same reason the ratchet's is: it exists from t0.
	Baseline string
	State    LadderState
	Policy   LadderPolicy
}

// LadderTransition is what one observation implies.
type LadderTransition struct {
	HighWater string
	// ReturnPct is the watermark's gross return in percent, for the audit trail.
	ReturnPct string
	NextState LadderState
	// Baseline is the composed protected baseline, never below the input's.
	Baseline      string
	Raised        bool
	WinningReason string
	Action        Action
	Reason        string
	// RungPromotedTo is the rung reached *this* evaluation, or NoRung. It is not
	// the cumulative maximum — that is NextState.ActivatedRung
	// (profit_ladder.py:143-146).
	RungPromotedTo     int
	Proposal           Proposal
	Suppressed         string
	CancelPendingFirst bool
}

// OpenLadder is the t0 state of a position under the ladder policy.
type OpenLadder struct {
	Baseline    string
	HighWater   string
	InitialRisk string
	State       LadderState
}

// OpenLadderState computes the state a ladder position opens with.
//
// The baseline is the entry decision's stop — the same t0 rule as the ratchet,
// and the reason a ladder position is protected before its first rung, which in
// the original it is not.
func OpenLadderState(entry, initialStop string, policy LadderPolicy) (OpenLadder, error) {
	if err := policy.Validate(); err != nil {
		return OpenLadder{}, err
	}
	e, s, risk, err := riskOf(entry, initialStop)
	if err != nil {
		return OpenLadder{}, err
	}
	return OpenLadder{
		Baseline:    formatPrice(s),
		HighWater:   formatPrice(e),
		InitialRisk: formatPrice(risk),
		State: LadderState{
			PolicyID:        policy.PolicyID,
			ActivatedRung:   NoRung,
			TakenRatioTotal: "0",
			PendingAction:   ActionNone,
			PendingRung:     NoRung,
		},
	}, nil
}

// EvaluateLadder judges one observation against the rung table.
func EvaluateLadder(in LadderInput) (LadderTransition, error) {
	if err := in.Policy.Validate(); err != nil {
		return LadderTransition{}, err
	}
	if in.State.PolicyID != in.Policy.PolicyID {
		// profit_ladder.py:216-220. A state hydrated under one rung set and judged
		// against another would compare a percent to the wrong table.
		return LadderTransition{}, refusal("ladder policy", fmt.Sprintf(
			"ladder state/policy mismatch: state=%s policy=%s", in.State.PolicyID, in.Policy.PolicyID))
	}
	entry, err := positive("entry price", in.EntryPrice)
	if err != nil {
		return LadderTransition{}, err
	}
	observed, err := positive("observed price", in.ObservedPrice)
	if err != nil {
		return LadderTransition{}, err
	}
	previousHigh, err := positive("high water", in.HighWater)
	if err != nil {
		return LadderTransition{}, err
	}
	previousBaseline, err := positive("baseline", in.Baseline)
	if err != nil {
		return LadderTransition{}, err
	}
	if _, err := fraction("taken ratio total", in.State.TakenRatioTotal); err != nil {
		return LadderTransition{}, err
	}
	if in.State.ActivatedRung < NoRung || in.State.ActivatedRung >= len(in.Policy.Rungs) {
		return LadderTransition{}, refusal("active rung", fmt.Sprintf(
			"%d is not a rung of a %d-rung policy", in.State.ActivatedRung, len(in.Policy.Rungs)))
	}

	// 1. The watermark, and the return it represents.
	probe := previousHigh
	if observed.Cmp(probe) > 0 {
		probe = observed
	}
	returnPct := percentOf(probe, entry)

	out := LadderTransition{
		HighWater:      formatPrice(probe),
		ReturnPct:      formatRMultiple(returnPct),
		NextState:      in.State,
		Baseline:       formatPrice(previousBaseline),
		WinningReason:  CandidatePrevious,
		Action:         ActionNone,
		Reason:         ReasonNoRungPromotion,
		RungPromotedTo: NoRung,
	}

	// 2. The highest rung the watermark has reached
	//    (`_highest_reached_rung_index`, profit_ladder.py:250-259).
	newIndex := in.State.ActivatedRung
	for i, rung := range in.Policy.Rungs {
		target, err := parseRat("rung target percent", rung.TargetPct)
		if err != nil {
			return LadderTransition{}, refusal("rung target percent", err.Error())
		}
		if i > newIndex && returnPct.Cmp(target) >= 0 {
			newIndex = i
		}
	}

	// 3. The baseline. Composed through the same R4 module the ratchet uses, so a
	//    ladder position's protection and a ratchet position's obey one rule.
	if newIndex > NoRung {
		lock, err := lockPrice(entry, in.Policy.Rungs[newIndex].StopPct)
		if err != nil {
			return LadderTransition{}, err
		}
		candidates := BuildCandidates(in.Baseline, "", lock)
		if newIndex == len(in.Policy.Rungs)-1 && in.Policy.RunnerTrailPct != "" {
			trailPct, err := parseRat("runner trail percent", in.Policy.RunnerTrailPct)
			if err != nil {
				return LadderTransition{}, refusal("runner trail percent", err.Error())
			}
			multiplier := new(big.Rat).Sub(one, new(big.Rat).Quo(trailPct, big.NewRat(100, 1)))
			trail := new(big.Rat).Mul(probe, multiplier)
			candidates = append(candidates, StopCandidate{
				Name: CandidateHighWaterRunner, Price: formatPrice(trail),
			})
		}
		composed, err := ComputeProtectedStop(candidates, in.Baseline)
		if err != nil {
			return LadderTransition{}, err
		}
		price, err := parseRat("composed baseline", composed.Price)
		if err != nil {
			return LadderTransition{}, refusal("composed baseline", err.Error())
		}
		out.Baseline = composed.Price
		out.WinningReason = composed.WinningReason
		out.Raised = price.Cmp(previousBaseline) > 0
	}

	// 4. The state promotion. It happens whether or not an order is proposed —
	//    the original's decision-time field — and it happens even while a
	//    proposal is unresolved, because protection is not what is being
	//    de-duplicated.
	if newIndex > in.State.ActivatedRung {
		out.NextState.ActivatedRung = newIndex
		out.RungPromotedTo = newIndex
	}

	// A finished ladder judges nothing further. Completed moves at execution
	// time, so this is the state a fill left behind.
	if in.State.Completed {
		out.Action, out.Reason, out.RungPromotedTo = ActionNone, ReasonCompleted, NoRung
		out.NextState.ActivatedRung = in.State.ActivatedRung
		return out, nil
	}

	// 5. The breach, against the baseline this evaluation just composed
	//    (`_check_protected_stop`, profit_ladder.py:239-246).
	baseline, err := parseRat("baseline", out.Baseline)
	if err != nil {
		return LadderTransition{}, refusal("baseline", err.Error())
	}
	if observed.Cmp(baseline) < 0 {
		out.Reason = ReasonStopBreached
		if in.State.PendingAction == ActionLadderStop {
			out.Suppressed = SuppressedPending
			return out, nil
		}
		out.Action = ActionLadderStop
		out.Proposal = Proposal{Action: ActionLadderStop, Level: rungLabel(newIndex), Ratio: "1"}
		out.CancelPendingFirst = in.State.PendingAction != ActionNone
		return out, nil
	}

	// 6. The promotion's own outcome (`_build_promotion_transition`,
	//    profit_ladder.py:262-292).
	if out.RungPromotedTo == NoRung {
		return out, nil
	}
	rung := in.Policy.Rungs[newIndex]
	switch {
	case newIndex == len(in.Policy.Rungs)-1 && in.Policy.FinalTakeFull:
		out.Action, out.Reason = ActionLadderTakeProfit, ReasonFinalRungTakeAll
		out.Proposal = Proposal{Action: ActionLadderTakeProfit, Level: rungLabel(newIndex), Ratio: "1"}
	case isPositive(rung.PartialRatio):
		out.Action, out.Reason = ActionLadderPartial, ReasonRungPartialTake
		out.Proposal = Proposal{
			Action: ActionLadderPartial, Level: rungLabel(newIndex), Ratio: rung.PartialRatio,
		}
	default:
		// State-only: the rung raised the lock and asks for no order.
		out.Action, out.Reason = ActionLadderHoldStopPromoted, ReasonStateOnly
		return out, nil
	}

	// 7. The pending suppression. Applied last, so the state promotion above has
	//    already happened — the original's LADDER_HOLD_STOP_PROMOTED with a
	//    quantity of zero.
	if in.State.PendingAction != ActionNone {
		out.Action, out.Reason = ActionLadderHoldStopPromoted, ReasonBlockedByPending
		out.Proposal = Proposal{}
		out.Suppressed = SuppressedPending
	}
	return out, nil
}

// Changed reports whether this evaluation moved anything worth an `exit_events`
// row (task 7.4).
func (t LadderTransition) Changed(previousHighWater string) bool {
	return t.Raised || t.HighWater != previousHighWater || t.RungPromotedTo != NoRung ||
		!t.Proposal.Zero()
}

// LockPrice is the protected stop one rung implies: entry × (1 + pct ÷ 100).
//
// Exported because the resolution path needs it too: rolling a refused rung back
// must not roll the baseline back with it, and the caller checking that needs the
// same arithmetic rather than its own.
func LockPrice(entryPrice, stopPct string) (string, error) {
	entry, err := positive("entry price", entryPrice)
	if err != nil {
		return "", err
	}
	return lockPrice(entry, stopPct)
}

func lockPrice(entry *big.Rat, stopPct string) (string, error) {
	pct, err := parseRat("rung stop percent", stopPct)
	if err != nil {
		return "", refusal("rung stop percent", err.Error())
	}
	factor := new(big.Rat).Add(one, new(big.Rat).Quo(pct, hundred))
	return formatPrice(new(big.Rat).Mul(entry, factor)), nil
}

// percentOf is the original's `gross_return_pct` (profit_ladder.py:137-139),
// computed exactly.
func percentOf(price, entry *big.Rat) *big.Rat {
	return new(big.Rat).Mul(new(big.Rat).Quo(new(big.Rat).Sub(price, entry), entry), hundred)
}

func isPositive(value string) bool {
	r, err := parseRatio("value", value)
	return err == nil && r.Sign() > 0
}

// rungLabel renders a rung index for `exit_states.pending_level`, the one column
// that carries a ratchet level under RATCHET and a rung index under LADDER
// (D7 — one column, because a position has one policy).
func rungLabel(index int) string { return strconv.Itoa(index) }

// RungIndex reads a rung label back. The error is a refusal because a
// `pending_level` that is neither a level nor an index is a corrupted row, not a
// state to guess at.
func RungIndex(label string) (int, error) {
	n, err := strconv.Atoi(label)
	if err != nil {
		return NoRung, refusal("pending level", fmt.Sprintf("%q is not a rung index", label))
	}
	if n < 0 {
		return NoRung, refusal("pending level", fmt.Sprintf("%q is not a rung index", label))
	}
	return n, nil
}
