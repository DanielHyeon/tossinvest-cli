package console

import (
	"context"
	"errors"
	"net/http"

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

type performanceHistoryPage struct {
	chrome
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
