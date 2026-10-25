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

	"golang.org/x/net/html"
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
// what is important on it. Sizes are fluid (clamp, change a172), so the order is
// checked at both ends of the range: the narrowest and the widest viewport.
func TestTheHeadingStepsAreDistinguishable(t *testing.T) {
	lo, hi := map[string]float64{}, map[string]float64{}
	for _, level := range []string{"h1", "h2", "h3"} {
		min, max, ok := declaredFontSize(pageTemplates, level)
		if !ok {
			t.Errorf("%s declares no font size, so its step is whatever the browser picks", level)
			continue
		}
		lo[level], hi[level] = min, max
	}
	bodyLo, bodyHi, ok := declaredBodyFontSize(pageTemplates)
	if !ok {
		t.Fatal("the stylesheet declares no body font size")
	}
	lo["body"], hi["body"] = bodyLo, bodyHi

	for name, sizes := range map[string]map[string]float64{"narrowest": lo, "widest": hi} {
		if !(sizes["h1"] > sizes["h2"] && sizes["h2"] > sizes["h3"] && sizes["h3"] > sizes["body"]) {
			t.Errorf("the heading steps are not distinguishable at the %s viewport: %v", name, sizes)
		}
	}
	if lo["body"] < 15 {
		t.Errorf("body text is %vpx at its smallest; below 15 the console stops being readable at arm's length",
			lo["body"])
	}
}

var (
	retiredStatusRule = regexp.MustCompile(`\.state-badge\s*[{,]`)
	// 크기는 `font-size: var(--text-N)` 로 모듈러 스케일 토큰을 읽고, 토큰은 rem clamp(min, 선호, max) 이다.
	// 평범한 rem 값도 그대로 읽는다.
	varSizePattern   = regexp.MustCompile(`font(?:-size)?:\s*var\((--text-[a-z0-9-]+)\)`)
	remSizePattern   = regexp.MustCompile(`font-size:\s*([0-9.]+)rem`)
	clampSizePattern = regexp.MustCompile(`clamp\(\s*([0-9.]+)rem\s*,[^;]*?,\s*([0-9.]+)rem\s*\)`)
)

// tokenSize reads `--name: clamp(min, preferred, max)` from the stylesheet, in px.
func tokenSize(sheet, name string) (lo, hi float64, ok bool) {
	m := regexp.MustCompile(regexp.QuoteMeta(name) + `:\s*` + clampSizePattern.String()).FindStringSubmatch(sheet)
	if m == nil {
		return 0, 0, false
	}
	return remRange(m[1], m[2])
}

// declaredFontSize reads a bare element rule's font-size range, in px.
func declaredFontSize(sheet, element string) (lo, hi float64, ok bool) {
	i := strings.Index(sheet, "\n"+element+" {")
	if i < 0 {
		return 0, 0, false
	}
	rest := sheet[i:]
	j := strings.Index(rest, "}")
	if j < 0 {
		return 0, 0, false
	}
	rule := rest[:j]
	if m := varSizePattern.FindStringSubmatch(rule); m != nil {
		return tokenSize(sheet, m[1])
	}
	if m := remSizePattern.FindStringSubmatch(rule); m != nil {
		return remRange(m[1], m[1])
	}
	return 0, 0, false
}

func declaredBodyFontSize(sheet string) (lo, hi float64, ok bool) {
	return declaredFontSize(sheet, "body")
}

