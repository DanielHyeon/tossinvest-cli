package console

import (
	"regexp"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

func TestA057DashboardAndPositionsShareSimpleHoldingProjection(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	line, recovery := ratchetViewSnapshot(t, "pos-managed", 1, "10", "obs-a057",
		"72000", "74000", "70100", "0.25", exitpolicy.LevelHalfRisk)
	writeViewSnapshot(t, h.journal, line, recovery, "2026-07-27T00:59:50Z")
	h.holdings.rows = append(h.holdings.rows, domain.Position{
		MarketType: "US", Symbol: "AAPL", Name: "Apple", Quantity: 2,
		AveragePrice: 200, CurrentPrice: 210, MarketValue: 420, UnrealizedPnL: 20, ProfitRate: .05,
	})
	h.authenticate(t)

	positions := h.page(t, "/positions") // fills the lazy holdings cache once
	beforeDashboard := h.holdings.count()
	dashboard := h.page(t, "/dashboard")
	if got := h.holdings.count(); got != beforeDashboard {
		t.Fatalf("dashboard refreshed broker holdings: calls %d -> %d", beforeDashboard, got)
	}
	if !strings.Contains(dashboard, "엔진 원장에 포지션이 없는 보유가 있다") {
		t.Error("dashboard does not explain why a broker-only holding has no protection lines")
	}

	for path, page := range map[string]string{"/positions": positions, "/dashboard": dashboard} {
		for _, header := range []string{"종목", "수량", "평균가", "현재가", "라인", "총금액", "미실현 PnL"} {
			if !strings.Contains(page, `scope="col">`+header+`</th>`) {
				t.Errorf("%s simple holdings header missing %q", path, header)
			}
		}

		row := positionHTMLRow(t, page, "005930")
		detailAt := strings.Index(row, `<details class="row-details">`)
		if detailAt < 0 {
			t.Fatalf("%s managed holding has no secondary disclosure: %s", path, row)
		}
		primary, detail := row[:detailAt], row[detailAt:]
		for _, want := range []string{
			"삼성전자", "005930", `data-label="수량"`, "10", `data-label="평균가"`,
			`data-label="현재가"`, "익절", line.NextTarget, "손절", line.InitialStop,
			"추적 회수", line.NextProtection, "기준", line.CurrentProtection,
			"고점", line.HighWater, `data-label="총금액"`, `data-label="미실현 PnL"`,
			"엔진 관리", "평가 완료",
		} {
			if !strings.Contains(primary, want) {
				t.Errorf("%s primary holding row lacks %q: %s", path, want, primary)
			}
		}
		for _, hidden := range []string{line.DecisionID, line.SnapshotID, line.ObservationID, "자격 근거", "원장 수량"} {
			if strings.Contains(primary, hidden) {
				t.Errorf("%s primary row exposes secondary detail %q: %s", path, hidden, primary)
			}
			if !strings.Contains(detail, hidden) {
				t.Errorf("%s disclosure lost secondary detail %q: %s", path, hidden, detail)
			}
		}
		if !strings.Contains(detail, `<summary>상세 보기</summary>`) {
			t.Errorf("%s disclosure lacks concise summary", path)
		}
		if !strings.Contains(detail, "계좌") || !strings.Contains(detail, "*********8901") {
			t.Errorf("%s disclosure lost the masked account identity: %s", path, detail)
		}

		us := positionHTMLRow(t, page, "AAPL")
		for _, want := range []string{"Apple", "AAPL", "US", "USD", `data-label="수량"`,
			`data-label="평균가"`, `data-label="현재가"`, `data-label="총금액"`, `data-label="미실현 PnL"`} {
			if !strings.Contains(us, want) {
				t.Errorf("%s US holding lacks common hierarchy %q: %s", path, want, us)
			}
		}
	}
}

func TestA057RawExitEvidenceStaysOutOfPrimaryLines(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	positions := h.page(t, "/positions")
	dashboard := h.page(t, "/dashboard")

	for path, page := range map[string]string{"/positions": positions, "/dashboard": dashboard} {
		row := positionHTMLRow(t, page, "005930")
		detailAt := strings.Index(row, `<details class="row-details">`)
		if detailAt < 0 {
			t.Fatalf("%s legacy holding has no disclosure", path)
		}
		primary, detail := row[:detailAt], row[detailAt:]
		for _, raw := range []string{"69500", "74000", "HALF_RISK", "intent-77"} {
			if strings.Contains(primary, raw) {
				t.Errorf("%s promotes raw evidence %q into the primary line: %s", path, raw, primary)
			}
		}
		for _, want := range []string{"익절", "손절", "추적 회수", "기준", "고점", "근거 없음"} {
			if !strings.Contains(primary, want) {
				t.Errorf("%s unknown primary line lacks %q: %s", path, want, primary)
			}
		}
		for _, want := range []string{"실효 미확인", "69500", "74000"} {
			if !strings.Contains(detail, want) {
				t.Errorf("%s disclosure lost stored evidence %q: %s", path, want, detail)
			}
		}
	}
}

func TestA057HoldingsViewsStayInputFreeAccessibleAndResponsive(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	_ = h.page(t, "/positions")

	for _, path := range []string{"/dashboard", "/positions"} {
		page := h.page(t, path)
		for _, want := range []string{
			`<caption>보유 종목과 보호 상태</caption>`, `scope="col"`, `scope="row"`, `class="data-table positions-table"`,
			`.row-details summary { min-height: 44px`, `@media (max-width: 720px)`,
			`.data-table tr > * { width: 100% !important; min-width: 0; }`, `:focus-visible`,
			`.data-table td > *, .data-table th[scope="row"] > * { grid-column: 2; min-width: 0; }`,
			`font-size: 0.9rem; font-variant-numeric: tabular-nums`,
			`.detail-grid { display: grid; grid-template-columns: 1fr;`,
		} {
			if !strings.Contains(page, want) {
				t.Errorf("%s accessibility/responsive contract missing %q", path, want)
			}
		}
		lower := strings.ToLower(page)
		for _, banned := range []string{"<form", "<input", "<textarea", "<select", "contenteditable", "<script", "javascript:"} {
			if strings.Contains(lower, banned) {
				t.Errorf("%s contains input/script surface %q", path, banned)
			}
		}
		for _, input := range regexp.MustCompile(`<input[^>]*>`).FindAllString(page, -1) {
			if !strings.Contains(input, `type="hidden"`) {
				t.Errorf("%s contains visible input: %s", path, input)
			}
		}
	}
}
