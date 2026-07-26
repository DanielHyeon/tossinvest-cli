package exitpolicy

// candidate.go is the protected-stop composition: StockOS
// `packages/trading/stockos_trading/exit/protected_stop_candidate.py`, ported.
//
// The original exists because two call paths (the common ladder runner and the
// Parker adaptive exit) both need "what is the best protected stop right now",
// and the R4 invariant — the stop never decreases — has to be enforced in
// exactly one place. TossOS has the same requirement from the other direction:
// exit-policy specifies the composition literally (새 기준선 = max(이전 기준선,
// 레벨 후보, 레벨 ≥ BREAKEVEN일 때 실질 본전)), and a max written at each call
// site is a max that can disagree with itself.
//
// # What was ported and what was not
//
// The original builds nine candidates. Three are ported:
//
//	previous          — the R4 anchor (protected_stop_candidate.py:35)
//	real_breakeven    — the cost-inclusive break-even (…:38)
//	baseline_ratchet  — the level's own candidate (…:39)
//
// Six are not: `fixed_floor` and `r_multiple` are the Parker strategy's floors,
// and `vwap_lower`, `ema9_close`, `atr_trail` and `distribution_lock` are all
// derived from indicators. Both groups are signal-layer inputs, of which this
// change admits zero (P3). They are absent rather than stubbed, because a
// candidate that is always None is a max() term nobody can test.
//
// The tuple order of the three that remain is the original's relative order, so
// the tie-break is unchanged: when two candidates propose the same price the one
// earlier in the tuple wins, which is why a BREAKEVEN promotion — where the
// level candidate *is* the real break-even — is attributed to `real_breakeven`
// and not to `baseline_ratchet`.
//
// # The one behavioural difference
//
// The original accepts `previous_stop=None` and raises ValueError when no anchor
// exists at all (…:118). TossOS has no such state: `exit_states.baseline_price`
// is NOT NULL from t0 and its first value is the entry decision's stop (D5's
// first correction). The anchor is therefore required here, and its absence is a
// refusal of the inputs rather than a case the composition has to handle.

import (
	"fmt"
	"math/big"
)

// The candidate names. They are the original's strings, because they are written
// into `exit_events` and read by an operator comparing the two systems.
const (
	CandidatePrevious        = "previous"
	CandidateRealBreakeven   = "real_breakeven"
	CandidateBaselineRatchet = "baseline_ratchet"
)

// StopCandidate is one candidate price tagged by the reason it exists. An empty
// Price is the original's None: the candidate does not apply to this evaluation.
type StopCandidate struct {
	Name  string
	Price string
}

// Valid reports whether the candidate proposes a price.
func (c StopCandidate) Valid() bool { return c.Price != "" }

// StopComputation is the outcome of a composition.
type StopComputation struct {
	// Price is the composed baseline, never below Previous.
	Price string
	// WinningReason names the candidate that won the max, for the audit trail.
	WinningReason string
	// Rejected are the candidates that did not win, in the order they lost.
	Rejected []StopCandidate
}

// BuildCandidates converts the raw optionals into the tagged tuple, in the
// original's order.
func BuildCandidates(previous, realBreakeven, baselineRatchet string) []StopCandidate {
	return []StopCandidate{
		{Name: CandidatePrevious, Price: previous},
		{Name: CandidateRealBreakeven, Price: realBreakeven},
		{Name: CandidateBaselineRatchet, Price: baselineRatchet},
	}
}

// ComputeProtectedStop returns max(previous, valid candidates) with the reason
// tag, and never a price below previous.
//
// The R4 floor is applied after the max rather than folded into it, exactly as
// the original does, so that a candidate which lost *to the floor* is reported as
// rejected rather than silently dropped: an operator reading `winning_reason =
// previous` needs to be able to see what proposed something lower.
func ComputeProtectedStop(candidates []StopCandidate, previous string) (StopComputation, error) {
	prev, err := parseRat("previous baseline", previous)
	if err != nil {
		return StopComputation{}, refusal("previous baseline", err.Error())
	}

	var (
		winner      *StopCandidate
		winnerPrice *big.Rat
		rejected    []StopCandidate
	)
	for i := range candidates {
		c := candidates[i]
		if !c.Valid() {
			continue
		}
		price, err := parseRat(c.Name+" candidate", c.Price)
		if err != nil {
			return StopComputation{}, refusal(c.Name+" candidate", err.Error())
		}
		if winner == nil || price.Cmp(winnerPrice) > 0 {
			if winner != nil {
				rejected = append(rejected, *winner)
			}
			winner, winnerPrice = &c, price
			continue
		}
		rejected = append(rejected, c)
	}

	if winner == nil {
		// Only the anchor exists. The original returns it unchanged with the
		// `previous` tag; so does this.
		return StopComputation{Price: formatPrice(prev), WinningReason: CandidatePrevious}, nil
	}
	if winnerPrice.Cmp(prev) < 0 {
		return StopComputation{
			Price:         formatPrice(prev),
			WinningReason: CandidatePrevious,
			Rejected:      append(rejected, *winner),
		}, nil
	}
	return StopComputation{
		Price:         formatPrice(winnerPrice),
		WinningReason: winner.Name,
		Rejected:      rejected,
	}, nil
}

// refusal wraps a malformed input into the package's typed refusal.
func refusal(field, detail string) error {
	return &RefusalError{Field: field, Detail: detail}
}

// RefusalError is what an unusable input produces. It is a refusal to judge and
// not a judgement of "no change": exit-policy requires the observation loop to
// alert on it (판정 거부 + 알림), and a caller that could not tell a refusal from
// a quiet no-op would alert on neither.
type RefusalError struct {
	Field  string
	Detail string
}

func (e *RefusalError) Error() string {
	return fmt.Sprintf("exitpolicy: %s: %s", e.Field, e.Detail)
}

// Is makes every refusal match ErrRefused, so callers test the class rather than
// the field.
func (e *RefusalError) Is(target error) bool { return target == ErrRefused }
