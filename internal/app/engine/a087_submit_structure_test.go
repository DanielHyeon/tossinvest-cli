package engine

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

func TestA087SubmitUsesOneChoiceAndMarksOnlyAfterPlace(t *testing.T) {
	file, submit := a087ExitLoopFunction(t, "submit")
	var choiceAt, floorAt, lineChoice, intentChoice, placeAt, markAt token.Pos
	ast.Inspect(submit.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch a087CallName(call) {
		case "exitLimitPrice":
			choiceAt = call.Pos()
		case "applyFloor":
			floorAt = call.Pos()
		case "exitOrderLineFor":
			if len(call.Args) >= 6 && a087SelectorIs(call.Args[5], "choice", "") {
				lineChoice = call.Pos()
			}
		case "sellIntent":
			if len(call.Args) >= 3 && a087SelectorIs(call.Args[2], "choice", "price") {
				intentChoice = call.Pos()
			}
		case "Place":
			placeAt = call.Pos()
		case "markProtectiveAttempt":
			markAt = call.Pos()
		}
		return true
	})
	if choiceAt == token.NoPos || floorAt == token.NoPos || choiceAt >= floorAt {
		t.Fatalf("가격 결정=%s applyFloor=%s; 가격 결정은 모든 제출 갈래보다 먼저여야 함",
			file.Position(choiceAt), file.Position(floorAt))
	}
	if lineChoice == token.NoPos || intentChoice == token.NoPos {
		t.Fatalf("submit choice flow: line=%s intent=%s", file.Position(lineChoice), file.Position(intentChoice))
	}
	if placeAt == token.NoPos || markAt == token.NoPos || markAt <= placeAt {
		t.Fatalf("Place=%s mark=%s; 보호 시도 표시는 Place 반환 뒤여야 함", file.Position(placeAt), file.Position(markAt))
	}
}

func TestA087RunSendsLowerLimitTokenBetweenObserveAndSleep(t *testing.T) {
	file, run := a087ExitLoopFunction(t, "Run")
	var observeAt, tokenAt, sleepAt token.Pos
	ast.Inspect(run.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch a087CallName(call) {
		case "ObserveOnce":
			observeAt = call.Pos()
		case "cycleComplete":
			tokenAt = call.Pos()
		case "Sleep":
			sleepAt = call.Pos()
		}
		return true
	})
	if observeAt == token.NoPos || tokenAt == token.NoPos || sleepAt == token.NoPos || !(observeAt < tokenAt && tokenAt < sleepAt) {
		t.Fatalf("ObserveOnce=%s token=%s Sleep=%s; token은 관측 뒤 대기 앞이어야 함",
			file.Position(observeAt), file.Position(tokenAt), file.Position(sleepAt))
	}
}

func a087ExitLoopFunction(t *testing.T, name string) (*token.FileSet, *ast.FuncDecl) {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(filepath.Dir(here), "exitloop.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name.Name == name {
			return fset, function
		}
	}
	t.Fatalf("ExitObserver.%s AST를 찾지 못함", name)
	return nil, nil
}

func a087CallName(call *ast.CallExpr) string {
	switch function := call.Fun.(type) {
	case *ast.Ident:
		return function.Name
	case *ast.SelectorExpr:
		return function.Sel.Name
	default:
		return ""
	}
}

func a087SelectorIs(expression ast.Expr, root, selector string) bool {
	if selector == "" {
		name, ok := expression.(*ast.Ident)
		return ok && name.Name == root
	}
	selection, ok := expression.(*ast.SelectorExpr)
	if !ok || selection.Sel.Name != selector {
		return false
	}
	name, ok := selection.X.(*ast.Ident)
	return ok && name.Name == root
}
