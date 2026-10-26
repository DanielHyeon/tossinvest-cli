package console

import (
	"strings"
	"testing"
)

func TestA174ExplainSummaryRestoresMobileDisclosureMarker(t *testing.T) {
	css := styleBlock(t)
	mobileAt := strings.Index(css, "@media (max-width: 720px)")
	if mobileAt < 0 {
		t.Fatal("모바일 CSS 블록이 없다")
	}
	mobile := css[mobileAt:]
	flex := strings.Index(mobile, "button, summary, nav a, .filter-group a")
	restore := strings.Index(mobile, "details.explain summary { display: list-item; }")
	if flex < 0 || restore < flex {
		t.Error("모바일 블록이 summary를 inline-flex로 만든 뒤 explain summary를 list-item으로 되돌려야 한다")
	}
}
