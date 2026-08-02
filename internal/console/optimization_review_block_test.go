package console

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	strategyopt "github.com/JungHoonGhae/tossinvest-cli/internal/optimization"
	"golang.org/x/net/html"
)

func newReviewOptimizationStore(t *testing.T) (*strategyopt.Store, *clock.Fake) {
	t.Helper()
	clk := clock.NewFake(time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC))
	registry, err := strategyopt.CoreRegistry(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	store, err := strategyopt.Open(context.Background(), strategyopt.Options{
		Path: filepath.Join(t.TempDir(), strategyopt.DatabaseFileName), Registry: registry,
		Clock: clk, Actor: "review:test",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, clk
}

func previewCapabilityFromHTML(t *testing.T, page string) string {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	nodes := nodesByAttribute(doc, "input", "name", "capability")
	if len(nodes) != 1 {
		t.Fatalf("capability inputs=%d, want 1", len(nodes))
	}
	capability, _ := htmlAttribute(nodes[0], "value")
	if capability == "" {
		t.Fatal("preview capability is empty")
	}
	return capability
}

func assertOptimizationResponseSecurity(t *testing.T, response, normal *http.Response) string {
	t.Helper()
	for _, name := range []string{"Content-Security-Policy", "Referrer-Policy", "X-Content-Type-Options"} {
		if got, want := response.Header.Get(name), normal.Header.Get(name); got == "" || got != want {
			t.Errorf("%d %s=%q, want normal response %q", response.StatusCode, name, got, want)
		}
	}
	page := body(t, response)
	for _, marker := range []string{`<html lang="ko">`, `name="viewport"`} {
		if !strings.Contains(page, marker) {
			t.Errorf("%d refusal lacks %s", response.StatusCode, marker)
		}
	}
	return page
}

func TestRiskPreviewWaitsThreeSecondsAndAppliesSameCapabilityExactlyOnce(t *testing.T) {
	store, clk := newReviewOptimizationStore(t)
	h := newDashboardHarness(t, func(options *Options) {
		options.Optimization = store
		options.Now = clk.Now
	})
	h.authenticate(t)
	normal := h.get(t, "/optimization?category=exit-protection")

	previewResponse := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "base_version": {"1"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {exitpolicy.CommonLadderBalanced},
	})
	if previewResponse.StatusCode != http.StatusOK {
		t.Fatalf("preview status=%d", previewResponse.StatusCode)
	}
	if got := previewResponse.Header.Get("Content-Security-Policy"); got != optimizationPreviewCSP ||
		!strings.Contains(got, "script-src 'sha256-") {
		t.Fatalf("preview CSP=%q, want the exact static countdown hash", got)
	}
	previewPage := body(t, previewResponse)
	for _, marker := range []string{
		`data-not-before-ms="1785574803000"`, `aria-live="polite"`, `3초 남음`,
		`data-risk-submit disabled`, `data-risk-confirm`, `window.setTimeout`,
		`button.disabled=remaining!==0||!confirmed`, `confirmation.addEventListener("change",tick)`,
	} {
		if !strings.Contains(previewPage, marker) {
			t.Errorf("t0 preview lacks %q", marker)
		}
	}
	previewDocument, err := html.Parse(strings.NewReader(previewPage))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(nodesByAttribute(previewDocument, "section", "aria-readonly", "true")); got != 1 {
		t.Fatalf("read-only before/after review sections=%d, want 1", got)
	}
	if got := len(nodesByAttribute(previewDocument, "section", "data-risk-preview", "")); got != 1 {
		t.Fatalf("sticky approval blocks=%d, want 1", got)
	}
	checkboxes := 0
	walkHTML(previewDocument, func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}
		if node.Data == "textarea" || node.Data == "select" {
			t.Errorf("risk preview exposes forbidden <%s>", node.Data)
		}
		if _, found := htmlAttribute(node, "contenteditable"); found {
			t.Errorf("risk preview exposes contenteditable on <%s>", node.Data)
		}
		if node.Data == "input" {
			kind, _ := htmlAttribute(node, "type")
			if kind == "checkbox" {
				checkboxes++
			} else if kind != "hidden" {
				t.Errorf("risk preview exposes arbitrary input type %q", kind)
			}
		}
	})
	if checkboxes != 1 || len(nodesByName(previewDocument, "button")) != 1 {
		t.Fatalf("risk confirmation controls: checkboxes=%d buttons=%d, want 1 each",
			checkboxes, len(nodesByName(previewDocument, "button")))
	}
	scripts := nodesByName(previewDocument, "script")
	if len(scripts) != 1 || scripts[0].FirstChild == nil {
		t.Fatalf("countdown scripts=%d, want one static inline script", len(scripts))
	}
	scriptDigest := sha256.Sum256([]byte(scripts[0].FirstChild.Data))
	wantScriptSource := "'sha256-" + base64.StdEncoding.EncodeToString(scriptDigest[:]) + "'"
	if !strings.Contains(previewResponse.Header.Get("Content-Security-Policy"), wantScriptSource) {
		t.Fatalf("preview CSP does not authorize its exact script bytes: want %s", wantScriptSource)
	}
	if strings.Contains(optimizationPreviewScript, "capability") {
		t.Fatal("countdown script is allowed to mutate the approval capability")
	}
	capability := previewCapabilityFromHTML(t, previewPage)

	early := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "action": {"apply"}, "capability": {capability}, "confirm": {"yes"},
	})
	if early.StatusCode != http.StatusTooEarly {
		t.Fatalf("early apply=%d, want 425", early.StatusCode)
	}
	assertOptimizationResponseSecurity(t, early, normal)
	view, err := store.Read(context.Background())
	if err != nil || view.Snapshot.Version != 1 || len(view.Audit) != 0 {
		t.Fatalf("early POST changed state: view=%+v err=%v", view, err)
	}

	clk.Advance(3 * time.Second)
	applied := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "action": {"apply"}, "capability": {capability}, "confirm": {"yes"},
	})
	if applied.StatusCode != http.StatusOK {
		t.Fatalf("t+3 apply final status=%d", applied.StatusCode)
	}
	after, err := store.Read(context.Background())
	if err != nil || after.Snapshot.Version != 2 || len(after.Audit) != 1 {
		t.Fatalf("t+3 apply=%+v err=%v", after, err)
	}
	h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "action": {"apply"}, "capability": {capability}, "confirm": {"yes"},
	})
	replayed, err := store.Read(context.Background())
	if err != nil || replayed.Snapshot.Version != 2 || len(replayed.Audit) != 1 {
		t.Fatalf("same capability applied more than once: %+v err=%v", replayed, err)
	}
}

