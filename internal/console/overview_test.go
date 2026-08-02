package console

// overview_test.go covers the operator overview screen (change
// console-operator-overview).
//
// The claims worth testing here are the ones an operator would be misled by if
// they were wrong:
//
//	nothing is invented   a value nobody read renders as unmeasured with the
//	                      reason, never as zero and never as an unexplained dash.
//	no broker call        the longest-lived screen in the console must not be the
//	                      one that spends the rate budget.
//	no currency mixing    KRW and USD are never added, in any panel.
//	the day is the market's the boundary is local midnight, and the screen prints
//	                      which one it used.
//	nothing to press      no form, no POST, no confirmation friction of any kind.

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// --- fixtures ------------------------------------------------------------------------

// fakeGateLimits is the Guardian-limits seam, as cmd/tossctl would fill it.
type fakeGateLimits struct {
	limits GateLimits
	err    error
}

func (f fakeGateLimits) GateLimits() (GateLimits, error) { return f.limits, f.err }

// overviewNow is the instant the day-boundary tests are taken at: 23:30 UTC,
// which is 08:30 the NEXT morning in Seoul and 19:30 the SAME evening in New
// York. It is chosen because a UTC "today" and a Seoul "today" disagree about a
// whole trading session there, which is the failure D6 exists to prevent.
var overviewNow = time.Date(2026, 7, 27, 23, 30, 0, 0, time.UTC)

// overviewTripsFixture is three KR round trips and one US one.
//
//	kr-yesterday  closed 14:00 KST on the 27th. Seoul says yesterday. UTC says
//	              today, because the UTC day began at 09:00 KST on the 27th — an
//	              implementation using UTC midnight reports a 500,000 loss the
//	              operator already lived through as this morning's.
//	kr-today-win  closed 01:00 KST on the 28th. Seoul's today.
//	kr-today-loss the same morning, on the other side.
//	us-today      closed 14:00 ET on the 27th, which is New York's today.
const overviewTripsFixture = `
INSERT INTO decisions (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
                       risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
VALUES ('d-ov','123-45-678901',0,'EXPOSURE_RAISING','RISK_INTENT','{}','h-ov',NULL,'{}','n-ov',
        '2026-07-26T00:30:00Z','2026-07-26T00:40:00Z');

INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('kr-yesterday','123-45-678901','kr','005930',9,'d-ov','CLOSED','0','0','2026-07-27T05:00:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('kr-yesterday','-500000','-1.0','500000','10',3600,NULL,NULL,'2026-07-27T05:00:00Z');

INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('kr-today-win','123-45-678901','kr','035420',9,'d-ov','CLOSED','0','0','2026-07-27T16:00:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('kr-today-win','120000','1.2','100000','5',600,NULL,NULL,'2026-07-27T16:00:00Z');

INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('kr-today-loss','123-45-678901','kr','000660',9,'d-ov','CLOSED','0','0','2026-07-27T17:00:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('kr-today-loss','-20000','-0.2','100000','3',300,NULL,NULL,'2026-07-27T17:00:00Z');

INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('us-today','123-45-678901','us','AAPL',9,'d-ov','CLOSED','0','0','2026-07-27T18:00:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('us-today','250.50','0.9','280','2',900,NULL,NULL,'2026-07-27T18:00:00Z');
`

// overviewBrokenClosedAtFixture is one KR round trip whose closed_at is not a
// timestamp. The ledger opens, answers, and hands back a value the screen cannot
// date — which is a different thing from a ledger that would not open.
const overviewBrokenClosedAtFixture = `
INSERT INTO decisions (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
                       risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
VALUES ('d-bad','123-45-678901',0,'EXPOSURE_RAISING','RISK_INTENT','{}','h-bad',NULL,'{}','n-bad',
        '2026-07-26T00:30:00Z','2026-07-26T00:40:00Z');

INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('kr-broken','123-45-678901','kr','005930',9,'d-bad','CLOSED','0','0','not-a-timestamp');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('kr-broken','1000','0.1','10000','1',60,NULL,NULL,'not-a-timestamp');
`

// overviewForeignMarketFixture is a round trip in a market this screen does not
// split by. Before the 기타/미분류 row it was on /history and nowhere on the
// overview.
const overviewForeignMarketFixture = `
INSERT INTO decisions (id, account_ref, generation, safety_class, preimage_kind, risk_preimage,
                       risk_hash, client_order_id, limits_json, nonce, issued_at, expires_at)
VALUES ('d-jp','123-45-678901',0,'EXPOSURE_RAISING','RISK_INTENT','{}','h-jp',NULL,'{}','n-jp',
        '2026-07-26T00:30:00Z','2026-07-26T00:40:00Z');

INSERT INTO positions (id, account_ref, market, symbol, instance_seq, entry_decision_id, state,
                       quantity, avg_price, closed_at)
VALUES ('jp-today','123-45-678901','jp','7203',9,'d-jp','CLOSED','0','0','2026-07-27T18:00:00Z');
INSERT INTO trade_outcomes (position_id, realized_pnl_after_costs, realized_r, initial_risk,
                            initial_quantity, held_seconds, exit_ratchet_level, exit_rung, closed_at)
VALUES ('jp-today','9000','0.4','20000','5',900,NULL,NULL,'2026-07-27T18:00:00Z');
`

// overviewHarness is the dashboard harness with the clock pinned and the
// Guardian-limits seam wired.
type overviewHarness struct {
	*dashboardHarness
}

func newOverviewHarness(t *testing.T, tweak ...func(*Options)) *overviewHarness {
	t.Helper()
	fixed := append([]func(*Options){}, tweak...)
	h := &overviewHarness{dashboardHarness: newDashboardHarness(t, fixed...)}
	// The clock is the dashboard harness's own, moved to overviewNow, so a test
	// can push it past the cache TTL — a frozen clock would hide a refresh that
	// only happens once the reading is stale, which is exactly the refresh the
	// zero-call rule is about.
	h.clock.advance(overviewNow.Sub(h.clock.Now()))
	h.authenticate(t)
	return h
}

// marketRow finds one market's row by label.
func todayRow(panel todayPanel, market string) todayMarketRow {
	for _, r := range panel.Markets {
		if r.Market == market {
			return r
		}
	}
	return todayMarketRow{}
}

func accountRow(panel accountPanel, market string) accountMarketRow {
	for _, r := range panel.Markets {
		if r.Market == market {
			return r
		}
	}
	return accountMarketRow{}
}

// --- the cache read that does not refresh (task 2.1) --------------------------------

// TestThereIsAReadOfTheBrokerCacheThatNeitherRefreshesNorInventsAReason.
//
// get() refreshes whenever the request is unheld and the TTL has passed, so the
// only way it reached zero broker calls was hold=true — and hold=true is rendered
// as "검증 중 — 갱신 보류" by the screen. On a machine where no verification is
// running that sentence is a fabrication, and it is a fabrication about why an
// operator's account numbers are old.
func TestThereIsAReadOfTheBrokerCacheThatNeitherRefreshesNorInventsAReason(t *testing.T) {
	counter := &countingHoldings{rows: brokerRows()}
	cache := newHoldingsCache(counter, holdingsTTL)
	now := time.Date(2026, 7, 28, 1, 0, 0, 0, time.UTC)

	snap := cache.peek(now)

	if n := counter.count(); n != 0 {
		t.Errorf("peek made %d broker call(s); it is the read that does not spend the budget", n)
	}
	if snap.Held || snap.HeldReason != "" {
		t.Errorf("peek reported Held=%v %q; no verification is running and the screen must not say one is",
			snap.Held, snap.HeldReason)
	}
	if !snap.Wired {
		t.Error("peek reported the broker unwired; one was supplied")
	}
	if snap.Present {
		t.Error("peek reported a reading on a cache nothing has ever filled")
	}
}

// TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing.
//
// The other half: once /positions has filled the cache, the overview shows those
// numbers with their age, and still costs nothing.
func TestPeekServesWhatTheLastRefreshFoundAndStillCallsNothing(t *testing.T) {
	counter := &countingHoldings{rows: brokerRows()}
	cache := newHoldingsCache(counter, holdingsTTL)
	now := time.Date(2026, 7, 28, 1, 0, 0, 0, time.UTC)

	if snap := cache.get(t.Context(), now, false, ""); !snap.Present {
		t.Fatal("the first unheld get did not fill the cache")
	}
	if n := counter.count(); n != 1 {
		t.Fatalf("the first get made %d broker call(s), want 1", n)
	}

	later := now.Add(90 * time.Second)
	snap := cache.peek(later)

	if n := counter.count(); n != 1 {
		t.Errorf("peek took the call count to %d; a peek past the TTL still refreshes nothing", n)
	}
	if !snap.Present || len(snap.Rows) != len(brokerRows()) {
		t.Errorf("peek returned Present=%v with %d row(s); it serves the last reading",
			snap.Present, len(snap.Rows))
	}
	if snap.Age != 90*time.Second {
		t.Errorf("peek reported an age of %s, want 90s; the screen prints how old the numbers are", snap.Age)
	}
	if snap.Held {
		t.Error("peek reported the refresh withheld; nothing withheld it")
	}
}

// --- the broker budget (task 4.2) ---------------------------------------------------

// TestTheOverviewMakesNoBrokerCall.
//
// This screen reloads itself every TTL and an operator leaves it open all day.
// If it refreshed the cache, the longest-lived tab in the console would become a
// permanent standing charge on the account's rate budget — and the scarce thing
// that budget protects is a supervised live verification, where a lost step costs
// a person another market window.
func TestTheOverviewMakesNoBrokerCall(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	// Fill the cache the way it is meant to be filled, then forget the cost.
	h.page(t, "/positions")
	before := h.holdings.count()

	// Past the TTL, which is when a refreshing read would fire. A frozen clock
	// would let a get() slip through this test unnoticed.
	for i := 0; i < 3; i++ {
		h.clock.advance(2 * holdingsTTL)
		h.page(t, "/dashboard")
	}
	if got := h.holdings.count(); got != before {
		t.Errorf("three overview renders took the broker call count from %d to %d; the overview "+
			"reads the cache and refreshes nothing", before, got)
	}
}

// TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt.
//
// The cost of making no broker call, written down: with /positions never opened
// the account panel is empty forever, so it has to say which of the five reasons
// applies and it has to carry the way out. never_fetched and verify_suspended are
// different afternoons — the second one is a fabrication when nothing is running,
// and it was the natural way to reach zero calls before peek existed.
func TestAColdBrokerCacheSaysNotYetReadAndLinksToTheScreenThatFillsIt(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	page := h.page(t, "/dashboard")
	if h.holdings.count() != 0 {
		t.Fatalf("a cold overview made %d broker call(s)", h.holdings.count())
	}

	view := h.overview(t.Context())
	row := accountRow(view.Account, "KR")
	if row.Unreadable.Code() != string(reasonNeverFetched) {
		t.Errorf("a never-filled cache renders as %q, want %q", row.Unreadable.Code(), reasonNeverFetched)
	}
	if strings.Contains(page, "갱신 보류") {
		t.Error("the overview claims a verification is suspending the refresh; none is running")
	}
	// The ROW's link, not any link. The intro paragraph carries href="/positions"
	// unconditionally, so `Contains(page, href="/positions")` was true whether or
	// not the row offered a way out — a reviewer deleted the row-level anchor and
	// the whole package still passed. D4 required this link in the spec precisely
	// because "산문에만 있으면 링크 없는 구현도 계약을 만족한다", and the test had the
	// same defect the design predicted for the prose.
	if !strings.Contains(page, overviewFillLink) {
		t.Errorf("the unread account row carries no link to the screen that fills it (looked for %q); "+
			"the number stays empty for as long as the operator believes the screen", overviewFillLink)
	}

	// And the row-level link is a consequence of the row being unread, not a
	// fixture of the markup: once the cache is filled it is gone.
	filled := newOverviewHarness(t)
	seedJournal(t, filled.journal)
	filled.holdings.rows = brokerRows()
	filled.page(t, "/positions")
	if got := filled.page(t, "/dashboard"); strings.Contains(got, overviewFillLink) {
		t.Error("the row-level link is rendered on a filled cache too, so its presence proves nothing " +
			"about the unread row")
	}
}

// overviewFillLink is the row-level anchor the account panel renders beside a
// never_fetched row. It is spelled here so the assertion is about the row and not
// about the page happening to mention /positions somewhere.
const overviewFillLink = `<a href="/positions">포지션 화면을 열면 이 값이 채워진다</a>`

// TestTheOverviewReadsTheBrokerCacheWithPeekAndNothingElse.
//
// Counting calls catches only half the trap D4 names. The other half —
// get(ctx, now, true, "검증 중 — 갱신 보류") — also costs nothing, and it carries a
// held reason into the snapshot that no verification produced. This screen
// happens to discard that field today, so no runtime assertion can see the
// difference; the day anything on the overview renders holdingsSnapshot.HeldReason
// (templates_portfolio.go's "brokerstate" already does, on the other screen) the
// fabrication is in front of an operator with no test failing. The read this
// screen is allowed is peek, and that is checked where it is visible: in the
// source.
func TestTheOverviewReadsTheBrokerCacheWithPeekAndNothingElse(t *testing.T) {
	code := strings.Join(nonCommentLines(packageFiles(t)["overview.go"]), "\n")
	if !strings.Contains(code, "c.holdings.peek(") {
		t.Error("overview.go no longer reads the broker cache with peek; the guard is checking a " +
			"claim nobody makes any more")
	}
	if strings.Contains(code, "c.holdings.get(") {
		t.Error("overview.go reads the broker cache with get. Unheld, that refreshes and puts the " +
			"console's longest-lived tab on the rate budget; held, it costs nothing but stamps the " +
			"snapshot with a suspended verification that is not running")
	}
}

// TestAFailedBrokerReadIsNotTheSameAsAnEmptyCache.
func TestAFailedBrokerReadIsNotTheSameAsAnEmptyCache(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	h.holdings.err = errBrokerDown

	// Let the positions screen take the failure, so the cache holds the error.
	h.page(t, "/positions")

	row := accountRow(h.overview(t.Context()).Account, "KR")
	if row.Unreadable.Code() != string(reasonBrokerReadFailed) {
		t.Errorf("a failed read renders as %q, want %q — the operator's move is to fix the "+
			"credentials, not to open another screen", row.Unreadable.Code(), reasonBrokerReadFailed)
	}
}

