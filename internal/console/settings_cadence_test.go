package console

// settings_cadence_test.go pins the navigation and the settings classification
// (change a055-console-settings-cadence §2 and §3).
//
// The defects being fixed:
//
//	twelve navigation items, two of which named a SECTION of a screen rather
//	than the screen, so the rest of that screen was reachable only by accident;
//
//	three screens with routes and no navigation entry at all, reachable only
//	from inside another screen;
//
//	one settings page ordered by the history of the changes that built it, so
//	nothing on it told the operator which of those settings could be undone.

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// fullSettingsHarness wires every settings seam at once.
//
// The default harness wires none, and an unwired card renders a reason instead
// of a form — so a check for "is this control on two tabs" run against it would
// find no controls anywhere and pass. Every claim about where a control lives
// has to be made against a console that actually has them.
func fullSettingsHarness(t *testing.T, tweak ...func(*Options)) *harness {
	t.Helper()
	updater := &fakeSystemUpdater{view: validUpdateView()}
	options := append([]func(*Options){func(o *Options) {
		o.Settings = &fakeSettings{}
		o.Limits = &fakeLimits{gate: fullLimits()}
		o.TradingPolicy = &fakeTrading{block: fullPolicy()}
		o.Gate = &fakeGate{}
		o.EngineBoot = &fakeAutostart{}
		o.SystemUpdater = updater
		o.ReleaseDownloader = &fakeReleaseDownloader{release: validSignedRelease()}
		o.ReleaseCandidateStager = updater
		o.AcquireUpdateEngineLock = func() (func(), error) { return func() {}, nil }
		o.CheckUpdateVerifyActivity = func() error { return nil }
		o.Relaunch = func(int) error { return nil }
	}}, tweak...)
	return newHarness(t, options...)
}

// --- §2 navigation ---------------------------------------------------------------

// TestTheNavigationSaysWhatEachScreenAnswers (task 2.1).
func TestTheNavigationSaysWhatEachScreenAnswers(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	bar := mainNav(t, body(t, h.get(t, "/dashboard")))

	items := navItems(bar)
	if len(items) != 6 {
		t.Errorf("the navigation has %d items, want the six the operating flow has "+
			"(발굴 → 발주 → 보유 → 종결 → 설정):\n%v", len(items), items)
	}
	for _, item := range items {
		if !strings.Contains(item, "<small>") {
			t.Errorf("navigation item %q carries no description; the labels alone do not "+
				"separate 검증 콘솔 from 검증, 이력 from 성과 이력, or 포지션 from 포지션 정책", item)
		}
	}
	// The order is the engine's, and it is the whole reason six items are easier
	// than twelve: an order that can be learned is an order nobody has to remember.
	for i, want := range []string{"개요", "신호", "주문", "포지션", "이력", "설정"} {
		if i < len(items) && !strings.Contains(items[i], ">"+want+"<") {
			t.Errorf("navigation item %d is %q, want %q", i, items[i], want)
		}
	}
	// The current screen is marked, for a reader who cannot see the underline.
	if !strings.Contains(bar, `aria-current="page"`) {
		t.Error("the navigation does not mark the current screen")
	}
}

// mainNav is the top navigation bar, by its label.
func mainNav(t *testing.T, page string) string {
	t.Helper()
	i := strings.Index(page, `<nav aria-label="주요 화면">`)
	if i < 0 {
		t.Fatal("the page renders no top navigation")
	}
	j := strings.Index(page[i:], "</nav>")
	if j < 0 {
		t.Fatal("the top navigation is not closed")
	}
	return page[i : i+j]
}

var navAnchor = regexp.MustCompile(`(?s)<a href="[^"]*".*?</a>`)

func navItems(bar string) []string { return navAnchor.FindAllString(bar, -1) }

// TestNoScreenIsReachableOnlyFromInsideAnother (task 2.2).
//
// Three screens were in that position: /performance-history, /strategy-runtime
// and /strategy-runtime/market-schedule. The check is reachability from the
// navigation in at most two hops — a top-level item, or an entry point on one of
// its sub-screens — because that is exactly what the requirement says.
func TestNoScreenIsReachableOnlyFromInsideAnother(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	// The structure the requirement describes, and not a general crawl: a
	// top-level item, or an explicit entry point on one of that item's
	// sub-screens. Only the settings tabs are expanded as sub-screens, so a
	// screen cannot become "reachable" by being linked from some unrelated row
	// three clicks away.
	reachable := map[string]bool{}
	var topLevel []string
	for _, href := range hrefsIn(mainNav(t, body(t, h.get(t, "/dashboard")))) {
		reachable[href] = true
		topLevel = append(topLevel, href)
	}
	for _, href := range topLevel {
		for _, next := range hrefsIn(body(t, h.get(t, href))) {
			reachable[next] = true
		}
	}
	for _, tab := range settingsTabPaths {
		if !reachable[tab] {
			t.Errorf("the settings tab %s is not linked from the navigation item", tab)
			continue
		}
		for _, next := range hrefsIn(body(t, h.get(t, tab))) {
			reachable[next] = true
		}
	}

	for _, screen := range consoleScreens {
		if reachable[screen.path] {
			continue
		}
		t.Errorf("%s is not reachable from the navigation in two hops; a screen you can only "+
			"find by already being on another screen is a screen nobody finds", screen.path)
	}
}

