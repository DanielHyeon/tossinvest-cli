package reconcile

// external.go folds the account's own holdings into the position projection
// (change add-core-domain task 6.3; reconciliation delta "외부 포지션의 투영
// 편입", design D4).
//
// # Why a position the engine never opened has to be in the projection
//
// The obvious answer is "so the ledger stops lying about the account", and it is
// not the operative one. The operative one is that the exit path reads the
// projection to decide *how much* to liquidate: 청산 수량 판정이 실제 보유를
// 알아야 한다(SHALL). A projection that omits three shares the account holds is
// a projection that will size a flatten for the wrong quantity, and the RECONCILE
// confirmed floor is computed against it.
//
// So the holding is folded in — as an adjustment event, because the account is
// the authority and an append-only event is how that authority reaches a row
// that nothing but a fill may otherwise write (position-ledger: Position 행 직접
// 덮어쓰기는 금지된다 SHALL NOT).
//
// # And why it must not become the exit policy's problem
//
// It is folded in with `entry_decision_id` NULL, and NULL is not a missing value
// here: it is the fact that no decision justifies the position. No decision means
// no entry stop, no entry stop means no t0 baseline, and D5's first correction is
// that the baseline *is* the entry stop. A ratchet with no baseline would have to
// invent one, and inventing a stop for somebody else's shares is a worse failure
// than not managing them. `ExitEligible` on the stored row is that rule, and this
// file refuses to fold a holding into an instance that already carries a decision
// rather than quietly widening what the exit policy manages.
//
// The exit loop itself is a later wave's; what this file owes it is that the rows
// exist and are marked.
//
// # And why a person is told
//
// 발견 시 알림을 발송한다(SHALL). The engine has just discovered trading it did
// not do on an account it is trading. The alert fires when the adjustment
// actually lands and not on every pass afterwards: the reconciliation loop runs
// every 30 seconds, and an alert that repeats forever is an alert nobody reads.
// The adjustment id is derived from what the adjustment is, so a re-collection
// that recomputes the same difference is recognised rather than re-announced.
//
// # Broker-behaviour claims
//
// One, tagged: whether the holdings snapshot carries a market dimension is
// `[미측정]` (design D4). A holding that names no market is therefore refused
// unless the caller declares the account's venue — the projection is keyed by
// market, and a guessed one puts the position on a venue an operator would then
// go looking for it on.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/position"
)

// AdjustmentStore is the journal half of the ingest.
//
// Two methods, and both halves of compare-and-append are visible in them: the
// watermark is read here and re-checked inside ApplyPositionAdjustment's own
// transaction. *journal.Journal satisfies it.
type AdjustmentStore interface {
	FillWatermark(ctx context.Context, symbol string) (int64, error)
	ApplyPositionAdjustment(ctx context.Context, req journal.AdjustmentRequest) (journal.AdjustmentResult, error)
}

// ExternalPositionAlert is what an externally acquired holding tells the
// operator.
type ExternalPositionAlert struct {
	AccountRef string
	Market     string
	Symbol     string
	// Quantity is the account's own — the value the projection converged to.
	Quantity string
	// PositionID is the instance the adjustment opened, so the operator can look
	// the whole thing up.
	PositionID string
	// BrokerAsOf is when the snapshot this came from was current.
	BrokerAsOf string
	// ExitEligible is the operationally important half of the alert: whether the
	// engine will protect the position it has just discovered.
	//
	// It used to be hardcoded false, which was true while an external holding
	// could never become managed. Since adopt-external-positions it is the
	// eligibility predicate's answer, because a re-reconciliation of an already
	// adopted holding must not tell an operator it is unprotected.
	ExitEligible bool
}