// TestARunningVerificationIsANoteAndNotTheReasonTheCacheIsCold.
//
// The overview refreshes nothing (D4), so a verification in progress withholds
// nothing FROM THIS SCREEN. Rendering verify_suspended here told the operator
// "끝나면 다시 읽힌다" — false twice over: nothing was being held back, and the
// value will not appear when the run ends either, because only /positions ever
// fills this cache. A cold cache is never_fetched whether or not a run is alive,
// which is what brokerReadable's own doc comment has said since it was written.
//
// The run is still worth mentioning, because it suspends the way out: opening
// /positions will not refresh either until it finishes. That belongs beside the
// panel as a note about the other screen, not on the value as its reason.
func TestARunningVerificationIsANoteAndNotTheReasonTheCacheIsCold(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	holdRunLock(t, h.runlock, overviewNow)

	view := h.overview(t.Context())
	row := accountRow(view.Account, "KR")
	if row.Unreadable.Code() != string(reasonNeverFetched) {
		t.Errorf("with a fresh run marker and a cold cache the reason is %q, want %q — this screen "+
			"refreshes nothing, so nothing is suspended for it and nothing will be read when the "+
			"run ends", row.Unreadable.Code(), reasonNeverFetched)
	}
	if !view.VerifyRunning {
		t.Fatal("a fresh run marker is not reported as a verification in progress")
	}

	page := h.page(t, "/dashboard")
	if !strings.Contains(page, "실계좌 검증이 실행 중이다") {
		t.Error("the screen does not mention the verification at all; the operator is told to open " +
			"/positions and that screen will not refresh until the run ends")
	}
	if strings.Contains(page, string(reasonVerifySuspended)) {
		t.Error("the screen still names verify_suspended as a reason a value on it is missing; it " +
			"withholds nothing")
	}
	if strings.Contains(page, "끝나면 다시 읽힌다") {
		t.Error("the screen promises the value will be read when the verification ends; only opening " +
			"/positions fills this cache")
	}
}

// TestAColdCacheWithNoVerificationSaysNothingAboutOne is the other side: the note
// is a reading of the run marker and not decoration.
func TestAColdCacheWithNoVerificationSaysNothingAboutOne(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	view := h.overview(t.Context())
	if view.VerifyRunning {
		t.Fatal("no run marker was written and the screen reports a verification in progress")
	}
	if strings.Contains(h.page(t, "/dashboard"), "실계좌 검증이 실행 중이다") {
		t.Error("the screen announces a verification nobody started")
	}
}

// --- no currency mixing (task 4.4) ---------------------------------------------------

// TestTheOverviewNeverAddsAcrossMarkets.
//
// domain.Position carries a MarketType and no currency at all, so a KR row is
// KRW and a US row is USD. One number made by adding them is not a total, it is
// an arithmetic accident with a currency symbol nobody can write next to it.
func TestTheOverviewNeverAddsAcrossMarkets(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	h.holdings.rows = append(brokerRows(), usHolding())
	h.page(t, "/positions")

	view := h.overview(t.Context())
	kr := accountRow(view.Account, "KR")
	us := accountRow(view.Account, "US")

	if !kr.Value.Known() || !us.Value.Known() {
		t.Fatalf("both markets should have values: KR=%q US=%q", kr.Value.Value(), us.Value.Value())
	}
	if kr.Value.Value() == us.Value.Value() {
		t.Errorf("both markets report %q; the rows are not separate", kr.Value.Value())
	}
	// KRW 730,000 + 1,330,000 + 394,500 = 2,454,500 and USD 3,000. A total would
	// be 2,457,500, which is a number in no currency.
	if strings.Contains(h.page(t, "/dashboard"), "2,457,500") {
		t.Error("the account panel prints a cross-market total; KRW and USD were added together")
	}

	todayKR := todayRow(view.Today, "KR")
	todayUS := todayRow(view.Today, "US")
	if todayKR.Currency == todayUS.Currency {
		t.Errorf("both today rows are denominated %q; the panel is not split by market", todayKR.Currency)
	}
}

// TestAHoldingInNeitherMarketIsNamedRatherThanDropped.
//
// The two market rows are matched by label, and anything that matched neither
// used to fall through the loop with no row and no marker — while /positions went
// on listing it. Two screens in one console then disagree about what the account
// holds, and it is D10's rule ("a value that disappears without a label is not an
// unmeasured marker, it is concealment") broken on a symbol rather than on a
// cell. The row is unmeasured rather than summed because nothing here knows what
// currency it is in, and inventing one is what D6 forbids.
func TestAHoldingInNeitherMarketIsNamedRatherThanDropped(t *testing.T) {
	h := newOverviewHarness(t)
	seedEngineJournal(t, h.journal, overviewForeignMarketFixture)
	h.holdings.rows = append(brokerRows(), domain.Position{
		Symbol: "7203", Name: "Toyota", MarketType: "JP", Quantity: 100, AveragePrice: 2000,
		CurrentPrice: 2100, MarketValue: 210000, UnrealizedPnL: 10000,
	})
	h.page(t, "/positions")

	view := h.overview(t.Context())

	other := accountRow(view.Account, overviewOtherMarket)
	if other.Market != overviewOtherMarket {
		t.Fatalf("the JP holding produced no row at all: %+v", view.Account.Markets)
	}
	if !other.Blocked || other.Unreadable.Known() {
		t.Errorf("the 기타 row claims a measured value: %+v", other)
	}
	if !strings.Contains(other.Unreadable.Why(), "JP:7203") {
		t.Errorf("the 기타 row does not name the holding it stands for: %q", other.Unreadable.Why())
	}

	otherToday := todayRow(view.Today, overviewOtherMarket)
	if otherToday.Market != overviewOtherMarket {
		t.Fatalf("the JP round trip produced no today row at all: %+v", view.Today.Markets)
	}
	if !otherToday.Blocked || !strings.Contains(otherToday.Unreadable.Why(), "JP:7203") {
		t.Errorf("the 기타 today row does not name the round trip it stands for: %+v", otherToday)
	}

	// And the KR row is untouched: the catch-all is not a place rows go to be
	// double-counted.
	kr := accountRow(view.Account, "KR")
	if !kr.Value.Known() || strings.Contains(kr.Value.Value(), "210,000") {
		t.Errorf("the JP holding leaked into the KR row: %q", kr.Value.Value())
	}

	page := h.page(t, "/dashboard")
	if !strings.Contains(page, overviewOtherMarket) {
		t.Error("the screen does not render the 기타/미분류 row")
	}
	if !strings.Contains(page, "JP:7203") {
		t.Error("the screen does not name the symbol the operator would have to look for on /positions")
	}
}

