package console

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

// a052: a seed-only exit row is useful operator evidence, but it is not an
// evaluated protection line. The screen must show both truths at once.
func TestPositionsSeparateStoredExitEvidenceFromEffectiveProtection(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
		VALUES ('pos-a052-raw','123-45-678901','us','A052RAW',1,'decision-a052','OPEN','2','123.45','2026-07-27T00:30:00Z');
		INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
		 high_water,ratchet_level,taken_ratio_total,completed,updated_at)
		VALUES ('pos-a052-raw','RATCHET','123.45','117.27','6.18','118.88','130.01','NONE','0',0,'2026-07-27T00:59:00Z');`)

	h.authenticate(t)
	row := positionHTMLRow(t, h.page(t, "/positions"), "A052RAW")
	for _, want := range []string{
		"원장 기록 · 실효 미확인",
		"t0 진입가 <strong>123.45</strong>",
		"최초 손절 <strong>117.27</strong>",
		"원장 기준선 <strong>118.88</strong>",
		"원장 high-water <strong>130.01</strong>",
		"현재 보호선 <strong>—</strong>",
		"다음 익절 <strong>—</strong>",
	} {
		if !strings.Contains(row, want) {
			t.Errorf("seed-only position row lacks %q: %s", want, row)
		}
	}
}

func TestA052TradingViewsStay375pxResponsiveAndInputFree(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState(), runtime: positionpolicy.ManagementRuntime{
		Effective: positionpolicy.NewAdoptionSettings(false, .05, nil, nil, ""), EffectiveKnown: true,
	}}
	h := newDashboardHarness(t, func(o *Options) { o.PositionPolicies = commander })
	seedJournal(t, h.journal)
	h.authenticate(t)
	for _, path := range []string{"/positions", "/position-management"} {
		page := h.page(t, path)
		for _, responsive := range []string{`name="viewport"`, `@media (max-width: 720px)`,
			`.data-table tr > * { width: 100% !important; min-width: 0; }`} {
			if !strings.Contains(page, responsive) {
				t.Errorf("%s lacks the mobile contract covering a 375px viewport: %q", path, responsive)
			}
		}
		for _, banned := range []string{`type="text"`, `type="number"`, "<textarea", "contenteditable",
			`action="/reconcile`, `name="reason"`} {
			if strings.Contains(strings.ToLower(page), strings.ToLower(banned)) {
				t.Errorf("%s contains free-form/reconcile surface %q", path, banned)
			}
		}
		for _, input := range regexp.MustCompile(`<input[^>]*>`).FindAllString(page, -1) {
			if !strings.Contains(input, `type="hidden"`) {
				t.Errorf("%s contains visible input: %s", path, input)
			}
		}
	}
}

func TestPositionManagementShowsActualDesiredEffectiveAndEscapedBlocks(t *testing.T) {
	state := managedPolicyState()
	state.Market, state.Symbol, state.ExitEligible = "us", "AAPL", false
	runtime := positionpolicy.ManagementRuntime{
		Effective:      positionpolicy.NewAdoptionSettings(false, .05, []string{"AAPL"}, nil, ""),
		EffectiveKnown: true,
		BlockSource:    positionpolicy.AdoptionBlockingTrackerProjection,
		Blocks: []positionpolicy.ReconcileBlock{{
			Scope: positionpolicy.ScopeAccount, Reason: "QUANTITY_MISMATCH",
			Detail:    "<script>alert('unsafe')</script>",
			StartedAt: time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC), Permanent: true,
		}},
	}
	commander := &fakePositionPolicyCommander{state: state, runtime: runtime}
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = commander
		o.Settings = &fakeSettings{block: config.Adoption{Enabled: true, DefaultStopPct: .03,
			IncludeSymbols: []string{"AAPL"}, ExcludeSymbols: []string{"TSLA"}}}
	})
	h.authenticate(t)
	page := h.page(t, "/position-management")
	for _, want := range []string{
		"기본값", "저장값 · desired", "실행값 · effective", "ON", "3%", "OFF", "5%",
		"AAPL", "TSLA", "QUANTITY_MISMATCH", "계좌 전체", "영구 차단", "대사 차단",
		positionpolicy.AdoptionBlockingTrackerProjection,
		"&lt;script&gt;alert(&#39;unsafe&#39;)&lt;/script&gt;",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("position-management lacks %q", want)
		}
	}
	for _, banned := range []string{"편입 불가", "<script>alert", `type="text"`, `type="number"`, "<textarea", "contenteditable"} {
		if strings.Contains(strings.ToLower(page), strings.ToLower(banned)) {
			t.Errorf("position-management contains forbidden %q", banned)
		}
	}
}

