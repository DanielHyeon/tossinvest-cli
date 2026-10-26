package console

import (
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
	"golang.org/x/net/html"
)

func TestOptimizationCandidateFiltersRenderDormantReadOnlyEvidenceState(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	page := h.page(t, "/optimization?category=candidate-filters")
	for _, want := range []string{
		`id="candidate-filters"`, "후보 필터", "KR", "US", "regular", "seen_late",
		"extended", "near_high", "미승인", "passed 구조적 0", "verdict 비활성",
		"legacy-unapproved", "2.0", "not_measured", "desired", "effective",
		"다음 candidate 평가", "before/after", "CAS", "주문·RiskIntent·LIVE 상태는 변경하지 않는다",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("candidate filters page does not contain %q", want)
		}
	}
	section := candidateFilterNode(t, page)
	walkHTML(section, func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}
		switch node.Data {
		case "form", "textarea", "input":
			t.Errorf("candidate filters exposes forbidden <%s> control", node.Data)
		}
		if _, found := htmlAttribute(node, "contenteditable"); found {
			t.Errorf("candidate filters exposes contenteditable on <%s>", node.Data)
		}
	})
}

func TestCandidateFilterCardsContainExactMarketMetricMatrix(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	section := candidateFilterNode(t, h.page(t, "/optimization?category=candidate-filters"))
	tables := a175NodesByClass(section, "table", "veto-matrix")
	markets := candidate.CandidateFilterMarkets()
	if len(tables) != 1 {
		t.Fatalf("candidate-filter matrices = %d, want exactly 1", len(tables))
	}
	if _, found := htmlAttribute(tables[0], "aria-readonly"); found {
		t.Error("native candidate-filter table must not rely on aria-readonly")
	}
	if got := a175ColumnHeaderCount(tables[0]); got != len(markets)+1 {
		t.Errorf("candidate-filter matrix columns = %d, want key plus %d markets", got, len(markets))
	}
	matrix := candidateFilterMatrixFromMarkets(markets)
	if got := a175MatrixDataRowCount(tables[0]); got != len(matrix.Rows) {
		t.Errorf("candidate-filter matrix rows = %d, want %d registry union keys", got, len(matrix.Rows))
	}
	for _, row := range matrix.Rows {
		metric := row.Key
		if got := exactElementTextCount(tables[0], "code", metric); got != 1 {
			t.Errorf("%s matrix key count = %d, want exactly 1", metric, got)
		}
	}
}

func TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	page := h.page(t, "/optimization?category=candidate-filters")
	section := candidateFilterNode(t, page)
	tables := a175NodesByClass(section, "table", "veto-matrix")
	markets := candidate.CandidateFilterMarkets()
	if len(tables) != 1 {
		t.Fatalf("candidate-filter matrices = %d, want exactly 1", len(tables))
	}
	if got := a175ColumnHeaderCount(tables[0]); got != len(markets)+1 {
		t.Errorf("candidate-filter matrix scope=col headers = %d, want %d", got, len(markets)+1)
	}
	if got := a175MatrixDataRowCount(tables[0]); got != len(candidateFilterMatrixFromMarkets(markets).Rows) {
		t.Errorf("candidate-filter matrix scope=row headers = %d, want registry union key count", got)
	}
	for _, want := range []string{`aria-label="후보 필터 시장 전환"`, `href="#candidate-filters-KR"`,
		`href="#candidate-filters-US"`, `aria-label="후보 필터 KR US 행렬"`, `<dl>`} {
		if !strings.Contains(page, want) {
			t.Errorf("candidate filter section lacks accessible/mobile marker %q", want)
		}
	}
}

func candidateFilterNode(t *testing.T, page string) *html.Node {
	t.Helper()
	document, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatalf("parse optimization HTML: %v", err)
	}
	nodes := nodesByAttribute(document, "section", "id", "candidate-filters")
	if len(nodes) != 1 {
		t.Fatalf("candidate-filters sections = %d, want exactly 1", len(nodes))
	}
	return nodes[0]
}

func walkHTML(root *html.Node, visit func(*html.Node)) {
	visit(root)
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		walkHTML(child, visit)
	}
}

func htmlAttribute(node *html.Node, key string) (string, bool) {
	for _, attribute := range node.Attr {
		if strings.EqualFold(attribute.Key, key) {
			return attribute.Val, true
		}
	}
	return "", false
}

func nodesByName(root *html.Node, name string) []*html.Node {
	var out []*html.Node
	walkHTML(root, func(node *html.Node) {
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, name) {
			out = append(out, node)
		}
	})
	return out
}

func nodesByAttribute(root *html.Node, name, key, value string) []*html.Node {
	var out []*html.Node
	walkHTML(root, func(node *html.Node) {
		if node.Type != html.ElementNode || !strings.EqualFold(node.Data, name) {
			return
		}
		if got, found := htmlAttribute(node, key); found && got == value {
			out = append(out, node)
		}
	})
	return out
}

func exactElementTextCount(root *html.Node, name, want string) int {
	count := 0
	for _, node := range nodesByName(root, name) {
		var text strings.Builder
		walkHTML(node, func(descendant *html.Node) {
			if descendant.Type == html.TextNode {
				text.WriteString(descendant.Data)
			}
		})
		if strings.TrimSpace(text.String()) == want {
			count++
		}
	}
	return count
}
