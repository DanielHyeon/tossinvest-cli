package candidate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeSource is a source under the test's control: what it answers, whether it
// fails, and what budget it reports.
type fakeSource struct {
	id       SourceID
	rows     []Row
	err      error
	budget   Budget
	fellBack bool
	calls    int
}

func (f *fakeSource) ID() SourceID { return f.id }

func (f *fakeSource) Read(ctx context.Context, market string) (Reading, error) {
	f.calls++
	if f.err != nil {
		return Reading{}, f.err
	}
	return Reading{Rows: f.rows, ViaFallback: f.fellBack, Budget: f.budget}, nil
}

func row(symbol string, rank, total int) Row {
	return Row{
		Symbol: symbol, Rank: rank, RankTotal: total,
		TradingValue: "1000000", TradingVolume: "500", Price: "70000",
	}
}

// --- task 2.1: the official sources have to be sufficient on their own ----------

// TestOfficialSourcesAloneProduceCandidates is spec Requirement 5's hard half.
//
// WTS sessions expire — the roadmap treats that as normal, not exceptional — so a
// configuration in which discovery stops when WTS does is a configuration that
// stops on an ordinary day. This is the test that fails if the official adapters
// ever become a supplement to the WTS ones rather than the other way round.
func TestOfficialSourcesAloneProduceCandidates(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}},
		},
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if result.Candidates != 1 {
		t.Fatalf("official rankings alone produced %d candidates, want 1", result.Candidates)
	}
	if result.Degraded {
		t.Error("a scan whose every source answered is reported as degraded")
	}

	cands, err := s.Candidates(ctx, t0)
	if err != nil {
		t.Fatalf("Candidates: %v", err)
	}
	if len(cands) != 1 || cands[0].Symbol != "005930" {
		t.Fatalf("stored candidates = %v, want one 005930", cands)
	}
	if len(cands[0].Sources) != 1 || cands[0].Sources[0] != SourceOfficialTradingValue {
		t.Errorf("sources = %v, want [%s]", cands[0].Sources, SourceOfficialTradingValue)
	}
}

// --- task 2.3: a source that fails must not take the scan with it ----------------

// TestWTSFailingEntirelyStillYieldsCandidates is the spec scenario "WTS 실패에도
// 후보가 나온다", and the sentence after it: the degradation is recorded rather
// than absorbed.
func TestWTSFailingEntirelyStillYieldsCandidates(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}},
			&fakeSource{id: SourceWTSPopular, err: errors.New("session expired")},
			&fakeSource{id: SourceWTSTheme, err: errors.New("session expired")},
		},
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if result.Candidates != 1 {
		t.Fatalf("candidates = %d, want 1 — two failing WTS sources stopped the scan",
			result.Candidates)
	}
	if !result.Degraded {
		t.Error("two sources failed and the scan does not report itself degraded")
	}
	if result.Attempted != 3 || result.Responded != 1 {
		t.Errorf("completeness = %d/%d, want 1/3", result.Responded, result.Attempted)
	}
	if len(result.Missing) != 2 {
		t.Fatalf("missing = %v, want the two WTS sources", result.Missing)
	}
	// The reason travels with the source id. "degraded: true" alone is a flag an
	// operator cannot act on — spec Requirement "강등을 빠진 원천 이름으로 말한다".
	for _, m := range result.Missing {
		if m.Reason == "" {
			t.Errorf("%s went missing with no reason", m.Source)
		}
	}

	// And the candidate itself carries the completeness of the scan that made it.
	cands, err := s.Candidates(ctx, t0)
	if err != nil {
		t.Fatalf("Candidates: %v", err)
	}
	if len(cands) != 1 {
		t.Fatalf("stored %d candidates, want 1", len(cands))
	}
	if cands[0].SourcesAttempted != 3 || cands[0].SourcesResponded != 1 {
		t.Errorf("candidate completeness = %d/%d, want 1/3",
			cands[0].SourcesResponded, cands[0].SourcesAttempted)
	}
	if !cands[0].Degraded {
		t.Error("the candidate does not carry the degradation of the scan that produced it")
	}
	if cands[0].Complete() {
		t.Error("Complete() is true for a candidate from a 1-of-3 scan")
	}
}

