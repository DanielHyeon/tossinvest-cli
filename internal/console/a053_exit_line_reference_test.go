package console

import (
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

func TestPositionsLeadWithManagedLegacyReferenceAcrossKRAndUS(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	for _, row := range []struct{ id, market, symbol, entry, stop, baseline, high string }{
		{"pos-a053-kr", "kr", "A053KR", "4095", "3890.25", "3890.25", "4135"},
		{"pos-a053-us", "us", "A053US", "36.44", "34.618", "36.87728", "37.6758"},
	} {
		execRawArgs(t, h.journal, `
			INSERT INTO positions(id,account_ref,market,symbol,instance_seq,adoption_id,state,quantity,avg_price,opened_at)
			VALUES (?,?,?,?,1,?,'OPEN','2',?,'2026-07-31T00:00:00Z');`,
			row.id, "123-45-678901", row.market, row.symbol, "adopt-"+row.id, row.entry)
		execRawArgs(t, h.journal, `
			INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
			 high_water,ratchet_level,taken_ratio_total,completed,updated_at,lifecycle_generation)
			VALUES (?,'LADDER',?,?, '1',?,?,'NONE','0',0,'2026-07-31T00:01:00Z',1);`,
			row.id, row.entry, row.stop, row.baseline, row.high)
	}
	h.authenticate(t)
	page := h.page(t, "/positions")
	for _, tc := range []struct{ symbol, market, currency, stop, baseline string }{
		{"A053KR", "KR", "KRW", "3890.25", "3890.25"},
		{"A053US", "US", "USD", "34.618", "36.87728"},
	} {
		row := positionHTMLRow(t, page, tc.symbol)
		for _, want := range []string{tc.market, tc.currency, "저장된 원장 기준선 · 현재 실효 미확인",
			"손절 <strong>" + tc.stop + "</strong>", "기준 <strong>" + tc.baseline + "</strong>",
			"익절 <strong>—</strong>"} {
			if !strings.Contains(row, want) {
				t.Errorf("%s row lacks %q: %s", tc.symbol, want, row)
			}
		}
		if strings.Contains(row, "현재 보호선 <strong>"+tc.baseline+"</strong>") {
			t.Errorf("%s legacy baseline became actionable: %s", tc.symbol, row)
		}
	}
}

func TestUSPendingAndBlockedShowEffectiveStopPlanWithoutPrice(t *testing.T) {
	for _, tc := range []struct {
		name   string
		blocks []positionpolicy.ReconcileBlock
		status string
	}{
		{name: "pending", status: "편입 예약됨"},
		{name: "blocked", blocks: []positionpolicy.ReconcileBlock{{Scope: positionpolicy.ScopeAccount,
			Reason: "QUANTITY_MISMATCH", Permanent: true}}, status: "대사 차단"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := managedPolicyState()
			state.PositionID, state.Market, state.Symbol, state.ExitEligible = "pos-a053-plan", "us", "AAPL", false
			runtime := positionpolicy.ManagementRuntime{EffectiveKnown: true,
				Effective: positionpolicy.NewAdoptionSettings(false, .03, []string{"AAPL"}, nil, ""),
				Blocks:    tc.blocks}
			h := newDashboardHarness(t, func(o *Options) {
				o.PositionPolicies = &fakePositionPolicyCommander{state: state, runtime: runtime}
				o.Settings = &fakeSettings{block: config.Adoption{DefaultStopPct: .07, IncludeSymbols: []string{"AAPL"}}}
			})
			h.holdings.rows = append(h.holdings.rows, domain.Position{MarketType: "US", Symbol: "AAPL",
				Quantity: 1, AveragePrice: 200, CurrentPrice: 201, MarketValue: 201})
			seedJournal(t, h.journal)
			execRaw(t, h.journal, `INSERT INTO positions(id,account_ref,market,symbol,instance_seq,state,quantity,avg_price,opened_at)
				VALUES ('pos-a053-plan','123-45-678901','us','AAPL',1,'OPEN','1','200','2026-08-01T00:00:00Z');`)
			h.authenticate(t)
			row := positionHTMLRow(t, h.page(t, "/positions"), "AAPL")
			for _, want := range []string{"US", "USD", tc.status, "기준선 미생성 · 엔진 보호 미적용",
				"현재 실행 중 엔진 정책: 최초 손절폭 <strong>3%</strong>", "가격은 편입 시 확정"} {
				if !strings.Contains(row, want) {
					t.Errorf("row lacks %q: %s", want, row)
				}
			}
			for _, synthetic := range []string{"194", "194.97", "187", "187.23", "7%"} {
				if strings.Contains(row, synthetic) {
					t.Errorf("row contains desired/synthetic value %q: %s", synthetic, row)
				}
			}
		})
	}
}

