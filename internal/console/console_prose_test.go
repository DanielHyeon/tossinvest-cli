package console

// console_prose_test.go pins what may be folded and what may not (change
// a055-console-settings-cadence §5 and §6).
//
// # Why the never-fold list is checked by class and not by phrase
//
// The first draft of this enumerated eight Korean sentences and matched them as
// substrings. A check like that dies silently: reword one clause, and it stops
// finding anything and passes. Nobody learns that the guard is gone, because a
// guard that finds nothing looks exactly like a guard with nothing to find.
//
// So the rule is mechanical and about position: an element carrying .notice or
// .danger may not appear inside a disclosure. New warnings are protected
// automatically, and a rewording changes nothing.

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
)

// reloadingScreens are the screens with a meta refresh in the default harness.
// They are the ones a native fold does not work on.
var reloadingScreens = []string{"/positions", "/orders", "/signals"}

// TestNoWarningIsHiddenInsideADisclosure (task 5.2).
func TestNoWarningIsHiddenInsideADisclosure(t *testing.T) {
	for _, build := range []*harness{fullSettingsHarness(t), newHarness(t)} {
		build.authenticate(t)
		for _, screen := range consoleScreens {
			page := body(t, build.get(t, screen.path))
			for _, hidden := range warningsInsideDisclosures(page) {
				t.Errorf("%s folds a %s away. The eight things that may never be folded are "+
					"warnings, and a warning behind a triangle is a warning nobody reads:\n%.200s",
					screen.path, hidden.class, hidden.text)
			}
		}
	}
}

type foldedWarning struct{ class, text string }

var openDetails = regexp.MustCompile(`<details\b`)

// warningsInsideDisclosures finds .notice and .danger elements nested inside a
// <details>, counting nesting so an inner close does not end an outer element.
func warningsInsideDisclosures(page string) []foldedWarning {
	var found []foldedWarning
	depth := 0
	for i := 0; i < len(page); {
		switch {
		case openDetails.MatchString(page[i:min(len(page), i+9)]) && strings.HasPrefix(page[i:], "<details"):
			depth++
			i += len("<details")
		case strings.HasPrefix(page[i:], "</details>"):
			if depth > 0 {
				depth--
			}
			i += len("</details>")
		case depth > 0 && strings.HasPrefix(page[i:], `class="notice"`):
			found = append(found, foldedWarning{"notice", page[i:min(len(page), i+220)]})
			i += len(`class="notice"`)
		case depth > 0 && strings.HasPrefix(page[i:], `class="danger"`):
			found = append(found, foldedWarning{"danger", page[i:min(len(page), i+220)]})
			i += len(`class="danger"`)
		default:
			i++
		}
	}
	return found
}

// TestTheNestingAwareScanFindsAWarningInsideAnInnerDisclosure.
//
// The scan is the thing every claim above rests on, so it is measured rather
// than trusted: a depth counter that treated the first </details> as the end of
// everything would report clean on exactly the markup this console writes.
func TestTheNestingAwareScanFindsAWarningInsideAnInnerDisclosure(t *testing.T) {
	for _, tc := range []struct {
		name string
		page string
		want int
	}{
		{"clean", `<p class="notice">a</p><details><p class="muted">b</p></details>`, 0},
		{"folded", `<details><summary>s</summary><p class="notice">a</p></details>`, 1},
		{"nested", `<details><details><p class="danger">a</p></details></details>`, 1},
		{"after a close", `<details><p>x</p></details><p class="notice">a</p>`, 0},
		{"inner close does not end the outer",
			`<details><details></details><p class="notice">a</p></details>`, 1},
	} {
		if got := len(warningsInsideDisclosures(tc.page)); got != tc.want {
			t.Errorf("%s: found %d, want %d", tc.name, got, tc.want)
		}
	}
}

// TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay (tasks 5.3, 5.4).
func TestAReloadingScreenFoldsWithTheURLAndOffersNoOtherWay(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if !strings.Contains(page, `http-equiv="refresh"`) {
		t.Fatal("the positions screen stopped reloading; this test is about the screens that do")
	}
	if strings.Contains(page, `<details class="explain"`) {
		t.Error("a reloading screen has an explanatory <details>. Its triangle does not change " +
			"the URL, so it opens on the click and closes on the next tick — which is worse " +
			"than not folding at all")
	}
	if !strings.Contains(page, `class="explain-link"`) {
		t.Fatal("the positions screen has no URL-driven disclosure at all")
	}

	// Closed by default, open when the URL says so, and the link that opens it is
	// on the page rather than something a test had to know.
	if strings.Contains(page, "백그라운드 폴러는 없다") {
		t.Error("the fold renders open with no parameter asking for it")
	}
	href := hrefFor(t, page, "holdings-basis")
	opened := h.page(t, href)
	if !strings.Contains(opened, "백그라운드 폴러는 없다") {
		t.Errorf("GET %s does not open the disclosure it names", href)
	}

	// And it survives the reload, which is the whole point: the meta refresh
	// reloads the current URL, and the parameter is in it.
	if !strings.Contains(opened, `http-equiv="refresh"`) {
		t.Error("the opened screen stopped reloading")
	}
	again := h.page(t, href)
	if !strings.Contains(again, "백그라운드 폴러는 없다") {
		t.Error("the disclosure closed on the second render of the same URL")
	}
}

