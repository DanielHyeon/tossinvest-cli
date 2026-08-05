package console

// line_cadence_test.go covers change a080: the protection line is redrawn on the
// engine's cadence instead of the broker cache TTL.
//
// The change turns on one measured fact — that the cache, and not the reload
// period, is what bounds the broker call rate. If that is false the change is not
// allowed, so it is the first test here.

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling.
//
// Spec "rate budget 보호" states the guarantee as a ceiling: one open tab costs at
// most one holdings call per TTL. Before a080 the spec also required the reload
// period itself to be at least the TTL, which reads as though the period were what
// enforced the ceiling. It is not — holdingsCache.get refreshes only once its last
// attempt is a TTL old:
//
//	if !hold && (!c.attempted || now.Sub(c.tried) >= c.ttl) {
//		c.refreshLocked(ctx, now)
//	}
//
// so reloading six times as often reaches the cache six times as often and the
// broker exactly as often. This test asserts the ceiling directly, which is the
// property the budget actually needs.
func TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	const window = 4 * holdingsTTL
	reloads := int(window / lineRefreshInterval)

	for i := 0; i < reloads; i++ {
		h.page(t, "/positions")
		h.clock.advance(lineRefreshInterval)
	}

	// The first render fetches, then at most one more per TTL elapsed.
	want := int(window/holdingsTTL) + 1
	if got := h.holdings.count(); got > want {
		t.Errorf("%d reload(s) over %s made %d broker call(s), want at most %d — "+
			"the cache TTL and not the reload period must bound the rate",
			reloads, window, got, want)
	}
	if reloads <= want {
		t.Fatalf("the window reloads %d time(s) and permits %d call(s); this test cannot "+
			"distinguish a bounded rate from an unbounded one", reloads, want)
	}
}

// TestTheOverviewSpendsNothingHoweverOftenItReloads.
//
// The overview reads the cache with peek and makes no broker call at all (change
// console-operator-overview D4, printed on the screen itself). Redrawing it more
// often must not turn the console's longest-lived tab into a spender.
//
// The clock advance carries weight here rather than decorating the loop. get
// refreshes once its last attempt is a TTL old, so only a window longer than the
// TTL can tell peek from get — the window is checked below, and the closing
// assertion proves the same harness does count a call when a get-reading screen
// asks for one. Without that last step a counter wired to nothing would pass.
func TestTheOverviewSpendsNothingHoweverOftenItReloads(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)

	const redraws = 10
	if window := redraws * lineRefreshInterval; window <= holdingsTTL {
		t.Fatalf("the redraw window is %s and the cache TTL is %s; a get-reading screen "+
			"would not refresh inside it, so this test cannot tell peek from get",
			window, holdingsTTL)
	}
	for i := 0; i < redraws; i++ {
		h.page(t, "/dashboard")
		h.clock.advance(lineRefreshInterval)
	}

	if got := h.holdings.count(); got != 0 {
		t.Errorf("the overview made %d broker call(s) across %d reloads, want 0", got, redraws)
	}

	// The positions screen reads the same cache with get. One render of it must
	// move the counter the assertion above depends on.
	h.page(t, "/positions")
	if got := h.holdings.count(); got == 0 {
		t.Error("a get-reading screen made no broker call either; the counter this test " +
			"reads is not wired to the broker and the zero above means nothing")
	}
}

// TestTheRedrawCadenceIsTheEnginesAndNotTheCaches.
//
// The separation is the change. Asserting the number alone would still pass if
// somebody re-derived it from holdingsTTL, so the test names both facts.
func TestTheRedrawCadenceIsTheEnginesAndNotTheCaches(t *testing.T) {
	if lineRefreshInterval != 5*time.Second {
		t.Errorf("lineRefreshInterval = %s, want the engine's 5s observation period",
			lineRefreshInterval)
	}
	if lineRefreshInterval >= holdingsTTL {
		t.Errorf("the redraw cadence (%s) is not shorter than the cache TTL (%s); "+
			"then a080 changed nothing an operator can see", lineRefreshInterval, holdingsTTL)
	}
	for name, got := range map[string]int{
		"positions": (positionsPage{}).RefreshSeconds(),
		"overview":  (overviewPage{}).RefreshSeconds(),
	} {
		if got != lineRefreshSeconds() {
			t.Errorf("%s reloads every %ds, want the engine's %ds", name, got, lineRefreshSeconds())
		}
	}
}

// TestBothLineScreensRenderTheSameCadence.
//
// Two screens drawing one holding's protection line must not disagree about how
// current it is, which is the same principle that keeps their values joined at one
// boundary (decoratePositionRows).
func TestBothLineScreensRenderTheSameCadence(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	want := `<meta http-equiv="refresh" content="` + strconv.Itoa(lineRefreshSeconds()) + `">`
	for _, path := range []string{"/positions", "/dashboard"} {
		if page := h.page(t, path); !strings.Contains(page, want) {
			t.Errorf("%s does not reload at the engine cadence", path)
		}
	}
}

// reloadPeriodIn returns the period the status strip's reload cell names, in
// seconds.
//
// It lives here rather than beside cellOf in status_strip_test.go because a080 is
// why it exists: the strip's cell used to be checked with
// strings.Contains(cell, strconv.Itoa(want)+"초마다"), and containment is blind by
// suffix. Once this change made the two line screens say 5초마다, a cell reading
// the signals screen's 15초마다 satisfied that check for want == 5. Reading the
// number out and comparing it fails on the value instead.
func reloadPeriodIn(t *testing.T, path, cell string) int {
	t.Helper()
	i := strings.Index(cell, "초마다")
	if i < 0 {
		t.Fatalf("%s: the reload cell names no period: %s", path, cell)
	}
	digits := cell[:i]
	start := len(digits)
	for start > 0 && digits[start-1] >= '0' && digits[start-1] <= '9' {
		start--
	}
	if start == len(digits) {
		t.Fatalf("%s: 초마다 in the reload cell has no number before it: %s", path, cell)
	}
	n, err := strconv.Atoi(digits[start:])
	if err != nil {
		t.Fatalf("%s: the reload cell's period is not a number: %v", path, err)
	}
	return n
}

// TestTheLineScreensStillShipNoScript.
//
// a080's first design fetched a fragment with a hash-pinned script. That is not
// available here: "JavaScript, CSP nonce, `unsafe-inline` script, 외부 CSS/폰트
// 도입" is a stated Non-Goal of the change that built these screens
// (archive/2026-07-31-streamline-trading-views), and the operator-console spec
// holds it as a CSP regression scenario. This test states the constraint from
// a080's side so that a future attempt to speed the screens up finds the reason
// here rather than rediscovering it from three failing tests elsewhere.
func TestTheLineScreensStillShipNoScript(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	for _, path := range []string{"/positions", "/dashboard"} {
		resp := h.get(t, path)
		if page := strings.ToLower(body(t, resp)); strings.Contains(page, "<script") {
			t.Errorf("%s ships a script; a080 buys its cadence without one", path)
		}
		if got := resp.Header.Get("Content-Security-Policy"); got != consoleHTMLCSP {
			t.Errorf("%s relaxed the CSP to %q", path, got)
		}
	}
}
