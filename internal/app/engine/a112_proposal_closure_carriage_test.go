//go:build tossos_testseams

package engine

// a112 태스크 8.8.4 항목 2(Manager 판정 2026-10-04, Q-B2 = (ii) · 계산/판정 분리): `collectMarket` 의 닫힘 갈래 13 이 싣는 활성화.
// 앞 판은 관문(`familyGateFor`)을 제안 적재 **뒤**에 계산해, 그 앞 일곱 닫힘(RouteNotReady ~ ProductionFault)이 영값 활성화를 실었다 —
// 그 주기의 레인 관측이 관문과 다른 승격(없음)을 본다. 이제 관문 **계산만** RouteNotReady 가드 직후로 옮긴다:
//   - 판정 자리(실패 kind · 우선순위 · FAMILY_GATE_CLOSED 의 자리)는 그대로 — 아래 census 가 13 갈래의 kind 순서를 못 박는다.
//   - 옮긴 값은 닫힘 갈래가 싣는 활성화로만 쓰인다.
//   - RouteNotReady 는 영값을 싣는다(사유: 경로 · 스케줄 권한이 준비되지 않은 주기에는 관문을 계산할 결속 값 — 경로 매니페스트 digest ·
//     보정 · 달력 — 이 없거나 믿을 수 없다; 그 갈래는 관문 계산 전이며 활성화 적재 호출 0 이 그 증거).
// 두 층: (1) 구조 census(AST) — 13 닫힘의 kind 순서와 관문 대입 자리가 정확히 표와 같다(새 갈래는 표 편집을 강제). 충돌 · 미해결 선택
// 두 갈래는 입력으로 닿기 어렵다(조정자는 색인한 신원만 고른다) — census 만 잰다. (2) 행동 — 닿을 수 있는 닫힘 여덟을 관문 세 모양(검증 ·
// 되돌림 · 미선언)으로 몰아 싣는 활성화 · 사유 · 활성화 적재 호출 수를 잰다.

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyproposal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

const a112GateMarker = "GATE"

// 13 닫힘의 실패 kind 를 소스 순서대로, 관문 계산 자리와 함께(2026-10-04 판정 뒤의 정본 순서).
var a112CollectMarketClosureOrder = []string{
	"StrategyProposalRouteNotReady",
	a112GateMarker,
	"StrategyProposalFXNotReady",
	"StrategyProposalInternalFailure",  // 적재기 설정
	"StrategyProposalAuthorityInvalid", // 제안 공개 열쇠
	"StrategyProposalInternalFailure",  // 경로 종목 중복
	"StrategyProposalAuthorityInvalid", // 제안 적재 실패 · digest 불일치
	"StrategyProposalProductionFault",
	"StrategyProposalFamilyGateClosed",
	"StrategyProposalInternalFailure", // 계보 신원 충돌
	"StrategyProposalQueueOverflow",
	"StrategyProposalArbitrationRefused",
	"StrategyProposalInternalFailure", // 미해결 선택
	"StrategyProposalNoAcceptedScope",
}

