package candidate

// watch.go is the scan cycle: what one turn of discovery does besides reading the
// sources, and the loop that repeats it.
//
// # Everything here is a thing that fails quietly if it is left out
//
// The cycle is not just "call Collect on a timer". Four obligations attach to it,
// and each one is invisible when it is missing:
//
//	retention        PruneObservations and PruneStale existed and nobody called
//	                 them, so the default behaviour of this store was unbounded
//	                 growth (D16). Cleanup is part of the cycle rather than an
//	                 option, because the growth is on the ledger's filesystem.
//	the space floor  when that filesystem fills, the write that fails is whichever
//	                 one gets there first. D16 fixes the order: discovery stops,
//	                 and says so, while there is still room. The ledger is orders,
//	                 fills and reconciliation.
//	the engine yield spec Requirement 7 ranks engine > verification > discovery.
//	                 The middle one has the run lock; the top one had a sentence.
//	                 The engine's delay reaches stop-loss immediacy through fill
//	                 detection, so this yield is code rather than an operator's
//	                 judgement — and it is recorded, because a scan that silently
//	                 slowed down has lost the earliness it exists for.
//	the backoff      the official ranking has no WTS fallback, so a 429 removes the
//	                 source outright. The retreat is written down for the same
//	                 reason: a scan that quietly polls slower reports a calm market.
//
// # Nothing here reads a wall clock
//
// Every instant comes from the store's injected clock and every wait comes from the
// same one. That is not a style rule: metrics_test.go's guard lists eleven spellings
// of binding work to a clock a test cannot move, and a backoff ladder is naturally
// written with three of them (time.After, time.NewTicker, time.Sleep). A loop
// written that way makes its own retreats untestable, which is exactly the code
// whose behaviour has to be proven.
//
// # And it still cannot order anything
//
// A cycle reads Sources, writes observations and computes verdicts. The Source
// interface has one method and it takes a market and returns rows; there is no way
// to express anything else through it. The dependency closure of this package is
// still {internal/clock} — the run lock and the engine marker are file-based
// signals this package is handed as functions, not packages it imports.

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// --- the retreat (task 5.4) --------------------------------------------------------

// ErrHeldByBackoff means a source was not read because it is inside a retreat.
//
// It is a distinct sentinel from ErrRateLimited, and the distinction is a
// measurement rather than a nicety. The count of 429s is the only evidence anybody
// has of the undocumented RANKING limit (D13 decision 2), and counting our own
// retreat as another 429 would inflate that count by however many ticks the ladder
// happens to cover — turning one rate limit into four and the measurement into a
// function of our own backoff length.
var ErrHeldByBackoff = errors.New("candidate: source is inside its backoff window")

// BackoffLadder is the retreat the design guarantees: 30 → 60 → 120 → 300 seconds.
//
// It is not decoration and it is not local. DefaultStalenessTTL and MaxRankPriorAge
// are both twice its last rung, on the argument that a retreat that long is normal
// operation and must not expire the candidates it happens during. A ladder that
// climbed somewhere else would move two bounds that were set against it, silently.
var BackoffLadder = []time.Duration{
	30 * time.Second,
	60 * time.Second,
	120 * time.Second,
	300 * time.Second,
}

// Retreat is one source backing off, as a recorded fact.
//
// spec Requirement 7: 백오프 사실은 스캔 결과에 남아야 한다 — 조용히 느려진 스캔은
// 조기성을 잃고도 그것을 보고하지 않는다. Discovery's product is the resolution of
// first_seen_at, and a retreat spends it. A retreat nobody wrote down is
// indistinguishable from a market that went quiet.
type Retreat struct {
	// Source is who retreated.
	Source SourceID
	// Step is the rung of the ladder, 1-based. It is on the record because the
	// difference between one 429 and a source that has been failing for ten minutes
	// is the difference between a blip and an outage.
	Step int
	// For is how long this rung lasts, and Since/Until are the window it covers.
	For          time.Duration
	Since, Until time.Time
}

// Backoff is the per-source retreat state.
//
// It is safe for concurrent use for Schedule's reason: the loop advances it and a
// caller rendering a screen may read it, and Go's unrecoverable "concurrent map
// read and map write" in a process that also runs the engine would take fill
// detection down with it.
type Backoff struct {
	mu       sync.Mutex
	retreats map[SourceID]Retreat
}