func TestPositionManagementKeepsEffectiveUnknownWhenRuntimeFails(t *testing.T) {
	commander := &fakePositionPolicyCommander{state: managedPolicyState(), runtimeError: errors.New("engine socket unavailable")}
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = commander
		o.Settings = &fakeSettings{block: config.Adoption{Enabled: true, DefaultStopPct: .03,
			IncludeSymbols: []string{"AAPL"}}}
	})
	h.authenticate(t)
	page := h.page(t, "/position-management")
	if !strings.Contains(page, "저장값 · desired") || !strings.Contains(page, "ON") ||
		!strings.Contains(page, "3%") || !strings.Contains(page, "실행값 · effective") ||
		!strings.Contains(page, "알 수 없음") || !strings.Contains(page, "engine socket unavailable") {
		t.Fatalf("desired/effective unknown contract missing: %s", page)
	}
}

func TestPositionsUseSharedReconcileBlockedStatusForUSCandidate(t *testing.T) {
	state := managedPolicyState()
	state.PositionID, state.Market, state.Symbol, state.ExitEligible = "pos-a052-us", "us", "AAPL", false
	runtime := positionpolicy.ManagementRuntime{
		Effective:      positionpolicy.NewAdoptionSettings(false, .03, []string{"AAPL"}, nil, ""),
		EffectiveKnown: true,
		Blocks: []positionpolicy.ReconcileBlock{{Scope: positionpolicy.ScopeAccount,
			Reason: "QUANTITY_MISMATCH", Detail: "holdings quantity differs", Permanent: true}},
		BlockSource: positionpolicy.AdoptionBlockingTrackerProjection,
	}
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = &fakePositionPolicyCommander{state: state, runtime: runtime}
		o.Settings = &fakeSettings{block: config.Adoption{DefaultStopPct: .03, IncludeSymbols: []string{"AAPL"}}}
	})
	h.holdings.rows = append(h.holdings.rows, domain.Position{MarketType: "US", Symbol: "AAPL", Name: "Apple",
		Quantity: 1, AveragePrice: 200, CurrentPrice: 201, MarketValue: 201, UnrealizedPnL: 1, ProfitRate: .005})
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES ('pos-a052-us','123-45-678901','us','AAPL',1,'OPEN','1','200','2026-07-27T00:30:00Z');`)
	h.authenticate(t)
	row := positionHTMLRow(t, h.page(t, "/positions"), "AAPL")
	for _, want := range []string{"대사 차단", "QUANTITY_MISMATCH", "holdings quantity differs", "관리 편입"} {
		if !strings.Contains(row, want) {
			t.Errorf("US blocked row lacks %q: %s", want, row)
		}
	}
	if strings.Contains(row, "편입 불가") || strings.Contains(row, "미국 시장 미지원") {
		t.Errorf("US row claims unsupported market: %s", row)
	}
}

func TestReleasedLifecycleWinsAcrossTradingViews(t *testing.T) {
	state := managedPolicyState()
	state.PositionID, state.Market, state.Symbol = "pos-a052-released", "us", "A052REL"
	state.Status = positionpolicy.StatusReleased
	state.ExitEligible = true
	state.Provenance = positionpolicy.ProvenanceExternalAdoption
	state.Eligibility = positionpolicy.EligibilityExternalLifecycle
	runtime := positionpolicy.ManagementRuntime{
		Effective:      positionpolicy.NewAdoptionSettings(true, .03, []string{"A052REL"}, nil, ""),
		EffectiveKnown: true,
		Blocks: []positionpolicy.ReconcileBlock{{Scope: positionpolicy.ScopeAccount,
			Reason: "QUANTITY_MISMATCH", Permanent: true}},
		BlockSource: positionpolicy.AdoptionBlockingTrackerProjection,
	}
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = &fakePositionPolicyCommander{state: state, runtime: runtime}
	})
	h.holdings.rows = append(h.holdings.rows, domain.Position{MarketType: "US", Symbol: "A052REL",
		Name: "Released", Quantity: 1, AveragePrice: 100, CurrentPrice: 101, MarketValue: 101})
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,adoption_id,state,quantity,avg_price,opened_at)
		VALUES ('pos-a052-released','123-45-678901','us','A052REL',1,'adopt-a052','OPEN','1','100','2026-07-27T00:30:00Z');
		INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
		 high_water,ratchet_level,taken_ratio_total,completed,updated_at)
		VALUES ('pos-a052-released','RATCHET','70000','68000','2000','68000','70000','NONE','0',0,'2026-07-27T00:59:00Z');`)
	line, recovery := ratchetViewSnapshot(t, "pos-a052-released", 1, "1", "obs-a052-release",
		"71000", "71000", "68000", "0", exitpolicy.LevelNone)
	writeViewSnapshot(t, h.journal, line, recovery, "2026-07-27T00:59:50Z")
	h.authenticate(t)

	for _, tc := range []struct {
		path      string
		positions bool
	}{
		{path: "/positions", positions: true},
		{path: "/position-management"},
	} {
		row := h.page(t, tc.path)
		if tc.positions {
			row = positionHTMLRow(t, row, "A052REL")
		}
		for _, want := range []string{"관리 외(운영자 해제)", "UNMANAGED", "OPERATOR_RELEASED"} {
			if !strings.Contains(row, want) {
				t.Errorf("%s released row lacks %q: %s", tc.path, want, row)
			}
		}
		for _, banned := range []string{"엔진 관리", "대사 차단으로 대기", "ADOPTION_PENDING", "RECONCILE_BLOCKED"} {
			if strings.Contains(row, banned) {
				t.Errorf("%s released row contains %q: %s", tc.path, banned, row)
			}
		}
		if tc.positions && (!strings.Contains(row, "현재 보호선 <strong>—</strong>") ||
			strings.Contains(row, line.DecisionID)) {
			t.Errorf("released row exposes actionable snapshot: %s", row)
		}
	}
}