func TestPositionsSuppressCrossLifecycleExitEvidence(t *testing.T) {
	state := managedPolicyState()
	state.PositionID, state.Market, state.Symbol, state.AdoptionGeneration = "pos-a053-gen", "us", "GENUS", 2
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = &fakePositionPolicyCommander{state: state, runtime: positionpolicy.ManagementRuntime{
			EffectiveKnown: true, Effective: positionpolicy.NewAdoptionSettings(true, .03, nil, nil, ""),
		}}
	})
	seedJournal(t, h.journal)
	execRaw(t, h.journal, `
		INSERT INTO positions(id,account_ref,market,symbol,instance_seq,adoption_id,state,quantity,avg_price,opened_at)
		VALUES ('pos-a053-gen','123-45-678901','us','GENUS',1,'adopt-old','OPEN','1','100','2026-07-31T00:00:00Z');
		INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
		 high_water,ratchet_level,taken_ratio_total,completed,updated_at,lifecycle_generation)
		VALUES ('pos-a053-gen','LADDER','SENTINEL_ENTRY','SENTINEL_STOP','1','SENTINEL_BASELINE',
		 'SENTINEL_HIGH','NONE','0',0,'2026-07-31T00:01:00Z',1);`)
	h.authenticate(t)
	row := positionHTMLRow(t, h.page(t, "/positions"), "GENUS")
	if !strings.Contains(row, "기준선 근거 세대 불일치") {
		t.Fatalf("mismatch explanation absent: %s", row)
	}
	for _, forbidden := range []string{"SENTINEL_ENTRY", "SENTINEL_STOP", "SENTINEL_BASELINE", "SENTINEL_HIGH"} {
		if strings.Contains(row, forbidden) {
			t.Errorf("cross-lifecycle evidence leaked %q: %s", forbidden, row)
		}
	}
}

func TestPositionsReferenceViewRemainsInputFree(t *testing.T) {
	h := newDashboardHarness(t)
	seedJournal(t, h.journal)
	h.authenticate(t)
	page := strings.ToLower(h.page(t, "/positions"))
	for _, forbidden := range []string{"<form", "<input", "<select", "<textarea", "<button", "contenteditable"} {
		if strings.Contains(page, forbidden) {
			t.Errorf("positions added input surface %q", forbidden)
		}
	}
}

func TestPositionsShowRuntimeUnknownWithoutDesiredFallback(t *testing.T) {
	h := newDashboardHarness(t, func(o *Options) {
		o.PositionPolicies = &fakePositionPolicyCommander{runtimeError: errors.New("engine socket unavailable")}
		o.Settings = &fakeSettings{block: config.Adoption{DefaultStopPct: .07, IncludeSymbols: []string{"A053UNK"}}}
	})
	h.holdings.rows = append(h.holdings.rows, domain.Position{MarketType: "US", Symbol: "A053UNK",
		Quantity: 1, AveragePrice: 200, CurrentPrice: 201, MarketValue: 201})
	seedJournal(t, h.journal)
	h.authenticate(t)
	row := positionHTMLRow(t, h.page(t, "/positions"), "A053UNK")
	for _, want := range []string{"기준선·정책 폭 알 수 없음", "실행 중 엔진 설정 또는 관리 상태를 확인할 수 없음",
		"손절 <strong>—</strong>", "기준 <strong>—</strong>", "익절 <strong>—</strong>"} {
		if !strings.Contains(row, want) {
			t.Errorf("runtime-unknown row lacks %q: %s", want, row)
		}
	}
	if strings.Contains(row, "7%") {
		t.Fatalf("desired stop percentage leaked into runtime-unknown row: %s", row)
	}
}

func TestPositionsSuppressCorruptAndLifecycleUnverifiedRawEvidence(t *testing.T) {
	for _, tc := range []struct {
		name      string
		corrupt   bool
		wantLabel string
	}{
		{name: "corrupt", corrupt: true, wantLabel: "exit 정책 정보가 불완전하여 표시하지 않는다"},
		{name: "lifecycle unverified", wantLabel: "관리 세대 확인 불가"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			forbidden := []string{"SENTINEL_ENTRY", "SENTINEL_STOP", "SENTINEL_BASELINE", "SENTINEL_HIGH"}
			h := newDashboardHarness(t, func(o *Options) {
				o.PositionPolicies = &fakePositionPolicyCommander{state: positionpolicy.State{PositionID: "another-position"},
					runtime: positionpolicy.ManagementRuntime{EffectiveKnown: true}}
			})
			seedJournal(t, h.journal)
			execRaw(t, h.journal, `
				INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at)
				VALUES ('pos-a053-unsafe','123-45-678901','us','A053UNSAFE',1,'decision','OPEN','1','100','2026-08-01T00:00:00Z');
				INSERT INTO exit_states(position_id,policy_kind,entry_price,initial_stop,initial_risk,baseline_price,
				 high_water,ratchet_level,taken_ratio_total,completed,updated_at,lifecycle_generation)
				VALUES ('pos-a053-unsafe','RATCHET','SENTINEL_ENTRY','SENTINEL_STOP','1','SENTINEL_BASELINE',
				 'SENTINEL_HIGH','NONE','0',0,'2026-08-01T00:01:00Z',1);`)
			if tc.corrupt {
				execRaw(t, h.journal, `UPDATE exit_states SET snapshot_status='EVALUATED' WHERE position_id='pos-a053-unsafe';`)
				// Keep the lifecycle seam unwired so this case reaches corruption classification.
				h.opts.PositionPolicies = nil
			} else {
				line, recovery := ratchetViewSnapshot(t, "pos-a053-unsafe", 1, "1", "obs-a053-unsafe",
					"72000", "72000", "68000", "0", exitpolicy.LevelNone)
				writeViewSnapshot(t, h.journal, line, recovery, "2026-08-02T00:59:50Z")
				forbidden = append(forbidden, line.EntryPrice, line.InitialStop, line.ObservedPrice,
					line.CurrentProtection, line.NextTarget, line.SnapshotID, line.DecisionID, line.ObservationID)
			}
			h.authenticate(t)
			row := positionHTMLRow(t, h.page(t, "/positions"), "A053UNSAFE")
			if !strings.Contains(row, tc.wantLabel) {
				t.Fatalf("unsafe evidence explanation missing: %s", row)
			}
			for _, value := range forbidden {
				if value != "" && strings.Contains(row, value) {
					t.Errorf("unsafe exit evidence leaked %q: %s", value, row)
				}
			}
		})
	}
}