var hrefPattern = regexp.MustCompile(`href="(/[^"]*)"`)

// hrefsIn is every same-origin link in a fragment, with the fragment identifier
// and the query dropped: they are the same screen.
func hrefsIn(fragment string) []string {
	var out []string
	for _, m := range hrefPattern.FindAllStringSubmatch(fragment, -1) {
		path := m[1]
		if i := strings.IndexAny(path, "#?"); i >= 0 {
			path = path[:i]
		}
		if path != "" {
			out = append(out, path)
		}
	}
	return out
}

// TestTheOldPathsStillRender (task 2.3).
//
// An item leaving the navigation and a route disappearing are different things,
// and only the first one happened. Documents, bookmarks and printed links point
// at all of these.
func TestTheOldPathsStillRender(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, path := range []string{
		pathVerifyConsole, "/optimization", "/performance-history", "/position-management",
		"/strategy-runtime", "/strategy-runtime/market-schedule", "/verify", "/report",
	} {
		if got := h.get(t, path).StatusCode; got != http.StatusOK {
			t.Errorf("GET %s = %d after leaving the navigation; the route was supposed to stay", path, got)
		}
	}
}

// --- §3 the four tabs -------------------------------------------------------------

// settingsTabPaths is the four sub-screens, in bar order.
var settingsTabPaths = []string{
	pathSettingsStanding, pathSettingsDaily, pathSettingsStrategy, pathSettingsTools,
}

// TestTheFourSettingsTabsAreRegisteredGetRoutes (task 3.1).
func TestTheFourSettingsTabsAreRegisteredGetRoutes(t *testing.T) {
	seen := map[string]route{}
	for _, r := range registeredRoutes(t) {
		seen[r.Path] = r
	}
	for _, path := range settingsTabPaths {
		r, ok := seen[path]
		if !ok {
			t.Errorf("%s is not registered", path)
			continue
		}
		if !r.Session {
			t.Errorf("%s is registered without session0", path)
		}
		if r.CSRFGated {
			t.Errorf("%s is behind the CSRF gate; a read route that demands a POST is a page "+
				"nobody can open", path)
		}
	}
}

// TestTheTabsAddNoStateChangingRoute (task 3.1, second half).
//
// The classification moves where a form is rendered. It must not add a way to
// change something — the console's list of state-changing acts is a safety
// invariant, and four new GET screens are not an addition to it.
func TestTheTabsAddNoStateChangingRoute(t *testing.T) {
	tabs := map[string]bool{}
	for _, path := range settingsTabPaths {
		tabs[path] = true
	}
	for _, path := range consoleStateChanging {
		if tabs[path] {
			t.Errorf("%s is a settings tab and is in the state-changing list", path)
		}
	}
	for _, finding := range routeTableFindings(registeredRoutes(t)) {
		t.Error(finding)
	}
}

// TestOpeningSettingsLandsOnTheDailyTab (task 3.2).
func TestOpeningSettingsLandsOnTheDailyTab(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	resp := h.noRedirect(t, pathSettings)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("GET %s = %d, want 303", pathSettings, resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != pathSettingsDaily {
		t.Errorf("GET %s redirects to %q; the default entry is the reversible, daily screen — "+
			"the fewest clicks belong to the one opened most", pathSettings, got)
	}
}

// TestTheOldAdoptionLinkLandsOnTheStandingTab (task 3.2).
//
// Two halves, because a fragment never reaches a server. `?section=adoption` is
// the half the server can answer; the other half is that the daily tab — where
// the browser carries a bare `#adoption` — has something at that anchor saying
// where the section went.
func TestTheOldAdoptionLinkLandsOnTheStandingTab(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	resp := h.noRedirect(t, pathSettings+"?section=adoption")
	if got := resp.Header.Get("Location"); got != pathSettingsStanding+"#adoption" {
		t.Errorf("?section=adoption redirects to %q, want %q",
			got, pathSettingsStanding+"#adoption")
	}

	daily := body(t, h.get(t, pathSettingsDaily))
	anchor := strings.Index(daily, `id="adoption"`)
	if anchor < 0 {
		t.Fatal("the default tab has no #adoption anchor; a bookmark to /settings#adoption lands " +
			"on it and finds nothing")
	}
	if !strings.Contains(daily[anchor:anchor+400], pathSettingsStanding+"#adoption") {
		t.Error("the #adoption anchor does not say where the section went")
	}
}

// TestEachSettingControlAppearsOnExactlyOneTab (task 3.3).
//
// A control is a named input inside a form. The same control on two tabs means
// the operator has to work out which one is real; the same VALUE on two tabs,
// read-only, is required rather than banned — TestATabShowsAnotherTabsValueAndLinksToIt
// is the other half.
func TestEachSettingControlAppearsOnExactlyOneTab(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)

	owner := map[string]string{}
	for _, path := range settingsTabPaths {
		for control := range controlsOn(body(t, h.get(t, path))) {
			if first, seen := owner[control]; seen {
				t.Errorf("the control %s is on both %s and %s; one of them is not the real one",
					control, first, path)
				continue
			}
			owner[control] = path
		}
	}
	if len(owner) == 0 {
		t.Fatal("no settings control was found at all; the scan is not reading the forms")
	}
}