// TestTwoMarketsAloneProduceNoOtherRow is the negative control: the catch-all is
// a consequence of an unmatched row and not a permanent fixture that would make
// the assertion above pass on an empty account.
func TestTwoMarketsAloneProduceNoOtherRow(t *testing.T) {
	h := newOverviewHarness(t)
	seedEngineJournal(t, h.journal, overviewTripsFixture)
	h.holdings.rows = append(brokerRows(), usHolding())
	h.page(t, "/positions")

	view := h.overview(t.Context())
	if len(view.Account.Markets) != 2 {
		t.Errorf("the account panel has %d rows for a KR+US account: %+v",
			len(view.Account.Markets), view.Account.Markets)
	}
	if len(view.Today.Markets) != 2 {
		t.Errorf("the today panel has %d rows for a KR+US account: %+v",
			len(view.Today.Markets), view.Today.Markets)
	}
	if strings.Contains(h.page(t, "/dashboard"), overviewOtherMarket) {
		t.Error("the screen renders an 기타/미분류 row for an account that has nothing in one")
	}
}

// --- the day boundary (task 4.6) -----------------------------------------------------

// TestTodayIsEachMarketsOwnLocalMidnightAndTheScreenSaysWhichOne.
//
// trade_outcomes.closed_at is UTC. At 08:30 in Seoul the UTC day began at 09:00
// KST *yesterday*, so a UTC boundary counts the whole of yesterday's Korean
// session as today — the fixture's 500,000 loss is exactly that trade, and an
// operator reading it as this morning's would think they had already spent the
// day's loss budget.
func TestTodayIsEachMarketsOwnLocalMidnightAndTheScreenSaysWhichOne(t *testing.T) {
	h := newOverviewHarness(t)
	seedEngineJournal(t, h.journal, overviewTripsFixture)

	view := h.overview(t.Context())
	kr := todayRow(view.Today, "KR")

	if kr.NoTrades || !kr.Trips.Known() {
		t.Fatalf("the KR row has no counted trips: %+v", kr)
	}
	if kr.Trips.Value() != "2" {
		t.Errorf("KR counted %s round trip(s) today, want 2; yesterday's Seoul session is not today",
			kr.Trips.Value())
	}
	if kr.PnL.Value() != "100,000" {
		t.Errorf("KR realised %s today, want 100,000; a UTC boundary would say -400,000 by folding "+
			"yesterday's 500,000 loss into this morning", kr.PnL.Value())
	}
	if kr.Wins.Value() != "1" || kr.Losses.Value() != "1" {
		t.Errorf("KR win/loss = %s/%s, want 1/1", kr.Wins.Value(), kr.Losses.Value())
	}

	us := todayRow(view.Today, "US")
	if us.PnL.Value() != "250.5" {
		t.Errorf("US realised %s today, want 250.5", us.PnL.Value())
	}

	page := h.page(t, "/dashboard")
	if !strings.Contains(page, "Asia/Seoul") || !strings.Contains(page, "America/New_York") {
		t.Error("the screen does not print which midnight each market counted from; choosing a " +
			"boundary and hiding the choice are different mistakes")
	}
	if !strings.Contains(kr.Boundary, "Asia/Seoul") {
		t.Errorf("the KR boundary is %q and does not name the zone it used", kr.Boundary)
	}
}

// TestAMarketWithNoClosedRoundTripsSaysSoRatherThanZero.
//
// Zero realised P&L and no trades look identical on a screen and mean opposite
// things: one is a flat day, the other is a day nothing happened on.
func TestAMarketWithNoClosedRoundTripsSaysSoRatherThanZero(t *testing.T) {
	h := newOverviewHarness(t)
	seedEngineJournal(t, h.journal, overviewTripsFixture)

	// The US market's fixture trip is today; KR's are too. Use the plain fixture,
	// whose trips all closed days ago, to get an empty day in both markets.
	empty := newOverviewHarness(t)
	seedJournal(t, empty.journal)

	for _, market := range []string{"KR", "US"} {
		row := todayRow(empty.overview(t.Context()).Today, market)
		if !row.NoTrades {
			t.Errorf("%s: a day with no closed round trips is not marked 거래 없음: %+v", market, row)
		}
		if row.PnL.Known() {
			t.Errorf("%s: an empty day reports a realised P&L of %q; it has none to report",
				market, row.PnL.Value())
		}
	}
	if !strings.Contains(empty.page(t, "/dashboard"), "거래 없음") {
		t.Error("the screen does not say 거래 없음 for a market that closed nothing today")
	}
}

// --- open orders (task 4.7) ----------------------------------------------------------

// TestTheOpenOrderCountIsUnmeasuredAndNotZero.
//
// This change's own rule, applied to this change. The order-reading seam lands in
// console-orders-screen; a placeholder 0 here would read as "미체결 없음" to
// somebody deciding whether it is safe to start the engine.
func TestTheOpenOrderCountIsUnmeasuredAndNotZero(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	count := h.overview(t.Context()).Open.Count
	if count.Known() {
		t.Errorf("the open-order count reports %q; no seam in this build can read it", count.Value())
	}
	if count.Code() != string(reasonSeamUnwired) {
		t.Errorf("the open-order count is unmeasured for %q, want %q", count.Code(), reasonSeamUnwired)
	}
	if !strings.Contains(h.page(t, "/dashboard"), string(reasonSeamUnwired)) {
		t.Error("the screen does not name seam_unwired beside the open-order count")
	}
}

// --- the safety panel (task 4.8) -----------------------------------------------------

// TestTheLeftoverPanelReadsBothMarkets.
//
// c.snapshot() is the KR reading and the two markets keep separate evidence
// files. A leftover panel that reads one market hides the other market's
// leftovers — which is the single thing this panel exists to show, and a leftover
// eats the next verification's exposure cap before any step can send anything.
func TestTheLeftoverPanelReadsBothMarkets(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	seedUSLeftover(t, usRecordPath(h.record))

	view := h.overview(t.Context())
	var found bool
	for _, m := range view.Safety.Markets {
		if m.Market == "US" && m.Leftovers() {
			found = true
		}
	}
	if !found {
		t.Fatalf("the US market's leftover is not on the panel: %+v", view.Safety.Markets)
	}

	page := h.page(t, "/dashboard")
	if !strings.Contains(page, usLeftoverID) {
		t.Error("the screen does not name the US leftover; a KR-only panel hides exactly this")
	}
	if !strings.Contains(page, "노출 상한") {
		t.Error("the screen does not say that a leftover fills the next verification's exposure cap")
	}
}

// --- Guardian (task 4.9) -------------------------------------------------------------

