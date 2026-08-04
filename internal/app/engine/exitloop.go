package engine

// exitloop.go is the exit observation loop (add-core-domain tasks 7.4 and 7.5):
// the component that turns internal/exitpolicy's pure judgement into a
// protection decision the account actually feels.
//
// # What one cycle is
//
//	 1. read the account's positions and its open exit states
//	 2. open a state for a newly held position, alert on one that cannot have one
//	 3. one price read for every held symbol, through the Retrier
//	 4. per position: evaluate, persist the judgement, and — if it proposed
//	    something — arm, mint, attach, submit
//
// Steps 1, 2 and 4 are local. Step 3 is the only broker call the loop makes, and
// it is one call for every held symbol rather than one per symbol
// (exit-policy: 보유 심볼 fan-out).
//
// # The rate budget (§0.4)
//
// One `GET /api/v1/prices` per interval, default 5 s, for the whole held set:
// **0.2 req/s**, independent of how many positions are open. The retry matrix's
// steady-state total (archive/2026-07-26-harden-execution-base/retry-matrix.md
// §3) was under 0.5 req/s; this loop takes it to under 0.7 req/s, and the
// retry worst case is bounded by the query budget (8 s) exactly as every other
// GET is. Nothing here polls per position: a fan-out that scaled with the
// portfolio would make the budget a function of a number nobody bounded.
//
// # Deference to fill detection
//
// Fill detection has an SLO and this loop does not, and when the two compete the
// order is fixed: **fill detection first** (exit-policy: 체결 감지 SLO에 양보).
// The reason is not politeness. Fill detection is what moves `taken_ratio_total`
// and resolves a pending proposal; a loop that starved it would go on judging
// against a position it believes is larger than it is, and would re-propose
// take-profits against quantity that has already been sold. So: when the
// injected [SLOPressure] reports the detector is behind, the cycle is skipped
// whole — the price read is not made, the judgement is not run, and the outage
// clock keeps running, which means sustained deference eventually escalates
// exactly as sustained failure does. That last part is deliberate: "we chose not
// to look" and "we could not look" leave a position equally unprotected.
//
// # The outage ladder
//
//	15 s   the landed staleness table blocks new entries (execgw.QueryPrice)
//	60 s   EXIT_OBSERVATION_OUTAGE → ENTRY_BLOCKED + a critical alert
//
// The first rung costs nothing here: [execgw.Retrier] stamps freshness on every
// success and the entry gate expires it on its own. Wiring that stamp is the
// point of routing the price read through the Retrier at all — without it the
// price query is never fresh, and a gate-ON engine refuses every entry with
// QUERY_STALE.
//
// The second rung is this file's, and it is the fourth producer of an automatic
// tightening (issues.md, task 3.2). It fires once per outage: the transition is
// idempotent (a no-op once the mode is already there) but the alert is not, and
// a critical alert re-published every five seconds is a transport budget spent
// on one fact.
//
// # What this file does not do
//
// It does not decide anything about protection — internal/exitpolicy does, from
// values — and it does not write a baseline, a watermark or a level itself:
// every such write goes through journal.RecordExitJudgement, in the same
// transaction as the arming and the history row. It also holds no state that a
// restart would need. The armed proposal, the level, the watermark and the
// baseline are all rows; the only fields on the observer are alert latches and
// timers, whose loss on restart re-raises the alert rather than losing it.

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// DefaultExitObservationInterval is the observation period (exit-policy: 주기
// 기본 5초).
const DefaultExitObservationInterval = 5 * time.Second

// DefaultExitObservationOutage is how long observation may fail before the
// account is tightened (exit-policy: staleness 임계 기본 60초).
//
// It is four intervals, which is what makes a single failed read — or a single
// deferred cycle — not an incident while a persistent one is.
const DefaultExitObservationOutage = 60 * time.Second

// DefaultExitLiquidationDelayBound is how long a breach liquidation may be held
// back by an unsettled mutation on the same symbol before it is a critical alert
// (exit-policy: 지연이 유계를 넘으면 critical 알림).
//
// The bound is the IN_DOUBT resolution procedure's own: a mutation that has not
// settled within it is a mutation the resolver has already failed to settle, so
// the delay is no longer "an order is in flight" but "nobody knows what the
// account holds" — and a liquidation waiting on that is the residual risk
// exit-policy names as 브로커측 보호 도입 전의 잔존 리스크.
const DefaultExitLiquidationDelayBound = 30 * time.Second

// PriceReader is the market read the loop makes. Context.Official satisfies it.
type PriceReader interface {
	Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)
}

// ReductionIssuer mints the RISK_REDUCING decision a proposal is submitted
// under. *execgw.RiskGuardian satisfies it.
//
// It is the Guardian path and not the plain issuer on purpose: exit-policy
// requires a proposal to travel "Guardian 위험 감소 경로 경유", which is what
// checks the one thing an exit can get wrong — selling more than the account
// holds.
type ReductionIssuer interface {
	IssueReduction(ctx context.Context, req execgw.ReductionIssuance) (execgw.Issued, error)
}

// ExitSubmitter is the order surface. *execgw.Gateway satisfies it, and nothing
// else in this package may: a proposal that reached the broker without a
// persisted decision would be an order with no authority behind it.
type ExitSubmitter interface {
	Place(ctx context.Context, req execgw.PlaceRequest) (execgw.Outcome, error)
	Cancel(ctx context.Context, req execgw.CancelRequest) (execgw.Outcome, error)
}

// ExitAlerter receives the loop's operator alerts. *obs.Notifier satisfies it.
type ExitAlerter interface {
	Notify(ctx context.Context, e obs.Event) error
}

// SLOPressure reports whether fill detection is behind its objective, which is
// the condition this loop yields to. *filldetect.Detector does not satisfy it
// directly — the adapter is in the engine's wiring, so this package does not
// depend on the detector.
type SLOPressure interface {
	FillDetectionBehind() bool
}

// FloorSource supplies the RECONCILE confirmed floor for one symbol (task 7.5).
//
// The bool is "this symbol is under RECONCILE, so the floor applies". A symbol
// that is not is not capped at all: the confirmed floor is the bound on what an
// automated path may sell *while the engine and the account disagree*, and
// applying it in normal operation would put a broker round trip in front of
// every liquidation.
type FloorSource interface {
	ConfirmedFloor(ctx context.Context, market, symbol string) (riskcalc.ConfirmedFloor, bool, error)
}

// ExitObserverOptions is the loop's wiring.
type ExitObserverOptions struct {
	// Journal owns every durable thing the loop touches. Required.
	Journal *journal.Journal
	// Prices is the market read. Required.
	Prices PriceReader
	// Retrier runs that read under the matrix and stamps its freshness.
	// Required — see the file header on QUERY_STALE.
	Retrier *execgw.Retrier
	// Issuer mints the reduction decisions. Required.
	Issuer ReductionIssuer
	// Submit is the order surface. Required.
	Submit ExitSubmitter
	// Alerts receives the operator alerts. Optional only so the loop can be unit
	// tested; an engine without one judges silently.
	Alerts ExitAlerter
	// Log receives the structured lines. Optional.
	Log *obs.Logger
	// Costs is the shared cost model, for the real break-even the BREAKEVEN
	// ratchet level is composed from. Required: an unconfigured model drops
	// break-even onto the entry price, which would promote the baseline early.
	Costs costs.Model
	// Floor supplies the RECONCILE cap. Optional; nil caps nothing, which is the
	// right behaviour for a build with no reconciliation loop running.
	Floor FloorSource
	// SLO is the deference source. Optional; nil never defers.
	SLO SLOPressure
	// Escalate persists the outage tightening. Optional; nil leaves the alert as
	// the only consequence. *journal.Journal satisfies it.
	Escalate execgw.ModeEscalation
	// Announcer is told about a transition that changed something. Optional.
	Announcer journal.ModeAnnouncer

	// AccountRef scopes everything. Required.
	AccountRef string
	// Clock drives the interval and every duration judgement. Defaults to the
	// system clock.
	Clock clock.Clock
	// Interval is the observation period. Zero takes the default.
	Interval time.Duration
	// OutageAfter is the tightening threshold. Zero takes the default.
	OutageAfter time.Duration
	// DelayBound is the in-flight delay alert threshold. Zero takes the default.
	DelayBound time.Duration

	// Ratchet overrides the trigger table. Nil takes exitpolicy's default.
	Ratchet *exitpolicy.RatchetConfig
	// Ladder overrides the rung table. Nil takes exitpolicy's default.
	Ladder *exitpolicy.LadderPolicy
	// CommonPolicy is the startup snapshot applied only when a new exit state is
	// opened. Empty preserves legacy RATCHET.
	CommonPolicy string

	// NewID mints intent ids. Defaults to 128 bits of randomness.
	NewID func() string
}

