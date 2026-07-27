package console

// portfolio_label_test.go pins the status column after the 2026-07-27 UX
// decision (console-adoption-controls, 포지션 가시성 delta): the column header
// says 관리 편입, and an unmanaged row's one label follows its checkbox — 관리
// 외(미편입) unchecked, 관리 편입 checked. The checked label is a designation,
// not protection, so the 편입 예약됨 note stays beside it.

import (
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

// TestTheStatusColumnHeaderSaysAdoption: the header is 관리 편입, in those
// words — the old 관리 header is gone.
func TestTheStatusColumnHeaderSaysAdoption(t *testing.T) {
	seam := &fakeSettings{block: config.Adoption{DefaultStopPct: 0.05}}
	h := settingsHarness(t, seam)
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")
	if !strings.Contains(page, "<th>관리 편입</th>") {
		t.Error("the status column header does not say 관리 편입")
	}
	if strings.Contains(page, "<th>관리</th>") {
		t.Error("the old 관리 column header is still rendered")
	}
}

// TestAnUnmanagedRowsLabelFollowsItsCheckbox: before designation the row says
// 관리 외(미편입); after checking, that same row's label is 관리 편입 — one
// label per row, switched by the checkbox state, with the reservation note
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
