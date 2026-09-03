package strategyworker

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func restoreFixture(t *testing.T) (clock.Clock, LatchRestore) {
	t.Helper()
	return clock.NewFake(time.Date(2026, 9, 3, 6, 0, 0, 0, time.UTC)), LatchRestore{
		Key: Key{Market: strategyrouter.MarketKR, Family: strategyrouter.FamilyBreakoutRetest,
			LaneID: "kr_short_breakout_retest_v1", LaneVersion: "v1"},
		LatchID:       "lane-latch:KR:BREAKOUT_RETEST:kr_short_breakout_retest_v1:v1:1",
		LatchRevision: 1, Reason: "breakout evidence store unavailable", Abnormal: true,
		ObservedAt: time.Date(2026, 9, 2, 6, 0, 0, 0, time.UTC),
	}
}

// TestAProductionLaneBornFromADurableRecordIsLatched 는 5.3.3 의 첫 문장이다:
// 레인은 열린 채로 태어나는 것이 아니라 **기록에서** 태어난다.
func TestAProductionLaneBornFromADurableRecordIsLatched(t *testing.T) {
	fake, restore := restoreFixture(t)
	lanes, unmatched := ProductionLanesFrom(fake, []LatchRestore{restore})
	if len(unmatched) != 0 {
		t.Fatalf("생산 레인 열쇠인데 붙지 않았다: %+v", unmatched)
	}
	if len(lanes) != len(ProductionWorkers()) {
		t.Fatalf("레인 %d 개 — 여덟이어야 한다", len(lanes))
	}

	latchedSeen := 0
	for _, lane := range lanes {
		if lane.Key() != restore.Key {
			if lane.Latched() || lane.Health() != LaneHealthy {
				t.Fatalf("이웃 레인 %v 가 함께 잠겼다", lane.Key().Parts())
			}
			continue
		}
		latchedSeen++
		if !lane.Latched() || lane.Health() != LaneLatched {
			t.Fatalf("기록이 있는 레인이 열린 채로 태어났다: latched=%v health=%s", lane.Latched(), lane.Health())
		}
		if lane.FirstFailure() != restore.Reason {
			t.Fatalf("첫 원인이 %q — 기록이 말한 것이 아니다", lane.FirstFailure())
		}
		if lane.LatchRevision() != restore.LatchRevision {
			t.Fatalf("latch revision %d — 기록이 말한 것이 아니다", lane.LatchRevision())
		}
		// 잠긴 레인은 사이클을 받지도, 돌지도 않는다.
		if trigger := lane.Offer(); trigger != TriggerDisabled {
			t.Fatalf("잠긴 레인이 투입을 받았다: %s", trigger)
		}
		if cycle := lane.Run(strategyrouter.FamilyActivation{}, Input{}); cycle.Outcome != OutcomeLatched {
			t.Fatalf("잠긴 레인의 사이클 결과가 %s 다", cycle.Outcome)
		}
	}
	if latchedSeen != 1 {
		t.Fatalf("기록 하나가 레인 %d 개에 붙었다", latchedSeen)
	}
}

// TestARestoredLatchDoesNotRestoreTheFailureCounter 는 되살리지 **않는** 것을
// 못 박는다. 계수기는 "지금 몇 번 연달아 실패하는 중인가"이고, 다시 선 프로세스는
// 아직 한 번도 안 돌았다.
func TestARestoredLatchDoesNotRestoreTheFailureCounter(t *testing.T) {
	fake, restore := restoreFixture(t)
	lanes, _ := ProductionLanesFrom(fake, []LatchRestore{restore})
	for _, lane := range lanes {
		if lane.ConsecutiveFailures() != 0 {
			t.Fatalf("레인 %v 가 계수기 %d 로 태어났다", lane.Key().Parts(), lane.ConsecutiveFailures())
		}
	}
}

