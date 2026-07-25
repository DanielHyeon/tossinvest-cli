package reconcile

// recovery.go is the restart sequence (harden-execution-base task 3.5).
//
// # Why a process that just started must not trade
//
// A crash can happen between "the request left" and "the answer arrived". The
// journal makes that window discoverable — an attempt found in DISPATCH_STARTED
// may or may not have executed — but discoverable is not the same as resolved.
// Until it is resolved, the engine does not know its own position, and a strategy
// that opens a new one on top is doubling an exposure it cannot see.
//
// So the gate is latched shut in the constructor, not somewhere a caller has to
// remember to call:
//
//	New()      → the entry gate is already blocked (recovery_incomplete)
//	Run()      → journal → resolution → account → reconstruction
//	completed  → the latch is released, and only then
//
// A Recovery that is never run therefore fails closed. That is the whole design:
// the dangerous state is the default, and it takes work to leave it.
//
// # What is deliberately not here
//
// Nothing in this file cancels, amends or places anything. Recovery observes and
// resolves; it never trades its way out of an ambiguity. The one action it
// delegates — IN_DOUBT resolution — is itself observation-only by construction
// (internal/execgw/indoubt.go has no mutator at all).
//
// Exits stay open throughout. The latch is an *entry* gate (§0.3): an operator
// who starts the engine into a mess must still be able to liquidate through it.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// Stabilisation tunes how the recovery snapshot is corroborated.
type Stabilisation struct {
	// Interval is the minimum gap between two counted snapshots. Zero takes
	// DefaultStabilisationInterval.
	Interval time.Duration
	// Required is how many consecutive agreeing snapshots are needed. Zero takes
	// DefaultStabilisationCount.
	Required int
	// MaxAttempts bounds how many snapshots are taken before giving up. An
	// account that never settles is not a reason to trade on the last reading.
	MaxAttempts int
}

func (s Stabilisation) withDefaults() Stabilisation {
	if s.Interval <= 0 {
		s.Interval = DefaultStabilisationInterval
	}
	if s.Required <= 0 {
		s.Required = DefaultStabilisationCount
	}
	if s.MaxAttempts <= 0 {
		s.MaxAttempts = 5
	}
	return s
}

// Options are the recovery's dependencies.
type Options struct {
	// Journal is the record of what the previous process was doing. Required.
	Journal *journal.Journal
	// Resolver settles the attempts whose outcome the crash left unknown.
	// Required: without it an IN_DOUBT attempt would stay unexamined, and
	// recovery would "complete" over an unknown live order.
	Resolver *execgw.Resolver
	// Collector reads the account. Required.
	Collector *Collector
	// Gate is latched shut until recovery completes. Required — a recovery that
	// cannot block anything is a report, not a safety mechanism.
	Gate *execgw.EntryGate
	// Clock drives the stabilisation waits. Defaults to clock.System().
	Clock      clock.Clock
	AccountRef string
	Comparer   Comparer
	Stabilise  Stabilisation
}

// Recovery is the restart sequence.
type Recovery struct {
	opts     Options
	stab     Stabilisation
	complete bool
}

// New arms the recovery: the entry gate is blocked before this function returns.
//
// Arming in the constructor rather than in Run is the point. A caller that builds
// a Recovery and forgets to run it gets an engine that refuses to open positions,
// which is the failure everyone can live with.
func New(opts Options) (*Recovery, error) {
	switch {
	case opts.Journal == nil:
		return nil, errors.New("reconcile: recovery needs the journal — " +
			"there is nothing else that knows what the previous process was doing")
	case opts.Resolver == nil:
		return nil, errors.New("reconcile: recovery needs an IN_DOUBT resolver")
	case opts.Collector == nil:
		return nil, errors.New("reconcile: recovery needs an account collector")
	case opts.Gate == nil:
		return nil, errors.New("reconcile: recovery needs an entry gate to hold shut")
	}
	r := &Recovery{opts: opts, stab: opts.Stabilise.withDefaults()}
	r.opts.Gate.Block(execgw.ReasonRecoveryIncomplete,
		"the restart recovery sequence has not completed; new entries stay blocked until it does")
	return r, nil
}

// Report is what the sequence found and did.
type Report struct {
	StartedAt   time.Time
	CompletedAt time.Time

	// NotDispatched are attempts that were provably never sent and were closed.
	NotDispatched []string
	// MovedToInDoubt are attempts the crash caught mid-dispatch.
	MovedToInDoubt []string
	// Resolutions is the outcome of every resolution attempted.
	Resolutions []execgw.Resolution
	// Unresolved are attempts that could not be settled by observation. Their
	// symbols stay blocked by the resolver's own latch until an operator acts.
	Unresolved []string
	// StillPending are attempts left in a state recovery does not settle —
	// an ACK the previous process never confirmed. They block their symbol
	// through the journal, not through this sequence.
	StillPending []journal.BlockedSymbol

	// Snapshot is the stabilised account reading.
	Snapshot Snapshot
	// SnapshotsTaken is how many readings it took to stabilise.
	SnapshotsTaken int
	// Local is the reconstructed local belief.
	Local LocalState
	// Diff is the comparison of the two.
	Diff Diff

	// Complete reports that the sequence finished. Only then is the entry latch
	// released — and even then, a blocking Diff immediately latches a different
	// one (task 3.6).
	Complete bool
}

// ErrRecoveryIncomplete means the sequence did not finish. The gate stays shut.
var ErrRecoveryIncomplete = errors.New("reconcile: restart recovery did not complete")

