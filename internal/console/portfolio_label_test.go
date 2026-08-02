package console

// portfolio_label_test.go pins the status presentation after the 2026-07-27 UX
// decision (console-adoption-controls, 포지션 가시성 delta). The compact
// holdings table keeps the adoption state beside the instrument instead of in
// a dedicated column, and an unmanaged row's one label follows its setting —
// 관리 외(미편입) before designation, 관리 편입 after. The latter is a
// designation, not protection, so the 편입 예약됨 note stays beside it.

import (
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// TestTheStatusColumnHeaderSaysAdoption preserves the adoption status while the
// compact seven-column table moves it under 종목 and dedicates 라인 to the five
// canonical exit references.
func TestTheStatusColumnHeaderSaysAdoption(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	for _, header := range []string{`<th scope="col">종목</th>`, `<th scope="col">라인</th>`} {
		if !strings.Contains(page, header) {
			t.Errorf("the compact holdings table is missing %s", header)
		}
	}
	if strings.Contains(page, `<th scope="col">관리 편입</th>`) {
		t.Error("adoption status still consumes a dedicated table column")
	}
	row := rowFor(t, page, "000660")
	if !strings.Contains(row, "관리 외(미편입)") {
		t.Errorf("the compact instrument cell lost the adoption status:\n%s", row)
	}
}

// TestAnUnmanagedRowsLabelFollowsItsCheckbox: before designation the row says
// 관리 외(미편입); after the explicit action, that same row's label is 관리 편입 — one
// label per row, switched by the stored setting, with the reservation note
// still present because the label reports a designation and not protection.
func TestAnUnmanagedRowsLabelFollowsItsCheckbox(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	row := rowFor(t, h.page(t, "/positions"), "000660")
	if !strings.Contains(row, "관리 외(미편입)") {
		t.Errorf("an undesignated unmanaged row must be labelled 관리 외(미편입):\n%s", row)
	}

	h.post(t, "/settings/include", url.Values{"csrf": {h.csrf}, "symbol": {"000660"}})

	row = rowFor(t, h.page(t, "/positions"), "000660")
	if strings.Contains(row, "관리 외(미편입)") {
		t.Errorf("a designated row still carries the unmanaged label:\n%s", row)
	}
	if !strings.Contains(row, "관리 편입") {
		t.Errorf("a designated row's label must read 관리 편입:\n%s", row)
	}
	if !strings.Contains(row, "편입 예약됨") {
		t.Errorf("the reservation note must stay beside the checked label:\n%s", row)
	}
}
