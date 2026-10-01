//go:build tossos_testseams

package engine

// a112 태스크 6.3 — dispatch 검증 순서 보존과 lease preimage 의 계보(Manager 판정 (A)+(C), 2026-10-01).
//
// 판정의 요지: lease 행에 family/중재 계보 열을 **더하지 않는다**(원장 스키마 v33 은 활성화 0 에서 YAGNI). 계보는 이미
//   ② family — lease 의 LaneID 가 정본 표(strategyrouter.ProductionLaneFamily)로 family 하나를 유일하게 함의하고,
//   ③ 중재 계보 — **lease 발급 앞**(1차 레그 권한의 소유자 범위 선택 + 봉인 identity 대조, admission 커밋 전)에서 조정 결과와 대조된다.
// **닫힌 지점은 발급 시점이다.** lease 발급 뒤 ~ transport 전에 조정 결과가 바뀌는 것은 lease 행에 계보가 없어 검출되지 않는다 —
// FinalAuthorityCheck 는 스케줄 재검증과 가족 만료만 본다(이 시험들은 transport 시점 계보 재대조를 주장하지 않는다; ROADMAP 「a112 이월」
// 의 (B) 스키마 열 추가가 활성화 로트의 선행).
//
// ①과 ④는 이미 서 있는 순서 · 재검사를 못 박는다. 기존 시험과 겹치는 조각은 review 「6.3」 절에 이름으로 결속한다.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

// ① dispatch 의 검증 순서(읽기 전용 경계 전부 → admission 커밋 → lease 발급 · claim → transport)와 transport 직전 최종 검사의 순서
// (스케줄 재검증이 가족 만료보다 먼저 — 8.7.2 Codex 재리뷰 P1)를 소스 순서로 얼린다. 순서를 바꾸는 편집은 이 목록을 바꿔야 한다.
func TestTheDispatchValidationOrderIsFrozen(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "strategy_dispatch_cycle.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	interest := map[string]bool{"validateStrategyFirstLegResult": true, "strategyFirstLegPlaceIntent": true, "ObserveStrategyProtection": true,
		"ProtectionReadyMinGeneration": true, "LeaseCeiling": true, "ObserveStrategyEntryGate": true, "dispatchOwner": true, "admit": true,
		"LookupDecision": true, "forScope": true, "IssueVerifiedFirstLegStrategyDispatchLease": true, "ClaimStrategyDispatchLease": true,
		"PlaceClaimedStrategy": true, "revalidateSchedule": true}
	name := func(call *ast.CallExpr) string {
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			return fun.Name
		case *ast.SelectorExpr:
			return fun.Sel.Name
		}
		return ""
	}
	var outer, final []string
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || function.Name.Name != "dispatch" || function.Recv == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if literal, ok := node.(*ast.FuncLit); ok {
				ast.Inspect(literal.Body, func(inner ast.Node) bool {
					if call, ok := inner.(*ast.CallExpr); ok && interest[name(call)] {
						final = append(final, name(call))
					}
					return true
				})
				return false
			}
			if call, ok := node.(*ast.CallExpr); ok && interest[name(call)] {
				outer = append(outer, name(call))
			}
			return true
		})
	}
	// ProtectionReadyMinGeneration 이 둘인 것은 하한 비교와 그 거절 문구가 각자 부르기 때문이다(측정값 — 첫 판의 기대가 하나로 틀렸다).
	wantOuter := []string{"validateStrategyFirstLegResult", "strategyFirstLegPlaceIntent", "ObserveStrategyProtection",
		"ProtectionReadyMinGeneration", "ProtectionReadyMinGeneration", "LeaseCeiling", "ObserveStrategyEntryGate", "dispatchOwner", "admit", "LookupDecision", "forScope",
		"IssueVerifiedFirstLegStrategyDispatchLease", "ClaimStrategyDispatchLease", "strategyFirstLegPlaceIntent", "PlaceClaimedStrategy"}
	wantFinal := []string{"revalidateSchedule", "LeaseCeiling"}
	if strings.Join(outer, ",") != strings.Join(wantOuter, ",") || strings.Join(final, ",") != strings.Join(wantFinal, ",") {
		t.Fatalf("dispatch order changed:\n outer=%v\n  want=%v\n final=%v\n  want=%v", outer, wantOuter, final, wantFinal)
	}
}