func TestActualApplyCASConflictRendersInputFreeRecoveryWithoutRetry(t *testing.T) {
	store, clk := newReviewOptimizationStore(t)
	view, err := store.Read(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	preview := func(optionID string) strategyopt.Preview {
		candidate, err := store.Preview(context.Background(), strategyopt.PreviewRequest{
			BaseVersion: view.Snapshot.Version, Category: strategyopt.CategoryExitProtection,
			Changes: map[string]string{"exit.common-policy": optionID},
			Source:  strategyopt.SourceServerPreset, Reason: strategyopt.ReasonServerPreset,
		})
		if err != nil {
			t.Fatal(err)
		}
		return candidate
	}
	winner := preview(exitpolicy.CommonLadderBalanced)
	stale := preview(exitpolicy.CommonLadderRunner)
	clk.Advance(3 * time.Second)
	if _, err := store.Apply(context.Background(), strategyopt.ApplyRequest{Capability: winner.Capability, Confirmed: true}); err != nil {
		t.Fatal(err)
	}

	h := newDashboardHarness(t, func(options *Options) {
		options.Optimization = store
		options.Now = clk.Now
	})
	h.authenticate(t)
	normal := h.get(t, "/optimization?category=exit-protection")
	conflict := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "action": {"apply"}, "capability": {stale.Capability}, "confirm": {"yes"},
	})
	if conflict.StatusCode != http.StatusPreconditionFailed {
		t.Fatalf("stale apply=%d, want 412", conflict.StatusCode)
	}
	page := assertOptimizationResponseSecurity(t, conflict, normal)
	for _, marker := range []string{
		"자동 retry하지 않았다", "읽기 전용으로 보존", "Attempted", "Latest desired", "Latest effective",
		"v1", "v2", exitpolicy.CommonLadderRunner, exitpolicy.CommonLadderBalanced,
		"최신값으로 돌아가기", "보존한 draft로 새 preview",
	} {
		if !strings.Contains(page, marker) {
			t.Errorf("412 recovery lacks %q", marker)
		}
	}
	for _, forbidden := range []string{`type="text"`, `type="number"`, `type="range"`, "textarea", "contenteditable"} {
		if strings.Contains(strings.ToLower(page), strings.ToLower(forbidden)) {
			t.Errorf("412 recovery exposes forbidden input %q", forbidden)
		}
	}
	afterConflict, err := store.Read(context.Background())
	if err != nil || afterConflict.Snapshot.Version != 2 || len(afterConflict.Audit) != 1 {
		t.Fatalf("412 retried automatically: %+v err=%v", afterConflict, err)
	}

	invalid := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "base_version": {"2"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {"INVENTED"},
	})
	if invalid.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid candidate=%d, want 400", invalid.StatusCode)
	}
	assertOptimizationResponseSecurity(t, invalid, normal)

	repreview := h.post(t, "/optimization/exit-policy", url.Values{
		"csrf": {h.csrf}, "base_version": {"2"}, "category": {"exit-protection"},
		"setting_key": {"exit.common-policy"}, "option_id": {exitpolicy.CommonLadderRunner},
	})
	if repreview.StatusCode != http.StatusOK || previewCapabilityFromHTML(t, body(t, repreview)) == stale.Capability {
		t.Fatal("new preview did not mint a distinct current-version capability")
	}
	final, _ := store.Read(context.Background())
	if final.Snapshot.Version != 2 || len(final.Audit) != 1 {
		t.Fatalf("new preview mutated settings: %+v", final)
	}
}

func TestOptimizationScrollableTablesHaveKeyboardFocusIndicator(t *testing.T) {
	if !strings.Contains(pageTemplates, ".table-scroll:focus-visible") {
		t.Fatal("tabindex=0 table scroll region has no focus-visible style")
	}
	commander := newFakeOptimizationCommander(t)
	h := newDashboardHarness(t, func(options *Options) { options.Optimization = commander })
	h.authenticate(t)
	page := h.page(t, "/optimization?category=exit-protection")
	if !strings.Contains(page, `class="table-scroll" role="region"`) || !strings.Contains(page, `tabindex="0"`) {
		t.Fatal("scrollable optimization table is not keyboard focusable")
	}
}
