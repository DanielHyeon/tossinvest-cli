package console

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

type fakePerformanceReader struct {
	view  performance.DashboardView
	query performance.Query
	calls int
}

func (f *fakePerformanceReader) Dashboard(_ context.Context, query performance.Query) (performance.DashboardView, error) {
	f.calls++
	f.query = query
	f.view.Query = query
	return f.view, nil
}

func TestPerformanceHistoryUsesOnlyServerFixedFiltersAndExplainsEveryMetric(t *testing.T) {
	reader := &fakePerformanceReader{view: performance.DashboardView{
		States: performance.StateCounts{Complete: 7, LinkMissing: 2, NotMeasured: 3, InsufficientSample: 1},
		Aggregates: []performance.Aggregate{{
			Market: "kr", LaneID: "krx_parker_vwap_conservative_v1", LaneVersion: "lane/v1",
			PolicyID: "COMMON_LADDER_BALANCED", PolicyVersion: "policy/v1", Samples: 7,
			Status: performance.StatusInsufficientSample, SemanticsVersion: performance.SemanticsVersion,
			ObservationProvenance: "existing-position@v1",
			Metrics: []performance.MetricSummary{{Key: "markout_5", Label: "5분 비용 후 markout",
				Help: "진입 뒤 기존 관측입니다.", Unit: "%", Value: "1.2", Samples: 6,
				Status: performance.StatusComplete, Provenance: "existing-position@v1"}},
		}},
	}}
	h := newDashboardHarness(t, func(options *Options) { options.Performance = reader })
	h.authenticate(t)
	page := h.page(t, "/performance-history?market=us&lane=invented&period=999&complete=false")
	for _, want := range []string{
		"최근 30일", "전체 시장", "전체 lane", "complete lineage only",
		"link_missing", "not_measured", "insufficient_sample", "추천 근거로 사용 불가",
		"5분 비용 후 markout", "진입 뒤 기존 관측입니다.", "6건", "existing-position@v1",
		`data-label="provenance"><code>existing-position@v1`,
		"lane-performance/v1", "주문, 설정 저장, lane 토글 또는 LIVE 승인",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("performance page missing %q", want)
		}
	}
	if reader.calls != 1 || reader.query.PeriodDays != 30 || reader.query.Market != performance.AllMarkets ||
		reader.query.Lane != performance.AllLanes || !reader.query.CompleteOnly {
		t.Fatalf("reader calls=%d query=%+v", reader.calls, reader.query)
	}
	for _, forbidden := range []string{"<input", "<button", "<form", "<textarea", "contenteditable", "type=number"} {
		if strings.Contains(strings.ToLower(page), forbidden) {
			t.Errorf("performance page contains input/mutation control %q", forbidden)
		}
	}
}

func TestPerformanceHistoryIsMethodReadOnlyMobileAccessibleAndCSPCompatible(t *testing.T) {
	reader := &fakePerformanceReader{view: performance.DashboardView{Aggregates: []performance.Aggregate{{
		LaneID: "lane", PolicyID: "policy", Status: performance.StatusComplete,
		Metrics: []performance.MetricSummary{{Label: "성과", Unit: "R", Value: "1", Samples: 1, Status: performance.StatusComplete}},
	}}}}
	h := newDashboardHarness(t, func(options *Options) { options.Performance = reader })
	h.authenticate(t)
	if response := h.post(t, "/performance-history", nil); response.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST status=%d want 405", response.StatusCode)
	}
	if reader.calls != 0 {
		t.Fatal("POST reached performance reader")
	}
	response := h.get(t, "/performance-history")
	body := body(t, response)
	for _, want := range []string{`name="viewport"`, `aria-label="고정된 성과 조회 범위"`, `scope="col"`, `class="data-table"`} {
		if !strings.Contains(body, want) {
			t.Errorf("accessible/mobile markup missing %q", want)
		}
	}
	const wantCSP = "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"
	if got := response.Header.Get("Content-Security-Policy"); got != wantCSP {
		t.Errorf("CSP=%q want=%q", got, wantCSP)
	}
	for _, forbidden := range []string{"<script", "onclick=", "onchange="} {
		if strings.Contains(strings.ToLower(body), forbidden) {
			t.Errorf("CSP-incompatible markup contains %q", forbidden)
		}
	}
}
