package candidate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
)

// t0 is the instant every test in this package counts from. The clock is
// injected everywhere (D5), so no test ever waits for one.
var t0 = time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC)

func openStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "candidates.db")
	s, err := Open(context.Background(), Options{
		Path:     path,
		FSProber: FixedFSProber(FSInfo{Name: "ext4"}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, path
}

func obs(sym string, at time.Time, source SourceID, r Reported) Observation {
	return Observation{
		Market: "KR", Symbol: sym, Source: source, ObservedAt: at, Reported: r,
	}
}

// --- task 1.1: the source's word and ours are different columns ----------------

// TestAReportedValueIsNeverConfusedWithADerivedOne is D6 as a type.
//
// The screen and any later analysis has to be able to say which number came off
// the wire and which one we computed, because the remedy differs: a wrong
// reported value means the source contract moved, a wrong derived value means our
// arithmetic is broken. A single flat struct cannot answer that question and the
// answer is not recoverable afterwards.
func TestAReportedValueIsNeverConfusedWithADerivedOne(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	in := obs("005930", t0, SourceOfficialTradingValue, Reported{
		Rank: 3, RankTotal: 100, TradingValue: "1000000", TradingVolume: "500", Price: "70000",
	})
	if err := s.RecordObservations(ctx, []Observation{in}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}

	got, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("read back %d observations, want 1", len(got))
	}
	if got[0].Reported.TradingValue != "1000000" {
		t.Errorf("Reported.TradingValue = %q, want the source's own value",
			got[0].Reported.TradingValue)
	}
	// DayHigh is the one the ranking list never carries (D10/D13). Absent must
	// read back as absent, not as zero: a zero day high would make every
	// near_high comparison pass.
	if got[0].Reported.DayHigh != "" {
		t.Errorf("Reported.DayHigh = %q, want empty — the ranking list carries none",
			got[0].Reported.DayHigh)
	}
}

// TestTheSourceOfEveryObservationSurvivesTheRoundTrip pins D4/D14's raw material:
// a candidate's `sources` list can only be honest if each row remembers who
// raised it, including whether the value arrived through the hybrid fallback.
func TestTheSourceOfEveryObservationSurvivesTheRoundTrip(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 1, RankTotal: 50}),
		{
			Market: "KR", Symbol: "005930", Source: SourceOfficialTradingValue,
			ObservedAt: t0.Add(15 * time.Second),
			Reported:   Reported{Rank: 1, RankTotal: 50},
			// D14: the same source id, reached a different way. Recording this
			// as an ordinary official response is what makes over-polling
			// invisible.
			ViaFallback: true,
		},
	})
	if err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}

	got, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("read back %d observations, want 2", len(got))
	}
	if got[0].ViaFallback {
		t.Error("the first observation is marked as a fallback; it was not")
	}
	if !got[1].ViaFallback {
		t.Error("the fallback observation reads back as a direct official response — " +
			"D14: 자동 우회는 성공이 아니라 강등이다")
	}
}

// --- task 1.2: two scans, two rows, across a restart --------------------------

// TestASecondScanDoesNotOverwriteTheFirst is the spec scenario "두 번째 스캔이 첫
// 스캔을 덮어쓰지 않는다". Without it there is no time axis at all and the whole
// change is one more point-in-time query.
func TestASecondScanDoesNotOverwriteTheFirst(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	for i, at := range []time.Time{t0, t0.Add(15 * time.Second), t0.Add(30 * time.Second)} {
		o := obs("005930", at, SourceOfficialTradingValue, Reported{
			Rank: 5 - i, RankTotal: 100, TradingValue: "1000",
		})
		if err := s.RecordObservations(ctx, []Observation{o}); err != nil {
			t.Fatalf("RecordObservations #%d: %v", i, err)
		}
	}

	got, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("read back %d observations, want 3 — a scan overwrote an earlier one", len(got))
	}
	for i := 1; i < len(got); i++ {
		if !got[i].ObservedAt.After(got[i-1].ObservedAt) {
			t.Fatalf("observations came back out of order: %v then %v",
				got[i-1].ObservedAt, got[i].ObservedAt)
		}
	}
}

// TestTheHistoryOutlivesTheProcess is the spec scenario "재시작 후에도 이력이
// 남는다", and the proposal's whole argument for persistence: an in-memory
// first_seen_at makes every symbol newly discovered at every restart, which turns
// an early-discovery system into a late one that always answers "not late".
func TestTheHistoryOutlivesTheProcess(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "candidates.db")
	prober := FixedFSProber(FSInfo{Name: "ext4"})

	first, err := Open(ctx, Options{Path: path, FSProber: prober})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := first.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 1, RankTotal: 100}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := first.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	second, err := Open(ctx, Options{Path: path, FSProber: prober})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() { _ = second.Close() }()

	got, err := second.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations after restart: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("the new process read %d observations, want 1 — the history did not survive", len(got))
	}
	cands, err := second.Candidates(ctx, t0.Add(time.Minute))
	if err != nil {
		t.Fatalf("Candidates after restart: %v", err)
	}
	if len(cands) != 1 || !cands[0].FirstSeenAt.Equal(t0) {
		t.Fatalf("first_seen_at after restart = %v, want %v", cands, t0)
	}
}

// TestTheStoreRefusesAFilesystemThatCannotPromiseAWrite carries the ledger's own
// rule across (D2): the same partial-write exposure applies here, so the same
// allowlist does. It refuses rather than degrading, because a store that quietly
// loses rows loses exactly the first_seen_at this change is about.
func TestTheStoreRefusesAFilesystemThatCannotPromiseAWrite(t *testing.T) {
	dir := t.TempDir()
	_, err := Open(context.Background(), Options{
		Path:     filepath.Join(dir, "candidates.db"),
		FSProber: FixedFSProber(FSInfo{Name: "fuse/fuseblk", Magic: MagicFUSE}),
	})
	if !errors.Is(err, ErrFilesystemNotAllowed) {
		t.Fatalf("Open on fuseblk = %v, want ErrFilesystemNotAllowed", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "candidates.db")); statErr == nil {
		t.Error("a refused mount was written to anyway")
	}
}

// --- task 1.3: the discovery store does not hold the engine's ledger lock -------

// TestAScanningStoreDoesNotBlockTheEngineFromStarting is D2 stated as the thing
// an operator would actually notice. A discovery scan holding the ledger lock
// would stop `tossctl engine run` from coming up — an observation feature
// blocking an execution feature, which is the wrong direction.
func TestAScanningStoreDoesNotBlockTheEngineFromStarting(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	s, err := Open(ctx, Options{
		Path:     filepath.Join(dir, "candidates.db"),
		FSProber: FixedFSProber(FSInfo{Name: "ext4"}),
	})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = s.Close() }()

	// A scan in progress: rows being written, the handle open.
	for i := 0; i < 20; i++ {
		o := obs("00593"+string(rune('0'+i%10)), t0.Add(time.Duration(i)*time.Second),
			SourceOfficialTradingValue, Reported{Rank: i + 1, RankTotal: 100})
		if err := s.RecordObservations(ctx, []Observation{o}); err != nil {
			t.Fatalf("RecordObservations: %v", err)
		}
	}

	lock, err := enginelock.Acquire(dir)
	if err != nil {
		t.Fatalf("the engine could not take its lock while a scan was running: %v", err)
	}
	lock.Release()
}

// TestTheStoreIsItsOwnFile pins the other half of D2: same directory, different
// file. Sharing the journal's file would share its lock however carefully the
// code avoided the lock's name.
func TestTheStoreIsItsOwnFile(t *testing.T) {
	_, path := openStore(t)
	if filepath.Base(path) == "journal.db" {
		t.Fatal("the candidate store is the journal file")
	}
	if DBFileName == "journal.db" {
		t.Fatalf("DBFileName = %q; the discovery store must not be the ledger", DBFileName)
	}
}

// --- task 3.4c: the store can be asked for one source's rows --------------------

