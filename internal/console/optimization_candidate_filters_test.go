package console

import (
	"regexp"
	"strings"
	"testing"
)

func TestOptimizationCandidateFiltersRenderDormantReadOnlyEvidenceState(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	page := h.page(t, "/optimization")
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
	section := candidateFilterSection(t, page)
	if regexp.MustCompile(`(?i)<(?:input|textarea)\b[^>]*(?:type=["']?(?:text|number)|contenteditable)`).MatchString(section) {
		t.Fatalf("candidate filters exposes free text/number editing:\n%s", section)
	}
	if strings.Contains(section, `<form`) || strings.Contains(section, `type="hidden"`) {
		t.Fatalf("unapproved candidate filters expose an apply transport:\n%s", section)
	}
}

func TestCandidateFilterCardsRemainMobileAndAccessibilityFriendly(t *testing.T) {
	h := newDashboardHarness(t, func(*Options) {})
	h.authenticate(t)
	section := candidateFilterSection(t, h.page(t, "/optimization"))
	if strings.Contains(section, "<table") {
		t.Fatal("candidate filter cards use a wide table instead of wrapping definition lists")
	}
	for _, want := range []string{`aria-label="후보 필터 시장 전환"`, `href="#candidate-filters-KR"`,
		`href="#candidate-filters-US"`, `aria-readonly="true"`, `<dl>`} {
		if !strings.Contains(section, want) {
			t.Errorf("candidate filter section lacks accessible/mobile marker %q", want)
		}
	}
}

func candidateFilterSection(t *testing.T, page string) string {
	t.Helper()
	start := strings.Index(page, `<section id="candidate-filters"`)
	if start < 0 {
		t.Fatal("candidate-filters section absent")
	}
	rest := page[start:]
	end := strings.Index(rest, `</section>`)
	if end < 0 {
		t.Fatal("candidate-filters section is not closed")
	}
	return rest[:end+len(`</section>`)]
}
