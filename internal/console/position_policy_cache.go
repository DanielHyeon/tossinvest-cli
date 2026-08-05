package console

// position_policy_cache.go is change a081: the two protection-line screens stop
// reaching the engine process once per render.
//
// # What a render used to cost the engine
//
// decoratePositionRows asked the engine twice, every time it ran, with nothing in
// between:
//
//	runtime, _ = c.opts.PositionPolicies.Runtime(ctx)
//	if states, err := c.opts.PositionPolicies.List(ctx); err == nil {
//
// Neither is a local call. Runtime re-reads the descriptor and dials the engine's
// Unix endpoint afresh on every invocation (cmd/tossctl/position_policy_commander.go),
// and List is served inside the engine by a SELECT on the journal handle the
// engine owns — the one whose pool is a single connection, because that is what
// makes "BEGIN IMMEDIATE then commit then submit" hold without SQLITE_BUSY races
// (internal/journal/journal.go). The exit loop's judgement transaction runs on
// that same connection, and the loop sleeps *after* its work rather than on a
// ticker, so time taken from the connection is added to the interval between
// stop-loss judgements.
//
// That made the number of open console screens, and how often each one redrew, an
// input to how quickly a stop is evaluated. It should not be one at any reload
// period. This was found by the independent review of change a080 (review.md F1)
// and is fixed here, ahead of it, because it is a defect a080 revealed rather than
// one a080 introduced: two open tabs already pay it every thirty seconds today.
//
// # The two reads have different costs, so they get different intervals
//
// The first draft of this file cached both on one thirty-second interval and
// argued that "nothing behind these two reads moves at render granularity". Three
// independent reviews showed that was false, and that the two reads are not
// alike:
//
//	List     is the one that contends. It is a SELECT on the engine's single
//	         write connection, which is the whole safety argument above. What it
//	         carries — a position's lifecycle status and adoption generation —
//	         moves when an operator applies something (which drops this cache
//	         outright, see invalidate) or when the engine adopts.
//
//	Runtime  touches no database at all. It reads the engine instance's startup
//	         adoption settings, fixed in its constructor, plus the reconcile
//	         tracker's current blocks (internal/app/engine/position_policy_command.go).
//	         It cannot delay a stop judgement, so it needs no safety bound — and
//	         those blocks are live tracker state on the reconcile loop's own
//	         cadence, so holding them for half a minute makes the screen say
//	         "편입 예약됨" about a holding the engine has in fact stopped adopting.
//	         That is optimistic-wrong, which is the direction that matters.
//
// So each interval is set by its own argument rather than by a shared number.
// engineLifecycleInterval bounds contention; engineRuntimeInterval keeps the
// fast-moving half honest and costs only a socket dial.
//
// # Why the two are no longer expired as a pair
//
// The draft kept them together so a render would see "one moment". That argument
// does not survive contact with the third reader: c.positions() reads the
// console's own journal handle on every render, so the row a reading is joined
// against is always fresher than the reading. There is no one moment to preserve
// — there are three, and there always were.
//
// What is preserved instead is the direction of the disagreement. Where the
// lifecycle list is older than the journal, the row it fails to match renders
// 관리 여부 불명 with no protection line rather than a protected one: a stale list
// can only withhold a verdict, never invent one. attachPositionExitLines makes the
// same choice about a generation mismatch. Both are fail-closed, and the tests
// below pin that rather than an unattainable simultaneity.
//
// # Why this is not holdingsCache
//
// holdingsCache keeps its last good reading when a refresh fails and prints its
// age beside it, because a position size an hour old is still information about
// somebody's money. An adoption snapshot is not that kind of fact. It is a claim
// about what is protecting the account *now*, and serving a dead engine's last
// success would have the screen assert an effective policy nobody is maintaining
// — which the operator-console spec forbids in as many words ("runtime unavailable인
// non-managed 행은 desired를 effective로 위장하지 않고 UNKNOWN").
//
// So the unit stored here is one attempt's outcome, errors included. A failure is
// served for the interval exactly as a success is, and the screen goes back to
// saying "알 수 없음" within one interval of the engine going away. Failures count
// against the interval too: the moment the engine cannot answer is the worst
// moment to open a fresh socket to it on every render, which is the same call
// holdings.go made about a broker returning 429.
//
// # Why the refresh does not run on the request's context
//
// A reading is shared, and the request that happens to take it is not. If the
// browser abandons a render — a reload during load, a closed tab, a speculative
// navigation — its context dies mid-read, and recording that failure would blank
// the protection line on both screens, on a healthy engine, for a whole interval,
// with no way for the operator to force a retry. Two of the three reviews
// reproduced exactly that.
//
// "The engine could not answer" and "this HTTP request went away" are different
// facts and only the first belongs in a shared cache. The refresh therefore runs
// on a detached context with a bound of its own, so what is stored is always an
// answer about the engine.

