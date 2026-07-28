package candidate

// notdue_test.go is the §5 review's P0 and P1-1: the two things that go wrong
// when a source is absent from a pass because the SCHEDULE said so rather than
// because the source failed.
//
// They are the same mistake made in two places and they destroy the same record.
//
//	P0    an empty panel — nothing due at all — reached Collect, which answered
//	      ErrNoSourceAnswered because that is true of a panel with no sources in
//	      it. Cycle returned it, Watch called OnError, and tossctl's OnError
//	      returns false: the loop ended. Nobody promotes after that, the implicit
//	      cooling clock runs at last_seen_at + DefaultStalenessTTL and the expiry
//	      DefaultCoolingTTL after that, so inside forty minutes every
//	      first_seen_at in the store is gone — and the operator was told "no
//	      source answered", which reads as a broker or a market problem.
//	P1-1  a not-due source was absent from the cooling-eligibility panel too, so
//	      coverageAnswered read it as "a source that is gone" and stopped
//	      requiring it to answer. A candidate its other supporter no longer lists
//	      was then cooled by a scan that never asked the source which raised it —
//	      task 2.8's rule (냉각은 침묵이 아니라 증거로 한다) violated by
//	      construction rather than by a race.
//
// Both are reachable in the shipped wiring the moment the engine starts:
// engineYieldFactor doubles the official sources' 15s to 30s while the loop still
// ticks at DefaultWatchInterval, and candidatesrc.Panel gives --market US exactly
// three official sources at one interval.

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// --- driving the loop ----------------------------------------------------------

// watchRun is one driven session of Watch.
type watchRun struct {
	cycles []CycleResult
	// errs is every error the loop reported through OnError. A quiet cycle must
	// produce none: it is the schedule working, not a market failure.
	errs []error
	err  error
}

// driveWatch runs Watch against a fake clock and advances that clock one second
// at a time until the loop finishes.
//
// A second at a time rather than a jump to the instant the test expects: the
// loop's own choice of wait is part of what is under test here, and a driver that
// advanced by the amount it predicted would pass whatever the loop actually did.
// Every cycle's At comes from the same fake, so the cadence is read off the
// results rather than assumed by the driver.
func driveWatch(t *testing.T, s *Store, clk *clock.Fake, opts WatchOptions) watchRun {
	t.Helper()

	var (
		mu  sync.Mutex
		run watchRun
	)
	userCycle, userError := opts.OnCycle, opts.OnError
	opts.OnCycle = func(r CycleResult) {
		mu.Lock()
		run.cycles = append(run.cycles, r)
		mu.Unlock()
		if userCycle != nil {
			userCycle(r)
		}
	}
	opts.OnError = func(err error) bool {
		mu.Lock()
		run.errs = append(run.errs, err)
		mu.Unlock()
		if userError != nil {
			return userError(err)
		}
		// The default is the strict one: a reported error ends the session. A
		// quiet cycle that reached here at all would end it.
		return false
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Watch(ctx, s, opts) }()

	var advanced time.Duration
	idleSince := time.Now()
	for {
		select {
		case err := <-done:
			mu.Lock()
			defer mu.Unlock()
			run.err = err
			return run
		default:
		}
		if clk.Sleepers() > 0 {
			idleSince = time.Now()
			clk.Advance(time.Second)
			advanced += time.Second
			if advanced > 2*time.Hour {
				mu.Lock()
				n := len(run.cycles)
				mu.Unlock()
				t.Fatalf("the loop has not finished after two hours of fake time; %d cycles ran", n)
			}
			continue
		}
		if time.Since(idleSince) > 10*time.Second {
			mu.Lock()
			n := len(run.cycles)
			mu.Unlock()
			t.Fatalf("the loop neither parked in the injected clock nor finished; %d cycles ran", n)
		}
		time.Sleep(time.Millisecond)
	}
}

// quietCycles counts the turns on which the schedule had nothing due.
func quietCycles(run watchRun) int {
	n := 0
	for _, c := range run.cycles {
		if c.Quiet {
			n++
		}
	}
	return n
}