// ② lease 행은 계보의 레인을 그대로 싣고, 그 레인은 정본 표에서 family 하나를 유일하게 함의한다 — family 열이 따로 없어도 lease 가 어느
// 가족의 것인지 모호하지 않다. 여덟 생산 레인 전부에서 레인 → 가족이 단사(두 가족이 한 레인을 나누지 않음)인지도 잰다.
func TestALeaseNamesItsLaneAndTheLaneNamesExactlyOneFamily(t *testing.T) {
	seen := map[string]strategyrouter.Family{}
	for _, descriptor := range strategyflow.Descriptors() {
		family, ok := strategyrouter.ProductionLaneFamily(descriptor.Market, descriptor.LaneID)
		if !ok || !family.Known() {
			t.Fatalf("production lane %s/%s has no canonical family", descriptor.Market, descriptor.LaneID)
		}
		key := string(descriptor.Market) + "/" + descriptor.LaneID
		if previous, dup := seen[key]; dup && previous != family {
			t.Fatalf("lane %s names two families %s and %s", key, previous, family)
		}
		seen[key] = family
	}
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
		result := proposals.forMarket(market).entries[0].authority.Proposal()
		if _, err := cycle.dispatch(context.Background(), deliverForTest(t, result)); err != nil {
			t.Fatalf("%s arrangement: the genuine proposal was not dispatched: %v", market, err)
		}
		if len(spy.calls) != 1 {
			t.Fatalf("%s gateway calls=%d, want 1", market, len(spy.calls))
		}
		lease, err := j.LookupStrategyDispatchLease(context.Background(), spy.calls[0].Lease.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		if lease.LaneID != result.Lineage.LaneID || lease.LaneVersion != result.Lineage.LaneVersion || lease.CampaignID != result.Lineage.CampaignID {
			t.Fatalf("%s lease lane=%s/%s campaign=%s, lineage %s/%s/%s", market, lease.LaneID, lease.LaneVersion, lease.CampaignID,
				result.Lineage.LaneID, result.Lineage.LaneVersion, result.Lineage.CampaignID)
		}
		if family, ok := strategyrouter.ProductionLaneFamily(result.Lineage.Market, lease.LaneID); !ok || !family.Known() {
			t.Fatalf("%s lease lane %s does not name a family", market, lease.LaneID)
		}
	}
}

// ③ 같은 소유자 범위의 다른 계보(다른 캠페인으로 봉인된 제안)가 dispatch 쪽에서 건너와도 **발급 시점**에 거절된다: 1차 레그 권한이 조립의
// 중재 결과와 봉인 identity 를 대조해 admission 커밋 전에 멈추므로 원장의 예약 · lease 행 0, 게이트웨이(브로커 경로) 호출 0.
// 이 시험이 단언하는 닫힌 지점은 이것 하나다 — lease 발급 뒤의 계보 변경은 다루지 않는다(머리말).
func TestALineageDriftIsRefusedAtIssuanceBeforeAnyReservationLeaseOrBrokerCall(t *testing.T) {
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			cycle, proposals, j, spy := pairedStrategyDispatchCycleFixture(t)
			winner := proposals.forMarket(market).entries[0].authority.Proposal()
			entry, stop, target := "100", "95", "120"
			if market == StrategyMarketUS {
				entry, stop, target = "10000", "9500", "12000"
			}
			now := cycle.schedule.observedAt
			drifted, err := strategyflow.AcceptedResultForAuthorityTest(riskLoaderDescriptor(t, market), winner.Lineage.AccountRef,
				winner.Lineage.Symbol, "campaign-a112-6-3-drift", winner.Quantity, entry, stop, target, now.Add(-time.Second), now.Add(time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			batch := strategyproposal.ProductionBatchAuthorityForTest("sha256:drift-"+string(market), map[string]strategyflow.Result{drifted.Lineage.Symbol: drifted})
			sealed, ok := batch.For(drifted.Lineage.Symbol)
			if !ok || sealed.Proposal().Lineage.Identity == winner.Lineage.Identity {
				t.Fatal("arrangement: the drifted proposal must be a different sealed lineage in the same owner scope")
			}
			before := a112DispatchRowCounts(t, j.Path())
			_, err = cycle.dispatch(context.Background(), deliverForTest(t, sealed.Proposal()))
			if err == nil || !strings.Contains(err.Error(), "production proposal identity changed") {
				t.Fatalf("drifted lineage dispatch err=%v, want the issuance-time identity refusal", err)
			}
			if after := a112DispatchRowCounts(t, j.Path()); after != before {
				t.Fatalf("a refused lineage left rows: before=%v after=%v (reservations, leases)", before, after)
			}
			if len(spy.calls) != 0 {
				t.Fatalf("a refused lineage reached the gateway %d time(s)", len(spy.calls))
			}
		})
	}
}

