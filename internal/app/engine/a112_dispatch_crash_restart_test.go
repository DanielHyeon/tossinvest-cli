//go:build tossos_testseams

package engine

// a112 태스크 6.5 (Manager 판정 2026-10-01 A65): crash/retry — dispatch 사슬의 네 지점에서 프로세스가 죽은 것처럼 멈추고, 원장을 **닫고 다시 열고**
// (재시작), 새 Guardian · 새 dispatch owner(새 epoch) · 새 파도로 같은 제안을 다시 전달한다. 단언: 브로커 요청은 재시작을 넘어 중복되지 않고,
// HELD 용량은 잘못 풀리지 않으며, owner · 캠페인 · 결속 · lease 는 둘이 되지 않는다.
//
//	① handoff 뒤 · admission 앞(진입 관문 관측 실패로 멈춤) → 재시작 뒤 정상 발급: 브로커 1 · 한 세트.
//	② admission(q_final · 캠페인 · 결속) 커밋 뒤 · lease 발급 앞(lease INSERT 에 RAISE 트리거) → 재시작 뒤 브로커 0 · HELD 유지 · lease 0.
//	③ lease 발급 · claim 뒤 · SUBMITTING 앞(송신 직전 멈춤) → 재시작 뒤 브로커 0 · lease CLAIMED · HELD 유지.
//	④ SUBMITTING 중(전송 시작 뒤 멈춤 — 브로커에 실주문이 있을 수 있다) → 재시작 뒤 추가 브로커 0 · lease SUBMITTING · HELD 유지.
//
// ② · ③ · ④ 의 사슬은 재시작 뒤 **다시 이어지지 않는다**: 전달 몸통이 claim 된 캠페인을 건너뛰고(6.4 층 ①), 생산에 lease 복구 · 결과 대사
// 호출자가 0 이다(`journal/strategy_dispatch_runtime.go:178` "No constructor for that authority exists in this build",
// `DiscoverStrategyDispatchRecovery` · `RecoverClaimedStrategyDispatchLease` 생산 호출 0). 안전 방향(해제 · 중복 없음)이지만 그 종목은 막힌 채다 —
// ④ 는 브로커 결과를 알아낼 경로가 0 인 UNKNOWN_BROKER_STATE 의 실체다. 활성화 로트의 면제 불가 선행(tasks 6.5 · ROADMAP).

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

const a112CrashSymbol = "005930"

// a112CrashLedger 는 그 종목의 첫 레그 흔적과 용량을 원장 파일에서 직접 센다(읽기 전용 연결).
type a112CrashLedger struct {
	bindings, campaigns, owners, held, released, leases int
	leaseStates                                         []string
}

func a112ReadCrashLedger(t *testing.T, path string) a112CrashLedger {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var out a112CrashLedger
	for query, target := range map[string]*int{
		`SELECT count(*) FROM strategy_first_leg_bindings WHERE symbol=?`:                                             &out.bindings,
		`SELECT count(*) FROM position_campaigns WHERE symbol=?`:                                                      &out.campaigns,
		`SELECT count(*) FROM risk_bucket_owners WHERE symbol=?`:                                                      &out.owners,
		`SELECT count(*) FROM risk_bucket_reservations WHERE symbol=? AND state='HELD' AND held_minor=reserved_minor`: &out.held,
		`SELECT count(*) FROM risk_bucket_reservations WHERE symbol=? AND state='RELEASED'`:                           &out.released,
		`SELECT count(*) FROM strategy_dispatch_leases WHERE symbol=?`:                                                &out.leases,
	} {
		if err := db.QueryRow(query, a112CrashSymbol).Scan(target); err != nil {
			t.Fatalf("%s: %v", query, err)
		}
	}
	rows, err := db.Query(`SELECT state FROM strategy_dispatch_leases WHERE symbol=? ORDER BY lease_id`, a112CrashSymbol)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		if err := rows.Scan(&state); err != nil {
			t.Fatal(err)
		}
		out.leaseStates = append(out.leaseStates, state)
	}
	return out
}

// restart 는 프로세스 재시작을 흉내 낸다: 원장 핸들을 닫고 같은 파일을 다시 열어 새 Guardian · 새 파도(새 dispatch owner 조정자)를 세운다.
// Gateway 스파이는 브로커 쪽이라 재시작을 넘어 그대로다(브로커 요청 총합을 센다).
func (fixture *a112TradingFixture) restart(t *testing.T) *clock.Fake {
	t.Helper()
	path := fixture.journal.Path()
	if err := fixture.journal.Close(); err != nil {
		t.Fatalf("closing the journal for the restart: %v", err)
	}
	journalClock := clock.NewFake(fixture.clk.Now())
	handle, err := journal.Open(context.Background(), journal.Options{Path: path, Clock: journalClock,
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt})})
	if err != nil {
		t.Fatalf("reopening the journal after the crash: %v", err)
	}
	t.Cleanup(func() { _ = handle.Close() })
	guardian, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{Journal: handle, Clock: fixture.clk, AccountRef: "acct-risk-loader",
		Policy: risk.DefaultPolicy(), Costs: costs.DefaultModel(), PolicyVersion: "engine.automation_gate/risk-policy-v1"})
	if err != nil {
		t.Fatal(err)
	}
	fixture.journal, fixture.guardian = handle, guardian
	fixture.wave(t)
	return journalClock
}