// TestOneSourcesReadingsCanBeReadWithoutTheOthers is the read D9's last section
// requires the store to have.
//
// Observations returns every source's rows for a symbol interleaved in time
// order, and that is the right answer for "show me everything we saw". It is the
// wrong input for a rate: TradingValue is a different cumulative per source, so a
// difference taken across the interleaving measures the gap between two sources'
// conventions. The metrics are keyed by (market, symbol, source), so the store has
// to be able to answer that key — and the mixed read has to stay available and
// stay obviously mixed, rather than being quietly narrowed.
func TestOneSourcesReadingsCanBeReadWithoutTheOthers(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{
			Rank: 1, RankTotal: 100, TradingValue: "10000000000"}),
		obs("005930", t0.Add(15*time.Second), SourceWTSPopular, Reported{
			Rank: 4, RankTotal: 30, TradingValue: "9500000000"}),
		obs("005930", t0.Add(30*time.Second), SourceOfficialTradingValue, Reported{
			Rank: 1, RankTotal: 100, TradingValue: "10600000000"}),
		obs("005930", t0.Add(45*time.Second), SourceWTSPopular, Reported{
			Rank: 3, RankTotal: 30, TradingValue: "9800000000"}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}

	all, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("the mixed read returned %d rows, want 4", len(all))
	}

	only, err := s.SourceObservations(ctx, "KR", "005930", SourceOfficialTradingValue, time.Time{})
	if err != nil {
		t.Fatalf("SourceObservations: %v", err)
	}
	if len(only) != 2 {
		t.Fatalf("the official series returned %d rows, want 2 — it carries another source's "+
			"readings, and differencing them would measure the difference between the two", len(only))
	}
	for i, o := range only {
		if o.Source != SourceOfficialTradingValue {
			t.Errorf("row %d is from %s", i, o.Source)
		}
		if i > 0 && o.ObservedAt.Before(only[i-1].ObservedAt) {
			t.Errorf("rows came back %v then %v; the documented order is oldest first",
				only[i-1].ObservedAt, o.ObservedAt)
		}
	}

	// `since` still narrows, so a metric can ask for the last two minutes of one
	// source rather than reading two trading days of rows to use three of them.
	recent, err := s.SourceObservations(ctx, "KR", "005930",
		SourceOfficialTradingValue, t0.Add(30*time.Second))
	if err != nil {
		t.Fatalf("SourceObservations(since): %v", err)
	}
	if len(recent) != 1 {
		t.Fatalf("the windowed read returned %d rows, want 1", len(recent))
	}
}

// --- task 1.4: lifecycle, and the one rule that makes the veto un-bypassable ---

// TestAReEntryWithinTheCoolingTTLKeepsTheOriginalFirstSeenAt is the spec scenario
// "냉각 중 재진입은 최초 시각을 유지한다", and D1 calls it the single path by which
// the chase veto could be bypassed: a symbol that has already doubled drops out of
// the ranking list for one scan, comes back, and is "just discovered".
func TestAReEntryWithinTheCoolingTTLKeepsTheOriginalFirstSeenAt(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if err := s.Cool(ctx, "KR", "005930", t0.Add(time.Minute)); err != nil {
		t.Fatalf("Cool: %v", err)
	}

	back, err := s.Promote(ctx, "KR", "005930", t0.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("re-Promote: %v", err)
	}
	if !back.FirstSeenAt.Equal(t0) {
		t.Errorf("first_seen_at after a re-entry = %v, want the original %v — "+
			"a symbol that left the list for one scan became newly discovered",
			back.FirstSeenAt, t0)
	}
	if back.State != StateActive {
		t.Errorf("state after a re-entry = %v, want %v", back.State, StateActive)
	}
}

// TestAReEntryAfterExpiryIsANewCandidate is the other half of the spec's pair.
// Keeping first_seen_at forever would be the opposite defect: a symbol observed
// once in the morning would still count as "seen early" at the close.
func TestAReEntryAfterExpiryIsANewCandidate(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	cooled := t0.Add(time.Minute)
	if err := s.Cool(ctx, "KR", "005930", cooled); err != nil {
		t.Fatalf("Cool: %v", err)
	}

	after := cooled.Add(DefaultCoolingTTL).Add(time.Second)
	back, err := s.Promote(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("re-Promote after expiry: %v", err)
	}
	if !back.FirstSeenAt.Equal(after) {
		t.Errorf("first_seen_at after expiry = %v, want the new pass %v", back.FirstSeenAt, after)
	}
}

// TestExpiryIsReadFromTheClockAndNotFromASweeper keeps the lifecycle honest when
// nothing is running: a candidate that cooled at lunch must read as expired in the
// afternoon whether or not any scan happened in between. A state that only
// advances when a sweeper runs is a state that lies on a restarted process.
func TestExpiryIsReadFromTheClockAndNotFromASweeper(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	cooled := t0.Add(time.Minute)
	if err := s.Cool(ctx, "KR", "005930", cooled); err != nil {
		t.Fatalf("Cool: %v", err)
	}

	states := map[time.Time]State{
		cooled.Add(time.Second):                     StateCooled,
		cooled.Add(DefaultCoolingTTL - time.Second): StateCooled,
		cooled.Add(DefaultCoolingTTL):               StateExpired,
		cooled.Add(2 * DefaultCoolingTTL):           StateExpired,
	}
	for at, want := range states {
		got, err := s.Candidates(ctx, at)
		if err != nil {
			t.Fatalf("Candidates(%v): %v", at, err)
		}
		if len(got) != 1 {
			t.Fatalf("Candidates(%v) returned %d, want 1", at, len(got))
		}
		if got[0].State != want {
			t.Errorf("at %v the state is %v, want %v", at.Sub(cooled), got[0].State, want)
		}
	}
}

// TestTheStoredInstantOrdersLikeTime is the section-1 review's P0, pinned.
//
// observed_at is TEXT, so every range and ORDER BY in the store is a byte
// comparison. RFC3339Nano — which is what stamp used — drops trailing zeros from
// the fraction, and 'Z' (0x5A) sorts after '.' (0x2E), so "09:00:00Z" compares
// GREATER than "09:00:00.5Z" even though it is half a second earlier.
//
// Nothing failed while that was true. PruneObservations deleted rows newer than
// its cutoff, Observations(since) dropped rows from inside the window it was
// asked for, and the rows came back out of order — and every test in this file
// used whole-second instants, which is precisely the branch where all the
// fractions are identical and none of it can show.
func TestTheStoredInstantOrdersLikeTime(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	// Deliberately mixed: whole seconds beside fractions, written out of order.
	instants := []time.Time{
		t0.Add(500 * time.Millisecond),
		t0.Add(-time.Second),
		t0,
		t0.Add(900 * time.Millisecond),
		t0.Add(2 * time.Second),
	}
	for _, at := range instants {
		if err := s.RecordObservations(ctx, []Observation{
			obs("005930", at, SourceOfficialTradingValue, Reported{Rank: 1, RankTotal: 100}),
		}); err != nil {
			t.Fatalf("RecordObservations(%v): %v", at, err)
		}
	}

	// The window asked for is [t0, ∞). Three rows are in it: t0, t0+0.5, t0+0.9,
	// t0+2 — four. The one before t0 is not.
	got, err := s.Observations(ctx, "KR", "005930", t0)
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(got) != 4 {
		var have []string
		for _, o := range got {
			have = append(have, o.ObservedAt.Format(time.RFC3339Nano))
		}
		t.Fatalf("Observations(since=%v) returned %d rows %v, want 4 — "+
			"the stored instant is not a usable range key",
			t0.Format(time.RFC3339Nano), len(got), have)
	}
	for i := 1; i < len(got); i++ {
		if got[i].ObservedAt.Before(got[i-1].ObservedAt) {
			t.Fatalf("rows came back %v then %v; the documented order is oldest first",
				got[i-1].ObservedAt, got[i].ObservedAt)
		}
	}

	// And the prune must take exactly what is older than its cutoff: the one row
	// before t0, and nothing that carries a fraction inside the same second.
	pruned, err := s.PruneObservations(ctx, t0)
	if err != nil {
		t.Fatalf("PruneObservations: %v", err)
	}
	if pruned != 1 {
		t.Fatalf("PruneObservations(before=%v) deleted %d rows, want 1 — "+
			"retention is eating rows it was told to keep", t0.Format(time.RFC3339Nano), pruned)
	}
}

// --- the two ways the lifecycle can run backwards -------------------------------

// TestAnInstantFromThePastCannotExpireALiveCandidate is the second D1 bypass, the
// one that arrives through the back door.
//
// Cool used to accept any instant. A cooling instant from the past starts the TTL
// in the past, so a candidate that has been running — and doubling — all day reads
// as expired, and the next promotion mints a fresh first_seen_at for it. seen_late
// then answers "no" forever. An NTP step, a replayed reading or a source-supplied
// timestamp is all it takes, and nothing about it looks like a failure.
func TestAnInstantFromThePastCannotExpireALiveCandidate(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	// Six hours of the symbol actually staying on the list: a scan every five
	// minutes, each one re-promoting it. This is what "live all day" is — two
	// promotions six hours apart is a six-hour gap, and the staleness rule is
	// right to expire that.
	last := t0
	for at := t0; !at.After(t0.Add(6 * time.Hour)); at = at.Add(5 * time.Minute) {
		if _, err := s.Promote(ctx, "KR", "005930", at); err != nil {
			t.Fatalf("Promote at %v: %v", at, err)
		}
		last = at
	}

	if err := s.Cool(ctx, "KR", "005930", t0.Add(5*time.Minute)); err == nil {
		t.Error("Cool accepted an instant from five minutes after the open on a candidate " +
			"last seen six hours later; that expires a live candidate")
	}

	got, _, err := s.Candidate(ctx, "KR", "005930", last.Add(time.Minute))
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if got.State != StateActive {
		t.Errorf("state after the refused Cool = %v, want %v", got.State, StateActive)
	}
	if !got.FirstSeenAt.Equal(t0) {
		t.Errorf("first_seen_at = %v, want %v", got.FirstSeenAt, t0)
	}
}

