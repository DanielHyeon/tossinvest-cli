package console

import (
	"strings"
	"testing"
)

func TestA058DetailActionUsesTrailingPopupColumn(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	for _, path := range []string{"/positions", "/dashboard"} {
		page := h.page(t, path)
		for _, want := range []string{
			`<col class="position-col-detail">`,
			`<th scope="col">상세</th>`,
			`data-label="상세" class="detail-cell"`,
			`.position-col-detail { width: 7%; }`,
			`justify-content: space-between;`,
			`.detail-cell { min-height: 66px;`,
			`.detail-cell .row-details[open] { position: fixed;`,
			`.detail-cell .row-details[open] summary { position: sticky;`,
			`top: 0; min-height: 44px;`,
			`.detail-cell .row-details[open] .detail-grid { align-content: start;`,
		} {
			if !strings.Contains(page, want) {
				t.Errorf("%s trailing popup contract missing %q", path, want)
			}
		}

		row := positionHTMLRow(t, page, "005930")
		pnlAt := strings.Index(row, `data-label="미실현 PnL"`)
		detailAt := strings.Index(row, `data-label="상세"`)
		popupAt := strings.Index(row, `<details class="row-details">`)
		if pnlAt < 0 || detailAt < 0 || popupAt < 0 || !(pnlAt < detailAt && detailAt < popupAt) {
			t.Fatalf("%s detail action is not the final cell after PnL: %s", path, row)
		}
		if strings.Contains(strings.ToLower(row), "<script") || strings.Contains(strings.ToLower(row), "<input") {
			t.Fatalf("%s popup added script or input surface: %s", path, row)
		}
		if strings.Contains(page, "\n.row-details[open]") {
			t.Fatalf("%s popup styling escaped the holdings detail cell scope", path)
		}
	}
}
