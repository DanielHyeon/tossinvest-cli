package console

import (
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

var inlineEventHandler = regexp.MustCompile(`(?i)\son[a-z]+\s*=`)

func TestPositionsControlsWorkUnderTheDeployedCSPWithoutScript(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	resp := h.get(t, "/positions")
	page := body(t, resp)
	for name, forbidden := range map[string]bool{
		"inline event handler": inlineEventHandler.MatchString(page),
		"script element":       strings.Contains(strings.ToLower(page), "<script"),
		"javascript URL":       strings.Contains(strings.ToLower(page), "javascript:"),
	} {
		if forbidden {
			t.Errorf("positions contains a CSP-incompatible %s", name)
		}
	}
	for _, path := range []string{"/positions", "/orders"} {
		response := h.post(t, path, url.Values{})
		if response.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("POST %s = %d, want 405", path, response.StatusCode)
		}
	}
	for _, path := range []string{"/positions", "/orders"} {
		page := body(t, h.get(t, path))
		lower := strings.ToLower(page)
		for _, forbidden := range []string{"<form", "<input", "<textarea", "<select", "<button", "contenteditable"} {
			if strings.Contains(lower, forbidden) {
				t.Errorf("%s contains forbidden input surface %q", path, forbidden)
			}
		}
	}
	if !strings.Contains(page, `/optimization?category=position-management`) {
		t.Error("positions lacks the canonical position-management context link")
	}
}

func TestRemoteTradingViewsKeepTheDenyByDefaultCSP(t *testing.T) {
	rig := newRemoteTestRig(t)
	cookie := remoteLogin(t, rig.console, "10.8.0.14:44000", "mobile")
	request := remoteRequest("GET", "/positions", "10.8.0.14:44001", "mobile", nil)
	request.AddCookie(cookie)
	response := serveRemote(rig.console, request)
	if response.Code != 200 {
		t.Fatalf("remote positions status = %d", response.Code)
	}
	csp := response.Header().Get("Content-Security-Policy")
	for _, want := range []string{"default-src 'none'", "form-action 'self'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP %q is missing %q", csp, want)
		}
	}
}

func TestManagedPositionPrimaryRowShowsTheProtectionDecision(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	start := strings.Index(page, `class="position-row" data-symbol="005930"`)
	if start < 0 {
		t.Fatal("managed position has no primary position-row")
	}
	end := strings.Index(page[start:], "</tr>")
	if end < 0 {
		t.Fatal("managed position primary row does not close")
	}
	row := page[start : start+end]
	for _, want := range []string{
		"엔진 관리",
		"평가손익",
		"수익률",
		"진입가",
		"현재 보호선",
		"근거 없음",
		"다음 익절",
		"예상 수량",
	} {
		if !strings.Contains(row, want) {
			t.Errorf("managed position primary row is missing %q", want)
		}
	}
	for _, raw := range []string{"69500", "HALF_RISK", "intent-77"} {
		if strings.Contains(row, raw) {
			t.Errorf("managed legacy row exposes raw exit value %q", raw)
		}
	}
	if !strings.Contains(page, `<caption>보유 종목과 보호 상태</caption>`) ||
		!strings.Contains(page, `scope="col"`) || !strings.Contains(page, `scope="row"`) {
		t.Error("positions table lacks an accessible caption/header relationship")
	}
}

func TestOrdersUseACompactReadOnlyPrimaryRowWithTraceDetails(t *testing.T) {
	reader := &countingOrders{lists: OrdersReading{
		Open:   []OrderRecord{livePlainOrder("order-123", "005930")},
		Closed: []OrderRecord{filledPlainOrder("order-456", "AAPL")},
	}}
	h := newOrdersHarness(t, reader)
	h.authenticate(t)
	page := body(t, h.get(t, "/orders"))

	for _, want := range []string{
		`class="data-table orders-table"`,
		`<caption>주문 목록</caption>`,
		`scope="col"`,
		`scope="row"`,
		`class="order-row"`,
		`<details class="row-details">`,
		`<summary>주문 추적 정보</summary>`,
		"주문번호",
		"평균체결가",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("orders compact layout is missing %q", want)
		}
	}
	for _, forbidden := range []string{"<form", "<button", `method="post"`} {
		if strings.Contains(page, forbidden) {
			t.Errorf("orders gained an action surface %q", forbidden)
		}
	}
}

