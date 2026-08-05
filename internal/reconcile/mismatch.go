package reconcile

// mismatch.go decides what a disagreement means and how long it lasts
// (harden-execution-base task 3.6).
//
// # Entries close, exits never do
//
// Every block in this file is an *entry* block. Nothing here can stop a cancel or
// a liquidation, and that is not an oversight to be tidied up later — it is
// §0.3. The failure being guarded against is "the engine adds exposure it cannot
// account for". Trapping the engine in a position it cannot exit would replace
// that failure with a worse one, on the same account, at a worse moment.
//
// # Why retrying has a limit
//
// Most disagreements are timing: a fill that landed between two reads, a
// settlement that has not propagated. Those disappear on the next reconciliation,
// which is why the first response is to wait 30 seconds and look again rather
// than to page a human.
//
// A disagreement that survives three of those is a different animal. Whatever it
// is, "look again" has now been shown not to fix it, and continuing to loop would
// be an engine that reports a problem forever and does nothing about it. So the
// third consecutive failure is terminal: it is marked permanent, escalated to
// account scope, and only an operator clears it. A success at any point resets
// the counter, because the counter is measuring *consecutive* failure — the thing
// that distinguishes a stuck account from a busy one.
//
// # The state table
//
// The spec asks for the block scopes and release conditions to be defined "as a
// state table with the reason codes". BlockRules() below is that table, as data
// rather than as a comment, and a test asserts every reason this package can
// produce has a row in it.
//
//	condition                          reason code                        scope    auto release          manual release
//	---------------------------------  ---------------------------------  -------  --------------------  ---------------
//	restart recovery unfinished        recovery_incomplete                account  recovery completes    —
//	quantity disagreement              reconciliation_mismatch            symbol   adjusted reconcile    operator
//	order the account does not show    reconciliation_mismatch            symbol   adjusted reconcile    operator
//	3 consecutive failed reconciles    reconciliation_mismatch_permanent  account  never                 operator
//	broker snapshot contradiction      unknown_broker_state               symbol   never                 operator
//
// # "Adjusted reconcile", and why an agreeing read is not enough
//
// The two symbol rows used to say "clean reconcile": the next comparison that
// agreed released the block. Task 6.3 narrowed that to what the reconciliation
// delta requires — 비영구 차단의 자동 해제는 조정 이벤트가 반영된 뒤의 재조회
// 일치에만 근거하며(SHALL) … 조정 없이 우연히 일치한 단발 관측은 [해제하지
// 못한다](SHALL NOT).
//
// The distinction is between a reading and a cause. "The account and the engine
// now agree" can be true because the disagreement was settled, and it can be
// true because one poll happened to catch a moment when the difference was
// invisible — a fill in flight, a settlement mid-propagation, a snapshot taken
// between two halves of a round trip. Releasing on the reading treats those two
// identically, and the second one re-opens entries onto an exposure the engine
// still cannot account for.
//
// So a release needs two things: something wrote an adjustment that converged
// the projection to the account's value (AdjustmentApplied below, called by
// whatever applied it), and a *later* comparison agreed. The release is recorded
// with cause ADJUSTMENT_APPLIED, so the history says which of the two facts
// closed the block.
//
// # Which re-read, and which symbol (a083)
//
// "The re-read after the adjustment" has two words the first implementation
// could not check, and both of them cost the rule its reachability.
//
// *After*: a credit carries the as-of of the comparison it was computed from,
// and only an observation of a strictly later comparison may spend it. Without
// the stamp, "the next observation" was the approximation — and the driver's
// cycle converges a disagreeing comparison and then observes that same
// comparison, so the credit was spent by the read it was answering, before any
// re-read existed. ADJUSTMENT_APPLIED was unreachable for the whole of its life:
// the production ledger recorded zero of them against eight blocks that never
// lifted and six an operator had to clear by hand.
//
// *For that symbol*: expiry is judged per symbol, not per comparison. A credit
// dies when a later comparison still disagrees about **its** symbol. An
// unrelated symbol's disagreement is not an answer to it, and discarding the
// credit over one would strand a symbol that has already converged and agreed —
// nothing writes an adjustment for a symbol the comparison agrees about, so it
// could never earn another credit.
//
// Both directions fail closed. An as-of that is missing, unreadable, or out of
// order keeps the block, because a block that stays costs entries and a block
// that lifts early costs exposure the engine cannot account for.
//
// What this costs: a disagreement that resolves itself with nothing written
// stays blocked until an operator clears it. That is deliberate and it is the
// conservative direction (§0.9) — the loop's own answer to a quantity
// disagreement is to converge the projection to the account, so the ordinary
// path always writes something.
//
// # The scope caveat
//
// The symbol-scoped rows are recorded at symbol scope here and *also* latch the
// account-wide entry gate, because execgw.EntryGate has no symbol dimension yet
// (issues.md, 2026-07-26). Blocking wider than the table says is the safe
// direction — it stops more trading, not less — and narrowing it is wiring work
// for task 4.2. EntryAllowed below is the symbol-aware check that wiring will
// call.
//
// # The journal is the authority (extend-execution-contract task 4.1)
//
// Everything above describes what this tracker *decides*. What it decides is no
// longer where the truth lives: an active disagreement is a row in the
// journal's reconcile_states, and the block set below is a projection of those
// rows. Observe writes through to the journal, Restore rebuilds the projection
// from it, and a restart therefore comes back still blocked — which is the
// spec's "재시작이 차단을 잃어서는 안 된다"(SHALL).
//
// The tracker only ever releases states it entered itself (cause
// QUANTITY_MISMATCH, via ExpectCause). A clean quantity comparison is not
// evidence about an identifier conflict somebody else recorded, and releasing
// one with the other is how a block silently disappears.

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// Scope is how wide a block reaches.
type Scope string

