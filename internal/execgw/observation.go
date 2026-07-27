package execgw

// observation.go is the recording seam for change add-net-rr-measurement
// (tasks 4.1, 4.1b, 4.2, 4.5).
//
// # Where the seam is, and why it is where it is
//
// The Guardian is the only place both halves of an entry verdict exist. A chain
// refusal is built here and returned to the caller; today nothing writes it down,
// so half of every judgement the account makes leaves no trace at all. An issued
// verdict is written to `decisions`, but without the two ratios or the rate set
// they were computed under.
//
// So the recording goes here. Three constraints shape it, and all three come from
// the same instinct — the measurement must never be able to cost a trade:
//
//	outside the transaction  An observation inside RecordDecisionAndReserve would
//	                         roll the decision back when it failed. A full disk
//	                         would then refuse an entry the chain allowed, which is
//	                         a measurement defect stopping trading.
//	off the critical path    Recording synchronously before IssueEntry returns puts
//	                         a journal write in front of Gateway submission. A slow
//	                         write would eat the 60-second decision TTL and lose an
//	                         entry that was authorised. So the hand-off is
//	                         non-blocking and the write happens on the observer's
//	                         own goroutine.
//	never a verdict          A failed observation produces no reason code and
//	                         changes nothing the caller sees. It is counted, in a
//	                         store that does not share a failure domain with the one
//	                         that just failed, and alerted at SeverityNormal.
//
// # Entry only
//
// IssueReduction shares the evaluateChain seam and is deliberately not wired: an
// exit carries no stop and no target, so its observation would be a row whose
// every measured column is empty. observation_scope_test.go pins that, and pins
// that risk.Evaluate itself stays free of any journal dependency — a pure chain is
// what lets the counterfactual harness run it without a database.