// ManagedCloseAlert is what a managed position going to zero outside the engine
// tells the operator (change adopt-external-positions, design A7).
//
// It is a separate event from a fold: the engine was protecting this position,
// the shares are gone, and no trade outcome was frozen because there is no sell
// leg the engine can price. All three of those are facts an operator would
// otherwise have to infer from a screen that quietly stopped listing a symbol.
type ManagedCloseAlert struct {
	AccountRef string
	Market     string
	Symbol     string
	// PositionID is the instance whose exit state was completed.
	PositionID string
	// PrevQuantity is what the projection held before the convergence.
	PrevQuantity string
	// BrokerAsOf is when the snapshot that showed the zero was current.
	BrokerAsOf string
	// Adopted reports that the closed position was one the engine had adopted
	// rather than one it opened.
	Adopted bool
}

// Alerter receives the operator alerts this package raises.
//
// It is an interface here rather than a direct dependency on internal/obs so
// that the reconciliation loop decides the grade and the transport, and so this
// package can be tested without one. The engine's wiring adapts obs.Notifier to
// it.
type Alerter interface {
	ExternalPositionFound(ctx context.Context, alert ExternalPositionAlert) error
	ManagedPositionClosedExternally(ctx context.Context, alert ManagedCloseAlert) error
}

// IngestedPosition is one holding the ingest folded in.
type IngestedPosition struct {
	Market string
	Symbol string
	// Quantity is the account's, which is what the projection now holds.
	Quantity string
	// PositionID is the instance the adjustment landed on.
	PositionID string
	// Applied is false when this exact adjustment was already on disk: a later
	// pass of the loop recomputing the same difference, or a retry after a crash.
	Applied bool
	// ExitEligible reports whether the exit policy may manage the instance. It is
	// the single predicate's answer, not a constant: false for a holding this
	// file has just folded in, and true for one a previous cycle adopted and this
	// one is re-reconciling.
	ExitEligible bool
	// Adopted reports that the eligibility above comes from an adoption record.
	// It is what tells the adoption judgement "this one is already managed" from
	// "this one is a candidate", without a second read of the projection.
	Adopted bool
}

// IngestReport is what one ingest did.
type IngestReport struct {
	// Folded is every holding that reached the projection, in the order the diff
	// listed them.
	Folded []IngestedPosition
	// Alerted is how many operator alerts were raised — one per newly folded
	// holding, none for one that was already on disk.
	Alerted int
}

// Ingestor folds external holdings into the projection.
type Ingestor struct {
	// Journal is where the adjustment is recorded. Required.
	Journal AdjustmentStore
	// Alert receives one notification per newly folded holding. Optional only so
	// the ingest can be unit-tested; an engine without one folds positions in
	// silently, which is the half of the requirement that is about a person.
	Alert Alerter
	// AccountRef scopes the adjustment. Falls back to the diff's when empty.
	AccountRef string
	// DefaultMarket is the venue a holding that names none is recorded under.
	// Empty means "refuse it", which is the honest answer when nothing in the
	// system knows where the position is.
	DefaultMarket string
}