// TestEverySourceFailingIsAnError separates "we looked and found nothing" from
// "we could not look". Zero candidates from a healthy panel is information; zero
// candidates because nothing answered is not, and reporting them the same way is
// how an empty screen comes to mean "quiet market".
func TestEverySourceFailingIsAnError(t *testing.T) {
	s, _ := openStore(t)
	_, err := Collect(context.Background(), s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, err: errors.New("429")},
			&fakeSource{id: SourceWTSPopular, err: errors.New("session expired")},
		},
	})
	if !errors.Is(err, ErrNoSourceAnswered) {
		t.Fatalf("Collect with every source failing = %v, want ErrNoSourceAnswered", err)
	}
}

// --- task 2.4: one symbol, two sources, one candidate ---------------------------

// TestTwoSourcesRaisingOneSymbolMakeOneCandidate is D1's identity rule.
//
// The identity is (market, symbol) and not the source, so a symbol on both the
// official trading-value ranking and the WTS popularity list is one candidate
// that two sources supported — not two candidates racing each other through the
// rest of the pipeline with half the evidence each.
func TestTwoSourcesRaisingOneSymbolMakeOneCandidate(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}},
			&fakeSource{id: SourceWTSPopular, rows: []Row{row("005930", 7, 30)}},
		},
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if result.Candidates != 1 {
		t.Fatalf("candidates = %d, want 1 — the same symbol became two", result.Candidates)
	}
	// Both readings are kept: they are different sources' answers about the same
	// symbol at the same instant, and section 3 needs them apart.
	if result.Observations != 2 {
		t.Errorf("observations = %d, want 2 — a source's reading was dropped", result.Observations)
	}

	cands, err := s.Candidates(ctx, t0)
	if err != nil {
		t.Fatalf("Candidates: %v", err)
	}
	if len(cands) != 1 {
		t.Fatalf("stored %d candidates, want 1", len(cands))
	}
	if len(cands[0].Sources) != 2 {
		t.Errorf("sources = %v, want both", cands[0].Sources)
	}
}

// --- task 2.5: the fallback is a degradation, and only prices have one ----------

// TestAFallbackReadingIsRecordedAsOne is D14 for the source that actually has a
// fallback. Over-polling the official prices endpoint produces no failure at all
// — the hybrid client answers from WTS and the screen looks normal — so the only
// way the source mix changing becomes visible is if the reading says so.
func TestAFallbackReadingIsRecordedAsOne(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialPrices, rows: []Row{row("005930", 0, 0)}, fellBack: true},
		},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	obs, err := s.Observations(ctx, MarketKR, "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("recorded %d observations, want 1", len(obs))
	}
	if !obs[0].ViaFallback {
		t.Error("a reading the hybrid client filled from WTS reads back as a direct official one")
	}
}

// TestARankingCannotClaimToHaveFallenBack is D14's 2026-07-28 correction, as a
// guard.
//
// hybrid's Rankings, MarketIndicatorPrices, MarketIndicatorCandles and
// MarketInvestorTrading are official-only — the code says so in as many words:
// "no WTS fallback (WTS popularity ranking is a different dataset)". So a ranking
// reading marked ViaFallback is not a degradation that happened; it is a wiring
// mistake, and recording it as a degradation would hide the mistake behind a
// plausible-looking fact.
func TestARankingCannotClaimToHaveFallenBack(t *testing.T) {
	s, _ := openStore(t)
	_, err := Collect(context.Background(), s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}, fellBack: true},
		},
	})
	if !errors.Is(err, ErrFallbackNotPossible) {
		t.Fatalf("a ranking claiming a fallback = %v, want ErrFallbackNotPossible", err)
	}
}

// --- task 2.5b: a ranking 429 removes the source outright ------------------------

// TestARankingFailureIsALossAndNotADegradedFallback.
//
// The correction to D14 inverts the risk for rankings. There is no fallback to
// absorb a 429, so the source is simply gone — and spec Requirement 5 names the
// official ranking as the one that must suffice alone. The rate at which this
// happens is the measurement that tells us the undocumented RANKING limit, but
// only if the loss is counted rather than smoothed over.
func TestARankingFailureIsALossAndNotADegradedFallback(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, err: ErrRateLimited},
			&fakeSource{id: SourceWTSPopular, rows: []Row{row("005930", 3, 30)}},
		},
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if len(result.Missing) != 1 || result.Missing[0].Source != SourceOfficialTradingValue {
		t.Fatalf("missing = %v, want the official ranking", result.Missing)
	}
	if !result.Missing[0].RateLimited {
		t.Error("a 429 on the official ranking is not counted as a rate-limited loss; " +
			"the loss rate is how the undocumented RANKING limit gets measured")
	}
	if !result.Degraded {
		t.Error("losing the source that must suffice alone is not reported as degradation")
	}
}