func a112DispatchRowCounts(t *testing.T, path string) [2]int {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var counts [2]int
	for index, table := range []string{"risk_bucket_reservations", "strategy_dispatch_leases"} {
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&counts[index]); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
	}
	return counts
}

// ④ transport 직전 최종 검사는 스케줄 재검증의 거절(활성화 세대 · 만료 · 매니페스트 digest · 달력 · desired revision 이 admission 때와
// 다름)을 그대로 돌려준다 — 게이트웨이는 그 오류에서 브로커 바이트를 보내지 않는다(execgw
// TestStrategyGatewayRechecksSourceBackedActivationAfterSubmittingFencePairedKRUS 가 places=0 을 잰다).
func TestTheFinalCheckRefusesWhenTheScheduleRevalidatorSeesDrift(t *testing.T) {
	for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {
		cycle, proposals, _, spy := pairedStrategyDispatchCycleFixture(t)
		drift := errors.New("a112 6.3 signed scheduler activation no longer matches dispatch admission")
		cycle.revalidateSchedule = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority) error { return drift }
		result := proposals.forMarket(market).entries[0].authority.Proposal()
		if _, err := cycle.dispatch(context.Background(), deliverForTest(t, result)); err != nil {
			t.Fatalf("%s arrangement: dispatch err=%v", market, err)
		}
		if len(spy.calls) != 1 || spy.calls[0].FinalAuthorityCheck == nil {
			t.Fatalf("%s gateway request without a final check", market)
		}
		if err := spy.calls[0].FinalAuthorityCheck(context.Background()); !errors.Is(err, drift) {
			t.Fatalf("%s final check err=%v, want the revalidator's drift refusal", market, err)
		}
	}
}

// 6.3 잔여 (c) — 생산 재검증기의 drift 판정을 축마다 잰다(Manager 판정 2026-10-01). 판정은 strategyScheduleStillMatchesAdmission 으로 의미 무변경
// 이동됐다(영수증 lot-6.3/move-receipt.txt).

// a112ScheduleDriftCondition 은 이동 전(HEAD 76dc35f7 의 재검증 클로저) 조건식의 go/types 철자다 — 영수증에서 옮겨 얼렸다.
const a112ScheduleDriftCondition = "!fresh.snapshot.Ready || fresh.restore.Activation == nil || expected.restore.Activation == nil || " +
	"fresh.desired.Revision != expected.desired.Revision || fresh.calendar.Version != expected.calendar.Version || " +
	"fresh.snapshot.ActivationManifestDigest != expected.snapshot.ActivationManifestDigest || " +
	"fresh.restore.Activation.Generation() != expected.restore.Activation.Generation() || " +
	"!fresh.restore.Activation.ExpiresAt().Equal(expected.restore.Activation.ExpiresAt())"

