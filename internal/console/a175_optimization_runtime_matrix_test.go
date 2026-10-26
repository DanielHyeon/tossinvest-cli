package console

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"golang.org/x/net/html"
)

func TestOptimizationHasOneNavigation(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)

	overview := a175HTML(t, h.page(t, "/optimization"))
	navs := nodesByAttribute(overview, "nav", "class", "optimization-nav")
	if len(navs) != 1 {
		t.Fatalf("optimization navigation = %d, want 1", len(navs))
	}
	for _, href := range []string{"/strategy-runtime", "/strategy-runtime/market-schedule", "/performance-history"} {
		if got := a175DescendantHrefCount(navs[0], href); got != 1 {
			t.Errorf("optimization navigation link %q = %d, want 1", href, got)
		}
	}
	if got := len(a175NodesByClass(overview, "dl", "status-strip")); got != 1 {
		t.Errorf("overview KPI block = %d, want 1", got)
	}
	if got := len(nodesByAttribute(overview, "div", "class", "category-summary")); got != 0 {
		t.Errorf("overview category-card grid = %d, want 0", got)
	}

	for _, category := range []strategyopt.Category{
		strategyopt.CategoryExitProtection,
		strategyopt.CategoryPositionManagement,
		strategyopt.CategoryCandidateFilters,
		strategyopt.CategoryStrategyRuntime,
		strategyopt.CategoryPerformanceHistory,
	} {
		document := a175HTML(t, h.page(t, "/optimization?category="+string(category)))
		if got := len(a175NodesByClass(document, "dl", "status-strip")); got != 0 {
			t.Errorf("%s KPI block = %d, want 0 outside overview", category, got)
		}
		if got := len(nodesByAttribute(document, "div", "class", "category-summary")); got != 0 {
			t.Errorf("%s category-card grid = %d, want 0", category, got)
		}
	}
}

func TestOptimizationRelatedLinksAreLabelledPlainList(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	document := a175HTML(t, h.page(t, "/optimization"))
	navs := nodesByAttribute(document, "nav", "class", "optimization-nav")
	if len(navs) != 1 {
		t.Fatalf("optimization navigation = %d, want 1", len(navs))
	}
	related := a175NodesByClass(navs[0], "div", "optimization-related")
	if len(related) != 1 {
		t.Fatalf("optimization related link groups = %d, want 1", len(related))
	}
	if got := len(a175NodesByClass(related[0], "p", "optimization-related-title")); got != 1 {
		t.Errorf("related links title = %d, want 1", got)
	}
	lists := nodesByName(related[0], "ul")
	if len(lists) != 1 {
		t.Fatalf("related link lists = %d, want 1", len(lists))
	}
	for _, href := range []string{"/strategy-runtime", "/strategy-runtime/market-schedule", "/performance-history"} {
		if got := a175DescendantHrefCount(lists[0], href); got != 1 {
			t.Errorf("related plain link %q = %d, want 1", href, got)
		}
	}
}

func TestPresetCompareDesktopStylesFillContainer(t *testing.T) {
	for _, want := range []string{
		`.preset-compare { width: 100%; table-layout: fixed; }`,
		`.preset-compare th, .preset-compare td { min-width: 0;`,
		`.preset-compare th small { display: block; max-inline-size: 32ch; overflow-wrap: anywhere; }`,
	} {
		if !strings.Contains(pageTemplates, want) {
			t.Errorf("preset comparison desktop style lacks %q", want)
		}
	}
}

func TestOptimizationCategorySummaryFailsClosedForUnavailableLifecycle(t *testing.T) {
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = failedOptimizationCommander{} })
	h.authenticate(t)
	document := a175HTML(t, h.page(t, "/optimization?category=candidate-filters"))
	summaries := a175NodesByClass(document, "p", "optimization-summary")
	if len(summaries) != 1 {
		t.Fatalf("unavailable lifecycle summaries = %d, want 1", len(summaries))
	}
	text := a175ElementText(summaries[0])
	if !strings.Contains(text, "entry 확인 불가") {
		t.Errorf("unavailable lifecycle entry = %q, want 확인 불가", text)
	}
	if strings.Contains(text, "entry OFF") {
		t.Errorf("unavailable lifecycle entry is rendered as OFF: %q", text)
	}
}

