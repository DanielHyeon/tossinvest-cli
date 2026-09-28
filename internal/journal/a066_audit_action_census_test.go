package journal

// a066 5.5(design D8) — 원장이 소유한 audit action 문자열의 얼린 열거표.
//
// 완화·전이의 audit 줄은 사람이 나중에 grep 하는 유일한 흔적임. 그래서 (1) action 문자열은 AuditAction* 상수로만
// 선언하고, (2) 원장 코드의 모든 RecordAction 호출은 그 상수 하나를 첫 인자로 넘기며(문자열 리터럴·변수 금지),
// (3) 선언된 상수는 전부 쓰이고 값이 서로 다르며, (4) 그 집합은 아래 표와 정확히 같아야 함.
// 새 audit action 을 더하면 이 표에 한 줄을 더하는 것이 곧 등록임 — 조용히 늘거나 바뀌지 않게 함.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

var frozenJournalAuditActions = map[string]string{
	"AuditActionOperatingMode":       "operating_mode.transition",
	"AuditActionReservationRelease":  "risk_reservation.release",
	"AuditActionEntryLockRelease":    "risk_bucket.entry_lock_release",
	"AuditActionOverageLatchRelease": "risk_bucket.overage_latch_release",
}

func TestJournalAuditActionsAreAFrozenCensus(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	declared := map[string]string{}
	used := map[string]int{}
	calls := 0
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ValueSpec:
				for i, id := range node.Names {
					if !strings.HasPrefix(id.Name, "AuditAction") {
						continue
					}
					lit, ok := valueAt(node, i).(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						t.Errorf("%s: %s is not a string literal constant", fset.Position(id.Pos()), id.Name)
						continue
					}
					value, _ := strconv.Unquote(lit.Value)
					declared[id.Name] = value
				}
			case *ast.CallExpr:
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "RecordAction" {
					return true
				}
				calls++
				if len(node.Args) == 0 {
					t.Errorf("%s: RecordAction without an action", fset.Position(node.Pos()))
					return true
				}
				id, ok := node.Args[0].(*ast.Ident)
				if !ok || !strings.HasPrefix(id.Name, "AuditAction") {
					t.Errorf("%s: RecordAction's action is not an AuditAction* constant", fset.Position(node.Pos()))
					return true
				}
				used[id.Name]++
			}
			return true
		})
	}
	if calls == 0 {
		t.Fatal("no RecordAction call found — the census measured nothing")
	}
	if len(declared) != len(frozenJournalAuditActions) {
		t.Errorf("declared audit actions = %v, frozen = %v", sortedKeys(declared), sortedKeys(frozenJournalAuditActions))
	}
	values := map[string]string{}
	for name, want := range frozenJournalAuditActions {
		if got, ok := declared[name]; !ok || got != want {
			t.Errorf("%s = %q (declared=%v), frozen %q", name, got, ok, want)
		}
		if used[name] == 0 {
			t.Errorf("%s is declared but no RecordAction call uses it", name)
		}
		if other, dup := values[want]; dup {
			t.Errorf("%s and %s share the audit action %q", name, other, want)
		}
		values[want] = name
	}
	for name := range used {
		if _, ok := frozenJournalAuditActions[name]; !ok {
			t.Errorf("RecordAction uses %s, which is not in the frozen census", name)
		}
	}
}

func valueAt(spec *ast.ValueSpec, i int) ast.Expr {
	if i < len(spec.Values) {
		return spec.Values[i]
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