// NewBackoff returns an empty retreat record.
func NewBackoff() *Backoff {
	return &Backoff{retreats: map[SourceID]Retreat{}}
}

// Note records a rate-limited read and returns the retreat now in force.
//
// The ladder climbs one rung per 429 and stops at the top rather than wrapping:
// five minutes is the longest retreat the design plans for, and the two staleness
// bounds are derived from that being true.
func (b *Backoff) Note(id SourceID, at time.Time) Retreat {
	b.mu.Lock()
	defer b.mu.Unlock()
	step := b.retreats[id].Step + 1
	if step > len(BackoffLadder) {
		step = len(BackoffLadder)
	}
	d := BackoffLadder[step-1]
	r := Retreat{Source: id, Step: step, For: d, Since: at.UTC(), Until: at.UTC().Add(d)}
	b.retreats[id] = r
	return r
}

// Recovered clears a source's ladder after a successful read.
//
// Without it one bad minute leaves a source at five-minute intervals for the rest of
// the session, which is the earliness this change buys being spent on a 429 that is
// long over.
func (b *Backoff) Recovered(id SourceID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.retreats, id)
}

// Held reports that a source is inside its retreat window at `at`, with the retreat
// that says so.
func (b *Backoff) Held(id SourceID, at time.Time) (Retreat, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	r, ok := b.retreats[id]
	if !ok || !at.Before(r.Until) {
		return Retreat{}, false
	}
	return r, true
}

// Active returns every retreat still in force at `at`, in source order.
func (b *Backoff) Active(at time.Time) []Retreat {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []Retreat
	for _, r := range b.retreats {
		if at.Before(r.Until) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Source < out[j].Source })
	return out
}

// heldSource stands in for a source the cycle did not call because it is retreating.
//
// It is in the panel rather than removed from it, and that is the load-bearing part.
// Cooling is a claim about evidence (task 2.8): a candidate may be cooled only when
// every source that raised it answered and none of them listed it. A source removed
// from the panel reads as "a source that is gone" and stops being required to
// answer, so its candidates would be cooled by a scan that never looked at them —
// and the cooling clock then expires them, taking every first_seen_at with it. That
// is precisely the destruction a 429 must not cause.
//
// Present-and-failing is the honest shape: we meant to read it, we did not, and
// nothing it used to raise may be cooled on this pass.
type heldSource struct {
	id      SourceID
	retreat Retreat
}

func (h heldSource) ID() SourceID { return h.id }

func (h heldSource) Read(context.Context, string) (Reading, error) {
	// The wording deliberately avoids the anchors isRateLimited matches. This is our
	// own retreat, not a new refusal from the server, and counting it as one would
	// corrupt the only measurement of the undocumented limit.
	return Reading{}, fmt.Errorf("%w: rung %d of %d, until %s",
		ErrHeldByBackoff, h.retreat.Step, len(BackoffLadder),
		h.retreat.Until.UTC().Format(time.RFC3339))
}

// --- the space floor (task 5.3c, D16) ----------------------------------------------

// DefaultFreeSpaceFloor is the contract's `free_space_floor_mb: 512`.
//
// Below it discovery stops writing. The number is not a capacity estimate — D11 puts
// the raw table at roughly a gigabyte a day, so half a gigabyte is well under one
// session's growth. It is a margin for the ledger: enough room left for orders,
// fills and reconciliation to keep being recorded while somebody deals with the
// disk, on the principle that the side which fails is always this one.
const DefaultFreeSpaceFloor int64 = 512 << 20

// SpaceReport is what the cycle learned about the room left on the store's
// filesystem.
type SpaceReport struct {
	// Path is the directory that was probed.
	Path string
	// Checked is the absent-versus-zero flag. False means the probe did not answer,
	// which is not the same fact as "no space" and not the same fact as "plenty" —
	// see Below.
	Checked bool
	// Free is the bytes available to this user, meaningful only when Checked.
	Free int64
	// Floor is what Free was compared against.
	Floor int64
	// Why names the reason there is no measurement. Empty exactly when Checked.
	Why string
}

// Below reports that the measured free space is under the floor.
//
// It requires the measurement, so it is never true of a probe that failed. That is
// why Cycle halts on `!Checked || Below()` rather than on Below() alone: an
// unmeasured filesystem is not a filesystem with room, and this is the one resource
// D16 says discovery must run out of first.
func (s SpaceReport) Below() bool { return s.Checked && s.Free < s.Floor }

