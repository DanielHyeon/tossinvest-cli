package console

// status_strip_test.go pins the shared shell (change a054-console-status-shell).
//
// The claims here are the ones an operator would be misled by if they were
// wrong:
//
//	same place, same words   the four facts are in one row on every screen, not
//	                         re-worded per template as they were before.
//	unwired is not stopped   a build without the marker seam and a machine with
//	                         the engine down are different problems.
//	the time is this screen's the engine marker's heartbeat is not the age of the
//	                         numbers on the page.
//	no tone without fault    a hold the spec mandates is an explanation, not a
//	                         warning; a warning that is always on is unread.
//	the reload cannot lie    the strip's period and the meta tag are one value,
//	                         so a screen that stops reloading cannot look as
//	                         though it still does.

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

// consoleScreens is every screen that renders the shared shell.
//
// reload is the meta-refresh period the default harness must produce, and zero
// means the screen does not reload itself there. The two verification screens
// are zero in this table because nothing is running: their two seconds are a
// property of a working run, not of the screen.
//
// The root path is deliberately absent. It is a redirect and not a screen
// (change a054); TestTheRootPathAnswersWithTheOverview is what covers it.
var consoleScreens = []struct {
	path   string
	reload int
}{
	{pathVerifyConsole, 0},
	// The two protection-line screens reload on the engine's observation cadence
	// and not on the holdings TTL (change a080). The periods are separate constants
	// now because they answer separate questions — how often the broker may be
	// asked, and how late an operator may learn that a baseline moved.
	{"/dashboard", lineRefreshSeconds()},
	{"/positions", lineRefreshSeconds()},
	{"/orders", int(ordersTTL / time.Second)},
	{"/signals", int(candidate.DefaultWatchInterval / time.Second)},
	{"/history", 0},
	// The settings screen is four sub-screens (change a055). "/settings" itself is
	// absent for the reason the root path is: it is a redirect now, and
	// TestOpeningSettingsLandsOnTheDailyTab is what covers it.
	{pathSettingsStanding, 0},
	{pathSettingsDaily, 0},
	{pathSettingsStrategy, 0},
	{pathSettingsTools, 0},
	{"/optimization", 0},
	{"/performance-history", 0},
	{"/position-management", 0},
	{"/strategy-runtime", 0},
	{"/strategy-runtime/market-schedule", 0},
	{"/verify", 0},
	{"/report", 0},
}

// TestEveryScreenRendersTheSameStatusStrip (task 2.2).
func TestEveryScreenRendersTheSameStatusStrip(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		for _, want := range []string{
			`class="console-status"`, `aria-label="콘솔 상태"`,
			"<dt>엔진</dt>", "<dt>시장 세션</dt>", "<dt>데이터 시각</dt>", "<dt>자동 재로드</dt>",
			"data-engine-state=", "data-session-state=", "data-data-mode=", "data-reload=",
		} {
			if !strings.Contains(page, want) {
				t.Errorf("%s: the shared status strip is missing %q", screen.path, want)
			}
		}
		// The strip carries no form. Approving, starting and stopping stay on the
		// screens that own them; a control in the chrome would be a control on
		// every screen at once.
		strip := stripOf(t, screen.path, page)
		for _, forbidden := range []string{"<form", "<button", "<input"} {
			if strings.Contains(strip, forbidden) {
				t.Errorf("%s: the status strip contains %q; it is a reading, not a control",
					screen.path, forbidden)
			}
		}
	}
}

// stripOf cuts the status strip out of a rendered page.
func stripOf(t *testing.T, path, page string) string {
	t.Helper()
	i := strings.Index(page, `<dl class="console-status"`)
	if i < 0 {
		t.Fatalf("%s: no status strip in the rendered page", path)
	}
	j := strings.Index(page[i:], "</dl>")
	if j < 0 {
		t.Fatalf("%s: the status strip is not closed", path)
	}
	return page[i : i+j]
}

// cellOf returns the strip cell whose dd carries the given attribute.
//
// It cuts the strip out first. Searching the whole page finds the stylesheet —
// every one of these attributes is also a CSS selector — and a test that asserts
// against `.console-status [data-engine-state=stopped] { … }` passes whatever the
// markup says.
func cellOf(t *testing.T, page, attr string) string {
	t.Helper()
	strip := stripOf(t, "page", page)
	i := strings.Index(strip, attr)
	if i < 0 {
		t.Fatalf("no %s cell in the status strip", attr)
	}
	rest := strip[i:]
	if j := strings.Index(rest, "</dd>"); j >= 0 {
		return rest[:j]
	}
	return rest
}