// requireLiveCandidate is the assertion every test in this file ends with: the
// candidate is still active and still carries the instant it was first seen at.
//
// That is the thing the dead loop destroyed, and it is destroyed silently — the
// store keeps answering, the screen keeps rendering, and the number that made
// discovery worth building is simply a later one.
func requireLiveCandidate(t *testing.T, s *Store, symbol string, at, firstSeen time.Time) {
	t.Helper()
	c, found, err := s.Candidate(context.Background(), MarketKR, symbol, at)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if !found {
		t.Fatalf("%s is not in the store at all", symbol)
	}
	if c.State != StateActive {
		t.Errorf("%s is %s at %s; a turn on which nothing was due cooled or expired it",
			symbol, c.State, at)
	}
	if !c.FirstSeenAt.Equal(firstSeen) {
		t.Errorf("first_seen_at = %s, want %s — the life was restarted, which is what makes "+
			"the whole record worthless: the next crossing looks like an early sighting",
			c.FirstSeenAt, firstSeen)
	}
}

// --- P0: a turn with nothing due is the schedule working ------------------------

// TestATurnWithNothingDueIsNotAMarketFailure is the defect at its smallest.
//
// Collect answers ErrNoSourceAnswered for a panel of no sources, which is true of
// what it was handed and false about the market. CycleResult's own documentation
// already says the two facts are different — "a retreat is a source we lost, a
// not-due source is the schedule working" — and NotDue never reached the error
// decision.
func TestATurnWithNothingDueIsNotAMarketFailure(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	opts := cycleOpts(MarketKR, src)
	opts.Schedule = NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})

	first, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	if first.Scan.Candidates != 1 {
		t.Fatalf("the first cycle raised %d candidates; the test cannot measure what a quiet "+
			"turn does to a candidate that is not there", first.Scan.Candidates)
	}

	// One second later nothing is due. Every source's interval is fifteen.
	clk.Advance(time.Second)
	calls := src.calls
	quiet, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("a cycle on which nothing was due returned %v. The schedule declining to "+
			"read a source is not a market that could not be read, and a loop that stops here "+
			"stops promoting — which expires every first_seen_at in the store inside forty "+
			"minutes while the operator is told the sources went quiet", err)
	}
	if !quiet.Quiet {
		t.Error("the cycle read nothing and does not report that nothing was due; the fact " +
			"that separates it from a failed scan is not on the result")
	}
	if len(quiet.NotDue) != 1 || quiet.NotDue[0] != SourceOfficialTradingValue {
		t.Errorf("not due = %v, want the one source named", quiet.NotDue)
	}
	if src.calls != calls {
		t.Errorf("a not-due source was read anyway (%d calls, was %d)", src.calls, calls)
	}
	if quiet.Scan.Cooled != 0 {
		t.Errorf("a turn that asked nobody cooled %d candidates", quiet.Scan.Cooled)
	}
	if quiet.Scan.Degraded {
		t.Error("a turn on which nothing was due reports itself degraded; nothing was lost, " +
			"and an operator reading 강등 goes looking for a source that is fine")
	}
	// And the assessment still ran: a quiet turn that reported an empty market
	// would be the same lie one level up.
	if quiet.Vetoes.Total != 1 {
		t.Errorf("the quiet turn assessed %d candidates, want 1 — a turn that read nothing "+
			"still holds everything the last one found", quiet.Vetoes.Total)
	}
	requireLiveCandidate(t, s, "005930", clk.Now(), t0)
}