// TestOnlyOneGuardianAxisIsComputedAndTheRestSayWhyNot.
//
// OpenExposure and the per-order figures live in the engine's memory and in no
// table. A console that computed something plausible from the holdings would
// render it as a measured limit consumption, and a screen that makes an operator
// trust a protection that is not there is worse than a screen with no limits on
// it at all.
func TestOnlyOneGuardianAxisIsComputedAndTheRestSayWhyNot(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) {
		o.GateLimits = fakeGateLimits{limits: GateLimits{
			MaxOrderQuantity:   100,
			MaxOrderNotional:   5_000_000,
			MaxTotalExposure:   20_000_000,
			MaxDailyLossAmount: 1_000_000,
			MaxDailyLossRatio:  0.02,
			Currency:           "KRW",
		}}
	})
	seedEngineJournal(t, h.journal, overviewTripsFixture)

	guardian := h.overview(t.Context()).Safety.Guardian
	if !guardian.Wired {
		t.Fatal("the limits seam was wired and the panel says otherwise")
	}

	// WHICH axis, not how many. Counting alone passes when the computed one and a
	// structural one are swapped — and the swap that matters is exactly the one
	// this test exists to prevent: an open-exposure figure derived from the
	// holdings, rendered as a measured limit consumption.
	var computed, structural []string
	for _, axis := range guardian.Axes {
		if axis.Structural != "" {
			structural = append(structural, axis.Label)
			continue
		}
		computed = append(computed, axis.Label)
	}
	wantComputed := []string{"오늘 실현손익 대 일일 손실 한도(금액)"}
	wantStructural := []string{"계좌 개방 노출", "1회 주문 수량", "1회 주문 금액"}
	if !reflect.DeepEqual(computed, wantComputed) {
		t.Errorf("the computed axes are %v, want %v — the daily realised loss is the only one the "+
			"ledger holds a record of", computed, wantComputed)
	}
	if !reflect.DeepEqual(structural, wantStructural) {
		t.Errorf("the structurally-unmeasured axes are %v, want %v", structural, wantStructural)
	}
	for _, axis := range guardian.Axes {
		if axis.Structural == "" {
			continue
		}
		if len(axis.Rows) != 0 {
			t.Errorf("the structural axis %q carries %d row(s); it has no record to put in one",
				axis.Label, len(axis.Rows))
		}
	}

	for _, limit := range guardian.Limits {
		if !limit.Value.Known() {
			t.Errorf("the configured limit %q reads as unmeasured; it was in the config", limit.Label)
		}
	}
	page := h.page(t, "/dashboard")
	if !strings.Contains(page, "구조적으로 미측정") {
		t.Error("the screen does not say the other axes are structurally unmeasured")
	}
	if !strings.Contains(page, "1,000,000") {
		t.Error("the daily-loss ceiling is not on the screen")
	}
}

// TestAnUnreadableGateConfigIsNeitherZeroNorUnlimited.
func TestAnUnreadableGateConfigIsNeitherZeroNorUnlimited(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) {
		o.GateLimits = fakeGateLimits{err: errBrokerDown}
	})
	seedJournal(t, h.journal)

	guardian := h.overview(t.Context()).Safety.Guardian
	for _, limit := range guardian.Limits {
		if limit.Value.Known() {
			t.Errorf("%q reads %q from a config nobody could parse", limit.Label, limit.Value.Value())
		}
		if strings.TrimSpace(limit.Value.Why()) == "" {
			t.Errorf("%q is unmeasured with no reason beside it", limit.Label)
		}
		// Neither of the two lies available here: a ceiling of zero, and no
		// ceiling at all. The first would read as "nothing may be sent" and the
		// second as "anything may".
		for _, lie := range []string{"0", "무제한", "미설정"} {
			if limit.Value.Value() == lie {
				t.Errorf("%q renders %q for a limit that was never read", limit.Label, lie)
			}
		}
	}
	if !strings.Contains(h.page(t, "/dashboard"), "한도를 읽지 못했다") {
		t.Error("the screen does not say the limits were unread")
	}
}

// TestTheDailyLossAxisIsNotComparedAcrossCurrencies.
//
// max_daily_loss_amount is one number in one currency and today's realised P&L is
// a number per market. Dividing one by the other across a currency boundary
// produces a percentage of nothing, and a percentage is exactly what an operator
// would act on.
//
// Both fixtures matter, and the second one is the one that used to be missing.
// The currency test ran AFTER the no-trades test, so a US market that had closed
// nothing rendered "0 / 1,000,000" — a KRW ceiling opposite an invented USD zero,
// as a measured consumption, with no note. The old fixture gave US a trip so the
// branch was never reached, and the old assertion (no "소진" in the value) would
// have passed on "0 / 1,000,000" anyway. Two misses in one test, which is why the
// assertions below are on what the row renders rather than on one absent word.
func TestTheDailyLossAxisIsNotComparedAcrossCurrencies(t *testing.T) {
	krwCeiling := func(o *Options) {
		o.GateLimits = fakeGateLimits{limits: GateLimits{
			MaxDailyLossAmount: 1_000_000, Currency: "KRW",
		}}
	}

	for _, tc := range []struct {
		name string
		seed func(*testing.T, *overviewHarness)
		// wantUS is what the foreign row must render, in full.
		wantUS string
	}{
		{
			name:   "the foreign market traded",
			seed:   func(t *testing.T, h *overviewHarness) { seedEngineJournal(t, h.journal, overviewTripsFixture) },
			wantUS: "250.5 USD / 1,000,000 KRW",
		},
		{
			// The case the fixture never reached. A quiet foreign session is the
			// most likely state of this row and the least likely to be looked at.
			name:   "the foreign market closed nothing",
			seed:   func(t *testing.T, h *overviewHarness) { seedJournal(t, h.journal) },
			wantUS: "거래 없음 USD / 1,000,000 KRW",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newOverviewHarness(t, krwCeiling)
			tc.seed(t, h)

			axis := h.overview(t.Context()).Safety.Guardian.Axes[0]
			var us guardianAxisRow
			var found bool
			for _, row := range axis.Rows {
				if row.Market == "US" {
					us, found = row, true
				}
			}
			if !found {
				t.Fatalf("the axis has no US row: %+v", axis.Rows)
			}
			if us.Consumed.Value() != tc.wantUS {
				t.Errorf("the US row renders %q, want %q — a KRW ceiling and a USD figure are shown "+
					"side by side and nothing is divided by anything", us.Consumed.Value(), tc.wantUS)
			}
			if !strings.Contains(us.Note, "통화") {
				t.Errorf("the US row does not say why no ratio was computed: %q", us.Note)
			}
			// The two lies this row can tell: a ratio, and a zero standing in for a
			// figure nobody measured.
			if strings.Contains(us.Consumed.Value(), "소진") {
				t.Errorf("the US row reports a consumption of a KRW ceiling: %q", us.Consumed.Value())
			}
			if strings.HasPrefix(us.Consumed.Value(), "0 /") {
				t.Errorf("the US row renders %q; today's realised P&L for that market is not zero, "+
					"it is a number in another currency or no number at all", us.Consumed.Value())
			}
		})
	}
}

// TestAGateConfigNobodyCouldReadDoesNotClaimTheCurrencyIsUnset.
//
// guardianPanelFrom sets Wired before it calls the seam and returns from the
// error branch before assigning Currency, and the template rendered the currency
// line on {{if .Wired}} — so a config nobody could parse printed 한도 통화 (미지정).
// That is a claim about the file's contents, made by a branch that never read the
// file, and it is the exact collapse of "unread" into "unset" that GateLimits's
// own doc comment says this screen does not make.
func TestAGateConfigNobodyCouldReadDoesNotClaimTheCurrencyIsUnset(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) {
		o.GateLimits = fakeGateLimits{err: errBrokerDown}
	})
	seedJournal(t, h.journal)

	guardian := h.overview(t.Context()).Safety.Guardian
	if !guardian.Wired {
		t.Fatal("a seam was injected and the panel says otherwise")
	}
	if guardian.Read {
		t.Fatal("the seam returned an error and the panel reports the limits as read")
	}

	page := h.page(t, "/dashboard")
	if strings.Contains(page, "한도 통화 <strong>") || strings.Contains(page, "(미지정)") {
		t.Error("the screen states the limit currency for a config it could not read; 미지정 is what " +
			"the file says and this branch never opened the file")
	}
	if !strings.Contains(page, "미지정과는 다르다") {
		t.Error("the screen does not distinguish an unread currency from an unset one")
	}

	// And the reason carries a code now, like every other unmeasured cell: the one
	// whose fix is "edit a file" was the only one an operator could not grep for
	// (I-2, Manager ruling).
	for _, limit := range guardian.Limits {
		if limit.Value.Code() != string(reasonConfigUnreadable) {
			t.Errorf("%q is unmeasured for %q, want %q", limit.Label, limit.Value.Code(),
				reasonConfigUnreadable)
		}
		if !strings.Contains(limit.Value.Why(), errBrokerDown.Error()) {
			t.Errorf("%q lost the parse error the seam returned: %q", limit.Label, limit.Value.Why())
		}
	}
	if !strings.Contains(page, string(reasonConfigUnreadable)) {
		t.Error("the screen does not print the config_unreadable code beside the unread ceilings")
	}
}