const (
	// ScopeAccount blocks every new entry on the account.
	ScopeAccount Scope = "account"
	// ScopeMarket blocks new entries in one market.
	ScopeMarket Scope = "market"
	// ScopeSymbol blocks new entries in one symbol.
	ScopeSymbol Scope = "symbol"
)

// Release names how a block goes away.
type Release string

const (
	// ReleaseOnAdjustedReconcile — an adjustment converged the projection and the
	// re-read after it agreed (add-core-domain task 6.3).
	//
	// It replaced a plain "the next agreeing reconciliation clears it", and the
	// difference is the whole of the delta's release rule: an agreeing read is a
	// reading, and a block is released by a cause. See the state table below.
	ReleaseOnAdjustedReconcile Release = "adjusted_reconcile"
	// ReleaseOnRecoveryComplete — finishing the restart sequence clears it.
	ReleaseOnRecoveryComplete Release = "recovery_complete"
	// ReleaseOperatorOnly — no automatic path out.
	ReleaseOperatorOnly Release = "operator_only"
)

// BlockRule is one row of the state table.
type BlockRule struct {
	Condition string
	Reason    execgw.ReasonCode
	Scope     Scope
	Auto      Release
	// Manual reports whether an operator can clear it by hand.
	Manual bool
}

// BlockRules returns the state table the spec requires, as data.
//
// Keeping it as a value rather than as documentation means the rules can be
// asserted against: a reason code the tracker can emit but that has no row here
// is a block nobody has decided how to release, which is how an engine ends up
// permanently and inexplicably shut.
func BlockRules() []BlockRule {
	return []BlockRule{
		{
			Condition: "the restart recovery sequence has not finished",
			Reason:    execgw.ReasonRecoveryIncomplete,
			Scope:     ScopeAccount,
			Auto:      ReleaseOnRecoveryComplete,
			Manual:    false,
		},
		{
			Condition: "the account and the engine disagree about a quantity",
			Reason:    execgw.ReasonReconcileMismatch,
			Scope:     ScopeSymbol,
			Auto:      ReleaseOnAdjustedReconcile,
			Manual:    true,
		},
		{
			Condition: "reconciliation failed the configured number of times in a row",
			Reason:    execgw.ReasonReconcilePermanent,
			Scope:     ScopeAccount,
			Auto:      ReleaseOperatorOnly,
			Manual:    true,
		},
		{
			Condition: "a broker snapshot contradicted what was already observed",
			Reason:    execgw.ReasonBrokerStateUnknown,
			Scope:     ScopeSymbol,
			Auto:      ReleaseOperatorOnly,
			Manual:    true,
		},
	}
}

// RuleFor returns the state-table row for a reason code.
func RuleFor(reason execgw.ReasonCode) (BlockRule, bool) {
	for _, rule := range BlockRules() {
		if rule.Reason == reason {
			return rule, true
		}
	}
	return BlockRule{}, false
}

// Block is one active entry block.
type Block struct {
	Scope   Scope
	Account string
	Market  string
	Symbol  string
	// Cause is the exact journal RECONCILE cause that owns this block. Reason is
	// the coarser execution-gate category; Cause preserves producer authority so
	// a quantity comparison cannot release another producer's state.
	Cause  string
	Reason execgw.ReasonCode
	Detail string
	Since  time.Time
	// Release is the state table's automatic release condition.
	Release Release
	// Permanent reports that no automatic path clears this block.
	Permanent bool
	// pending reports that the block is conservative in-memory evidence whose
	// durable EnterReconcile write has not committed yet. It stays enforced and
	// is retried on every matching observation until the journal confirms it.
	pending bool
}

// Key identifies a block for deduplication.
func (b Block) Key() string {
	switch b.Scope {
	case ScopeAccount:
		return string(b.Scope) + "|" + b.Account
	case ScopeMarket:
		return string(b.Scope) + "|" + b.Account + "|" + strings.ToLower(strings.TrimSpace(b.Market))
	case ScopeSymbol:
		return string(b.Scope) + "|" + b.Account + "|" + strings.ToUpper(strings.TrimSpace(b.Symbol))
	default:
		return string(b.Scope) + "|" + b.Account + "|" + b.Market + "|" + b.Symbol
	}
}