// TestTheEngineYieldDoesNotEndTheDiscoveryLoop is the first of the three inputs.
//
// engineYieldFactor turns the official sources' 15s into 30s while the loop is
// still ticking at DefaultWatchInterval, so the tick after the engine starts has
// nothing due. For `--market US` that is unconditional: candidatesrc.Panel drops
// the WTS source for any market except KR, so the US panel is exactly three
// official sources at one interval and there is nothing else that could be due.
func TestTheEngineYieldDoesNotEndTheDiscoveryLoop(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	opts := WatchOptions{
		CycleOptions: cycleOpts(MarketKR, src),
		Interval:     DefaultWatchInterval,
		Cycles:       3,
	}
	opts.Schedule = NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})
	// The engine starts after the first turn. Both closures run in the loop's own
	// goroutine, which is the only one that touches this flag.
	engineRunning := false
	opts.EngineRunning = func(time.Time) bool { return engineRunning }
	opts.OnCycle = func(CycleResult) { engineRunning = true }

	run := driveWatch(t, s, clk, opts)
	if run.err != nil {
		t.Fatalf("the loop ended with %v. The engine starting is the ordinary case, and this "+
			"is the loop that keeps first_seen_at alive", run.err)
	}
	if len(run.errs) != 0 {
		t.Errorf("the loop reported %v; the schedule declining to read a source is not an "+
			"error and an operator reading stderr at 3am cannot act on it", run.errs)
	}
	if len(run.cycles) != 3 {
		t.Fatalf("cycles = %d, want 3 — the loop stopped when the engine started", len(run.cycles))
	}
	if got := quietCycles(run); got != 1 {
		t.Errorf("quiet cycles = %d, want exactly 1 (the turn straight after the yield "+
			"doubled the interval); cadence %s", got, cycleCadence(run))
	}
	if !run.cycles[1].Quiet {
		t.Errorf("cycle 2 is not the quiet one; cadence %s", cycleCadence(run))
	}
	if !run.cycles[1].EngineYield {
		t.Error("the quiet turn does not record that discovery had yielded to the engine, " +
			"which is the reason it was quiet")
	}
	if run.cycles[2].Scan.Observations == 0 {
		t.Error("the turn after the quiet one recorded nothing; the loop survived and stopped " +
			"reading, which reaches the same destruction more slowly")
	}
	requireLiveCandidate(t, s, "005930", clk.Now(), t0)
}

// TestATickBelowTheSourceIntervalDoesNotEndTheDiscoveryLoop is the second input,
// and the flag help invites it: "a floor of 3s".
//
// Under the fix the tick and the schedule are no longer two numbers that have to
// match. The loop never waits less than the operator's tick and never wakes
// before the schedule would let a source be read, so a tick below every source's
// interval stops manufacturing quiet turns rather than merely surviving them.
func TestATickBelowTheSourceIntervalDoesNotEndTheDiscoveryLoop(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	opts := WatchOptions{
		CycleOptions: cycleOpts(MarketKR, src),
		Interval:     MinWatchInterval,
		Cycles:       3,
	}
	opts.Schedule = NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})

	run := driveWatch(t, s, clk, opts)
	if run.err != nil {
		t.Fatalf("the loop ended with %v after being asked for the 3s tick its own flag help "+
			"offers", run.err)
	}
	if len(run.cycles) != 3 {
		t.Fatalf("cycles = %d, want 3", len(run.cycles))
	}
	for i, want := range []time.Duration{15 * time.Second, 15 * time.Second} {
		if got := run.cycles[i+1].At.Sub(run.cycles[i].At); got != want {
			t.Errorf("the loop waited %s between cycle %d and %d, want %s — the tick has to "+
				"lengthen to the instant a source may next be read rather than spinning "+
				"through turns on which nothing is due; cadence %s",
				got, i+1, i+2, want, cycleCadence(run))
		}
	}
	if got := quietCycles(run); got != 0 {
		t.Errorf("quiet cycles = %d, want 0; cadence %s", got, cycleCadence(run))
	}
	for i, c := range run.cycles {
		if c.Scan.Observations == 0 {
			t.Errorf("cycle %d recorded nothing", i+1)
		}
	}
	requireLiveCandidate(t, s, "005930", clk.Now(), t0)
}

// steppedClock is a wall clock that can be moved without moving the timer a
// sleeper is waiting on.
//
// That is the shape of the third input and it cannot be modelled with the fake
// alone: clock.Fake's Sleep parks against an absolute deadline on the same axis
// as Now, whereas the real one sleeps on a monotonic timer and reads a wall clock
// that NTP or a resumed VM can step underneath it. So the loop's wait comes from
// the fake and the instant every cycle is stamped with comes from the fake plus a
// skew the test can change.
type steppedClock struct {
	fake *clock.Fake
	mu   sync.Mutex
	skew time.Duration
}

func (c *steppedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.fake.Now().Add(c.skew)
}

func (c *steppedClock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }

func (c *steppedClock) Sleep(ctx context.Context, d time.Duration) error {
	return c.fake.Sleep(ctx, d)
}

func (c *steppedClock) step(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.skew += d
}