func TestTradingViewsCarryWholePageResponsiveAndFocusContracts(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	page := h.page(t, "/positions")

	for _, want := range []string{
		`@media (max-width: 720px)`,
		`header > div { flex-wrap: wrap;`,
		`nav { display: flex; flex-wrap: wrap;`,
		`min-height: 44px`,
		`:focus-visible`,
		`overflow-wrap: anywhere`,
		`aria-current="page"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("responsive/accessibility contract missing %q", want)
		}
	}
}

func TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA(t *testing.T) {
	const darkSection = "#1d1d22"
	darkStart := strings.Index(pageTemplates, "@media (prefers-color-scheme: dark) {")
	if darkStart < 0 {
		t.Fatal("dark semantic media block is missing")
	}
	darkEnd := strings.Index(pageTemplates[darkStart+1:], "@media (prefers-color-scheme: dark)")
	if darkEnd < 0 {
		t.Fatal("dark responsive media block is missing")
	}
	darkCSS := pageTemplates[darkStart : darkStart+1+darkEnd]
	for name, contract := range map[string]struct {
		selector string
		token    string
	}{
		"fresh status":          {selector: ".ok { color: #22c55e; }", token: "#22c55e"},
		"stale or error status": {selector: ".bad { color: #f43f5e; }", token: "#f43f5e"},
	} {
		if !strings.Contains(darkCSS, contract.selector) {
			t.Errorf("dark semantic CSS is missing %s contract %q", name, contract.selector)
		}
		if ratio := wcagContrastRatio(t, contract.token, darkSection); ratio < 4.5 {
			t.Errorf("%s token %s contrast on %s = %.2f:1, want at least 4.5:1", name, contract.token, darkSection, ratio)
		}
	}
}

func wcagContrastRatio(t *testing.T, foreground, background string) float64 {
	t.Helper()
	light, dark := relativeLuminance(t, foreground), relativeLuminance(t, background)
	if light < dark {
		light, dark = dark, light
	}
	return (light + 0.05) / (dark + 0.05)
}

func relativeLuminance(t *testing.T, raw string) float64 {
	t.Helper()
	if len(raw) != 7 || raw[0] != '#' {
		t.Fatalf("invalid six-digit CSS color %q", raw)
	}
	channels := make([]float64, 3)
	for index := range channels {
		value, err := strconv.ParseUint(raw[1+index*2:3+index*2], 16, 8)
		if err != nil {
			t.Fatalf("invalid CSS color %q: %v", raw, err)
		}
		channel := float64(value) / 255
		if channel <= 0.04045 {
			channels[index] = channel / 12.92
		} else {
			channels[index] = math.Pow((channel+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*channels[0] + 0.7152*channels[1] + 0.0722*channels[2]
}

// TestRenderTradingViewFixtures writes only the deterministic fake broker and
// journal pages when a visual-audit directory is explicitly supplied. Normal
// test runs skip it; live account data can never enter the design evidence.
func TestRenderTradingViewFixtures(t *testing.T) {
	out := os.Getenv("TOSSOS_UI_FIXTURE_DIR")
	if out == "" {
		t.Skip("set TOSSOS_UI_FIXTURE_DIR for a visual audit")
	}
	if err := os.MkdirAll(out, 0o700); err != nil {
		t.Fatal(err)
	}

	positions := settingsHarness(t, &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}})
	seedJournal(t, positions.journal)
	positions.authenticate(t)

	reader := &countingOrders{lists: OrdersReading{
		Open:        []OrderRecord{livePlainOrder("order-fixture-open", "005930")},
		Closed:      []OrderRecord{filledPlainOrder("order-fixture-closed", "AAPL")},
		Conditional: []ConditionalRecord{watchingConditional("condition-fixture", "035420")},
	}}
	orders := newOrdersHarness(t, reader)
	orders.authenticate(t)

	pages := map[string]string{
		"positions.html": positions.page(t, "/positions"),
		"orders.html":    body(t, orders.get(t, "/orders")),
	}
	for name, page := range pages {
		if err := os.WriteFile(filepath.Join(out, name), []byte(page), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