// TestAPromotionCannotRunBackwards is the same rule on the other write.
func TestAPromotionCannotRunBackwards(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0.Add(time.Hour)); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if _, err := s.Promote(ctx, "KR", "005930", t0); err == nil {
		t.Error("Promote accepted an instant an hour older than the candidate's last reading")
	}
}

// TestACandidateNobodyCooledDoesNotStayActiveForever is the never-cooled gap.
//
// Explicit cooling needs a writer, and the writer is the scan process. Kill it —
// or let the symbol simply stop appearing on any list — and no Cool is ever
// written. Deriving the state from cooled_at alone left such a candidate active
// indefinitely, still holding a first_seen_at from a session that ended days ago,
// so seen_late would answer "we have known about this since Monday" about a symbol
// nobody had looked at since.
func TestACandidateNobodyCooledDoesNotStayActiveForever(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	for _, tc := range []struct {
		after time.Duration
		want  State
		why   string
	}{
		{time.Minute, StateActive, "a live scan"},
		{5 * time.Minute, StateActive, "inside the longest planned backoff (300s)"},
		{DefaultStalenessTTL - time.Second, StateActive, "still inside the staleness window"},
		{DefaultStalenessTTL, StateCooled, "nobody refreshed it; it cools implicitly"},
		{DefaultStalenessTTL + DefaultCoolingTTL, StateExpired, "and then expires"},
		{72 * time.Hour, StateExpired, "three days later it is certainly not active"},
	} {
		got, found, err := s.Candidate(ctx, "KR", "005930", t0.Add(tc.after))
		if err != nil || !found {
			t.Fatalf("Candidate(+%v): %v found=%v", tc.after, err, found)
		}
		if got.State != tc.want {
			t.Errorf("+%v: state = %v, want %v (%s)", tc.after, got.State, tc.want, tc.why)
		}
	}

	// And the promotion after that gap is a new candidate, not a three-day-old one.
	back, err := s.Promote(ctx, "KR", "005930", t0.Add(72*time.Hour))
	if err != nil {
		t.Fatalf("Promote after the gap: %v", err)
	}
	if !back.FirstSeenAt.Equal(t0.Add(72 * time.Hour)) {
		t.Errorf("first_seen_at after a three-day gap = %v, want the new pass; "+
			"seen_late would report this symbol as known since %v",
			back.FirstSeenAt, t0)
	}
}

// --- D4: provenance belongs to the candidate it describes ------------------------

// TestANewCandidateDoesNotInheritTheDeadOnesSources.
//
// Promote used to update three columns and leave sources, completeness and
// degraded alone, so a candidate whose life began one second ago reported the
// source mix of a scan that ended half an hour earlier. That is D4's completeness
// accounting describing the wrong scan — the specific failure it exists to prevent.
func TestANewCandidateDoesNotInheritTheDeadOnesSources(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if err := s.NoteSources(ctx, "KR", "005930",
		[]SourceID{SourceWTSPopular, SourceOfficialTradingValue}, 5, 2, true); err != nil {
		t.Fatalf("NoteSources: %v", err)
	}
	cooled := t0.Add(time.Minute)
	if err := s.Cool(ctx, "KR", "005930", cooled); err != nil {
		t.Fatalf("Cool: %v", err)
	}

	after := cooled.Add(DefaultCoolingTTL).Add(time.Second)
	fresh, err := s.Promote(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("Promote after expiry: %v", err)
	}
	if len(fresh.Sources) != 0 {
		t.Errorf("a new candidate carries %v from the expired one", fresh.Sources)
	}
	if fresh.SourcesAttempted != 0 || fresh.SourcesResponded != 0 || fresh.Degraded {
		t.Errorf("a new candidate reports %d/%d degraded=%v from a scan that ended 30 minutes ago",
			fresh.SourcesResponded, fresh.SourcesAttempted, fresh.Degraded)
	}

	stored, _, err := s.Candidate(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if len(stored.Sources) != 0 || stored.Degraded {
		t.Errorf("the stored row still carries %v degraded=%v", stored.Sources, stored.Degraded)
	}
}

// TestRecordingSourcesOneAtATimeKeepsThemAll.
//
// Candidate.Sources is documented as the sources that have supported this
// candidate — D1's "one symbol raised by two sources is one candidate two sources
// supported". A caller recording per source must therefore accumulate, not
// overwrite the previous one.
func TestRecordingSourcesOneAtATimeKeepsThemAll(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if err := s.NoteSources(ctx, "KR", "005930",
		[]SourceID{SourceOfficialTradingValue}, 5, 5, false); err != nil {
		t.Fatalf("NoteSources #1: %v", err)
	}
	if err := s.NoteSources(ctx, "KR", "005930",
		[]SourceID{SourceWTSPopular}, 5, 5, false); err != nil {
		t.Fatalf("NoteSources #2: %v", err)
	}

	got, _, err := s.Candidate(ctx, "KR", "005930", t0.Add(time.Minute))
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if len(got.Sources) != 2 {
		t.Fatalf("sources = %v, want both — the second write erased the first", got.Sources)
	}
}

// TestRecordingSourcesAgainstNothingIsAnError.
//
// D4's claim is that a reduced source panel is never hidden. A provenance write
// that lands nowhere and returns nil hides it more completely than a failed scan
// would: the caller believes it recorded 2-of-5, and there is no row saying so.
func TestRecordingSourcesAgainstNothingIsAnError(t *testing.T) {
	s, _ := openStore(t)
	err := s.NoteSources(context.Background(), "KR", "999999",
		[]SourceID{SourceWTSPopular}, 5, 2, true)
	if !errors.Is(err, ErrNoCandidate) {
		t.Fatalf("NoteSources against an unknown candidate = %v, want ErrNoCandidate", err)
	}
}

// --- D8: a rank with no list length cannot be normalised --------------------------

func TestARankWithoutItsListLengthIsRefused(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	// rank/rankTotal is the percentile input (D8). A zero total makes it +Inf, and
	// +Inf clears every threshold it is compared against — a maximal signal
	// manufactured out of a missing field.
	err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 5, RankTotal: 0}),
	})
	if err == nil {
		t.Error("an observation ranked 5 of an unstated list was accepted")
	}

	// A source that is not a ranking at all reports neither, and that is fine.
	if err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialPrices, Reported{Price: "70000"}),
	}); err != nil {
		t.Errorf("a non-ranking observation was refused: %v", err)
	}
}

// --- D11: pruning raw rows must not prune the claim they supported -------------

// TestPruningRawObservationsLeavesTheCandidateSummary is the spec scenario "원
// 관측을 강등해도 최초 발견 시각은 남는다". D11 predicts exactly how this breaks:
// somebody tidies up a store that grows by millions of rows a day and takes
// first_seen_at with it, which deletes the change's entire claim without failing
// anything.
func TestPruningRawObservationsLeavesTheCandidateSummary(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 1, RankTotal: 100}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	pruned, err := s.PruneObservations(ctx, t0.Add(48*time.Hour))
	if err != nil {
		t.Fatalf("PruneObservations: %v", err)
	}
	if pruned != 1 {
		t.Fatalf("pruned %d rows, want 1", pruned)
	}

	left, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("%d raw observations survived the prune", len(left))
	}

	cands, err := s.Candidates(ctx, t0.Add(time.Minute))
	if err != nil {
		t.Fatalf("Candidates: %v", err)
	}
	if len(cands) != 1 {
		t.Fatalf("the prune deleted the candidate summary; %d left, want 1", len(cands))
	}
	if !cands[0].FirstSeenAt.Equal(t0) {
		t.Errorf("first_seen_at after the prune = %v, want %v", cands[0].FirstSeenAt, t0)
	}
}

// --- D17: the expansion baseline is a column, not a query over prunable rows ----