// TestABackwardClockStepDoesNotEndTheDiscoveryLoop is the third input.
//
// Schedule.due compares `!at.Before(last.Add(every))`, so an instant that moves
// backwards by at least one interval makes every source not-due at once — one
// NTP correction, one resumed VM. It is also the input with no operator in it:
// nobody asked for anything, and the loop would have ended with "no source
// answered" on a market that was answering fine.
func TestABackwardClockStepDoesNotEndTheDiscoveryLoop(t *testing.T) {
	fake := clock.NewFake(t0)
	clk := &steppedClock{fake: fake}
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{pricedRow("005930", 12, 100, "10000")}}
	opts := WatchOptions{
		CycleOptions: cycleOpts(MarketKR, src),
		Interval:     DefaultWatchInterval,
		Cycles:       3,
	}
	opts.Schedule = NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})
	stepped := false
	opts.OnCycle = func(CycleResult) {
		if !stepped {
			stepped = true
			clk.step(-20 * time.Minute)
		}
	}

	run := driveWatch(t, s, fake, opts)
	if run.err != nil {
		t.Fatalf("the loop ended with %v after the clock stepped backwards", run.err)
	}
	if len(run.cycles) != 3 {
		t.Fatalf("cycles = %d, want 3; cadence %s", len(run.cycles), cycleCadence(run))
	}
	if !run.cycles[1].Quiet {
		t.Errorf("the turn after the step is not reported as one with nothing due; cadence %s",
			cycleCadence(run))
	}
	if run.cycles[1].Scan.Cooled != 0 {
		t.Errorf("the turn after the step cooled %d candidates", run.cycles[1].Scan.Cooled)
	}
	if run.cycles[2].Scan.Observations == 0 {
		t.Error("the loop never read a source again after the step")
	}
	requireLiveCandidate(t, s, "005930", clk.Now(), t0)
}

// --- P1-1: a source nobody asked cannot vouch -----------------------------------

// TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised is task 2.8's
// rule with its other hole closed.
//
// heldSource was given the present-and-failing shape precisely so a 429'd source
// could not vouch. The other reason a source is absent from a pass — the schedule
// has not come round to it — never got the same treatment: Cycle handed Collect
// only the due sources, so coverageAnswered saw a not-due supporter as a source
// that is gone and stopped requiring it to answer.
//
// The wiring reaches this on alternate ticks the moment the engine starts: the
// yield puts the official rankings at 30s against WTS's 10s, so the 30-row
// popularity list — the most volatile source in the panel — decides cooling on
// its own.
func TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	gainers := &fakeSource{
		id:   SourceOfficialGainers,
		rows: []Row{pricedRow("005930", 3, 100, "10000")},
	}
	popular := &fakeSource{
		id: SourceWTSPopular,
		rows: []Row{
			pricedRow("005930", 7, 30, "10000"),
			pricedRow("000660", 1, 30, "20000"),
		},
	}
	opts := cycleOpts(MarketKR, gainers, popular)
	opts.Schedule = NewSchedule(map[SourceID]Interval{
		SourceOfficialGainers: {Every: 15 * time.Second, Floor: 5 * time.Second},
		SourceWTSPopular:      {Every: 5 * time.Second, Floor: 3 * time.Second},
	})

	if _, err := Cycle(ctx, s, opts); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}
	c, found, err := s.Candidate(ctx, MarketKR, "005930", clk.Now())
	if err != nil || !found {
		t.Fatalf("Candidate: %v (found %v)", err, found)
	}
	if len(c.Sources) != 2 {
		t.Fatalf("005930 has supporters %v, want both; the test measures what happens when "+
			"only one of them is asked", c.Sources)
	}

	// Five seconds on: the popularity list is due, the gainers ranking is not, and
	// the popularity list no longer carries 005930.
	popular.rows = []Row{pricedRow("000660", 1, 30, "20000")}
	clk.Advance(5 * time.Second)
	res, err := Cycle(ctx, s, opts)
	if err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	if len(res.NotDue) != 1 || res.NotDue[0] != SourceOfficialGainers {
		t.Fatalf("not due = %v, want the gainers ranking — the arrangement under test is one "+
			"supporter asked and one not", res.NotDue)
	}
	if res.Scan.Cooled != 0 {
		t.Errorf("the scan cooled %d candidates while a supporter had not been asked. "+
			"냉각은 침묵이 아니라 증거로 한다 (task 2.8): a source that did not answer must "+
			"not vouch, and the schedule not reaching a source is not evidence about the "+
			"symbols it raises", res.Scan.Cooled)
	}
	requireLiveCandidate(t, s, "005930", clk.Now(), t0)

	// And the guard is narrow: a candidate every asked source still lists is
	// untouched, and one whose only supporter answered and dropped it still cools.
	// Otherwise the repair would simply make cooling unreachable, which is the
	// staleness fallback taking over — the path store.go reserves for the scan
	// that died.
	if _, found, ferr := s.Candidate(ctx, MarketKR, "000660", clk.Now()); ferr != nil || !found {
		t.Fatalf("000660: %v (found %v)", ferr, found)
	}
}