var (
	formBlock = regexp.MustCompile(`(?s)<form[^>]*action="([^"]*)"[^>]*>(.*?)</form>`)
	inputName = regexp.MustCompile(`name="([^"]*)"`)
)

// controlsOn is every (action, field) pair the page submits, minus the CSRF
// token, which every form carries by design.
func controlsOn(page string) map[string]bool {
	out := map[string]bool{}
	for _, form := range formBlock.FindAllStringSubmatch(page, -1) {
		for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {
			if field[1] == "csrf" || field[1] == "tier" {
				continue
			}
			out[form[1]+" · "+field[1]] = true
		}
	}
	return out
}

// TestATabShowsAnotherTabsValueAndLinksToIt (task 3.3, second half).
func TestATabShowsAnotherTabsValueAndLinksToIt(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	daily := body(t, h.get(t, pathSettingsDaily))

	if !strings.Contains(daily, "automation gate") {
		t.Error("the daily tab does not say whether the gate is on; a ceiling means nothing " +
			"without knowing whether the thing it constrains is running")
	}
	if !strings.Contains(daily, pathSettingsStanding+"#gate") {
		t.Error("the daily tab shows the gate state with no link to the tab that has the switch")
	}
	if strings.Contains(daily, `action="/settings/gate"`) {
		t.Error("the gate switch is duplicated onto the daily tab")
	}
}

// TestEachTabHeaderStatesTheOperatingState (task 3.4).
func TestEachTabHeaderStatesTheOperatingState(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	for _, path := range settingsTabPaths {
		page := body(t, h.get(t, path))
		if !strings.Contains(page, `<h1>설정 <span class="muted">`) {
			t.Errorf("%s has no title naming which tab it is", path)
		}
		if !strings.Contains(page, `class="page-intro"`) {
			t.Errorf("%s has no one-line description of what belongs here", path)
		}
		state := blockOf(t, page, `class="console-summary settings-state"`, "</dl>")
		for _, want := range []string{"자동화 게이트", "엔진", "Guardian 한도"} {
			if !strings.Contains(state, want) {
				t.Errorf("%s: the header does not state %q:\n%s", path, want, state)
			}
		}
	}
}

// TestTheEntryPointsCarryTheTargetsCurrentLine (task 3.5).
func TestTheEntryPointsCarryTheTargetsCurrentLine(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)

	for path, wantLinks := range map[string][]string{
		pathSettingsStrategy: {"/optimization", "/position-management", "/strategy-runtime"},
		pathSettingsTools:    {pathVerifyConsole, "/verify", "/report"},
	} {
		page := body(t, h.get(t, path))
		for _, href := range wantLinks {
			if !strings.Contains(page, `href="`+href+`"`) {
				t.Errorf("%s does not link %s", path, href)
			}
		}
		summaries := regexp.MustCompile(`data-entry-summary="[^"]*">([^<]*)<`).FindAllStringSubmatch(page, -1)
		if len(summaries) != len(wantLinks) {
			t.Errorf("%s has %d entry summaries for %d links; a tab of bare links is one more "+
				"empty screen", path, len(summaries), len(wantLinks))
		}
		for _, summary := range summaries {
			if strings.TrimSpace(summary[1]) == "" {
				t.Errorf("%s has an entry point with an empty summary", path)
			}
		}
	}
}