func TestTheThirteenProposalClosuresKeepTheirOrderAndTheGateIsComputedRightAfterRouteReadiness(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "strategy_proposal_authority.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "collectMarket" && fn.Recv != nil {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("collectMarket not found")
	}
	var order []string
	gateAssignments, failLiteralsCarryGate, successCarriesGate := 0, 0, 0
	ast.Inspect(body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.AssignStmt:
			if len(x.Lhs) == 1 {
				if id, ok := x.Lhs[0].(*ast.Ident); ok && id.Name == "gate" && x.Tok == token.ASSIGN {
					call, ok := x.Rhs[0].(*ast.CallExpr)
					sel, ok2 := call.Fun.(*ast.SelectorExpr)
					if !ok || !ok2 || sel.Sel.Name != "familyGateFor" {
						t.Errorf("gate is assigned from something other than familyGateFor")
					}
					gateAssignments++
					order = append(order, a112GateMarker)
				}
			}
		case *ast.CallExpr:
			if id, ok := x.Fun.(*ast.Ident); ok && id.Name == "fail" && len(x.Args) == 1 {
				if reason, ok := x.Args[0].(*ast.Ident); ok {
					order = append(order, reason.Name)
				} else {
					order = append(order, "<non-constant reason>")
				}
			}
		case *ast.CompositeLit:
			if id, ok := x.Type.(*ast.Ident); ok && id.Name == "strategyProposalMarketAuthority" {
				for _, element := range x.Elts {
					kv, ok := element.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "activation" {
						if sel, ok := kv.Value.(*ast.SelectorExpr); ok && sel.Sel.Name == "activation" {
							if base, ok := sel.X.(*ast.Ident); ok && base.Name == "gate" {
								if a112HasKey(x, "entries") {
									successCarriesGate++
								} else {
									failLiteralsCarryGate++
								}
							}
						}
					}
				}
			}
		}
		return true
	})
	if strings.Join(order, ",") != strings.Join(a112CollectMarketClosureOrder, ",") {
		t.Fatalf("collectMarket closure order:\n got  %v\n want %v\n(a new or moved closing branch, or a moved gate computation, must update this table and review 8.8.4)", order, a112CollectMarketClosureOrder)
	}
	if gateAssignments != 1 || failLiteralsCarryGate != 1 || successCarriesGate != 1 {
		t.Fatalf("gate assignments=%d, fail-closure literals carrying gate.activation=%d, success literals carrying it=%d — want 1, 1, 1",
			gateAssignments, failLiteralsCarryGate, successCarriesGate)
	}
}

func a112HasKey(literal *ast.CompositeLit, name string) bool {
	for _, element := range literal.Elts {
		if kv, ok := element.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok && key.Name == name {
				return true
			}
		}
	}
	return false
}