// Covers reports whether this block stops a new entry in a market and symbol.
func (b Block) Covers(market, symbol string) bool {
	switch b.Scope {
	case ScopeAccount:
		return true
	case ScopeMarket:
		return strings.EqualFold(b.Market, market)
	case ScopeSymbol:
		return strings.EqualFold(b.Symbol, symbol)
	default:
		return true
	}
}

// DefaultReconcileInterval is the minimum wait before reconciling again
// (reconciliation spec: 기본 30초).
const DefaultReconcileInterval = 30 * time.Second

// DefaultMaxFailures is how many consecutive failures make a mismatch permanent.
const DefaultMaxFailures = 3

// ReconcileStore is the durable half of the RECONCILE state.
//
// It is an interface rather than *journal.Journal so a test can drive the
// tracker without a database, and so the coupling is visibly three methods
// wide. *journal.Journal satisfies it.
type ReconcileStore interface {
	EnterReconcile(ctx context.Context, req journal.EnterReconcileRequest) (journal.ReconcileState, bool, error)
	ReleaseReconcile(ctx context.Context, req journal.ReleaseReconcileRequest) (journal.ReconcileState, bool, error)
	ReleaseReconciles(ctx context.Context, reqs []journal.ReleaseReconcileRequest) ([]journal.ReconcileState, error)
	ActiveReconcileStates(ctx context.Context) ([]journal.ReconcileState, error)
}

// Tracker holds the mismatch state across reconciliations.
type Tracker struct {
	// Clock drives the interval. Defaults to clock.System().
	Clock clock.Clock
	// Gate is latched while any block is active. Optional but expected: without
	// it the tracker reports blocks nobody enforces.
	Gate *execgw.EntryGate
	// Journal is where the blocks actually live. Optional only so unit tests of
	// the promotion arithmetic can run without a database — a production
	// tracker without one forgets every disagreement on restart, which is the
	// failure task 4.1 exists to remove.
	Journal ReconcileStore
	// MinInterval is the minimum wait between reconciliations. Zero takes
	// DefaultReconcileInterval.
	MinInterval time.Duration
	// MaxFailures is how many consecutive failures make a mismatch permanent.
	// Zero takes DefaultMaxFailures.
	MaxFailures int
	AccountRef  string

	mu        sync.Mutex
	failures  int
	lastAt    time.Time
	observed  bool
	blocks    map[string]Block
	permanent bool
	// adjusted maps a symbol to the as-of of the *comparison the adjustment was
	// computed from*. It is what turns an agreeing comparison from a reading into
	// a release, and the stamp is what makes "the re-read after the adjustment"
	// something this type can check rather than something its caller has to
	// arrange (a083).
	adjusted map[string]string
}

// AdjustmentApplied records that an adjustment converged the projection of a
// symbol to the account's value.
//
// It is the first half of the delta's release rule; the second half is the
// comparison that follows. The caller is whatever applied the adjustment — the
// external-position ingest in this package, or the loop that converges a
// quantity disagreement — because the tracker cannot tell from a diff whether
// anything was written.
//
// # comparison, and why it is not optional
//
// comparison is the as-of of the diff the adjustment was computed from. Without
// it the rule "the re-read *after* the adjustment" has no way to tell a re-read
// from the read it is answering, and the driver hands both to Observe: it
// converges a disagreeing diff and then observes that same diff in the same
// cycle. A credit that any following observation could spend is therefore spent
// by the pre-adjustment comparison, before the re-read exists — which made
// ADJUSTMENT_APPLIED unreachable in production for the whole of its life
// (a083; the ledger recorded zero of them and eight blocks that never lifted).
//
// Requiring the argument means a caller cannot fail to say which comparison it
// is standing on. An unparseable or empty one is never spendable, which keeps
// the block — the conservative direction.
//
// Deliberately in-memory and deliberately not restored. A restart comes back
// with the blocks (they are journal rows) and without the credits, so the first
// comparison after a restart cannot release anything: the process that applied
// the adjustment is gone, and the conservative direction is to make the next
// convergence say so again.
func (t *Tracker) AdjustmentApplied(comparison string, symbols ...string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.adjusted == nil {
		t.adjusted = map[string]string{}
	}
	for _, symbol := range symbols {
		key := strings.ToUpper(strings.TrimSpace(symbol))
		if key == "" {
			continue
		}
		t.adjusted[key] = strings.TrimSpace(comparison)
	}
}

