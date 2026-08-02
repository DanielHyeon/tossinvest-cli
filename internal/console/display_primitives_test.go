package console

// display_primitives_test.go pins the console's display vocabulary (change
// a054-console-status-shell, §4).
//
// # What is NOT tested here, and why
//
// The spec sentence this section inherits talks about documentElement.scrollWidth
// at 375px. This repository has no test that measures a layout — no browser, no
// headless engine, nothing that could produce that number — and writing the
// sentence down anyway would add a SHALL that passes because nothing evaluates
// it. The automatic checks below are the four things a rendered string can
// actually answer, and the browser measurement is a one-time human task recorded
// as evidence (tasks.md 4.7).

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestEveryTableIsResponsiveOrScrolls (task 4.2).
//
// Two ways to be responsive and no third. .data-table becomes a card stack under
// 720px using each cell's data-label; a table whose cells have no data-label
// cannot use it, and those go in a scroll region instead — which contains the
// overflow inside the table rather than letting the page scroll sideways.
func TestEveryTableIsResponsiveOrScrolls(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		for _, offset := range tableOffsets(page) {
			tag := page[offset:min(offset+120, len(page))]
			if strings.Contains(strings.SplitN(tag, ">", 2)[0], "data-table") {
				continue
			}
			if inScrollRegion(page[:offset]) {
				continue
			}
			t.Errorf("%s renders a table that is neither .data-table nor inside a scroll region: %.80s",
				screen.path, tag)
		}
	}
}

// TestEveryTableInTheTemplatesIsResponsiveOrScrolls (task 4.2, the other half).
//
// The rendered check above can only see tables a harness actually produces, and
// most of this console's tables need a seam wired before they appear at all —
// removing a scroll wrapper from one of those changes nothing a request can
// observe. This reads the markup instead, so a table nobody has a fixture for is
// still covered.
func TestEveryTableInTheTemplatesIsResponsiveOrScrolls(t *testing.T) {
	sources := map[string]string{
		"pageTemplates":              pageTemplates,
		"portfolioTemplates":         portfolioTemplates,
		"settingsTemplates":          settingsTemplates,
		"overviewTemplates":          overviewTemplates,
		"ordersTemplates":            ordersTemplates,
		"signalsTemplates":           signalsTemplates,
		"optimizationTemplates":      optimizationTemplates,
		"positionPolicyTemplates":    positionPolicyTemplates,
		"openAPIOnboardingTemplates": openAPIOnboardingTemplates,
	}
	for name, source := range sources {
		for _, offset := range tableOffsets(source) {
			tag := source[offset:min(offset+120, len(source))]
			if strings.Contains(strings.SplitN(tag, ">", 2)[0], "data-table") {
				continue
			}
			if inScrollRegion(source[:offset]) {
				continue
			}
			t.Errorf("%s declares a table that is neither .data-table nor inside a scroll region: %.80s",
				name, tag)
		}
	}
}

func tableOffsets(page string) []int {
	var out []int
	for i := 0; ; {
		j := strings.Index(page[i:], "<table")
		if j < 0 {
			return out
		}
		out = append(out, i+j)
		i += j + len("<table")
	}
}

// inScrollRegion reports that the text immediately before a table opens a scroll
// wrapper — the established shape is the div and the table adjacent, so nothing
// but whitespace may sit between them.
func inScrollRegion(before string) bool {
	trimmed := strings.TrimRight(before, " \t\r\n")
	return strings.HasSuffix(trimmed, ">") &&
		strings.Contains(lastTag(trimmed), `class="table-scroll"`)
}

func lastTag(s string) string {
	i := strings.LastIndex(s, "<")
	if i < 0 {
		return ""
	}
	return s[i:]
}

// TestTheNarrowViewportConditionsHold (task 4.1).
func TestTheNarrowViewportConditionsHold(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		if !strings.Contains(page, `name="viewport"`) {
			t.Errorf("%s renders no viewport meta", screen.path)
		}
		if !strings.Contains(page, "@media (max-width: 720px)") {
			t.Errorf("%s applies no narrow-viewport rules", screen.path)
		}
		for _, wide := range fixedPixelWidths(page) {
			t.Errorf("%s fixes a width at %dpx, wider than the narrowest viewport this console "+
				"is read on", screen.path, wide)
		}
	}
}

