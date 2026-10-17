package strategyworker

// a112 태스크 7.3 — 투영 패키지(strategyprojection)는 import 0 인 잎이라 레인 목록을 스스로 들고 있다(기본 `lanes[8]`).
// 옮겨 적은 표이므로 양쪽을 못 박는다: 골든 쪽은 strategyprojection 시험이, **생산 레인 목록** 쪽은 이 시험이 잰다.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
)

func TestTheProjectionLaneTableIsTheProductionLaneList(t *testing.T) {
	lanes := ProductionLanes(clock.NewFake(laneNow))
	table := strategyprojection.DormantSnapshot(laneNow).Lanes
	if len(lanes) != len(table) {
		t.Fatalf("production lanes=%d projection table=%d", len(lanes), len(table))
	}
	for index, lane := range lanes {
		key, row := lane.Key(), table[index]
		if string(key.Market) != string(row.Market) || string(key.Family) != row.Family || key.LaneID != row.LaneID ||
			key.LaneVersion != row.LaneVersion || string(lane.Horizon()) != row.Horizon || string(lane.Runtime()) != string(row.Runtime) {
			t.Fatalf("lane %d: production %+v horizon=%s runtime=%s, projection %+v", index, key, lane.Horizon(), lane.Runtime(), row)
		}
	}
}

// 레인 어휘 census: 이 패키지의 Trigger · Start · Outcome · LaneHealth 상수 전부가 투영이 받는 어휘와 같아야 한다(AST — 손으로 고르지
// 않음). 이 패키지가 값을 하나 더하면 투영이 그것을 몰라 Validate 가 스냅숏 전체를 거절하므로, 빠짐도 지어냄도 실패다.
func TestTheProjectionLaneVocabularyIsEveryWorkerConstant(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string][]string{}
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, 0)
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
				ident, ok := value.Type.(*ast.Ident)
				if !ok {
					continue
				}
				for _, literal := range value.Values {
					if basic, ok := literal.(*ast.BasicLit); ok {
						text, err := strconv.Unquote(basic.Value)
						if err != nil {
							t.Fatal(err)
						}
						found[ident.Name] = append(found[ident.Name], text)
					}
				}
			}
		}
	}
	for typeName, projection := range map[string][]string{
		"Trigger": strategyprojection.LaneTriggers(), "Start": strategyprojection.LaneStarts(),
		"Outcome": strategyprojection.LaneOutcomes(), "LaneHealth": strategyprojection.LaneHealths(),
	} {
		worker, want := append([]string(nil), found[typeName]...), append([]string(nil), projection...)
		sort.Strings(worker)
		sort.Strings(want)
		if len(worker) == 0 || !reflect.DeepEqual(worker, want) {
			t.Errorf("%s: worker constants=%v, projection vocabulary=%v", typeName, worker, want)
		}
	}
}

// 새 읽기 접근자는 worker 에 위임할 뿐이다 — 활성화가 없으면 OFF(골든 서술자 값).
func TestTheLaneReadAccessorsDelegateToItsWorker(t *testing.T) {
	workers := ProductionWorkers()
	for index, lane := range ProductionLanes(clock.NewFake(laneNow)) {
		worker := workers[index]
		none := strategyrouter.FamilyActivation{}
		if lane.Horizon() != worker.Horizon() || lane.Runtime() != worker.Runtime() ||
			lane.Desired(none) != strategyrouter.StateOff || lane.Effective(none) != strategyrouter.StateOff {
			t.Fatalf("lane %v accessors disagree with its worker or are not OFF without activation", lane.Key())
		}
	}
}
