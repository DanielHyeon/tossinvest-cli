package exitpolicy

// adoption.go is the arithmetic and the bounds of a *synthetic* t0 — the one an
// externally acquired position is adopted into exit management with (change
// adopt-external-positions, design A2).
//
// # Why the number is synthetic and what that costs
//
// Every other t0 in this package comes from an entry decision: the stop the
// Guardian sized against is the baseline, and R is measured in a risk somebody
// decided to take. An adopted position has no such record — the shares were
// bought by hand — so the baseline has to be manufactured from the only two
// things that are known at adoption time: the price observed at that instant and
// a configured fraction.
//
//	EntryPrice   = the observation taken immediately before the adoption
//	InitialStop  = EntryPrice × (1 − pct)
//
// The cost basis is deliberately not in either formula. Anchoring t0 on the
// original purchase price would put every long-held winner outside its own
// ±pct band on the first observation and liquidate it on the spot, which is what
// design A2 calls manage-forward and rejects the alternative to.
//
// # The lower bound is a safety rail, not a preference
//
// `MinStopPct` is 0.02. Below that the protective band is narrower than the
// scale of the things that move a price without meaning anything: observation
// noise between two five-second ticks, and the round-trip cost of the exit
// itself (internal/costs' documented ceiling is MAX_RATE = 0.05 per side). A
// protection band inside that scale is not protection, it is a device that
// liquidates on the first tick — so a configuration that asks for one is refused
// rather than honoured.
//
// The upper bound is 1: at pct = 1 the stop is zero and at pct > 1 it is
// negative, and neither is a price. `riskOf` refuses both anyway, which is the
// second half of the same rail: an extremely small pct whose stop rounds up to
// the entry price itself produces no risk per unit, and the open refuses it
// fail-closed rather than freezing a zero denominator.

import (
	"fmt"
	"math/big"
)

// MinStopPct and MaxStopPct bound `adoption.default_stop_pct`: 0.02 ≤ pct < 1.
// See the file header for why each end is where it is.
const (
	MinStopPct = 0.02
	MaxStopPct = 1.0
)

// ErrStopPctOutOfRange is the configuration refusal. It is a named error because
// the caller's answer to it is specified — the whole adoption feature stays off
// (exit-policy: "설정이 거부되고 편입은 전면 비활성으로 남는다") — and a caller
// that has to match on a message cannot implement that reliably.
type ErrStopPctOutOfRange struct{ Value float64 }

func (e *ErrStopPctOutOfRange) Error() string {
	return fmt.Sprintf(
		"exitpolicy: adoption stop fraction %v is outside [%v, %v): below %v the protective band is "+
			"narrower than observation noise and the round-trip cost, which makes it a device that "+
			"liquidates on the first tick rather than a stop",
		e.Value, MinStopPct, MaxStopPct, MinStopPct)
}

// ValidateStopPct reports whether a configured fraction may be used.
func ValidateStopPct(pct float64) error {
	if !(pct >= MinStopPct && pct < MaxStopPct) {
		return &ErrStopPctOutOfRange{Value: pct}
	}
	return nil
}

// SyntheticStop is `observed × (1 − pct)`, rendered by this package's protective
// rule.
//
// The rounding direction is formatPrice's: a protective price rounds up, so a
// stop that does not terminate in decimal lands a hair *above* where the exact
// arithmetic put it and therefore triggers marginally sooner. §0.9 permits
// movement in that direction and not the other.
//
// The fraction is a float64 because that is what a JSON configuration carries;
// it is converted to an exact rational before it touches a price, so the price
// arithmetic itself never sees a binary float. `[기존 제약 — 설정 값의 소스가
// JSON float64]`.
func SyntheticStop(observed string, pct float64) (string, error) {
	if err := ValidateStopPct(pct); err != nil {
		return "", err
	}
	price, err := positive("observed price", observed)
	if err != nil {
		return "", err
	}
	fraction := new(big.Rat).SetFloat64(pct)
	if fraction == nil {
		return "", refusal("adoption stop fraction", fmt.Sprintf("%v is not a finite fraction", pct))
	}
	stop := new(big.Rat).Mul(price, new(big.Rat).Sub(one, fraction))
	rendered := formatPrice(stop)

	// The rounded stop has to remain a stop. formatPrice rounds up, so a fraction
	// small enough for the whole band to vanish inside the rendering scale would
	// produce `stop == entry` — no risk per unit, an undefined R, and a position
	// whose first observation is already at its baseline. riskOf refuses it; this
	// asks first so the refusal names the fraction rather than the price.
	if _, _, _, err := riskOf(observed, rendered); err != nil {
		return "", err
	}
	return rendered, nil
}