// hrefFor is the link that opens one named disclosure.
func hrefFor(t *testing.T, page, id string) string {
	t.Helper()
	i := strings.Index(page, `data-explain="`+id+`"`)
	if i < 0 {
		t.Fatalf("the page has no disclosure named %q", id)
	}
	m := regexp.MustCompile(`href="([^"]*)"`).FindStringSubmatch(page[i:])
	if m == nil {
		t.Fatalf("the disclosure %q has no link", id)
	}
	return strings.ReplaceAll(m[1], "&amp;", "&")
}

// TestOnlyOneDisclosureOpensAtATime (task 5.3).
func TestOnlyOneDisclosureOpensAtATime(t *testing.T) {
	h := newOrdersHarness(t, &countingOrders{})
	h.authenticate(t)
	page := body(t, h.get(t, "/orders?explain=open-basis"))

	if !strings.Contains(page, "101번째의 살아 있는 주문") {
		t.Error("?explain=open-basis did not open the disclosure it names")
	}
	if strings.Contains(page, "거래 이력 화면이 대신하지 못하는 정보다") {
		t.Error("?explain=open-basis opened a second disclosure as well")
	}
}

// TestAnUnknownFoldParameterIsIgnoredRatherThanRefused (task 5.5).
func TestAnUnknownFoldParameterIsIgnoredRatherThanRefused(t *testing.T) {
	h := newOrdersHarness(t, &countingOrders{})
	h.authenticate(t)

	resp := h.get(t, "/orders?explain=nothing-by-this-name")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("an unrecognised fold id answered %d; the only thing it can be wrong about is "+
			"which paragraph is visible", resp.StatusCode)
	}
	page := body(t, resp)
	for _, folded := range []string{"101번째의 살아 있는 주문", "거래 이력 화면이 대신하지 못하는 정보다"} {
		if strings.Contains(page, folded) {
			t.Errorf("an unknown id opened %q", folded)
		}
	}
}

// TestTheFoldParameterReachesNoJudgementOrRecord (task 5.5, second half).
//
// It is display only. The claim is made against a save: the same POST with and
// without the parameter has to produce the same write and the same answer.
func TestTheFoldParameterReachesNoJudgementOrRecord(t *testing.T) {
	seam := &fakeLimits{gate: fullLimits()}
	h := limitsHarness(t, seam)
	h.authenticate(t)

	h.post(t, "/settings/limits/preset", tierForm(h.csrf, "kr-small-live"))
	h.post(t, "/settings/limits/preset?explain=anything", tierForm(h.csrf, "kr-small-live"))

	writes := seam.writes()
	if len(writes) != 2 {
		t.Fatalf("writes = %d, want 2", len(writes))
	}
	if writes[0] != writes[1] {
		t.Errorf("the fold parameter changed what was written:\n%+v\n%+v", writes[0], writes[1])
	}
}

// TestANonReloadingScreenUsesNativeDisclosures (task 5.4).
//
// The settings screens have no meta refresh, so there is nothing for a fold to
// survive and a URL parameter would be state introduced for no reason.
func TestANonReloadingScreenUsesNativeDisclosures(t *testing.T) {
	h := fullSettingsHarness(t)
	h.authenticate(t)
	for _, path := range settingsTabPaths {
		page := body(t, h.get(t, path))
		if strings.Contains(page, `http-equiv="refresh"`) {
			t.Errorf("%s reloads itself; then its folds would have to be URL-driven", path)
		}
		if strings.Contains(page, `class="explain-link"`) {
			t.Errorf("%s uses a URL-driven fold on a screen that never reloads; that is state "+
				"in the address bar buying nothing", path)
		}
	}
	if !strings.Contains(body(t, h.get(t, pathSettingsStanding)), `<details class="explain">`) {
		t.Error("the standing tab folds nothing at all; the 근거 prose was supposed to go behind " +
			"a native disclosure")
	}
}

// TestTheConsoleSpeaksInOneRegister (task 5.6).
//
// One screen used 합쇼체 while every other used 해라체, which reads as a
// different person talking on every third screen.
func TestTheConsoleSpeaksInOneRegister(t *testing.T) {
	polite := regexp.MustCompile(`(습니다|입니다|합니다|됩니다|립니다|칩니다|세요|십시오)`)
	h := fullSettingsHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		for _, found := range polite.FindAllStringIndex(page, -1) {
			t.Errorf("%s speaks in 합쇼체 (%q); every other screen uses 해라체 and a console that "+
				"changes voice reads as a different program",
				screen.path, strings.TrimSpace(page[max(0, found[0]-30):found[1]]))
		}
	}
}

// tierForm is the preset form's body.
func tierForm(csrf, tier string) map[string][]string {
	return map[string][]string{"csrf": {csrf}, "tier": {tier}}
}
