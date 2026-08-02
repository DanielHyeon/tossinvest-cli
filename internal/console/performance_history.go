package console

import (
	"context"
	"net/http"

	"github.com/JungHoonGhae/tossinvest-cli/internal/performance"
)

// PerformanceReader is deliberately read-only. The console receives neither a
// performance Store nor a journal/broker/config handle, only this one fixed
// query operation.
type PerformanceReader interface {
	Dashboard(context.Context, performance.Query) (performance.DashboardView, error)
}

type performanceHistoryPage struct {
	chrome
	Unwired bool
	LoadErr string
	View    performance.DashboardView
}

func (c *Console) handlePerformanceHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다",
			"성과 이력은 조회만 한다. 주문·설정·lane·LIVE 승인을 변경하지 않는다.")
		return
	}
	page := performanceHistoryPage{chrome: c.chromeOnRequest("performance-history"), Unwired: c.opts.Performance == nil}
	if c.opts.Performance != nil {
		view, err := c.opts.Performance.Dashboard(r.Context(), performance.DefaultQuery(c.now()))
		if err != nil {
			page.LoadErr = err.Error()
		} else {
			page.View = view
		}
	}
	c.render(w, "performance-history", page)
}