func TestEveryReachableProposalClosureCarriesTheGatesActivationExceptRouteNotReady(t *testing.T) {
	now := time.Date(2026, 9, 3, 1, 2, 3, 0, time.UTC)
	verified := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 7, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	modes := []struct {
		name       string
		activation strategyrouter.FamilyActivation
		err        error
	}{
		{"verified", verified, nil},
		{"rolled back (declared but unusable)", strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationUnavailable},
		{"undeclared", strategyrouter.FamilyActivation{}, strategyrouter.ErrProductionFamilyActivationUndeclared},
	}
	type knobs struct {
		loader *strategyProposalAuthorityLoader
		routes strategyRouteMarketAuthority
		fx     strategyFXMarketAuthority
	}
	resultFor := func(config strategyproposal.ProductionConfig, symbol string) strategyflow.Result {
		var descriptor strategyflow.Descriptor
		for _, d := range strategyflow.Descriptors() {
			if d.LaneID == continuationlane.KRContinuationLaneID {
				descriptor = d
			}
		}
		result, err := strategyflow.AcceptedResultForAuthorityTest(descriptor, config.AccountRef, symbol, "campaign-KR", 8, "100", "90", "120",
			now.Add(-time.Second), now.Add(time.Minute))
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	cases := []struct {
		name   string
		reason StrategyProposalReason
		mutate func(*knobs)
		loads  int // 활성화 적재 호출 수
	}{
		{"route not ready", StrategyProposalRouteNotReady, func(k *knobs) { k.routes.snapshot.Ready = false }, 0},
		{"fx not ready", StrategyProposalFXNotReady, func(k *knobs) { k.fx.snapshot.Ready = false }, 1},
		{"loader misconfigured", StrategyProposalInternalFailure, func(k *knobs) { k.loader.accountRef = "" }, 1},
		{"proposal public key invalid", StrategyProposalAuthorityInvalid, func(k *knobs) {
			getenv := k.loader.getenv
			k.loader.getenv = func(name string) string {
				if name == strategyProposalPublicKeyEnv {
					return "not-a-key"
				}
				return getenv(name)
			}
		}, 1},
		{"duplicate routed symbol", StrategyProposalInternalFailure, func(k *knobs) {
			k.routes.entries = append(append([]strategyRouteEntryAuthority(nil), k.routes.entries...), k.routes.entries[0])
		}, 1},
		{"proposal load failed", StrategyProposalAuthorityInvalid, func(k *knobs) {
			k.loader.load = func(context.Context, strategyproposal.ProductionConfig, []strategyproposal.ProductionTarget, interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
				return strategyproposal.ProductionBatchAuthority{}, errors.New("a112 8.8.4 load failure")
			}
		}, 1},
		{"a scope lost its proposal", StrategyProposalProductionFault, func(k *knobs) {
			k.loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
				values := map[string][]strategyflow.Result{}
				for _, target := range targets {
					values[target.Approved.Symbol()] = []strategyflow.Result{resultFor(config, target.Approved.Symbol())}
				}
				return strategyproposal.ProductionBatchAuthorityWithFaultForTest(config.ManifestDigest, values, strategyproposal.ProductionAbsence{
					Symbol: "005930", LaneID: continuationlane.KRContinuationLaneID, Reason: strategyproposal.ProductionAbsenceEvidenceReplay}), nil
			}
		}, 1},
		{"no scope accepted (post-gate control)", StrategyProposalNoAcceptedScope, func(k *knobs) {
			k.loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, _ []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
				return strategyproposal.ProductionBatchAuthorityMultiLaneForTest(config.ManifestDigest, map[string][]strategyflow.Result{}), nil
			}
		}, 1},
	}
	for _, mode := range modes {
		for _, tc := range cases {
			t.Run(mode.name+"/"+tc.name, func(t *testing.T) {
				runtime, _ := familyGateFixture(t)
				loader := testStrategyProposalLoader(t).withStrategyLanes(runtime)
				loader.load = func(_ context.Context, config strategyproposal.ProductionConfig, targets []strategyproposal.ProductionTarget, _ interfaceOfficialFX) (strategyproposal.ProductionBatchAuthority, error) {
					return arbitrationBatch(t, config, targets, now, "005930", []string{continuationlane.KRContinuationLaneID}), nil
				}
				loads := 0
				loader.loadActivation = func(context.Context, StrategyMarket, strategyScheduleMarketAuthority, strategyRouteMarketAuthority, time.Time) (strategyrouter.FamilyActivation, error) {
					loads++
					return mode.activation, mode.err
				}
				k := knobs{loader: loader,
					routes: arbitrationRoutePair(t, now, familyScoresForTest(strategyrouter.MarketKR), "005930", continuationlane.KRContinuationLaneID).forMarket(StrategyMarketKR),
					fx:     proposalFXPair(now).forMarket(StrategyMarketKR)}
				tc.mutate(&k)
				got := k.loader.collectMarket(context.Background(), routeReadySchedulePair(now).forMarket(StrategyMarketKR), k.routes, k.fx, now)
				if got.snapshot.Ready || got.snapshot.Reason != tc.reason {
					t.Fatalf("reason=%s ready=%v, want closed with %s", got.snapshot.Reason, got.snapshot.Ready, tc.reason)
				}
				if loads != tc.loads {
					t.Fatalf("activation loads=%d, want %d", loads, tc.loads)
				}
				wantVerified := mode.err == nil && tc.reason != StrategyProposalRouteNotReady
				if got.familyActivation().Verified() != wantVerified {
					t.Fatalf("carried activation verified=%v, want %v", got.familyActivation().Verified(), wantVerified)
				}
				if wantVerified && got.familyActivation().Generation() != verified.Generation() {
					t.Fatalf("carried generation=%d, want %d", got.familyActivation().Generation(), verified.Generation())
				}
			})
		}
	}
}