// creditUsableBy reports whether a credit recorded against comparison `credit`
// may be spent by an observation of comparison `observed`.
//
// Strictly later, and nothing else. Equal means this is the very comparison the
// adjustment was computed from, and an observation of it is not a re-read —
// that equality is the whole of the a083 defect. Earlier means the two cannot be
// ordered the way the rule needs. Either side missing or unreadable means the
// order cannot be established at all. All three keep the block, which is the
// conservative direction: a block that stays costs entries, and a block that
// lifts early costs exposure the engine cannot account for.
func creditUsableBy(credit, observed string) bool {
	creditAt, err := time.Parse(time.RFC3339, strings.TrimSpace(credit))
	if err != nil {
		return false
	}
	observedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(observed))
	if err != nil {
		return false
	}
	return observedAt.After(creditAt)
}

// symbolsInDispute is the set of symbols this comparison still disagrees about.
//
// It is what makes credit expiry per symbol rather than per comparison. A credit
// is answered by what the re-read says about *its* symbol; an unrelated symbol's
// disagreement is not an answer to it, and discarding the credit over one would
// strand a symbol that has converged and agreed with no way to earn another
// credit — nothing writes an adjustment for a symbol the comparison agrees about.
func symbolsInDispute(diff Diff) map[string]bool {
	out := make(map[string]bool, len(diff.Quantities)+len(diff.MissingOrders))
	for _, mismatch := range diff.Quantities {
		out[strings.ToUpper(strings.TrimSpace(mismatch.Symbol))] = true
	}
	for _, missing := range diff.MissingOrders {
		out[strings.ToUpper(strings.TrimSpace(missing.Symbol))] = true
	}
	return out
}

// Outcome is what one observation did.
type Outcome struct {
	// Failures is the consecutive-failure count after this observation.
	Failures int
	// Permanent reports that the mismatch is now permanent.
	Permanent bool
	// Blocked reports whether any block is active after this observation.
	Blocked bool
	// Added are blocks this observation raised.
	Added []Block
	// Cleared are blocks this observation released, with cause
	// ADJUSTMENT_APPLIED.
	Cleared []Block
	// AwaitingAdjustment are blocks a comparison agreed about and that stayed
	// anyway, because nothing has converged their symbol's projection.
	//
	// They are reported rather than silent because "the account agrees and the
	// engine is still blocked" is the state an operator will ask about, and the
	// answer is that agreement is not what releases.
	AwaitingAdjustment []Block
	// NextDueAt is the earliest instant a re-reconciliation may run.
	NextDueAt time.Time
}

