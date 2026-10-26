package console

import (
	"strings"
	"testing"

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
	if got := len(nodesByAttribute(overview, "div", "class", "category-summary")); got != 1 {
		t.Errorf("overview KPI block = %d, want 1", got)
	}

	for _, category := range []strategyopt.Category{
		strategyopt.CategoryExitProtection,
		strategyopt.CategoryPositionManagement,
		strategyopt.CategoryCandidateFilters,
		strategyopt.CategoryStrategyRuntime,
		strategyopt.CategoryPerformanceHistory,
	} {
		document := a175HTML(t, h.page(t, "/optimization?category="+string(category)))
		if got := len(nodesByAttribute(document, "div", "class", "category-summary")); got != 0 {
			t.Errorf("%s KPI block = %d, want 0 outside overview", category, got)
		}
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
	if len(tables) != 1 {
		t.Fatalf("candidate filter matrices = %d, want 1", len(tables))
	}
	if got := a175MatrixDataRowCount(tables[0]); got != 3 {
		t.Errorf("candidate filter matrix rows = %d, want 3 registry keys", got)
	}
	if got := a175ColumnHeaderCount(tables[0]); got != 3 {
		t.Errorf("candidate filter matrix columns = %d, want key plus KR and US", got)
	}
	for _, reason := range []string{"sample_count", "missing_rate", "evidence_digest", "human_activation_record", "not_measured", "미승인"} {
		if !strings.Contains(a175ElementText(tables[0]), reason) {
			t.Errorf("candidate filter matrix omits state or reason %q", reason)
		}
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
	for _, safety := range []string{"UNWIRED", "NOT_CONFIGURED", "UNKNOWN"} {
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
}

func TestSignalsCommonCopyOnceAndShadowBandsFolded(t *testing.T) {
	page := renderSignals(t, newSignalsHarness(t, &stubSignals{reading: oneMarket(neverChecked("005930"))}))
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