// ExitObserver is the loop.
type ExitObserver struct {
	opts    ExitObserverOptions
	clk     clock.Clock
	ratchet exitpolicy.RatchetConfig
	ladder  exitpolicy.LadderPolicy

	// lastObserved is when a price read last succeeded. Zero means "not yet",
	// which is measured from start so a loop that never reads escalates.
	lastObserved time.Time
	// startedAt anchors that measurement.
	startedAt time.Time
	// outageRaised stops the critical alert repeating for one outage.
	outageRaised bool
	// unmanaged latches the external-position alert, one per position id
	// (exit-policy: 발견 시 알림, not "every five seconds").
	unmanaged map[string]bool
	// refused latches the refusal-to-judge alert per position.
	refused map[string]bool
	// quarantineAnnounced latches the quarantine-creation alert per quarantine
	// row — position, generation and version — so a release followed by a
	// re-quarantine announces again. See exit_quarantine_announce.go. It is
	// initialised lazily on first use; a nil map already reads as "nothing
	// announced", which is the correct starting state.
	quarantineAnnounced map[string]bool
	// delayedSince is when a symbol's liquidation first found the symbol busy.
	delayedSince map[string]time.Time
	// delayAlerted stops that alert repeating within one delay.
	delayAlerted map[string]bool
	// observationSequence distinguishes fallback observations when the quote
	// source supplies no FetchedAt and a test/frozen clock does not advance.
	observationSequence atomic.Uint64
}

// NewExitObserver validates the wiring and snapshots the common selection used
// only for opening new states. Active ladder states carry their own policy ID
// and are resolved independently on every judgement.
func NewExitObserver(opts ExitObserverOptions) (*ExitObserver, error) {
	switch {
	case opts.Journal == nil:
		return nil, errors.New("engine: the exit observation loop needs the journal that owns the exit states")
	case opts.Prices == nil:
		return nil, errors.New("engine: the exit observation loop needs a price read; a loop with no observation " +
			"holds every judgement forever and never says so")
	case opts.Retrier == nil:
		return nil, errors.New("engine: the exit observation loop needs the Retrier — it is what stamps the " +
			"price query's freshness, and without the stamp every entry is refused as QUERY_STALE")
	case opts.Issuer == nil:
		return nil, errors.New("engine: the exit observation loop needs a reduction issuer; a proposal must " +
			"travel the Guardian risk-reducing path")
	case opts.Submit == nil:
		return nil, errors.New("engine: the exit observation loop needs the execution gateway")
	case strings.TrimSpace(opts.AccountRef) == "":
		return nil, errors.New("engine: the exit observation loop is scoped to one account; none was named")
	case !opts.Costs.Configured():
		return nil, errors.New("engine: the exit observation loop needs a cost model; an unconfigured one " +
			"prices the round trip as free and drops the break-even baseline onto the entry price")
	}

	ratchet := exitpolicy.DefaultRatchetConfig()
	if opts.Ratchet != nil {
		ratchet = *opts.Ratchet
	}
	if err := ratchet.Validate(); err != nil {
		return nil, fmt.Errorf("engine: the configured ratchet table is not a ratchet: %w", err)
	}
	ladder := exitpolicy.DefaultLadderPolicy()
	if opts.Ladder != nil {
		ladder = *opts.Ladder
	}
	if err := ladder.Validate(); err != nil {
		return nil, fmt.Errorf("engine: the configured rung table is not a ladder: %w", err)
	}
	if id := strings.TrimSpace(opts.CommonPolicy); id != "" {
		if _, ok := exitpolicy.CommonPolicyByID(id); !ok {
			return nil, fmt.Errorf("engine: unknown common exit policy %q", id)
		}
		opts.CommonPolicy = id
	}

	clk := opts.Clock
	if clk == nil {
		clk = clock.System()
	}
	if opts.NewID == nil {
		opts.NewID = randomExitID
	}
	return &ExitObserver{
		opts:         opts,
		clk:          clk,
		ratchet:      ratchet,
		ladder:       ladder,
		startedAt:    clk.Now(),
		unmanaged:    map[string]bool{},
		refused:      map[string]bool{},
		delayedSince: map[string]time.Time{},
		delayAlerted: map[string]bool{},
	}, nil
}

// Interval is the observation period this observer runs at.
func (o *ExitObserver) Interval() time.Duration {
	if o.opts.Interval > 0 {
		return o.opts.Interval
	}
	return DefaultExitObservationInterval
}

func (o *ExitObserver) outageAfter() time.Duration {
	if o.opts.OutageAfter > 0 {
		return o.opts.OutageAfter
	}
	return DefaultExitObservationOutage
}

func (o *ExitObserver) delayBound() time.Duration {
	if o.opts.DelayBound > 0 {
		return o.opts.DelayBound
	}
	return DefaultExitLiquidationDelayBound
}

// Run observes until the context ends.
//
// It returns the context's error and never a judgement's: a failed cycle is the
// hold exit-policy specifies (관측 실패 시 … 판정은 보류된다), and a loop that
// exited on one would remove the protection it exists to provide. What a
// repeated failure produces is the outage ladder, not a return.
func (o *ExitObserver) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		o.reportCycle(o.ObserveOnce(ctx))
		if err := o.clk.Sleep(ctx, o.Interval()); err != nil {
			return err
		}
	}
}

// reportCycle writes the line a failed cycle is owed (change a064).
//
// ExitCycle.Err's declaration has always said "It is reported and not returned
// by Run". Nothing reported it: this loop discarded the whole cycle, and it is
// registered in the runtime with no Health, so the supervisor's degradation
// ladder could not count cycle failures either. The failure that made this
// visible was a quarantine — the judgement transaction created one, returned its
// sentinel, and the creating cycle left no trace anywhere.
//
// A log line and not an alert. See obs.EventExitCycleFailed for why: a cycle
// failure has no single meaning, and grading it critical would let a transient
// ledger error stop a live account. The conditions that mean a position is
// actually unprotected raise their own critical events from where they happen.
//
// Successful cycles say nothing. One line every five seconds forever is not
// observability.
func (o *ExitObserver) reportCycle(cycle ExitCycle) {
	if cycle.Err == nil {
		return
	}
	o.logErr(obs.EventExitCycleFailed, cycle.Err,
		"the exit observation cycle did not complete; the loop retries on the next interval")
}

// ExitCycle is what one observation did, for tests and for the status surface.
type ExitCycle struct {
	// Deferred reports that the cycle yielded to fill detection.
	Deferred bool
	// Observed is how many symbols a price came back for.
	Observed int
	// Judged is how many positions were evaluated.
	Judged int
	// Proposed is how many proposals were armed and submitted.
	Proposed int
	// Opened is how many exit states were created this cycle.
	Opened int
	// Unmanaged is how many held positions the policy will not manage.
	Unmanaged int
	// Escalated reports that this cycle tightened the operating mode.
	Escalated bool
	// Err is the cycle's first failure, if any. It is reported and not returned
	// by Run: the loop holds rather than stops.
	Err error
}