import (
	"context"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// engineLifecycleInterval is how long one lifecycle listing is shared.
//
// This is the read that competes with the exit loop for the engine's write
// connection, so this is the number the safety argument is about. Thirty seconds
// is what the two screens' shipped reload period already makes the effective rate,
// so nothing gets slower than it is today; what changes is that the bound now
// holds however often they redraw and however many are open. With two tabs open
// the console reaches the engine less often than it does today, not more.
//
// It is deliberately not derived from holdingsTTL. They are the same number today
// and they answer different questions about different resources — how often this
// console may spend the account's broker budget, and how often it may take the
// engine's write connection. a080's issues.md I1 records what happened the last
// time one constant stood for two decisions.
const engineLifecycleInterval = 30 * time.Second

// engineRuntimeInterval is how long one runtime reading is shared.
//
// Runtime touches no database, so no safety bound applies to it and the only cost
// it has is a descriptor read and a Unix dial. What decides this number is the
// other direction: the reconcile blocks it carries are live tracker state, and a
// screen that reports a stale block tells the operator a holding is queued for
// adoption when the engine has stopped adopting it.
//
// The engine's own exit observation period is the floor on how fresh any of this
// can usefully be, so that is what this matches (engine.DefaultExitObservationInterval,
// duplicated rather than imported — the console runs as its own process and must
// not compile against the engine runtime).
const engineRuntimeInterval = 5 * time.Second

// enginePolicyTimeout bounds one refresh.
//
// The shipped clients already carry five-second timeouts of their own, but
// PositionPolicyCommander is an exported seam and a substituted implementation
// need not. holdings.go makes the same argument about a broker that has stopped
// answering: it must not hold an HTTP handler open.
const enginePolicyTimeout = 10 * time.Second

// positionPolicyReader is the two reads this cache is allowed to make.
//
// The cache takes this rather than the whole PositionPolicyCommander so that it
// cannot name Preview or Apply — a caching wrapper around the commander would put
// every caller behind the cache silently, including the command screen that must
// decide an operator's action on a reading taken for that action. Who reads
// through the cache stays visible at the call site.
type positionPolicyReader interface {
	Runtime(context.Context) (positionpolicy.ManagementRuntime, error)
	List(context.Context) ([]positionpolicy.State, error)
}

// enginePolicyReading is what one render is given.
//
// The two halves are taken on their own intervals and are therefore not
// necessarily of the same instant — see the file comment on why that simultaneity
// was never available anyway.
//
// The errors are carried rather than resolved because their handling is the
// caller's: Runtime's error leaves a zero-value runtime, which is what renders
// "unknown", and List's error is what leaves the lifecycle map nil.
type enginePolicyReading struct {
	Runtime    positionpolicy.ManagementRuntime
	RuntimeErr error
	States     []positionpolicy.State
	StatesErr  error
}

// positionPolicyCache holds the most recent reading of each half.
//
// The mutex is held across the reads, like holdingsCache's. Two screens rendering
// at the same moment then cost one reading rather than two: the second waits,
// finds the reading fresh, and returns it. A slow engine delaying both pages is
// the right way round — the console has one operator and no concurrency worth
// protecting, and the engine's write connection has a stop-loss on it.
type positionPolicyCache struct {
	reader            positionPolicyReader
	lifecycleInterval time.Duration
	runtimeInterval   time.Duration

	mu sync.Mutex

	runtime          positionpolicy.ManagementRuntime
	runtimeErr       error
	runtimeTaken     time.Time
	runtimeAttempted bool

	states          []positionpolicy.State
	statesErr       error
	statesTaken     time.Time
	statesAttempted bool
}

func newPositionPolicyCache(reader positionPolicyReader, lifecycle, runtime time.Duration) *positionPolicyCache {
	if reader == nil {
		return nil
	}
	if lifecycle <= 0 {
		lifecycle = engineLifecycleInterval
	}
	if runtime <= 0 {
		runtime = engineRuntimeInterval
	}
	return &positionPolicyCache{reader: reader, lifecycleInterval: lifecycle, runtimeInterval: runtime}
}

// read returns the current reading, refreshing either half whose last attempt is
// its own interval old.
//
// A nil cache is the unwired build and answers with an empty reading without
// reaching anything. "Is the engine wired" stays the question it was before this
// cache existed — decoratePositionRows still asks it of Options.PositionPolicies,
// so a caller that unwires the seam after construction gets the unwired rendering
// and not a reading held over from when it was wired.
func (c *positionPolicyCache) read(ctx context.Context, now time.Time) enginePolicyReading {
	if c == nil || c.reader == nil {
		return enginePolicyReading{}
	}

	// Detached from the caller: what is stored here is shared by every render in
	// the interval, so it must be an answer about the engine and never a record
	// of one browser walking away mid-request.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), enginePolicyTimeout)
	defer cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.runtimeAttempted || now.Sub(c.runtimeTaken) >= c.runtimeInterval {
		c.refreshRuntimeLocked(ctx, now)
	}
	if !c.statesAttempted || now.Sub(c.statesTaken) >= c.lifecycleInterval {
		c.refreshLifecycleLocked(ctx, now)
	}
	return c.snapshotLocked()
}

// invalidate drops both readings so the next render takes fresh ones.
//
// The console calls this after a policy mutation of its own succeeds. An operator
// who has just released a position and cannot see it released is looking at a bug,
// not at a cache. Only successes invalidate: a refused apply changed nothing, and
// dropping the reading on refusal would let a repeated bad request defeat the
// bound this cache exists to hold.
func (c *positionPolicyCache) invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runtime, c.runtimeErr, c.runtimeTaken, c.runtimeAttempted = positionpolicy.ManagementRuntime{}, nil, time.Time{}, false
	c.states, c.statesErr, c.statesTaken, c.statesAttempted = nil, nil, time.Time{}, false
}