// --- task 2.6: the budget is read from the response, not from the 429 ------------

// TestTheRemainingBudgetIsRecordedWhenTheCallSucceeds is spec's "남은 예산이
// 기록된다" and D13 decision 2.
//
// Nothing in this repository reads X-RateLimit-* today, so the budget is learned
// only when a 429 arrives — which is exactly too late, and for the ranking source
// (no fallback) it is too late by one whole source.
func TestTheRemainingBudgetIsRecordedWhenTheCallSucceeds(t *testing.T) {
	s, _ := openStore(t)
	reset := t0.Add(time.Minute)

	result, err := Collect(context.Background(), s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{
				id:     SourceOfficialTradingValue,
				rows:   []Row{row("005930", 1, 100)},
				budget: Budget{Limit: 10, Remaining: 3, Reset: reset, Reported: true},
			},
		},
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	got, ok := result.Budgets[SourceOfficialTradingValue]
	if !ok {
		t.Fatal("a 200 response reported its remaining budget and the scan did not keep it")
	}
	if got.Remaining != 3 || got.Limit != 10 || !got.Reset.Equal(reset) {
		t.Errorf("budget = %+v, want limit 10 remaining 3 reset %v", got, reset)
	}
	if got.Tight(2) {
		t.Error("3 of 10 remaining reads as tight against a floor of 2")
	}
	if !got.Tight(5) {
		t.Error("3 of 10 remaining does not read as tight against a floor of 5")
	}
}

// TestASourceThatReportsNoBudgetIsNotRecordedAsZero is the absent-vs-zero rule
// applied to the rate budget. WTS reports no such header at all; storing that as
// "0 calls remaining" would make the scheduler back off from a source that never
// asked it to.
func TestASourceThatReportsNoBudgetIsNotRecordedAsZero(t *testing.T) {
	s, _ := openStore(t)
	result, err := Collect(context.Background(), s, CollectOptions{
		Market: MarketKR,
		At:     t0,
		Sources: []Source{
			&fakeSource{id: SourceWTSPopular, rows: []Row{row("005930", 1, 30)}},
		},
	})
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	got := result.Budgets[SourceWTSPopular]
	if got.Reported {
		t.Error("a source that sent no rate headers is recorded as having reported a budget")
	}
	if got.Tight(5) {
		t.Error("an unreported budget reads as tight; unmeasured is not zero")
	}
}

// --- task 2.7: the interval belongs to the source, not to the market -------------

// TestASlowSourceDoesNotHoldBackAFastOne is the spec scenario "원천마다 다른
// 간격을 가진다".
//
// The limits differ by orders of magnitude — WTS has no observed limit, the
// official RANKING group is not in the published table at all — so one interval
// per market ties every source to the strictest one and throws away the earliness
// this change exists to buy.
func TestASlowSourceDoesNotHoldBackAFastOne(t *testing.T) {
	sched := NewSchedule(map[SourceID]Interval{
		SourceWTSPopular:           {Every: 5 * time.Second, Floor: 3 * time.Second},
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})

	if !sched.Due(SourceWTSPopular, t0) || !sched.Due(SourceOfficialTradingValue, t0) {
		t.Fatal("nothing is due on the first tick")
	}
	sched.Ran(SourceWTSPopular, t0)
	sched.Ran(SourceOfficialTradingValue, t0)

	at := t0.Add(5 * time.Second)
	if !sched.Due(SourceWTSPopular, at) {
		t.Error("the fast source is not due at its own interval")
	}
	if sched.Due(SourceOfficialTradingValue, at) {
		t.Error("the slow source came due at the fast source's interval")
	}
	if !sched.Due(SourceOfficialTradingValue, t0.Add(15*time.Second)) {
		t.Error("the slow source is not due at its own interval")
	}
}

// TestAnIntervalBelowTheFloorIsRefused keeps the rate-budget promise honest: the
// floor is the spec's, and a configuration that undercuts it is a configuration
// that spends the engine's budget on discovery.
func TestAnIntervalBelowTheFloorIsRefused(t *testing.T) {
	sched := NewSchedule(map[SourceID]Interval{
		SourceWTSPopular: {Every: time.Second, Floor: 3 * time.Second},
	})
	if got := sched.Every(SourceWTSPopular); got != 3*time.Second {
		t.Errorf("a 1s interval under a 3s floor resolved to %v, want the floor", got)
	}
}