func TestRelayCategoriesAreOneLineAndKeepTheirLinks(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	commander.view.Audit = []strategyopt.AuditEvent{{ID: 1, AuditID: "audit-a175", Version: 12,
		CandidateID: "candidate-a175", Key: "exit.common-policy", Actor: "verified:operator",
		Reason: strategyopt.ReasonServerPreset}}
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)

	for _, test := range []struct {
		category strategyopt.Category
		heading  string
		links    []string
	}{
		{strategyopt.CategoryPositionManagement, "종목별 관리", []string{"/position-management"}},
		{strategyopt.CategoryStrategyRuntime, "전략·실행", []string{"/strategy-runtime", "/strategy-runtime/market-schedule"}},
		{strategyopt.CategoryPerformanceHistory, "성과·이력", []string{"/performance-history"}},
	} {
		t.Run(string(test.category), func(t *testing.T) {
			document := a175HTML(t, h.page(t, "/optimization?category="+string(test.category)))
			section := a175SectionByHeading(t, document, test.heading)
			if got := a175DirectElementCount(section, "p"); got != 1 {
				t.Errorf("%s relay summary paragraphs = %d, want 1", test.category, got)
			}
			for _, href := range test.links {
				if got := a175DescendantHrefCount(section, href); got != 1 {
					t.Errorf("%s relay link %q = %d, want 1", test.category, href, got)
				}
			}
		})
	}

	performance := a175HTML(t, h.page(t, "/optimization?category=performance-history"))
	if got := len(nodesByAttribute(performance, "div", "aria-label", "설정 변경 감사 이력")); got != 1 {
		t.Errorf("performance audit table regions = %d, want 1 with audit fixture", got)
	}
}

func TestExitPresetsCompareInOneTable(t *testing.T) {
	policies := exitpolicy.RegisteredCommonPolicies()
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := optimizationPolicyOnly(t, h.page(t, "/optimization?category=exit-protection"))
	document := a175HTML(t, page)
	tables := a175NodesByClass(document, "table", "preset-compare")
	if len(tables) != 1 {
		t.Fatalf("preset comparison tables = %d, want 1", len(tables))
	}
	if got := a175ColumnHeaderCount(tables[0]); got != len(policies)+1 {
		t.Errorf("preset comparison columns = %d, want %d", got, len(policies)+1)
	}

	forms := a175FormsForAction(document, "/optimization/exit-policy")
	if len(forms) != len(policies) {
		t.Fatalf("preset forms = %d, want %d", len(forms), len(policies))
	}
	for _, form := range forms {
		for _, name := range []string{"csrf", "base_version", "category", "setting_key", "option_id"} {
			if got := a175HiddenInputCount(form, name); got != 1 {
				t.Errorf("preset form hidden %s = %d, want 1", name, got)
			}
		}
	}
	for _, fragment := range []string{
		`<form method="post" action="/optimization/exit-policy" class="choice-action">`,
		`<input type="hidden" name="csrf" value="{{$.CSRF}}">`,
		`<input type="hidden" name="base_version" value="{{$.Snapshot.Version}}">`,
		`<input type="hidden" name="category" value="exit-protection">`,
		`<input type="hidden" name="setting_key" value="exit.common-policy">`,
		`<input type="hidden" name="option_id" value="{{.ID}}">`,
		`<button type="submit" {{if .Selected}}disabled aria-disabled="true"{{end}}>{{if .Selected}}현재 desired · 선택됨{{else}}이 preset 선택 · 미리보기{{end}}</button>`,
	} {
		if !strings.Contains(optimizationTemplates, fragment) {
			t.Errorf("exit-policy form contract changed: %s", fragment)
		}
	}
	text := a175ElementText(tables[0])
	if strings.Contains(text, "999") {
		t.Error("runner sentinel 999 is exposed in comparison table")
	}
	if got := strings.Count(text, "고정 목표 없음"); got != 1 {
		t.Errorf("runner sentinel label = %d, want 1", got)
	}

	readonly := newDashboardHarness(t, func(*Options) {})
	readonly.authenticate(t)
	readonlyPage := optimizationPolicyOnly(t, readonly.page(t, "/optimization?category=exit-protection"))
	readonlyDocument := a175HTML(t, readonlyPage)
	if got := len(a175FormsForAction(readonlyDocument, "/optimization/exit-policy")); got != 0 {
		t.Errorf("read-only preset forms = %d, want 0", got)
	}
	if got := strings.Count(readonlyPage, "읽기 전용 · owner descriptor 확인 필요"); got != len(policies) {
		t.Errorf("read-only preset labels = %d, want %d", got, len(policies))
	}
}