func TestVirtualReleasedDefaultIsNotAnOperatorRelease(t *testing.T) {
	state := managedPolicyState()
	state.PositionID, state.Market, state.Symbol = "pos-a052-virtual", "us", "A052VIRTUAL"
	state.Status, state.Version, state.ExitEligible = positionpolicy.StatusReleased, 0, false
	runtime := positionpolicy.ManagementRuntime{EffectiveKnown: true,
		Effective: positionpolicy.NewAdoptionSettings(false, .05, nil, nil, "")}
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = &fakePositionPolicyCommander{state: state, runtime: runtime}
	})
	h.holdings.rows = append(h.holdings.rows, domain.Position{MarketType: "US", Symbol: "A052VIRTUAL",
		Name: "Virtual", Quantity: 1, AveragePrice: 100, CurrentPrice: 101, MarketValue: 101})
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
		VALUES ('pos-a052-virtual','123-45-678901','us','A052VIRTUAL',1,'OPEN','1','100','2026-07-27T00:30:00Z');`)
	h.authenticate(t)
	for _, page := range []string{
		positionHTMLRow(t, h.page(t, "/positions"), "A052VIRTUAL"),
		h.page(t, "/position-management"),
	} {
		if !strings.Contains(page, "관리 외(미편입)") || !strings.Contains(page, "NOT_SELECTED") ||
			strings.Contains(page, "OPERATOR_RELEASED") {
			t.Errorf("virtual release default misclassified: %s", page)
		}
	}
}