// TestTheEngineTakesPrecedenceOverDiscovery is spec Requirement 7's top priority,
// which had an ordering and no mechanism until the section-1 review.
//
// The engine's delay reaches stop-loss immediacy through fill detection, so this
// yield is not something an operator remembers to do — it is something the code
// does.
func TestTheEngineTakesPrecedenceOverDiscovery(t *testing.T) {
	sched := NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
	})
	quiet := sched.Every(SourceOfficialTradingValue)

	sched.YieldToEngine(true)
	busy := sched.Every(SourceOfficialTradingValue)
	if busy <= quiet {
		t.Errorf("with the engine running the interval is %v, was %v — discovery did not yield",
			busy, quiet)
	}

	sched.YieldToEngine(false)
	if got := sched.Every(SourceOfficialTradingValue); got != quiet {
		t.Errorf("after the engine stopped the interval stayed at %v, want %v", got, quiet)
	}
}

// --- the reading that does not become a candidate --------------------------------

// TestASymbolThatLeavesEveryListCools closes the loop between a scan and the
// lifecycle: a candidate the sources stopped raising must be cooled by the scan
// that noticed, not left for the staleness fallback. The fallback exists for the
// scan that died, not as the normal path.
func TestASymbolThatLeavesEveryListCools(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	present := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{present},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	// Next scan: the symbol is gone from the list, another one is there.
	present.rows = []Row{row("000660", 1, 100)}
	at := t0.Add(15 * time.Second)
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: at, Sources: []Source{present},
	}); err != nil {
		t.Fatalf("Collect #2: %v", err)
	}

	gone, found, err := s.Candidate(ctx, MarketKR, "005930", at)
	if err != nil || !found {
		t.Fatalf("Candidate: %v found=%v", err, found)
	}
	if gone.State != StateCooled {
		t.Errorf("a symbol that left every list is %v, want %v", gone.State, StateCooled)
	}
	if !gone.FirstSeenAt.Equal(t0) {
		t.Errorf("cooling moved first_seen_at to %v", gone.FirstSeenAt)
	}
}

// TestAScanDoesNotCoolASymbolItDidNotLookFor is the other half, and the one that
// would quietly destroy the record: a scan whose ranking source failed has not
// established that anything left the list. Cooling on that basis would cool every
// candidate on every 429, and the cooling clock would then expire them — so a
// rate limit would erase the first_seen_at of everything being tracked.
func TestAScanDoesNotCoolASymbolItDidNotLookFor(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}},
		},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	at := t0.Add(15 * time.Second)
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: at,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, err: ErrRateLimited},
			&fakeSource{id: SourceWTSPopular, rows: []Row{row("000660", 1, 30)}},
		},
	}); err != nil {
		t.Fatalf("Collect #2: %v", err)
	}

	still, found, err := s.Candidate(ctx, MarketKR, "005930", at)
	if err != nil || !found {
		t.Fatalf("Candidate: %v found=%v", err, found)
	}
	if still.State != StateActive {
		t.Errorf("a 429 on the ranking source cooled a candidate it never looked for: state = %v",
			still.State)
	}
}

// TestOneSurvivingSupporterIsNotEnoughToCool is where the narrow rule and the
// obvious one visibly disagree.
//
// A candidate raised by two sources, of which one answers this time and does not
// raise it. The obvious rule ("some source answered and did not see it") cools it.
// The narrow rule does not, because the source that is missing is the one that
// might still have been raising it — and cooling starts a clock that ends in
// expiry, which discards first_seen_at. When the evidence is partial the
// conservative answer is to keep the candidate, because the cost of keeping a
// stale candidate is a row on a screen and the cost of dropping a live one is the
// entire claim this package makes about it.
func TestOneSurvivingSupporterIsNotEnoughToCool(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	ranking := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}}
	popular := &fakeSource{id: SourceWTSPopular, rows: []Row{row("005930", 4, 30)}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{ranking, popular},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	// Next scan: WTS answers and no longer lists it, the official ranking is gone.
	ranking.err = ErrRateLimited
	popular.rows = []Row{row("000660", 1, 30)}
	at := t0.Add(15 * time.Second)
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: at, Sources: []Source{ranking, popular},
	}); err != nil {
		t.Fatalf("Collect #2: %v", err)
	}

	got, found, err := s.Candidate(ctx, MarketKR, "005930", at)
	if err != nil || !found {
		t.Fatalf("Candidate: %v found=%v", err, found)
	}
	if got.State != StateActive {
		t.Errorf("state = %v, want %v — one of two supporters was rate limited, so "+
			"nothing established that this symbol left", got.State, StateActive)
	}
}

