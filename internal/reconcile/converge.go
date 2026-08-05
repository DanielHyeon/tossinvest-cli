package reconcile

// converge.go folds a *quantity mismatch* back onto the account's own number
// (add-core-domain task 7.5 succession; issues.md, task 6.3 "운영상 대가").
//
// # Why this file has to exist
//
// Task 6.3 made the automatic release of a symbol block conditional on evidence:
// a block lifts on "the re-read after an adjustment agreed", never on a
// coincidence that agreed by itself (SHALL/SHALL NOT). That is the right rule and
// it has a price, which 6.3 named and left for this task:
//
//	루프의 정상 경로는 수량 불일치를 조정으로 수렴시키므로(계좌가 권위) 항상
//	무언가를 쓴다. 4차 대사 루프가 수량 불일치에 대해서도 조정을 발행하지 않으면
//	이 대가가 일상화된다.
//
// The Ingestor covers holdings the projection has never heard of. This covers
// the other half — a symbol the engine *does* hold, in a quantity the account
// disagrees with — and without it `ADJUSTMENT_APPLIED` is unreachable on the
// path it was designed for, so every ordinary quantity disagreement becomes an
// operator ticket.
//
// # The account is the authority, and that is the whole rule
//
// There is no arithmetic here. `QuantityMismatch.Authority()` is the broker's
// number, this writes it, and the projection converges to it. What the
// adjustment carries is the *evidence*: the value it was computed against
// (compare-and-append), the fill watermark it was computed at, and the broker
// snapshot's as-of. Any of those moving means the difference was computed
// against a world that no longer exists, and the commit discards it.
//
// # What it refuses rather than guesses
//
//   - A symbol whose local quantity is spread over instances in more than one
//     market. The comparison is symbol-level (`[미측정 — 보유 스냅샷의 market
//     차원]`) and the adjustment is instance-level, so there is no non-arbitrary
//     way to say which venue moved. Refusing leaves the block up, which is the
//     conservative direction.
//   - A symbol with no non-CLOSED instance at all. That is the Ingestor's case,
//     not this one, and folding it in here would open a second writer for the
//     same event.
//
// # Provenance
//
// `UNKNOWN`, from internal/position.Classify: the engine has an instance, so it
// is not external, and nobody declared it, so it is not manual. "The journal does
// not explain this" is the honest name for it.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// ConvergeStore is the journal half of a convergence.
//
// It is AdjustmentStore plus the projection read, because unlike an external
// holding a mismatch has to be resolved onto an instance that already exists.
// *journal.Journal satisfies it.
type ConvergeStore interface {
	AdjustmentStore
	Positions(ctx context.Context, accountRef string) ([]journal.Position, error)
}

// AdjustmentCrediter is told which comparison an adjustment was computed from
// and which symbols it was applied to, which is what makes a *later* re-read
// able to release their block (Tracker.AdjustmentApplied). *Tracker satisfies it.
//
// The comparison is not optional and not inferable. The reconciliation driver
// converges a diff and then observes that same diff in the same cycle, so a
// crediter that only learns "some adjustment happened" cannot tell the re-read
// from the read it is answering — which is how the automatic release stayed
// unreachable in production (a083).
type AdjustmentCrediter interface {
	AdjustmentApplied(comparison string, symbols ...string)
}

// ConvergedPosition is one mismatch this pass settled.
type ConvergedPosition struct {
	Market string
	Symbol string
	// Local is what the projection held; Quantity is what it holds now, which is
	// the account's own number.
	Local    string
	Quantity string
	// PositionID is the instance the adjustment landed on.
	PositionID string
	// Applied is false when this exact adjustment was already on disk — a retry
	// after a crash, or a later pass recomputing the same difference.
	Applied bool
	// ClosedExitState reports that converging to the account's number took a
	// position the engine was protecting to zero, so its exit state was completed
	// with an ADJUSTMENT_CLOSED event and no outcome was frozen (design A7).
	ClosedExitState bool
}

// ConvergeReport is what one pass did.
type ConvergeReport struct {
	// Converged lists the mismatches that reached the projection.
	Converged []ConvergedPosition
	// Refused lists the ones this pass would have had to guess at, with why.
	// They are reported rather than returned as an error: one unattributable
	// symbol must not stop the others from converging.
	Refused map[string]string
	// Credited is the symbols handed to the crediter, sorted.
	Credited []string
	// Closed counts the managed positions this pass found had gone to zero
	// outside the engine.
	Closed int
}

// Converger folds quantity mismatches onto the account's numbers.
type Converger struct {
	// Journal records the adjustment and supplies the projection. Required.
	Journal ConvergeStore
	// Credit is told which symbols were adjusted, so the next re-read can
	// release their block with ADJUSTMENT_APPLIED. Optional only so the
	// convergence can be unit-tested; an engine without one converges the
	// projection and leaves the block up until an operator clears it.
	Credit AdjustmentCrediter
	// Alert receives the notification that a managed position was closed outside
	// the engine. Optional only so the convergence can be unit-tested; an engine
	// without one completes the exit state silently, which leaves an operator to
	// discover from an empty screen that the engine stopped protecting something.
	Alert Alerter
	// AccountRef scopes the adjustment. Falls back to the diff's when empty.
	AccountRef string
}

