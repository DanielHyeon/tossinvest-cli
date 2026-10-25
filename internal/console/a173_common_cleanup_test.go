package console

// a173_common_cleanup_test.go pins change a173 (콘솔 공통 정리).
//
// 다섯 계약을 전 화면 × 여러 fixture 로 잰다:
//
//	원인 1회      같은 원인의 설명 문장은 화면당 1회, 미측정 칸마다 사유 코드
//	알림 3단계    떠 있는 상자(.notice/.danger)는 화면당 3개 이하
//	식별자 접기   change ID·절대경로·64자 hex 는 details.tech 안에만
//	경계 배지     h1 옆 배지 1개 + 접힌 경계 블록 1개 이하
//	표시줄 한 줄  상태 표시줄 칸의 설명은 한 줄·영문 원문 오류 없음
//
// "모든 X 가 맞아야 한다" 는 표본이 0이면 자동으로 참이다. 그래서 빈 fixture 만이 아니라
// 원장·주문·브로커 오류를 채운 fixture 를 함께 돌리고, 표본 수가 0이면 시험이 실패하게 했다.

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a173Fixture 는 한 설정의 콘솔이고 get 은 인증된 GET 본문이다.
type a173Fixture struct {
	name string
	get  func(path string) string
}

const a173CredentialError = "verify: no Open API credentials — orders are placed through the official API only, never the web session"

// a173Fixtures 는 빈 콘솔 · 전 설정 · 원장 없음/있음 · 브로커 오류 · 주문 있음/오류 일곱 상태다.
func a173Fixtures(t *testing.T) []a173Fixture {
	t.Helper()
	var out []a173Fixture
	add := func(name string, h *harness) {
		h.authenticate(t)
		out = append(out, a173Fixture{name, func(path string) string { return body(t, h.get(t, path)) }})
	}

	add("empty", newHarness(t))
	add("settings-full", fullSettingsHarness(t))
	add("journal-absent", newDashboardHarness(t).harness)

	seeded := newDashboardHarness(t)
	seedJournal(t, seeded.journal)
	add("journal-seeded", seeded.harness)

	down := newDashboardHarness(t)
	seedJournal(t, down.journal)
	down.holdings.err = errors.New(a173CredentialError)
	add("broker-down", down.harness)

	liveJournal := filepath.Join(t.TempDir(), journal.DBFileName)
	seedJournal(t, liveJournal)
	live := newOrdersHarness(t, &countingOrders{lists: OrdersReading{
		Open: []OrderRecord{livePlainOrder("o-1", "005930")},
	}}, func(o *Options) { o.JournalPath = liveJournal })
	add("orders-live", live.harness)

	bad := newOrdersHarness(t, &countingOrders{err: errors.New(a173CredentialError)})
	add("orders-error", bad.harness)
	return out
}

// a173Screens 는 공통 껍데기 화면 전부다. 하위 화면(최적화 카테고리)은 따로 더한다.
func a173Screens() []string {
	var paths []string
	for _, s := range consoleScreens {
		paths = append(paths, s.path)
	}
	return append(paths,
		"/optimization?category=exit-protection", "/optimization?category=position-management",
		"/optimization?category=strategy-runtime", "/optimization?category=performance-history",
		"/optimization?category=candidate-filters")
}

var a173StyleBlock = regexp.MustCompile(`(?s)<style>.*?</style>`)

// a173Text 는 스타일·태그를 걷어 낸 본문 글자다.
func a173Text(page string) string {
	page = a173StyleBlock.ReplaceAllString(page, "")
	return strings.Join(strings.Fields(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(page, " ")), " ")
}

// --- ① 원인 1회 -------------------------------------------------------------------------

var a173UnmeasuredCell = regexp.MustCompile(`class="muted unmeasured"[^>]*>미측정</span> · <code>([a-z_]+)</code>`)

// TestEachUnmeasuredCauseIsExplainedOnce (task 2.1).
func TestEachUnmeasuredCauseIsExplainedOnce(t *testing.T) {
	cells, explained := 0, 0
	for _, fx := range a173Fixtures(t) {
		for _, path := range a173Screens() {
			page := fx.get(path)
			// 칸 수: 미측정 글자가 칸 표식(class="muted unmeasured") 없이 나오면 사유 코드 없는 미측정이다.
			marked := strings.Count(page, `class="muted unmeasured"`)
			coded := len(a173UnmeasuredCell.FindAllStringSubmatch(page, -1))
			if marked != coded {
				t.Errorf("%s %s: 미측정 칸 %d개 중 사유 코드가 붙은 것은 %d개", fx.name, path, marked, coded)
			}
			cells += marked

			seen := map[string]bool{}
			for _, m := range a173UnmeasuredCell.FindAllStringSubmatch(page, -1) {
				seen[m[1]] = true
			}
			// 칸이 코드로 가리키는 원인은 그 화면에 문장이 정확히 1회 있어야 한다(0이면 사유를 잃은 것).
			for code := range seen {
				sentence := unmeasuredSentences[unmeasuredReason(code)]
				if unmeasuredReason(code) == reasonDiscoveryUnreadable {
					continue // 신호 화면 전용 코드: 문장은 detail 로 배너에 1회 실린다.
				}
				if sentence == "" {
					t.Errorf("%s %s: 열거에 없는 사유 코드 %q", fx.name, path, code)
					continue
				}
				if n := strings.Count(a173Text(page), a173Text(sentence)); n != 1 {
					t.Errorf("%s %s: %s 설명 문장이 %d회 나온다, 1회여야 한다", fx.name, path, code, n)
				}
				explained++
			}
			// 칸이 없는 화면에서도 문장이 둘 이상이면 안 된다.
			for code, sentence := range unmeasuredSentences {
				if n := strings.Count(a173Text(page), a173Text(sentence)); n > 1 {
					t.Errorf("%s %s: %s 설명 문장이 %d회 반복된다", fx.name, path, code, n)
				}
			}
		}
	}
	if cells == 0 || explained == 0 {
		t.Fatalf("표본이 비었다(미측정 칸 %d, 설명 %d) — 전칭 판정이 공허하게 통과한다", cells, explained)
	}
}

