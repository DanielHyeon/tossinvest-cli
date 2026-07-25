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
//	condition                          reason code                        scope    auto release        manual release
//	---------------------------------  ---------------------------------  -------  ------------------  ---------------
//	restart recovery unfinished        recovery_incomplete                account  recovery completes  —
//	quantity disagreement              reconciliation_mismatch            symbol   clean reconcile     operator
//	order the account does not show    reconciliation_mismatch            symbol   clean reconcile     operator
//	3 consecutive failed reconciles    reconciliation_mismatch_permanent  account  never               operator
//	broker snapshot contradiction      unknown_broker_state               symbol   never               operator
//
// # The scope caveat
//
// The symbol-scoped rows are recorded at symbol scope here and *also* latch the
// account-wide entry gate, because execgw.EntryGate has no symbol dimension yet
// (issues.md, 2026-07-26). Blocking wider than the table says is the safe
// direction — it stops more trading, not less — and narrowing it is wiring work
// for task 4.2. EntryAllowed below is the symbol-aware check that wiring will
// call.

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
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
	// ReleaseOnCleanReconcile — the next agreeing reconciliation clears it.
	ReleaseOnCleanReconcile Release = "clean_reconcile"
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
			Auto:      ReleaseOnCleanReconcile,
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
	Reason  execgw.ReasonCode
	Detail  string
	Since   time.Time
	// Release is the state table's automatic release condition.
	Release Release
	// Permanent reports that no automatic path clears this block.
	Permanent bool
}

// Key identifies a block for deduplication.
func (b Block) Key() string {
	return string(b.Scope) + "|" + b.Market + "|" + b.Symbol + "|" + string(b.Reason)
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

// Tracker holds the mismatch state across reconciliations.
type Tracker struct {
	// Clock drives the interval. Defaults to clock.System().
	Clock clock.Clock
	// Gate is latched while any block is active. Optional but expected: without
	// it the tracker reports blocks nobody enforces.
	Gate *execgw.EntryGate
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
	// Cleared are blocks this observation released.
	Cleared []Block
	// NextDueAt is the earliest instant a re-reconciliation may run.
	NextDueAt time.Time
}

// Observe records the result of one reconciliation.
func (t *Tracker) Observe(diff Diff) Outcome {
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
		// failure, which is what separates a stuck account from a busy one.
		t.failures = 0
		for key, block := range t.blocks {
			if block.Permanent {
				continue
			}
			delete(t.blocks, key)
			out.Cleared = append(out.Cleared, block)
		}
	} else {
		t.failures++
		for _, block := range blocksFor(diff, t.AccountRef, now) {
			if _, exists := t.blocks[block.Key()]; !exists {
				t.blocks[block.Key()] = block
				out.Added = append(out.Added, block)
			}
		}
		if t.failures >= maxFailures && !t.permanent {
			t.permanent = true
			permanent := Block{
				Scope:     ScopeAccount,
				Account:   t.AccountRef,
				Reason:    execgw.ReasonReconcilePermanent,
				Since:     now,
				Release:   ReleaseOperatorOnly,
				Permanent: true,
				Detail: fmt.Sprintf(
					"reconciliation disagreed %d times in a row; looking again has been shown not to resolve it",
					t.failures),
			}
			t.blocks[permanent.Key()] = permanent
			out.Added = append(out.Added, permanent)
		}
	}

	out.Failures = t.failures
	out.Permanent = t.permanent
	out.Blocked = len(t.blocks) > 0
	out.NextDueAt = now.Add(interval)
	active := t.snapshotBlocks()
	t.mu.Unlock()

	sortBlocks(out.Added)
	sortBlocks(out.Cleared)
	t.syncGate(active)
	return out
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
func (t *Tracker) Resolve(operator, note string) error {
	if strings.TrimSpace(operator) == "" {
		return fmt.Errorf("reconcile: clearing a mismatch requires the operator's identity")
	}
	if strings.TrimSpace(note) == "" {
		return fmt.Errorf("reconcile: clearing a mismatch requires a note explaining what was verified")
	}

	t.mu.Lock()
	t.permanent = false
	t.failures = 0
	t.blocks = map[string]Block{}
	t.mu.Unlock()

	if t.Gate != nil {
		t.Gate.Clear(execgw.ReasonReconcilePermanent)
		t.Gate.Clear(execgw.ReasonReconcileMismatch)
	}
	return nil
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
			Reason:  execgw.ReasonReconcileMismatch,
			Since:   now,
			Release: ReleaseOnCleanReconcile,
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
			Reason:  execgw.ReasonReconcileMismatch,
			Since:   now,
			Release: ReleaseOnCleanReconcile,
			Detail: fmt.Sprintf("order %s on %s is not in the account's open list",
				missing.OrderID, missing.Symbol),
		})
	}
	return out
}

// syncGate mirrors the block set onto the account-wide entry gate.
//
// Precedence matters: a permanent mismatch and an ordinary one are cleared by
// different things, so latching the ordinary reason over a permanent one would
// make a clean reconciliation appear to release a block that an operator has to
// clear.
func (t *Tracker) syncGate(active []Block) {
	if t.Gate == nil {
		return
	}
	var permanent, ordinary *Block
	for i := range active {
		switch {
		case active[i].Permanent && permanent == nil:
			permanent = &active[i]
		case !active[i].Permanent && ordinary == nil:
			ordinary = &active[i]
		}
	}

	if permanent != nil {
		t.Gate.Block(execgw.ReasonReconcilePermanent, permanent.Detail)
	}
	if ordinary != nil {
		t.Gate.Block(execgw.ReasonReconcileMismatch, ordinary.Detail)
	} else {
		t.Gate.Clear(execgw.ReasonReconcileMismatch)
	}
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
