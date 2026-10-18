package console

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// stripHypothetical verifies the only exempt section before checking the unchanged
// policy controls. Unexpected controls must fail, not disappear from the check.
func stripHypothetical(page string) (string, error) {
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		return "", err
	}
	sections := nodesByAttribute(doc, "section", "aria-labelledby", "hypothetical-title")
	if len(sections) != 1 {
		return "", fmt.Errorf("hypothetical sections: %d", len(sections))
	}
	section := sections[0]
	forms := nodesByName(section, "form")
	if len(forms) != 1 {
		return "", fmt.Errorf("hypothetical forms: %d", len(forms))
	}
	method, _ := htmlAttribute(forms[0], "method")
	action, _ := htmlAttribute(forms[0], "action")
	if method != "post" || action != "/optimization/hypothetical" {
		return "", fmt.Errorf("unexpected form target")
	}
	expected := map[string]bool{"csrf": true, "market": true, "entry": true, "low": true, "atr": true, "tick": true, "min": true, "max": true, "budget": true, "max_shares": true, "max_notional": true}
	seen := map[string]int{}
	var bad error
	walkHTML(section, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		if _, ok := htmlAttribute(n, "contenteditable"); ok {
			bad = fmt.Errorf("contenteditable")
		}
		for _, attr := range []string{"form", "formaction", "formmethod"} {
			if _, ok := htmlAttribute(n, attr); ok {
				bad = fmt.Errorf("unexpected %s", attr)
			}
		}
		for _, attr := range n.Attr {
			if strings.HasPrefix(strings.ToLower(attr.Key), "on") {
				bad = fmt.Errorf("unexpected event handler")
			}
		}
		switch n.Data {
		case "textarea", "a", "script", "iframe":
			bad = fmt.Errorf("unexpected %s", n.Data)
		case "input", "select":
			name, _ := htmlAttribute(n, "name")
			if !expected[name] || (n.Data == "select") != (name == "market") {
				bad = fmt.Errorf("unexpected field %q", name)
			}
			if n.Data == "input" {
				kind, _ := htmlAttribute(n, "type")
				if (name == "csrf" && kind != "hidden") || (name != "csrf" && kind != "" && kind != "text") {
					bad = fmt.Errorf("unexpected input type")
				}
			}
			inside := false
			for p := n.Parent; p != nil; p = p.Parent {
				if p == forms[0] {
					inside = true
					break
				}
			}
			if !inside {
				bad = fmt.Errorf("control outside form")
			}
			seen[name]++
		}
	})
	if bad != nil {
		return "", bad
	}
	for name := range expected {
		if seen[name] != 1 {
			return "", fmt.Errorf("field %s count %d", name, seen[name])
		}
	}
	buttons := nodesByName(section, "button")
	if len(buttons) != 1 {
		return "", fmt.Errorf("unexpected buttons")
	}
	kind, _ := htmlAttribute(buttons[0], "type")
	if kind != "submit" {
		return "", fmt.Errorf("unexpected button type")
	}
	section.Parent.RemoveChild(section)
	var b bytes.Buffer
	if err := html.Render(&b, doc); err != nil {
		return "", err
	}
	return b.String(), nil
}

func optimizationPolicyOnly(t *testing.T, page string) string {
	t.Helper()
	stripped, err := stripHypothetical(page)
	if err != nil {
		t.Fatal(err)
	}
	return stripped
}

func TestA154ContractScopeRejectsForgedControls(t *testing.T) {
	h := newDashboardHarness(t, func(o *Options) { o.Optimization = newFakeOptimizationCommander(t) })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	if _, err := stripHypothetical(page); err != nil {
		t.Fatal(err)
	}
	for _, injection := range []string{
		`<input name="setting_key" value="fixed_007">`,
		`<button type="submit" formaction="/optimization/exit-policy">apply</button>`,
		`<form method="post" action="/optimization/exit-policy"></form>`,
	} {
		mutated := strings.Replace(page, `</form>`, `</form>`+injection, 1)
		if mutated == page {
			t.Fatal("mutation did not apply")
		}
		if _, err := stripHypothetical(mutated); err == nil {
			t.Fatalf("accepted forged control %s", injection)
		}
	}
}