// TestARecordThatNamesNoLaneInThisBuildIsHandedBackNotSwallowed 는 조용히 버리지
// 않는다는 것이다. 붙지 않은 기록은 사람이 잠가 둔 것이 아무 일도 하지 않으면서
// 기록으로만 남는 상태이고, 그것을 보는 것은 부르는 쪽의 몫이다.
func TestARecordThatNamesNoLaneInThisBuildIsHandedBackNotSwallowed(t *testing.T) {
	fake, restore := restoreFixture(t)
	for name, mutate := range map[string]func(*LatchRestore){
		"이 빌드에 없는 lane_id": func(r *LatchRestore) { r.Key.LaneID = "kr_short_breakout_retest_v2" },
		"이 빌드에 없는 버전":      func(r *LatchRestore) { r.Key.LaneVersion = "v9" },
		"이유 없는 기록":         func(r *LatchRestore) { r.Reason = "" },
		"revision 0":       func(r *LatchRestore) { r.LatchRevision = 0 },
	} {
		stale := restore
		mutate(&stale)
		lanes, unmatched := ProductionLanesFrom(fake, []LatchRestore{stale})
		if len(unmatched) != 1 {
			t.Fatalf("%s: 붙지 않은 기록 %d 개 — 조용히 버렸다", name, len(unmatched))
		}
		for _, lane := range lanes {
			if lane.Latched() {
				t.Fatalf("%s: 붙지 않은 기록이 레인 %v 를 잠갔다", name, lane.Key().Parts())
			}
		}
	}
}

// TestNothingInThisPackageCanOpenALatchedLane 은 이 패키지의 **모든 비시험
// 파일**을 훑어 `latched` 를 거짓으로 만드는 자리를 센다.
//
// 왜 셈인가. "잠금을 못 연다"는 문장은 한 함수의 성질이 아니라 패키지의
// 성질이고, 행동 시험은 자기가 부른 함수만 본다. 5.3.3 이 잠금을 되살리는 길을
// 새로 열었으므로(`restoreLatch`), 부호만 뒤집으면 여는 길이 된다 — 그 편집이
// 조용히 지나가면 durable latch 는 이름만 남는다.
//
// 얼린 목록은 **거짓으로 만드는 자리 0개**다. 잠금은 이 패키지 밖의 기록이
// 지우고, 그 판정은 더 큰 서명 활성화 세대를 요구하는 트리거가 한다.
func TestNothingInThisPackageCanOpenALatchedLane(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package: %v", err)
	}
	fset := token.NewFileSet()
	scanned := 0
	assignments := []string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		scanned++
		for _, decl := range file.Decls {
			fn, isFunc := decl.(*ast.FuncDecl)
			if !isFunc || fn.Body == nil {
				continue
			}
			owner := fn.Name.Name
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				assign, isAssign := node.(*ast.AssignStmt)
				if !isAssign {
					return true
				}
				for index, target := range assign.Lhs {
					if !strings.HasSuffix(types.ExprString(target), ".latched") || index >= len(assign.Rhs) {
						continue
					}
					assignments = append(assignments,
						owner+" = "+types.ExprString(assign.Rhs[index]))
				}
				return true
			})
		}
	}
	if scanned < 4 {
		t.Fatalf("비시험 파일 %d 개만 훑었다 — 셈의 범위가 무너졌다", scanned)
	}
	sort.Strings(assignments)
	want := []string{"latch = true", "restoreLatch = true"}
	if strings.Join(assignments, "\n") != strings.Join(want, "\n") {
		t.Fatalf("`latched` 를 쓰는 자리가 바뀌었다.\n got:\n  %s\nwant:\n  %s\n\n"+
			"거짓을 쓰는 자리가 하나라도 생기면 durable latch 는 이름만 남는다 — 이 패키지가"+
			" 스스로 잠금을 열 수 있으면, 더 큰 서명 활성화 세대를 요구하는 기록은 우회된다.",
			strings.Join(assignments, "\n  "), strings.Join(want, "\n  "))
	}
}

// TestARestoredLaneStillCannotUnlatchItself 는 위 셈의 행동 쪽이다. 성공한
// 사이클은 계수기만 지우고 잠금은 못 연다 — 되살린 잠금도 마찬가지다.
func TestARestoredLaneStillCannotUnlatchItself(t *testing.T) {
	fake, restore := restoreFixture(t)
	lanes, _ := ProductionLanesFrom(fake, []LatchRestore{restore})
	var latched *Lane
	for _, lane := range lanes {
		if lane.Key() == restore.Key {
			latched = lane
		}
	}
	latched.Succeed()
	if !latched.Latched() {
		t.Fatal("성공 한 번이 되살린 잠금을 열었다 — 복구 조건은 증거이지 성공이 아니다")
	}
	if cycle := latched.Run(strategyrouter.FamilyActivation{}, Input{}); cycle.Outcome != OutcomeLatched {
		t.Fatalf("성공 뒤 사이클 결과가 %s 다", cycle.Outcome)
	}
}