// ObserveOnce runs one cycle. It is exported for the tests and for a tracer
// run, which drives the loop by hand rather than by clock.
func (o *ExitObserver) ObserveOnce(ctx context.Context) ExitCycle {
	var cycle ExitCycle
	observation := cycleObservation{
		at: o.clk.Now().UTC(), sequence: o.observationSequence.Add(1),
	}

	if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {
		// Yield the whole cycle, and let the outage clock keep running. See the
		// file header: an unobserved position is unprotected whether the reason
		// was a failure or a choice.
		cycle.Deferred = true
		o.checkOutage(ctx, &cycle)
		return cycle
	}

	states, err := o.workingSet(ctx, &cycle)
	if err != nil {
		cycle.Err = err
		return cycle
	}
	if len(states) == 0 {
		// Nothing is held, so there is nothing to observe and nothing to be
		// unprotected. The outage clock is reset rather than left running: an
		// account with no positions that has not read a price is not in an
		// outage, and escalating it would block entries for a condition that
		// only entering could clear.
		o.lastObserved = o.clk.Now()
		o.outageRaised = false
		return cycle
	}

	quotes, err := o.observe(ctx, symbolsOf(states))
	if err != nil {
		cycle.Err = err
		o.checkOutage(ctx, &cycle)
		return cycle
	}
	o.lastObserved = o.clk.Now()
	o.outageRaised = false
	cycle.Observed = len(quotes)

	for _, state := range states {
		quote, ok := quotes[strings.ToUpper(strings.TrimSpace(state.position.Symbol))]
		if !ok {
			// A symbol the price read did not answer for is a symbol this cycle
			// did not observe. Hold it; the outage clock is per account and the
			// next cycle may answer.
			continue
		}
		cycle.Judged++
		if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {
			cycle.Err = err
		}
	}
	return cycle
}

// managed is one position and the exit state that governs it.
type managed struct {
	position       journal.Position
	state          journal.ExitState
	policyIdentity exitpolicy.PolicyIdentity
	identityErr    error
}

// workingSet reconciles the held positions with the exit states, opening the
// ones that are missing and alerting on the ones that cannot have one.
func (o *ExitObserver) workingSet(ctx context.Context, cycle *ExitCycle) ([]managed, error) {
	positions, err := o.opts.Journal.Positions(ctx, o.opts.AccountRef)
	if err != nil {
		return nil, err
	}
	stateResults, err := o.opts.Journal.OpenExitStateResults(ctx, o.opts.AccountRef)
	if err != nil {
		return nil, err
	}
	byPosition := make(map[string]journal.ExitStateResult, len(stateResults))
	for _, result := range stateResults {
		byPosition[result.State.PositionID] = result
	}

	var out, refused []managed
	for _, p := range positions {
		if p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) {
			continue
		}
		if !p.ExitEligible() {
			// exit-policy: entry 결정이 없는 포지션은 exit 정책의 대상이 아니며
			// 발견 시 알림을 발송한다. There is no stop to build a baseline out
			// of and no risk to measure R against, so the honest answer is to
			// name it rather than to invent one.
			cycle.Unmanaged++
			o.alertUnmanaged(ctx, p)
			continue
		}
		result, ok := byPosition[p.ID]
		state := result.State
		if !ok {
			opened, err := o.openState(ctx, p)
			if err != nil {
				if cycle.Err == nil {
					cycle.Err = err
				}
				continue
			}
			if opened.PositionID == "" {
				// The state exists and is completed: the position closed and
				// reopened is a new instance, and a completed policy judges
				// nothing further.
				continue
			}
			cycle.Opened++
			state = opened
		}
		if result.Corruption != nil {
			q, qerr := o.opts.Journal.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq,
				"stored_snapshot_corrupt", result.Corruption.Error())
			if qerr != nil {
				if cycle.Err == nil {
					cycle.Err = qerr
				}
				continue
			}
			refused = append(refused, managed{position: p, state: state,
				identityErr: fmt.Errorf("%w (quarantine version %d)", result.Corruption, q.Version)})
			o.announceQuarantine(ctx, p, q)
			continue
		}
		if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil {
			if cycle.Err == nil {
				cycle.Err = qerr
			}
			continue
		} else if active {
			refused = append(refused, managed{position: p, state: state,
				identityErr: fmt.Errorf("%w (version %d): %s", journal.ErrExitSnapshotQuarantined, q.Version, q.Reason)})
			continue
		}
		identity, identityErr := managedPolicyIdentity(state, p.Adopted())
		if identityErr != nil {
			q, qerr := o.opts.Journal.QuarantineExitSnapshot(ctx, p.ID, p.InstanceSeq,
				"legacy_policy_identity_unknown", identityErr.Error())
			if qerr != nil {
				if cycle.Err == nil {
					cycle.Err = qerr
				}
				continue
			}
			refused = append(refused, managed{position: p, state: state,
				identityErr: fmt.Errorf("%w (quarantine version %d)", identityErr, q.Version)})
			o.announceQuarantine(ctx, p, q)
			continue
		}
		out = append(out, managed{
			position: p, state: state, policyIdentity: identity, identityErr: identityErr,
		})
	}
	// Quarantined rows come last. alertRefused may synchronously retry a
	// publisher, so every valid position—including an emergency breach—is
	// recorded/armed/submitted before alert delivery can wait.
	return append(out, refused...), nil
}

func managedPolicyIdentity(state journal.ExitState, adopted bool) (exitpolicy.PolicyIdentity, error) {
	identity := state.PolicyIdentity
	if strings.TrimSpace(identity.ID) != "" || strings.TrimSpace(identity.Version) != "" ||
		strings.TrimSpace(identity.Digest) != "" {
		if err := identity.Validate(); err != nil {
			return exitpolicy.PolicyIdentity{}, err
		}
		return identity, nil
	}
	if state.PolicyKind == journal.ExitPolicyLadder {
		return exitpolicy.LegacyLadderPolicyIdentity(state.PolicyID, adopted)
	}
	return exitpolicy.LegacyRatchetPolicyIdentity()
}

