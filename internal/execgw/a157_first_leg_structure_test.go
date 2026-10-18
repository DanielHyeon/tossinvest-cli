package execgw_test

// a157 구조 시험(design D6 — freeze 리뷰 Opus P1-3): 행동 시험은 「결속 판정을 앞에 두고 뒤에도 그대로 둔」 구현(복제)과 원래 구현을
// 가를 수 없음 — 둘 다 같은 거절을 냄. 그래서 PrecheckQFinalCampaignFirstLeg 의 AST 에서 직접 셈:
//   - 진입 판정 `g.PrecheckQFinalEntry` 호출은 정확히 하나,
//   - 결속 거절의 field 문자 여섯과 `journal.QFinalPolicyVersion` 호출은 각각 정확히 한 번이고, 모두 그 호출보다 앞.
//   - 결속 판정을 담은 닫힌 함수(FuncLit)는 함수 본문에 정확히 하나이고, 정의하는 자리에서 **즉시 한 번** 호출되며(`CallExpr.Fun` 이 FuncLit),
//     그 호출은 함수 본문 최상위 대입문의 오른쪽이고(반복 · defer · go · if 안이 아님) 그 대입문은 진입 판정 호출이 든 최상위 문장보다 앞.
//     리터럴 개수만 세면 「이름 붙인 닫힌 함수를 앞뒤로 두 번 부르기」 · 「유일한 호출을 사이징 뒤로 미루기」가 뚫림(구현 리뷰 P2 — 둘 다 행동이 같음).

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

func TestA157ThePrecheckJudgesTheBindingOnceAndBeforeSizing(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "riskguardian_first_leg.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var target *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == "PrecheckQFinalCampaignFirstLeg" && fn.Recv != nil {
			target = fn
		}
	}
	if target == nil {
		t.Fatal("PrecheckQFinalCampaignFirstLeg not found in riskguardian_first_leg.go")
	}
	// 결속 거절의 field 문자 — B3 · B4 · B5 · B6 · B7 · B8(B2 는 journal.QFinalPolicyVersion 의 오류를 그대로 냄).
	fields := map[string][]token.Pos{"prospective_generation": nil, "first_leg_price_unit": nil, "first_leg_lineage": nil,
		"position_generation": nil, "first_leg_strategy": nil, "weekly_reservation": nil}
	var sizing, sizingAnyReceiver, policyVersion []token.Pos
	ast.Inspect(target.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.CallExpr:
			selector, ok := value.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			receiver, _ := selector.X.(*ast.Ident)
			switch {
			case selector.Sel.Name == "PrecheckQFinalEntry":
				// 어느 수신자로 부르든 셈 — 다른 이름으로 진입 판정을 한 번 더 부르는 모양도 하나가 아님.
				sizingAnyReceiver = append(sizingAnyReceiver, value.Pos())
				if receiver != nil && receiver.Name == "g" {
					sizing = append(sizing, value.Pos())
				}
			case receiver != nil && receiver.Name == "journal" && selector.Sel.Name == "QFinalPolicyVersion":
				policyVersion = append(policyVersion, value.Pos())
			}
		case *ast.BasicLit:
			if value.Kind != token.STRING {
				return true
			}
			if text, err := strconv.Unquote(value.Value); err == nil {
				if _, watched := fields[text]; watched {
					fields[text] = append(fields[text], value.Pos())
				}
			}
		}
		return true
	})
	if len(sizing) != 1 || len(sizingAnyReceiver) != 1 {
		t.Fatalf("PrecheckQFinalEntry calls: on g=%d, any receiver=%d — want exactly one", len(sizing), len(sizingAnyReceiver))
	}
	call := sizing[0]
	if len(policyVersion) != 1 || policyVersion[0] > call {
		t.Errorf("journal.QFinalPolicyVersion at %v — want exactly one, before PrecheckQFinalEntry at %v",
			positions(fset, policyVersion), fset.Position(call))
	}
	for field, at := range fields {
		if len(at) != 1 || at[0] > call {
			t.Errorf("binding field %q at %v — want exactly one, before PrecheckQFinalEntry at %v", field, positions(fset, at), fset.Position(call))
		}
	}

	// 닫힌 함수는 하나뿐이고 즉시 호출됨 — 그 FuncLit 이 바로 즉시 호출의 Fun 이어야 함(다른 이름으로 다시 부를 길이 없음).
	var literals []*ast.FuncLit
	var invoked []*ast.CallExpr
	ast.Inspect(target.Body, func(node ast.Node) bool {
		switch value := node.(type) {
		case *ast.FuncLit:
			literals = append(literals, value)
		case *ast.CallExpr:
			if _, ok := value.Fun.(*ast.FuncLit); ok {
				invoked = append(invoked, value)
			}
		}
		return true
	})
	if len(literals) != 1 || len(invoked) != 1 || invoked[0].Fun != literals[0] {
		t.Fatalf("func literals=%d immediately-invoked=%d — want exactly one binding closure, invoked once where it is defined", len(literals), len(invoked))
	}
	// 그 즉시 호출이 최상위 대입문의 오른쪽이고, 진입 판정이 든 최상위 문장보다 앞인지 — 최상위 문장의 순서가 곧 실행 순서임.
	bindingAt, sizingAt := -1, -1
	for index, statement := range target.Body.List {
		if assign, ok := statement.(*ast.AssignStmt); ok && len(assign.Rhs) == 1 && assign.Rhs[0] == ast.Expr(invoked[0]) {
			bindingAt = index
		}
		if statement.Pos() <= call && call < statement.End() {
			sizingAt = index
		}
	}
	if bindingAt < 0 || sizingAt < 0 || bindingAt >= sizingAt {
		t.Errorf("binding closure statement #%d, PrecheckQFinalEntry statement #%d — want the immediately-invoked binding closure "+
			"as a top-level assignment strictly before the sizing statement", bindingAt, sizingAt)
	}
}

func positions(fset *token.FileSet, at []token.Pos) []string {
	out := make([]string, 0, len(at))
	for _, pos := range at {
		out = append(out, fset.Position(pos).String())
	}
	return out
}