// TestTheEngineCellSeparatesUnwiredFromStopped (task 2.2).
//
// The two were one word before the strip existed on most screens, and they send
// an operator to different places: one is a build that cannot see the engine,
// the other is a machine that is not running it.
func TestTheEngineCellSeparatesUnwiredFromStopped(t *testing.T) {
	unwired := newHarness(t)
	unwired.authenticate(t)
	cell := cellOf(t, body(t, unwired.get(t, "/dashboard")), `data-engine-state=`)
	if !strings.Contains(cell, `data-engine-state="unwired"`) || !strings.Contains(cell, "미배선") {
		t.Errorf("a build with no marker path does not say 미배선: %s", cell)
	}
	if strings.Contains(cell, "정지") {
		t.Errorf("a build with no marker path is reported as stopped: %s", cell)
	}

	wired := newEngineHarness(t)
	wired.authenticate(t)
	cell = cellOf(t, body(t, wired.get(t, "/dashboard")), `data-engine-state=`)
	if !strings.Contains(cell, `data-engine-state="stopped"`) || !strings.Contains(cell, "정지") {
		t.Errorf("a wired build with no live marker does not say 정지: %s", cell)
	}

	holdEngineMarker(t, wired.marker, engineNow)
	cell = cellOf(t, body(t, wired.get(t, "/dashboard")), `data-engine-state=`)
	if !strings.Contains(cell, `data-engine-state="running"`) || !strings.Contains(cell, "실행 중") {
		t.Errorf("a live marker does not say 실행 중: %s", cell)
	}
}

// TestEachScreenKeepsItsOwnReloadPeriod (task 2.1 ①).
//
// The shared shell supplies a default period of zero, and a screen's own method
// shadows it. Delete that method and nothing fails to compile — the screen just
// quietly stops updating. This table is what notices.
func TestEachScreenKeepsItsOwnReloadPeriod(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		want := `<meta http-equiv="refresh" content="` + strconv.Itoa(screen.reload) + `">`
		switch screen.reload {
		case 0:
			if strings.Contains(page, `http-equiv="refresh"`) {
				t.Errorf("%s reloads itself; this screen renders values nothing makes younger",
					screen.path)
			}
		default:
			if !strings.Contains(page, want) {
				t.Errorf("%s does not reload every %ds", screen.path, screen.reload)
			}
		}
	}
}

// TestTheReloadCellAndTheMetaTagAreOneFact (task 2.1 ②).
//
// Refresh moved onto the embedded chrome, so a screen that forgets to set it
// goes silent with no compile error and no test failure anywhere else. The strip
// would then be the only thing on the page still claiming a period. Making the
// cell read .Refresh directly is what stops the two from disagreeing, and this
// is the test that would catch it if somebody copied the period into Go instead.
func TestTheReloadCellAndTheMetaTagAreOneFact(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		cell := cellOf(t, page, "data-reload=")
		on := strings.Contains(page, `http-equiv="refresh"`)
		if on != strings.Contains(cell, `data-reload="on"`) {
			t.Errorf("%s: the meta refresh says on=%v and the strip disagrees: %s",
				screen.path, on, cell)
		}
		if !on && !strings.Contains(cell, "걸려 있지 않음") {
			t.Errorf("%s: a screen with no reload does not say so: %s", screen.path, cell)
		}
		if on {
			// Read the number out rather than asking whether the cell contains
			// "<n>초마다". Containment is blind by suffix: once a080 made the line
			// screens say 5초마다, a cell reading 15초마다 satisfied Contains for
			// screen.reload == 5, and a mutation pinning the template to the
			// signals period passed. Two of the five reloading screens differ only
			// in that leading digit.
			if got := reloadPeriodIn(t, screen.path, cell); got != screen.reload {
				t.Errorf("%s: the strip says %ds and the meta tag uses %ds: %s",
					screen.path, got, screen.reload, cell)
			}
		}
	}
}