// openState creates the protection state of a position that has just started
// being held.
//
// # Two sources, branched on the record and not on a fallback
//
// exit-policy requires the opening path to accept both (SHALL): an engine-entered
// position's t0 comes from its entry decision, an adopted one's from its adoption
// record. The branch is on which record the projection says justifies the
// position, because for an adopted position the decision lookup is not a lookup
// that fails — there is nothing to look up, and a path that tried and fell back
// would turn "the ledger is broken" and "this position was adopted" into the same
// code path.
//
// For the engine-entered arm the frozen pair is the *entry decision's* limit and
// stop, not the fill's average. Both are known at t0, both are what the Guardian
// chain sized and judged against, and their difference is therefore the initial
// risk the whole R scale is expressed in. The average fill price moves while the
// entry is still filling; freezing a denominator that moves would make the same
// price mean two different R values over the life of one position.
//
// For the adopted arm the pair is the observation the adoption was judged on and
// the synthetic stop derived from it (design A2's manage-forward t0). The cost
// basis is in neither: anchoring on it would put every long-held winner outside
// its own band on the first observation.
func (o *ExitObserver) openState(ctx context.Context, p journal.Position) (journal.ExitState, error) {
	if p.Adopted() {
		return o.openAdoptedState(ctx, p)
	}
	decision, err := o.opts.Journal.LookupDecision(ctx, p.EntryDecisionID)
	if err != nil {
		return journal.ExitState{}, fmt.Errorf(
			"engine: reading the entry decision of position %s: %w", p.ID, err)
	}
	preimage, err := journal.ParsePreimage(decision.PreimageKind, decision.RiskPreimage)
	if err != nil {
		return journal.ExitState{}, fmt.Errorf(
			"engine: reading the entry decision of position %s: %w", p.ID, err)
	}
	entry, ok := preimage.(journal.RiskIntent)
	if !ok {
		return journal.ExitState{}, fmt.Errorf(
			"engine: the entry decision of position %s is a %s and carries no stop price",
			p.ID, decision.PreimageKind)
	}

	kind, policyID := journal.ExitPolicyRatchet, ""
	if o.opts.CommonPolicy != "" {
		kind, policyID = journal.ExitPolicyLadder, o.opts.CommonPolicy
	}
	state, err := o.opts.Journal.OpenExitState(ctx, journal.ExitStateSeed{
		PositionID:  p.ID,
		PolicyKind:  kind,
		PolicyID:    policyID,
		EntryPrice:  entry.EntryPrice,
		InitialStop: entry.StopPrice,
	})
	switch {
	case errors.Is(err, journal.ErrExitStateExists):
		// It exists but was not in the open set, so it is completed.
		return journal.ExitState{}, nil
	case err != nil:
		return journal.ExitState{}, fmt.Errorf("engine: opening the exit state of %s: %w", p.ID, err)
	}
	o.log(obs.EventEngineStarted, false,
		obs.FieldSymbol, p.Symbol,
		"position_id", p.ID,
		"baseline", state.Baseline,
		"initial_risk", state.InitialRisk,
		obs.FieldDetail, "the exit policy is now managing this position")
	return state, nil
}

// openAdoptedState is the adopted arm. It reads no decision, because there is
// none: the adoption record is the whole justification and the journal seeds the
// row from it in one transaction.
//
// A crash between the adoption commit and this call leaves a position with an
// adoption and no exit state. That is precisely the state this arm recovers
// from — the next cycle finds the position eligible, finds no open state, and
// completes the opening from the record already on disk (exit-policy: 편입 기록
// 영속 후 크래시). Nothing is re-observed and nothing is re-adopted.
func (o *ExitObserver) openAdoptedState(ctx context.Context, p journal.Position) (journal.ExitState, error) {
	state, err := o.opts.Journal.OpenAdoptedExitState(ctx, p.ID)
	switch {
	case errors.Is(err, journal.ErrExitStateExists):
		// It exists but was not in the open set, so it is completed.
		return journal.ExitState{}, nil
	case err != nil:
		return journal.ExitState{}, fmt.Errorf(
			"engine: opening the exit state of adopted position %s: %w", p.ID, err)
	}
	o.log(obs.EventEngineStarted, false,
		obs.FieldSymbol, p.Symbol,
		"position_id", p.ID,
		"adoption_id", p.AdoptionID,
		"baseline", state.Baseline,
		"initial_risk", state.InitialRisk,
		obs.FieldDetail, "the exit policy is now managing this adopted position")
	return state, nil
}

type observedQuote struct {
	Price     string
	FetchedAt time.Time
}

type cycleObservation struct {
	at       time.Time
	sequence uint64
}

// observe performs the one price read of the cycle.
func (o *ExitObserver) observe(ctx context.Context, symbols []string) (map[string]observedQuote, error) {
	var quotes []domain.Quote
	err := o.opts.Retrier.Query(ctx, execgw.QueryPrice, func(ctx context.Context) error {
		var qerr error
		quotes, qerr = o.opts.Prices.Prices(ctx, symbols)
		return qerr
	})
	if err != nil {
		return nil, err
	}

	out := make(map[string]observedQuote, len(quotes))
	for _, q := range quotes {
		if q.Last <= 0 {
			// A quote with no last trade is not an observation. Recording it as
			// zero would read as a total collapse and liquidate the position.
			continue
		}
		out[strings.ToUpper(strings.TrimSpace(q.Symbol))] = observedQuote{
			Price: decimalOf(q.Last), FetchedAt: q.FetchedAt,
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("engine: the price read answered for none of %d held symbol(s)", len(symbols))
	}
	return out, nil
}

// checkOutage runs the escalation ladder's second rung.
func (o *ExitObserver) checkOutage(ctx context.Context, cycle *ExitCycle) {
	since := o.lastObserved
	if since.IsZero() {
		since = o.startedAt
	}
	if o.clk.Now().Sub(since) < o.outageAfter() {
		return
	}
	if o.outageRaised {
		return
	}
	o.outageRaised = true

	o.alert(ctx, obs.Event{
		Type:  obs.EventExitObservationOutage,
		Key:   string(obs.EventExitObservationOutage) + "|" + o.opts.AccountRef,
		Title: "exit observation has been down for " + o.clk.Now().Sub(since).String(),
		Body: "no price has been observed for the held positions since " + since.UTC().Format(time.RFC3339) +
			"; with no broker-resident stop an unobserved position is an unprotected one, so new entries " +
			"are blocked until observation recovers and an operator relaxes the mode",
		Fields: map[string]any{
			obs.FieldAccount: o.opts.AccountRef,
			"outage_seconds": int(o.clk.Now().Sub(since).Seconds()),
		},
	})

	if o.opts.Escalate == nil || strings.TrimSpace(o.opts.AccountRef) == "" {
		return
	}
	_, changed, err := o.opts.Escalate.EscalateOperatingMode(ctx, o.opts.AccountRef,
		journal.ModeTriggerExitObservationOutage, o.opts.Announcer)
	if err != nil {
		o.logErr(obs.EventOperatingMode, err,
			"the exit observation outage did not reach the operating mode, so a restart would lift the block")
		return
	}
	cycle.Escalated = changed
}

// judge evaluates one position and acts on what the evaluation asked for.
func (o *ExitObserver) judge(ctx context.Context, m managed, quote observedQuote,
	observation cycleObservation, cycle *ExitCycle) error {
	if m.identityErr != nil {
		o.alertRefused(ctx, m, m.identityErr)
		return nil
	}
	breakEven, err := o.breakEven(m)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}

	switch m.state.PolicyKind {
	case journal.ExitPolicyLadder:
		return o.judgeLadder(ctx, m, quote, observation, breakEven, cycle)
	default:
		return o.judgeRatchet(ctx, m, quote, observation, breakEven, cycle)
	}
}

// breakEven is the cost-inclusive break-even sell price of the position, which
// is the BREAKEVEN level's candidate.
//
// It is priced against the frozen entry price and the position's *current*
// quantity, because that is what a sale would actually be charged on. §0.3 is
// not in play: this number can only raise a protective baseline, never withhold
// a liquidation.
func (o *ExitObserver) breakEven(m managed) (string, error) {
	market, err := clock.ParseMarket(m.position.Market)
	if err != nil {
		return "", err
	}
	quantity := strings.TrimSpace(m.position.Quantity)
	if quantity == "" || isZeroQuantity(quantity) {
		return "", fmt.Errorf("position %s holds no quantity to price a break-even against", m.position.ID)
	}
	return o.opts.Costs.BreakEvenSellPrice(m.state.EntryPrice, quantity, market)
}