// --- what the section-2 review found ---------------------------------------------

// TestTwoSourcesCannotShareAnID is the section-2 review's P0, as a guard.
//
// `responded` and `Budgets` are keyed by source id and the cooling rule is "every
// source that raised this candidate answered". With two sources under one id, a
// source that answers vouches for one that did not — so a 429 on one official
// ranking cooled candidates that only the other ranking had ever raised, and the
// cooling clock then expired them. The panel is checked before anything is read
// because the failure is invisible once it is running.
func TestTwoSourcesCannotShareAnID(t *testing.T) {
	s, _ := openStore(t)
	_, err := Collect(context.Background(), s, CollectOptions{
		Market: MarketKR, At: t0,
		Sources: []Source{
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}},
			&fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("000660", 1, 100)}},
		},
	})
	if err == nil {
		t.Fatal("a panel with two sources under one id was accepted")
	}
}

// TestAnEmptyReadingIsNotEvidenceOfAbsence.
//
// A ranking that returns zero rows — out of hours, a server blip, a body whose
// symbols were all blank — has answered about nothing. Counting that as coverage
// cools the entire watchlist on a single 200, which reaches the same destruction
// as the 429 case with nothing having failed at all.
func TestAnEmptyReadingIsNotEvidenceOfAbsence(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{src},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	src.rows = nil
	at := t0.Add(15 * time.Second)
	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: at, Sources: []Source{src},
	})
	if err != nil {
		t.Fatalf("Collect #2: %v", err)
	}
	if result.Cooled != 0 {
		t.Errorf("an empty 200 cooled %d candidates", result.Cooled)
	}
	if !result.Degraded {
		t.Error("a source that answered with no rows is not reported as a degradation")
	}

	got, found, err := s.Candidate(ctx, MarketKR, "005930", at)
	if err != nil || !found {
		t.Fatalf("Candidate: %v found=%v", err, found)
	}
	if got.State != StateActive {
		t.Errorf("state = %v, want %v", got.State, StateActive)
	}
}

// TestOneRejectedSymbolDoesNotAbortTheMarket.
//
// Promote refuses an instant older than the candidate's last reading, which a
// clock step backwards produces. Aborting there left the remaining symbols
// un-promoted and skipped cooling — and because the loop is sorted, the same
// symbol blocked every subsequent scan for as long as the clock was behind, so a
// one-second NTP correction expired the whole watchlist.
func TestOneRejectedSymbolDoesNotAbortTheMarket(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	src := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{
		row("000660", 1, 3), row("005930", 2, 3), row("035720", 3, 3),
	}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0.Add(time.Second), Sources: []Source{src},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	// The clock steps back one second. 000660 sorts first and will be refused.
	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{src},
	})
	if err != nil {
		t.Fatalf("Collect #2: %v", err)
	}
	if len(result.Rejected) != 3 {
		t.Errorf("rejected = %v, want all three refused and reported", result.Rejected)
	}
	if result.Rejected[0].Reason == "" {
		t.Error("a rejected symbol carries no reason")
	}
	// The pass completed: the counters describe what happened rather than
	// stopping at the first refusal.
	if result.Observations != 3 {
		t.Errorf("observations = %d, want 3", result.Observations)
	}
}

