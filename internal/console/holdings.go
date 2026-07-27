package console

// holdings.go is the console's whole relationship with a broker (change
// add-operator-dashboard, task 1.3).
//
// # The interface is the safety property
//
// Until this file the console reached an account only through the StartVerify it
// was handed, and internal/verifylive owned every request. The dashboard needs
// one more thing — what the account is actually holding — and the temptation is
// to reuse verifylive.Broker, which is already wired in cmd/tossctl and already
// has a Holdings method.
//
// That would hand a read-only screen PlaceOrder, CancelOrder, ModifyOrder and
// three conditional-order mutations. Nothing would call them today. The interface
// below declares one method instead, and static_test.go fails if a second one
// appears: "the console cannot place an order" becomes a property of the type it
// is given rather than of what its handlers happen to do.
//
// # The rate budget, spelled out (§0.4)
//
// This is the first broker call the console has ever made, so the accounting is
// explicit and it is enforced by holdingsCache rather than described:
//
//	lazy        a refresh happens because a page was requested — by the operator,
//	            or by the positions screen re-opening itself at the TTL (change
//	            refresh-positions-screen). There is no server-side poller: an open
//	            positions tab costs at most one call per TTL, a closed one costs
//	            nothing, and a verification in progress suspends even that.
//	one call    a refresh is exactly one GET /api/v1/holdings. The current price
//	            comes from that response's lastPrice; a per-symbol quote fan-out
//	            would turn one call into one-per-holding for a number the first
//	            response already carried.
//	TTL         holdingsTTL and no shorter than 15 seconds. Refresh-mashing and a
//	            second browser tab are free, and the page shows the cache's age so
//	            nobody has to guess how old the number is.
//	yielding    while a live verification is walking the catalogue, the refresh is
//	            withheld entirely. The verification is the scarce thing: a lost
//	            step costs a person another supervised run, and a dashboard is
//	            worth nothing in particular (internal/runlock's package doc makes
//	            the same trade for the soak).

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
)

// HoldingsReader is the console's entire view of a broker: one read.
//
// Every method added here is a capability the operator console gains, so there
// is exactly one and static_test.go's
// TestTheConsoleBrokerInterfaceDeclaresNothingButReads fails if that changes.
// cmd/tossctl supplies the implementation; this package imports no client of any
// kind.
type HoldingsReader interface {
	// Holdings returns the account's stock positions. An empty symbol asks for
	// all of them, which is the only way the console calls it — one call per
	// refresh is the rate-budget contract, and a per-symbol loop would break it.
	Holdings(ctx context.Context, symbol string) ([]domain.Position, error)
}

// holdingsTTL is how long one broker reading is served from memory.
//
// The operator-console spec fixes a floor of 15 seconds; this is double it,
// because the numbers on the screen are a position's size and cost basis rather
// than a tape, and the page prints the reading's age next to them. A shorter TTL
// would buy precision nobody asked for out of a budget a live verification needs.
const holdingsTTL = 30 * time.Second

// holdingsTimeout bounds one refresh. A broker that has stopped answering must
// not hold an HTTP handler open: the page renders the previous reading, or an
// honest "no data", and says why.
const holdingsTimeout = 10 * time.Second

// holdingsSnapshot is one reading of the account, as the page sees it.
//
// Every field is a state the screen has words for. Present is false in three
// different situations — no wiring, nothing fetched yet, a failed first fetch —
// and each of them has its own sentence, because "no holdings" and "we could not
// ask" are very different things to tell somebody about their money.
type holdingsSnapshot struct {
	// Wired reports that a broker was supplied at all.
	Wired bool
	// Present reports that a reading exists (possibly an old one).
	Present bool
	// Rows is the reading. Nil is a legitimate answer: an account can hold
	// nothing.
	Rows []domain.Position
	// At is when the reading was taken, and Age how old it is now.
	At  time.Time
	Age time.Duration
	// Error is the last refresh failure, empty when the last attempt succeeded.
	Error string
	// Held reports that the refresh was withheld, and HeldReason says by what.
	Held       bool
	HeldReason string
}

// Stale reports a reading older than the TTL. It is only ever true alongside
// Held: an unheld request refreshes instead.
func (s holdingsSnapshot) Stale() bool { return s.Present && s.Age > holdingsTTL }

// AgeSeconds is the reading's age for the page, rounded to a whole second.
func (s holdingsSnapshot) AgeSeconds() int { return int(s.Age.Round(time.Second) / time.Second) }

