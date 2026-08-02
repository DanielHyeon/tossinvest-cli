package console

import (
	"strings"
	"testing"
)

func TestA058HoldingsUseStockOSDensityAndExplicitGrid(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	_ = h.page(t, "/positions")

	for _, path := range []string{"/positions", "/dashboard"} {
		page := h.page(t, path)
		for _, want := range []string{
			`<colgroup>`,
			`<col class="position-col-instrument">`,
			`<col class="position-col-quantity">`,
			`<col class="position-col-average">`,
			`<col class="position-col-current">`,
			`<col class="position-col-lines">`,
			`<col class="position-col-value">`,
			`<col class="position-col-pnl">`,
			`.position-col-instrument { width: 31%; }`,
			`.position-col-quantity { width: 6%; }`,
			`.position-col-average, .position-col-current { width: 11%; }`,
			`.position-col-lines { width: 18%; }`,
			`.position-col-value { width: 12%; }`,
			`.position-col-pnl { width: 11%; }`,
			`.positions-table { font-size: 12px; line-height: 18px; }`,
			`.positions-table thead { font-size: 10px; line-height: 15px; }`,
			`.positions-table th, .positions-table td { padding: 8px 12px; }`,
			`class="number-column" scope="col">수량</th>`,
			`.positions-table .number-column { text-align: right; font-variant-numeric: tabular-nums; }`,
		} {
			if !strings.Contains(page, want) {
				t.Errorf("%s compact holdings contract missing %q", path, want)
			}
		}
	}
}

func TestA058VerboseExitReasonIsOnlyInDetails(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	page := h.page(t, "/positions")
	row := positionHTMLRow(t, page, "005930")

	detailAt := strings.Index(row, `<details class="row-details">`)
	if detailAt < 0 {
		t.Fatal("holding row has no native details disclosure")
	}
	primary, detail := row[:detailAt], row[detailAt:]
	const verbose = "이전 원장에는 exit snapshot 근거가 없다"
	if strings.Contains(primary, verbose) {
		t.Fatalf("verbose exit reason still expands the primary scan row: %s", primary)
	}
	if !strings.Contains(detail, verbose) {
		t.Fatalf("details disclosure lost the verbose exit reason: %s", detail)
	}
	for _, want := range []string{"저장된 원장 기준선", "현재 실효 미확인", "익절", "손절", "추적 회수", "기준", "고점"} {
		if !strings.Contains(primary, want) {
			t.Errorf("primary row lost concise safety fact %q: %s", want, primary)
		}
	}
}