// TestASupporterThatLeftThePanelDoesNotBlockCoolingForever.
//
// Candidate.Sources accumulates over a candidate's life, so a source that raised
// it once and has since been removed from the configuration — an operator
// dropping WTS, which the design treats as routine — would never appear in the
// responded set again. Without the panel guard those candidates become
// permanently un-coolable and can only leave through the staleness fallback,
// which is reserved for the scan that died.
func TestASupporterThatLeftThePanelDoesNotBlockCoolingForever(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	ranking := &fakeSource{id: SourceOfficialTradingValue, rows: []Row{row("005930", 1, 100)}}
	popular := &fakeSource{id: SourceWTSPopular, rows: []Row{row("005930", 4, 30)}}
	if _, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: t0, Sources: []Source{ranking, popular},
	}); err != nil {
		t.Fatalf("Collect: %v", err)
	}

	// WTS is dropped from the configuration entirely, and the symbol leaves the
	// official ranking.
	ranking.rows = []Row{row("000660", 1, 100)}
	at := t0.Add(15 * time.Second)
	result, err := Collect(ctx, s, CollectOptions{
		Market: MarketKR, At: at, Sources: []Source{ranking},
	})
	if err != nil {
		t.Fatalf("Collect #2: %v", err)
	}
	if result.Cooled != 1 {
		t.Fatalf("cooled = %d, want 1 — a source that no longer exists is not missing evidence",
			result.Cooled)
	}
}

// TestTheScheduleIsSafeToUseFromTwoGoroutines.
//
// YieldToEngine exists because something outside the scan loop watches whether
// the engine is running, so the setter and the readers are different goroutines
// by construction. An unsynchronised map here is not a subtle race — it is Go's
// unrecoverable "concurrent map read and map write", which in a process that also
// runs the engine would take fill detection down with it.
func TestTheScheduleIsSafeToUseFromTwoGoroutines(t *testing.T) {
	sched := NewSchedule(map[SourceID]Interval{
		SourceOfficialTradingValue: {Every: 15 * time.Second, Floor: 5 * time.Second},
		SourceWTSPopular:           {Every: 5 * time.Second, Floor: 3 * time.Second},
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 2000; i++ {
			sched.YieldToEngine(i%2 == 0)
			sched.Ran(SourceWTSPopular, t0.Add(time.Duration(i)*time.Second))
		}
	}()
	for i := 0; i < 2000; i++ {
		_ = sched.Due(SourceWTSPopular, t0.Add(time.Duration(i)*time.Second))
		_ = sched.Every(SourceOfficialTradingValue)
	}
	<-done
}

// TestAnUnconfiguredSourceIsNotUnlimited.
//
// A source absent from the interval map used to answer "due" on every tick with
// no floor and no engine yield — uncapped polling of an account budget shared
// with the engine, produced by a config typo or by a panel gaining a source
// before its interval was written down. Section 3 adds candles at five calls a
// second, which is both the most expensive source and the likeliest to arrive
// that way.
func TestAnUnconfiguredSourceIsNotUnlimited(t *testing.T) {
	sched := NewSchedule(map[SourceID]Interval{
		SourceWTSPopular: {Every: 5 * time.Second, Floor: 3 * time.Second},
	})
	sched.Ran(SourceOfficialCandles, t0)

	if got := sched.Every(SourceOfficialCandles); got <= 0 {
		t.Fatalf("an unconfigured source has an interval of %v — it is due on every tick", got)
	}
	if sched.Due(SourceOfficialCandles, t0.Add(time.Second)) {
		t.Error("an unconfigured source is due one second after it ran")
	}
	quiet := sched.Every(SourceOfficialCandles)
	sched.YieldToEngine(true)
	if sched.Every(SourceOfficialCandles) <= quiet {
		t.Error("an unconfigured source does not yield to the engine either")
	}
}

// TestARequestIDContainingTheDigits429IsNotARateLimit.
//
// The 429 count is the measurement of an undocumented limit, and section 5 backs
// off on it. An unanchored substring search matched any error text containing
// those three digits — a Korean stock code, a trace id, a request id echoed
// inside an API error body, all of which reach err.Error() because the official
// client embeds the whole response.
func TestARequestIDContainingTheDigits429IsNotARateLimit(t *testing.T) {
	for _, text := range []string{
		"wts popularity ranking: symbol 429000 not listed",
		"MARKET_TRADING_AMOUNT ranking: request 8f429ab timed out",
		`official: api error 400: {"traceId":"a429b"}`,
	} {
		if isRateLimited(errors.New(text)) {
			t.Errorf("%q was counted as a rate limit", text)
		}
	}
	for _, text := range []string{
		"official: status 429",
		"http 429 from the ranking endpoint",
		"429 too many requests",
		"rate limit exceeded",
	} {
		if !isRateLimited(errors.New(text)) {
			t.Errorf("%q was not counted as a rate limit", text)
		}
	}
	if !isRateLimited(ErrRateLimited) {
		t.Error("the sentinel itself is not counted")
	}
}