// --- the cleanup (task 5.2, D16) -----------------------------------------------------

// CleanupReport is what the cycle's retention sweep did.
type CleanupReport struct {
	// Ran is false only when the store refused the sweep.
	Ran bool
	// Retention is the window that was applied.
	Retention time.Duration
	// Pruned counts the raw observations deleted. The candidate summaries are
	// outside the statement's reach by construction (D11), which is what keeps a
	// tidy-up from deleting this package's entire claim.
	Pruned int64
	// Summaries counts the candidate summaries deleted: the other half of D11's
	// split, which had no enforcer until task 5.8. They go one retention window
	// after the life ended rather than on expiry, because that is when the raw
	// readings they describe are gone and the summary joins to nothing.
	//
	// It is a separate count from Pruned because the two answer different
	// questions. A large Pruned is a busy market; a large Summaries is a lot of
	// lives that ended, which is D3's "늦게 본 비율" accounting losing its rows —
	// and the count is the last place that fact is visible.
	Summaries int64
	// Checkpointed reports that the write-ahead log was reclaimed, and Busy that it
	// could not be because a reader was attached. The console's candidate screen is
	// exactly such a reader, so busy is an ordinary outcome rather than a failure.
	Checkpointed bool
	Busy         bool
	// Pages is how many WAL pages went back into the table.
	Pages int64
	// Why names a sweep that could not run. Empty when Ran.
	Why string
}

// --- the assessment ------------------------------------------------------------------

// DefaultAssessHistory is how far back a cycle loads readings to assess from.
//
// Ten minutes, and it is the same derivation as DefaultStalenessTTL and
// MaxRankPriorAge rather than a coincidence: past twice the longest planned backoff,
// a gap is not the backoff ladder and every metric here already refuses the reading
// on its own bound. The acceleration needs two windows (sixty seconds), near_high
// refuses a candle older than MaxVetoInputAge (two minutes), and the rank move
// refuses a prior reading older than MaxRankPriorAge (ten). Loading more would cost
// the query and change no answer.
const DefaultAssessHistory = 10 * time.Minute

// Verdict is one candidate's complete read-only assessment as of an instant.
//
// The three measurements travel beside the verdict rather than being folded into it,
// because a screen showing "unmeasured (NO_DAY_HIGH)" has to be able to show what
// *was* read — D6's rule that the reading and anything derived from it are different
// columns, applied to the whole assembly.
type Verdict struct {
	// Summary is the candidate with its two stored firsts.
	Summary Summary
	// Chase is the three-state verdict. Under D18 two of its three codes are
	// THRESHOLD_ABSENT, so Passed() is unreachable — an honest result rather than an
	// outage, and the one wrong repair is to count an absent threshold as a pass.
	Chase Chase
	// Sighting, Expansion and Range are the inputs the verdict was computed from.
	Sighting  Sighting
	Expansion Expansion
	Range     RangePosition
	// SeenLateBand and ExtendedBand are the shadow record for the two codes that
	// have no threshold (task 4.10). They decide nothing — see band.go.
	SeenLateBand, ExtendedBand ShadowBand
	// Accelerations is one entry per source series this candidate has, because D9
	// makes the series (market, symbol, source): two sources' cumulative figures
	// differenced against each other measure the sources rather than the market.
	Accelerations []Acceleration
}

// AssessOptions configures one market's assessment.
type AssessOptions struct {
	// Market is what to assess.
	Market string
	// At is the instant to assess as of. Input ages are measured against it.
	At time.Time
	// Thresholds are the veto's numbers. Two of the three have none (D18) and
	// their vetoes are unmeasured rather than clear.
	Thresholds VetoThresholds
	// History is how far back to load readings. Zero uses DefaultAssessHistory.
	History time.Duration
	// AccelerationWindow is D9's window. Zero uses DefaultAccelerationWindow.
	AccelerationWindow time.Duration
}