// refreshRuntimeLocked takes the one runtime reading an interval is allowed.
//
// taken is stamped before the read rather than after, so a slow or failing engine
// is asked once per interval and not once per render that waited on it.
func (c *positionPolicyCache) refreshRuntimeLocked(ctx context.Context, now time.Time) {
	c.runtimeTaken, c.runtimeAttempted = now, true
	c.runtime, c.runtimeErr = c.reader.Runtime(ctx)
}

// refreshLifecycleLocked takes the one listing an interval is allowed.
func (c *positionPolicyCache) refreshLifecycleLocked(ctx context.Context, now time.Time) {
	c.statesTaken, c.statesAttempted = now, true
	c.states, c.statesErr = c.reader.List(ctx)
}

// snapshotLocked hands out a copy of everything a render could hold onto.
//
// The slices are shared by every render inside the interval and the rows built
// from them outlive this lock. The elements are display facts nothing in this
// package writes to, but the headers are copied so one render's slicing cannot
// reach the next one's.
func (c *positionPolicyCache) snapshotLocked() enginePolicyReading {
	reading := enginePolicyReading{
		Runtime:    c.runtime,
		RuntimeErr: c.runtimeErr,
		States:     append([]positionpolicy.State(nil), c.states...),
		StatesErr:  c.statesErr,
	}
	reading.Runtime.Blocks = append([]positionpolicy.ReconcileBlock(nil), c.runtime.Blocks...)
	return reading
}