// TestTheDataCellSaysWhichOfThreeThingsItMeans (task 2.3).
func TestTheDataCellSaysWhichOfThreeThingsItMeans(t *testing.T) {
	// ② a screen that reads its files while rendering. The render instant IS the
	// reading instant here, so printing it is not an invention.
	plain := newHarness(t)
	plain.authenticate(t)
	cell := cellOf(t, body(t, plain.get(t, pathSettingsDaily)), "data-data-mode=")
	if !strings.Contains(cell, `data-data-mode="on-request"`) || !strings.Contains(cell, "요청 시 읽음") {
		t.Errorf("a render-time screen does not say it read on request: %s", cell)
	}
	if !strings.Contains(cell, `data-data-tone=""`) {
		t.Errorf("a render-time reading carries a tone; it cannot be stale: %s", cell)
	}

	// ① a cache-backed screen, with its own recorded instant and age.
	cached := newDashboardHarness(t)
	cached.authenticate(t)
	cell = cellOf(t, cached.page(t, "/positions"), "data-data-mode=")
	if !strings.Contains(cell, `data-data-mode="cache"`) {
		t.Errorf("the positions screen does not report a cache reading: %s", cell)
	}
	if !strings.Contains(cell, "초 전)") {
		t.Errorf("the cache cell does not print the reading's age: %s", cell)
	}

	// ③ a read that failed. The reason is the fact; the last success is printed
	// only where there is one.
	failing := newDashboardHarness(t, func(o *Options) { o.Holdings = nil })
	failing.authenticate(t)
	cell = cellOf(t, failing.page(t, "/positions"), "data-data-mode=")
	if !strings.Contains(cell, `data-data-mode="unavailable"`) || !strings.Contains(cell, "읽지 못함") {
		t.Errorf("an unread screen does not say so: %s", cell)
	}
	if strings.Contains(cell, "초 전)") {
		t.Errorf("an unread screen printed an age for a reading it does not have: %s", cell)
	}
}

// TestTheFreshnessToneComesFromTheScreensOwnTTL (task 2.4).
//
// No new constant: the thresholds are the screen's own cache period, so a screen
// cannot warn about an age it itself considers current.
func TestTheFreshnessToneComesFromTheScreensOwnTTL(t *testing.T) {
	h := newDashboardHarness(t)
	h.authenticate(t)
	h.page(t, "/positions") // fills the cache at the frozen instant

	for _, tc := range []struct {
		name string
		age  time.Duration
		want string
	}{
		{"inside the TTL", holdingsTTL / 2, "ok"},
		{"past the TTL", holdingsTTL + time.Second, "warn"},
		{"past twice the TTL", 2*holdingsTTL + time.Second, "stale"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// The overview peeks and never refreshes, so its age is the one that can
			// actually cross a threshold while the screen stays open.
			h.clock.advance(tc.age)
			defer h.clock.advance(-tc.age)
			cell := cellOf(t, h.page(t, "/dashboard"), "data-data-mode=")
			if !strings.Contains(cell, `data-data-tone="`+tc.want+`"`) {
				t.Errorf("age %v is not graded %q: %s", tc.age, tc.want, cell)
			}
		})
	}
}

// TestAKnownRefreshHoldGetsAReasonInsteadOfATone (task 2.5).
//
// A verification running is the spec telling the caches to stand down, and the
// discovery reading advances on a tick this console neither causes nor observes.
// Grading either against a clock puts a warning on a system doing exactly what
// it was told to do, and a warning that is always on is one nobody reads.
func TestAKnownRefreshHoldGetsAReasonInsteadOfATone(t *testing.T) {
	// The discovery screen, with the store wired and a reading four hours old:
	// long past any tick interval, and still not a fault.
	h := newSignalsHarness(t, &stubSignals{reading: SignalsReading{
		At:      signalsNow.Add(-4 * time.Hour),
		Markets: []SignalsMarket{{Market: "KR", Why: "관측 없음"}},
	}})
	h.authenticate(t)
	cell := cellOf(t, body(t, h.get(t, "/signals")), "data-data-mode=")
	if strings.Contains(cell, `data-data-tone="warn"`) || strings.Contains(cell, `data-data-tone="stale"`) {
		t.Errorf("the discovery reading is graded against a tick this console cannot cause: %s", cell)
	}
	if !strings.Contains(cell, "발굴 스캔 tick") {
		t.Errorf("the discovery cell does not say what advances it: %s", cell)
	}

	// A verification in flight: the overview says why, and says nothing about
	// rot. The cache is old because the spec said to leave it alone.
	held := newDashboardHarness(t)
	held.authenticate(t)
	held.page(t, "/positions")
	held.startAndWait(t)
	held.clock.advance(3 * holdingsTTL)
	cell = cellOf(t, held.page(t, "/dashboard"), "data-data-mode=")
	if !strings.Contains(cell, "갱신 보류") {
		t.Errorf("a held refresh does not say it is held: %s", cell)
	}
	if strings.Contains(cell, `data-data-tone="stale"`) {
		t.Errorf("a refresh the spec mandates holding is reported as a fault: %s", cell)
	}
}