func remRange(a, b string) (lo, hi float64, ok bool) {
	x, errA := strconv.ParseFloat(a, 64)
	y, errB := strconv.ParseFloat(b, 64)
	return x * 16, y * 16, errA == nil && errB == nil
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

// TestNoSideTabAccentBorders (change a172, Impeccable 리뷰 1~3).
//
// 알림·카드·내비 한쪽에 3px 색 띠를 두는 "side-tab" 은 AI 티의 대표 징후이고 규칙 20 이
// 금지한다. 띠는 border-left 로도, inset box-shadow 로도 만들 수 있어 둘 다 막는다.
// 1px 구분선(border-left: 1px / 0)은 정상이므로 3px 이상만 센다.
func TestNoSideTabAccentBorders(t *testing.T) {
	start := strings.Index(pageTemplates, `{{define "style"}}`)
	end := strings.Index(pageTemplates[start:], "{{end}}")
	if start < 0 || end < 0 {
		t.Fatal("stylesheet block not found")
	}
	css := pageTemplates[start : start+end]
	sideTab := regexp.MustCompile(`border-(?:left|right)\s*:\s*[2-9]px|inset\s+-?[2-9]px\s+0\s`)
	if m := sideTab.FindAllString(css, -1); len(m) != 0 {
		t.Errorf("stylesheet carries side-tab accent borders: %q", m)
	}
}

// styleBlock returns the console stylesheet (the "style" template body).
func styleBlock(t *testing.T) string {
	t.Helper()
	start := strings.Index(pageTemplates, `{{define "style"}}`)
	if start < 0 {
		t.Fatal("stylesheet block not found")
	}
	end := strings.Index(pageTemplates[start:], "{{end}}")
	return pageTemplates[start : start+end]
}

// TestMotionContract (change a172 4단계, 안전 불변식 4).
//
// 모션은 compositor 속성만 쓰고, 상태를 바꾸는 POST 폼·승인 영역의 버튼은 전환이 0 이어야
// 한다. 꾸밈이 주문·승인·청산·토글 버튼의 즉시성을 늦추지 못하게 이 시험이 막는다.
func TestMotionContract(t *testing.T) {
	css := styleBlock(t)

	// (a) transition: all 금지
	if regexp.MustCompile(`transition(?:-property)?\s*:\s*[^;{}]*\ball\b`).MatchString(css) {
		t.Error("stylesheet transitions `all`")
	}
	// (b) 전환·키프레임 속성 목록에 레이아웃·paint-heavy 속성 0건. 허용: translate scale opacity
	// transform background-color color. 이 밖의 이름이 하나라도 있으면 실패.
	allowed := map[string]bool{"translate": true, "scale": true, "opacity": true, "transform": true,
		"rotate": true, "background-color": true, "color": true, "none": true}
	for _, m := range regexp.MustCompile(`(?m)transition\s*:\s*([^;}]*)`).FindAllStringSubmatch(css, -1) {
		for _, part := range strings.Split(m[1], ",") {
			fields := strings.Fields(part)
			if len(fields) == 0 {
				continue
			}
			if !allowed[fields[0]] {
				t.Errorf("transition animates %q, which is not a compositor/hover-colour property", fields[0])
			}
		}
	}
	for _, m := range regexp.MustCompile(`@keyframes\s+[a-z-]+\s*\{([^@]*?)\}\s*\n`).FindAllStringSubmatch(css, -1) {
		for _, prop := range regexp.MustCompile(`([a-z-]+)\s*:`).FindAllStringSubmatch(m[1], -1) {
			if !allowed[prop[1]] {
				t.Errorf("@keyframes animates %q", prop[1])
			}
		}
	}
	// blur·backdrop-filter 금지
	if regexp.MustCompile(`backdrop-filter|filter\s*:\s*blur|\bblur\(`).MatchString(css) {
		t.Error("stylesheet uses blur or backdrop-filter")
	}
	// (c) will-change 는 전체 2곳 이하
	if n := strings.Count(css, "will-change"); n > 2 {
		t.Errorf("will-change appears %d times, want at most 2", n)
	}
	// (d) reduced-motion 블록이 시간 토큰을 0 으로 만들고 움직임 크기를 리셋
	rm := regexp.MustCompile(`@media \(prefers-reduced-motion: reduce\) \{[^@]*?\n\}`).FindString(css)
	if rm == "" {
		t.Fatal("prefers-reduced-motion block is missing")
	}
	for _, want := range []string{"--dur-1: 0.01ms", "--dur-2: 0.01ms", "--dur-3: 0.01ms", "--nudge: 0px", "--press: 1", "--pop-from: 1"} {
		if !strings.Contains(rm, want) {
			t.Errorf("reduced-motion block does not reset %s", want)
		}
	}
	// (e) 제외 컨테이너: POST 폼 버튼·승인 바·승인 버튼 묶음·확인 체크·승인 안내는 transition: none
	exclusion := regexp.MustCompile(`(?s)form:not\(\[method=get\]\) button[^{]*\{[^}]*\}`).FindString(css)
	if !strings.Contains(exclusion, "transition: none") {
		t.Fatalf("mutating-form buttons are not pinned to transition: none: %q", exclusion)
	}
	for _, container := range []string{".sticky-save", ".approval-actions", ".confirm-check", ".console-approval"} {
		if !strings.Contains(exclusion, container) {
			t.Errorf("exclusion rule does not cover %s", container)
		}
	}
	// 일반 버튼 전환 규칙은 같은 컨테이너를 :not() 으로도 뺀다(규칙 순서가 바뀌어도 전환이 새지 않게).
	if !strings.Contains(css, `button:not(:where(form:not([method=get]) *`) {
		t.Error("button transitions do not exclude mutating forms by selector")
	}
	// (f) 스프링 이징은 linear() 이고, 단일 cubic-bezier 곡선은 쓰지 않음
	if !strings.Contains(css, "--ease-spring: linear(") || !strings.Contains(css, "--ease-settle: linear(") || !strings.Contains(css, "--ease-pop: linear(") {
		t.Error("spring easing tokens are not linear() curves")
	}
	if !strings.Contains(css, "@supports (animation-timing-function: linear(0, 1))") {
		t.Error("linear() easings have no @supports guard with a fallback")
	}
	if strings.Contains(css, "cubic-bezier") || strings.Contains(css, "ease-in-out") {
		t.Error("stylesheet uses a single-curve easing")
	}
}

// TestAmbientMotionContract (change a172 6단계, 사람 결정: CSS 만, JS·CSP 완화 없음).
//
// 앰비언트 배경은 정보(엔진·데이터 톤)를 나르는 정적 그라디언트 덩어리이고, 움직이는 것은 덩어리의
// translate·rotate·opacity 뿐이다. 그라디언트·배경 위치·필터를 애니메이션하면 매 프레임 paint 가 되고,
// 새로고침 화면(2초)에서는 애니메이션이 재시작해 튀므로 멈춰야 하며, 호버 조명은 주문·승인 영역에
// 닿지 않아야 한다. 이 시험이 그 계약을 고정한다.
func TestAmbientMotionContract(t *testing.T) {
	css := styleBlock(t)

	// (a) 키프레임: 속성은 compositor 화이트리스트뿐이고, 앰비언트 키프레임이 실제로 존재
	allowed := map[string]bool{"translate": true, "rotate": true, "scale": true, "opacity": true, "transform": true,
		"from": true, "to": true}
	ambient := regexp.MustCompile(`@keyframes\s+(amb-[a-z-]+)\s*\{([^@]*?)\}\s*\n`).FindAllStringSubmatch(css, -1)
	if len(ambient) < 2 {
		t.Fatalf("ambient keyframes amb-* = %d, want at least 2", len(ambient))
	}
	for _, m := range ambient {
		for _, prop := range regexp.MustCompile(`([a-z-]+)\s*:`).FindAllStringSubmatch(m[2], -1) {
			if !allowed[prop[1]] {
				t.Errorf("@keyframes %s animates %q", m[1], prop[1])
			}
		}
	}
	// (b) 합성 비용이 큰 것 금지: filter·mix-blend-mode·배경/그라디언트 애니메이션
	for _, banned := range []string{"mix-blend-mode", "background-position", "background-size", "backdrop-filter"} {
		if strings.Contains(css, banned) {
			t.Errorf("stylesheet uses %s", banned)
		}
	}
	if regexp.MustCompile(`(^|[;{\s])filter\s*:`).MatchString(css) {
		t.Error("stylesheet uses filter")
	}
	// (c) 장식은 클릭을 가로채지 않고 콘텐츠 아래에 있음
	pseudo := regexp.MustCompile(`(?s)body > header::before[^{]*\{[^}]*\}`).FindString(css)
	for _, want := range []string{"pointer-events: none", "z-index: -1", "radial-gradient(", "color-mix("} {
		if !strings.Contains(pseudo, want) {
			t.Errorf("ambient blob rule lacks %q: %q", want, pseudo)
		}
	}
	header := regexp.MustCompile(`(?m)^body > header \{[^}]*\}`).FindString(css)
	if !strings.Contains(header, "overflow: clip") {
		t.Errorf("header does not clip the ambient blobs: %q", header)
	}
	if !strings.Contains(pseudo+css, "width: min(42%, 30rem)") || !strings.Contains(css, "height: min(170%, 22rem)") {
		t.Error("ambient blob size is not capped")
	}
	// (d) 새로고침 화면(상태 스트립이 data-reload=on 을 렌더)은 애니메이션 정지
	refresh := regexp.MustCompile(`(?s)body:has\(\[data-reload=on\]\)[^{]*\{[^}]*\}`).FindString(css)
	if !strings.Contains(refresh, "animation: none") || !strings.Contains(refresh, ".console-status") {
		t.Errorf("refreshing screens are not pinned to animation: none: %q", refresh)
	}
	// (e) 톤: 알림색 토큰이 --tone 으로 이어지고, 가장 나쁜 톤이 마지막에 이겨야 함
	iok, iwarn, ibad := strings.Index(css, "--tone: var(--ok)"), strings.Index(css, "--tone: var(--warn)"), strings.Index(css, "--tone: var(--bad)")
	if iok < 0 || iwarn < iok || ibad < iwarn {
		t.Errorf("tone rules must be ordered ok < warn < bad (got %d %d %d)", iok, iwarn, ibad)
	}
	// (f) 방어 미디어
	if !regexp.MustCompile(`(?s)@media \(prefers-reduced-motion: reduce\) \{[^@]*?animation: none`).MatchString(css) {
		t.Error("reduced-motion does not stop the ambient animation")
	}
	rt := regexp.MustCompile(`(?s)@media \(prefers-reduced-transparency: reduce\) \{.*?\n\}`).FindString(css)
	if !strings.Contains(rt, "background: none") || !strings.Contains(rt, ".settings-tabs a)::after") || !strings.Contains(rt, ".console-status::before") {
		t.Errorf("prefers-reduced-transparency does not clear both the ambient blobs and the hover glow: %q", rt)
	}
	// (g) hover 조명: opacity 로만 페이드하고, 대상은 렌더된 화면에서 POST 폼을 품지 않은 이동 링크뿐이어야 함.
	// 문자열(제외 선택자 존재)만 보면 .tier-card·.choice-tile 처럼 폼을 품은 카드가 조명을 받는 실제 위반을 통과시킨다.
	glow := regexp.MustCompile(`(?s):is\(([^)]*)\):not\(:where\(([^{]*)\)\)::after\s*\{[^}]*\}`).FindStringSubmatch(css)
	if glow == nil {
		t.Fatal("hover glow rule is missing")
	}
	if !strings.Contains(glow[0], "opacity: 0") || !strings.Contains(glow[0], "transition: opacity") {
		t.Errorf("hover glow does not fade by opacity only: %q", glow[0])
	}
	for _, container := range []string{".sticky-save", ".approval-actions", ".confirm-check", ".console-approval", "form:not([method=get])"} {
		if !strings.Contains(glow[2], container) {
			t.Errorf("hover glow rule does not exclude %s: %q", container, glow[2])
		}
	}
	if !strings.Contains(css, "(hover: hover)") {
		t.Error("hover glow is not limited to hover-capable pointers")
	}
	for _, banned := range []string{".tier-card", ".choice-tile"} {
		if strings.Contains(glow[1], banned) {
			t.Errorf("hover glow targets %s, which holds a POST form", banned)
		}
	}
	h := newSettingsWiredHarness(t)
	h.authenticate(t)
	checked := 0
	for _, screen := range append([]struct {
		path   string
		reload int
	}{}, consoleScreens...) {
		doc, err := html.Parse(strings.NewReader(h.page(t, screen.path)))
		if err != nil {
			t.Fatalf("parse %s: %v", screen.path, err)
		}
		for _, target := range strings.Split(glow[1], ",") {
			fields := strings.Fields(strings.TrimSpace(target))
			class, tag := strings.TrimPrefix(fields[0], "."), ""
			if len(fields) > 1 {
				tag = fields[1]
			}
			walkHTML(doc, func(n *html.Node) {
				if n.Type != html.ElementNode || !hasClass(n, class) {
					return
				}
				// 최적화 nav·설정 탭은 컨테이너 클래스 아래의 <a> 가 대상이다.
				var hits []*html.Node
				if tag == "" {
					hits = []*html.Node{n}
				} else {
					walkHTML(n, func(c *html.Node) {
						if c.Type == html.ElementNode && c.Data == tag {
							hits = append(hits, c)
						}
					})
				}
				for _, hit := range hits {
					checked++
					walkHTML(hit, func(c *html.Node) {
						if c.Type == html.ElementNode && c.Data == "form" {
							t.Errorf("%s: hover glow target %q contains a <form>", screen.path, target)
						}
					})
				}
			})
		}
	}
	if checked == 0 {
		t.Error("no rendered element matched the hover glow selectors; the check proved nothing")
	}
	// (h) 스크립트 0, CSP 불변: script-src 는 여전히 없음
	if strings.Contains(strings.ToLower(css), "<script") || strings.Contains(consoleHTMLCSP, "script-src") {
		t.Error("the console permits scripts")
	}
	// (i) will-change 총량
	if n := strings.Count(css, "will-change"); n > 2 {
		t.Errorf("will-change appears %d times, want at most 2", n)
	}
}

func hasClass(n *html.Node, class string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

// TestEveryPostFormButtonIsExcludedFromMotion walks the real templates: every
// <form method="post"> button sits inside a form[method="post"], so the exclusion
// selector covers it by construction. This pins that the selector's attribute form
// matches the markup the templates emit. The stylesheet selects mutating forms as
// `form:not([method=get])` on purpose, never by spelling the POST attribute: read-only screens
// are pinned to contain no `method="post"`/`method=post` string anywhere in the page, CSS
// included. Fail-safe direction: a form that does not say GET is treated as mutating.
func TestEveryPostFormButtonIsExcludedFromMotion(t *testing.T) {
	if !strings.Contains(pageTemplates+optimizationTemplates, `<form method="post"`) {
		t.Fatal("templates no longer emit <form method=\"post\">; update the motion exclusion selector")
	}
	// 경로 의존 방어(R4): 폼 안 버튼, 제출 입력, formmethod/formaction 이 붙은 어떤 요소든
	// 전환 규칙의 :not() 목록과 transition: none 규칙에 모두 들어 있어야 한다.
	css := styleBlock(t)
	for _, selector := range []string{"[formmethod]", "[formaction]", "form button", "form input[type=submit]"} {
		if got := strings.Count(css, selector); got < 3 {
			t.Errorf("exclusion selector %q appears %d times, want it in all three :not() lists and the transition: none rule", selector, got)
		}
		exclusion := regexp.MustCompile(`(?s)form:not\(\[method=get\]\) button[^{]*\{[^}]*\}`).FindString(css)
		if !strings.Contains(exclusion, selector) {
			t.Errorf("transition: none rule does not list %q", selector)
		}
	}
}