// TakenAt renders the reading's timestamp.
func (s holdingsSnapshot) TakenAt() string {
	if !s.Present {
		return "—"
	}
	return s.At.UTC().Format("2006-01-02 15:04:05Z")
}

// holdingsCache is the lazy, single-flight cache in front of one HoldingsReader.
//
// The mutex is held across the fetch on purpose. Two tabs refreshing at the same
// moment then cost one broker call rather than two: the second waits, finds the
// reading fresh and returns it. The cost is that one slow broker delays both
// pages, which is the right way round — a local console with one operator has no
// concurrency to protect, and the rate budget does.
type holdingsCache struct {
	reader HoldingsReader
	ttl    time.Duration

	mu      sync.Mutex
	rows    []domain.Position
	at      time.Time
	present bool
	lastErr string
	// tried is when the last refresh was attempted, successful or not. The TTL
	// bounds attempts rather than successes: a broker that is answering 429 would
	// otherwise be asked again on every page load, which is the one moment the
	// budget is already in trouble.
	tried     time.Time
	attempted bool
}

func newHoldingsCache(reader HoldingsReader, ttl time.Duration) *holdingsCache {
	if ttl <= 0 {
		ttl = holdingsTTL
	}
	return &holdingsCache{reader: reader, ttl: ttl}
}

// get returns the current reading, refreshing it if it is older than the TTL.
//
// hold is the verification-in-progress signal. When it is set nothing is
// fetched, whatever the age: the cached reading is returned with the reason, and
// a cold cache stays cold and says so. That is the "검증 중 — 갱신 보류" /
// "검증 중 — 데이터 없음" pair in the spec, and it is one branch rather than two
// because the difference is only whether anything was cached first.
func (c *holdingsCache) get(ctx context.Context, now time.Time, hold bool, holdReason string) holdingsSnapshot {
	if c == nil || c.reader == nil {
		return holdingsSnapshot{Held: hold, HeldReason: holdReason}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
		c.refreshLocked(ctx, now)
	}

	snap := holdingsSnapshot{
		Wired:      true,
		Present:    c.present,
		Rows:       append([]domain.Position(nil), c.rows...),
		At:         c.at,
		Error:      c.lastErr,
		Held:       hold,
		HeldReason: holdReason,
	}
	if c.present {
		snap.Age = now.Sub(c.at)
		if snap.Age < 0 {
			snap.Age = 0
		}
	}
	return snap
}

// refreshLocked performs the one broker call a refresh is allowed.
//
// A failure keeps the previous reading rather than discarding it. An operator
// looking at a five-minute-old holdings list with "브로커 조회 실패" next to it is
// better informed than one looking at an empty table, and the age is on the page
// either way.
func (c *holdingsCache) refreshLocked(ctx context.Context, now time.Time) {
	c.tried, c.attempted = now, true

	fetchCtx, cancel := context.WithTimeout(ctx, holdingsTimeout)
	defer cancel()

	rows, err := c.reader.Holdings(fetchCtx, "")
	if err != nil {
		c.lastErr = err.Error()
		return
	}
	c.rows, c.at, c.present, c.lastErr = rows, now, true, ""
}

// --- the yielding rule ------------------------------------------------------------

// verifyHold answers one question: is a supervised live verification spending
// this account's rate budget right now?
//
// It is answered from two different places because the two cases are genuinely
// different, and a single mechanism would get one of them wrong:
//
//	this process   the run is an object in memory. Asking a file about a thing
//	               this process is doing would be a race with its own writer, and
//	               it would keep answering "no" for the first second of every run.
//	another one    there is no IPC and there is not going to be. internal/runlock
//	               already publishes exactly this fact as a file's modification
//	               time, with a staleness bound, precisely so a crashed
//	               verification costs one pause instead of wedging the reader
//	               forever. The console reads it and never writes it.
func (c *Console) verifyHold(now time.Time) (bool, string) {
	if run := c.currentRun(); run != nil && !run.finished() {
		return true, "이 콘솔이 실계좌 검증을 실행 중이다"
	}
	path := strings.TrimSpace(c.opts.RunLockPath)
	if path == "" {
		return false, ""
	}
	fresh, at := runlock.Fresh(path, now, runlock.StaleAfter)
	if !fresh {
		return false, ""
	}
	return true, "다른 프로세스가 실계좌 검증을 실행 중이다 (마커 " +
		at.UTC().Format("15:04:05Z") + ", " + runlock.StaleAfter.String() + " 신선도 상한)"
}
