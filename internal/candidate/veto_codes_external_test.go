package candidate_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/candidate"
)

func TestOrderedVetoCodesReturnsAnExternallyMutableCopy(t *testing.T) {
	first := candidate.OrderedVetoCodes()
	first[0] = candidate.VetoNearHigh
	second := candidate.OrderedVetoCodes()
	if second != [3]candidate.VetoCode{candidate.VetoSeenLate, candidate.VetoExtended, candidate.VetoNearHigh} {
		t.Fatalf("external mutation changed package order: %v", second)
	}
}

func TestCandidateExportsNoMutableVetoCodeVariable(t *testing.T) {
	source, err := os.ReadFile("veto.go")
	if err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "veto.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				if ast.IsExported(name.Name) && name.Name == "VetoCodes" {
					t.Fatal("VetoCodes remains an exported mutable package variable")
				}
			}
		}
	}
}