// TestTheCandidateSummaryCarriesTheFirstPrice is D17's schema half.
//
// D11's retention table promised the candidate summary would keep the "극값" until
// the candidate expired. The candidates table had no such column — timestamps,
// sources, counters and degraded, and nothing else — so the baseline behind the
// `extended` veto lived only in `observations`, which PruneObservations empties
// every two days and SourceObservations narrows on request.
func TestTheCandidateSummaryCarriesTheFirstPrice(t *testing.T) {
	s, _ := openStore(t)
	cols := candidateColumns(t, s)
	for _, want := range []string{"first_price", "first_price_at", "first_price_source"} {
		if !cols[want] {
			t.Errorf("candidates has no %s column; the extended veto's baseline lives only in "+
				"the prunable observations table", want)
		}
	}
}

func candidateColumns(t *testing.T, s *Store) map[string]bool {
	t.Helper()
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT name FROM pragma_table_info('candidates')`)
	if err != nil {
		t.Fatalf("pragma_table_info: %v", err)
	}
	defer func() { _ = rows.Close() }()
	cols := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		cols[name] = true
	}
	return cols
}

// TestPruningTheRawRowsDoesNotRebaseTheExpansionBaseline is D17's behavioural
// half, and the §3 review's P0 measured end to end.
//
// Before the column existed:
//
//	before the prune  measured=true  first="10000"  gain=200.0%
//	after the prune   measured=true  first="20000"  gain=50.0%   why=""
//
// A candidate that had tripled reported a 50% expansion, with Measured true and
// no reason given, on exactly the candidates that had run longest — the ones
// `extended` exists for. D17 predicted the failure would degrade to unmeasured;
// it did not, and a wrong number is worse than a missing one because D10 stops
// the missing one from being read as a pass.
func TestPruningTheRawRowsDoesNotRebaseTheExpansionBaseline(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	// The candidate stays on the lists for the whole fifty hours. That is what
	// makes it the candidate D17 is about: it outlives the 48-hour raw retention
	// window while keeping its first_seen_at, so it is the one whose baseline used
	// to disappear. A gap instead of a scan here would expire it, and an expired
	// candidate is a different candidate with a new baseline by design — see
	// TestTheBaselineFollowsFirstSeenAtThroughCoolingAndExpiry.
	instants := []time.Time{t0, t0.Add(49 * time.Hour), t0.Add(50 * time.Hour)}
	prices := map[time.Time]string{
		instants[0]: "10000", instants[1]: "20000", instants[2]: "30000",
	}
	for at := t0; !at.After(instants[2]); at = at.Add(9 * time.Minute) {
		if _, err := s.Promote(ctx, "KR", "005930", at); err != nil {
			t.Fatalf("Promote at %v: %v", at, err)
		}
	}
	for _, at := range instants {
		if err := s.RecordObservations(ctx, []Observation{
			obs("005930", at, SourceOfficialPrices, Reported{Price: prices[at]}),
		}); err != nil {
			t.Fatalf("RecordObservations: %v", err)
		}
		// What a scan does on every priced reading: write it and let the store
		// decide whether it is the first one.
		if _, err := s.NoteFirstPrice(ctx, "KR", "005930", prices[at], at, SourceOfficialPrices); err != nil {
			t.Fatalf("NoteFirstPrice at %v: %v", at, err)
		}
	}
	if c, _, err := s.Candidate(ctx, "KR", "005930", instants[2]); err != nil || c.State != StateActive {
		t.Fatalf("the candidate is %v at +50h (%v); the scenario needs it alive", c.State, err)
	}

	baseline, found, err := s.Baseline(ctx, "KR", "005930")
	if err != nil || !found {
		t.Fatalf("Baseline: %v found=%v", err, found)
	}
	all, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	before := MeasureExpansion(baseline, all)
	if before.FirstPrice != "10000" || before.GainPct != "200" {
		t.Fatalf("before the prune: first=%q gain=%s%%, want 10000 and 200%%",
			before.FirstPrice, before.GainPct)
	}

	newest := instants[len(instants)-1]
	if _, err := s.PruneObservations(ctx, newest.Add(-DefaultRawRetention)); err != nil {
		t.Fatalf("PruneObservations: %v", err)
	}
	left, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations after the prune: %v", err)
	}
	if len(left) != 2 {
		t.Fatalf("%d raw rows survived the prune, want 2", len(left))
	}

	pruned, found, err := s.Baseline(ctx, "KR", "005930")
	if err != nil || !found {
		t.Fatalf("Baseline after the prune: %v found=%v", err, found)
	}
	after := MeasureExpansion(pruned, left)
	if after.FirstPrice != "10000" || !after.FirstAt.Equal(t0) {
		t.Errorf("baseline after the prune = %q at %v, want 10000 at %v — retention re-based it",
			after.FirstPrice, after.FirstAt, t0)
	}
	if after.GainPct != "200" {
		t.Errorf("gain after the prune = %s%%, want 200%% — a candidate that tripled reports %s%%",
			after.GainPct, after.GainPct)
	}
	if after.FirstSource != SourceOfficialPrices {
		t.Errorf("the baseline's source = %q, want %q — D17 keeps the instant and the source "+
			"beside the value so a reader can tell a 20-minute-old baseline from a 3-day-old one",
			after.FirstSource, SourceOfficialPrices)
	}
}

// TestTheBaselineIsSetOnceAndNotOverwritten.
//
// Every priced reading is written; only the first one lands. A store that took
// the latest instead would re-base the baseline on every scan, which is the same
// defect as the prune with a shorter fuse.
func TestTheBaselineIsSetOnceAndNotOverwritten(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()
	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	first, err := s.NoteFirstPrice(ctx, "KR", "005930", "10000", t0, SourceWTSPopular)
	if err != nil {
		t.Fatalf("NoteFirstPrice: %v", err)
	}
	if first.Price != "10000" || !first.At.Equal(t0) || first.Source != SourceWTSPopular {
		t.Fatalf("the first baseline came back as %+v", first)
	}

	again, err := s.NoteFirstPrice(ctx, "KR", "005930", "30000", t0.Add(time.Hour), SourceOfficialPrices)
	if err != nil {
		t.Fatalf("NoteFirstPrice #2: %v", err)
	}
	if again.Price != "10000" || !again.At.Equal(t0) || again.Source != SourceWTSPopular {
		t.Errorf("the second write moved the baseline to %+v; it is set once and never "+
			"overwritten while the candidate lives", again)
	}

	stored, _, err := s.Baseline(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if stored.Price != "10000" {
		t.Errorf("the stored baseline is %q, want 10000", stored.Price)
	}
}

// TestAnAbsentBaselineIsNotAZeroOne.
//
// The same rule every decimal in this package follows: NULL means no observation
// has carried a price yet, and it must not read as a price of zero — a baseline
// of zero is a denominator that makes every ratio infinite and every threshold
// pass.
func TestAnAbsentBaselineIsNotAZeroOne(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()
	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	got, found, err := s.Baseline(ctx, "KR", "005930")
	if err != nil || !found {
		t.Fatalf("Baseline: %v found=%v", err, found)
	}
	if got.Recorded() {
		t.Errorf("a candidate nobody has priced reports a baseline of %q", got.Price)
	}
	if got.Price != "" {
		t.Errorf("the absent baseline reads as %q", got.Price)
	}

	// And an unknown candidate is a different answer from a candidate with no
	// baseline. Both are unrecorded; only one is a lifecycle question.
	if _, found, err := s.Baseline(ctx, "KR", "999999"); err != nil || found {
		t.Errorf("Baseline of an unknown candidate: found=%v err=%v", found, err)
	}
	if _, err := s.NoteFirstPrice(ctx, "KR", "999999", "1000", t0, SourceOfficialPrices); !errors.Is(err, ErrNoCandidate) {
		t.Errorf("NoteFirstPrice against an unknown candidate = %v, want ErrNoCandidate", err)
	}
	// An empty price is not a baseline either.
	if _, err := s.NoteFirstPrice(ctx, "KR", "005930", "  ", t0, SourceOfficialPrices); err == nil {
		t.Error("an empty price was accepted as a baseline")
	}
}

// TestTheBaselineFollowsFirstSeenAtThroughCoolingAndExpiry is D17's lifecycle
// claim, and it is D1's rule applied to the second field of the same grade.
//
// Kept through cooling and re-entry: dropping it there would reopen the bypass D1
// closed, one field along. A symbol that has already doubled leaves the ranking
// list for one scan, comes back, and gets a fresh baseline — so `extended`
// answers "not extended" for it from then on, exactly as `seen_late` would have
// answered "not late" if first_seen_at moved.
//
// Reset on expiry: the next pass is a different candidate, and measuring its
// expansion from a dead candidate's price is measuring across two lives.
func TestTheBaselineFollowsFirstSeenAtThroughCoolingAndExpiry(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if _, err := s.NoteFirstPrice(ctx, "KR", "005930", "10000", t0, SourceOfficialPrices); err != nil {
		t.Fatalf("NoteFirstPrice: %v", err)
	}

	cooled := t0.Add(time.Minute)
	if err := s.Cool(ctx, "KR", "005930", cooled); err != nil {
		t.Fatalf("Cool: %v", err)
	}
	if got, _, err := s.Baseline(ctx, "KR", "005930"); err != nil || got.Price != "10000" {
		t.Fatalf("the baseline during cooling = %q (%v), want 10000", got.Price, err)
	}

	back, err := s.Promote(ctx, "KR", "005930", cooled.Add(time.Minute))
	if err != nil {
		t.Fatalf("re-Promote: %v", err)
	}
	if !back.FirstSeenAt.Equal(t0) {
		t.Fatalf("first_seen_at after a re-entry = %v, want %v", back.FirstSeenAt, t0)
	}
	kept, _, err := s.Baseline(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("Baseline after the re-entry: %v", err)
	}
	if kept.Price != "10000" || !kept.At.Equal(t0) || kept.Source != SourceOfficialPrices {
		t.Errorf("the baseline after a re-entry = %+v, want the original 10000 at %v — a symbol "+
			"that had already doubled would come back with a fresh baseline and never read "+
			"as extended again", kept, t0)
	}

	// Expiry. The staleness clock alone gets there: nobody re-promoted it.
	after := cooled.Add(time.Minute).Add(DefaultStalenessTTL).Add(DefaultCoolingTTL)
	state, _, err := s.Candidate(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if state.State != StateExpired {
		t.Fatalf("state at +%v = %v, want %v", after.Sub(t0), state.State, StateExpired)
	}
	fresh, err := s.Promote(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("Promote after expiry: %v", err)
	}
	if !fresh.FirstSeenAt.Equal(after) {
		t.Fatalf("first_seen_at after expiry = %v, want the new pass %v", fresh.FirstSeenAt, after)
	}
	reset, _, err := s.Baseline(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("Baseline after expiry: %v", err)
	}
	if reset.Recorded() {
		t.Errorf("a new candidate inherited the expired one's baseline (%+v); its expansion "+
			"would be measured across two lives", reset)
	}
	if !reset.At.IsZero() || reset.Source != "" {
		t.Errorf("the baseline's instant and source survived expiry: %+v", reset)
	}

	// And the new life can set its own.
	if _, err := s.NoteFirstPrice(ctx, "KR", "005930", "90000", after, SourceWTSPopular); err != nil {
		t.Fatalf("NoteFirstPrice after expiry: %v", err)
	}
	if got, _, err := s.Baseline(ctx, "KR", "005930"); err != nil || got.Price != "90000" {
		t.Errorf("the new life's baseline = %q (%v), want 90000", got.Price, err)
	}
}

// --- task 4.9: the first sighting's rank is a column too -------------------------

// TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry is the baseline's
// lifecycle claim applied to the field task 4.9 adds, and it is the same D1 rule
// on a third fact of the same grade.
//
// Kept through cooling and re-entry: dropping it there reopens D1's bypass one
// field along. A symbol that had climbed to 5th leaves the ranking list for one
// scan, comes back, and records 5th as where we first saw it — so seen_late
// answers "we caught it early" from then on.
//
// Reset on expiry: the next pass is a different candidate, and D20's whole finding
// is that a dead life's position answers for a live one in the unsafe direction.
func TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 148, 150, t0, SourceOfficialTradingValue); err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}

	cooled := t0.Add(time.Minute)
	if err := s.Cool(ctx, "KR", "005930", cooled); err != nil {
		t.Fatalf("Cool: %v", err)
	}
	if got, _, err := s.FirstRank(ctx, "KR", "005930"); err != nil || got.Rank != 148 {
		t.Fatalf("the first rank during cooling = %d (%v), want 148", got.Rank, err)
	}

	back, err := s.Promote(ctx, "KR", "005930", cooled.Add(time.Minute))
	if err != nil {
		t.Fatalf("re-Promote: %v", err)
	}
	if !back.FirstSeenAt.Equal(t0) {
		t.Fatalf("first_seen_at after a re-entry = %v, want %v", back.FirstSeenAt, t0)
	}
	// The re-entry offers its current — much higher — position, and the store
	// refuses it because one is already recorded.
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 5, 150,
		cooled.Add(time.Minute), SourceOfficialTradingValue); err != nil {
		t.Fatalf("NoteFirstRank on the re-entry: %v", err)
	}
	kept, _, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("FirstRank after the re-entry: %v", err)
	}
	if kept.Rank != 148 || kept.Total != 150 || !kept.At.Equal(t0) ||
		kept.Source != SourceOfficialTradingValue {
		t.Errorf("the first rank after a re-entry = %+v, want the original 148 of 150 at %v — a "+
			"symbol that had climbed to 5th would come back recording 5th as where we first "+
			"saw it, and seen_late would clear for it from then on", kept, t0)
	}

	// Expiry. The staleness clock alone gets there: nobody re-promoted it.
	after := cooled.Add(time.Minute).Add(DefaultStalenessTTL).Add(DefaultCoolingTTL)
	state, _, err := s.Candidate(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if state.State != StateExpired {
		t.Fatalf("state at +%v = %v, want %v", after.Sub(t0), state.State, StateExpired)
	}
	fresh, err := s.Promote(ctx, "KR", "005930", after)
	if err != nil {
		t.Fatalf("Promote after expiry: %v", err)
	}
	reset, _, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("FirstRank after expiry: %v", err)
	}
	if reset.Recorded() {
		t.Errorf("a new candidate inherited the expired one's first rank (%+v); D20's finding is "+
			"that this answers a live candidate's seen_late with a dead one's position", reset)
	}
	if !reset.At.IsZero() || reset.Source != "" {
		t.Errorf("the first rank's instant and source survived expiry: %+v", reset)
	}
	// The new life records its own, and it is the position it actually came back at.
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 5, 150, after, SourceOfficialGainers); err != nil {
		t.Fatalf("NoteFirstRank after expiry: %v", err)
	}
	again, _, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil || again.Rank != 5 {
		t.Fatalf("the new life's first rank = %d (%v), want 5", again.Rank, err)
	}
	sighting := MeasureFirstSighting(fresh, again, nil)
	if !sighting.Measured || sighting.PercentilePct != "96.666666666666" {
		t.Errorf("the new life's sighting = measured %v at %s%%, want 5 of 150",
			sighting.Measured, sighting.PercentilePct)
	}
}

// TestARankFromOutsideTheIdentityWindowIsNotStored is the write half of the
// backstop.
//
// The ordinary caller is a scan loop that offers a ranked reading on every tick,
// and the first one it offers is not necessarily the first sighting: a candidate
// raised by prices or candles carries no position until some ranking list picks it
// up, which may be hours later. Storing that would be D17's late baseline with no
// safe direction — so it is not stored, it is not an error, and seen_late stays
// unmeasured rather than answering from a later moment.
func TestARankFromOutsideTheIdentityWindowIsNotStored(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	late := t0.Add(DefaultStalenessTTL)
	got, err := s.NoteFirstRank(ctx, "KR", "005930", 4, 150, late, SourceOfficialTradingValue)
	if err != nil {
		t.Fatalf("NoteFirstRank outside the window = %v, want no error; a candidate first raised "+
			"by a source that carries no position is ordinary", err)
	}
	if got.Recorded() {
		t.Errorf("a reading %v after first_seen_at was stored as the first sighting: %+v",
			DefaultStalenessTTL, got)
	}
	stored, _, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil || stored.Recorded() {
		t.Errorf("the store kept %+v (%v), want nothing", stored, err)
	}
	// One second inside it is stored, so the boundary is where it says it is.
	inside := t0.Add(DefaultStalenessTTL - time.Second)
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 4, 150, inside, SourceOfficialTradingValue); err != nil {
		t.Fatalf("NoteFirstRank inside the window: %v", err)
	}
	if stored, _, _ := s.FirstRank(ctx, "KR", "005930"); stored.Rank != 4 {
		t.Errorf("a reading one second inside the window was not stored: %+v", stored)
	}
}

// TestARankOfZeroIsNotAFirstSighting. Absent is not zero, and a rank of 0 with a
// list length is not a position at the top of the list — percentileOf would make
// it exactly that.
func TestARankOfZeroIsNotAFirstSighting(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()
	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	for _, bad := range []struct{ rank, total int }{{0, 150}, {12, 0}, {-1, 150}, {0, 0}, {151, 150}} {
		if _, err := s.NoteFirstRank(ctx, "KR", "005930", bad.rank, bad.total, t0,
			SourceOfficialTradingValue); err == nil {
			t.Errorf("NoteFirstRank accepted %d of %d", bad.rank, bad.total)
		}
	}
	// And the other refusals a provenance write cannot make silently.
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 12, 150, time.Time{},
		SourceOfficialTradingValue); err == nil {
		t.Error("NoteFirstRank accepted a first sighting with no instant")
	}
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 12, 150, t0, "  "); err == nil {
		t.Error("NoteFirstRank accepted a first sighting with no source")
	}
	if _, err := s.NoteFirstRank(ctx, "KR", "999999", 12, 150, t0,
		SourceOfficialTradingValue); !errors.Is(err, ErrNoCandidate) {
		t.Errorf("NoteFirstRank against an unknown candidate = %v, want ErrNoCandidate", err)
	}
	if got, found, err := s.FirstRank(ctx, "KR", "999999"); err != nil || found || got.Recorded() {
		t.Errorf("FirstRank of an unknown candidate = %+v found=%v (%v)", got, found, err)
	}
}

// TestPruningRawObservationsLeavesTheFirstRankToo extends D11's promise to the
// column task 4.9 adds. It is the case the column exists for: the rows the rank
// came from are exactly what retention removes, and seen_late is a question about
// the longest-running candidates.
func TestPruningRawObservationsLeavesTheFirstRankToo(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialTradingValue, Reported{Rank: 148, RankTotal: 150, Price: "10000"}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	c, err := s.Promote(ctx, "KR", "005930", t0)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 148, 150, t0, SourceOfficialTradingValue); err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}
	if _, err := s.PruneObservations(ctx, t0.Add(DefaultRawRetention)); err != nil {
		t.Fatalf("PruneObservations: %v", err)
	}
	rows, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("%d raw rows survived the prune, want 0", len(rows))
	}
	first, _, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("FirstRank after the prune: %v", err)
	}
	if first.Rank != 148 || first.Total != 150 || !first.At.Equal(t0) {
		t.Errorf("the first rank after the prune = %+v, want 148 of 150 at %v", first, t0)
	}
	sighting := MeasureFirstSighting(c, first, rows)
	if !sighting.Measured {
		t.Errorf("seen_late is unmeasurable once the rows are gone (%q); the column exists so "+
			"that the longest-running candidates are the ones it can still answer for",
			sighting.Why)
	}
}

// TestPruningRawObservationsLeavesTheBaselineToo extends D11's promise to the
// column D17 added: PruneObservations reaches `observations` and nothing else, so
// the tidy-up that would have taken first_seen_at cannot take the first price
// either.
func TestPruningRawObservationsLeavesTheBaselineToo(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()

	if err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialPrices, Reported{Price: "10000"}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	if _, err := s.Promote(ctx, "KR", "005930", t0); err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if _, err := s.NoteFirstPrice(ctx, "KR", "005930", "10000", t0, SourceOfficialPrices); err != nil {
		t.Fatalf("NoteFirstPrice: %v", err)
	}
	if _, err := s.PruneObservations(ctx, t0.Add(48*time.Hour)); err != nil {
		t.Fatalf("PruneObservations: %v", err)
	}

	got, _, err := s.Baseline(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if got.Price != "10000" || !got.At.Equal(t0) {
		t.Errorf("the baseline after the prune = %+v, want 10000 at %v", got, t0)
	}
}

// --- the schema ladder ------------------------------------------------------------

// v1Schema is the schema exactly as SchemaVersion 1 wrote it. It is spelled out
// rather than derived, because a migration test that builds its "old" database
// from the current source cannot fail: it would migrate a v2 file and call it a
// v1 one.
const v1Schema = `
CREATE TABLE IF NOT EXISTS observations (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	market         TEXT    NOT NULL,
	symbol         TEXT    NOT NULL,
	source         TEXT    NOT NULL,
	observed_at    TEXT    NOT NULL,
	rank           INTEGER NOT NULL DEFAULT 0,
	rank_total     INTEGER NOT NULL DEFAULT 0,
	newly_listed   INTEGER NOT NULL DEFAULT 0,
	trading_value  TEXT,
	trading_volume TEXT,
	price          TEXT,
	day_high       TEXT,
	day_low        TEXT,
	via_fallback   INTEGER NOT NULL DEFAULT 0,
	CHECK (newly_listed IN (0,1)),
	CHECK (via_fallback IN (0,1)),
	CHECK (rank >= 0 AND rank_total >= 0)
) STRICT;
CREATE INDEX IF NOT EXISTS idx_observations_symbol_time
	ON observations(market, symbol, observed_at);
CREATE INDEX IF NOT EXISTS idx_observations_time ON observations(observed_at);
CREATE TABLE IF NOT EXISTS candidates (
	market             TEXT NOT NULL,
	symbol             TEXT NOT NULL,
	first_seen_at      TEXT NOT NULL,
	last_seen_at       TEXT NOT NULL,
	cooled_at          TEXT,
	sources            TEXT NOT NULL DEFAULT '',
	sources_attempted  INTEGER NOT NULL DEFAULT 0,
	sources_responded  INTEGER NOT NULL DEFAULT 0,
	degraded           INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (market, symbol),
	CHECK (degraded IN (0,1))
) STRICT;
CREATE TABLE IF NOT EXISTS store_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;
INSERT INTO store_meta (key, value) VALUES ('schema_version', '1');
`

// writeV1Store builds a schema-1 database with one candidate and its raw rows,
// the way a build from before D17 would have left one.
func writeV1Store(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", dsn(path, time.Second))
	if err != nil {
		t.Fatalf("opening a v1 store: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(v1Schema); err != nil {
		t.Fatalf("creating the v1 schema: %v", err)
	}
	_, err = db.Exec(`INSERT INTO candidates
		(market, symbol, first_seen_at, last_seen_at, sources, sources_attempted, sources_responded, degraded)
		VALUES ('KR','005930',?,?,'official_prices',5,5,0)`,
		stamp(t0), stamp(t0.Add(time.Minute)))
	if err != nil {
		t.Fatalf("seeding the v1 candidate: %v", err)
	}
	for i, price := range []string{"10000", "11000"} {
		_, err = db.Exec(`INSERT INTO observations
			(market, symbol, source, observed_at, rank, rank_total, newly_listed, price, via_fallback)
			VALUES ('KR','005930','official_prices',?,0,0,0,?,0)`,
			stamp(t0.Add(time.Duration(i)*time.Minute)), price)
		if err != nil {
			t.Fatalf("seeding a v1 observation: %v", err)
		}
	}
	// Two ranked rows for the same candidate: one that could be its first sighting
	// and one from half an hour later, so the v3 backfill has both to choose from.
	for _, r := range []struct {
		at   time.Time
		rank int
	}{{t0, 12}, {t0.Add(30 * time.Minute), 4}} {
		_, err = db.Exec(`INSERT INTO observations
			(market, symbol, source, observed_at, rank, rank_total, newly_listed, price, via_fallback)
			VALUES ('KR','005930','official_rankings_trading_value',?,?,150,0,'10000',0)`,
			stamp(r.at), r.rank)
		if err != nil {
			t.Fatalf("seeding a v1 ranked observation: %v", err)
		}
	}
	// And a second candidate whose only surviving ranked row is an hour after it
	// was first seen — the shape the backfill has to leave alone.
	_, err = db.Exec(`INSERT INTO candidates
		(market, symbol, first_seen_at, last_seen_at, sources, sources_attempted, sources_responded, degraded)
		VALUES ('KR','000660',?,?,'official_rankings_trading_value',5,5,0)`,
		stamp(t0), stamp(t0.Add(time.Hour)))
	if err != nil {
		t.Fatalf("seeding the second v1 candidate: %v", err)
	}
	_, err = db.Exec(`INSERT INTO observations
		(market, symbol, source, observed_at, rank, rank_total, newly_listed, price, via_fallback)
		VALUES ('KR','000660','official_rankings_trading_value',?,9,150,0,'20000',0)`,
		stamp(t0.Add(time.Hour)))
	if err != nil {
		t.Fatalf("seeding the second v1 candidate's observation: %v", err)
	}
}

// TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows.
//
// Additive and nullable, which is WORKFLOW §0.6's preference for a schema change:
// nothing already in the file is rewritten, so the step cannot lose a
// first_seen_at, and that is the field this package exists to protect.
func TestAStoreLeftByAnOlderBuildOpensMigratesAndKeepsItsRows(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "candidates.db")
	writeV1Store(t, path)

	s, err := Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatalf("Open on a v1 store: %v", err)
	}
	defer func() { _ = s.Close() }()

	cols := candidateColumns(t, s)
	for _, want := range []string{
		"first_price", "first_price_at", "first_price_source",
		"first_rank", "first_rank_total", "first_rank_at", "first_rank_source",
	} {
		if !cols[want] {
			t.Errorf("after the migration candidates still has no %s", want)
		}
	}

	var version string
	if err := s.db.QueryRowContext(ctx,
		`SELECT value FROM store_meta WHERE key = 'schema_version'`).Scan(&version); err != nil {
		t.Fatalf("reading the schema version: %v", err)
	}
	if version != fmt.Sprint(SchemaVersion) {
		t.Errorf("schema version after the migration = %q, want %d — a store two versions behind "+
			"climbs both rungs in order", version, SchemaVersion)
	}

	// The rows it had are the rows it has.
	got, _, err := s.Candidate(ctx, "KR", "005930", t0.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("Candidate: %v", err)
	}
	if !got.FirstSeenAt.Equal(t0) {
		t.Errorf("first_seen_at across the migration = %v, want %v", got.FirstSeenAt, t0)
	}
	if len(got.Sources) != 1 || got.Sources[0] != SourceOfficialPrices {
		t.Errorf("sources across the migration = %v", got.Sources)
	}
	if got.SourcesAttempted != 5 || got.SourcesResponded != 5 {
		t.Errorf("completeness across the migration = %d/%d",
			got.SourcesResponded, got.SourcesAttempted)
	}
	rows, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(rows) != 4 {
		t.Fatalf("%d raw rows survived the migration, want 4", len(rows))
	}

	// The backfill: the oldest surviving priced row becomes the baseline.
	//
	// Leaving it NULL for the next scan to fill would be strictly worse — that
	// reading is newer than everything already in the table, so it would set a
	// later baseline and understate every expansion measured against it. This is
	// the oldest evidence that exists at the moment of the upgrade, and
	// first_price_at is written with it so a reader can see how old that is.
	baseline, found, err := s.Baseline(ctx, "KR", "005930")
	if err != nil || !found {
		t.Fatalf("Baseline after the migration: %v found=%v", err, found)
	}
	if baseline.Price != "10000" || !baseline.At.Equal(t0) || baseline.Source != SourceOfficialPrices {
		t.Errorf("the backfilled baseline = %+v, want 10000 at %v from %q",
			baseline, t0, SourceOfficialPrices)
	}

	// And it is set once from here on, like any other.
	if _, err := s.NoteFirstPrice(ctx, "KR", "005930", "50000", t0.Add(time.Hour), SourceWTSPopular); err != nil {
		t.Fatalf("NoteFirstPrice: %v", err)
	}
	if again, _, _ := s.Baseline(ctx, "KR", "005930"); again.Price != "10000" {
		t.Errorf("a later reading overwrote the backfilled baseline: %q", again.Price)
	}

	// The v3 backfill, which is deliberately narrower than the v2 one.
	//
	// D17's price backfill takes the earliest surviving priced row outright and
	// relies on first_price_at plus the read-time comparison to catch a late one. A
	// rank has no safe direction to be wrong in, so this rung applies the identity
	// window on the way in too: the row at t0 is inside it and becomes the first
	// rank, the row half an hour later is not and does not.
	first, found, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil || !found {
		t.Fatalf("FirstRank after the migration: %v found=%v", err, found)
	}
	if first.Rank != 12 || first.Total != 150 || !first.At.Equal(t0) ||
		first.Source != SourceOfficialTradingValue {
		t.Errorf("the backfilled first rank = %+v, want 12 of 150 at %v from %q",
			first, t0, SourceOfficialTradingValue)
	}

	// The candidate whose only surviving ranked row is an hour late keeps none.
	// Unmeasured is the honest answer there, and it is the one D10 makes safe.
	late, found, err := s.FirstRank(ctx, "KR", "000660")
	if err != nil || !found {
		t.Fatalf("FirstRank of the second candidate: %v found=%v", err, found)
	}
	if late.Recorded() {
		t.Errorf("a row an hour after first_seen_at was backfilled as the first sighting: %+v "+
			"— a rank has no safe direction to be wrong in", late)
	}
	sighting := MeasureFirstSighting(
		Candidate{Key: Key{Market: "KR", Symbol: "000660"}, FirstSeenAt: t0}, late, nil)
	if sighting.Measured {
		t.Errorf("seen_late is measurable for the candidate the backfill left alone: %+v", sighting)
	}

	// And the backfilled one is set once from here on, like the baseline.
	if _, err := s.NoteFirstRank(ctx, "KR", "005930", 3, 150, t0, SourceWTSPopular); err != nil {
		t.Fatalf("NoteFirstRank: %v", err)
	}
	if again, _, _ := s.FirstRank(ctx, "KR", "005930"); again.Rank != 12 {
		t.Errorf("a later reading overwrote the backfilled first rank: %d", again.Rank)
	}

	// Reopening is a no-op rather than a second climb.
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatalf("reopening the migrated store: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	if got, _, err := reopened.Baseline(ctx, "KR", "005930"); err != nil || got.Price != "10000" {
		t.Errorf("the baseline after reopening = %q (%v), want 10000", got.Price, err)
	}
}

// v2Schema is the schema exactly as SchemaVersion 2 wrote it: v1 plus D17's three
// baseline columns. Spelled out rather than derived for v1Schema's reason — a
// ladder test that builds its "old" database from the current source migrates a
// current file and calls it an old one.
const v2Schema = `
CREATE TABLE IF NOT EXISTS observations (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	market         TEXT    NOT NULL,
	symbol         TEXT    NOT NULL,
	source         TEXT    NOT NULL,
	observed_at    TEXT    NOT NULL,
	rank           INTEGER NOT NULL DEFAULT 0,
	rank_total     INTEGER NOT NULL DEFAULT 0,
	newly_listed   INTEGER NOT NULL DEFAULT 0,
	trading_value  TEXT,
	trading_volume TEXT,
	price          TEXT,
	day_high       TEXT,
	day_low        TEXT,
	via_fallback   INTEGER NOT NULL DEFAULT 0,
	CHECK (newly_listed IN (0,1)),
	CHECK (via_fallback IN (0,1)),
	CHECK (rank >= 0 AND rank_total >= 0)
) STRICT;
CREATE INDEX IF NOT EXISTS idx_observations_symbol_time
	ON observations(market, symbol, observed_at);
CREATE INDEX IF NOT EXISTS idx_observations_time ON observations(observed_at);
CREATE TABLE IF NOT EXISTS candidates (
	market             TEXT NOT NULL,
	symbol             TEXT NOT NULL,
	first_seen_at      TEXT NOT NULL,
	last_seen_at       TEXT NOT NULL,
	cooled_at          TEXT,
	sources            TEXT NOT NULL DEFAULT '',
	sources_attempted  INTEGER NOT NULL DEFAULT 0,
	sources_responded  INTEGER NOT NULL DEFAULT 0,
	degraded           INTEGER NOT NULL DEFAULT 0,
	first_price        TEXT,
	first_price_at     TEXT,
	first_price_source TEXT,
	PRIMARY KEY (market, symbol),
	CHECK (degraded IN (0,1))
) STRICT;
CREATE TABLE IF NOT EXISTS store_meta (
	key   TEXT PRIMARY KEY,
	value TEXT NOT NULL
) STRICT;
INSERT INTO store_meta (key, value) VALUES ('schema_version', '2');
`

// writeV2Store builds a schema-2 database: a candidate that already has its
// baseline recorded and its ranked rows still in the raw table, the way a build
// from between D17 and task 4.9 would have left one.
func writeV2Store(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", dsn(path, time.Second))
	if err != nil {
		t.Fatalf("opening a v2 store: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(v2Schema); err != nil {
		t.Fatalf("creating the v2 schema: %v", err)
	}
	_, err = db.Exec(`INSERT INTO candidates
		(market, symbol, first_seen_at, last_seen_at, cooled_at, sources,
		 sources_attempted, sources_responded, degraded,
		 first_price, first_price_at, first_price_source)
		VALUES ('KR','005930',?,?,NULL,'official_rankings_trading_value',5,4,1,
		        '10000',?,'official_prices')`,
		stamp(t0), stamp(t0.Add(time.Minute)), stamp(t0))
	if err != nil {
		t.Fatalf("seeding the v2 candidate: %v", err)
	}
	for _, r := range []struct {
		at   time.Time
		rank int
	}{{t0, 12}, {t0.Add(time.Minute), 8}} {
		_, err = db.Exec(`INSERT INTO observations
			(market, symbol, source, observed_at, rank, rank_total, newly_listed, price, via_fallback)
			VALUES ('KR','005930','official_rankings_trading_value',?,?,150,0,'10000',0)`,
			stamp(r.at), r.rank)
		if err != nil {
			t.Fatalf("seeding a v2 observation: %v", err)
		}
	}
	// A second candidate promoted an hour later, with a ranked row from nine
	// minutes before that — the gap between two lives, or the drift of a symbol
	// nobody had promoted yet. It is inside the symmetric read-time window and it
	// is the shape D20 names, so the backfill must not take it.
	crossed := t0.Add(time.Hour)
	_, err = db.Exec(`INSERT INTO candidates
		(market, symbol, first_seen_at, last_seen_at, sources, sources_attempted, sources_responded, degraded)
		VALUES ('KR','000660',?,?,'official_rankings_trading_value',5,5,0)`,
		stamp(crossed), stamp(crossed))
	if err != nil {
		t.Fatalf("seeding the second v2 candidate: %v", err)
	}
	_, err = db.Exec(`INSERT INTO observations
		(market, symbol, source, observed_at, rank, rank_total, newly_listed, price, via_fallback)
		VALUES ('KR','000660','official_rankings_trading_value',?,148,150,0,'10000',0)`,
		stamp(crossed.Add(-9*time.Minute)))
	if err != nil {
		t.Fatalf("seeding the gap observation: %v", err)
	}
}

// TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows is the second rung on its own.
//
// The v1 test climbs both, which is the case that proves the ladder runs in order
// — and it is exactly the case that cannot catch a rung that only works when the
// one below it has just run in the same transaction. Most stores in the field are
// at v2, so v2 → v3 is the step an upgrade actually takes.
func TestAStoreAtSchemaTwoOpensMigratesAndKeepsItsRows(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "candidates.db")
	writeV2Store(t, path)

	s, err := Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatalf("Open on a v2 store: %v", err)
	}
	defer func() { _ = s.Close() }()

	cols := candidateColumns(t, s)
	for _, want := range []string{
		"first_rank", "first_rank_total", "first_rank_at", "first_rank_source",
	} {
		if !cols[want] {
			t.Errorf("after the migration candidates still has no %s", want)
		}
	}
	var version string
	if err := s.db.QueryRowContext(ctx,
		`SELECT value FROM store_meta WHERE key = 'schema_version'`).Scan(&version); err != nil {
		t.Fatalf("reading the schema version: %v", err)
	}
	if version != fmt.Sprint(SchemaVersion) {
		t.Errorf("schema version after the migration = %q, want %d", version, SchemaVersion)
	}

	// Everything the file already had is still there and unchanged. Additive and
	// nullable is WORKFLOW §0.6's preference precisely so that this holds.
	got, found, err := s.Candidate(ctx, "KR", "005930", t0.Add(2*time.Minute))
	if err != nil || !found {
		t.Fatalf("Candidate: %v found=%v", err, found)
	}
	if !got.FirstSeenAt.Equal(t0) || !got.LastSeenAt.Equal(t0.Add(time.Minute)) {
		t.Errorf("the lifecycle instants across the migration = %v / %v, want %v / %v",
			got.FirstSeenAt, got.LastSeenAt, t0, t0.Add(time.Minute))
	}
	if got.SourcesAttempted != 5 || got.SourcesResponded != 4 || !got.Degraded {
		t.Errorf("completeness across the migration = %d/%d degraded=%v, want 4/5 degraded",
			got.SourcesResponded, got.SourcesAttempted, got.Degraded)
	}
	baseline, _, err := s.Baseline(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("Baseline: %v", err)
	}
	if baseline.Price != "10000" || !baseline.At.Equal(t0) {
		t.Errorf("the baseline across the migration = %+v, want 10000 at %v", baseline, t0)
	}
	rows, err := s.Observations(ctx, "KR", "005930", time.Time{})
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("%d raw rows survived the migration, want 2", len(rows))
	}

	// And the new rung's own work: the earliest ranked row inside the identity
	// window becomes the first rank, so seen_late answers for a candidate that
	// predates the column instead of collapsing to unmeasured.
	first, _, err := s.FirstRank(ctx, "KR", "005930")
	if err != nil {
		t.Fatalf("FirstRank: %v", err)
	}
	if first.Rank != 12 || first.Total != 150 || !first.At.Equal(t0) {
		t.Errorf("the backfilled first rank = %+v, want 12 of 150 at %v", first, t0)
	}
	sighting := MeasureFirstSighting(got, first, rows)
	if !sighting.Measured || sighting.PercentilePct != "92" {
		t.Errorf("the sighting after the migration = measured %v at %s%%, want 92%%",
			sighting.Measured, sighting.PercentilePct)
	}

	// And the row from the gap is not backfilled, which is the whole point of the
	// rung being narrower than D17's. The read-time window is symmetric because a
	// scan can stamp the promotion and the reading a moment apart; at migration
	// time the earlier side is where D20's two shapes live and nothing downstream
	// could tell a baked-in wrong rank from a right one.
	gap, _, err := s.FirstRank(ctx, "KR", "000660")
	if err != nil {
		t.Fatalf("FirstRank of the second candidate: %v", err)
	}
	if gap.Recorded() {
		t.Errorf("a ranked row nine minutes before first_seen_at was backfilled as the first "+
			"sighting: %+v — that is the gap between two lives, and it reads 148th for a "+
			"candidate that crossed near the top", gap)
	}

	// Reopening is a no-op rather than a second climb.
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	reopened, err := Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatalf("reopening the migrated store: %v", err)
	}
	defer func() { _ = reopened.Close() }()
	if again, _, err := reopened.FirstRank(ctx, "KR", "005930"); err != nil || again.Rank != 12 {
		t.Errorf("the first rank after reopening = %d (%v), want 12", again.Rank, err)
	}
}

// TestAStoreFromANewerBuildIsRefused keeps the other end of the ladder honest: an
// older binary must not read a file whose columns it does not know about as a
// file whose values are absent.
func TestAStoreFromANewerBuildIsRefused(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "candidates.db")
	s, err := Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE store_meta SET value = ? WHERE key = 'schema_version'`,
		fmt.Sprint(SchemaVersion+1)); err != nil {
		t.Fatalf("stamping a future version: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	_, err = Open(ctx, Options{Path: path, FSProber: FixedFSProber(FSInfo{Name: "ext4"})})
	if !errors.Is(err, ErrSchemaTooNew) {
		t.Fatalf("Open on a newer schema = %v, want ErrSchemaTooNew", err)
	}
}

// --- the source id is normalised on both sides of the write ----------------------

// TestAPaddedSourceIdIsNotASilentWarmUp.
//
// Market and symbol were normalised on every read from the beginning and the
// source was not, so a padded or mis-cased id selected no rows and returned no
// error. Downstream an empty series is WARMING_UP — "the scan has just started,
// nothing needs fixing" — for something that will never start.
func TestAPaddedSourceIdIsNotASilentWarmUp(t *testing.T) {
	s, _ := openStore(t)
	ctx := context.Background()
	if err := s.RecordObservations(ctx, []Observation{
		obs("005930", t0, SourceOfficialPrices, Reported{Price: "70000"}),
	}); err != nil {
		t.Fatalf("RecordObservations: %v", err)
	}
	for _, id := range []SourceID{" official_prices", "official_prices ", "Official_Prices"} {
		got, err := s.SourceObservations(ctx, "KR", "005930", id, time.Time{})
		if err != nil {
			t.Fatalf("SourceObservations(%q): %v", id, err)
		}
		if len(got) != 1 {
			t.Errorf("SourceObservations(%q) returned %d rows and no error; downstream that is "+
				"WARMING_UP — \"the scan just started\" — for a wiring defect", id, len(got))
		}
	}
	// And the write side agrees with the read side, so a padded id cannot be
	// stored in a spelling no read will ever find.
	if err := s.RecordObservations(ctx, []Observation{
		obs("000660", t0, " Official_Prices ", Reported{Price: "200000"}),
	}); err != nil {
		t.Fatalf("RecordObservations with a padded source: %v", err)
	}
	got, err := s.SourceObservations(ctx, "KR", "000660", SourceOfficialPrices, time.Time{})
	if err != nil {
		t.Fatalf("SourceObservations: %v", err)
	}
	if len(got) != 1 || got[0].Source != SourceOfficialPrices {
		t.Errorf("a padded id was written verbatim and is unreachable: %d rows", len(got))
	}

	if _, err := s.SourceObservations(ctx, "KR", "005930", "  ", time.Time{}); err == nil {
		t.Error("an empty source id was accepted and answered with silence")
	}
	if _, err := s.SourceSeries(ctx, "KR", "005930", "  ", time.Time{}); err == nil {
		t.Error("SourceSeries accepted an empty source id")
	}
}