func TestCandidateFiltersRenderAsMatrix(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	section := candidateFilterNode(t, h.page(t, "/optimization?category=candidate-filters"))
	tables := a175NodesByClass(section, "table", "veto-matrix")
	markets := candidate.CandidateFilterMarkets()
	if len(tables) != 1 {
		t.Fatalf("candidate filter matrices = %d, want 1", len(tables))
	}
	if got := a175MatrixDataRowCount(tables[0]); got != len(markets[0].Filters) {
		t.Errorf("candidate filter matrix rows = %d, want %d registry keys", got, len(markets[0].Filters))
	}
	if got := a175ColumnHeaderCount(tables[0]); got != len(markets)+1 {
		t.Errorf("candidate filter matrix columns = %d, want key plus %d markets", got, len(markets))
	}
	for _, reason := range []string{"sample_count", "missing_rate", "evidence_digest", "human_activation_record", "not_measured", "미승인"} {
		if !strings.Contains(a175ElementText(tables[0]), reason) {
			t.Errorf("candidate filter matrix omits state or reason %q", reason)
		}
	}
}

func TestCandidateFilterMatrixMergesEqualMarketDescriptors(t *testing.T) {
	markets := candidate.CandidateFilterMarkets()
	matrix := candidateFilterMatrixFromMarkets(markets)
	if len(matrix.Rows) != len(markets[0].Filters) {
		t.Fatalf("candidate filter matrix rows = %d, want %d registry keys", len(matrix.Rows), len(markets[0].Filters))
	}
	for _, row := range matrix.Rows {
		if !row.Common {
			t.Errorf("equal descriptors for %q were not merged", row.Key)
		}
	}

	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	section := candidateFilterNode(t, h.page(t, "/optimization?category=candidate-filters"))
	table := a175NodesByClass(section, "table", "veto-matrix")[0]
	cells := a175NodesByClass(table, "td", "matrix-common")
	if len(cells) != len(matrix.Rows) {
		t.Fatalf("common matrix cells = %d, want %d", len(cells), len(matrix.Rows))
	}
	for _, cell := range cells {
		wantSpan := strconv.Itoa(len(markets))
		if got, ok := htmlAttribute(cell, "colspan"); !ok || got != wantSpan {
			t.Errorf("common matrix cell colspan = %q, want %s", got, wantSpan)
		}
	}
}

func TestCandidateFilterMatrixKeepsDifferentMarketValuesSeparate(t *testing.T) {
	markets := candidate.CandidateFilterMarkets()
	markets[1].Filters[0].Help = "US 별도 설명"
	matrix := candidateFilterMatrixFromMarkets(markets)
	if len(matrix.Rows) == 0 {
		t.Fatal("candidate-filter matrix rows = 0")
	}
	row := matrix.Rows[0]
	if row.Common {
		t.Fatal("different KR and US descriptors were merged")
	}
	if row.KR.Market != candidate.MarketKR || row.US.Market != candidate.MarketUS {
		t.Errorf("separate market labels = %q / %q, want KR / US", row.KR.Market, row.US.Market)
	}
	if row.KR.Help == row.US.Help || row.US.Help != "US 별도 설명" {
		t.Errorf("separate market help = %q / %q, want distinct values", row.KR.Help, row.US.Help)
	}
}