// TestALedgerValueThatWillNotParseIsNotTheSameAsAnUnreadableLedger.
//
// journal_unreadable's advice is "start the engine so it migrates, or update the
// console". Neither does anything for a ledger that opened, answered, and handed
// back a closed_at that is not a timestamp — the row is what has to be looked at,
// and that is a seventh thing to tell somebody.
func TestALedgerValueThatWillNotParseIsNotTheSameAsAnUnreadableLedger(t *testing.T) {
	h := newOverviewHarness(t)
	seedEngineJournal(t, h.journal, overviewBrokenClosedAtFixture)

	view := h.overview(t.Context())
	if !view.Journal.Readable() {
		t.Fatal("the fixture was supposed to be a ledger that opens perfectly well")
	}
	row := todayRow(view.Today, "KR")
	if !row.Blocked {
		t.Fatalf("the KR row summed a day it could not date: %+v", row)
	}
	if row.Unreadable.Code() != string(reasonJournalValueUnparsable) {
		t.Errorf("an unparsable ledger value renders as %q, want %q — %q sends the operator to restart "+
			"an engine that would change nothing", row.Unreadable.Code(),
			reasonJournalValueUnparsable, reasonJournalUnreadable)
	}
	if !strings.Contains(row.Unreadable.Why(), "not-a-timestamp") {
		t.Errorf("the reason does not name the value that would not parse: %q", row.Unreadable.Why())
	}
	if !strings.Contains(h.page(t, "/dashboard"), string(reasonJournalValueUnparsable)) {
		t.Error("the screen does not print the code beside the row")
	}
}

// --- panel isolation (task 4.11) -----------------------------------------------------

// TestAnUnreadableJournalEmptiesOnlyItsOwnPanels.
//
// The operator opens this screen precisely when they do not know what is running,
// which is the moment a missing ledger is most likely and a blank page least
// useful.
func TestAnUnreadableJournalEmptiesOnlyItsOwnPanels(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) {
		o.GateLimits = fakeGateLimits{limits: GateLimits{MaxDailyLossAmount: 1_000_000, Currency: "KRW"}}
	})
	// No journal file at all.
	h.holdings.rows = brokerRows()
	h.page(t, "/positions")

	view := h.overview(t.Context())

	if view.Today.Journal.Readable() {
		t.Fatal("the fixture was supposed to have no journal")
	}
	for _, market := range []string{"KR", "US"} {
		row := todayRow(view.Today, market)
		if !row.Blocked || row.Unreadable.Code() != string(reasonJournalUnreadable) {
			t.Errorf("%s today reads %+v; an absent ledger is journal_unreadable, not zero", market, row)
		}
	}
	if !view.Recent.Blocked {
		t.Error("the recent panel does not report the ledger as unread")
	}

	// And the panels that do not need the ledger still work.
	kr := accountRow(view.Account, "KR")
	if !kr.Value.Known() {
		t.Errorf("the account panel lost its broker numbers to a missing ledger: %+v", kr)
	}
	if kr.Managed.Known() {
		t.Error("the managed split claims to be measured without a ledger")
	}
	if len(view.Safety.Markets) != 2 {
		t.Error("the safety panel stopped reading the evidence records")
	}
	for _, limit := range view.Safety.Guardian.Limits {
		if !limit.Value.Known() {
			t.Errorf("the Guardian limit %q was lost to a missing ledger", limit.Label)
		}
	}
	page := h.page(t, "/dashboard")
	if !strings.Contains(page, "개요") {
		t.Error("the screen did not render at all")
	}
}

// TestTheLedgerIsOpenedOncePerRenderAndTheNoticePrintsOnce.
//
// The account panel called livePositions and the today/recent panels called a
// second reader, so one render opened the journal twice and produced two
// journalViews. Two readings of a file the engine is writing to can disagree, and
// the screen then showed one panel's holdings split against another panel's
// 원장 미판독, with the ledger-state notice underneath both of them.
func TestTheLedgerIsOpenedOncePerRenderAndTheNoticePrintsOnce(t *testing.T) {
	code := strings.Join(nonCommentLines(packageFiles(t)["overview.go"]), "\n")
	if n := strings.Count(code, "c.openJournal("); n != 1 {
		t.Errorf("overview.go opens the journal %d times; one render is one handle and one verdict", n)
	}
	if strings.Contains(code, "c.livePositions(") {
		t.Error("overview.go calls livePositions, which opens the journal itself; that is the second " +
			"open the panels used to disagree across")
	}

	h := newOverviewHarness(t)
	// No journal at all: the state every panel has to agree about.
	page := h.page(t, "/dashboard")
	if n := strings.Count(page, "엔진 미가동 — 원장 파일이 아직 없다"); n != 1 {
		t.Errorf("the ledger-state notice is printed %d times, want 1", n)
	}

	view := h.overview(t.Context())
	if view.Account.Journal.State != view.Today.Journal.State ||
		view.Today.Journal.State != view.Recent.Journal.State {
		t.Errorf("the three panels disagree about the ledger: account=%q today=%q recent=%q",
			view.Account.Journal.State, view.Today.Journal.State, view.Recent.Journal.State)
	}
}

// TestAnUnwiredGateLimitsSeamOnlyDarkensItsOwnPanel (task 2.5).
func TestAnUnwiredGateLimitsSeamOnlyDarkensItsOwnPanel(t *testing.T) {
	h := newOverviewHarness(t) // no GateLimits
	seedEngineJournal(t, h.journal, overviewTripsFixture)

	view := h.overview(t.Context())
	if view.Safety.Guardian.Wired {
		t.Fatal("no seam was wired and the panel says one was")
	}
	for _, limit := range view.Safety.Guardian.Limits {
		if limit.Value.Code() != string(reasonSeamUnwired) {
			t.Errorf("%q is unmeasured for %q, want seam_unwired", limit.Label, limit.Value.Code())
		}
	}
	if row := todayRow(view.Today, "KR"); !row.Trips.Known() {
		t.Errorf("the today panel went dark because the limits seam is missing: %+v", row)
	}
	page := h.page(t, "/dashboard")
	if page == "" {
		t.Error("the screen did not render without the limits seam")
	}
	// The sentence that says the console does not judge the gate lived inside
	// {{if .Wired}}, so the build that shows no limits at all — the one where a
	// reader is most likely to wonder what the console thinks about the gate —
	// was the one build that never said it.
	if !strings.Contains(page, "게이트가 열려 있는지는 여기서 판정하지 않는다") {
		t.Error("an unwired build never states that the console does not decide whether the gate is " +
			"open; the sentence is about this console and not about the seam")
	}
	if strings.Contains(page, "한도 통화") {
		t.Error("an unwired build states something about the limit currency; no seam read any config")
	}
}