// --- ② 알림 3단계 -----------------------------------------------------------------------

var a173Box = regexp.MustCompile(`class="[^"]*\b(?:notice|danger)\b`)

// TestFloatingNoticesAreAtMostThree (task 2.2).
func TestFloatingNoticesAreAtMostThree(t *testing.T) {
	samples := 0
	for _, fx := range a173Fixtures(t) {
		for _, path := range a173Screens() {
			page := a173StyleBlock.ReplaceAllString(fx.get(path), "")
			samples++
			if n := len(a173Box.FindAllString(page, -1)); n > 3 {
				t.Errorf("%s %s: 떠 있는 상자 %d개 — 3개 이하여야 한다", fx.name, path, n)
			}
		}
	}
	if samples == 0 {
		t.Fatal("표본 0")
	}
}

// --- ③ 개발자 식별자 --------------------------------------------------------------------

var (
	a173TechBlock = regexp.MustCompile(`(?s)<details class="tech".*?</details>`)
	a173ChangeID  = regexp.MustCompile(`\ba0\d{2,3}\b|\bconsole-[a-z]+-[a-z]+\b`)
	a173AbsPath   = regexp.MustCompile(`(^|[\s(>"'])/(?:home|mnt|tmp|var|etc|root|Users)/[\w./-]+`)
	a173Hex64     = regexp.MustCompile(`\b[0-9a-f]{64}\b`)
)

// TestOperatorBodyCarriesNoDeveloperIdentifiers (task 2.3).
func TestOperatorBodyCarriesNoDeveloperIdentifiers(t *testing.T) {
	techBlocks := 0
	reported := map[string]bool{}
	for _, fx := range a173Fixtures(t) {
		for _, path := range a173Screens() {
			raw := fx.get(path)
			techBlocks += len(a173TechBlock.FindAllString(raw, -1))
			text := a173Text(a173TechBlock.ReplaceAllString(raw, ""))
			for name, re := range map[string]*regexp.Regexp{
				"change ID": a173ChangeID, "절대경로": a173AbsPath, "64자 hex": a173Hex64,
			} {
				if m := re.FindString(text); m != "" && !reported[path+name] {
					reported[path+name] = true
					t.Errorf("%s %s: 운영 본문에 %s 가 있다: %q", fx.name, path, name, m)
				}
			}
		}
	}
	if techBlocks == 0 {
		t.Fatal("details.tech 가 한 번도 렌더되지 않았다 — 식별자를 어디에 옮겼는지 시험이 볼 수 없다")
	}

	// 접은 뒤에도 값은 DOM 에 남는다: 최적화 카테고리 라벨의 change ID, 원장 경로.
	fixtures := a173Fixtures(t)
	opt := fixtures[0].get("/optimization")
	if inTech := strings.Join(a173TechBlock.FindAllString(opt, -1), " "); !a173ChangeID.MatchString(inTech) {
		t.Errorf("최적화 화면의 기술 정보 블록에 change ID 가 없다 — 접은 값이 사라졌다")
	}
	absent := fixtures[2].get("/positions")
	if inTech := strings.Join(a173TechBlock.FindAllString(absent, -1), " "); !a173AbsPath.MatchString(inTech) {
		t.Errorf("포지션 화면의 기술 정보 블록에 원장 경로가 없다")
	}
}

// --- ④ 경계 배지 ------------------------------------------------------------------------

var (
	a173Badge  = regexp.MustCompile(`</h1>\s*<span class="status-pill scope">(읽기 전용|편집 가능)</span>`)
	a173Intro  = regexp.MustCompile(`(?s)</h1>.*?(?:<section|<h2)`)
	a173ScopeD = regexp.MustCompile(`(?s)<details class="scope">.*?</details>`)
)

