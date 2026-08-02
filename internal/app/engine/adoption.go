package engine

// adoption.go is the last step of the reconciliation cycle: deciding which of
// the account's holdings the engine takes into exit management, and doing it
// (change adopt-external-positions task 2.2; exit-policy "외부 취득 포지션의
// 자동 편입", design A2/A4).
//
// # The order of the gates, and why each one is where it is
//
// A holding is judged against these in order, and the order is the point: every
// gate above a transition state is a *finding* the operator is told about, and
// every transition state is silent because it will resolve by itself.
//
//	already managed     an entry decision or an adoption already justifies it.
//	                    Nothing to do; an adopted one is additionally checked for
//	                    an external increase.
//	RECONCILE           the engine and the account disagree about this symbol.
//	                    Transition state — silent. Adopting into a disagreement
//	                    would freeze a t0 against a quantity nobody agrees on.
//	stale snapshot      the stable view is older than the freshness bound.
//	                    Transition state — silent; the next cycle re-reads.
//	adoption off        a finding: the engine is trading beside a position it
//	                    will not protect, and §0.2 keeps that alert regardless of
//	                    the toggle (design A4).
//	excluded            a finding: the operator asked for this one to be left
//	                    alone, and it is still unprotected.
//	otherwise           a candidate.
//
// # The observation, and the fifteen seconds
//
// Candidates are priced in one batched read taken immediately before the
// adoption transactions (design A6: 후보 전체를 한 번의 배치 호출로). The
// observation's age is re-checked per candidate against the price staleness
// bound, and a candidate whose observation has gone stale between the read and
// its own transaction is deferred rather than adopted (exit-policy SHALL:
// staleness 15초 초과 시 편입을 연기한다).
//
// The reason is the whole of manage-forward. `synthetic_stop` is frozen from
// that observation, and a stop frozen from a price the market has already left
// behind is a stop the first exit observation crosses — the adoption would read
// as an instant liquidation of somebody's long-held shares.
//
// # What this file does not do
//
// It issues no sell proposal. The adoption transaction records the t0 and opens
// the exit state, and that is all (SHALL NOT — design A2). The first exit
// judgement happens in the exit observation loop, on its own schedule, against
// its own observation — and a proposal from *that* is normal exit behaviour, not
// an artefact of adopting.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// candidate is one holding that passed every gate.
type candidate struct {
	position journal.Position
	holding  reconcile.Holding
}

// judgeHoldings runs the gates over the stable snapshot's holdings and adopts
// what passes.
func (d *ReconcileDriver) judgeHoldings(ctx context.Context, snapshot reconcile.Snapshot,
	cycle *ReconcileCycle) {
	stale := d.opts.SnapshotStaleness
	if stale <= 0 {
		stale = DefaultAdoptionStaleness
	}
	fresh := snapshot.Age(d.clk.Now()) <= stale

	var (
		candidates []candidate
		unmanaged  []journal.Position
	)
	for _, holding := range snapshot.Holdings {
		market := strings.ToLower(strings.TrimSpace(holding.Market))
		if market == "" {
			market = strings.ToLower(strings.TrimSpace(d.opts.DefaultMarket))
		}
		symbol := strings.ToUpper(strings.TrimSpace(holding.Symbol))
		if symbol == "" || market == "" || isZeroQuantity(holding.Quantity) {
			continue
		}

		p, err := d.opts.Journal.CurrentPosition(ctx, d.opts.AccountRef, market, symbol)
		if err != nil {
			// No instance yet: the fold that creates one runs earlier in this same
			// cycle, so this is a holding whose fold failed or was refused. It is
			// already reported through that path and is not a second finding.
			continue
		}
		if p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) {
			continue
		}

		if p.ExitEligible() {
			if p.Adopted() {
				d.checkExternalIncrease(ctx, p)
			}
			continue
		}

		// --- transition states: silent -------------------------------------
		if d.blocked(market, symbol) {
			continue
		}
		if !fresh {
			continue
		}

		// --- findings ------------------------------------------------------
		// Exclusion is judged first: it wins over the global switch AND over a
		// per-symbol designation (exit-policy: exclude가 항상 우선 — 동시 등재는
		// 편입하지 않는다), so the alert can name the list the operator wrote.
		if d.opts.Adoption.Excludes(symbol) {
			unmanaged = append(unmanaged, p)
			continue
		}
		// A candidate is either globally admitted or individually designated
		// (change console-adoption-controls). The designated path runs through
		// the same gates above — nothing about RECONCILE, freshness or the
		// Stabiliser is relaxed for it.
		if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {
			unmanaged = append(unmanaged, p)
			continue
		}
		candidates = append(candidates, candidate{position: p, holding: holding})
	}

	adopted := d.adopt(ctx, candidates, cycle)
	for _, c := range candidates {
		if !adopted[c.position.ID] {
			// A candidate the engine could not adopt is a position it is not
			// protecting, which is the same finding as an excluded one
			// (exit-policy: 제외 목록 심볼·편입 실패는 알림이 남는다).
			unmanaged = append(unmanaged, c.position)
		}
	}

	for _, p := range unmanaged {
		cycle.Unmanaged++
		d.alertUnmanaged(ctx, p)
	}
}