func TestCandidateFilterMatrixPreservesMarketOnlyKeys(t *testing.T) {
	markets := candidate.CandidateFilterMarkets()
	krOnly := markets[0].Filters[0]
	krOnly.Key, krOnly.Label = "kr_only", "KR 전용 필터"
	usOnly := markets[1].Filters[0]
	usOnly.Key, usOnly.Label = "us_only", "US 전용 필터"
	markets[0].Filters = []candidate.CandidateFilterDescriptor{krOnly}
	markets[1].Filters = []candidate.CandidateFilterDescriptor{usOnly}

	matrix := candidateFilterMatrixFromMarkets(markets)
	if got := len(matrix.Rows); got != 2 {
		t.Fatalf("market-only matrix rows = %d, want union 2", got)
	}
	if matrix.Rows[0].Key != "kr_only" || matrix.Rows[1].Key != "us_only" {
		t.Fatalf("market-only matrix key order = %q, %q; want KR then US first-seen union", matrix.Rows[0].Key, matrix.Rows[1].Key)
	}
	for _, row := range matrix.Rows {
		if row.Common {
			t.Errorf("market-only key %q was merged as common", row.Key)
		}
	}
	if !matrix.Rows[0].KR.Registered || matrix.Rows[0].US.Registered {
		t.Errorf("KR-only registration = KR:%t US:%t, want true/false", matrix.Rows[0].KR.Registered, matrix.Rows[0].US.Registered)
	}
	if matrix.Rows[1].KR.Registered || !matrix.Rows[1].US.Registered {
		t.Errorf("US-only registration = KR:%t US:%t, want false/true", matrix.Rows[1].KR.Registered, matrix.Rows[1].US.Registered)
	}
	var missingCell bytes.Buffer
	if err := pages.ExecuteTemplate(&missingCell, "candidate-filter-matrix-cell", matrix.Rows[0].US); err != nil {
		t.Fatalf("render unregistered market cell: %v", err)
	}
	if !strings.Contains(missingCell.String(), "이 시장에 등록 없음") {
		t.Error("unregistered market cell lacks explicit unavailable label")
	}
	if strings.Contains(missingCell.String(), "미승인 — 숫자 값 없음") {
		t.Error("unregistered market cell renders a zero-value descriptor as approved data")
	}
}

func TestCandidateFilterMatrixKeepsCategoryDifferenceSeparate(t *testing.T) {
	markets := candidate.CandidateFilterMarkets()
	markets[1].Filters[0].Category = "candidate-filters-us"
	matrix := candidateFilterMatrixFromMarkets(markets)
	if len(matrix.Rows) == 0 {
		t.Fatal("candidate-filter matrix rows = 0")
	}
	if matrix.Rows[0].Common {
		t.Fatal("descriptor with a different category was merged")
	}
}

func TestCandidateFilterMatrixFoldsDescriptionsOnly(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	section := candidateFilterNode(t, h.page(t, "/optimization?category=candidate-filters"))
	details := a175NodesByClass(section, "details", "filter-description")
	previewContract := candidate.CandidateFilterMarkets()[0].Filters[0].PreviewContract
	if len(details) != 3 {
		t.Fatalf("candidate filter description details = %d, want 3", len(details))
	}
	for _, item := range details {
		if _, open := htmlAttribute(item, "open"); open {
			t.Error("candidate filter descriptions are open by default")
		}
		for _, forbidden := range []string{"notice", "danger"} {
			if len(a175NodesByClass(item, "p", forbidden)) != 0 {
				t.Errorf("candidate filter description contains .%s", forbidden)
			}
		}
	}
	for _, required := range []string{"미승인", "not_measured", "sample_count", "CAS 필수"} {
		if !strings.Contains(a175ElementText(section), required) {
			t.Errorf("candidate filter leaves required value or reason hidden: %q", required)
		}
		for _, detail := range details {
			if strings.Contains(a175ElementText(detail), required) {
				t.Errorf("candidate filter folds required value or reason %q", required)
			}
		}
	}
	if got := strings.Count(a175ElementText(section), previewContract); got != len(details) {
		t.Errorf("visible preview contracts = %d, want %d", got, len(details))
	}
	for _, detail := range details {
		if strings.Contains(a175ElementText(detail), previewContract) {
			t.Error("candidate filter folds preview contract into description details")
		}
	}
}

