package execgw

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestA165SubmitSettlesWithItsCallerContext(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "gateway.go", nil, 0)
	if err != nil {
		t.Fatalf("parse gateway.go: %v", err)
	}
	var settlementArgs []ast.Expr
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "DispatchVerified" || len(call.Args) == 0 {
			return true
		}
		settlementArgs = append(settlementArgs, call.Args[0])
		return true
	})
	if len(settlementArgs) != 1 {
		t.Fatalf("DispatchVerified call count=%d, want one non-strategy settlement", len(settlementArgs))
	}
	argument, ok := settlementArgs[0].(*ast.Ident)
	if !ok || argument.Name != "ctx" {
		t.Fatalf("DispatchVerified first argument=%T %v, want submit caller ctx", settlementArgs[0], settlementArgs[0])
	}
}

func TestA165LatchCheckFollowsTheExposureReturn(t *testing.T) {
	body, err := os.ReadFile("gateway.go")
	if err != nil {
		t.Fatalf("read gateway.go: %v", err)
	}
	source := string(body)
	exit := strings.Index(source, "if !plan.raisesExposure {\n\t\treturn nil, nil\n\t}")
	latch := strings.Index(source, "g.cancelRefusalLatch(ctx, plan)")
	if exit < 0 || latch < 0 || latch < exit {
		t.Fatalf("checkSymbolFree latch position is not after the no-exposure return: exit=%d latch=%d", exit, latch)
	}
}