// Observe records the result of one reconciliation.
//
// It takes a context and returns an error because the block set it computes is
// written through to the journal before it is reported: a promotion that only
// exists in memory is one a restart loses. A write failure is returned with the
// outcome, and the caller must treat it as "the block may not be durable" —
// which for an entry gate means staying blocked, not proceeding.
func (t *Tracker) Observe(ctx context.Context, diff Diff) (Outcome, error) {
	now := t.clock().Now()
	interval := t.interval()
	maxFailures := t.maxFailures()

	t.mu.Lock()
	if t.blocks == nil {
		t.blocks = map[string]Block{}
	}
	t.lastAt = now
	t.observed = true

	var out Outcome
	if !diff.BlocksEntry() {
		// A success resets the counter — the counter measures *consecutive*
		// failure, which is what separates a stuck account from a busy one. The
		// counter and the block are separate decisions: the counter is about how
		// the reconciliation *process* is doing, the block is about whether the
		// engine knows its exposure, and only the second one needs a cause.
		t.failures = 0
		for _, block := range t.blocks {
			if block.Cause != journal.ReconcileCauseQuantityMismatch ||
				block.Release != ReleaseOnAdjustedReconcile {
				// The quantity comparer owns only QUANTITY_MISMATCH. Snapshot and
				// attribution producers may use the same coarse gate reason, but an
				// agreeing quantity read is not evidence that their state is over.
				continue
			}
			symbol := strings.ToUpper(strings.TrimSpace(block.Symbol))
			if !creditUsableBy(t.adjusted[symbol], diff.AsOf) {
				// Either nothing was written for this symbol, or what was written
				// belongs to *this* comparison rather than to one before it. Both are
				// agreement with nothing behind it. See the release rule at the top of
				// the file.
				out.AwaitingAdjustment = append(out.AwaitingAdjustment, block)
				continue
			}
			out.Cleared = append(out.Cleared, block)
		}
	} else {
		t.failures++
		for _, block := range blocksFor(diff, t.AccountRef, now) {
			if _, exists := t.blocks[block.Key()]; !exists {
				block.pending = true
				t.blocks[block.Key()] = block
				out.Added = append(out.Added, block)
			}
		}
		if t.failures >= maxFailures && !t.permanent {
			permanent := Block{
				Scope:     ScopeAccount,
				Account:   t.AccountRef,
				Cause:     journal.ReconcileCauseQuantityMismatch,
				Reason:    execgw.ReasonReconcilePermanent,
				Since:     now,
				Release:   ReleaseOperatorOnly,
				Permanent: true,
				Detail: fmt.Sprintf(
					"reconciliation disagreed %d times in a row; looking again has been shown not to resolve it",
					t.failures),
			}
			if _, exists := t.blocks[permanent.Key()]; !exists {
				permanent.pending = true
				t.permanent = true
				t.blocks[permanent.Key()] = permanent
				out.Added = append(out.Added, permanent)
			}
		}
	}

	sortBlocks(out.Added)
	sortBlocks(out.Cleared)
	sortBlocks(out.AwaitingAdjustment)

	// Additions are already exposed conservatively in memory. A failed enter is
	// kept pending and retried here on the next matching observation. Releases
	// are only published after the journal confirms each exact-cause closure.
	// Mirror the conservative pre-persist set first: a slow or failed journal
	// write must never leave the execution gateway open after the mismatch is
	// already known.
	t.syncGate(t.snapshotBlocks())
	toPersist := out
	added := make(map[string]bool, len(out.Added))
	for _, block := range out.Added {
		added[block.Key()] = true
	}
	for _, block := range t.blocks {
		if block.pending && !added[block.Key()] {
			toPersist.Added = append(toPersist.Added, block)
		}
	}
	persisted, persistErr := t.persist(ctx, toPersist)
	for _, block := range persisted.durable {
		block.pending = false
		t.blocks[block.Key()] = block
	}
	for _, block := range persisted.authoritative {
		t.blocks[block.Key()] = block
	}
	for _, block := range persisted.released {
		delete(t.blocks, block.Key())
	}
	out.Cleared = persisted.released
	t.permanent = hasPermanentQuantityAccountBlock(t.blocks)

	// Credit accounting. Three outcomes, and which one a credit gets is decided
	// per symbol rather than per comparison (a083):
	//
	//	untouched  this observation is not later than the credit's comparison, so it
	//	           is not the re-read the rule means. Nothing is decided about it.
	//	refuted    a later comparison still disagrees about that symbol. The
	//	           adjustment did not settle it; releasing now would need a new one.
	//	answered   a later comparison agrees about that symbol. The credit is spent
	//	           by the release when one commits, and kept when it does not — a
	//	           storage failure must not lose the evidence that earned the
	//	           release, and neither must a comparison that is clean for this
	//	           symbol while another symbol keeps the whole diff blocking.
	disputed := symbolsInDispute(diff)
	committed := make(map[string]bool, len(persisted.released))
	for _, block := range persisted.released {
		committed[strings.ToUpper(strings.TrimSpace(block.Symbol))] = true
	}
	for symbol, comparison := range t.adjusted {
		if !creditUsableBy(comparison, diff.AsOf) {
			continue
		}
		if disputed[symbol] || committed[symbol] {
			delete(t.adjusted, symbol)
		}
	}

	out.Failures = t.failures
	out.Permanent = t.permanent
	out.Blocked = len(t.blocks) > 0
	out.NextDueAt = now.Add(interval)
	active := t.snapshotBlocks()
	t.syncGate(active)
	t.mu.Unlock()

	sortBlocks(out.Cleared)
	return out, persistErr
}

// persist writes this observation's promotions and releases through to the
// journal, which is where the block actually lives.
//
// Entering is idempotent (the journal keeps the first observation for a scope
// that is already active) and releasing names ExpectCause, so a clean pass
// closes only the rows this tracker opened.
//
// The release cause is ADJUSTMENT_APPLIED rather than RECHECK_MATCHED because
// out.Cleared only ever carries blocks whose symbol an adjustment converged —
// which is exactly the distinction the delta asks the record to keep (신규
// release cause와 원인 기록을 남긴다 SHALL). It is also what the states this
// tracker did not enter need: the frozen projections task 6.1 records
// (position_projection.go) are QUANTITY_MISMATCH rows, and the adjustment that
// unfreezes one is what earns their release too.
type persistResult struct {
	released      []Block
	authoritative []Block
	durable       []Block
}