func (o *ExitObserver) judgeRatchet(ctx context.Context, m managed, quote observedQuote,
	observation cycleObservation, breakEven string, cycle *ExitCycle) error {
	activeIdentity, err := exitpolicy.RatchetPolicyIdentity(o.ratchet)
	if err != nil || !samePolicyIdentity(m.policyIdentity, activeIdentity) {
		if err == nil {
			err = fmt.Errorf("%w: position %s preserves %s@%s/%s, active ratchet is %s@%s/%s",
				exitpolicy.ErrPolicyIdentityConflict, m.position.ID,
				m.policyIdentity.ID, m.policyIdentity.Version, m.policyIdentity.Digest,
				activeIdentity.ID, activeIdentity.Version, activeIdentity.Digest)
		}
		o.alertRefused(ctx, m, err)
		return nil
	}
	snapshotContext, err := o.snapshotContext(m, quote, observation)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}
	evaluation := exitpolicy.RatchetSnapshotInput{
		Context: snapshotContext,
		Input: exitpolicy.RatchetInput{
			Entry:           m.state.EntryPrice,
			InitialStop:     m.state.InitialStop,
			ObservedPrice:   quote.Price,
			HighWater:       m.state.HighWater,
			Baseline:        m.state.Baseline,
			RealBreakeven:   breakEven,
			TakenRatioTotal: m.state.TakenRatioTotal,
			Level:           exitpolicy.Level(m.state.RatchetLevel),
			PendingAction:   exitpolicy.Action(m.state.PendingAction),
			Config:          &o.ratchet,
		},
	}
	snapshot, err := exitpolicy.EvaluateRatchetSnapshot(evaluation)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}
	o.clearRefused(m.position.ID)

	snapshot = snapshot.ChangedFromState(m.state.HighWater, m.state.Baseline,
		exitpolicy.Level(m.state.RatchetLevel), exitpolicy.NoRung)
	if !snapshot.Changed {
		// exit_events is append-only and the loop runs every five seconds. A
		// judgement that moved nothing is not a judgement worth a row; the
		// evaluator computes the condition (issues.md, task 0.1) so this does not
		// re-derive it.
		return nil
	}

	return o.record(ctx, m, snapshot,
		exitpolicy.NewRatchetRecoveryPolicy(evaluation),
		quote, observation, cycle)
}

func (o *ExitObserver) judgeLadder(ctx context.Context, m managed, quote observedQuote,
	observation cycleObservation, breakEven string, cycle *ExitCycle) error {
	ladder, err := o.ladderFor(m)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}
	if err := o.checkLadderPolicyStillFits(m, ladder); err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}
	snapshotContext, err := o.snapshotContext(m, quote, observation)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}
	pendingRung := exitpolicy.NoRung
	if m.state.PendingLevel != "" {
		if idx, err := exitpolicy.RungIndex(m.state.PendingLevel); err == nil {
			pendingRung = idx
		}
	}
	evaluation := exitpolicy.LadderSnapshotInput{
		Context: snapshotContext,
		Input: exitpolicy.LadderInput{
			EntryPrice:    m.state.EntryPrice,
			InitialStop:   m.state.InitialStop,
			ObservedPrice: quote.Price,
			HighWater:     m.state.HighWater,
			Baseline:      m.state.Baseline,
			State: exitpolicy.LadderState{
				// The journal snapshot selects the immutable registry policy above,
				// so the evaluator's identity check catches a mismatched state/table.
				PolicyID:        m.policyIdentity.ID,
				PolicyVersion:   m.policyIdentity.Version,
				PolicyDigest:    m.policyIdentity.Digest,
				ActivatedRung:   m.state.ActiveRung,
				TakenRatioTotal: m.state.TakenRatioTotal,
				Completed:       m.state.Completed,
				PendingAction:   exitpolicy.Action(m.state.PendingAction),
				PendingRung:     pendingRung,
			},
			Policy: ladder,
		},
	}
	snapshot, err := exitpolicy.EvaluateLadderSnapshot(evaluation)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return nil
	}
	o.clearRefused(m.position.ID)

	snapshot = snapshot.ChangedFromState(m.state.HighWater, m.state.Baseline,
		exitpolicy.Level(m.state.RatchetLevel), m.state.ActiveRung)
	if !snapshot.Changed {
		return nil
	}
	return o.record(ctx, m, snapshot, exitpolicy.NewLadderRecoveryPolicy(evaluation),
		quote, observation, cycle)
}

func (o *ExitObserver) snapshotContext(m managed, quote observedQuote,
	fallback cycleObservation) (exitpolicy.SnapshotContext, error) {
	observationID, err := stableObservationID(o.opts.AccountRef, m, quote, fallback)
	if err != nil {
		return exitpolicy.SnapshotContext{}, err
	}
	return exitpolicy.SnapshotContext{
		PositionID: m.position.ID, PositionGeneration: m.position.InstanceSeq,
		ObservationID: observationID, RemainingQuantity: m.position.Quantity,
	}, nil
}

func stableObservationID(accountRef string, m managed, quote observedQuote,
	fallback cycleObservation) (string, error) {
	price, err := riskcalc.CanonicalDecimal(quote.Price)
	if err != nil {
		return "", err
	}
	stampKind, stamp, sequence := "fetched_at", quote.FetchedAt.UTC().Format(time.RFC3339Nano), ""
	if quote.FetchedAt.IsZero() {
		stampKind, stamp = "cycle", fallback.at.UTC().Format(time.RFC3339Nano)
		sequence = strconv.FormatUint(fallback.sequence, 10)
	}
	h := sha256.New()
	for _, field := range []string{
		strings.TrimSpace(accountRef), strings.ToLower(strings.TrimSpace(m.position.Market)),
		strings.ToUpper(strings.TrimSpace(m.position.Symbol)), strings.TrimSpace(m.position.ID),
		strconv.FormatInt(m.position.InstanceSeq, 10), stampKind, stamp, sequence, price,
	} {
		_, _ = fmt.Fprintf(h, "%d:", len(field))
		_, _ = h.Write([]byte(field))
	}
	return "obs_" + fmt.Sprintf("%x", h.Sum(nil)), nil
}

func samePolicyIdentity(a, b exitpolicy.PolicyIdentity) bool {
	return strings.TrimSpace(a.ID) == strings.TrimSpace(b.ID) &&
		strings.TrimSpace(a.Version) == strings.TrimSpace(b.Version) &&
		strings.TrimSpace(a.Digest) == strings.TrimSpace(b.Digest)
}

// checkLadderPolicyStillFits is defence in depth after policy-ID resolution: an
// active rung must still exist and its lock must be no stronger than the
// monotone baseline the row already holds.
func (o *ExitObserver) checkLadderPolicyStillFits(m managed, ladder exitpolicy.LadderPolicy) error {
	rung := m.state.ActiveRung
	if rung < 0 {
		return nil
	}
	if rung >= len(ladder.Rungs) {
		return fmt.Errorf(
			"position %s stands on rung %d and the configured table has %d; the rung table was replaced "+
				"under a live position and its stored index no longer names the same rung",
			m.position.ID, rung, len(ladder.Rungs))
	}
	lock, err := exitpolicy.LockPrice(m.state.EntryPrice, ladder.Rungs[rung].StopPct)
	if err != nil {
		return err
	}
	cmp, err := riskcalc.CompareDecimal(lock, m.state.Baseline)
	if err != nil {
		return err
	}
	if cmp > 0 {
		return fmt.Errorf(
			"position %s stands on rung %d, whose configured lock %s is above the baseline %s it should "+
				"already have set; the rung table was replaced under a live position",
			m.position.ID, rung, lock, m.state.Baseline)
	}
	return nil
}

func (o *ExitObserver) ladderFor(m managed) (exitpolicy.LadderPolicy, error) {
	id := strings.TrimSpace(m.state.PolicyID)
	if id == "" || id == "default_v1" {
		if o.ladder.PolicyID != "default_v1" {
			return exitpolicy.LadderPolicy{}, fmt.Errorf(
				"position %s has legacy policy default_v1 but the injected test policy is %s",
				m.position.ID, o.ladder.PolicyID)
		}
		return o.ladder, nil
	}
	policy, err := exitpolicy.CommonLadderForPosition(id, m.position.Adopted())
	if err != nil {
		return exitpolicy.LadderPolicy{}, fmt.Errorf("position %s: %w", m.position.ID, err)
	}
	return policy, nil
}

