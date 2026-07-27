package candidate

import (
	"context"
	"errors"
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