// narrowestViewport is the width the responsive contract is written against.
const narrowestViewport = 375

// fixedPixelWidths finds width declarations in px that exceed the narrowest
// viewport. rem is left alone: it scales with the reader's text size, and the one
// large rem width in this stylesheet is inside a scroll region on purpose.
// The leading class excludes max-width, which is how a media query states a
// breakpoint rather than how an element states a size.
var pixelWidthPattern = regexp.MustCompile(`[;{\s](?:min-width|width)\s*:\s*(\d+)px`)

func fixedPixelWidths(page string) []int {
	var out []int
	for _, m := range pixelWidthPattern.FindAllStringSubmatch(page, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil && n > narrowestViewport {
			out = append(out, n)
		}
	}
	return out
}

// TestOneNameForOneStatusDisplay (task 4.3).
//
// The direction of the merge is the point. .status-pill was used eight times and
// is asserted by an existing test (strategy_runtime_test.go); .state-badge was
// used twice and by nothing. Keeping the rarer name would have meant rewriting
// the common one's every use and breaking a test that had nothing to do with this
// change.
func TestOneNameForOneStatusDisplay(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		if strings.Contains(page, "state-badge") {
			t.Errorf("%s still carries the retired status class", screen.path)
		}
	}
	// A rule, not a mention: the merge is recorded in a comment that names the
	// retired class so a reader can still grep for why it went away.
	if retiredStatusRule.MatchString(pageTemplates) {
		t.Error("the stylesheet still defines the retired status class")
	}
	if !strings.Contains(pageTemplates, ".status-pill") {
		t.Fatal("the surviving status class is gone from the stylesheet")
	}
}

// TestTheHeadingStepsAreDistinguishable (task 4.4).
//
// Body text and a section heading at the same size is a screen that does not say
// what is important on it.
func TestTheHeadingStepsAreDistinguishable(t *testing.T) {
	sizes := map[string]float64{}
	for _, level := range []string{"h1", "h2", "h3"} {
		size, ok := declaredFontSize(pageTemplates, level)
		if !ok {
			t.Errorf("%s declares no font size, so its step is whatever the browser picks", level)
			continue
		}
		sizes[level] = size
	}
	body, ok := declaredBodyFontSize(pageTemplates)
	if !ok {
		t.Fatal("the stylesheet declares no body font size")
	}
	sizes["body"] = body

	if !(sizes["h1"] > sizes["h2"] && sizes["h2"] > sizes["h3"] && sizes["h3"] > sizes["body"]) {
		t.Errorf("the heading steps are not distinguishable: %v", sizes)
	}
	if sizes["body"] < 15 {
		t.Errorf("body text is %vpx; below 15 the console stops being readable at arm's length",
			sizes["body"])
	}
}

var (
	retiredStatusRule = regexp.MustCompile(`\.state-badge\s*[{,]`)
	remSizePattern    = regexp.MustCompile(`font-size:\s*([0-9.]+)rem`)
	bodyFontPattern   = regexp.MustCompile(`font:\s*([0-9.]+)px/`)
)

// declaredFontSize reads a bare element rule's font-size, in px.
func declaredFontSize(sheet, element string) (float64, bool) {
	i := strings.Index(sheet, "\n"+element+" { ")
	if i < 0 {
		return 0, false
	}
	rest := sheet[i:]
	j := strings.Index(rest, "}")
	if j < 0 {
		return 0, false
	}
	m := remSizePattern.FindStringSubmatch(rest[:j])
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	return v * 16, err == nil
}

func declaredBodyFontSize(sheet string) (float64, bool) {
	m := bodyFontPattern.FindStringSubmatch(sheet)
	if m == nil {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[1], 64)
	return v, err == nil
}

// The console renders no inline event handler on any screen, and this is what
// holds that line.
//
// a054 could only pin a shrinking inventory. Three `onsubmit="return confirm(…)"`
// handlers were on the settings screen — one per Guardian tier preset, plus the
// two system-update forms — and none of them had ever run: the response CSP is
// default-src 'none' with no script-src, which blocks inline handlers, and the
// header is pinned byte for byte below. Each was a confirmation step that existed
// in the source and not in the browser.
//
// a055 removed all three, which is why the exemption below is gone rather than
// narrowed. What took their place is the 적용 후 preview, which is rendered by
// the server and therefore actually visible. The claim here is now the simple
// one: no screen has any.