// record persists the judgement and, when it proposed something, submits it.
//
// The ordering is the crash contract (apply_hook.go): the judgement transaction
// arms the proposal, then the intent is minted, attached and submitted. Reversed
// — submit first, record after — a crash in between leaves an order the ledger
// cannot explain and the next observation proposes the same level again.
func (o *ExitObserver) record(ctx context.Context, m managed, snapshot exitpolicy.ExitLineSnapshot,
	recoveryPolicy exitpolicy.RecoveryPolicyDefinition, quote observedQuote,
	observation cycleObservation, cycle *ExitCycle) error {
	proposal := snapshot.ExecutableProposal()
	quantity := snapshot.ProjectedQuantity
	orderable := snapshot.Orderable && !proposal.Zero()
	judgement := journal.ExitJudgement{
		Snapshot: snapshot, RecoveryPolicy: recoveryPolicy,
		PositionID: m.position.ID, LifecycleGeneration: m.state.LifecycleGeneration,
		ObservedPrice: snapshot.ObservedPrice,
		HighWater:     snapshot.HighWater, Baseline: snapshot.CurrentProtection,
		RatchetLevel: string(snapshot.RatchetLevel), ActiveRung: snapshot.ActiveRung,
		Provenance: journal.ExitDecisionProvenance{
			ObservationID: snapshot.ObservationID, SnapshotID: snapshot.SnapshotID,
			DecisionID: snapshot.DecisionID, Policy: snapshot.Policy,
		},
	}
	if quote.FetchedAt.IsZero() {
		judgement.ObservationSource, judgement.ObservedAt = "cycle", observation.at
	} else {
		judgement.ObservationSource, judgement.ObservedAt = "quote_fetched_at", quote.FetchedAt
	}

	// Clearing the symbol comes *before* the arming, not after it, and the reason
	// is that one of the two things it clears is an arming.
	//
	//   - The working entry buy is the E31 succession (issues.md, task 6.1): a
	//     liquidation submitted while a buy is still working can be answered by
	//     an entry fill during the close, which the projection refuses as a
	//     forbidden transition and turns into a RECONCILE.
	//   - The outstanding take-profit is what CancelPendingFirst names. Two
	//     working sells on one position can oversell it, and the arming refuses
	//     a second proposal while the first is outstanding — so the first has to
	//     be resolved, and it can only be resolved once its order is off the book.
	//
	// A clear that does not complete does not stop the judgement: the watermark
	// and the baseline still advance, and only the proposal is withheld. §0.3 is
	// why the withholding is noticed rather than silent — noteDelay is what turns
	// a delayed liquidation into an alert once it passes its bound.
	if orderable && (snapshot.CancelPendingFirst || isFullExit(proposal)) {
		cleared, err := o.clearTheSymbol(ctx, m, snapshot.CancelPendingFirst)
		if err != nil {
			return err
		}
		if !cleared {
			o.noteDelay(ctx, m, "a working order on the symbol could not be taken off the book")
			orderable = false
			judgement.ArmSuppressedReason = journal.ArmSuppressedWorkingOrder
		} else {
			o.clearDelay(m.position.ID)
		}
	}

	intentID := ""
	if orderable {
		intentID = exitIntentID(snapshot.DecisionID)
		if intentID == "" {
			intentID = o.opts.NewID()
		}
		judgement.Proposal = &journal.ExitProposal{
			Action:     string(proposal.Action),
			Level:      proposal.Level,
			IntentID:   intentID,
			Provenance: judgement.Provenance,
		}
	}

	recorded, err := o.opts.Journal.RecordExitJudgementResult(ctx, judgement)
	if err != nil {
		if errors.Is(err, journal.ErrProposalPending) {
			// Another proposal is outstanding. The evaluator suppresses this in
			// the normal path; reaching here means the row moved under the read,
			// and holding is the conservative answer.
			return nil
		}
		if errors.Is(err, journal.ErrExitSnapshotQuarantined) {
			// The transaction quarantined this generation and committed before
			// returning. Announcing here is what makes the creating cycle
			// observable at all — the working set only sees the quarantine on the
			// next cycle, and this is the path all three of the 2026-08-03
			// quarantines came through (change a064).
			//
			// The error is still returned unchanged: the judgement was not
			// recorded, and the announcement decides nothing about that.
			o.announceQuarantineFromLedger(ctx, m.position)
		}
		return fmt.Errorf("engine: recording the exit judgement of %s: %w", m.position.ID, err)
	}
	if recorded.ArmedProposal == nil || recorded.ArmOutcome != journal.ExitArmArmed {
		return nil
	}
	intentID = recorded.ArmedProposal.IntentID
	cycle.Proposed++
	return o.submit(ctx, m, proposal, quantity, intentID, judgement.ObservedPrice,
		judgement.Provenance)
}

func exitIntentID(decisionID string) string {
	decisionID = strings.TrimSpace(decisionID)
	if !strings.HasPrefix(decisionID, "eld_") || len(decisionID) != len("eld_")+sha256.Size*2 {
		return ""
	}
	return "exit_" + strings.TrimPrefix(decisionID, "eld_")
}

// isFullExit reports a proposal that intends to close the position outright.
func isFullExit(p exitpolicy.Proposal) bool {
	return p.Action == exitpolicy.ActionBaselineBreach || p.Action == exitpolicy.ActionLadderStop ||
		p.Action == exitpolicy.ActionLadderTakeProfit
}

// submit takes an armed proposal to the broker.
func (o *ExitObserver) submit(ctx context.Context, m managed, proposal exitpolicy.Proposal,
	quantity, intentID, observed string, provenance journal.ExitDecisionProvenance) error {
	submitQuantity, capped, err := o.applyFloor(ctx, m, quantity)
	if err != nil {
		return err
	}
	if isZeroQuantity(submitQuantity) {
		// The confirmed floor authorises nothing — an absent or stale snapshot
		// while the account is in RECONCILE. Release the proposal so the same
		// level identity is proposed again once the floor lifts; keeping it armed
		// would suppress exactly the re-proposal exit-policy asks for.
		return o.release(ctx, m, journal.ProposalRefused)
	}

	issued, err := o.opts.Issuer.IssueReduction(ctx, execgw.ReductionIssuance{
		Intent: risk.Intent{
			AccountRef: o.opts.AccountRef,
			Market:     costs.Market(strings.ToLower(strings.TrimSpace(m.position.Market))),
			Symbol:     m.position.Symbol,
			Side:       risk.SideSell,
			Quantity:   submitQuantity,
		},
		Account: risk.AccountState{HeldQuantity: m.position.Quantity},
		Reason: fmt.Sprintf("exit policy %s at level %s (position %s)",
			proposal.Action, proposal.Level, m.position.ID),
	})
	if err != nil {
		o.alertProposalRefused(ctx, m, proposal, err.Error())
		return o.release(ctx, m, journal.ProposalRefused)
	}

	// Attach before submitting. The id is already in the judgement transaction;
	// this is the belt to that braces — a re-read of the row before an order
	// exists, so an attach that fails stops the submission rather than leaving an
	// order nothing can resolve.
	if err := o.opts.Journal.AttachExitIntent(ctx, m.position.ID, intentID); err != nil {
		return fmt.Errorf("engine: attaching the exit intent of %s: %w", m.position.ID, err)
	}

	intent, err := o.sellIntent(m, submitQuantity, observed)
	if err != nil {
		o.alertRefused(ctx, m, err)
		return o.release(ctx, m, journal.ProposalRefused)
	}
	out, err := o.opts.Submit.Place(ctx, execgw.PlaceRequest{
		Intent:         intent,
		Decision:       issued.Decision,
		IntentID:       intentID,
		ExitProvenance: &provenance,
	})
	switch {
	case out.State == journal.StateConfirmed:
		o.log(obs.EventOrderConfirmed, false,
			obs.FieldSymbol, m.position.Symbol,
			obs.FieldQuantity, submitQuantity,
			"action", string(proposal.Action),
			"level", proposal.Level,
			"capped", capped)
		return nil
	case out.State == journal.StateInDoubt || out.State == journal.StateUnresolvedInDoubt:
		// The order may exist. The proposal stays armed and the resolver settles
		// it; releasing here would let the next observation submit a second sell
		// on top of one that may already be live.
		return nil
	case out.Reason == execgw.ReasonSymbolInFlight:
		o.noteDelay(ctx, m, "another mutation on the symbol is in flight")
		return o.release(ctx, m, journal.ProposalCancelled)
	default:
		detail := out.Detail
		if detail == "" && err != nil {
			detail = err.Error()
		}
		o.alertProposalRefused(ctx, m, proposal, detail)
		return o.release(ctx, m, journal.ProposalRefused)
	}
}

