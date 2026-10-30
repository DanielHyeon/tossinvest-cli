package execgw

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
)

// a166TypedActionValues는 이름이 아니라 실제 typed Action 상수값으로 비교함.
func a166TypedActionValues() (map[string]string, error) {
	const dir = "../exitpolicy"
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		f, e := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if e != nil {
			return nil, e
		}
		include := entry.Name() == "ratchet.go"
		for _, decl := range f.Decls {
			g, ok := decl.(*ast.GenDecl)
			if !ok || g.Tok != token.CONST {
				continue
			}
			for _, spec := range g.Specs {
				v, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				if id, ok := v.Type.(*ast.Ident); ok && id.Name == "Action" {
					include = true
				}
				for _, name := range v.Names {
					if strings.HasPrefix(name.Name, "Action") {
						include = true
					}
				}
				for _, expr := range v.Values {
					if call, ok := expr.(*ast.CallExpr); ok {
						if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "Action" {
							include = true
						}
					}
				}
			}
		}
		if include {
			files = append(files, f)
		}
	}
	pkg, err := (&types.Config{Importer: importer.Default(), IgnoreFuncBodies: true}).Check("github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy", fset, files, nil)
	if err != nil {
		return nil, err
	}
	action, ok := pkg.Scope().Lookup("Action").(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("Action type missing")
	}
	values := map[string]string{}
	for _, name := range pkg.Scope().Names() {
		obj, ok := pkg.Scope().Lookup(name).(*types.Const)
		if ok && types.Identical(obj.Type(), action.Type()) {
			values[name] = constant.StringVal(obj.Val())
		}
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no Action constants")
	}
	return values, nil
}

// a166CheckActionCoverage는 실제 switch case의 반환 class까지 검사함.
func a166CheckActionCoverage(source []byte) error {
	values, err := a166TypedActionValues()
	if err != nil {
		return err
	}
	expected := map[string]bool{}
	for _, value := range values {
		if exitpolicy.Action(value).Orderable() {
			expected[value] = true
		}
	}
	if len(expected) == 0 {
		return fmt.Errorf("empty orderable set")
	}
	f, err := parser.ParseFile(token.NewFileSet(), "replay-origin.go", source, 0)
	if err != nil {
		return err
	}
	found := 0
	actual := map[string]bool{}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "replayOrigin" {
			continue
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			sw, ok := node.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			call, ok := sw.Tag.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Action" {
				return true
			}
			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "exitpolicy" {
				return true
			}
			arg, ok := call.Args[0].(*ast.Ident)
			if !ok || arg.Name != "action" {
				return true
			}
			found++
			for _, item := range sw.Body.List {
				clause := item.(*ast.CaseClause)
				if len(clause.List) == 0 {
					continue
				}
				if len(clause.Body) != 1 {
					err = fmt.Errorf("invalid action case body")
					return false
				}
				ret, ok := clause.Body[0].(*ast.ReturnStmt)
				if !ok || len(ret.Results) != 3 {
					err = fmt.Errorf("invalid action return")
					return false
				}
				priority, ok := ret.Results[0].(*ast.SelectorExpr)
				if !ok {
					err = fmt.Errorf("invalid action priority")
					return false
				}
				pname, ok := priority.X.(*ast.Ident)
				if !ok || pname.Name != "accountbudget" || (priority.Sel.Name != "Normal" && priority.Sel.Name != "Critical") {
					err = fmt.Errorf("unknown action priority")
					return false
				}
				truth, ok := ret.Results[1].(*ast.Ident)
				if !ok || truth.Name != "true" {
					err = fmt.Errorf("action not proven")
					return false
				}
				last, ok := ret.Results[2].(*ast.Ident)
				if !ok || last.Name != "nil" {
					err = fmt.Errorf("unexpected action error")
					return false
				}
				for _, expr := range clause.List {
					action, ok := expr.(*ast.SelectorExpr)
					if !ok {
						err = fmt.Errorf("nonconstant action case")
						return false
					}
					p, ok := action.X.(*ast.Ident)
					if !ok || p.Name != "exitpolicy" {
						err = fmt.Errorf("unexpected action case package")
						return false
					}
					value, ok := values[action.Sel.Name]
					if !ok || !expected[value] || actual[value] {
						err = fmt.Errorf("invalid or duplicate action %s", action.Sel.Name)
						return false
					}
					critical := value == string(exitpolicy.ActionBaselineBreach) || value == string(exitpolicy.ActionLadderStop)
					if critical != (priority.Sel.Name == "Critical") {
						err = fmt.Errorf("wrong priority for %s", action.Sel.Name)
						return false
					}
					actual[value] = true
				}
			}
			return false
		})
	}
	if err != nil {
		return err
	}
	if found != 1 {
		return fmt.Errorf("action switch count %d", found)
	}
	for value := range expected {
		if !actual[value] {
			return fmt.Errorf("unclassified orderable action %q", value)
		}
	}
	return nil
}
