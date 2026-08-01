package console

import (
	"context"
	"net/http"
	"strings"
	"testing"

	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"golang.org/x/net/html"
)

func TestOptimizationDOMHasNoArbitraryInputAndSubmitsOnlyOwnerOptions(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	allowedOptions := map[string]bool{}
	for _, field := range commander.view.Registry.Fields(strategyopt.CategoryExitProtection) {
		for _, option := range field.Descriptor.Options {
			allowedOptions[option.ID] = true
		}
	}
	walkHTML(doc, func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}
		switch node.Data {
		case "textarea":
			t.Errorf("optimization exposes forbidden <%s>", node.Data)
		case "input":
			kind, _ := htmlAttribute(node, "type")
			if kind != "hidden" && kind != "checkbox" {
				t.Errorf("optimization exposes arbitrary input type %q", kind)
			}
			name, _ := htmlAttribute(node, "name")
			value, _ := htmlAttribute(node, "value")
			if name == "option_id" && !allowedOptions[value] {
				t.Errorf("submitted option %q is not owner-registered", value)
			}
		}
		if _, found := htmlAttribute(node, "contenteditable"); found {
			t.Errorf("optimization exposes contenteditable on <%s>", node.Data)
		}
	})
	for _, forbidden := range []string{`type="text"`, `type="number"`, `type="range"`, "typed confirmation", "symbol 입력", "reason 입력"} {
		if strings.Contains(strings.ToLower(page), strings.ToLower(forbidden)) {
			t.Errorf("optimization page contains forbidden input affordance %q", forbidden)
		}
	}
}

func TestOptimizationReadOnlyCategoriesContainNoMutationControl(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	for _, category := range []strategyopt.Category{
		strategyopt.CategoryOverview, strategyopt.CategoryCandidateFilters,
		strategyopt.CategoryStrategyRuntime, strategyopt.CategoryPerformanceHistory,
	} {
		page := h.page(t, "/optimization?category="+string(category))
		doc, err := html.Parse(strings.NewReader(page))
		if err != nil {
			t.Fatal(err)
		}
		for _, element := range []string{"form", "input", "textarea", "button", "select"} {
			if got := len(nodesByName(doc, element)); got != 0 {
				t.Errorf("read-only category %s has %d <%s> controls", category, got, element)
			}
		}
	}
}

func TestOptimizationCategoryOrderMobileTouchFocusAndCSP(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	resp := h.get(t, "/optimization")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	page := body(t, resp)
	last := -1
	for _, category := range strategyopt.Categories() {
		needle := `href="/optimization?category=` + string(category.ID) + `"`
		index := strings.Index(page, needle)
		if index <= last {
			t.Errorf("category %s is out of fixed order", category.ID)
		}
		last = index
	}
	for _, contract := range []string{
		`name="viewport"`, "@media (max-width: 720px)", ".optimization-shell", "grid-template-columns: 1fr",
		"min-height: 44px", "overflow-x: auto", "a:focus-visible", "aria-label=\"최적화 카테고리\"",
	} {
		if !strings.Contains(page, contract) {
			t.Errorf("responsive/a11y contract lacks %q", contract)
		}
	}
	wantCSP := "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"
	if got := resp.Header.Get("Content-Security-Policy"); got != wantCSP {
		t.Errorf("CSP=%q, want %q", got, wantCSP)
	}
}

type failedOptimizationCommander struct{}

func (failedOptimizationCommander) Read(context.Context) (strategyopt.View, error) {
	return strategyopt.View{}, strategyopt.ErrInsufficientEvidence
}
func (failedOptimizationCommander) Preview(context.Context, strategyopt.PreviewRequest) (strategyopt.Preview, error) {
	return strategyopt.Preview{}, strategyopt.ErrInsufficientEvidence
}
func (failedOptimizationCommander) PreviewRollback(context.Context, strategyopt.RollbackPreviewRequest) (strategyopt.Preview, error) {
	return strategyopt.Preview{}, strategyopt.ErrInsufficientEvidence
}
func (failedOptimizationCommander) Apply(context.Context, strategyopt.ApplyRequest) (strategyopt.ApplyResult, error) {
	return strategyopt.ApplyResult{}, strategyopt.ErrInsufficientEvidence
}

func TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed(t *testing.T) {
	// The console is server rendered: it has no asynchronous client loading state
	// that could masquerade as a default. A failed bounded Read renders an error
	// and no setting form.
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = failedOptimizationCommander{} })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	for _, want := range []string{"lifecycle을 읽지 못했습니다", "모든 변경을 닫았습니다", "미검증/사용 불가"} {
		if !strings.Contains(page, want) {
			t.Errorf("fail-closed page lacks %q", want)
		}
	}
	if strings.Contains(page, `<form`) || strings.Contains(page, `<button`) {
		t.Error("read failure retained mutation controls")
	}
}