// release clears an armed proposal that produced no live order, which re-arms
// its level (exit-policy: 거부·취소 시 해당 레벨은 재발의 가능해진다).
func (o *ExitObserver) release(ctx context.Context, m managed, how journal.ProposalResolution) error {
	if err := o.opts.Journal.ResolveExitProposal(ctx, m.position.ID, how); err != nil {
		return fmt.Errorf("engine: releasing the exit proposal of %s: %w", m.position.ID, err)
	}
	return nil
}

// clearTheSymbol takes the working orders that would collide with a liquidation
// off the book, and reports whether the symbol is now clear.
//
// withPending widens it from "the working entry" to "every working order plus
// the proposal that owns one": that is CancelPendingFirst, and it ends by
// releasing the outstanding proposal, because the arming refuses a second one
// while the first is still recorded.
//
// The release is last and only on success. Releasing a proposal whose order is
// still live would re-arm its level over a working sell, which is the oversell
// the whole sequence exists to prevent.
func (o *ExitObserver) clearTheSymbol(ctx context.Context, m managed, withPending bool) (bool, error) {
	live, err := o.opts.Journal.LiveOrdersForSymbol(ctx,
		o.opts.AccountRef, m.position.Market, m.position.Symbol)
	if err != nil {
		return false, fmt.Errorf("engine: reading the working orders of %s: %w", m.position.Symbol, err)
	}
	clear := true
	for _, order := range live {
		buy := strings.EqualFold(strings.TrimSpace(order.Side), "BUY")
		if !buy && !withPending {
			continue
		}
		issued, err := o.opts.Issuer.IssueReduction(ctx, execgw.ReductionIssuance{
			Intent: risk.Intent{
				AccountRef: o.opts.AccountRef,
				Market:     costs.Market(strings.ToLower(strings.TrimSpace(order.Market))),
				Symbol:     order.Symbol,
				Side:       risk.SideSell,
				Quantity:   order.Quantity,
			},
			Account: risk.AccountState{HeldQuantity: order.Quantity},
			Reason: fmt.Sprintf("cancelling the working %s order %s before liquidating position %s",
				strings.ToUpper(strings.TrimSpace(order.Side)), order.OrderID, m.position.ID),
		})
		if err != nil {
			clear = false
			continue
		}
		quantity, qerr := floatOf("working order quantity", order.Quantity)
		price, perr := floatOf("working order price", order.Price)
		if qerr != nil || perr != nil {
			clear = false
			continue
		}
		out, err := o.opts.Submit.Cancel(ctx, execgw.CancelRequest{
			Intent: orderintent.CancelIntent{OrderID: order.OrderID, Symbol: order.Symbol},
			Order: execgw.OrderRef{
				Market:   order.Market,
				Side:     order.Side,
				Quantity: quantity,
				Price:    price,
				Currency: order.Currency,
			},
			Decision: issued.Decision,
		})
		if err != nil || out.State != journal.StateConfirmed {
			clear = false
		}
	}
	if !clear {
		return false, nil
	}
	if withPending && m.state.Pending() {
		if err := o.release(ctx, m, journal.ProposalCancelled); err != nil {
			return false, err
		}
	}
	return true, nil
}

// applyFloor caps a liquidation at the RECONCILE confirmed floor (task 7.5).
//
// The remainder is not remembered in this process, and that is the design rather
// than an omission. exit-policy asks that a capped remainder be re-proposed with
// the same level identity once the disagreement is resolved; a remainder held in
// memory would be lost by exactly the restart most likely to follow a
// reconciliation incident. What is durable is the *condition* — the baseline is
// still breached, the rung is still unfilled — and the next observation
// re-derives the same proposal from it, under the same level, from the ledger.
func (o *ExitObserver) applyFloor(ctx context.Context, m managed, quantity string) (string, bool, error) {
	if o.opts.Floor == nil {
		return quantity, false, nil
	}
	floor, applies, err := o.opts.Floor.ConfirmedFloor(ctx, m.position.Market, m.position.Symbol)
	if err != nil {
		// A floor that cannot be computed is a floor of zero, not an absent one:
		// the account is in RECONCILE and the engine has just failed to establish
		// what it may sell.
		o.logErr(obs.EventExitProposalCapped, err,
			"the confirmed floor of "+m.position.Symbol+" could not be computed; nothing is submitted")
		return "0", true, nil
	}
	if !applies {
		return quantity, false, nil
	}
	cmp, err := riskcalc.CompareDecimal(floor.Quantity, quantity)
	if err != nil {
		return "", false, fmt.Errorf("engine: comparing the confirmed floor of %s: %w", m.position.Symbol, err)
	}
	if cmp >= 0 {
		return quantity, false, nil
	}
	remainder, err := riskcalc.SubDecimal(quantity, floor.Quantity)
	if err != nil {
		return "", false, fmt.Errorf("engine: sizing the capped remainder of %s: %w", m.position.Symbol, err)
	}
	o.alert(ctx, obs.Event{
		Type:  obs.EventExitProposalCapped,
		Key:   string(obs.EventExitProposalCapped) + "|" + m.position.ID,
		Title: "the liquidation of " + m.position.Symbol + " was capped by the confirmed floor",
		Body: fmt.Sprintf(
			"%s of %s was proposed, the RECONCILE confirmed floor authorises %s (%s), and %s stays "+
				"unsold; the same level is re-proposed once the disagreement is resolved",
			quantity, m.position.Symbol, floor.Quantity, floor.Bound, remainder),
		Fields: map[string]any{
			obs.FieldSymbol:   m.position.Symbol,
			obs.FieldQuantity: floor.Quantity,
			"proposed":        quantity,
			"remainder":       remainder,
			"floor_bound":     floor.Bound,
		},
	})
	return floor.Quantity, true, nil
}

// sellIntent renders the order a proposal becomes.
//
// The limit price is the **observed price** — the one this judgement was made
// from — and not the baseline. Automated orders are LIMIT only (riskcalc's rule,
// kept on the exit side because a market sell has no price the ledger can record
// an intent against), and a limit has to be marketable to be an exit at all: a
// breach means the price is already *below* the baseline, so a sell limited at
// the baseline would rest above the market and protect nothing. Pricing at the
// observation makes the order marketable in both directions — a take-profit is
// observed above its trigger and a liquidation below its baseline — and it is
// the number the judgement itself is explained by.
//
// The residual is named rather than hidden: between the observation and the
// broker receiving the order the price can move further, and a limit does not
// chase it. That is the same sampling window design.md lists as the risk 2c's
// broker-resident stop closes, not a new one.
func (o *ExitObserver) sellIntent(m managed, quantity, observed string) (orderintent.PlaceIntent, error) {
	price := strings.TrimSpace(observed)
	if price == "" {
		price = strings.TrimSpace(m.state.Baseline)
	}
	if price == "" {
		return orderintent.PlaceIntent{}, fmt.Errorf(
			"position %s has no price to submit a liquidation at", m.position.ID)
	}
	qty, err := floatOf("liquidation quantity", quantity)
	if err != nil {
		return orderintent.PlaceIntent{}, err
	}
	limit, err := floatOf("liquidation limit price", price)
	if err != nil {
		return orderintent.PlaceIntent{}, err
	}
	return orderintent.PlaceIntent{
		Symbol:       m.position.Symbol,
		Market:       m.position.Market,
		Side:         "sell",
		OrderType:    "limit",
		Quantity:     qty,
		Price:        limit,
		CurrencyMode: currencyFor(m.position.Market),
	}, nil
}

