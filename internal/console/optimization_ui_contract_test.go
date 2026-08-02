package console

import (
	"context"
	"net/http"
	"strings"
	"testing"

	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"github.com/JungHoonGhae/tossinvest-cli/internal/settingmeta"
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
		"@media (max-width: 420px)", "min-height: 44px", "max-width: 100%", "overflow-x: auto",
		"a:focus-visible", "button:focus-visible", "aria-label=\"최적화 카테고리\"",
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

func TestOptimizationUsesOnePresetPreviewFlowWithoutClientDraft(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(nodesByName(doc, "form")); got != 3 {
		t.Fatalf("preset preview forms=%d, want exactly 3 server-defined choices", got)
	}
	if got := len(nodesByName(doc, "button")); got != 3 {
		t.Fatalf("preset preview buttons=%d, want exactly 3", got)
	}
	for _, forbidden := range []string{"select", "textarea", "script"} {
		if got := len(nodesByName(doc, forbidden)); got != 0 {
			t.Errorf("preset page has %d <%s> elements", got, forbidden)
		}
	}
	for _, marker := range []string{
		`aria-label="익절 보호 설정 적용 순서"`, "1 · preset 선택", "2 · before/after 확인",
		"3 · 3초 확인 후 적용", `data-lifecycle-state="ready"`, `data-evidence-state="unavailable"`,
		`href="/optimization?category=position-management"`, `href="/optimization?category=candidate-filters"`,
		`href="/optimization?category=strategy-runtime"`, `href="/optimization?category=performance-history"`,
	} {
		if !strings.Contains(page, marker) {
			t.Errorf("minimal preset flow lacks %q", marker)
		}
	}
	for _, inventedState := range []string{"localStorage", "sessionStorage", "data-dirty", "미저장 변경", "초기화"} {
		if strings.Contains(page, inventedState) {
			t.Errorf("single-field screen invented client draft state %q", inventedState)
		}
	}
}

func TestOptimizationConfigurationErrorIsReadOnlyAndSuppressesPresetControls(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	field, ok := commander.view.Registry.Field("exit.common-policy")
	if !ok {
		t.Fatal("core registry lacks exit.common-policy")
	}
	field.Descriptor.Description = ""
	registry, err := strategyopt.BuildRegistry(context.Background(), strategyopt.ProviderBinding{
		Category: strategyopt.CategoryExitProtection,
		Provider: strategyopt.StaticProvider{Owner: field.Descriptor.Provenance.OwnerChange,
			Fields: []settingmeta.FieldDescriptor{field.Descriptor}},
	})
	if err != nil {
		t.Fatal(err)
	}
	commander.view.Registry = registry
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"설정 오류 · 읽기 전용", "owner descriptor를 바로잡기 전", "읽기 전용 · owner descriptor 확인 필요"} {
		if !strings.Contains(page, marker) {
			t.Errorf("configuration error UI lacks %q", marker)
		}
	}
	if got := len(nodesByAttribute(doc, "fieldset", "aria-readonly", "true")); got != 1 {
		t.Errorf("read-only error fieldsets=%d, want 1", got)
	}
	for _, element := range []string{"form", "input", "button", "select", "textarea"} {
		if got := len(nodesByName(doc, element)); got != 0 {
			t.Errorf("configuration error page exposes %d <%s> mutation controls", got, element)
		}
	}
}

func TestOptimizationStaleEvidenceIsExplicitAndFailClosed(t *testing.T) {
	commander := newFakeOptimizationCommander(t)
	commander.view.Evidence = strategyopt.Evidence{Status: strategyopt.EvidenceStale, Missing: []string{"a049-window-expired"}}
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := h.page(t, "/optimization")
	for _, marker := range []string{`data-evidence-state="stale"`, "오래됨 · 근거 추천 중지", "근거 기반 추천 candidate를 만들지 않는다", "a049-window-expired"} {
		if !strings.Contains(page, marker) {
			t.Errorf("stale evidence state lacks %q", marker)
		}
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
func (failedOptimizationCommander) RecoverConflict(context.Context, string) (strategyopt.ConflictView, error) {
	return strategyopt.ConflictView{}, strategyopt.ErrInsufficientEvidence
}

func TestOptimizationLoadingModelIsServerBlockingAndReadFailureIsFailClosed(t *testing.T) {
	// The console is server rendered: it has no asynchronous client loading state
	// that could masquerade as a default. A failed bounded Read renders an error
	// and no setting form.
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = failedOptimizationCommander{} })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	for _, want := range []string{"lifecycle을 읽지 못했다", "모든 변경을 닫았다", "미검증/사용 불가"} {
		if !strings.Contains(page, want) {
			t.Errorf("fail-closed page lacks %q", want)
		}
	}
	if strings.Contains(page, `<form`) || strings.Contains(page, `<button`) {
		t.Error("read failure retained mutation controls")
	}
}