// Assess computes every live candidate's verdict for one market.
//
// # Live, not every row
//
// Active and cooled candidates are assessed; expired ones are not. An expired
// candidate is a life that ended (D1) and the next crossing starts a new one, so its
// stored firsts belong to a record that is closed — counting it would put a symbol
// nobody is watching into the tally an operator reads as "what is in front of me
// now". The rows themselves are kept: D3 requires the vetoed and the missed to stay
// countable, and PruneObservations is the only thing that ever removes them.
func Assess(ctx context.Context, store *Store, opts AssessOptions) ([]Verdict, error) {
	market := opts.Market
	history := opts.History
	if history <= 0 {
		history = DefaultAssessHistory
	}
	window := opts.AccelerationWindow
	if window <= 0 {
		window = DefaultAccelerationWindow
	}

	summaries, err := store.Summaries(ctx, opts.At)
	if err != nil {
		return nil, err
	}
	rows, err := store.ObservationsSince(ctx, market, opts.At.Add(-history))
	if err != nil {
		return nil, err
	}
	bySymbol := map[string][]Observation{}
	for _, o := range rows {
		bySymbol[o.Symbol] = append(bySymbol[o.Symbol], o)
	}

	var out []Verdict
	for _, summary := range summaries {
		if summary.Market != market || summary.State == StateExpired {
			continue
		}
		out = append(out, assessOne(summary, bySymbol[summary.Symbol], opts.At,
			opts.Thresholds, window))
	}
	return out, nil
}

