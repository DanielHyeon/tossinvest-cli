package execgw_test

// a157 구조 시험(design D6 — freeze 리뷰 Opus P1-3): 행동 시험은 「결속 판정을 앞에 두고 뒤에도 그대로 둔」 구현(복제)과 원래 구현을
// 가를 수 없음 — 둘 다 같은 거절을 냄. 그래서 PrecheckQFinalCampaignFirstLeg 의 AST 에서 직접 셈:
//   - 진입 판정 `g.PrecheckQFinalEntry` 호출은 정확히 하나,
//   - 결속 거절의 field 문자 여섯과 `journal.QFinalPolicyVersion` 호출은 각각 정확히 한 번이고, 모두 그 호출보다 앞.

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
}

func positions(fset *token.FileSet, at []token.Pos) []string {
	out := make([]string, 0, len(at))
	for _, pos := range at {
		out = append(out, fset.Position(pos).String())
	}
	return out
}