func (t *Tracker) persist(ctx context.Context, out Outcome) (persistResult, error) {
	if t.Journal == nil {
		return persistResult{
			released: append([]Block(nil), out.Cleared...),
			durable:  append([]Block(nil), out.Added...),
		}, nil
	}
	result := persistResult{
		durable:  make([]Block, 0, len(out.Added)),
		released: make([]Block, 0, len(out.Cleared)),
	}
	for _, block := range out.Added {
		expectedCause := firstNonEmpty(block.Cause, journal.ReconcileCauseQuantityMismatch)
		state, entered, err := t.Journal.EnterReconcile(ctx, journal.EnterReconcileRequest{
			AccountRef: firstNonEmpty(block.Account, t.AccountRef),
			Symbol:     block.Symbol,
			Cause:      expectedCause,
			Evidence:   block.Detail,
		})
		if err != nil {
			return result, fmt.Errorf("reconcile: persisting the %s block on %s: %w",
				block.Reason, scopeLabel(block), err)
		}
		if !entered && state.Cause != expectedCause {
			authoritative := blockFromReconcileState(state)
			result.authoritative = append(result.authoritative, authoritative)
			return result, fmt.Errorf(
				"reconcile: persisting the %s block on %s: durable scope is already owned by cause %s",
				block.Reason, scopeLabel(block), firstNonEmpty(state.Cause, "UNKNOWN"))
		}
		result.durable = append(result.durable, block)
	}
	for _, block := range out.Cleared {
		_, ok, err := t.Journal.ReleaseReconcile(ctx, journal.ReleaseReconcileRequest{
			AccountRef:  firstNonEmpty(block.Account, t.AccountRef),
			Symbol:      block.Symbol,
			Cause:       journal.ReconcileReleaseAdjustmentApplied,
			Evidence:    "an adjustment converged the projection and the reconciliation after it agreed",
			ExpectCause: firstNonEmpty(block.Cause, journal.ReconcileCauseQuantityMismatch),
		})
		if err != nil {
			return result, fmt.Errorf("reconcile: releasing the %s block on %s: %w",
				block.Reason, scopeLabel(block), err)
		}
		if !ok {
			return result, fmt.Errorf(
				"reconcile: releasing the %s block on %s: the journal did not confirm the exact-cause release",
				block.Reason, scopeLabel(block))
		}
		result.released = append(result.released, block)
	}
	return result, nil
}

// Restore rebuilds the tracker and the entry gate from the journal.
//
// This is the restart path. The tracker's own rows (cause QUANTITY_MISMATCH)
// become its block set again — an account-wide one is the permanent promotion,
// which is why the counter is restored to the threshold rather than to zero:
// coming back with failures=0 would let the next clean pass "release" a
// permanent block a human never cleared.
//
// The gate is then rebuilt from *every* active state, not only this tracker's,
// because the gate is the projection of the whole table.
func (t *Tracker) Restore(ctx context.Context) error {
	if t.Journal == nil {
		return nil
	}
	states, err := t.Journal.ActiveReconcileStates(ctx)
	if err != nil {
		return fmt.Errorf("reconcile: reading the active RECONCILE states: %w", err)
	}

	blocks := map[string]Block{}
	accountStates := make([]journal.ReconcileState, 0, len(states))
	permanent := false
	for _, state := range states {
		if t.AccountRef != "" && state.AccountRef != t.AccountRef {
			continue
		}
		accountStates = append(accountStates, state)
		block := blockFromReconcileState(state)
		permanent = permanent || (block.Scope == ScopeAccount &&
			block.Cause == journal.ReconcileCauseQuantityMismatch && block.Permanent)
		blocks[block.Key()] = block
	}

	t.mu.Lock()
	t.blocks = blocks
	t.permanent = permanent
	// A restart comes back with the blocks and without the credits: whatever
	// applied an adjustment before the crash is gone, and the first comparison
	// after a restart must not release on its word.
	t.adjusted = nil
	if permanent && t.failures < t.maxFailures() {
		t.failures = t.maxFailures()
	}
	if t.Gate != nil {
		t.Gate.RebuildReconcileProjection(accountStates)
	}
	t.mu.Unlock()
	return nil
}

// Refresh replaces the durable part of the runtime projection without losing
// pending fail-closed blocks, failure counters, or adjustment credits. Runtime
// writers outside this tracker (notably the fill projection hook) call it after
// committing a journal change, and the reconciliation driver calls it before
// adoption judgement.
func (t *Tracker) Refresh(ctx context.Context) error {
	if t.Journal == nil {
		return nil
	}
	// Serialize the authority read with Observe's journal write-through. Taking
	// this lock after the read would let an older empty snapshot overwrite a
	// block Observe had just committed and reopen the gate.
	t.mu.Lock()
	states, err := t.Journal.ActiveReconcileStates(ctx)
	if err != nil {
		t.mu.Unlock()
		return fmt.Errorf("reconcile: refreshing the active RECONCILE states: %w", err)
	}

	blocks := map[string]Block{}
	for key, block := range t.blocks {
		if block.pending {
			blocks[key] = block
		}
	}
	for _, state := range states {
		if t.AccountRef != "" && state.AccountRef != t.AccountRef {
			continue
		}
		block := blockFromReconcileState(state)
		blocks[block.Key()] = block
	}
	t.blocks = blocks
	t.permanent = hasPermanentQuantityAccountBlock(blocks)
	if t.permanent && t.failures < t.maxFailures() {
		t.failures = t.maxFailures()
	}
	active := t.snapshotBlocks()
	t.syncGate(active)
	t.mu.Unlock()
	return nil
}