// assessOne is one candidate's verdict from its summary and its retained readings.
func assessOne(summary Summary, rows []Observation, at time.Time,
	th VetoThresholds, window time.Duration) Verdict {

	v := Verdict{Summary: summary}
	v.Sighting = MeasureFirstSighting(summary.Candidate, summary.FirstRank, rows)
	v.Expansion = MeasureExpansion(summary.Baseline, rows)
	v.Range = MeasureRangePosition(rows)
	v.Chase = AssessChase(VetoInputs{
		Candidate:  summary.Candidate,
		Sighting:   v.Sighting,
		Expansion:  v.Expansion,
		Range:      v.Range,
		At:         at,
		Thresholds: th,
	})
	v.SeenLateBand = MeasureSeenLateBand(v.Sighting)
	v.ExtendedBand = MeasureExtendedBand(v.Expansion, summary.Candidate, at, th)

	// One series per source, never one series over all of them. D9 warns that the
	// mixed version compiles and produces a plausible number, and that the number is
	// the gap between two sources' conventions rather than the market's acceleration.
	bySource := map[SourceID][]Observation{}
	var order []SourceID
	for _, o := range rows {
		if _, seen := bySource[o.Source]; !seen {
			order = append(order, o.Source)
		}
		bySource[o.Source] = append(bySource[o.Source], o)
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	for _, id := range order {
		// The error is deliberately not consulted: NewSourceSeries puts its refusal
		// in the returned value as well, so measuring a refused series answers
		// MIXED_SERIES rather than WARMING_UP — the reason that means nothing needs
		// fixing. That is the property the sticky field exists for, and reading the
		// error here would only give this loop a second way to say the same thing.
		series, _ := NewSourceSeries(bySource[id]) //nolint:errcheck // see above
		v.Accelerations = append(v.Accelerations,
			Accelerate(series, FieldTradingValue, at, window))
	}
	return v
}

// --- one cycle -----------------------------------------------------------------------

// CycleOptions configures one turn of discovery.
type CycleOptions struct {
	// Market is the market to scan.
	Market string
	// Sources is the market's full panel. A source that cannot see this market must
	// be absent from it — see candidatesrc.Panel — because a responding source is
	// evidence about the market it was asked for.
	Sources []Source
	// Schedule decides which sources are due and carries the engine yield. Nil gets
	// a fresh one, which is right for a single scan and wrong for a loop; Watch
	// always supplies one.
	Schedule *Schedule
	// Backoff carries the retreat ladder. Nil gets a fresh one, same caveat.
	Backoff *Backoff
	// Thresholds are the veto's numbers.
	Thresholds VetoThresholds
	// Retention is how long raw observations are kept. Zero uses
	// DefaultRawRetention.
	Retention time.Duration
	// FreeSpaceFloor is the byte floor below which discovery stops writing. Zero
	// uses DefaultFreeSpaceFloor. A negative value is refused rather than treated as
	// "no floor": a floor nobody set is the default, never no floor at all — the
	// same rule VetoThresholds.MaxInputAge follows, and for the same reason.
	FreeSpaceFloor int64
	// EngineRunning reports whether the trading engine is running at an instant.
	// Nil means nobody is watching, which is not the same as "no engine" and is why
	// the yield is recorded rather than assumed.
	//
	// It is a function rather than a package because internal/enginelock is outside
	// this package's dependency closure, and the closure is the actual lock on the
	// isolation requirement (D7).
	EngineRunning func(at time.Time) bool
	// History and AccelerationWindow are Assess's, passed through.
	History, AccelerationWindow time.Duration
	// SkipAssessment omits the verdict pass. It exists for a caller that only wants
	// the scan written — the assessment is several queries per market — and it is
	// off by default because spec Requirement 8 puts the unmeasured count on the
	// scan output.
	SkipAssessment bool
}

// CycleResult is everything one turn did, including what it declined to do.
type CycleResult struct {
	// Market and At identify the turn.
	Market string
	At     time.Time
	// Halted reports that nothing was read and nothing was written, and HaltReason
	// says why in a sentence a person can act on. The only current cause is the free
	// space floor (D16).
	Halted     bool
	HaltReason string
	// Scan is what Collect did. Zero when Halted.
	Scan ScanResult
	// ScanErr is a scan that completed and still had something to report — a source
	// claiming a fallback it cannot have is the case. It is on the result rather
	// than returned, so a loop can keep going and still surface it.
	ScanErr string
	// Space is the free-space measurement this turn made.
	Space SpaceReport
	// Cleanup is the retention sweep, which runs on every turn (D16).
	Cleanup CleanupReport
	// Backoff is every retreat in force, and NotDue names the sources whose own
	// interval had not elapsed. The two are separate because they mean opposite
	// things about the panel: a retreat is a source we lost, a not-due source is the
	// schedule working.
	Backoff []Retreat
	NotDue  []SourceID
	// Quiet reports that the schedule had nothing due at all, so no source was read
	// and nothing was written.
	//
	// It is the other half of the sentence above, and it is a field rather than a
	// derivation from len(NotDue) because the fact it carries is what a caller has
	// to branch on. A turn with an empty panel is not a market that could not be
	// read: Collect answers ErrNoSourceAnswered for a panel of no sources, which is
	// true of what it was handed and false about the market, and a loop that stopped
	// there would stop promoting — every first_seen_at in the store then goes to the
	// implicit cooling clock and the expiry after it, inside forty minutes, while
	// the operator is told the sources went quiet (§5 review P0).
	//
	// Everything the turn owes besides the scan still happened: retention ran, the
	// free space was measured, and the assessment was taken — a quiet turn that
	// reported an empty market would be the same lie one level up.
	Quiet bool
	// EngineYield reports that discovery lengthened its intervals because the engine
	// is running (spec Requirement 7). Intervals is what they became.
	EngineYield bool
	Intervals   map[SourceID]time.Duration
	// Verdicts is every live candidate's assessment, and the three tallies are its
	// summary. Bands is keyed by the codes that have a shadow record — the two D18
	// left without a threshold.
	Verdicts  []Verdict
	Vetoes    VetoTally
	Crossings CrossingTally
	Bands     map[VetoCode]BandTally
}

// RateLimited reports that at least one source was lost to a 429 this turn.
func (r CycleResult) RateLimited() bool { return len(r.Scan.RateLimitedSources()) > 0 }

// Cycle runs one turn of discovery: enforce retention, check the room left, read
// what is due, record it, and assess what is there.
//
// The order is deliberate. Cleanup runs before the space check so that a store which
// has grown past the floor gets its own chance to come back under it — pruning and a
// WAL truncation are the two things that free space here, and running them after the
// halt would mean the halt was permanent. The space check then runs before the
// sources are read, not just before the write: reading and discarding would spend
// rate budget the engine shares on rows that could not be stored.
func Cycle(ctx context.Context, store *Store, opts CycleOptions) (CycleResult, error) {
	at := store.Now()
	sched := opts.Schedule
	if sched == nil {
		sched = NewSchedule(nil)
	}
	off := opts.Backoff
	if off == nil {
		off = NewBackoff()
	}
	floor := opts.FreeSpaceFloor
	if floor <= 0 {
		floor = DefaultFreeSpaceFloor
	}

	res := CycleResult{
		Market:    opts.Market,
		At:        at,
		Intervals: map[SourceID]time.Duration{},
		Bands:     map[VetoCode]BandTally{},
	}

	// 1. The engine outranks us, and the yield is a verb rather than an ordering.
	res.EngineYield = opts.EngineRunning != nil && opts.EngineRunning(at)
	sched.YieldToEngine(res.EngineYield)
	for _, src := range opts.Sources {
		res.Intervals[src.ID()] = sched.Every(src.ID())
	}

	// 2. Retention, on every turn, because the alternative is unbounded growth on
	//    the filesystem the ledger writes to (D16).
	res.Cleanup = runCleanup(ctx, store, opts.Retention)

	// 3. The room left. Unmeasured is not room.
	res.Space = measureSpace(store, floor)
	switch {
	case !res.Space.Checked:
		res.Halted = true
		res.HaltReason = fmt.Sprintf(
			"discovery is not writing: the free space of %s could not be measured (%s). "+
				"This store shares a filesystem with the order ledger, and an unmeasured "+
				"filesystem is not one with room on it",
			res.Space.Path, res.Space.Why)
		return res, nil
	case res.Space.Below():
		res.Halted = true
		res.HaltReason = fmt.Sprintf(
			"discovery is not writing: %s has %d bytes free, below the %d byte floor. "+
				"The order ledger is on this filesystem and it must not be the write that "+
				"fails, so this one stops first",
			res.Space.Path, res.Space.Free, res.Space.Floor)
		return res, nil
	}

	// 4. The panel for this turn: what is due, with the retreating ones present and
	//    failing rather than absent. See heldSource.
	panel := make([]Source, 0, len(opts.Sources))
	var notAsked []SourceID
	due := sched.DueSources(opts.Sources, at)
	dueIDs := make(map[SourceID]bool, len(due))
	for _, src := range due {
		dueIDs[src.ID()] = true
	}
	for _, src := range opts.Sources {
		id := src.ID()
		if !dueIDs[id] {
			res.NotDue = append(res.NotDue, id)
			notAsked = append(notAsked, id)
			continue
		}
		if retreat, held := off.Held(id, at); held {
			panel = append(panel, heldSource{id: id, retreat: retreat})
			continue
		}
		panel = append(panel, src)
		sched.Ran(id, at)
	}

	// 4b. Nothing due is the schedule working, and it is not a market failure.
	//
	// Collect answers ErrNoSourceAnswered for a panel of no sources, which is true
	// of what it was handed and says nothing about the market — and the loop above
	// it ended on that error. The three ways to get here are all ordinary: the
	// engine yield doubles the sources' interval while the tick stays where it was,
	// an operator asks for a tick below every source's interval, or the clock steps
	// backwards by at least one interval. See CycleResult.Quiet for what the loop
	// ending costs.
	//
	// The turn's other duties have already run — retention above, the free space
	// above that — and the assessment still runs below, because a quiet turn that
	// reported an empty market would be the same lie one level up.
	//
	// A panel that is empty because there are no sources CONFIGURED is not this: it
	// is a wiring defect, there is no schedule involved, and Collect's "0 source(s)
	// attempted" is the accurate thing to say about it. So the branch requires that
	// there was something to pass over.
	if len(panel) == 0 && len(opts.Sources) > 0 {
		res.Quiet = true
		res.Scan = ScanResult{
			Market:  strings.ToUpper(strings.TrimSpace(opts.Market)),
			At:      at.UTC(),
			Budgets: map[SourceID]Budget{},
		}
		res.Backoff = off.Active(at)
		return assessInto(ctx, store, opts, res, at)
	}

	scan, err := Collect(ctx, store, CollectOptions{
		Market: opts.Market, At: at, Sources: panel, NotAsked: notAsked,
	})
	res.Scan = scan
	if err != nil {
		if !errors.Is(err, ErrFallbackNotPossible) {
			return res, err
		}
		// A mis-wired source is a complete, usable result and a loud alarm. Returning
		// it as an error would stop a loop over a defect that does not stop the scan.
		res.ScanErr = err.Error()
	}

	// 5. The ladder moves on what the scan found, and it moves both ways.
	//
	// The recovery signal is Vouched rather than Budgets. A budget entry is written
	// for every source that answered, including the zero-row 200 this file refuses
	// to treat as coverage anywhere else — and load-shedding by returning an empty
	// list is a common way a rate-limited service degrades, so that reading is
	// exactly the one that must not take the ladder back to the bottom.
	for _, missing := range scan.Missing {
		if missing.RateLimited {
			off.Note(missing.Source, at)
		}
	}
	for _, id := range scan.Vouched {
		off.Recovered(id)
	}
	res.Backoff = off.Active(at)

	return assessInto(ctx, store, opts, res, at)
}

// assessInto adds the read-only assessment to a turn's result.
//
// It is shared by the ordinary path and the quiet one deliberately: a turn on
// which nothing was due still holds everything the last one found, and a caller
// that received zero verdicts from it would render an empty market.
func assessInto(ctx context.Context, store *Store, opts CycleOptions,
	res CycleResult, at time.Time) (CycleResult, error) {

	if opts.SkipAssessment {
		return res, nil
	}
	verdicts, err := Assess(ctx, store, AssessOptions{
		Market:             opts.Market,
		At:                 at,
		Thresholds:         opts.Thresholds,
		History:            opts.History,
		AccelerationWindow: opts.AccelerationWindow,
	})
	if err != nil {
		return res, err
	}
	res.Verdicts = verdicts
	res.Vetoes, res.Crossings, res.Bands = TallyVerdicts(verdicts)
	return res, nil
}

// TallyVerdicts reduces an assessment to the three counts a screen reads.
//
// It is exported because there are two screens. The console's discovery seam
// assembled the same three tallies by hand, so a fourth shadow band added here
// would have appeared on the scan output and not on /signals — and a band that is
// missing from a page renders as a band nobody crossed rather than as one nobody
// counted (§5 review P2).
func TallyVerdicts(in []Verdict) (VetoTally, CrossingTally, map[VetoCode]BandTally) {
	chases := make([]Chase, 0, len(in))
	var accelerations []Acceleration
	seenLate := make([]ShadowBand, 0, len(in))
	extended := make([]ShadowBand, 0, len(in))
	for _, v := range in {
		chases = append(chases, v.Chase)
		accelerations = append(accelerations, v.Accelerations...)
		seenLate = append(seenLate, v.SeenLateBand)
		extended = append(extended, v.ExtendedBand)
	}
	return TallyVetoes(chases), TallyCrossings(accelerations), map[VetoCode]BandTally{
		VetoSeenLate: TallyBands(VetoSeenLate, seenLate),
		VetoExtended: TallyBands(VetoExtended, extended),
	}
}

// runCleanup prunes the raw observations past their retention, sweeps the
// candidate summaries those observations have left behind, and reclaims the
// write-ahead log.
//
// What keeps a summary from being deleted while its readings are still on disk is
// the grace period, not the order of these two statements — both are evaluated
// against the same instant, so swapping them changes nothing. The order is for
// the reader: rows, then the summaries those rows no longer explain, then the WAL
// that both deletions left behind.
//
// A failure is reported rather than returned. The cleanup is a housekeeping duty of
// the cycle and a store that could not prune is still a store that can observe —
// refusing the whole turn would make the sweep's own failure the thing that stops
// discovery, which is a worse outcome than the growth it was preventing.
func runCleanup(ctx context.Context, store *Store, retention time.Duration) CleanupReport {
	if retention <= 0 {
		retention = DefaultRawRetention
	}
	out := CleanupReport{Retention: retention}
	pruned, err := store.PruneStale(ctx, retention)
	if err != nil {
		out.Why = err.Error()
		return out
	}
	out.Ran, out.Pruned = true, pruned

	// The grace period is the raw retention window: a summary stops being evidence
	// when the readings it describes are gone (D11, Manager ruling in design.md
	// "만료된 요약은 누가 지우는가"). Passing the same window rather than letting
	// PruneExpiredCandidates fall back to its own default keeps a caller who
	// shortens retention from being told the rows went and the summaries did not.
	summaries, err := store.PruneExpiredCandidates(ctx, store.Now(), retention)
	if err != nil {
		out.Why = err.Error()
		return out
	}
	out.Summaries = summaries

	busy, pages, err := store.Checkpoint(ctx)
	if err != nil {
		out.Why = err.Error()
		return out
	}
	out.Busy, out.Pages = busy, pages
	out.Checkpointed = !busy
	return out
}

// measureSpace probes the room left on the store's filesystem.
func measureSpace(store *Store, floor int64) SpaceReport {
	out := SpaceReport{Path: store.Path(), Floor: floor}
	free, err := store.FreeSpace()
	if err != nil {
		out.Why = err.Error()
		return out
	}
	out.Checked, out.Free = true, free
	return out
}

// --- the loop (task 5.2) -------------------------------------------------------------

// DefaultWatchInterval is the contract's `scheduler_tick_seconds: 15`.
const DefaultWatchInterval = 15 * time.Second

// MinWatchInterval is the floor the loop cannot be asked below.
//
// Three seconds, which is the shortest per-source floor in the contract
// (`sources.wts_rankings.floor_seconds: 3`). A loop ticking faster than the fastest
// source may be read would only ever produce cycles in which nothing is due, and the
// reason there is a floor at all is spec Requirement 7: discovery shares one
// account and one rate budget with the engine and the live verification, and being
// read-only does not make it free.
const MinWatchInterval = 3 * time.Second

// WatchInterval applies the floor.
//
// A zero or negative interval is the default and never "as fast as possible". The
// unset field switching a bound off is the failure shape this package keeps meeting
// — a zero DayHigh meaning "at the high", a zero threshold meaning "no threshold" —
// and an unset poll interval meaning "unlimited" is the same one aimed at a shared
// rate budget.
func WatchInterval(d time.Duration) time.Duration {
	switch {
	case d <= 0:
		return DefaultWatchInterval
	case d < MinWatchInterval:
		return MinWatchInterval
	default:
		return d
	}
}

// WatchOptions configures the repeated scan.
type WatchOptions struct {
	CycleOptions
	// Interval is the tick, subject to WatchInterval's floor.
	Interval time.Duration
	// Cycles bounds the run. Zero or negative runs until the context ends.
	Cycles int
	// OnCycle is called with each turn's result, in the loop's goroutine.
	OnCycle func(CycleResult)
	// OnError is called when a turn fails. Returning false stops the loop; nil
	// stops on the first error.
	//
	// It exists because the two failures are different: a store that cannot be
	// written is a reason to stop, and a market that could not be read for one tick
	// is not — and a loop that stopped on the second would end the session over a
	// blip.
	OnError func(error) bool
}

// Watch repeats Cycle until the context ends or the cycle budget runs out.
//
// The wait comes from the store's injected clock, which is the same clock every
// instant in the result came from. Two time sources would be one too many: the loop
// would sleep on wall time while its records were stamped from a fake, and a test
// that advanced the fake would measure a cadence the loop never had.
func Watch(ctx context.Context, store *Store, opts WatchOptions) error {
	interval := WatchInterval(opts.Interval)
	clk := store.Clock()

	for i := 0; opts.Cycles <= 0 || i < opts.Cycles; i++ {
		res, err := Cycle(ctx, store, opts.CycleOptions)
		if err != nil {
			if opts.OnError == nil || !opts.OnError(err) {
				return err
			}
		} else if opts.OnCycle != nil {
			opts.OnCycle(res)
		}
		if opts.Cycles > 0 && i == opts.Cycles-1 {
			break
		}
		if err := clk.Sleep(ctx, watchWait(opts.CycleOptions, interval, res.At)); err != nil {
			// A cancelled context is how a watch normally ends, and the caller is the
			// one that knows whether it asked for it.
			return err
		}
	}
	return nil
}

// watchWait is how long the loop sleeps after a turn.
//
// The tick and the per-source schedule are two different numbers and this is where
// they meet. `DefaultWatchInterval` and the official sources' interval are both 15
// seconds, which agreed right up until engineYieldFactor doubled one of them and
// not the other — and the loop then woke every other tick to find nothing due (§5
// review P0). Those turns are no longer errors, but they are still a retention
// sweep, a free-space probe and a whole market's assessment spent to learn that a
// source cannot be read yet.
//
// So: never sooner than the operator's tick, and never sooner than the schedule
// would allow any source to be read. The second is bounded by the first source
// that comes due — UntilNextDue is a minimum across the panel — so it can lengthen
// the wait to that instant and no further, whatever a single slow source is
// configured at.
//
// A nil schedule falls back to the tick. Cycle builds a fresh one per turn in that
// case, so every source is always due and there is nothing to wait for.
func watchWait(opts CycleOptions, tick time.Duration, at time.Time) time.Duration {
	if opts.Schedule == nil {
		return tick
	}
	until, known := opts.Schedule.UntilNextDue(opts.Sources, at)
	if !known || until <= tick {
		return tick
	}
	return until
}