func TestRuntimeGridUsesSingleColumnAt720px(t *testing.T) {
	if !strings.Contains(pageTemplates, "@media (max-width: 45rem) {\n  .runtime-grid { grid-template-columns: 1fr; }") {
		t.Error("runtime grid does not become one column at 720px")
	}
}

func TestStrategyRuntimeIsTwoColumnsWithFoldedContract(t *testing.T) {
	snapshot := strategyprojection.WithMarketFailure(consoleProjectionPair(t), strategyprojection.MarketKR,
		strategyprojection.RefusalRuntimeUnavailable, consoleProjectionNow)
	h := newHarness(t, func(options *Options) { options.StrategyRuntime = multiMarketRuntimeStub{snapshot: snapshot} })
	h.authenticate(t)
	document := a175HTML(t, body(t, h.get(t, "/strategy-runtime")))
	grids := a175NodesByClass(document, "div", "runtime-grid")
	if len(grids) != 1 {
		t.Fatalf("runtime grids = %d, want 1", len(grids))
	}
	for _, market := range []string{"KR", "US"} {
		if got := len(nodesByAttribute(grids[0], "article", "data-market", market)); got != 1 {
			t.Errorf("runtime %s blocks = %d, want 1", market, got)
		}
	}
	tech := a175NodesByClass(document, "details", "tech")
	if len(tech) != 1 {
		t.Fatalf("runtime technical details = %d, want 1", len(tech))
	}
	for _, want := range []string{"Schema", "golden"} {
		if !strings.Contains(a175ElementText(tech[0]), want) {
			t.Errorf("folded runtime contract omits %q", want)
		}
	}
	for _, field := range strategyprojection.Registry() {
		if !strings.Contains(a175ElementText(tech[0]), field.Key) {
			t.Errorf("folded runtime registry omits %q", field.Key)
		}
	}
	for _, safety := range []string{"UNWIRED", "UNKNOWN"} {
		if !strings.Contains(a175ElementText(document), safety) {
			t.Errorf("runtime omits safety state %q", safety)
		}
		for _, details := range tech {
			if strings.Contains(a175ElementText(details), safety) {
				t.Errorf("runtime folds safety state %q", safety)
			}
		}
	}
	us := nodesByAttribute(grids[0], "article", "data-market", "US")
	if len(us) != 1 || !strings.Contains(a175ElementText(us[0]), "evidence-US") {
		t.Error("KR failure hides or changes US runtime projection")
	}

	dormantHarness := newHarness(t)
	dormantHarness.authenticate(t)
	dormant := a175HTML(t, body(t, dormantHarness.get(t, "/strategy-runtime")))
	for _, safety := range []string{"OFF", "UNWIRED", "NOT_CONFIGURED", "UNKNOWN"} {
		if !strings.Contains(a175ElementText(dormant), safety) {
			t.Errorf("dormant runtime omits safety state %q", safety)
		}
		for _, details := range a175NodesByClass(dormant, "details", "tech") {
			if strings.Contains(a175ElementText(details), safety) {
				t.Errorf("dormant runtime folds safety state %q", safety)
			}
		}
	}
}

