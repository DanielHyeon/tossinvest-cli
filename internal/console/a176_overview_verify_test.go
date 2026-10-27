package console

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

func TestOverviewSummaryAnswersSixFactsAboveTheFold(t *testing.T) {
	h := newOverviewHarness(t, func(o *Options) { o.Holdings = nil })
	seedJournal(t, h.journal)
	page := h.page(t, "/dashboard")
	summary := a176Summary(t, page)

	if !strings.Contains(summary, `class="console-summary"`) {
		t.Fatalf("개요 요약이 세 묶음 구조가 아님:\n%s", summary)
	}
	for _, want := range []string{
		"엔진 상태", "계좌 보유", "오늘 실현손익", "살아 있는 주문", "안전·Guardian 한도", "최근 exit",
		`href="/verify-console"`, `href="/positions"`, `href="/history"`, `href="/orders"`,
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("개요 요약에 %q가 없음:\n%s", want, summary)
		}
	}
	for _, label := range []string{"KR", "US"} {
		if n := strings.Count(summary, label); n < 2 {
			t.Errorf("개요 요약이 시장 %s를 보유·오늘에 각각 분리하지 않음: %d회\n%s", label, n, summary)
		}
	}
	if !regexp.MustCompile(`미측정</span> · <code>[^<]+</code>`).MatchString(summary) {
		t.Errorf("미측정 칸에 사유 코드가 없음:\n%s", summary)
	}
}

func TestStructuralUnmeasuredIsExplainedOnce(t *testing.T) {
	h := newOverviewHarness(t)
	seedJournal(t, h.journal)
	page := h.page(t, "/dashboard")
	if got := strings.Count(page, "이 소진분은 엔진 프로세스가 메모리에서 계산해"); got != 1 {
		t.Fatalf("구조적 미측정 설명 = %d회, want 1회", got)
	}
}

func TestVerifyStepsAreFoldableAndListTextIsPreserved(t *testing.T) {
	p := verifyPage{Market: verifylive.MarketKR, Steps: a176StepText(t), ShowStart: true}
	items := p.StepItems()
	steps := verifylive.Steps()
	if len(items) != len(steps) {
		t.Fatalf("단계 조각 수 = %d, want %d", len(items), len(steps))
	}

	var rebuilt strings.Builder
	for i, item := range items {
		if i == 0 {
			rebuilt.WriteString(item.Preamble)
		}
		if item.ID != steps[i].ID {
			t.Errorf("조각 %d ID = %q, want %q", i, item.ID, steps[i].ID)
		}
		if !strings.HasPrefix(item.Header, string(steps[i].ID)) {
			t.Errorf("조각 %d 머리말이 %q로 시작하지 않음: %q", i, steps[i].ID, item.Header)
		}
		rebuilt.WriteString(item.Header)
		rebuilt.WriteByte('\n')
		rebuilt.WriteString(item.Body)
	}
	if got := rebuilt.String(); got != p.Steps {
		t.Fatalf("머리말과 조각을 이은 값이 원문과 다름\nwant:\n%s\ngot:\n%s", p.Steps, got)
	}

	rendered := a176RenderVerify(t, p)
	if got := strings.Count(rendered, `<details class="step">`); got != len(steps) {
		t.Errorf("접힌 단계 수 = %d, want %d", got, len(steps))
	}
	for _, item := range items {
		want := `<details class="step"><summary><code>` + html.EscapeString(item.Header) + `</code></summary><pre>` + html.EscapeString(item.Body) + `</pre></details>`
		if !strings.Contains(rendered, want) {
			t.Errorf("단계 본문이 접힌 영역 안에 보존되지 않음: %q", item.ID)
		}
	}

	missing := steps[len(steps)/2].ID
	boundary := "\n" + fmt.Sprintf("%-22s ", missing)
	broken := strings.Replace(p.Steps, boundary, "\n누락된 단계                  ", 1)
	brokenPage := verifyPage{Market: verifylive.MarketKR, Steps: broken, ShowStart: true}
	if got := brokenPage.StepItems(); got != nil {
		t.Fatalf("경계가 없는 문자열의 조각 = %#v, want nil", got)
	}
	fallback := a176RenderVerify(t, brokenPage)
	wantFallback := "<pre>" + html.EscapeString(broken) + "</pre>"
	if !strings.Contains(fallback, wantFallback) {
		t.Errorf("경계 오류가 종전 전체 pre 렌더로 돌아가지 않음:\n%s", fallback)
	}
}

func TestVerifyFormsAreByteIdentical(t *testing.T) {
	page := verifyPage{
		CSRF:      "a176-csrf",
		Market:    verifylive.MarketKR,
		Steps:     "fixture step text",
		Prompt:    "fixture prompt",
		Run:       &runView{Awaiting: true},
		ShowStart: true,
	}
	got := a176FormContract(t, a176RenderVerify(t, page))
	wantBytes, err := os.ReadFile("testdata/a176_verify_forms.golden.html")
	if err != nil {
		t.Fatal(err)
	}
	if want := string(wantBytes); got != want {
		t.Fatalf("검증 폼 마크업이 변경됨\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func a176StepText(t *testing.T) string {
	t.Helper()
	var steps strings.Builder
	verifylive.WriteSteps(&steps, false, false)
	return steps.String()
}

func a176RenderVerify(t *testing.T, page verifyPage) string {
	t.Helper()
	var rendered bytes.Buffer
	if err := pages.ExecuteTemplate(&rendered, "verify", page); err != nil {
		t.Fatal(err)
	}
	return rendered.String()
}

func a176Summary(t *testing.T, page string) string {
	t.Helper()
	start := strings.Index(page, `class="console-summary`)
	if start < 0 {
		t.Fatal("개요 요약이 없음")
	}
	end := strings.Index(page[start:], "</dl>")
	if end < 0 {
		t.Fatal("개요 요약의 끝이 없음")
	}
	return page[start : start+end+len("</dl>")]
}

func a176FormContract(t *testing.T, page string) string {
	t.Helper()
	const prompt = "<pre>fixture prompt</pre>"
	if !strings.Contains(page, prompt) {
		t.Fatalf("승인 프롬프트가 변경됨:\n%s", page)
	}
	var forms []string
	for start := 0; ; {
		begin := strings.Index(page[start:], "<form")
		if begin < 0 {
			break
		}
		begin += start
		end := strings.Index(page[begin:], "</form>")
		if end < 0 {
			t.Fatalf("폼 끝이 없음:\n%s", page[begin:])
		}
		end += begin + len("</form>")
		forms = append(forms, page[begin:end])
		start = end
	}
	if len(forms) == 0 {
		return ""
	}
	return prompt + "\n" + strings.Join(forms, "\n") + "\n"
}