// IngestExternalPositions folds every external holding in the diff into the
// projection, alerting on the ones that are new.
//
// A stale adjustment stops the ingest and is returned wrapping
// journal.ErrAdjustmentStale: the caller re-collects the snapshot rather than
// applying the rest of a view that has been shown to have moved. What was
// already folded is still in the report, because those adjustments committed.
func (in *Ingestor) IngestExternalPositions(ctx context.Context, diff Diff) (IngestReport, error) {
	var report IngestReport
	if len(diff.ExternalPos) == 0 {
		return report, nil
	}
	if in == nil || in.Journal == nil {
		return report, fmt.Errorf(
			"reconcile: folding an external holding in needs the journal that records the adjustment")
	}
	account := strings.TrimSpace(firstNonEmpty(in.AccountRef, diff.AccountRef))
	if account == "" {
		return report, fmt.Errorf("reconcile: an external-position ingest is scoped to an account; none was named")
	}
	asOf := strings.TrimSpace(diff.AsOf)
	if asOf == "" {
		return report, fmt.Errorf(
			"reconcile: the comparison carries no as-of, so an adjustment from it could not be "+
				"ordered against the fills it competes with (account %s)", account)
	}

	var alertErrs []error
	for _, external := range diff.ExternalPos {
		symbol := strings.ToUpper(strings.TrimSpace(external.Symbol))
		market := strings.ToLower(strings.TrimSpace(external.Market))
		if market == "" {
			market = strings.ToLower(strings.TrimSpace(in.DefaultMarket))
		}
		if symbol == "" || market == "" {
			return report, fmt.Errorf(
				"reconcile: the account holds %q of an unmarked symbol %q; the projection is keyed by "+
					"market and this one names none [미측정 — 보유 스냅샷의 market 차원]",
				external.Quantity, external.Symbol)
		}

		watermark, err := in.Journal.FillWatermark(ctx, symbol)
		if err != nil {
			return report, err
		}

		result, err := in.Journal.ApplyPositionAdjustment(ctx, journal.AdjustmentRequest{
			AccountRef: account,
			Market:     market,
			Symbol:     symbol,
			Kind:       string(position.KindExternal),
			// Zero by construction: a holding reaches ExternalPos precisely when the
			// projection has nothing non-CLOSED for that symbol. Carrying it
			// explicitly is what makes the commit re-check it.
			ExpectedPrevQuantity:  "0",
			ExpectedFillWatermark: watermark,
			NewQuantity:           external.Quantity,
			NewAvgPrice:           external.AveragePrice,
			BrokerAsOf:            asOf,
			Evidence: fmt.Sprintf(
				"the account holds %s of %s and no local instance explains it; folded in as an "+
					"external position with no entry decision, and therefore no exit policy",
				external.Quantity, symbol),
		})
		if errors.Is(err, journal.ErrAdjustmentStale) {
			return report, fmt.Errorf(
				"reconcile: the external holding of %s was computed against a view that has moved; "+
					"re-collect the snapshot: %w", symbol, err)
		}
		if err != nil {
			return report, fmt.Errorf("reconcile: folding in the external holding of %s: %w", symbol, err)
		}

		folded := IngestedPosition{
			Market:       market,
			Symbol:       symbol,
			Quantity:     result.Position.Quantity,
			PositionID:   result.Position.ID,
			Applied:      result.Applied,
			ExitEligible: result.Position.ExitEligible(),
			Adopted:      result.Position.Adopted(),
		}
		if strings.TrimSpace(result.Position.EntryDecisionID) != "" {
			// The instance the adjustment landed on carries an entry decision, so
			// this was not an external position at all — the comparison and the
			// projection disagree about whether the engine opened it. Widening what
			// the exit policy manages on the strength of that disagreement is the
			// one thing this path must not do.
			//
			// The test is `entry_decision_id` explicitly and NOT the eligibility
			// predicate (design A1). Since adopt-external-positions an eligible
			// position can also be one this engine adopted, and a fold landing on
			// *that* is the ordinary re-reconciliation path — the quantity
			// comparison a managed external position needs on every cycle. Guarding
			// on eligibility would refuse it and freeze the loop on exactly the
			// positions the change exists to manage.
			return report, fmt.Errorf(
				"reconcile: the holding of %s folded onto instance %s, which carries an entry decision; "+
					"an external position must not inherit one", symbol, folded.PositionID)
		}
		report.Folded = append(report.Folded, folded)

		if !folded.Applied || in.Alert == nil {
			continue
		}
		if err := in.Alert.ExternalPositionFound(ctx, ExternalPositionAlert{
			AccountRef: account, Market: market, Symbol: symbol,
			Quantity: folded.Quantity, PositionID: folded.PositionID,
			BrokerAsOf: asOf, ExitEligible: folded.ExitEligible,
		}); err != nil {
			// The adjustment is committed either way, so the fold is reported. An
			// undelivered alert is still a failure — the operator does not know the
			// engine is trading beside a position it will not protect — so it is
			// returned once everything that could be folded has been.
			alertErrs = append(alertErrs, fmt.Errorf(
				"reconcile: alerting on the external holding of %s: %w", symbol, err))
			continue
		}
		report.Alerted++
	}
	return report, errors.Join(alertErrs...)
}