import (
	"context"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

// EntryObserver receives one entry verdict's observation.
//
// Implementations MUST NOT block: this is called from the issuance path, between
// the ledger commit and the caller's Gateway submission, and any wait here is a
// wait in front of an authorised order. AsyncObserver is the implementation that
// honours it; a nil observer records nothing, which is the correct behaviour for
// a Guardian nobody wired one into.
type EntryObserver interface {
	ObserveEntry(journal.EntryObservation)
}

// observeEntry is the Guardian's single recording call site. Keeping it one
// function rather than three inline blocks is what makes "the observation cannot
// change the verdict" checkable by reading: it takes no error return and there is
// nothing for a caller to branch on.
func (g *RiskGuardian) observeEntry(obs journal.EntryObservation) {
	if g.observer == nil {
		return
	}
	g.observer.ObserveEntry(obs)
}

// entryObservation builds the row for one verdict.
//
// The ratios come from risk.MeasureEntry, which fills what it can and leaves the
// rest empty rather than failing — so this function has no error path and the call
// sites have no branch that could turn a measurement problem into a refusal.
func (g *RiskGuardian) entryObservation(
	in risk.Input, outcome string, verdict risk.Decision,
) journal.EntryObservation {
	ratios := risk.MeasureEntry(g.costs, in.Intent.Market,
		in.Intent.LimitPrice, in.Intent.StopPrice, in.Intent.TargetPrice)

	row := journal.EntryObservation{
		ID:              g.newID(),
		AccountRef:      g.accountRef,
		Market:          string(in.Intent.Market),
		Symbol:          in.Intent.Symbol,
		EntryPrice:      in.Intent.LimitPrice,
		StopPrice:       in.Intent.StopPrice,
		TargetPrice:     in.Intent.TargetPrice,
		BreakEvenPrice:  ratios.BreakEvenPrice,
		GrossRewardRisk: ratios.GrossRewardRisk,
		NetRewardRisk:   ratios.NetRewardRisk,
		// The scope is a constant because this build measures one: commission and
		// tax on both legs. Slippage is not in it, which is why the metric is
		// "수수료·세금 차감 후 RR" and not "cost-adjusted" — task 4.5.
		CostScope:            journal.CostScopeFeeTaxOnly,
		CostModelFingerprint: g.costs.Fingerprint(),
		Outcome:              outcome,
		ObservedAt:           in.Now,
	}
	if outcome == journal.OutcomeRefusedChain {
		// Reported by the verdict, never inferred from the reason: one reason code
		// is raised at 42 sites in internal/risk, so the inverse map is
		// many-to-one (risk.Decision.Step).
		row.StoppedStep = verdict.Step
		row.ReasonCode = string(verdict.Reason)
	}
	return row
}

// AsyncObserver writes observations on its own goroutine.
//
// The queue is bounded and a full queue drops rather than waits, which is the only
// answer consistent with the non-blocking contract: the alternative is the entry
// path waiting on the analysis path, and an analysis backlog must not become an
// execution delay. A drop is counted like any other loss.
type AsyncObserver struct {
	queue  chan journal.EntryObservation
	done   chan struct{}
	sink   ObservationSink
	losses LossCounter
}

// ObservationSink is the storage half. *journal.Journal satisfies it.
type ObservationSink interface {
	RecordEntryObservation(ctx context.Context, obs journal.EntryObservation) error
}

// LossCounter records a measurement that did not survive. It is an interface so
// that this package does not depend on the counter's implementation — the counter
// deliberately lives in a package with no journal dependency at all, and routing
// it through a narrow interface keeps that arrangement visible from here.
type LossCounter interface {
	Record(kind string, attrs ...any)
}

// Loss kinds, mirroring internal/measure/degrade's. Spelled as constants here
// rather than imported so that this package's dependency on the counter stays a
// one-method interface.
const (
	lossObservationWrite     = "observation_write_failed"
	lossRefusalUnrecoverable = "refusal_observation_lost"
)

// AsyncObserverOptions constructs one.
type AsyncObserverOptions struct {
	// Sink is where observations are written. Required.
	Sink ObservationSink
	// Losses counts what could not be written. Optional.
	Losses LossCounter
	// Depth bounds the queue. Zero uses DefaultObservationQueueDepth.
	Depth int
	// WriteTimeout bounds one write. Zero uses DefaultObservationWriteTimeout.
	WriteTimeout time.Duration
}

// DefaultObservationQueueDepth is how many observations may be in flight.
//
// Small on purpose. The queue exists to decouple one journal write from the entry
// path, not to buffer an outage: if writes are failing, a deep queue converts a
// visible loss into a delayed one and hides it behind memory.
const DefaultObservationQueueDepth = 64

// DefaultObservationWriteTimeout bounds one observation write. It is far under the
// decision TTL, but that is incidental — nothing waits on this — and it is here so
// a wedged write cannot stall the queue behind it forever.
const DefaultObservationWriteTimeout = 5 * time.Second

// NewAsyncObserver starts the writer. Close stops it.
func NewAsyncObserver(opts AsyncObserverOptions) *AsyncObserver {
	depth := opts.Depth
	if depth <= 0 {
		depth = DefaultObservationQueueDepth
	}
	timeout := opts.WriteTimeout
	if timeout <= 0 {
		timeout = DefaultObservationWriteTimeout
	}
	o := &AsyncObserver{
		queue:  make(chan journal.EntryObservation, depth),
		done:   make(chan struct{}),
		sink:   opts.Sink,
		losses: opts.Losses,
	}
	go o.run(timeout)
	return o
}

// ObserveEntry hands the row over without waiting. A full queue drops it.
func (o *AsyncObserver) ObserveEntry(row journal.EntryObservation) {
	select {
	case o.queue <- row:
	default:
		o.lost(row, "queue_full")
	}
}

// Close drains what is queued and stops the writer.
func (o *AsyncObserver) Close() {
	close(o.queue)
	<-o.done
}

func (o *AsyncObserver) run(timeout time.Duration) {
	defer close(o.done)
	for row := range o.queue {
		if o.sink == nil {
			o.lost(row, "no_sink")
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		err := o.sink.RecordEntryObservation(ctx, row)
		cancel()
		if err != nil {
			o.lost(row, err.Error())
		}
	}
}

// lost counts one. The kind distinguishes the two cases design D6 separates: an
// issued verdict's observation can be rebuilt from the decision preimage, and a
// refusal's cannot — there is no decision to rebuild from. Reporting them as one
// number would report a recoverable gap and a permanent one as the same fact.
func (o *AsyncObserver) lost(row journal.EntryObservation, reason string) {
	if o.losses == nil {
		return
	}
	kind := lossObservationWrite
	if row.Outcome != journal.OutcomeAllowedIssued {
		kind = lossRefusalUnrecoverable
	}
	o.losses.Record(kind,
		"outcome", row.Outcome,
		"decision_id", row.DecisionID,
		"symbol", row.Symbol,
		"error", reason)
}
