// Command extract-go-ast emits function-level structural evidence for a Go source file.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type nodeEvidence struct {
	ID   string   `json:"id,omitempty"`
	Kind string   `json:"kind"`
	At   location `json:"at"`
	Text string   `json:"text,omitempty"`
}

type functionEvidence struct {
	File         string         `json:"file"`
	SourceSHA256 string         `json:"source_sha256"`
	Package      string         `json:"package"`
	Function     string         `json:"function"`
	Receiver     string         `json:"receiver,omitempty"`
	Signature    string         `json:"signature"`
	Start        location       `json:"start"`
	End          location       `json:"end"`
	Branches     []nodeEvidence `json:"branches"`
	Returns      []nodeEvidence `json:"returns"`
	Calls        []nodeEvidence `json:"calls"`
	Assignments  []nodeEvidence `json:"assignments"`
	GoStatements []nodeEvidence `json:"go_statements"`
	Defers       []nodeEvidence `json:"defers"`
}

type functionSummary struct {
	File         string   `json:"file"`
	SourceSHA256 string   `json:"source_sha256"`
	Function     string   `json:"function"`
	Receiver     string   `json:"receiver,omitempty"`
	Start        location `json:"start"`
	End          location `json:"end"`
}

func pos(fset *token.FileSet, p token.Pos) location {
	v := fset.Position(p)
	return location{Line: v.Line, Column: v.Column}
}

func exprName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		prefix := exprName(value.X)
		if prefix == "" {
			return value.Sel.Name
		}
		return prefix + "." + value.Sel.Name
	case *ast.StarExpr:
		return "*" + exprName(value.X)
	case *ast.IndexExpr:
		return exprName(value.X)
	case *ast.IndexListExpr:
		return exprName(value.X)
	default:
		return ""
	}
}

func receiverName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	return strings.TrimPrefix(exprName(fn.Recv.List[0].Type), "*")
}

func functionName(fn *ast.FuncDecl) string {
	if recv := receiverName(fn); recv != "" {
		return recv + "." + fn.Name.Name
	}
	return fn.Name.Name
}

func signature(fn *ast.FuncDecl) string {
	name := functionName(fn)
	params := 0
	results := 0
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if len(field.Names) == 0 {
				params++
			} else {
				params += len(field.Names)
			}
		}
	}
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				results++
			} else {
				results += len(field.Names)
			}
		}
	}
	return fmt.Sprintf("%s(params=%d, results=%d)", name, params, results)
}

func analyze(filePath, wanted string) (functionEvidence, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return functionEvidence{}, err
	}
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filePath, source, parser.ParseComments)
	if err != nil {
		return functionEvidence{}, err
	}

	var matches []*ast.FuncDecl
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if strings.Contains(wanted, ".") {
			if functionName(fn) == wanted {
				matches = append(matches, fn)
			}
		} else if fn.Name.Name == wanted {
			matches = append(matches, fn)
		}
	}
	if len(matches) == 0 {
		return functionEvidence{}, fmt.Errorf("function %q not found in %s", wanted, filePath)
	}
	if len(matches) > 1 {
		return functionEvidence{}, fmt.Errorf("function %q is ambiguous; use Receiver.%s", wanted, wanted)
	}
	target := matches[0]

	result := functionEvidence{
		File:         filepath.Clean(filePath),
		SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(source)),
		Package:      parsed.Name.Name,
		Function:     target.Name.Name,
		Receiver:     receiverName(target),
		Signature:    signature(target),
		Start:        pos(fset, target.Pos()),
		End:          pos(fset, target.End()),
	}

	ast.Inspect(target.Body, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		at := pos(fset, node.Pos())
		switch value := node.(type) {
		case *ast.IfStmt:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "if", At: at})
			if value.Else != nil {
				result.Branches = append(
					result.Branches,
					nodeEvidence{Kind: "else", At: pos(fset, value.Else.Pos())},
				)
			}
		case *ast.ForStmt:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "for", At: at})
		case *ast.RangeStmt:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "range", At: at})
		case *ast.SwitchStmt:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "switch", At: at})
		case *ast.TypeSwitchStmt:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "type-switch", At: at})
		case *ast.SelectStmt:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "select", At: at})
		case *ast.CaseClause:
			result.Branches = append(result.Branches, nodeEvidence{Kind: "case", At: at})
		case *ast.ReturnStmt:
			result.Returns = append(result.Returns, nodeEvidence{Kind: "return", At: at})
		case *ast.CallExpr:
			result.Calls = append(result.Calls, nodeEvidence{Kind: "call", At: at, Text: exprName(value.Fun)})
		case *ast.AssignStmt:
			result.Assignments = append(result.Assignments, nodeEvidence{Kind: value.Tok.String(), At: at})
		case *ast.IncDecStmt:
			result.Assignments = append(result.Assignments, nodeEvidence{Kind: value.Tok.String(), At: at})
		case *ast.GoStmt:
			result.GoStatements = append(result.GoStatements, nodeEvidence{Kind: "go", At: at, Text: exprName(value.Call.Fun)})
		case *ast.DeferStmt:
			result.Defers = append(result.Defers, nodeEvidence{Kind: "defer", At: at, Text: exprName(value.Call.Fun)})
		}
		return true
	})
	for index := range result.Branches {
		result.Branches[index].ID = fmt.Sprintf("B%d", index+1)
	}
	return result, nil
}

func listFunctions(filePath string) ([]functionSummary, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filePath, source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(source))
	result := make([]functionSummary, 0)
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		result = append(result, functionSummary{
			File:         filepath.Clean(filePath),
			SourceSHA256: digest,
			Function:     fn.Name.Name,
			Receiver:     receiverName(fn),
			Start:        pos(fset, fn.Pos()),
			End:          pos(fset, fn.End()),
		})
	}
	return result, nil
}

func main() {
	filePath := flag.String("file", "", "Go source file")
	function := flag.String("func", "", "function name or Receiver.Method")
	list := flag.Bool("list", false, "list function source ranges")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()
	if *filePath == "" || (*function == "" && !*list) {
		fmt.Fprintln(os.Stderr, "usage: go run ./tools/logic-map --file path.go --func Function")
		os.Exit(2)
	}
	var result any
	var err error
	if *list {
		result, err = listFunctions(*filePath)
	} else {
		result, err = analyze(*filePath, *function)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var out []byte
	if *pretty {
		out, err = json.MarshalIndent(result, "", "  ")
	} else {
		out, err = json.Marshal(result)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