// blocked reports that the symbol is under RECONCILE.
func (d *ReconcileDriver) blocked(market, symbol string) bool {
	if d.opts.Tracker == nil {
		return false
	}
	return d.opts.Tracker.Permanent() || d.opts.Tracker.EntryAllowed(market, symbol) != nil
}

// adopt prices the candidates in one batched read and adopts what it can.
//
// It returns the set of position ids that were adopted, so the caller can report
// the rest as unmanaged: a failed adoption is a position the engine is not
// protecting, and silence about it would be the alert regression design A4
// exists to prevent.
func (d *ReconcileDriver) adopt(ctx context.Context, candidates []candidate,
	cycle *ReconcileCycle) map[string]bool {
	adopted := map[string]bool{}
	if len(candidates) == 0 {
		return adopted
	}

	quotes, readAt, err := d.observeCandidates(ctx, candidates)
	if err != nil {
		cycle.Deferred += len(candidates)
		if cycle.Err == nil {
			cycle.Err = err
		}
		return adopted
	}

	bound := d.opts.PriceStaleness
	if bound <= 0 {
		bound = DefaultAdoptionPriceStaleness
	}
	for _, c := range candidates {
		key := adoptionQuoteKey(c.position.Market, c.position.Symbol)
		observed, ok := quotes[key]
		if !ok {
			// The read answered for other symbols and not this one. There is no
			// t0 to freeze, so the holding waits for the next cycle.
			cycle.Deferred++
			continue
		}
		if age := d.clk.Now().Sub(readAt); age > bound {
			// The observation went stale between the read and this transaction.
			// Freezing a synthetic stop from it would be freezing a stop the
			// market may already have passed.
			cycle.Deferred += len(candidates) - len(adopted)
			d.logDeferred(c.position, fmt.Sprintf(
				"the price observation is %s old, past the %s bound", age, bound))
			return adopted
		}
		if d.adoptOne(ctx, c, observed) {
			adopted[c.position.ID] = true
			cycle.Adopted++
			continue
		}
		cycle.Deferred++
	}
	return adopted
}

// observeCandidates is the one batched price read.
func (d *ReconcileDriver) observeCandidates(ctx context.Context, candidates []candidate) (
	map[string]string, time.Time, error) {
	marketsBySymbol := map[string]map[string]struct{}{}
	symbols := make([]string, 0, len(candidates))
	for _, c := range candidates {
		symbol := strings.ToUpper(strings.TrimSpace(c.position.Symbol))
		market := strings.ToLower(strings.TrimSpace(c.position.Market))
		if symbol == "" || market == "" {
			continue
		}
		markets := marketsBySymbol[symbol]
		if markets == nil {
			markets = map[string]struct{}{}
			marketsBySymbol[symbol] = markets
			symbols = append(symbols, symbol)
		}
		markets[market] = struct{}{}
	}
	sort.Strings(symbols)

	// Stamped before the request, not after. The quote the broker returns is at
	// best as fresh as the instant it was asked for, and a retried read can take
	// the whole query budget — so measuring the observation's age from the moment
	// the read *started* is the conservative reading of "the observation
	// immediately before the transaction". Measuring from the reply would let a
	// slow read hand back a price the market has already left behind and call it
	// zero seconds old.
	readAt := d.clk.Now()

	var quotes []domain.Quote
	read := func(ctx context.Context) error {
		var err error
		quotes, err = d.opts.Prices.Prices(ctx, symbols)
		return err
	}
	var err error
	if d.opts.Retrier != nil {
		err = d.opts.Retrier.Query(ctx, execgw.QueryPrice, read)
	} else {
		err = read(ctx)
	}
	if err != nil {
		return nil, time.Time{}, fmt.Errorf(
			"engine: observing %d adoption candidate(s): %w", len(symbols), err)
	}

	out := make(map[string]string, len(quotes))
	for _, q := range quotes {
		if q.Last <= 0 {
			// A quote with no last trade is not an observation. Recording it as
			// zero would freeze a synthetic stop of zero.
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(q.Symbol))
		markets := marketsBySymbol[symbol]
		if len(markets) != 1 {
			// A symbol-only quote cannot prove which candidate market it belongs
			// to when the same symbol is present in more than one market.
			continue
		}
		market := ""
		for candidateMarket := range markets {
			market = candidateMarket
		}
		expectedCurrency, ok := adoptionCurrency(market)
		if !ok || !strings.EqualFold(strings.TrimSpace(q.Currency), expectedCurrency) {
			continue
		}
		out[adoptionQuoteKey(market, symbol)] = decimalOf(q.Last)
	}
	return out, readAt, nil
}

