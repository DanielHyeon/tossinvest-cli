package candidate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// watch_test.go is section 5's loop: the repeated scan, what it enforces on every
// turn, and what it has to write down about its own retreats.
//
// Four of the tests here are about a scan that keeps working while doing less than
// it says. A discovery that quietly polls slower has lost the earliness it exists
// for and reports nothing; a discovery that quietly fills the disk takes the ledger
// with it. Both look like a calm market on a screen.

// spaceProber is a filesystem the test can shrink. FixedFSProber cannot: it answers
// the same thing forever, and the whole point of the free-space floor is that the
// answer changes while the process runs.
type spaceProber struct {
	mu   sync.Mutex
	info FSInfo
	err  error
}

func newSpaceProber(free int64) *spaceProber {
	return &spaceProber{info: FSInfo{Name: "ext4", FreeBytes: free, FreeMeasured: true}}
}

func (p *spaceProber) Probe(string) (FSInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.info, p.err
}

func (p *spaceProber) setFree(free int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.info.FreeBytes, p.info.FreeMeasured, p.err = free, true, nil
}

func (p *spaceProber) fail(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.err = err
}

// openStoreOver opens a store over a prober the test controls, with an injected
// clock so a whole session can be driven without waiting for one.
func openStoreOver(t *testing.T, prober FSProber, clk clock.Clock) *Store {
	t.Helper()
	s, err := Open(context.Background(), Options{
		Path:     filepath.Join(t.TempDir(), "candidates.db"),
		FSProber: prober,
		Clock:    clk,
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// plentyOfSpace is a free-space figure comfortably above the floor.
const plentyOfSpace = 100 << 30

// cycleOpts is a cycle over one source, with the contract's thresholds.
func cycleOpts(market string, sources ...Source) CycleOptions {
	return CycleOptions{
		Market:     market,
		Sources:    sources,
		Schedule:   NewSchedule(nil),
		Backoff:    NewBackoff(),
		Thresholds: VetoThresholds{NearHighDistancePct: DefaultNearHighThresholdPct},
	}
}

// --- task 5.3c: the ledger must never be the one that runs out of room ----------

// TestDiscoveryStopsWritingBeforeTheLedgerRunsOutOfSpace is D16.
//
// The two stores share a directory and therefore a filesystem, by design — the
// discovery store follows the ledger's data-directory rule. D11 costed the
// discovery side at roughly a gigabyte a day and treated it as a query-performance
// problem; the capacity conclusion was never drawn. When that filesystem fills, the
// write that fails is the ledger's: orders, fills and reconciliation, which
// .claude/CLAUDE.md names as High-risk.
//
// So the order of failure is fixed rather than left to whoever gets there first.
// Discovery stops writing, and says so, while there is still room.
func TestDiscoveryStopsWritingBeforeTheLedgerRunsOutOfSpace(t *testing.T) {
	prober := newSpaceProber(plentyOfSpace)
	clk := clock.NewFake(t0)
	s := openStoreOver(t, prober, clk)
	ctx := context.Background()

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	opts := cycleOpts(MarketKR, src)

	first, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	if first.Halted {
		t.Fatalf("the first cycle halted with %d bytes free: %s", plentyOfSpace, first.HaltReason)
	}
	if first.Scan.Observations == 0 {
		t.Fatal("the first cycle wrote nothing; the test cannot tell a halt from a no-op")
	}

	// The disk fills. The floor is where discovery gives up, not where the
	// filesystem does.
	prober.setFree(DefaultFreeSpaceFloor - 1)
	clk.Advance(time.Minute)
	calls := src.calls

	halted, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	if !halted.Halted {
		t.Error("discovery kept writing below the free-space floor; the next write to fail " +
			"is the ledger's, and that one is orders and fills")
	}
	if halted.Scan.Observations != 0 {
		t.Errorf("%d observations were written below the floor", halted.Scan.Observations)
	}
	if src.calls != calls {
		t.Error("the sources were still read below the floor; the rate budget is shared with " +
			"the engine and the rows could not have been stored anyway")
	}
	if halted.HaltReason == "" {
		t.Error("discovery stopped and did not say why")
	}
	if !halted.Space.Below() {
		t.Errorf("space report = %+v, want it below the floor", halted.Space)
	}

	// And the store still holds what it recorded before the floor was reached: a
	// halt is a stop, not a rollback.
	rows, err := s.Observations(ctx, MarketKR, "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(rows) == 0 {
		t.Error("halting on space discarded the history that was already recorded")
	}
}

// TestSpaceItCouldNotMeasureIsNotSpaceItHas.
//
// Absent is not zero, and here it is not "plenty" either. A probe that fails is the
// one case where continuing is a guess about the exact resource D16 says discovery
// must lose first, so the unmeasured answer halts and names itself.
func TestSpaceItCouldNotMeasureIsNotSpaceItHas(t *testing.T) {
	prober := newSpaceProber(plentyOfSpace)
	clk := clock.NewFake(t0)
	s := openStoreOver(t, prober, clk)

	prober.fail(errors.New("statfs: no such device"))
	res, err := Cycle(context.Background(), s, cycleOpts(MarketKR,
		&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}}))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if !res.Halted {
		t.Error("a free-space probe that failed was treated as room to write in")
	}
	if res.Space.Checked {
		t.Error("an unmeasured space report claims to have been checked")
	}
	if res.Space.Why == "" {
		t.Error("the space report is unmeasured and names no reason")
	}
}

// --- task 5.2: retention is enforced by the scan, not by somebody's intention ---

// TestACycleEnforcesRetentionItselfRatherThanLeavingItToACaller is D16's first
// finding: PruneObservations and PruneStale are an API nobody calls, so the default
// behaviour of this store is unbounded growth on the ledger's filesystem.
func TestACycleEnforcesRetentionItselfRatherThanLeavingItToACaller(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	// A row from three days ago: outside DefaultRawRetention (two trading days).
	// It belongs to a different symbol from the one this cycle raises, so the sweep
	// and the scan are not measuring the same record.
	old := t0.Add(-72 * time.Hour)
	if err := s.RecordObservations(ctx, []Observation{
		obs("000660", old, SourceOfficialTradingValue, Reported{Rank: 5, RankTotal: 100, Price: "9000"}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	// And it is kept alive the whole three days rather than promoted once, which is
	// D17's case exactly: a candidate that outlives its own raw rows keeps
	// first_seen_at and loses only its baseline. Promoting once and walking away
	// would make this a candidate that expired three days ago, and since task 5.8
	// the same cleanup sweeps those — the claim being made here would then be about
	// a row the summary sweep is entitled to take, not about the raw one's reach.
	for at := old; at.Before(t0); at = at.Add(DefaultStalenessTTL) {
		if _, err := s.Promote(ctx, MarketKR, "000660", at); err != nil {
			t.Fatalf("Promote at %s: %v", at, err)
		}
	}

	res, err := Cycle(ctx, s, cycleOpts(MarketKR,
		&fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if !res.Cleanup.Ran {
		t.Fatal("the cycle did not run the cleanup; retention is a function nobody calls " +
			"and the default behaviour is unbounded growth")
	}
	if res.Cleanup.Pruned != 1 {
		t.Errorf("pruned = %d, want 1 — the three-day-old row is outside the retention window",
			res.Cleanup.Pruned)
	}

	// The candidate summary is outside the sweep's reach, which is the whole of the
	// two-layer split (D11): a tidy-up that took first_seen_at with it would delete
	// this package's entire claim without failing anything.
	c, found, err := s.Candidate(ctx, MarketKR, "000660", t0)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if !found {
		t.Fatal("pruning the raw rows deleted the candidate summary; the tidy-up took " +
			"first_seen_at with it and this package's whole claim went with it")
	}
	if !c.FirstSeenAt.Equal(old) {
		t.Errorf("first_seen_at = %s, want %s", c.FirstSeenAt, old)
	}
	if rows, rerr := s.Observations(ctx, MarketKR, "000660", time.Time{}); rerr != nil {
		t.Fatalf("Observations: %v", rerr)
	} else if len(rows) != 0 {
		t.Errorf("%d raw rows survived the retention window", len(rows))
	}
}

// TestTheCleanupReclaimsTheWriteAheadLogToo.
//
// D16's second finding. The WAL does not check itself back into the table while a
// long-lived reader is attached, and the console's candidate screen is exactly such
// a reader — so a store whose rows have been deleted can still be holding the bytes
// the deletion was supposed to release, on the filesystem the ledger writes to.
func TestTheCleanupReclaimsTheWriteAheadLogToo(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	res, err := Cycle(context.Background(), s, cycleOpts(MarketKR,
		&fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if !res.Cleanup.Checkpointed {
		t.Error("the cleanup pruned rows and left the write-ahead log to grow beside the table")
	}
}

// --- task 5.8: the summary retention D11 promised, enforced by the cycle ---------

// TestACycleSweepsTheSummariesItsOwnRetentionHasOrphaned.
//
// Task 5.2 gave the raw rows an enforcer inside the cycle for D16's reason —
// cleanup is not optional on a store that shares the ledger's filesystem — and
// left the other half of D11's split with none. The `candidates` table is one row
// per symbol rather than the hundreds a scan writes, so this is not about volume;
// it is that the growth had no upper bound at all, and the sweep that would have
// found it is the same one already running every turn.
//
// The two halves are asserted in one cycle deliberately: what makes the summary
// deletable is that the raw rows explaining it are gone, so a test that did not
// watch both would not be measuring the reason.
func TestACycleSweepsTheSummariesItsOwnRetentionHasOrphaned(t *testing.T) {
	ctx := context.Background()
	// `now` is far enough past the old life that its expiry (cooled_at + cooling)
	// is more than one raw-retention window behind.
	old := t0.Add(-96 * time.Hour)
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	if err := s.RecordObservations(ctx, []Observation{
		obs("000660", old, SourceOfficialTradingValue, Reported{Rank: 5, RankTotal: 100, Price: "9000"}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := s.Promote(ctx, MarketKR, "000660", old); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	res, err := Cycle(ctx, s, cycleOpts(MarketKR,
		&fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if !res.Cleanup.Ran {
		t.Fatal("the cycle did not run the cleanup at all")
	}
	if res.Cleanup.Summaries != 1 {
		t.Errorf("the cycle swept %d expired summaries, want 1; D11 says a summary lives until "+
			"the candidate expires and nothing was enforcing it, so the table grew without "+
			"bound on the filesystem the order ledger writes to", res.Cleanup.Summaries)
	}
	if _, found, ferr := s.Candidate(ctx, MarketKR, "000660", t0); ferr != nil {
		t.Fatalf("Candidate: %v", ferr)
	} else if found {
		t.Error("the summary of a life that ended four days ago is still stored; its raw " +
			"observations are gone, so it now joins to nothing")
	}

	// And the candidate this very cycle raised is untouched. A sweep that took it
	// would delete first_seen_at, which is the whole claim this package makes.
	if _, found, ferr := s.Candidate(ctx, MarketKR, "005930", t0); ferr != nil {
		t.Fatalf("Candidate: %v", ferr)
	} else if !found {
		t.Error("the sweep deleted the candidate the same cycle had just promoted")
	}
}

// --- task 5.3b: the top of the priority table finally gets a verb ---------------

// TestDiscoveryBacksOffWhileTheEngineIsRunningAndSaysSo is spec Requirement 7's
// correction.
//
// The ranking is engine > verification > discovery. The middle one has the run lock
// and the bottom one is the thing being ranked; the top one had a sentence and no
// mechanism. The engine's own delay reaches stop-loss immediacy through fill
// detection, which is safety invariant 4 — so this yield is code rather than an
// operator's judgement, and it is recorded, because a scan that silently slowed down
// has lost earliness and reported nothing.
func TestDiscoveryBacksOffWhileTheEngineIsRunningAndSaysSo(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	sched := NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})
	opts := cycleOpts(MarketKR, &fakeSource{
		id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")},
	})
	opts.Schedule = sched

	idle, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle with no engine: %v", err)
	}
	if idle.EngineYield {
		t.Fatal("discovery yielded with no engine running")
	}
	quiet := sched.Every(SourceOfficialTradingValue)
	if quiet != 15*time.Second {
		t.Fatalf("interval with no engine = %s, want 15s", quiet)
	}

	clk.Advance(time.Minute)
	opts.EngineRunning = func(time.Time) bool { return true }
	busy, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle with the engine running: %v", err)
	}
	if got := sched.Every(SourceOfficialTradingValue); got <= quiet {
		t.Errorf("interval while the engine runs = %s, was %s with it idle — discovery did "+
			"not back off, and the priority table's top entry still has no verb", got, quiet)
	}
	if !busy.EngineYield {
		t.Error("discovery backed off and did not record that it had; a retreat nobody wrote " +
			"down is indistinguishable from a market that went quiet")
	}
}

// --- task 5.4: the retreat is a fact, not a delay somebody notices --------------

// TestTheBackoffLadderIsTheOneTheDesignGuaranteed.
//
// 30 → 60 → 120 → 300 seconds. The ladder is not decoration: MaxRankPriorAge and
// DefaultStalenessTTL are both derived from its last rung, so a ladder that climbed
// somewhere else would silently move two bounds that were set against it.
func TestTheBackoffLadderIsTheOneTheDesignGuaranteed(t *testing.T) {
	want := []time.Duration{30 * time.Second, 60 * time.Second, 120 * time.Second, 300 * time.Second}
	if len(BackoffLadder) != len(want) {
		t.Fatalf("ladder = %v, want %v", BackoffLadder, want)
	}
	b := NewBackoff()
	at := t0
	for i, expect := range want {
		r := b.Note(SourceOfficialGainers, at)
		if r.Step != i+1 {
			t.Errorf("rung %d reported step %d", i+1, r.Step)
		}
		if r.For != expect {
			t.Errorf("rung %d = %s, want %s", i+1, r.For, expect)
		}
		if !r.Until.Equal(at.Add(expect)) {
			t.Errorf("rung %d until = %s, want %s", i+1, r.Until, at.Add(expect))
		}
		at = at.Add(expect)
	}
	// The top rung is a ceiling, not a wrap-around.
	if r := b.Note(SourceOfficialGainers, at); r.For != 300*time.Second {
		t.Errorf("past the top rung = %s, want it to stay at 300s", r.For)
	}
}

// TestA429IsRecordedInTheScanResultAndNotOnlySuffered is task 5.4's point.
//
// Spec: 백오프 사실은 스캔 결과에 남아야 한다 — 조용히 느려진 스캔은 조기성을 잃고도
// 그것을 보고하지 않는다. The official ranking has no WTS fallback, so a 429 removes
// the source outright; the rate at which that happens is the only measurement
// anybody has of an undocumented limit.
func TestA429IsRecordedInTheScanResultAndNotOnlySuffered(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	good := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	limited := &fakeSource{
		id:  SourceOfficialGainers,
		err: fmt.Errorf("%w: the gainers ranking", ErrRateLimited),
	}
	opts := cycleOpts(MarketKR, good, limited)

	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if len(res.Backoff) != 1 {
		t.Fatalf("backoff = %v, want one retreat for the rate-limited source", res.Backoff)
	}
	r := res.Backoff[0]
	if r.Source != SourceOfficialGainers {
		t.Errorf("retreat source = %q, want %q", r.Source, SourceOfficialGainers)
	}
	if r.For != BackoffLadder[0] {
		t.Errorf("retreat = %s, want the first rung %s", r.For, BackoffLadder[0])
	}
}

// TestASourceHeldByTheBackoffIsNotAskedAndDoesNotVouch.
//
// Two things at once, and the second is the dangerous one. A source inside its
// retreat must not be called — that is what a retreat is — and it must not be
// treated as a source that answered, because cooling is a claim about evidence
// (task 2.8). A backed-off ranking counted as coverage would cool every candidate
// only it raises, and the cooling clock then expires them: a transient rate limit
// would erase the record of everything we found early.
func TestASourceHeldByTheBackoffIsNotAskedAndDoesNotVouch(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	gainers := &fakeSource{id: SourceOfficialGainers, rows: []Row{pricedRow("005930", 3, 100, "10000")}}
	value := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("000660", 1, 100, "20000")}}
	opts := cycleOpts(MarketKR, gainers, value)

	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}

	// Now the gainers list 429s, so 005930 has no supporter that can answer.
	gainers.err = fmt.Errorf("%w: the gainers ranking", ErrRateLimited)
	clk.Advance(time.Minute)
	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	calls := gainers.calls

	// The third cycle is past the source's own interval and still inside the retreat
	// window, which is the only arrangement that tests the retreat rather than the
	// schedule.
	clk.Advance(20 * time.Second)
	third, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("third Cycle: %v", err)
	}
	if gainers.calls != calls {
		t.Error("a source inside its own backoff window was called anyway; the retreat is " +
			"a record and not a delay")
	}
	if third.Scan.Cooled != 0 {
		t.Errorf("cooled %d candidates while a supporter was backing off; one rate limit "+
			"would erase every first_seen_at it protects", third.Scan.Cooled)
	}
	c, found, err := s.Candidate(ctx, MarketKR, "005930", clk.Now())
	if err != nil || !found {
		t.Fatalf("Candidate: %v (found %v)", err, found)
	}
	if c.State != StateActive {
		t.Errorf("005930 is %s; a backed-off supporter vouched for a symbol it never looked at",
			c.State)
	}
	// And the held source is named in the result rather than silently absent.
	held := false
	for _, m := range third.Scan.Missing {
		if m.Source == SourceOfficialGainers {
			held = true
		}
	}
	if !held {
		t.Error("the held source is missing from the scan result; a panel that shrank without " +
			"saying so is the degradation D4 refuses to hide")
	}
}

// TestARecoveredSourceStartsTheLadderAgainFromTheBottom.
//
// A retreat that never reset would leave a source at five minutes for the rest of
// the session after one bad minute, which is the earliness this change exists to buy
// being spent on a 429 that is long over.
func TestARecoveredSourceStartsTheLadderAgainFromTheBottom(t *testing.T) {
	b := NewBackoff()
	b.Note(SourceOfficialGainers, t0)
	b.Note(SourceOfficialGainers, t0.Add(30*time.Second))
	b.Recovered(SourceOfficialGainers)
	if r := b.Note(SourceOfficialGainers, t0.Add(5*time.Minute)); r.Step != 1 {
		t.Errorf("step after a recovery = %d, want 1 — the ladder never came back down",
			r.Step)
	}
}

// --- task 5.2: the interval floor ------------------------------------------------

// TestARepeatedScanCannotBeAskedToPollFasterThanItsFloor.
//
// spec Requirement 7: 반복 스캔에는 간격 하한이 있어야 한다. Discovery shares one
// account and one rate budget with the engine and the live verification, and a
// read-only command is still capable of spending it.
func TestARepeatedScanCannotBeAskedToPollFasterThanItsFloor(t *testing.T) {
	for _, tc := range []struct {
		name string
		ask  time.Duration
		want time.Duration
	}{
		{"unset takes the contract's tick", 0, DefaultWatchInterval},
		{"below the floor is raised to it", time.Millisecond, MinWatchInterval},
		{"negative is the contract's tick, never unlimited", -time.Hour, DefaultWatchInterval},
		{"above the floor is honoured", time.Minute, time.Minute},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := WatchInterval(tc.ask); got != tc.want {
				t.Errorf("WatchInterval(%s) = %s, want %s", tc.ask, got, tc.want)
			}
		})
	}
}

// TestTheWatchLoopWaitsOnTheInjectedClock.
//
// Nothing in this package may bind work to a clock a test cannot move — the guard
// in metrics_test.go lists eleven spellings of doing it, and a backoff ladder is
// naturally written with three of them. The loop's wait comes from the same injected
// clock its instants do, so this test runs a session in microseconds.
func TestTheWatchLoopWaitsOnTheInjectedClock(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	var (
		mu      sync.Mutex
		seen    []time.Time
		src     = &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
		opts    = WatchOptions{CycleOptions: cycleOpts(MarketKR, src), Interval: time.Minute, Cycles: 3}
		ctx, cn = context.WithCancel(context.Background())
	)
	defer cn()
	opts.OnCycle = func(r CycleResult) {
		mu.Lock()
		defer mu.Unlock()
		seen = append(seen, r.At)
	}

	done := make(chan error, 1)
	go func() { done <- Watch(ctx, s, opts) }()

	for i := 0; i < 2; i++ {
		if !clk.WaitForSleepers(1, 2*time.Second) {
			t.Fatalf("the loop never parked in the injected clock's Sleep after cycle %d; "+
				"the wait is bound to a clock no test can move", i+1)
		}
		clk.Advance(time.Minute)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Watch: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Watch did not finish its three cycles")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(seen) != 3 {
		t.Fatalf("cycles = %d, want 3", len(seen))
	}
	for i, want := range []time.Time{t0, t0.Add(time.Minute), t0.Add(2 * time.Minute)} {
		if !seen[i].Equal(want) {
			t.Errorf("cycle %d ran at %s, want %s", i+1, seen[i], want)
		}
	}
}

// --- task 5.1: what the scan output has to say ----------------------------------

// TestAScanCountsUnmeasuredVetoesAndNeverCallsAnAbsentThresholdAPass is the
// counting half of D18's consequence.
//
// While seen_late and extended have no threshold, Chase.Passed() is unreachable and
// the tally's Passed is permanently 0. That is not an outage — it is what D18, D10
// and the spec's "세 사유가 모두 측정되었고 모두 위험하지 않을 때만" produce together.
// The one wrong repair is to count THRESHOLD_ABSENT as a pass, and this is the test
// that fails if somebody makes it.
func TestAScanCountsUnmeasuredVetoesAndNeverCallsAnAbsentThresholdAPass(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	res, err := Cycle(ctx, s, cycleOpts(MarketKR, &fakeSource{
		id: SourceOfficialTradingValue,
		rows: []Row{
			pricedRow("005930", 12, 100, "10000"),
			pricedRow("000660", 90, 100, "20000"),
		},
	}))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if res.Vetoes.Total != 2 {
		t.Fatalf("tallied %d verdicts, want 2", res.Vetoes.Total)
	}
	if res.Vetoes.Passed != 0 {
		t.Errorf("passed = %d, want 0 — two of the three vetoes have no approved threshold, "+
			"so no candidate can have cleared all three. Counting THRESHOLD_ABSENT as a pass "+
			"is the one wrong repair", res.Vetoes.Passed)
	}
	if res.Vetoes.Unmeasured != 2 {
		t.Errorf("unmeasured = %d, want 2 — the screen's default reading has to be "+
			"'mostly unchecked'", res.Vetoes.Unmeasured)
	}
	if got := res.Vetoes.Reasons[VetoThresholdAbsent]; got != 4 {
		t.Errorf("THRESHOLD_ABSENT = %d, want 4 (two codes over two candidates)", got)
	}
	// near_high has a threshold and no candle, so it is unmeasured under its own
	// name — the candle budget working as designed rather than an outage.
	if got := res.Vetoes.Reasons[VetoUnmeasured(LevelNoDayHigh)]; got != 2 {
		t.Errorf("NO_DAY_HIGH = %d, want 2", got)
	}
}

// TestAScanReportsTheShadowRecordForEveryCodeThatHasOne.
//
// Task 5.1 wants crossings per shadow threshold in the output, and the words are
// kept apart from "passed the veto" on purpose. Two records exist: §3.5's
// acceleration thresholds and task 4.10's bands for the two vetoes with no
// threshold. Both are on the result, and both count the candidates they could not
// measure.
func TestAScanReportsTheShadowRecordForEveryCodeThatHasOne(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	res, err := Cycle(ctx, s, cycleOpts(MarketKR, &fakeSource{
		id:   SourceOfficialTradingValue,
		rows: []Row{pricedRow("005930", 12, 100, "10000")},
	}))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}

	if res.Crossings.Total == 0 {
		t.Error("no acceleration series was tallied; the shadow crossing record is what T3.2's " +
			"thresholds will be derived from")
	}
	if len(res.Crossings.Crossed) != len(ShadowThresholds) {
		t.Errorf("crossed keys = %v, want exactly %v", res.Crossings.Crossed, ShadowThresholds)
	}

	for _, code := range []VetoCode{VetoSeenLate, VetoExtended} {
		tally, ok := res.Bands[code]
		if !ok {
			t.Fatalf("no shadow band tally for %s; D18 leaves it with no threshold and the "+
				"band record is the only thing that will ever give it one", code)
		}
		if tally.Total != 1 {
			t.Errorf("%s band total = %d, want 1", code, tally.Total)
		}
		sum := tally.Measured
		for _, n := range tally.NotMeasured {
			sum += n
		}
		if sum != tally.Total {
			t.Errorf("%s band: measured + not measured = %d, total = %d", code, sum, tally.Total)
		}
	}
	// The seen_late band is measurable on the first scan: the first sighting is the
	// reading that promoted the candidate, and section 5 now stores its position.
	if got := res.Bands[VetoSeenLate].Measured; got != 1 {
		t.Errorf("seen_late band measured = %d, want 1 — the stored first rank is what makes "+
			"it measurable, and this scan is what wrote it", got)
	}
}

// TestTheDegradationOnTheResultNamesTheSourcesThatWentMissing is spec Requirement
// 8's last clause: 사유 없는 불리언 하나는 대응할 수 없는 표시다.
func TestTheDegradationOnTheResultNamesTheSourcesThatWentMissing(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	res, err := Cycle(context.Background(), s, cycleOpts(MarketKR,
		&fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}},
		&fakeSource{id: SourceWTSPopular, err: errors.New("the web session expired")},
	))
	if err != nil {
		t.Fatalf("Cycle: %v", err)
	}
	if !res.Scan.Degraded {
		t.Fatal("a scan that lost a source did not report itself degraded")
	}
	if len(res.Scan.Missing) != 1 || res.Scan.Missing[0].Source != SourceWTSPopular {
		t.Fatalf("missing = %+v, want the WTS popularity source named", res.Scan.Missing)
	}
	if res.Scan.Missing[0].Reason == "" {
		t.Error("the missing source has no reason; an expired session and a rate limit call " +
			"for opposite responses")
	}
}