// TestASourceTheSchedulePassedOverIsNotASourceThatIsGone keeps the P1-1 repair
// from swallowing §2-5.
//
// A supporter removed from the configuration — an operator dropping WTS, which
// the design calls routine — must still allow cooling, or those candidates become
// permanently un-coolable and can only leave through the staleness fallback. The
// difference is that one source is in the market's panel and was passed over, and
// the other is not in the panel at all.
func TestASourceTheSchedulePassedOverIsNotASourceThatIsGone(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)
	ctx := context.Background()

	gainers := &fakeSource{
		id:   SourceOfficialGainers,
		rows: []Row{pricedRow("005930", 3, 100, "10000")},
	}
	popular := &fakeSource{
		id:   SourceWTSPopular,
		rows: []Row{pricedRow("005930", 7, 30, "10000")},
	}
	both := cycleOpts(MarketKR, gainers, popular)
	both.Schedule = NewSchedule(nil)
	if _, err := Cycle(ctx, s, both); err != nil {
		t.Fatalf("first Cycle: %v", err)
	}

	// The operator drops WTS: it is no longer in the panel at all. The gainers
	// ranking answers with a list that no longer carries the symbol — a list with
	// rows in it, because a zero-row reading responds without vouching (§2-2) and
	// would make this test pass for the wrong reason.
	gainers.rows = []Row{pricedRow("000660", 5, 100, "5000")}
	clk.Advance(time.Minute)
	onlyOfficial := cycleOpts(MarketKR, gainers)
	onlyOfficial.Schedule = NewSchedule(nil)
	res, err := Cycle(ctx, s, onlyOfficial)
	if err != nil {
		t.Fatalf("second Cycle: %v", err)
	}
	if res.Scan.Cooled != 1 {
		t.Errorf("cooled = %d, want 1 — a supporter that is no longer in the panel is a "+
			"source that is gone, not missing evidence, and treating it as the latter makes "+
			"the candidate un-coolable for the rest of its life", res.Scan.Cooled)
	}
}

// cycleCadence renders a run's instants and quiet flags for a failure message.
func cycleCadence(run watchRun) string {
	if len(run.cycles) == 0 {
		return "(no cycles)"
	}
	out := ""
	base := run.cycles[0].At
	for i, c := range run.cycles {
		if i > 0 {
			out += ", "
		}
		mark := "read"
		if c.Quiet {
			mark = "quiet"
		}
		out += c.At.Sub(base).String() + " " + mark
	}
	return "[" + out + "]"
}

// TestAPanelWithNoSourcesInItIsStillAnError keeps the quiet branch narrow.
//
// "Nothing was due" is a claim about a schedule, and a panel that is empty because
// nobody configured a source has no schedule in it — there is nothing to have
// passed over. Collect's "0 source(s) attempted" is the accurate sentence for that
// one, and it is also where the market itself gets validated.
func TestAPanelWithNoSourcesInItIsStillAnError(t *testing.T) {
	clk := clock.NewFake(t0)
	s := openStoreOver(t, newSpaceProber(plentyOfSpace), clk)

	res, err := Cycle(context.Background(), s, cycleOpts(MarketKR))
	if err == nil {
		t.Fatalf("a cycle over a panel of no sources succeeded (quiet = %v). That is a wiring "+
			"defect rather than a schedule, and reporting it as an ordinary quiet turn hides "+
			"a configuration with nothing in it", res.Quiet)
	}
	if !errors.Is(err, ErrNoSourceAnswered) {
		t.Errorf("error = %v, want ErrNoSourceAnswered", err)
	}
	if res.Quiet {
		t.Error("a panel with no sources configured reports itself as a turn on which nothing " +
			"was due")
	}
}