func adoptionQuoteKey(market, symbol string) string {
	return strings.ToLower(strings.TrimSpace(market)) + "\x00" +
		strings.ToUpper(strings.TrimSpace(symbol))
}

func adoptionCurrency(market string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(market)) {
	case "kr":
		return "KRW", true
	case "us":
		return "USD", true
	default:
		return "", false
	}
}

// adoptOne persists one adoption and opens its exit state.
//
// The two are separate transactions on purpose, and the ordering is the crash
// contract (exit-policy: 편입 기록 영속 후 크래시). The adoption commits first;
// if the process dies before the exit state opens, the position is left carrying
// an adoption and no protection, which the *next* cycle — or the exit
// observation loop's own working set — completes from the record on disk. The
// other order would leave an exit state whose t0 nothing explains.
func (d *ReconcileDriver) adoptOne(ctx context.Context, c candidate, observed string) bool {
	stop, err := exitpolicy.SyntheticStop(observed, d.opts.Adoption.DefaultStopPct)
	if err != nil {
		d.logDeferred(c.position, "the synthetic stop could not be derived: "+err.Error())
		return false
	}

	adoption, err := d.opts.Journal.AdoptPosition(ctx, journal.AdoptionRequest{
		PositionID: c.position.ID,
		Symbol:     c.position.Symbol,
		Market:     c.position.Market,
		Quantity:   c.position.Quantity,
		// The broker's own string, never a re-rendering of the float beside it.
		// Empty when the snapshot could not preserve one, which the record stores
		// as ABSENT.
		CostBasis:     c.holding.CostBasisRaw,
		ObservedPrice: observed,
		SyntheticStop: stop,
		ObservedAt:    journal.RFC3339(d.clk.Now()),
		ExitPolicyID:  d.opts.CommonPolicy,
	})
	if err != nil {
		d.logDeferred(c.position, "the adoption was refused: "+err.Error())
		return false
	}

	if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != nil {
		// The record is committed, so the position *is* adopted and the next pass
		// completes the opening. It is reported as adopted rather than deferred
		// for exactly that reason: telling the operator it was not would be
		// wrong, and the alert says the protection is not open yet.
		d.logDeferred(c.position, "the adoption is recorded but the exit state is not open yet: "+
			err.Error())
	}

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionAdopted,
		Key:   string(obs.EventExitPositionAdopted) + "|" + c.position.ID,
		Title: c.position.Symbol + " was adopted into exit management",
		Body: "the account holds it and no entry decision explains it, so the engine recorded an " +
			"adoption: from now on the ratchet, the partial take-profit and the stop apply to it as " +
			"they do to a position the engine opened. The baseline is measured from the price observed " +
			"at adoption, not from what the shares originally cost — so a long-held winner is protected " +
			"from here forward and its historical gain is not part of the R scale",
		Fields: map[string]any{
			obs.FieldAccount:  d.opts.AccountRef,
			obs.FieldSymbol:   c.position.Symbol,
			obs.FieldQuantity: adoption.Quantity,
			"market":          c.position.Market,
			"position_id":     c.position.ID,
			"adoption_id":     adoption.ID,
			"observed_price":  adoption.ObservedPrice,
			"synthetic_stop":  adoption.SyntheticStop,
			"cost_basis":      adoption.CostBasis,
			"cost_basis_src":  adoption.CostBasisSource,
			"observed_at":     adoption.ObservedAt,
			"exit_policy_id":  adoption.ExitPolicyID,
		},
	})
	// A position that has just been adopted is no longer unmanaged; drop the
	// latch so a *later* loss of eligibility would alert again.
	delete(d.unmanaged, c.position.ID)
	return true
}

