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
	for _, raw := range []string{"HALF_RISK", "intent-77"} {
		if strings.Contains(row, raw) {
			t.Errorf("managed legacy row exposes raw exit value %q", raw)
		}
	}
	if !strings.Contains(row, "원장 기준선 <strong>69500</strong>") ||
		!strings.Contains(row, "현재 보호선 <strong>—</strong>") ||
		!strings.Contains(row, "다음 익절 <strong>—</strong>") {
		t.Error("managed legacy row does not distinguish stored evidence from actionable prices")
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

// 두 테마는 `:root` 변수로만 색을 정한다(change a172). 한 함수가 라이트·다크를 모두 돌며
// 변수 값을 읽어 직접 대비를 잰다.
//   - 텍스트 토큰(잉크 1~3, 액센트, 의미색 fg) : 세 면(surface-0/1/2) 위에서 4.5:1 이상
//   - 잉크는 의미색 바탕(ok/warn/bad-bg) 위에서도 4.5:1, 의미색 fg 는 자기 bg 위에서 4.5:1
//   - UI 경계(line-strong, 포커스 링 accent, 의미색 line) : 세 면 위에서 3:1 이상
//
// line / line-soft 는 장식 구분선이라 대상이 아니다.
func TestTradingViewsThemeColorsMeetWCAGInBothThemes(t *testing.T) {
	rootBlock := regexp.MustCompile(`:root \{[^}]*\}`)
	darkStart := strings.Index(pageTemplates, "@media (prefers-color-scheme: dark) {")
	if darkStart < 0 {
		t.Fatal("dark theme media block is missing")
	}
	blocks := map[string]string{
		"light": rootBlock.FindString(pageTemplates[:darkStart]),
		"dark":  rootBlock.FindString(pageTemplates[darkStart:]),
	}
	hexVar := regexp.MustCompile(`(--[a-z0-9-]+):\s*(#[0-9a-fA-F]{6})`)
	for theme, block := range blocks {
		vars := map[string]string{}
		for _, m := range hexVar.FindAllStringSubmatch(block, -1) {
			vars[m[1]] = m[2]
		}
		get := func(name string) string {
			if vars[name] == "" {
				t.Fatalf("%s theme does not define %s", theme, name)
			}
			return vars[name]
		}
		surfaces := []string{"--surface-0", "--surface-1", "--surface-2"}
		check := func(token string, against []string, floor float64) {
			for _, bg := range against {
				if ratio := wcagContrastRatio(t, get(token), get(bg)); ratio < floor {
					t.Errorf("%s: %s %s on %s %s = %.2f:1, want at least %.1f:1", theme, token, get(token), bg, get(bg), ratio, floor)
				}
			}
		}
		for _, token := range []string{"--ink-1", "--ink-2", "--ink-3", "--accent", "--ok", "--warn", "--bad"} {
			check(token, surfaces, 4.5)
		}
		for _, token := range []string{"--ink-1", "--ink-2", "--ink-3"} {
			check(token, []string{"--ok-bg", "--warn-bg", "--bad-bg"}, 4.5)
		}
		for _, semantic := range []string{"ok", "warn", "bad"} {
			check("--"+semantic, []string{"--" + semantic + "-bg"}, 4.5)
			check("--"+semantic+"-line", surfaces, 3)
		}
		check("--line-strong", surfaces, 3)
		check("--accent", surfaces, 3)
		// 잉크는 1 → 3 으로 갈수록 옅어져야 한다(단조 스케일).
		first, second, third := wcagContrastRatio(t, get("--ink-1"), get("--surface-1")),
			wcagContrastRatio(t, get("--ink-2"), get("--surface-1")), wcagContrastRatio(t, get("--ink-3"), get("--surface-1"))
		if !(first > second && second > third) {
			t.Errorf("%s: ink steps are not monotonic against surface-1: %.2f %.2f %.2f", theme, first, second, third)
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
