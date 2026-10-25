package console

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

// PerformanceReader is deliberately read-only. The console receives neither a
// performance Store nor a journal/broker/config handle, only this one fixed
// query operation.
type PerformanceReader interface {
	Dashboard(context.Context, performance.Query) (performance.DashboardView, error)
	AttributionRows(context.Context, string, performance.AttributionQuery, int) ([]performance.Attribution, error)
}

type performanceAttributionRow struct {
	AccountLabel string
	Row          performance.Attribution
}

// performanceWindow는 성과 조회 기본값(일수·시장·lane·lineage)의 표시 정본임.
// 성과 이력 화면과 최적화 카테고리가 같은 값 객체에서 같은 문구를 렌더함.
// Unmeasured가 차있으면 조회 기본값을 얻지 못한 상태 → 0일로 꾸미지 않고 사유 코드만 표시함.
type performanceWindow struct {
	Parts      []string
	Unmeasured string
}

// newPerformanceWindow는 seam 배선 여부와 서버 조회 기본값으로 표시용 창을 만듦.
func newPerformanceWindow(wired bool, q performance.Query) performanceWindow {
	if !wired {
		return performanceWindow{Unmeasured: string(reasonSeamUnwired)}
	}
	market, lane := "전체 시장", "전체 lane"
	if q.Market != performance.AllMarkets {
		market = q.Market
	}
	if q.Lane != performance.AllLanes {
		lane = q.Lane
	}
	lineage := "complete lineage only"
	if !q.CompleteOnly {
		lineage = "all lineage"
	}
	return performanceWindow{Parts: []string{fmt.Sprintf("최근 %d일", q.PeriodDays), market, lane, lineage}}
}

// Text는 칩이 아닌 한 줄 표기(최적화 카테고리용). 칩과 같은 Parts에서 만듦.
func (w performanceWindow) Text() string {
	if w.Unmeasured != "" {
		return "미측정 · " + w.Unmeasured
	}
	return strings.Join(w.Parts, " · ")
}

type performanceHistoryPage struct {
	chrome
	Window         performanceWindow
	Unwired        bool
	LoadErr        string
	AttributionErr string
	View           performance.DashboardView
	Attributions   []performanceAttributionRow
}

func (c *Console) handlePerformanceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다",
			"성과 이력은 조회만 한다. 주문·설정·lane·LIVE 승인을 변경하지 않는다.")
		return
	}
	page := performanceHistoryPage{chrome: c.chromeOnRequest("performance-history"), Unwired: c.opts.Performance == nil}
	page.Window = newPerformanceWindow(c.opts.Performance != nil, performance.DefaultQuery(c.now()))
	if c.opts.Performance != nil {
		view, err := c.opts.Performance.Dashboard(r.Context(), performance.DefaultQuery(c.now()))
		if err != nil {
			page.LoadErr = err.Error()
		} else {
			page.View = view
			page.Attributions, err = c.readPerformanceAttributions(r.Context())
			if err != nil {
				page.AttributionErr = err.Error()
			}
		}
	}
	c.render(w, "performance-history", page)
}

func (c *Console) readPerformanceAttributions(ctx context.Context) ([]performanceAttributionRow, error) {
	if c == nil || c.opts.Performance == nil || c.opts.JournalPath == "" {
		return nil, nil
	}
	reader, err := journal.OpenReadOnly(ctx, journal.ReadOnlyOptions{Path: c.opts.JournalPath})
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	accounts, err := reader.AccountRefs(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]performanceAttributionRow, 0)
	for _, accountRef := range accounts {
		for _, market := range []string{"KR", "US"} {
			values, readErr := c.opts.Performance.AttributionRows(ctx, accountRef,
				performance.AttributionQuery{Market: market, IncludeLinkMissing: true}, performance.MaxAttributionQueryRows/2)
			if errors.Is(readErr, performance.ErrAttributionUnavailable) {
				continue
			}
			if readErr != nil {
				return nil, readErr
			}
			for _, value := range values {
				rows = append(rows, performanceAttributionRow{AccountLabel: attest.Mask(accountRef), Row: value})
			}
		}
	}
	return rows, nil
}