// TestEveryScreenHeadHasScopeBadgeAndFoldedBoundary (task 2.4).
func TestEveryScreenHeadHasScopeBadgeAndFoldedBoundary(t *testing.T) {
	fx := a173Fixtures(t)[1] // 설정 seam 이 전부 배선된 콘솔 — 폼이 있는 화면도 그려진다.
	editable, readonly := 0, 0
	for _, path := range a173Screens() {
		page := fx.get(path)
		if n := strings.Count(page, `status-pill scope"`); n != 1 {
			t.Errorf("%s: 경계 배지가 %d개 — h1 옆에 정확히 1개여야 한다", path, n)
		}
		m := a173Badge.FindStringSubmatch(page)
		if m == nil {
			t.Errorf("%s: 배지가 </h1> 바로 뒤에 없다", path)
		} else if m[1] == "편집 가능" {
			editable++
		} else {
			readonly++
		}
		if n := len(a173ScopeD.FindAllString(page, -1)); n > 1 {
			t.Errorf("%s: 접힌 경계 블록이 %d개 — 1개 이하여야 한다", path, n)
		}
		if n := strings.Count(page, `<details class="scope"><summary>이 화면의 경계</summary>`); n > 1 {
			t.Errorf("%s: 경계 summary 가 %d개", path, n)
		}
		// 첫 섹션 전의 머리 단락에는 "이 화면은 ~하지 않는다" 류 경계 문장이 없다(접힌 블록으로 이동).
		head := a173ScopeD.ReplaceAllString(a173Intro.FindString(page), "")
		if strings.Contains(a173Text(head), "이 화면은") || strings.Contains(a173Text(head), "이 화면에는") {
			t.Errorf("%s: 머리 단락에 경계 문장이 남아 있다: %.120s", path, a173Text(head))
		}
	}
	if editable == 0 || readonly == 0 {
		t.Fatalf("배지 표본이 한쪽뿐이다(편집 %d, 읽기 %d)", editable, readonly)
	}
}

// --- ⑤ 상태 표시줄 한 줄 ----------------------------------------------------------------

var a173Small = regexp.MustCompile(`(?s)<small[^>]*>(.*?)</small>`)

// TestStatusStripCellsAreOneLine (task 2.5).
func TestStatusStripCellsAreOneLine(t *testing.T) {
	samples := 0
	for _, fx := range a173Fixtures(t) {
		for _, path := range a173Screens() {
			strip := stripOf(t, path, fx.get(path))
			if strings.Contains(strip, "<br") {
				t.Errorf("%s %s: 상태 표시줄 칸에 줄바꿈이 있다", fx.name, path)
			}
			for _, m := range a173Small.FindAllStringSubmatch(strip, -1) {
				samples++
				if n := utf8.RuneCountInString(a173Text(m[1])); n > 40 {
					t.Errorf("%s %s: 칸 설명이 %d자(%q) — 40자 이하 한 줄이어야 한다", fx.name, path, n, a173Text(m[1]))
				}
			}
			for _, raw := range []string{"verify:", "TOSSCTL_", "Open API credentials", "context deadline"} {
				if strings.Contains(a173Text(strip), raw) {
					t.Errorf("%s %s: 상태 표시줄에 영문 원문 오류 %q 가 보인다", fx.name, path, raw)
				}
			}
		}
	}
	if samples == 0 {
		t.Fatal("표본 0 — 칸 설명(small)이 하나도 없다")
	}
}

// --- ⑥ CSS 계약 -------------------------------------------------------------------------

// TestProseWidthAndToggleSpacingRulesExist (task 2.6).
func TestProseWidthAndToggleSpacingRulesExist(t *testing.T) {
	css := a173Fixtures(t)[0].get("/dashboard")
	i := strings.Index(css, "<style>")
	j := strings.Index(css, "</style>")
	if i < 0 || j < i {
		t.Fatal("스타일 블록이 없다")
	}
	css = css[i:j]
	for _, want := range []string{"max-inline-size: 42rem", ".toggle-row", ".aside", "details.tech", "details.scope"} {
		if !strings.Contains(css, want) {
			t.Errorf("스타일에 %q 규칙이 없다", want)
		}
	}
	// 거래 정책 체크박스 하나하나가 자기 설명을 같은 묶음 안에 둔다.
	policy := a173Fixtures(t)[1].get(pathSettingsDaily)
	if n := strings.Count(policy, `class="toggle-row"`); n < 4 {
		t.Errorf("거래 정책 토글 묶음(toggle-row)이 %d개 — 네 개여야 한다", n)
	}
}

// TestStructuralUnmeasuredSentenceIsSaidOnce — 오늘 소진분 세 축이 같은 구조적 사유를 세 번 적던 것(a173).
func TestStructuralUnmeasuredSentenceIsSaidOnce(t *testing.T) {
	page := a173Text(a173Fixtures(t)[0].get("/dashboard"))
	if n := strings.Count(page, "이 소진분은 엔진 프로세스가 메모리에서 계산해"); n != 1 {
		t.Errorf("구조적 미측정 사유 문장이 %d회 — 1회여야 한다", n)
	}
}
