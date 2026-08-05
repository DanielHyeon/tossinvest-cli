package console

// portfolio_refresh_test.go covers change refresh-positions-screen: the page-level
// notices that replace the per-row repetition of the two page-global reasons, and
// the positions screen's browser reload instruction.
//
// The claims are the ones the change exists for:
//
//	said once      a reason shared by every row in the same state is one notice,
//	               not one paragraph per holding — the repetition was the report.
//	label per row  dedup moves the sentence, never the verdict: every unmanaged
//	               row still carries 관리 외(미편입) (spec: 라벨은 행마다 표시).
//	reload bounded an open tab costs at most one holdings call per TTL. Change
//	               a080 moved where that ceiling is enforced: the cache holds it,
//	               so the reload period is the engine's cadence rather than the
//	               TTL; the verification screens keep their two-second cadence.

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// fourthHolding is a second broker-only symbol, so that repetition — one
// paragraph per unmanaged holding — is observable at all.
func fourthHolding() domain.Position {
	return domain.Position{Symbol: "066570", Name: "LG전자", MarketType: "KR", Quantity: 5,
		AveragePrice: 90000, CurrentPrice: 91000, MarketValue: 455000, UnrealizedPnL: 5000,
		ProfitRate: 0.0111}
}

// TestTheJournalAbsenceNoticeAppearsOncePerPage.
//
// Spec scenario "같은 사유의 수동 보유 다수": two broker-only holdings share one
// state, so the reason is one page-level notice — while the label stays on every
// row, because the notice explains and the label judges.
func TestTheJournalAbsenceNoticeAppearsOncePerPage(t *testing.T) {
	h := newDashboardHarness(t)
	h.holdings.rows = append(h.holdings.rows, fourthHolding())
	seedJournal(t, h.journal)
	h.authenticate(t)

	page := h.page(t, "/positions")

	const notice = "엔진 원장에 포지션이 없는 보유"
	if n := strings.Count(page, notice); n != 1 {
		t.Errorf("the journal-absence notice appears %d time(s), want exactly 1 — "+
			"a shared state is one sentence, not one per holding", n)
	}
	if strings.Contains(page, "엔진 원장에 이 종목의 포지션이 없다") {
		t.Error("the per-row spelling of the journal-absence reason is still rendered; " +
			"the reason moved to a page-level notice")
	}
	// 000660, 066570 (broker only) and 035420 (a journal position with neither
	// record) all keep the label on their own rows, and the legend still explains
	// it below the table.
	if n := strings.Count(page, "관리 외(미편입)"); n < 4 {
		t.Errorf("the unmanaged label appears %d time(s), want one per unmanaged row plus "+
			"the legend — dedup moves the reason, never the label", n)
	}
	// The row-specific reason of 035420 stays on its row.
	if !strings.Contains(page, "편입 기록(adoption)도 없는 포지션이다") {
		t.Error("the row-specific reason (a journal position with neither justifying record) " +
			"no longer renders; only the page-global reasons were to move")
	}
}

// TestTheUnknownStateNoticeAppearsOncePerPage.
//
// When the journal could not be read every row is 관리 여부 불명 for the same
// reason, and that reason is one notice — the per-row repetition said the same
// sentence three times before this change.
func TestTheUnknownStateNoticeAppearsOncePerPage(t *testing.T) {
	h := newDashboardHarness(t)
	// No journal is seeded: the file does not exist, so the ledger never answers.
	h.authenticate(t)

	page := h.page(t, "/positions")

	if n := strings.Count(page, "관리 중이 아니라고 단정하지 않는다"); n != 1 {
		t.Errorf("the unknown-state notice appears %d time(s), want exactly 1", n)
	}
	if n := strings.Count(page, "관리 여부 불명"); n < 3 {
		t.Errorf("the unknown label appears %d time(s), want one per broker row — "+
			"the verdict stays on the rows", n)
	}
}

// TestThePositionsScreenAsksTheBrowserToReloadAtTheEngineCadence.
//
// Spec "rate budget 보호", as a080 restates it: what the budget requires is that
// one open tab cost at most one holdings call per TTL, and that ceiling is held by
// the cache rather than by the reload period — holdingsCache.get refreshes only
// once its last attempt is a TTL old, so a shorter reload reaches the cache more
// often and the broker exactly as often. The period is therefore the engine's
// observation cadence, and the ceiling is asserted directly by
// TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling.
//
// The history screen renders frozen values and keeps no reload instruction.
func TestThePositionsScreenAsksTheBrowserToReloadAtTheEngineCadence(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)

	if page := h.page(t, "/positions"); !strings.Contains(page,
		`<meta http-equiv="refresh" content="`+strconv.Itoa(lineRefreshSeconds())+`">`) {
		t.Error("the positions screen does not ask the browser to reload at the engine cadence")
	}
	if page := h.page(t, "/history"); strings.Contains(page, `http-equiv="refresh"`) {
		t.Error("the history screen asks the browser to reload; it renders frozen values " +
			"and the change deliberately left it manual")
	}
}

// TestTheVerificationScreensKeepTheirTwoSecondReload.
//
// The head template's period became a parameter for the positions screen; the
// screens that follow a running verification keep their two-second cadence, and
// this pins it through the same template the handlers execute.
func TestTheVerificationScreensKeepTheirTwoSecondReload(t *testing.T) {
	for _, tc := range []struct {
		name string
		page any
	}{
		{"verify", verifyPage{chrome: chrome{Nav: "verify", Refresh: true}}},
		{"dashboard", dashboardPage{chrome: chrome{Nav: "verify-console", Refresh: true}}},
	} {
		var buf bytes.Buffer
		if err := pages.ExecuteTemplate(&buf, "head", tc.page); err != nil {
			t.Fatalf("rendering head for the %s page: %v", tc.name, err)
		}
		if !strings.Contains(buf.String(), `<meta http-equiv="refresh" content="2">`) {
			t.Errorf("the %s page no longer reloads every two seconds while a run is working",
				tc.name)
		}
	}
}