// TestEverySettingsPostNamesTheFormItReturnsTo (task 3.6 / §4 ④).
//
// The table is read against the real route table, so a settings POST added
// without an entry fails here rather than silently answering on the default tab.
func TestEverySettingsPostNamesTheFormItReturnsTo(t *testing.T) {
	tabs := map[string]bool{}
	for _, path := range settingsTabPaths {
		tabs[path] = true
	}
	for _, r := range registeredRoutes(t) {
		if !strings.HasPrefix(r.Path, "/settings/") || tabs[r.Path] || !r.CSRFGated {
			continue
		}
		origin, ok := settingsOrigins[r.Path]
		if !ok {
			t.Errorf("%s saves something and names no card to answer beside; its result would "+
				"appear at the top of a screen with several forms on it", r.Path)
			continue
		}
		if !tabs[origin.Tab] {
			t.Errorf("%s returns to %q, which is not a settings tab", r.Path, origin.Tab)
		}
		if origin.Form == "" {
			t.Errorf("%s names no card", r.Path)
		}
	}
}

// TestTheSubmitContractIsUnchanged (task 3.6).
//
// The classification changes where a form is RENDERED. Every assertion here is
// about what a form still SENDS, which this change does not touch.
func TestTheSubmitContractIsUnchanged(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)

	fields := map[string]map[string]bool{}
	for _, path := range settingsTabPaths {
		for _, form := range formBlock.FindAllStringSubmatch(body(t, h.get(t, path)), -1) {
			if fields[form[1]] == nil {
				fields[form[1]] = map[string]bool{}
			}
			for _, field := range inputName.FindAllStringSubmatch(form[2], -1) {
				fields[form[1]][field[1]] = true
			}
		}
	}
	for action, want := range map[string][]string{
		"/settings/save":                   {"csrf", "enabled", "default_stop_percent", "exclude_symbols", "include_symbols"},
		"/settings/limits":                 {"csrf", "max_order_quantity", "max_order_notional", "max_total_exposure", "max_daily_loss_amount", "max_daily_loss_ratio", "limit_currency"},
		"/settings/limits/preset":          {"csrf", "tier"},
		"/settings/trading":                {"csrf", "place", "sell", "cancel", "allow_live_order_actions"},
		"/settings/gate":                   {"csrf", "enabled"},
		"/settings/autostart":              {"csrf", "enabled"},
		"/settings/system-update/install":  {"csrf", "reviewed_sha256"},
		"/settings/system-update/download": {"csrf"},
	} {
		got, ok := fields[action]
		if !ok {
			t.Errorf("no tab renders a form posting to %s", action)
			continue
		}
		for _, field := range want {
			if !got[field] {
				t.Errorf("%s no longer sends %q", action, field)
			}
		}
		if len(got) != len(want) {
			t.Errorf("%s sends %d fields, want %d: %v", action, len(got), len(want), got)
		}
	}
}

// TestASaveResultComesBackToTheFormThatCausedIt (§4 ④).
func TestASaveResultComesBackToTheFormThatCausedIt(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)

	resp := h.postNoRedirect(t, "/settings/trading", url.Values{"csrf": {h.csrf}, "sell": {"on"}})
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, pathSettingsDaily) {
		t.Errorf("a trading save lands on %q, want the tab that has the trading card", location)
	}
	for _, want := range []string{"form=trading", "#trading"} {
		if !strings.Contains(location, want) {
			t.Errorf("the redirect %q does not carry %q", location, want)
		}
	}

	page := body(t, h.post(t, "/settings/trading", url.Values{"csrf": {h.csrf}, "sell": {"on"}}))
	i := strings.Index(page, `data-save-result="trading"`)
	if i < 0 {
		t.Fatal("the save result is not rendered beside the card that produced it")
	}
	card := strings.LastIndex(page[:i], `data-settings-card=`)
	if card < 0 || !strings.Contains(page[card:i], `"trading"`) {
		t.Error("the save result is rendered outside the trading card")
	}
}

// TestAnUnmappedSettingsPostStillLandsOnASettingsTab (§4 ④, the fallback).
//
// An answer displayed beside its form is better than one at the top of the page,
// and one at the top of the page is much better than one that is lost.
func TestAnUnmappedSettingsPostStillLandsOnASettingsTab(t *testing.T) {
	origin := settingsOriginFor("/settings/something-nobody-mapped")
	if origin.Tab != pathSettingsDaily {
		t.Errorf("an unmapped save returns to %q, want the default tab", origin.Tab)
	}
	if origin.Form != "" {
		t.Errorf("an unmapped save claims to belong to card %q", origin.Form)
	}
}

// blockOf cuts out the fragment starting at marker and ending at terminator.
func blockOf(t *testing.T, page, marker, terminator string) string {
	t.Helper()
	i := strings.Index(page, marker)
	if i < 0 {
		t.Fatalf("the page has no %q", marker)
	}
	j := strings.Index(page[i:], terminator)
	if j < 0 {
		t.Fatalf("the block at %q is not closed", marker)
	}
	return page[i : i+j]
}
