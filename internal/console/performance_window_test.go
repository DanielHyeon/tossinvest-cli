package console

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

// 두 화면이 같은 조회 창 문구를 렌더하는지 보는 시험(a177).
// 성과 이력은 칩 4개, 최적화 카테고리는 한 줄이므로 " · " 로 이어 붙여 비교함.
var windowChipRE = regexp.MustCompile(`<span class="status-pill">([^<]*)</span>`)

func historyWindowChips(page string) string {
	// 서버 조회 기본값 섹션 안의 칩만 취함
	start := strings.Index(page, "서버 조회 기본값")
	end := strings.Index(page, "화면 필터일 뿐")
	if start < 0 || end < start {
		return ""
	}
	var parts []string
	for _, m := range windowChipRE.FindAllStringSubmatch(page[start:end], -1) {
		parts = append(parts, strings.TrimSpace(m[1]))
	}
	return strings.Join(parts, " · ")
}

func optimizationWindowLine(page string) string {
	m := regexp.MustCompile(`<strong>([^<]*) · 읽기 전용</strong>`).FindStringSubmatch(page)
	if m == nil {
		return ""
	}
	return m[1]
}

func TestPerformanceWindowIsSingleSourceAcrossScreens(t *testing.T) {
	t.Run("wired", func(t *testing.T) {
		h := newDashboardHarness(t, func(o *Options) { o.Performance = &fakePerformanceReader{} })
		h.authenticate(t)
		history := historyWindowChips(h.page(t, "/performance-history"))
		optimization := optimizationWindowLine(h.page(t, "/optimization?category=performance-history"))
		if len(history) == 0 || len(optimization) == 0 {
			t.Fatalf("창 문구가 렌더되지 않음: history=%q optimization=%q", history, optimization)
		}
		const want = "최근 30일 · 전체 시장 · 전체 lane · complete lineage only"
		if history != want || optimization != want {
			t.Fatalf("창 문구 불일치: history=%q optimization=%q want=%q", history, optimization, want)
		}
	})
	t.Run("unwired", func(t *testing.T) {
		h := newDashboardHarness(t)
		h.authenticate(t)
		history := historyWindowChips(h.page(t, "/performance-history"))
		optimization := optimizationWindowLine(h.page(t, "/optimization?category=performance-history"))
		if len(history) == 0 || len(optimization) == 0 {
			t.Fatalf("창 문구가 렌더되지 않음: history=%q optimization=%q", history, optimization)
		}
		want := "미측정 · " + string(reasonSeamUnwired)
		if history != want || optimization != want {
			t.Fatalf("미측정 형식 불일치: history=%q optimization=%q want=%q", history, optimization, want)
		}
		if strings.Contains(history, "0일") || strings.Contains(optimization, "0일") {
			t.Fatal("미측정 상태가 0일로 꾸며짐")
		}
	})
	// 변경 전 코드(칩이 .View.Query.PeriodDays 를 찍던 때)였다면 LoadErr 에서 "최근 0일" 로 FAIL 했을 시험임.
	t.Run("loaderr", func(t *testing.T) {
		h := newDashboardHarness(t, func(o *Options) { o.Performance = &fakePerformanceReader{err: errors.New("boom")} })
		h.authenticate(t)
		history := historyWindowChips(h.page(t, "/performance-history"))
		optimization := optimizationWindowLine(h.page(t, "/optimization?category=performance-history"))
		if len(history) == 0 || len(optimization) == 0 {
			t.Fatalf("창 문구가 렌더되지 않음: history=%q optimization=%q", history, optimization)
		}
		const want = "최근 30일 · 전체 시장 · 전체 lane · complete lineage only"
		if history != want || optimization != want {
			t.Fatalf("LoadErr 창 문구 불일치: history=%q optimization=%q want=%q", history, optimization, want)
		}
		if strings.Contains(history, "최근 0일") || strings.Contains(optimization, "최근 0일") {
			t.Fatal("LoadErr 상태가 0일로 찍힘")
		}
	})
}
