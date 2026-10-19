package console

import (
	"go/ast"
	"go/parser"
	"testing"
)

// sessionGuarded permits only the explicit no-store wrapper outside session0.
func sessionGuarded(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	isMethod := func(call *ast.CallExpr, name string) bool {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != name {
			return false
		}
		receiver, ok := sel.X.(*ast.Ident)
		return ok && receiver.Name == "c"
	}
	if isMethod(call, "hypotheticalNoStore") && len(call.Args) == 1 {
		call, ok = call.Args[0].(*ast.CallExpr)
		if !ok {
			return false
		}
	}
	return isMethod(call, "session0")
}

func TestA154StaticSessionGuardTransparency(t *testing.T) {
	for _, tc := range []struct {
		expression string
		want       bool
	}{
		{"c.session0(c.mutating(c.h))", true},
		{"c.hypotheticalNoStore(c.session0(c.mutating(c.h)))", true},
		{"c.hypotheticalNoStore(c.mutating(c.h))", false},
		{"other.hypotheticalNoStore(c.session0(c.h))", false},
		{"c.other(c.session0(c.h))", false},
		{"handler", false},
	} {
		expr, err := parser.ParseExpr(tc.expression)
		if err != nil || sessionGuarded(expr) != tc.want {
			t.Errorf("%s: want %v", tc.expression, tc.want)
		}
	}
}