// ConvergeQuantities applies one adjustment per quantity mismatch.
//
// A stale adjustment stops the pass and wraps journal.ErrAdjustmentStale, for
// the same reason the external ingest does: the rest of the diff was computed
// from the same view, and applying it after that view has been shown to have
// moved would be writing conclusions from a snapshot the ledger has already
// rejected. What committed before it stays in the report.
func (c *Converger) ConvergeQuantities(ctx context.Context, diff Diff) (ConvergeReport, error) {
	report := ConvergeReport{Refused: map[string]string{}}
	if len(diff.Quantities) == 0 {
		return report, nil
	}
	if c == nil || c.Journal == nil {
		return report, errors.New(
			"reconcile: converging a quantity mismatch needs the journal that records the adjustment")
	}
	account := strings.TrimSpace(firstNonEmpty(c.AccountRef, diff.AccountRef))
	if account == "" {
		return report, errors.New("reconcile: a convergence is scoped to an account; none was named")
	}
	asOf := strings.TrimSpace(diff.AsOf)
	if asOf == "" {
		return report, fmt.Errorf(
			"reconcile: the comparison carries no as-of, so an adjustment from it could not be ordered "+
				"against the fills it competes with (account %s)", account)
	}

	instances, err := c.Journal.Positions(ctx, account)
	if err != nil {
		return report, err
	}

	var (
		credited  []string
		alertErrs []error
	)
	for _, mismatch := range diff.Quantities {
		symbol := strings.ToUpper(strings.TrimSpace(mismatch.Symbol))
		market, reason := soleLiveMarket(instances, symbol)
		if reason != "" {
			report.Refused[symbol] = reason
			continue
		}

		watermark, err := c.Journal.FillWatermark(ctx, symbol)
		if err != nil {
			return report, err
		}
		result, err := c.Journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
			AccountRef: account,
			Market:     market,
			Symbol:     symbol,
			Kind: string(position.Classify(position.ProvenanceInputs{
				LocalInstance: true,
			})),
			ExpectedPrevQuantity:  mismatch.Local,
			ExpectedFillWatermark: watermark,
			NewQuantity:           mismatch.Authority(),
			// The account's cost basis is not part of a quantity comparison, and
			// "" leaves the projected basis alone rather than erasing it.
			NewAvgPrice: "",
			BrokerAsOf:  asOf,
			Evidence: fmt.Sprintf(
				"the account reports %s of %s and the projection held %s; the account is the authority "+
					"and the projection is converged to it",
				mismatch.Broker, symbol, mismatch.Local),
		})
		if errors.Is(err, journal.ErrAdjustmentStale) {
			return report, fmt.Errorf(
				"reconcile: the quantity mismatch on %s was computed against a view that has moved; "+
					"re-collect the snapshot: %w", symbol, err)
		}
		if err != nil {
			return report, fmt.Errorf("reconcile: converging the quantity of %s: %w", symbol, err)
		}

		report.Converged = append(report.Converged, ConvergedPosition{
			Market: market, Symbol: symbol, Local: mismatch.Local,
			Quantity: result.Position.Quantity, PositionID: result.Position.ID,
			Applied: result.Applied, ClosedExitState: result.ClosedExitState,
		})
		credited = append(credited, symbol)

		if result.ClosedExitState {
			report.Closed++
			// A position the engine was protecting went to zero somewhere the
			// engine cannot see. The exit state is completed and no outcome was
			// frozen (design A7), and both of those are facts an operator has to be
			// told rather than left to infer from a screen that quietly stops
			// listing the position.
			if c.Alert == nil {
				continue
			}
			if err := c.Alert.ManagedPositionClosedExternally(ctx, ManagedCloseAlert{
				AccountRef: account, Market: market, Symbol: symbol,
				PositionID:   result.Position.ID,
				PrevQuantity: mismatch.Local,
				BrokerAsOf:   asOf,
				Adopted:      result.Position.Adopted(),
			}); err != nil {
				alertErrs = append(alertErrs, fmt.Errorf(
					"reconcile: alerting that the managed position in %s was closed outside the engine: %w",
					symbol, err))
			}
		}
	}

	if len(credited) > 0 {
		sort.Strings(credited)
		report.Credited = credited
		if c.Credit != nil {
			// The credit carries the comparison it was computed from, so the tracker
			// can tell a re-read from the read this adjustment is answering. The
			// driver observes *this* diff later in the same cycle; only the next
			// cycle's comparison is the "re-read after the adjustment" the release
			// rule means, and the stamp is what says so (a083).
			//
			// A re-applied adjustment (Applied=false) is credited too: the projection
			// is converged either way, and what the release rule requires is that
			// something was written for this symbol before the re-read — not that
			// this process was the one that wrote it.
			c.Credit.AdjustmentApplied(asOf, credited...)
		}
	}
	// The adjustments committed either way, so the report stands. An undelivered
	// alert is still a failure — the operator does not know the engine stopped
	// protecting a position — so it is returned once everything that could be
	// converged has been, the same rule the ingest follows.
	return report, errors.Join(alertErrs...)
}

// soleLiveMarket returns the one venue a symbol's live quantity sits on, or the
// reason there is not exactly one.
func soleLiveMarket(instances []journal.Position, symbol string) (string, string) {
	markets := map[string]bool{}
	for _, p := range instances {
		if strings.ToUpper(strings.TrimSpace(p.Symbol)) != symbol {
			continue
		}
		if p.State == journal.PositionClosed {
			continue
		}
		markets[strings.ToLower(strings.TrimSpace(p.Market))] = true
	}
	switch len(markets) {
	case 1:
		for m := range markets {
			return m, ""
		}
		return "", ""
	case 0:
		return "", "the projection holds no live instance of this symbol; a holding with no local " +
			"instance is the external-position path, not a quantity mismatch"
	default:
		names := make([]string, 0, len(markets))
		for m := range markets {
			names = append(names, m)
		}
		sort.Strings(names)
		return "", fmt.Sprintf(
			"the projection holds this symbol on %s and the comparison is at symbol level, so which "+
				"venue moved cannot be established [미측정 — 보유 스냅샷의 market 차원]",
			strings.Join(names, " and "))
	}
}