// 옮긴 코드의 양쪽을 못 박는다: 판정 함수의 조건식은 이동 전 철자 그대로이고, 생산 클로저는 수집 한 문장 + 그 함수 호출 한 문장뿐이다(조건을
// 클로저에 다시 두거나 함수를 우회하면 실패).
func TestTheScheduleDriftJudgementMovedVerbatim(t *testing.T) {
	judgement, err := parser.ParseFile(token.NewFileSet(), "strategy_schedule_revalidation.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var conditions []string
	for _, decl := range judgement.Decls {
		if function, ok := decl.(*ast.FuncDecl); ok && function.Name.Name == "strategyScheduleStillMatchesAdmission" {
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if stmt, ok := node.(*ast.IfStmt); ok {
					conditions = append(conditions, types.ExprString(stmt.Cond))
				}
				return true
			})
		}
	}
	if len(conditions) != 1 || conditions[0] != a112ScheduleDriftCondition {
		t.Fatalf("drift judgement conditions=%q, want exactly the pre-move spelling %q", conditions, a112ScheduleDriftCondition)
	}
	supervisor, err := parser.ParseFile(token.NewFileSet(), "strategy_entry_supervisor.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var bodies [][]string
	ast.Inspect(supervisor, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		selector, ok := assign.Lhs[0].(*ast.SelectorExpr)
		literal, isLiteral := assign.Rhs[0].(*ast.FuncLit)
		if !ok || !isLiteral || selector.Sel.Name != "revalidateSchedule" {
			return true
		}
		statements := make([]string, 0, len(literal.Body.List))
		for _, stmt := range literal.Body.List {
			switch value := stmt.(type) {
			case *ast.AssignStmt:
				statements = append(statements, "assign "+types.ExprString(value.Lhs[0])+" := "+exprSpelling(value.Rhs[0]))
			case *ast.ReturnStmt:
				if len(value.Results) == 1 {
					statements = append(statements, "return "+exprSpelling(value.Results[0]))
				}
			default:
				statements = append(statements, fmt.Sprintf("%T", stmt))
			}
		}
		bodies = append(bodies, statements)
		return true
	})
	if len(bodies) != 1 || len(bodies[0]) != 2 || !strings.HasPrefix(bodies[0][0], "assign fresh := ") ||
		!strings.HasSuffix(bodies[0][0], ".collectMarket(checkCtx, market)") ||
		bodies[0][1] != "return strategyScheduleStillMatchesAdmission(fresh, expected)" {
		t.Fatalf("production revalidator closures=%q, want one closure: collect fresh, then return the drift judgement", bodies)
	}
}

// 축마다: admission 때와 같으면 통과, 한 축이라도 다르면 거절. 축 하나를 빼는 변이가 각각 잡힌다(lot-6.3/mutation-6.3.tsv A1~A8).
func TestTheScheduleDriftJudgementRefusesEveryAxis(t *testing.T) {
	approved := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	activation := func(revision uint64, approvedAt time.Time) *scheduler.Activation {
		return scheduler.ActivationForTest(scheduler.ActivationBinding{DesiredRevision: revision, ApprovedAt: approvedAt})
	}
	base := func() strategyScheduleMarketAuthority {
		return strategyScheduleMarketAuthority{market: StrategyMarketKR,
			desired:  scheduler.DesiredState{Revision: 3},
			calendar: scheduler.CalendarSnapshot{Version: "calendar-v1"},
			restore:  scheduler.RestoreResult{Activation: activation(3, approved)},
			snapshot: StrategyScheduleMarketSnapshot{Ready: true, ActivationManifestDigest: "sha256:" + strings.Repeat("a", 64)}}
	}
	if err := strategyScheduleStillMatchesAdmission(base(), base()); err != nil {
		t.Fatalf("control: an unchanged schedule was refused: %v", err)
	}
	for name, drift := range map[string]func(fresh, expected *strategyScheduleMarketAuthority){
		"A1 fresh not ready":             func(fresh, _ *strategyScheduleMarketAuthority) { fresh.snapshot.Ready = false },
		"A2 fresh activation missing":    func(fresh, _ *strategyScheduleMarketAuthority) { fresh.restore.Activation = nil },
		"A3 expected activation missing": func(_, expected *strategyScheduleMarketAuthority) { expected.restore.Activation = nil },
		"A4 desired revision":            func(fresh, _ *strategyScheduleMarketAuthority) { fresh.desired.Revision = 4 },
		"A5 calendar version":            func(fresh, _ *strategyScheduleMarketAuthority) { fresh.calendar.Version = "calendar-v2" },
		"A6 manifest digest": func(fresh, _ *strategyScheduleMarketAuthority) {
			fresh.snapshot.ActivationManifestDigest = "sha256:" + strings.Repeat("b", 64)
		},
		"A7 activation generation": func(fresh, _ *strategyScheduleMarketAuthority) { fresh.restore.Activation = activation(4, approved) },
		"A8 activation expiry": func(fresh, _ *strategyScheduleMarketAuthority) {
			fresh.restore.Activation = activation(3, approved.Add(time.Minute))
		},
	} {
		t.Run(name, func(t *testing.T) {
			fresh, expected := base(), base()
			drift(&fresh, &expected)
			if err := strategyScheduleStillMatchesAdmission(fresh, expected); err == nil ||
				!strings.Contains(err.Error(), "no longer matches dispatch admission") {
				t.Fatalf("drift on %s was not refused: %v", name, err)
			}
		})
	}
}
