//go:build tossos_testseams

package engine

// a112 태스크 7.3.1 R2(Manager 판정 2026-10-04): spec four-family-strategy-runtime 「SHADOW는 OFF state를 승격하지 않는 read-only runtime이다」
// (spec.md:88-93)는 오늘 **SHADOW 상태가 없어서** 성립한다 — 빈 표본 통과다. 그것을 핀 통과로 바꾼다:
//   (1) 재시작 시나리오(:91-93): 앞 프로세스가 검증된 활성화 아래 레인 ON 을 관측한 뒤 같은 원장으로 재시작하면 여덟 레인 전부
//       OFF/OFF/UNOBSERVED 이고, 원장에 dispatch lease 0 · 레인 잠금 기록 0 — 프로세스 안 상태가 되살아나지 않는다. 활성화는 원장이 아니라
//       배포가 핀한 파일에서 매 파도 다시 읽고(8.7.2) 그 파일을 쓰는 생산 코드는 0 이다(`TestOnlyTheAuthoringToolCanBuildActivationBytes`).
//   (2) 어휘 census: runtime 어휘는 정확히 {UNOBSERVED} — strategyrouter 의 `RuntimeState` 상수와 projection 의 `LaneRuntime` 상수. SHADOW 를
//       더하는 편집은 이 census 를 뒤집어야 한다(그 로트가 spec 의 허용 절 · 서명 shadow 매니페스트 · 골든 개정을 함께 세운다 — ROADMAP).
// 기존 결속: `TestEveryProductionWorkerIsBornDormantAndEmitsNothing`(strategyworker) · `TestDescriptorsShipKRAndUSTogetherDefaultOFF`(strategyrouter) ·
// `TestDormantAndUnavailableSnapshotsCarryEightUnobservedLanesAndTwoCoordinators`(strategyprojection — 미관측 레인 = OFF/OFF/UNOBSERVED).

import (
	"context"
	"database/sql"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestARestartAfterAnObservedPromotionComesBackOffOffUnobservedWithNothingWritten(t *testing.T) {
	c, lanes, fake := a112LaneProjectionContext(t)
	activation := strategyrouter.FamilyActivationForTest(strategyrouter.MarketKR, 1, strategyrouter.AllFourFamiliesForTest(strategyrouter.MarketKR))
	if err := lanes.evaluate(context.Background(), StrategyMarketKR, 1, activation, nil, strategyShadowBatch{}); err != nil {
		t.Fatal(err)
	}
	on := 0
	for _, lane := range a112Read(t, c).Lanes {
		if lane.Desired == strategyprojection.StateOn {
			on++
		}
	}
	if on != 4 {
		t.Fatalf("arrangement: the first process observed %d lanes ON, want the four KR lanes", on)
	}
	// 재시작: 같은 원장 · 새 Context(새 투영 저장소 · 새 레인 런타임) — 프로세스 안 상태는 없다.
	store, err := strategyprojection.NewStore(strategyprojection.DormantSnapshot(fake.Now()))
	if err != nil {
		t.Fatal(err)
	}
	restarted := &Context{Journal: c.Journal, AccountRef: c.AccountRef, strategyProjection: store}
	if _, err := restarted.productionStrategyLanes(context.Background(), fake); err != nil {
		t.Fatal(err)
	}
	snapshot := a112Read(t, restarted)
	if len(snapshot.Lanes) != 8 {
		t.Fatalf("lanes=%d after the restart, want 8", len(snapshot.Lanes))
	}
	for _, lane := range snapshot.Lanes {
		if lane.Desired != strategyprojection.StateOff || lane.Effective != strategyprojection.StateOff || lane.Runtime != strategyprojection.LaneRuntimeUnobserved {
			t.Fatalf("lane %s after the restart is %s/%s/%s, want OFF/OFF/UNOBSERVED", lane.LaneID, lane.Desired, lane.Effective, lane.Runtime)
		}
	}
	db, err := sql.Open("sqlite", "file:"+c.Journal.Path()+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, table := range []string{"strategy_dispatch_leases", "strategy_lane_latches", "strategy_lane_latch_recoveries"} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM ` + table).Scan(&n); err != nil || n != 0 {
			t.Fatalf("%s rows=%d err=%v, want 0 — the observed promotion must leave nothing for a restart to restore", table, n, err)
		}
	}
}

func TestTheRouterRuntimeStaysUnobservedAndOnlyTheProjectionAddsShadow(t *testing.T) {
	census := func(dir, typeName string) []string {
		t.Helper()
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil || len(files) == 0 {
			t.Fatalf("no Go files under %s (err=%v)", dir, err)
		}
		var values []string
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.CONST {
					continue
				}
				for _, spec := range gen.Specs {
					value := spec.(*ast.ValueSpec)
					if id, ok := value.Type.(*ast.Ident); ok && id.Name == typeName {
						for index := range value.Names {
							if lit, ok := value.Values[index].(*ast.BasicLit); ok {
								values = append(values, strings.Trim(lit.Value, `"`))
							} else {
								values = append(values, "<non-literal>")
							}
						}
					}
				}
			}
		}
		sort.Strings(values)
		return values
	}
	// a112 7.3.1 SHADOW 로트(브리프 v3.3 §7)가 이 census 의 주인이다: SHADOW 는 projection 어휘에만 열리고, router RuntimeState 는
	// MarketRecord 와 공유하므로 {UNOBSERVED} 그대로다. 어휘는 타입을 가진 상수로만 세운다(변환식 `LaneRuntime("…")` 은 아래에서 0 으로 잰다).
	for _, tc := range []struct{ dir, typeName, want string }{
		{"../../strategyrouter", "RuntimeState", "UNOBSERVED"},
		{"../../strategyprojection", "LaneRuntime", "SHADOW,UNOBSERVED"},
	} {
		if got := census(tc.dir, tc.typeName); strings.Join(got, ",") != tc.want {
			t.Errorf("%s %s constants=%v, want exactly [%s]", tc.dir, tc.typeName, got, tc.want)
		}
	}
	// 변환식 census: 생산 코드가 문자열을 LaneRuntime 으로 바꿔 새 값을 지어내지 않는다 — 엔진 투영의 worker runtime 변환 한 자리만.
	conversions := 0
	for _, dir := range []string{".", "../../strategyprojection"} {
		files, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					return true
				}
				name := ""
				switch fun := call.Fun.(type) {
				case *ast.Ident:
					name = fun.Name
				case *ast.SelectorExpr:
					name = fun.Sel.Name
				}
				if name == "LaneRuntime" {
					conversions++
					if _, literal := call.Args[0].(*ast.BasicLit); literal {
						t.Errorf("%s converts a string literal into LaneRuntime — the vocabulary is the typed constants", path)
					}
				}
				return true
			})
		}
	}
	if conversions != 1 {
		t.Errorf("LaneRuntime conversions=%d, want exactly the one worker-runtime conversion in strategyLaneProjection", conversions)
	}
}