// --- recent activity (task 4.10) -----------------------------------------------------

// TestTheRecentPanelShowsExitJudgementsAndNoFills.
//
// journal.ReadOnly has no fills accessor, so a recent-fills strip would need the
// ledger's read surface widened — which is console-orders-screen's work and a
// different risk grade.
func TestTheRecentPanelShowsExitJudgementsAndNoFills(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	view := h.overview(t.Context())
	if len(view.Recent.Events) == 0 {
		t.Fatal("the fixture has an exit event and the panel shows none")
	}
	page := h.page(t, "/dashboard")
	if !strings.Contains(page, "PARTIAL_TAKE") {
		t.Error("the recent panel does not render the judgement the ledger holds")
	}
	if !strings.Contains(page, "체결은 이 화면에 없다") {
		t.Error("the screen does not state that fills are absent and why")
	}
}

// --- the engine panel (task 4.3) -----------------------------------------------------

// TestTheEnginePanelSaysARefusalReasonIsProcessLocal.
//
// c.engineNote is filled only when start or stop was pressed in THIS console
// process. A freshly opened overview has no reason, and that is correct rather
// than a gap — writing it down here stops the next reader hunting for a persisted
// reason that was never persisted.
func TestTheEnginePanelSaysARefusalReasonIsProcessLocal(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	page := h.page(t, "/dashboard")
	if !strings.Contains(page, "기동 거부 사유 없음") {
		t.Error("a freshly opened overview does not explain that a refusal reason is process-local")
	}
	if !strings.Contains(page, "엔진") {
		t.Error("the engine panel is missing entirely")
	}
}

// --- the shape of the screen ----------------------------------------------------------

// TestTheOverviewReloadsAtTheCacheTTL (task 4.12).
func TestTheOverviewReloadsAtTheCacheTTL(t *testing.T) {
	want := int(holdingsTTL / time.Second)
	if got := (overviewPage{}).RefreshSeconds(); got != want {
		t.Errorf("the overview reloads every %ds, want %ds derived from holdingsTTL", got, want)
	}
	// Refresh used to be a method on this type and is now a field on the embedded
	// chrome, set by the handler (change a054). The type can no longer answer
	// "does this screen reload" on its own, and the render below is the stronger
	// question anyway: a handler that forgets to set it produces exactly this
	// failure, which the method form could not have caught.
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	rendered := h.page(t, "/dashboard")
	if !strings.Contains(rendered, `http-equiv="refresh"`) {
		t.Error("the rendered page carries no meta refresh")
	}
	if !strings.Contains(rendered, `content="`+strconv.Itoa(want)+`"`) {
		t.Errorf("the rendered meta refresh is not %ds", want)
	}
}

// TestTheOverviewHasNoFormAndNoConfirmationInput (task 4.13).
//
// The screen is a 관측창. There is nothing on it to press, therefore nothing to
// confirm, therefore no typed string, no second click and no extra approval —
// 사용자 지시 2026-07-27, which a review suggestion has revived once before.
func TestTheOverviewHasNoFormAndNoConfirmationInput(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) {
		o.GateLimits = fakeGateLimits{limits: GateLimits{MaxDailyLossAmount: 1, Currency: "KRW"}}
	})
	seedJournal(t, h.journal)
	seedUSLeftover(t, usRecordPath(h.record))

	page := h.page(t, "/dashboard")
	for _, banned := range []string{
		"<form", "</form>", "method=\"post\"", "<input", "<button", "<textarea",
		"type=\"text\"", "confirm(", "onclick", "onchange", "csrf",
	} {
		if strings.Contains(strings.ToLower(page), strings.ToLower(banned)) {
			t.Errorf("the overview contains %q; this screen has no act on it and therefore nothing "+
				"to confirm", banned)
		}
	}
}

// TestTheTwoScreensAreNamedApartAndNeitherIsTheOther (tasks 4.1, 5.3; moved by
// a054).
//
// Two screens answering different questions must not share a name: "이 빌드를
// 믿어도 되나" is the verification console's question and "계좌가 지금 어떤가" is
// the overview's.
//
// This test used to also assert that the verification console stayed on the root
// path, under the heading "a tab watching an approval window must not be moved to
// another screen". a054 moved it, and the guard that claim was really making is
// kept below rather than dropped: a tab left on the old root DOES now land
// somewhere else, and what makes that safe is that the shared status strip
// surfaces a waiting approval — with a link to the screen that can take it — from
// wherever the operator ends up. Three windows were lost in one day to controls
// not being where the operator was looking (M11, M18); a redirect that hid one
// would be repeating that, and this is the assertion that stops it.
func TestTheTwoScreensAreNamedApartAndNeitherIsTheOther(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	resp := h.get(t, pathVerifyConsole)
	if resp.StatusCode != 200 {
		t.Fatalf("GET %s = %d, want 200", pathVerifyConsole, resp.StatusCode)
	}
	root := body(t, resp)
	if !strings.Contains(root, "실계좌 검증 진행") {
		t.Errorf("%s no longer renders the verification console", pathVerifyConsole)
	}
	if !strings.Contains(root, ">검증 콘솔<") || !strings.Contains(root, ">개요<") {
		t.Error("the navigation does not name the two screens apart")
	}

	overview := h.page(t, "/dashboard")
	if !strings.Contains(overview, "<h1>개요</h1>") {
		t.Error("/dashboard is not the overview")
	}
	if strings.Contains(overview, "실계좌 검증 진행") {
		t.Error("/dashboard has become a second copy of the verification console")
	}

	// A tab left on the old root follows the redirect to a different screen. It
	// must still be told an approval is waiting, and how to reach it.
	h.startAndWait(t)
	landed := body(t, h.get(t, "/"))
	if !strings.Contains(landed, `data-pending-approval="true"`) {
		t.Error("a tab redirected off the old root is not told an approval is waiting")
	}
	if !strings.Contains(landed, `<a href="`+pathVerifyConsole+`">`) {
		t.Error("the waiting approval is announced with no way to reach it")
	}
}

// --- the datatype of "we could not read it" (tasks 3.1, 3.2, 3.3) ---------------------