func scopeLabel(b Block) string {
	if b.Symbol == "" {
		return "(account-wide)"
	}
	return b.Symbol
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Due reports whether enough time has passed to reconcile again.
//
// The interval is not throttling for its own sake: reconciling in a tight loop
// against a disagreement that is really a settlement in flight would burn the
// query budget (§0.4) and make the fill-detection reads stale, which blocks
// entries for a second reason and hides the first.
func (t *Tracker) Due(now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.observed {
		return true
	}
	return !now.Before(t.lastAt.Add(t.interval()))
}

// NextDueAt reports when the next reconciliation may run.
func (t *Tracker) NextDueAt() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.observed {
		return time.Time{}
	}
	return t.lastAt.Add(t.interval())
}

// Blocks lists the active blocks, in a stable order.
func (t *Tracker) Blocks() []Block {
	t.mu.Lock()
	blocks := t.snapshotBlocks()
	t.mu.Unlock()
	sortBlocks(blocks)
	return blocks
}

// Failures reports the consecutive-failure count.
func (t *Tracker) Failures() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.failures
}

// Permanent reports whether the mismatch has been marked permanent.
func (t *Tracker) Permanent() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.permanent
}

// EntryAllowed reports why a new entry in this market and symbol is refused, or
// nil when it is allowed.
//
// This is the symbol-aware check the state table describes. Until task 4.2 wires
// it into the gateway, the account-wide latch on the entry gate is what actually
// stops trading; this method is the narrower answer that wiring will use.
func (t *Tracker) EntryAllowed(market, symbol string) *execgw.RejectedError {
	for _, block := range t.Blocks() {
		if block.Covers(market, symbol) {
			return &execgw.RejectedError{Reason: block.Reason, Detail: block.Detail}
		}
	}
	return nil
}

// ExitAllowed always reports true.
//
// It is a method rather than an absence so that §0.3 is something a test can
// assert and a reader can find. Reducing exposure is never gated: a cancel or a
// liquidation makes the disagreement smaller, and refusing it would make the
// engine's safety mechanism the thing that traps it.
func (t *Tracker) ExitAllowed(string, string) bool { return true }

// Resolve is the operator's exit from a permanent mismatch.
//
// It requires an identity and a note for the same reason the journal's operator
// resolution does: this is a human overriding the machine's judgement about a
// live account, and the audit trail is the point.
//
// The journal rows are closed first. Clearing memory and leaving the rows open
// would resurrect every block on the next Restore, which looks to an operator
// like their resolution was ignored.
func (t *Tracker) Resolve(ctx context.Context, operator, note string) error {
	if strings.TrimSpace(operator) == "" {
		return fmt.Errorf("reconcile: clearing a mismatch requires the operator's identity")
	}
	if strings.TrimSpace(note) == "" {
		return fmt.Errorf("reconcile: clearing a mismatch requires a note explaining what was verified")
	}

	if t.Journal != nil {
		blocks := t.Blocks()
		requests := make([]journal.ReleaseReconcileRequest, 0, len(blocks))
		for _, block := range blocks {
			requests = append(requests, journal.ReleaseReconcileRequest{
				AccountRef:  firstNonEmpty(block.Account, t.AccountRef),
				Symbol:      block.Symbol,
				Cause:       journal.ReconcileReleaseOperator,
				Evidence:    "operator " + strings.TrimSpace(operator) + ": " + strings.TrimSpace(note),
				ExpectCause: firstNonEmpty(block.Cause, journal.ReconcileCauseQuantityMismatch),
			})
		}
		if _, err := t.Journal.ReleaseReconciles(ctx, requests); err != nil {
			return fmt.Errorf("reconcile: recording the atomic operator release: %w", err)
		}
	}

	t.mu.Lock()
	t.permanent = false
	t.failures = 0
	t.blocks = map[string]Block{}
	t.adjusted = nil
	t.syncGate(nil)
	t.mu.Unlock()
	return nil
}

func blockFromReconcileState(state journal.ReconcileState) Block {
	reason := execgw.ReconcileReasonFor(state.Cause)
	block := Block{
		Scope:     ScopeSymbol,
		Account:   state.AccountRef,
		Symbol:    state.Symbol,
		Cause:     state.Cause,
		Reason:    reason,
		Detail:    state.Evidence,
		Since:     state.EnteredAt,
		Release:   ReleaseOperatorOnly,
		Permanent: reason == execgw.ReasonReconcilePermanent,
	}
	if state.Cause == journal.ReconcileCauseQuantityMismatch {
		block.Release = ReleaseOnAdjustedReconcile
	}
	if state.AccountWide() {
		block.Scope = ScopeAccount
		block.Symbol = ""
		if state.Cause == journal.ReconcileCauseQuantityMismatch {
			block.Reason = execgw.ReasonReconcilePermanent
			block.Release = ReleaseOperatorOnly
			block.Permanent = true
		}
	}
	return block
}

func hasPermanentQuantityAccountBlock(blocks map[string]Block) bool {
	for _, block := range blocks {
		if block.Scope == ScopeAccount && block.Cause == journal.ReconcileCauseQuantityMismatch &&
			block.Permanent {
			return true
		}
	}
	return false
}