// alertUnmanaged reports a holding confirmed to be outside exit management.
//
// Latched per position: the loop runs every minute and an alert that repeats
// forever is an alert nobody reads. The latch is in memory, so a restart
// re-raises it — which is the right direction, because a restart is also when an
// operator is most likely to be looking.
//
// It fires regardless of `adoption.enabled` (SHALL — design A4). That toggle
// decides whether the engine *acts*; it does not decide whether the operator is
// told that the engine is trading beside a position it will not protect.
func (d *ReconcileDriver) alertUnmanaged(ctx context.Context, p journal.Position) {
	if d.unmanaged[p.ID] {
		return
	}
	d.unmanaged[p.ID] = true

	// The why-matrix (change console-adoption-controls design D2): every row is
	// a different actionable fact, and the one thing this switch may never do is
	// tell the owner of a tried-and-failed designation that "adoption is off" —
	// that sentence sends them to the wrong setting.
	why := "adoption is off and the symbol is not designated, so the engine records the holding " +
		"and leaves it alone"
	switch {
	case d.opts.Adoption.Rejected != "":
		why = "the adoption settings were refused, so adoption is off: " + d.opts.Adoption.Rejected
	case d.opts.Adoption.Excludes(p.Symbol):
		why = "the symbol is on adoption.exclude_symbols, so it is deliberately left unprotected"
	case d.opts.Adoption.Enabled:
		why = "adoption is on but this holding could not be adopted this cycle; it is unprotected " +
			"until it can be"
	case d.opts.Adoption.Included(p.Symbol):
		why = "the symbol is designated on adoption.include_symbols and the adoption was tried, but " +
			"this cycle failed; it is unprotected until a cycle succeeds"
	}

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|" + p.ID,
		Title: p.Symbol + " is held with no record justifying a stop, so the exit policy will not manage it",
		Body: "the position carries neither an entry decision nor an adoption record, so there is no " +
			"stop to build a baseline out of and no initial risk to measure R against — " + why,
		Fields: map[string]any{
			obs.FieldAccount:   d.opts.AccountRef,
			obs.FieldSymbol:    p.Symbol,
			obs.FieldQuantity:  p.Quantity,
			"market":           p.Market,
			"position_id":      p.ID,
			"adoption_enabled": d.opts.Adoption.Enabled,
		},
	})
}

// checkExternalIncrease reports an adopted position the owner has since bought
// more of (design A8).
//
// The t0 is deliberately not recomputed: `exit_states` freezes the entry, the
// initial risk and the initial quantity, and moving any of them would rewrite
// the denominator every R on that position has already been expressed in.
// Reporting it is what the operator can act on; re-sizing is a later change.
func (d *ReconcileDriver) checkExternalIncrease(ctx context.Context, p journal.Position) {
	if d.grown[p.ID] {
		return
	}
	adoption, err := d.opts.Journal.AdoptionOf(ctx, p.ID)
	if err != nil {
		return
	}
	cmp, err := riskcalc.CompareDecimal(p.Quantity, adoption.Quantity)
	if err != nil || cmp <= 0 {
		return
	}
	d.grown[p.ID] = true

	d.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|grown|" + p.ID,
		Title: p.Symbol + " grew after it was adopted, and the frozen t0 does not cover the increase",
		Body: "the account now holds more of this symbol than the adoption recorded. The exit state's " +
			"entry, initial risk and initial quantity stay frozen — recomputing them would rewrite the " +
			"denominator every R already reported for this position was measured in — so the extra " +
			"shares are protected by a stop sized for the original quantity",
		Fields: map[string]any{
			obs.FieldAccount:   d.opts.AccountRef,
			obs.FieldSymbol:    p.Symbol,
			obs.FieldQuantity:  p.Quantity,
			"market":           p.Market,
			"position_id":      p.ID,
			"adopted_quantity": adoption.Quantity,
			"adoption_id":      adoption.ID,
		},
	})
}

func (d *ReconcileDriver) logDeferred(p journal.Position, why string) {
	if d.opts.Log == nil {
		return
	}
	d.opts.Log.Event(obs.EventEngineHeartbeat,
		obs.FieldAccount, d.opts.AccountRef,
		obs.FieldSymbol, p.Symbol,
		"position_id", p.ID,
		obs.FieldDetail, "the adoption was deferred: "+why)
}