// TestTheUnmeasuredReasonsStayApart.
//
// Folding them into one boolean loses the only thing that makes an unmeasured
// value actionable: whether to wait, to fix something, to start the engine, to
// wire a seam, to open another screen, to edit a config file, or to go and look
// at the ledger row that will not parse.
//
// The list grew from five to seven and both additions were the same argument.
// I-2's ruling: the enumeration's job is to write down what the operator has to
// fix, without remainder, and a reason that exists only as a free sentence cannot
// be counted or tested for absence — so the next one gets written as a sentence
// too, and at that point the enumeration stops describing the surface. The
// config-read failure and the unparsable ledger value each send the operator
// somewhere none of the other codes does.
func TestTheUnmeasuredReasonsStayApart(t *testing.T) {
	all := []unmeasuredReason{
		reasonVerifySuspended, reasonBrokerReadFailed, reasonJournalUnreadable,
		reasonSeamUnwired, reasonNeverFetched, reasonConfigUnreadable,
		reasonJournalValueUnparsable,
	}
	if len(unmeasuredSentences) != len(all) {
		t.Errorf("%d reasons are worded, want %d", len(unmeasuredSentences), len(all))
	}
	seen := map[string]bool{}
	for _, r := range all {
		sentence, ok := unmeasuredSentences[r]
		if !ok || strings.TrimSpace(sentence) == "" {
			t.Errorf("%q has no sentence; an operator cannot act on a code", r)
			continue
		}
		if seen[sentence] {
			t.Errorf("%q shares its sentence with another reason; they ask for different things", r)
		}
		seen[sentence] = true

		got := unmeasuredFor(r)
		if got.Known() {
			t.Errorf("%q produced a measured reading", r)
		}
		if got.Code() != string(r) {
			t.Errorf("%q renders its code as %q", r, got.Code())
		}
	}
}

// TestNoUnmeasuredValueIsAReasonlessDash.
//
// A row that vanishes is not an unmeasured marker, it is concealment (D10 — the
// one thing deliberately not borrowed from StockOS, whose addDetail() drops a
// line when it has no value). A dash with no sentence beside it is the same
// failure with a placeholder.
func TestNoUnmeasuredValueIsAReasonlessDash(t *testing.T) {
	// The type cannot produce one, including from its zero value.
	for _, r := range []reading{
		{}, unmeasuredFor(reasonSeamUnwired), unmeasuredFor(unmeasuredReason("invented")),
		unmeasuredBecause(""), unmeasuredBecause("   "),
	} {
		if r.Known() {
			t.Errorf("%+v claims to be measured", r)
		}
		if strings.TrimSpace(r.Why()) == "" {
			t.Errorf("%+v renders an em dash with nothing beside it", r)
		}
	}

	// And no assembled screen produces one, in any of the states it can be in.
	for _, tc := range []struct {
		name string
		set  func(*testing.T, *overviewHarness)
	}{
		{"cold everything", func(*testing.T, *overviewHarness) {}},
		{"journal only", func(t *testing.T, h *overviewHarness) {
			seedEngineJournal(t, h.journal, overviewTripsFixture)
		}},
		{"broker failed", func(t *testing.T, h *overviewHarness) {
			seedJournal(t, h.journal)
			h.holdings.err = errBrokerDown
			h.page(t, "/positions")
		}},
		{"everything", func(t *testing.T, h *overviewHarness) {
			seedEngineJournal(t, h.journal, overviewTripsFixture)
			h.page(t, "/positions")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newOverviewHarness(t, func(o *Options) {
				o.GateLimits = fakeGateLimits{limits: GateLimits{MaxDailyLossAmount: 1_000_000, Currency: "KRW"}}
			})
			tc.set(t, h)
			for label, r := range allReadings(h.overview(t.Context())) {
				if !r.Known() && strings.TrimSpace(r.Why()) == "" {
					t.Errorf("%s is unmeasured with no reason beside it", label)
				}
			}
		})
	}
}

// allReadings collects every value on the screen, so the guard above cannot miss
// one by not knowing about it.
func allReadings(v overviewView) map[string]reading {
	out := map[string]reading{}
	for _, m := range v.Account.Markets {
		out["account/"+m.Market+"/unreadable"] = m.Unreadable
		out["account/"+m.Market+"/value"] = m.Value
		out["account/"+m.Market+"/pnl"] = m.PnL
		out["account/"+m.Market+"/symbols"] = m.Symbols
		out["account/"+m.Market+"/managed"] = m.Managed
		out["account/"+m.Market+"/unmanaged"] = m.Unmanaged
	}
	for _, m := range v.Today.Markets {
		out["today/"+m.Market+"/unreadable"] = m.Unreadable
		out["today/"+m.Market+"/trips"] = m.Trips
		out["today/"+m.Market+"/pnl"] = m.PnL
		out["today/"+m.Market+"/wins"] = m.Wins
		out["today/"+m.Market+"/losses"] = m.Losses
	}
	out["open/count"] = v.Open.Count
	out["recent/unreadable"] = v.Recent.Unreadable
	for _, l := range v.Safety.Guardian.Limits {
		out["guardian/limit/"+l.Label] = l.Value
	}
	for _, a := range v.Safety.Guardian.Axes {
		for _, r := range a.Rows {
			out["guardian/axis/"+a.Label+"/"+r.Market] = r.Consumed
		}
	}
	// Cells that were never filled because their row is blocked are the zero
	// reading, which is exactly the state this guard is about — but a blocked row
	// renders one reason instead of its cells, so those are dropped here rather
	// than reported as dashes the screen never draws.
	for _, m := range v.Account.Markets {
		if m.Blocked || m.Empty {
			for _, k := range []string{"value", "pnl", "symbols", "managed", "unmanaged"} {
				delete(out, "account/"+m.Market+"/"+k)
			}
		}
		if !m.Blocked {
			delete(out, "account/"+m.Market+"/unreadable")
		}
	}
	for _, m := range v.Today.Markets {
		if m.Blocked || m.NoTrades {
			for _, k := range []string{"trips", "pnl", "wins", "losses"} {
				delete(out, "today/"+m.Market+"/"+k)
			}
		}
		if !m.Blocked {
			delete(out, "today/"+m.Market+"/unreadable")
		}
	}
	if !v.Recent.Blocked {
		delete(out, "recent/unreadable")
	}
	return out
}

// --- helpers -------------------------------------------------------------------------

var errBrokerDown = context.DeadlineExceeded

// usHolding is one US position, so the account panel has both currencies in it.
func usHolding() domain.Position {
	return domain.Position{
		Symbol: "AAPL", Name: "Apple", MarketType: "US", Quantity: 10, AveragePrice: 250,
		CurrentPrice: 300, MarketValue: 3000, UnrealizedPnL: 500, ProfitRate: 0.2,
	}
}

// seedUSLeftover writes a US evidence record whose last run left an order live.
func seedUSLeftover(t *testing.T, path string) {
	t.Helper()
	rec, err := verifylive.OpenRecorder(path)
	if err != nil {
		t.Fatalf("OpenRecorder: %v", err)
	}
	defer rec.Close()

	now := overviewNow
	if err := rec.Append(verifylive.Entry{
		StepID: verifylive.StepOrderCancel, Title: "place and cancel",
		Verdict: verifylive.VerdictFail, Reason: "409 already-processing",
		AccountRef: attest.Mask("123-45-678901"), StartedAt: now, FinishedAt: now, Mutating: true,
		Artifacts: []verifylive.Artifact{{
			Kind: "order", ID: usLeftoverID, Symbol: "AAPL", CreatedAt: now,
		}},
	}); err != nil {
		t.Fatalf("Append: %v", err)
	}
}

// usLeftoverID is the identifier the screen has to print in full: an operator
// cancelling it by hand needs it.
const usLeftoverID = "US-STRANDED-1"