// Run performs the sequence in the order the spec fixes:
// journal resolution → account reads → state reconstruction.
//
// The order is causal, not cosmetic. Resolving the journal can *change* the
// account's meaning — an IN_DOUBT place that turns out to have executed is a
// position the snapshot must be compared against — so reading the account first
// would reconstruct state from a picture that is about to move.
func (r *Recovery) Run(ctx context.Context) (Report, error) {
	clk := r.clock()
	report := Report{StartedAt: clk.Now()}

	// 1. The journal's own restart rules: RECORDED closes as NOT_DISPATCHED
	//    (provably unsent), DISPATCH_STARTED becomes IN_DOUBT (may have executed).
	recovered, err := r.opts.Journal.RecoverPending(ctx)
	if err != nil {
		return report, fmt.Errorf("%w: applying the journal's restart rules: %v",
			ErrRecoveryIncomplete, err)
	}
	report.NotDispatched = recovered.NotDispatched
	report.MovedToInDoubt = recovered.InDoubt

	// 2. Resolve every attempt whose outcome is unknown. Observation only — the
	//    resolver has no mutator, so this cannot re-submit anything.
	pending, err := r.opts.Journal.PendingAttempts(ctx)
	if err != nil {
		return report, fmt.Errorf("%w: listing attempts to resolve: %v", ErrRecoveryIncomplete, err)
	}
	for _, rec := range pending {
		if rec.State != journal.StateInDoubt {
			// ACKED is the only other state that survives here: the broker named
			// an order and the previous process died before confirming it. It is
			// not ambiguous, it is unfinished, and the journal's own per-symbol
			// check keeps it blocked.
			blocked, berr := r.blockedSymbol(ctx, rec)
			if berr != nil {
				return report, fmt.Errorf("%w: %v", ErrRecoveryIncomplete, berr)
			}
			report.StillPending = append(report.StillPending, blocked)
			continue
		}
		res, rerr := r.opts.Resolver.Resolve(ctx, rec.ID)
		if rerr != nil {
			return report, fmt.Errorf("%w: resolving attempt %s: %v",
				ErrRecoveryIncomplete, rec.ID, rerr)
		}
		report.Resolutions = append(report.Resolutions, res)
		if res.State == journal.StateUnresolvedInDoubt {
			report.Unresolved = append(report.Unresolved, res.AttemptID)
		}
	}

	// 3. Read the account, and keep reading until it stops moving. A single
	//    reading taken while a fill is settling would reconstruct a position that
	//    is already wrong.
	snap, taken, err := r.stableSnapshot(ctx)
	report.SnapshotsTaken = taken
	if err != nil {
		return report, err
	}
	report.Snapshot = snap

	// 4. Reconstruct what the engine believes, and compare.
	local, err := LocalStateFromJournal(ctx, r.opts.Journal, r.opts.AccountRef)
	if err != nil {
		return report, fmt.Errorf("%w: reconstructing local state: %v", ErrRecoveryIncomplete, err)
	}
	report.Local = local
	report.Diff = r.opts.Comparer.Compare(snap, local)

	// 5. Recovery itself is finished. Releasing this latch is not a statement
	//    that everything is fine — a blocking diff latches its own reason, and
	//    an unresolved attempt has already latched the resolver's. It only says
	//    the engine now knows what it is looking at.
	report.CompletedAt = clk.Now()
	report.Complete = true
	r.complete = true
	r.opts.Gate.Clear(execgw.ReasonRecoveryIncomplete)

	if report.Diff.BlocksEntry() {
		r.opts.Gate.Block(execgw.ReasonReconcileMismatch,
			"recovery finished with the account and the engine disagreeing: "+report.Diff.Summary())
	}
	return report, nil
}

// Complete reports whether the sequence has finished.
func (r *Recovery) Complete() bool { return r.complete }

// stableSnapshot collects until the account stops changing.
func (r *Recovery) stableSnapshot(ctx context.Context) (Snapshot, int, error) {
	clk := r.clock()
	stabiliser := &Stabiliser{MinInterval: r.stab.Interval, Required: r.stab.Required}

	var taken int
	for attempt := 1; attempt <= r.stab.MaxAttempts; attempt++ {
		snap, err := r.opts.Collector.Collect(ctx)
		if err != nil {
			return Snapshot{}, taken, fmt.Errorf("%w: %v", ErrRecoveryIncomplete, err)
		}
		taken++
		if result := stabiliser.Offer(snap); result.Stable {
			return snap, taken, nil
		}
		if attempt == r.stab.MaxAttempts {
			break
		}
		if err := clk.Sleep(ctx, r.stab.Interval); err != nil {
			return Snapshot{}, taken, fmt.Errorf("%w: %v", ErrRecoveryIncomplete, err)
		}
	}
	// An account that will not settle is not a reason to trade on the last
	// reading; the gate stays shut and an operator looks.
	return Snapshot{}, taken, fmt.Errorf(
		"%w: the account did not produce %d agreeing snapshots in %d readings",
		ErrRecoveryIncomplete, r.stab.Required, taken)
}

func (r *Recovery) blockedSymbol(ctx context.Context, rec journal.AttemptRecord) (journal.BlockedSymbol, error) {
	intent, err := r.opts.Journal.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		return journal.BlockedSymbol{}, fmt.Errorf("reading the intent of attempt %s: %w", rec.ID, err)
	}
	return journal.BlockedSymbol{
		Market:    intent.Market,
		Symbol:    intent.Symbol,
		AttemptID: rec.ID,
		State:     rec.State,
	}, nil
}

func (r *Recovery) clock() clock.Clock {
	if r.opts.Clock == nil {
		return clock.System()
	}
	return r.opts.Clock
}