// TestAPendingApprovalIsVisibleFromEveryScreen (task 2.6).
//
// Three approval windows lapsed on 2026-07-27 because the controls were not
// where the operator was looking (measurements M11, M18). The window is short;
// the strip is on every screen; this is what makes finding it independent of
// which screen happens to be open.
func TestAPendingApprovalIsVisibleFromEveryScreen(t *testing.T) {
	h := newHarness(t)
	h.startAndWait(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		if !strings.Contains(page, `data-pending-approval="true"`) {
			t.Errorf("%s does not surface the waiting approval", screen.path)
		}
		if !strings.Contains(page, `<a href="`+pathVerifyConsole+`">`) {
			t.Errorf("%s surfaces the waiting approval with no link to the screen that approves it",
				screen.path)
		}
		i := strings.Index(page, `data-pending-approval="true"`)
		j := strings.Index(page[i:], "</p>")
		if j > 0 && strings.Contains(page[i:i+j], "<form") {
			t.Errorf("%s puts an approval form in the chrome; approving happens on the run's screen",
				screen.path)
		}
	}
}

// TestTheEngineMarkerTimeIsNotTheScreensDataTime (task 2.7).
//
// engineView.RefreshedAt is the marker's heartbeat. Printing it in the data cell
// would make "3초 전" a statement about a process's liveness while the reader
// takes it for the age of the numbers underneath.
func TestTheEngineMarkerTimeIsNotTheScreensDataTime(t *testing.T) {
	h := newEngineHarness(t)
	h.authenticate(t)
	holdEngineMarker(t, h.marker, engineNow)

	page := body(t, h.get(t, "/dashboard"))
	marker := renderInstant(engineNow)
	if cell := cellOf(t, page, "data-data-mode="); strings.Contains(cell, marker) {
		t.Errorf("the data cell prints the engine marker's instant %q: %s", marker, cell)
	}
}

// TestTheStatusStripAddsNoBrokerCall (task 2.8).
func TestTheStatusStripAddsNoBrokerCall(t *testing.T) {
	h := newDashboardHarness(t)
	h.authenticate(t)
	h.page(t, "/positions")
	after := h.holdings.count()

	// Every screen, twice, inside one TTL. The strip reads the engine marker and
	// stats the installed binary; neither is a broker.
	for range 2 {
		for _, screen := range consoleScreens {
			resp := h.get(t, screen.path)
			_ = body(t, resp)
		}
	}
	if got := h.holdings.count(); got != after {
		t.Errorf("the broker was called %d more times inside one TTL; the strip is spending "+
			"the budget it was supposed to only report on", got-after)
	}
}

// TestTheStripDoesNotBranchOnTheMarketHoursAdvisory (task 2.2, safety).
//
// The open/closed decision is the template's. static_test.go bans .Outside from
// every Go file but data.go and templates.go, and the reason survives this
// change: the window in which the broker accepts an order is unmeasured, so
// every Go read of this advisory is a place that rule has to go on being true.
func TestTheStripDoesNotBranchOnTheMarketHoursAdvisory(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	cell := cellOf(t, body(t, h.get(t, "/dashboard")), "data-session-state=")
	if !strings.Contains(cell, "advisory") || !strings.Contains(cell, "휴장일은 모르며") {
		t.Errorf("the session cell does not say it is advisory: %s", cell)
	}
	if !strings.Contains(cell, `data-session-state="open"`) &&
		!strings.Contains(cell, `data-session-state="closed"`) {
		t.Errorf("the session cell states no session: %s", cell)
	}
}

// unusedURLValues keeps the url import honest if the approval test is edited.
var _ = url.Values{}