// --- internals --------------------------------------------------------------

// blocksFor turns a diff into the blocks the state table says it raises.
func blocksFor(diff Diff, accountRef string, now time.Time) []Block {
	var out []Block
	for _, mismatch := range diff.Quantities {
		out = append(out, Block{
			Scope:   ScopeSymbol,
			Account: accountRef,
			Symbol:  mismatch.Symbol,
			Cause:   journal.ReconcileCauseQuantityMismatch,
			Reason:  execgw.ReasonReconcileMismatch,
			Since:   now,
			Release: ReleaseOnAdjustedReconcile,
			Detail: fmt.Sprintf(
				"the engine believes %s of %s, the account says %s; the account wins",
				mismatch.Local, mismatch.Symbol, mismatch.Broker),
		})
	}
	for _, missing := range diff.MissingOrders {
		out = append(out, Block{
			Scope:   ScopeSymbol,
			Account: accountRef,
			Market:  missing.Market,
			Symbol:  missing.Symbol,
			Cause:   journal.ReconcileCauseQuantityMismatch,
			Reason:  execgw.ReasonReconcileMismatch,
			Since:   now,
			Release: ReleaseOnAdjustedReconcile,
			Detail: fmt.Sprintf("order %s on %s is not in the account's open list",
				missing.OrderID, missing.Symbol),
		})
	}
	return out
}

// syncGate mirrors the block set onto the entry gate, each block at the scope the
// state table gives it (task 4.2).
//
// Until the gate had a symbol dimension, every block here was mirrored
// account-wide — safe, but it meant one symbol's disagreement stopped the engine
// trading anything at all (issues.md, 2026-07-26). Now an account-scoped row
// latches the account and a symbol-scoped row latches that symbol, which is what
// BlockRules() has said all along.
//
// Precedence still matters for the account latches: a permanent mismatch and an
// ordinary one are cleared by different things, so latching the ordinary reason
// over a permanent one would make a clean reconciliation appear to release a
// block that an operator has to clear.
func (t *Tracker) syncGate(active []Block) {
	if t.Gate == nil {
		return
	}
	accountWide := map[execgw.ReasonCode]Block{}
	for _, block := range active {
		if block.Scope == ScopeAccount {
			accountWide[block.Reason] = block
		}
	}

	// Symbol-scoped rows: block what is still disagreeing, release what is not.
	// Clearing first and re-blocking is wrong — it would drop the "since" of a
	// block that never went away — so the surviving set is computed and the
	// difference removed.
	surviving := make(map[string]bool, len(active))
	for _, block := range active {
		if block.Scope != ScopeSymbol {
			continue
		}
		surviving[strings.ToUpper(block.Symbol)+"|"+string(block.Reason)] = true
		t.Gate.BlockSymbol(block.Market, block.Symbol, block.Reason, block.Detail)
	}
	for _, existing := range t.Gate.SymbolBlocks() {
		if !isReconcileReason(existing.Reason) {
			// Somebody else's block (fill detection, the flatten saga). Releasing
			// it because *this* reconciliation is happy would be answering a
			// question nobody asked us.
			continue
		}
		if !surviving[strings.ToUpper(existing.Symbol)+"|"+string(existing.Reason)] {
			t.Gate.ClearSymbol(existing.Market, existing.Symbol, existing.Reason)
		}
	}

	for _, reason := range []execgw.ReasonCode{
		execgw.ReasonReconcilePermanent,
		execgw.ReasonReconcileMismatch,
	} {
		if block, ok := accountWide[reason]; ok {
			t.Gate.Block(reason, block.Detail)
		} else {
			t.Gate.Clear(reason)
		}
	}
}

// isReconcileReason reports whether a symbol block is one syncGate may release.
//
// Both reconcile reasons are owned by this full active-state projection. Other
// gate families use different reason codes and are left untouched.
func isReconcileReason(reason execgw.ReasonCode) bool {
	return reason == execgw.ReasonReconcileMismatch || reason == execgw.ReasonReconcilePermanent
}

// snapshotBlocks copies the block set. Callers must hold t.mu.
func (t *Tracker) snapshotBlocks() []Block {
	out := make([]Block, 0, len(t.blocks))
	for _, block := range t.blocks {
		out = append(out, block)
	}
	return out
}

func sortBlocks(blocks []Block) {
	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i].Permanent != blocks[j].Permanent {
			return blocks[i].Permanent
		}
		return blocks[i].Key() < blocks[j].Key()
	})
}

func (t *Tracker) clock() clock.Clock {
	if t.Clock == nil {
		return clock.System()
	}
	return t.Clock
}

func (t *Tracker) interval() time.Duration {
	if t.MinInterval <= 0 {
		return DefaultReconcileInterval
	}
	return t.MinInterval
}

func (t *Tracker) maxFailures() int {
	if t.MaxFailures <= 0 {
		return DefaultMaxFailures
	}
	return t.MaxFailures
}
