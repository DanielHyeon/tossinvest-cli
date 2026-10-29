package engine

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil(t *testing.T) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	dir := filepath.Dir(here)
	root := filepath.Clean(filepath.Join(dir, "..", "..", ".."))
	goFiles := a087ProductionGoFiles(t, root)
	fset := token.NewFileSet()
	var files []*ast.File
	for _, name := range goFiles {
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}
	info := &types.Info{
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	config := types.Config{Importer: a087GoListImporter(t, root, fset)}
	pkg, err := config.Check("github.com/JungHoonGhae/tossinvest-cli/internal/app/engine", fset, files, info)
	if err != nil {
		t.Fatalf("go/types engine 검사: %v", err)
	}
	constructor, ok := pkg.Scope().Lookup("newExitLowerLimitBoard").(*types.Func)
	if !ok {
		t.Fatal("하한가 판 생성자를 찾지 못함")
	}
	signature, ok := constructor.Type().(*types.Signature)
	if !ok || signature.Params().Len() != 6 {
		t.Fatalf("생성자 시그니처 = %v, admission 여섯째 인자가 아님", constructor.Type())
	}
	boardField := a087AdmissionField(t, signature.Results().At(0).Type())
	constructorDecl := a087FunctionDecl(files, "newExitLowerLimitBoard")
	if constructorDecl == nil {
		t.Fatal("생성자 선언 AST를 찾지 못함")
	}
	constructorStart, constructorEnd := constructorDecl.Pos(), constructorDecl.End()
	calls := 0
	for _, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				name, ok := call.Fun.(*ast.Ident)
				if !ok || info.Uses[name] != constructor {
					return true
				}
				calls++
				if function.Name.Name != "NewContext" {
					t.Errorf("%s: 하한가 판 생성은 NewContext 안에만 있어야 함", fset.Position(call.Pos()))
				}
				if len(call.Args) != 6 || !a087IsNil(call.Args[5]) {
					t.Errorf("%s: admission 생성자 인자는 리터럴 nil이어야 함", fset.Position(call.Pos()))
				}
				return true
			})
		}
	}
	if calls != 1 {
		t.Fatalf("하한가 판 생성 호출 = %d, want 1", calls)
	}

	for _, file := range files {
		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil || (node.Pos() >= constructorStart && node.End() <= constructorEnd) {
				return true
			}
			switch value := node.(type) {
			case *ast.AssignStmt:
				for _, lhs := range value.Lhs {
					if selection, ok := lhs.(*ast.SelectorExpr); ok && info.Selections[selection] != nil && info.Selections[selection].Obj() == boardField {
						t.Errorf("%s: 생성자 밖 admission 필드 대입은 금지됨", fset.Position(lhs.Pos()))
					}
				}
			case *ast.CompositeLit:
				for _, element := range value.Elts {
					pair, ok := element.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := pair.Key.(*ast.Ident)
					if ok && info.Uses[key] == boardField {
						t.Errorf("%s: 생성자 밖 admission 필드 초기화는 금지됨", fset.Position(pair.Pos()))
					}
				}
			}
			return true
		})
	}
}

func a087AdmissionField(t *testing.T, result types.Type) types.Object {
	t.Helper()
	pointer, ok := result.(*types.Pointer)
	if !ok {
		t.Fatalf("생성자 반환 = %v, 포인터가 아님", result)
	}
	named, ok := pointer.Elem().(*types.Named)
	if !ok {
		t.Fatalf("생성자 반환 원소 = %v, named struct가 아님", pointer.Elem())
	}
	structure, ok := named.Underlying().(*types.Struct)
	if !ok {
		t.Fatalf("생성자 반환 = %v, struct가 아님", named)
	}
	for index := 0; index < structure.NumFields(); index++ {
		field := structure.Field(index)
		if field.Name() == "admission" {
			return field
		}
	}
	t.Fatal("하한가 판 admission 필드를 찾지 못함")
	return nil
}

func a087FunctionDecl(files []*ast.File, name string) *ast.FuncDecl {
	for _, file := range files {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if ok && function.Name.Name == name {
				return function
			}
		}
	}
	return nil
}

func declarations(files []*ast.File) []ast.Decl {
	var out []ast.Decl
	for _, file := range files {
		out = append(out, file.Decls...)
	}
	return out
}

func a087IsNil(expression ast.Expr) bool {
	name, ok := expression.(*ast.Ident)
	return ok && name.Name == "nil"
}

func a087GoListImporter(t *testing.T, root string, fset *token.FileSet) types.Importer {
	t.Helper()
	lookup := func(path string) (io.ReadCloser, error) {
		command := exec.Command("go", "list", "-export", "-f", "{{.Export}}", path)
		command.Dir = root
		output, err := command.Output()
		if err != nil {
			return nil, err
		}
		export := strings.TrimSpace(string(bytes.TrimSpace(output)))
		return os.Open(export)
	}
	return importer.ForCompiler(fset, "gc", lookup)
}

func a087ProductionGoFiles(t *testing.T, root string) []string {
	t.Helper()
	command := exec.Command("go", "list", "-f", "{{join .GoFiles \"\\n\"}}", "./internal/app/engine")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatalf("go list engine: %v", err)
	}
	files := strings.Fields(string(output))
	if len(files) == 0 {
		t.Fatal("go list engine source files are empty")
	}
	return files
}