func TestSignalsCommonCopyOnceAndShadowBandsFolded(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: marketWithBands(map[candidate.VetoCode]candidate.BandTally{
		candidate.VetoSeenLate: bandTally(2, map[string]int{"0": 2}),
		candidate.VetoExtended: bandTally(3, map[string]int{"0": 3}),
	})}))
	document := a175HTML(t, page)
	if got := strings.Count(page, "그림자 지표는 기록만 하고 후보 판정에 참여하지 않는다."); got != 1 {
		t.Errorf("signals common copy = %d, want 1", got)
	}
	bands := a175NodesByClass(document, "details", "shadow-band")
	if len(bands) != 2 {
		t.Fatalf("shadow-band details = %d, want 2", len(bands))
	}
	for _, band := range bands {
		if _, open := htmlAttribute(band, "open"); open {
			t.Error("shadow-band details are open by default")
		}
	}
	vetoTables := nodesByAttribute(document, "div", "aria-label", "추격 veto 사유별 후보 수")
	if len(vetoTables) != 1 {
		t.Fatalf("veto reason tables = %d, want 1", len(vetoTables))
	}
	for _, band := range bands {
		if a175ContainsNode(band, vetoTables[0]) || strings.Contains(a175ElementText(band), "후보 1개") {
			t.Error("veto table or candidate count is folded into shadow-band details")
		}
	}
}

func TestSignalsCommonCopyIsAbsentWithoutMarketRows(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: SignalsReading{}}))
	if strings.Contains(page, "그림자 지표는 기록만 하고 후보 판정에 참여하지 않는다.") {
		t.Error("market rows가 없는데 그림자 지표 설명이 렌더링됨")
	}
}

func a175HTML(t *testing.T, page string) *html.Node {
	t.Helper()
	document, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatalf("parse HTML: %v", err)
	}
	return document
}

func a175DescendantHrefCount(root *html.Node, href string) int {
	count := 0
	walkHTML(root, func(node *html.Node) {
		if node.Type != html.ElementNode || node.Data != "a" {
			return
		}
		if got, ok := htmlAttribute(node, "href"); ok && got == href {
			count++
		}
	})
	return count
}

func a175NodesByClass(root *html.Node, name, class string) []*html.Node {
	var out []*html.Node
	walkHTML(root, func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == name && hasClass(node, class) {
			out = append(out, node)
		}
	})
	return out
}

func a175SectionByHeading(t *testing.T, root *html.Node, heading string) *html.Node {
	t.Helper()
	for _, section := range nodesByName(root, "section") {
		for _, node := range nodesByName(section, "h2") {
			if strings.TrimSpace(a175ElementText(node)) == heading {
				return section
			}
		}
	}
	t.Fatalf("section headed %q not found", heading)
	return nil
}

func a175DirectElementCount(root *html.Node, name string) int {
	count := 0
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == name {
			count++
		}
	}
	return count
}

func a175FormsForAction(root *html.Node, action string) []*html.Node {
	return nodesByAttribute(root, "form", "action", action)
}

func a175HiddenInputCount(root *html.Node, name string) int {
	count := 0
	for _, input := range nodesByAttribute(root, "input", "name", name) {
		if kind, ok := htmlAttribute(input, "type"); ok && kind == "hidden" {
			count++
		}
	}
	return count
}

func a175ColumnHeaderCount(table *html.Node) int {
	count := 0
	for _, header := range nodesByName(table, "th") {
		if scope, ok := htmlAttribute(header, "scope"); ok && scope == "col" {
			count++
		}
	}
	return count
}

func a175MatrixDataRowCount(table *html.Node) int {
	count := 0
	for _, row := range nodesByName(table, "tr") {
		if len(nodesByAttribute(row, "th", "scope", "row")) == 1 {
			count++
		}
	}
	return count
}

func a175ElementText(root *html.Node) string {
	var text strings.Builder
	walkHTML(root, func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
	})
	return text.String()
}

func a175ContainsNode(root, wanted *html.Node) bool {
	var found bool
	walkHTML(root, func(node *html.Node) {
		if node == wanted {
			found = true
		}
	})
	return found
}