// newSettingsWiredHarness renders the settings screen with its seams filled.
//
// The default harness does not: this console does not disable a form whose seam
// is missing, it declines to render the form at all, so an unwired settings
// screen has no handlers on it and a check run against one passes by finding
// nothing. The three handlers below only exist on a settings screen that works.
func newSettingsWiredHarness(t *testing.T) *dashboardHarness {
	t.Helper()
	return limitsHarness(t, &fakeLimits{})
}

// TestNoScreenSmugglesInAScript (task 4.6).
func TestNoScreenSmugglesInAScript(t *testing.T) {
	inlineHandler := regexp.MustCompile(`\son[a-z]+=`)
	h := newSettingsWiredHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		resp := h.get(t, screen.path)
		page := body(t, resp)
		for _, found := range inlineHandler.FindAllStringIndex(page, -1) {
			t.Errorf("%s renders the inline handler %q. Under this CSP it cannot run, which makes it "+
				"a step that exists in the source and not in the browser",
				screen.path, strings.TrimSpace(page[found[0]:found[1]]))
		}
		for _, forbidden := range []string{"<script", "javascript:"} {
			if strings.Contains(strings.ToLower(page), forbidden) {
				t.Errorf("%s renders %q", screen.path, forbidden)
			}
		}
		if got := resp.Header.Get("Content-Security-Policy"); got != consoleHTMLCSP {
			t.Errorf("%s answered with CSP %q", screen.path, got)
		}
		if strings.Contains(resp.Header.Get("Content-Security-Policy"), "script-src") {
			t.Errorf("%s permits scripts; the dead handlers above would stop being dead", screen.path)
		}
	}
}

// --- the overview's summary (§5) ---------------------------------------------

// TestTheOverviewAnswersFromTheTop (task 5.1).
//
// "무엇이 잘못됐는가" has to be answerable without scrolling, and every cell has
// to say where the detail behind it lives — a summary that cannot be followed is
// a number to worry about with nowhere to go.
func TestTheOverviewAnswersFromTheTop(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	page := h.page(t, "/dashboard")

	i := strings.Index(page, `class="console-summary"`)
	if i < 0 {
		t.Fatal("the overview renders no summary")
	}
	j := strings.Index(page[i:], "</dl>")
	summary := page[i : i+j]

	for _, want := range []string{
		"엔진", "계좌 보유", "오늘 실현손익", "미체결", "안전",
		`href="` + pathVerifyConsole + `"`, `href="/positions"`, `href="/history"`, `href="/orders"`,
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("the summary is missing %q", want)
		}
	}

	// Above the detail, not instead of it.
	for _, section := range []string{"<h2>계좌 보유", "<h2>오늘", "<h2>안전"} {
		k := strings.Index(page, section)
		if k < 0 {
			t.Errorf("the detail section %q was removed when the summary arrived", section)
			continue
		}
		if k < i {
			t.Errorf("the summary is below %q; it is supposed to be the first thing read", section)
		}
	}
}

// TestAnUnmeasuredSummaryCellSaysSo (task 5.2).
//
// Zero means "we looked and there was none". A cell nobody could read means
// something else entirely, and printing the first when the second is true is how
// an unreadable account comes to look like an empty one.
func TestAnUnmeasuredSummaryCellSaysSo(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) { o.Holdings = nil })
	seedJournal(t, h.journal)
	h.authenticate(t)
	page := h.page(t, "/dashboard")

	i := strings.Index(page, `class="console-summary"`)
	if i < 0 {
		t.Fatal("the overview renders no summary")
	}
	summary := page[i : i+strings.Index(page[i:], "</dl>")]
	if !strings.Contains(summary, "미측정") {
		t.Errorf("an unreadable holdings cache does not read as unmeasured in the summary:\n%s", summary)
	}
	if strings.Contains(summary, ">0<") {
		t.Errorf("the summary prints a zero for something nobody read:\n%s", summary)
	}
}

// TestTheSummaryCostsNoBrokerCall (task 5.3).
func TestTheSummaryCostsNoBrokerCall(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	before := h.holdings.count()
	for range 3 {
		h.page(t, "/dashboard")
	}
	if got := h.holdings.count(); got != before {
		t.Errorf("the overview made %d broker calls; it is the longest-lived tab in the console and "+
			"must not be the one that spends the budget", got-before)
	}
}