// a112CrashingGateway 는 송신 자리에서 프로세스가 죽은 모양을 만든다: beforeSubmitting 이면 SUBMITTING 앞에서, 아니면 SUBMITTING CAS 를
// 원장에 쓴 뒤(전송 시작 — 브로커가 주문을 받았을 수 있다) 멈춘다. 어느 쪽도 정상 응답을 돌려주지 않는다.
type a112CrashingGateway struct {
	*strategyDispatchGatewaySpy
	journal          *journal.Journal
	beforeSubmitting bool
	maybeSent        int
}

var errA112Crash = errors.New("a112 6.5 simulated process crash")

func (gateway *a112CrashingGateway) PlaceClaimedStrategy(ctx context.Context, request execgw.StrategyPlaceRequest) (execgw.Outcome, error) {
	if gateway.beforeSubmitting {
		return execgw.Outcome{}, errA112Crash
	}
	if _, err := gateway.journal.BeginStrategyDispatchSubmitting(ctx, request.Lease); err != nil {
		return execgw.Outcome{}, err
	}
	gateway.maybeSent++
	return execgw.Outcome{}, errA112Crash
}

func TestADispatchCrashAtEveryStepNeitherDuplicatesTheOrderNorReleasesTheCapacity(t *testing.T) {
	for _, tc := range []struct {
		name  string
		crash func(t *testing.T, fixture *a112TradingFixture) (maybeSent int, undo func())
		// 재시작 뒤 기대
		broker int
		leases []string
		// afterErr 가 있으면 재시작 뒤 전달이 그 오류로 멈추고 어떤 종목도 송신되지 않아야 한다(④: 전송 중 lease 가 새 dispatch owner 를 막는다).
		afterErr string
	}{
		{"① after handoff, before admission", func(t *testing.T, fixture *a112TradingFixture) (int, func()) {
			fixture.spy.failEntryGate = map[string]error{"kr": errA112Crash}
			return 0, func() { fixture.spy.failEntryGate = nil }
		}, 1, []string{"CLAIMED"}, ""},
		{"② after the admission commit, before lease issue", func(t *testing.T, fixture *a112TradingFixture) (int, func()) {
			db, err := sql.Open("sqlite", "file:"+fixture.journal.Path())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(`CREATE TRIGGER a112_crash_before_lease BEFORE INSERT ON strategy_dispatch_leases BEGIN SELECT RAISE(ABORT,'a112 crash'); END`); err != nil {
				t.Fatal(err)
			}
			return 0, func() {
				if _, err := db.Exec(`DROP TRIGGER a112_crash_before_lease`); err != nil {
					t.Fatal(err)
				}
				_ = db.Close()
			}
		}, 0, nil, ""},
		{"③ after lease claim, before SUBMITTING", func(t *testing.T, fixture *a112TradingFixture) (int, func()) {
			fixture.cycle.gateway = &a112CrashingGateway{strategyDispatchGatewaySpy: fixture.spy, journal: fixture.journal, beforeSubmitting: true}
			return 0, func() {}
		}, 0, []string{"CLAIMED"}, ""},
		{"④ in SUBMITTING (the broker may hold the order)", func(t *testing.T, fixture *a112TradingFixture) (int, func()) {
			crashing := &a112CrashingGateway{strategyDispatchGatewaySpy: fixture.spy, journal: fixture.journal}
			fixture.cycle.gateway = crashing
			return 0, func() {}
		}, 0, []string{"SUBMITTING"}, "strategy dispatch owner has active transport"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newA112TradingFixture(t, a112TradingOptions{})
			_, undo := tc.crash(t, &fixture)
			gateway := fixture.cycle.gateway
			crashErr := fixture.deliverKR(t)
			if !errors.Is(crashErr, errA112Crash) && (crashErr == nil || !strings.Contains(crashErr.Error(), "a112 crash")) {
				t.Fatalf("arrangement: the first wave did not stop at the crash point: %v", crashErr)
			}
			maybeSent := 0
			if crashing, ok := gateway.(*a112CrashingGateway); ok {
				maybeSent = crashing.maybeSent
			}
			before := a112ReadCrashLedger(t, fixture.journal.Path())
			undo()
			journalClock := fixture.restart(t)
			afterErr := fixture.deliverKR(t)
			switch {
			case tc.afterErr != "":
				// 전송 중(SUBMITTING) lease 가 남으면 새 프로세스는 그 lease 의 수명(+ 인수 유예) 동안 dispatch owner 를 얻지 못한다 — 그동안 **모든**
				// 전략 진입이 멈춘다(ErrStrategyDispatchOwnerBusy, journal/strategy_dispatch_runtime.go:213).
				if afterErr == nil || !strings.Contains(afterErr.Error(), tc.afterErr) || len(fixture.placedSymbols()) != 0 {
					t.Fatalf("after the restart err=%v placed=%v, want every strategy entry stopped by %q", afterErr, fixture.placedSymbols(), tc.afterErr)
				}
			case afterErr != nil && !strings.Contains(afterErr.Error(), "BUCKET_USAGE_STALE"):
				t.Fatalf("delivery after the restart: %v", afterErr)
			}
			after := a112ReadCrashLedger(t, fixture.journal.Path())
			sent := fixture.placedFor(a112CrashSymbol) + maybeSent
			t.Logf("before restart %+v · after %+v · broker requests %d (maybe-sent before the crash %d)", before, after, sent, maybeSent)
			// 브로커: ① 재시작 뒤 정상 발급 1. ②③ 0. ④ 크래시 전 전송 시작 1(브로커가 받았을 수 있음) — 재시작 뒤 추가 0.
			if fixture.placedFor(a112CrashSymbol) != tc.broker {
				t.Fatalf("broker requests after the restart=%d, want %d", fixture.placedFor(a112CrashSymbol), tc.broker)
			}
			if sent > 1 {
				t.Fatalf("broker requests across the restart=%d — a crash duplicated the order", sent)
			}
			// 한 세트: 결속 · 캠페인 · owner 하나, 용량 다섯 HELD, 풀린 것 0.
			if after.bindings != 1 || after.campaigns != 1 || after.owners != 1 || after.held != 5 || after.released != 0 {
				t.Fatalf("after the restart bindings=%d campaigns=%d owners=%d held=%d released=%d, want one first leg with its five holds intact",
					after.bindings, after.campaigns, after.owners, after.held, after.released)
			}
			if len(after.leaseStates) != len(tc.leases) {
				t.Fatalf("leases after the restart=%v, want %v", after.leaseStates, tc.leases)
			}
			for i := range tc.leases {
				if after.leaseStates[i] != tc.leases[i] {
					t.Fatalf("leases after the restart=%v, want %v", after.leaseStates, tc.leases)
				}
			}
			// 크래시가 남긴 상태를 재시작이 바꾸지 않았다(②③④): 용량 · lease 가 그대로.
			if tc.broker == 0 && (before.held != after.held || before.leases != after.leases || before.bindings != after.bindings) {
				t.Fatalf("the restart changed the stranded state: before %+v after %+v", before, after)
			}
			if tc.afterErr == "" {
				return
			}
			// ④ 의 둘째 국면: lease 수명(≤ 30s) + 인수 유예(20s)가 지나면 새 owner 는 선다 — 그러나 그 lease 는 SUBMITTING · HELD 그대로이고, 원장의
			// 복구 분류는 ATTESTED_OUTCOME_REQUIRED(「No constructor for that authority exists in this build」, journal/strategy_dispatch_runtime.go:178)다.
			ctx := context.Background()
			if _, err := fixture.journal.AcquireStrategyDispatchOwner(ctx, "a112-6.5-probe"); !errors.Is(err, journal.ErrStrategyDispatchOwnerBusy) {
				t.Fatalf("inside the lease lifetime a new owner was acquired (err=%v), want ErrStrategyDispatchOwnerBusy", err)
			}
			journalClock.Advance(time.Minute)
			owner, err := fixture.journal.AcquireStrategyDispatchOwner(ctx, "a112-6.5-probe")
			if err != nil {
				t.Fatalf("after the lease lifetime and takeover grace the owner was still refused: %v", err)
			}
			items, err := fixture.journal.DiscoverStrategyDispatchRecovery(ctx, owner)
			if err != nil {
				t.Fatal(err)
			}
			attested := 0
			for _, item := range items {
				if item.Lease.Symbol == a112CrashSymbol && item.Action == journal.StrategyDispatchRecoveryAttestedOutcome {
					attested++
				}
			}
			later := a112ReadCrashLedger(t, fixture.journal.Path())
			if attested != 1 || later.held != 5 || later.released != 0 || len(later.leaseStates) != 1 || later.leaseStates[0] != "SUBMITTING" {
				t.Fatalf("after takeover: recovery items=%+v ledger %+v, want one ATTESTED_OUTCOME_REQUIRED item and the SUBMITTING lease with its holds intact", items, later)
			}
		})
	}
}
