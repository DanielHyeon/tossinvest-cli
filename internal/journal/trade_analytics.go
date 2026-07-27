package journal

// trade_analytics.go splits the frozen outcomes by what justified the position
// (change adopt-external-positions task 2.6; trade-analytics "합성 R의 구분
// 집계").
//
// # Why the split is mandatory rather than nice to have
//
// Realised R is `pnl ÷ (initial_risk × initial_quantity)`, and after this change
// the denominator comes from two different kinds of number:
//
//	engine-entered   initial_risk = entry − the stop a Guardian sized against.
//	                 A *measured* risk: somebody decided to take it.
//	adopted          initial_risk = observation × default_stop_pct. A *synthetic*
//	                 risk: nobody decided anything, a configured fraction did.
//
// Averaging the two produces a number with no meaning. Changing
// `adoption.default_stop_pct` from 5 % to 10 % halves every adopted R without a
// single trade behaving differently, so a blended ΣR would move because a
// setting moved. trade-analytics therefore requires the two populations to be
// reported apart (SHALL), and requires a mixed figure to carry both sample
// counts (SHALL) — because the honest reading of a blended number is "these n
// are measured and these m are not".
//
// # The join is explicit
//
// `positions.adoption_id IS NOT NULL`, a declared reference column. No time
// window, no heuristic, and no new column on `trade_outcomes`: the outcome row
// is frozen and this classification is derived on read, which is the same rule
// the aggregates themselves follow.

import (
	"context"
	"fmt"
	"strings"
)

// Outcome sources. The strings are part of the answer an export renders, so they
// are appended to rather than renamed.
const (
	// OutcomeSourceEntry is a position the engine opened. Its R is measured.
	OutcomeSourceEntry = "ENTRY"
	// OutcomeSourceAdoption is a position the engine adopted. Its R is synthetic.
	OutcomeSourceAdoption = "ADOPTION"
)

// ClassifiedOutcomes is one account's frozen outcomes, split by source.
type ClassifiedOutcomes struct {
	// Entered are the engine's own round trips (measured R).
	Entered []TradeOutcome
	// Adopted are the adopted positions' round trips (synthetic R).
	Adopted []TradeOutcome
}

// ClassifiedTradeOutcomes returns one account's frozen outcomes split by what
// justified the position, oldest close first within each population.
func (j *Journal) ClassifiedTradeOutcomes(ctx context.Context, accountRef string) (
	ClassifiedOutcomes, error) {
	var out ClassifiedOutcomes
	rows, err := j.db.QueryContext(ctx, `
		SELECT o.position_id, o.realized_pnl_after_costs, o.realized_r, o.initial_risk,
		       o.initial_quantity, coalesce(o.held_seconds, 0),
		       coalesce(o.exit_ratchet_level, ''), o.exit_rung, o.closed_at,
		       p.adoption_id IS NOT NULL
		  FROM trade_outcomes o
		  JOIN positions p ON p.id = o.position_id
		 WHERE p.account_ref = ?
		 ORDER BY o.closed_at, o.rowid`, strings.TrimSpace(accountRef))
	if err != nil {
		return ClassifiedOutcomes{}, fmt.Errorf(
			"journal: listing the classified outcomes of %s: %w", accountRef, err)
	}
	defer rows.Close()

	for rows.Next() {
		outcome, adopted, err := scanClassifiedOutcome(rows)
		if err != nil {
			return ClassifiedOutcomes{}, err
		}
		if adopted {
			out.Adopted = append(out.Adopted, outcome)
			continue
		}
		out.Entered = append(out.Entered, outcome)
	}
	if err := rows.Err(); err != nil {
		return ClassifiedOutcomes{}, fmt.Errorf(
			"journal: listing the classified outcomes of %s: %w", accountRef, err)
	}
	return out, nil
}

func scanClassifiedOutcome(rows rowScanner) (TradeOutcome, bool, error) {
	var (
		o       TradeOutcome
		rung    any
		adopted int
	)
	if err := rows.Scan(&o.PositionID, &o.RealizedPnLAfterCosts, &o.RealizedR,
		&o.InitialRisk, &o.InitialQuantity, &o.HeldSeconds, &o.ExitRatchetLevel,
		&rung, &o.ClosedAt, &adopted); err != nil {
		return TradeOutcome{}, false, fmt.Errorf("journal: reading a classified outcome: %w", err)
	}
	o.ExitRung = -1
	if v, ok := rung.(int64); ok {
		o.ExitRung = int(v)
	}
	return o, adopted != 0, nil
}

// SegmentedAggregates is the summary trade-analytics requires: the two
// populations apart, and the blended figure *only* alongside both sample counts.
//
// The counts are fields of the same struct as Combined on purpose. A caller
// cannot obtain the mixed number without also holding the two n's, so "혼합
// 집계를 낼 때는 두 모집단의 표본 수를 병기한다" is a property of the type rather
// than of whoever renders it.
type SegmentedAggregates struct {
	// Entered is the measured-R population; Adopted the synthetic-R one.
	Entered TradeAggregates
	Adopted TradeAggregates
	// Combined blends both. It is meaningful only read together with the counts
	// below — see Mixed and Note.
	Combined TradeAggregates
}

// Mixed reports that both populations are non-empty, which is exactly when a
// blended figure needs its caveat.
func (s SegmentedAggregates) Mixed() bool {
	return s.Entered.Trades > 0 && s.Adopted.Trades > 0
}

// Note is the sample-count caveat, ready to render beside a blended figure. It
// is empty when there is nothing blended.
func (s SegmentedAggregates) Note() string {
	if !s.Mixed() {
		return ""
	}
	return fmt.Sprintf(
		"실측 R %d건 + 합성 R %d건 혼합 — 합성 R의 분모는 편입 시점 관측가 × default_stop_pct이므로 "+
			"두 모집단의 R은 같은 단위가 아니다",
		s.Entered.Trades, s.Adopted.Trades)
}

// AggregateBySource derives the three aggregates from a classified set.
func AggregateBySource(c ClassifiedOutcomes) SegmentedAggregates {
	all := make([]TradeOutcome, 0, len(c.Entered)+len(c.Adopted))
	all = append(all, c.Entered...)
	all = append(all, c.Adopted...)
	return SegmentedAggregates{
		Entered:  AggregateTradeOutcomes(c.Entered),
		Adopted:  AggregateTradeOutcomes(c.Adopted),
		Combined: AggregateTradeOutcomes(all),
	}
}