// --- alerts -------------------------------------------------------------------

func (o *ExitObserver) alertUnmanaged(ctx context.Context, p journal.Position) {
	if o.unmanaged[p.ID] {
		return
	}
	o.unmanaged[p.ID] = true
	o.alert(ctx, obs.Event{
		Type:  obs.EventExitPositionUnmanaged,
		Key:   string(obs.EventExitPositionUnmanaged) + "|" + p.ID,
		Title: p.Symbol + " is held with no entry decision, so the exit policy will not manage it",
		Body: "the position carries no entry decision, so there is no stop to build a baseline out of and " +
			"no initial risk to measure R against; it is reported and left alone",
		Fields: map[string]any{
			obs.FieldSymbol:   p.Symbol,
			obs.FieldQuantity: p.Quantity,
			"position_id":     p.ID,
		},
	})
}

// alertRefused reports a refusal to *judge*, which exit-policy separates from a
// judgement that decided nothing (판정 거부 + 알림).
func (o *ExitObserver) alertRefused(ctx context.Context, m managed, cause error) {
	if o.refused[m.position.ID] {
		return
	}
	o.refused[m.position.ID] = true
	field := ""
	var refusal *exitpolicy.RefusalError
	if errors.As(cause, &refusal) {
		field = refusal.Field
	}
	o.alert(ctx, obs.Event{
		Type:  obs.EventExitJudgementRefused,
		Key:   string(obs.EventExitJudgementRefused) + "|" + m.position.ID,
		Title: "the exit policy could not judge " + m.position.Symbol,
		Body: "the stored protection state or the observed price is not usable, so this position is not " +
			"being judged at all: " + cause.Error(),
		Fields: map[string]any{
			obs.FieldSymbol: m.position.Symbol,
			"position_id":   m.position.ID,
			"invariant":     field,
		},
	})
}

func (o *ExitObserver) clearRefused(positionID string) { delete(o.refused, positionID) }

func (o *ExitObserver) alertProposalRefused(ctx context.Context, m managed,
	proposal exitpolicy.Proposal, detail string) {
	o.alert(ctx, obs.Event{
		Type: obs.EventExitProposalRefused,
		Key: string(obs.EventExitProposalRefused) + "|" + m.position.ID + "|" +
			string(proposal.Action) + "|" + proposal.Level,
		Title: "the exit proposal for " + m.position.Symbol + " was not submitted",
		Body: fmt.Sprintf("%s at level %s was refused before it reached the broker: %s",
			proposal.Action, proposal.Level, detail),
		Fields: map[string]any{
			obs.FieldSymbol: m.position.Symbol,
			"position_id":   m.position.ID,
			"action":        string(proposal.Action),
			"level":         proposal.Level,
			obs.FieldDetail: detail,
		},
	})
}

// noteDelay starts (or continues) the in-flight delay clock for one symbol and
// raises the critical alert once the delay passes the bound.
func (o *ExitObserver) noteDelay(ctx context.Context, m managed, why string) {
	now := o.clk.Now()
	since, running := o.delayedSince[m.position.ID]
	if !running {
		o.delayedSince[m.position.ID] = now
		return
	}
	if now.Sub(since) < o.delayBound() || o.delayAlerted[m.position.ID] {
		return
	}
	o.delayAlerted[m.position.ID] = true
	o.alert(ctx, obs.Event{
		Type:  obs.EventExitLiquidationDelayed,
		Key:   string(obs.EventExitLiquidationDelayed) + "|" + m.position.ID,
		Title: "the liquidation of " + m.position.Symbol + " has been delayed past its bound",
		Body: fmt.Sprintf("%s, and it has been %s: until a broker-resident stop exists this delay is "+
			"unprotected exposure", why, now.Sub(since).Round(time.Second)),
		Fields: map[string]any{
			obs.FieldSymbol: m.position.Symbol,
			"position_id":   m.position.ID,
			"delay_seconds": int(now.Sub(since).Seconds()),
			obs.FieldDetail: why,
		},
	})
}

func (o *ExitObserver) clearDelay(positionID string) {
	delete(o.delayedSince, positionID)
	delete(o.delayAlerted, positionID)
}

func (o *ExitObserver) alert(ctx context.Context, e obs.Event) {
	if o.opts.Alerts == nil {
		return
	}
	if err := o.opts.Alerts.Notify(ctx, e); err != nil {
		o.logErr(e.Type, err, "the alert could not be made durable")
	}
}

func (o *ExitObserver) log(t obs.EventType, warn bool, args ...any) {
	if o.opts.Log == nil {
		return
	}
	if warn {
		o.opts.Log.Warn(t, args...)
		return
	}
	o.opts.Log.Event(t, args...)
}

func (o *ExitObserver) logErr(t obs.EventType, err error, detail string) {
	if o.opts.Log == nil {
		return
	}
	o.opts.Log.Error(t, err, obs.FieldAccount, o.opts.AccountRef, obs.FieldDetail, detail)
}

// --- small helpers -------------------------------------------------------------

func symbolsOf(states []managed) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(states))
	for _, m := range states {
		symbol := strings.ToUpper(strings.TrimSpace(m.position.Symbol))
		if symbol == "" || seen[symbol] {
			continue
		}
		seen[symbol] = true
		out = append(out, symbol)
	}
	sort.Strings(out)
	return out
}

// proposalQuantity turns a ratio of the remaining quantity into a whole number
// of units.
//
// It rounds **down**, and the arithmetic is math/big rational rather than
// float64 for the reason internal/risk gives: the boundary of "40 % of 5" is a
// rule, and a boundary decided by a binary expansion is not the rule anybody
// wrote. Rounding down is the conservative direction for a *partial*: selling
// more than the policy asked for would take protection the position was meant to
// keep, and the last unit belongs to the full liquidation that eventually comes.
func proposalQuantity(remaining, ratio string) (string, error) {
	return exitpolicy.ProjectWholeShares(remaining, ratio)
}

func isZeroQuantity(q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return true
	}
	zero, err := riskcalc.CompareDecimal(q, "0")
	return err != nil || zero <= 0
}

// decimalOf renders a broker float as the decimal string the ledger stores.
func decimalOf(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

// floatOf is the reverse, for the order intents upstream still expresses in
// float64. A value the engine itself produced always parses; one that does not
// is a refusal rather than a zero, because a quantity of zero is an order that
// silently does nothing.
func floatOf(field, s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("engine: %s %q is not a number: %w", field, s, err)
	}
	return v, nil
}

func currencyFor(market string) string {
	if strings.EqualFold(strings.TrimSpace(market), "us") {
		return "USD"
	}
	return "KRW"
}

func randomExitID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand does not fail on the platforms this ships to, and an id
		// derived from a weaker source would be an id an operator could guess.
		panic("engine: reading randomness for an exit intent id: " + err.Error())
	}
	return "exit-" + hex.EncodeToString(b[:])
}
