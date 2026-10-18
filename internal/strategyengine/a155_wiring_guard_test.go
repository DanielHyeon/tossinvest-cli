package strategyengine

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestA155ObservationGuardRejectsRuntimeAndForbiddenEngineReferences(t *testing.T) {
	source := `package p; import o "github.com/JungHoonGhae/tossinvest-cli/internal/strategyobserve"; var _ = o.Calculate`
	imported, err := parser.ParseFile(token.NewFileSet(), "mutation.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"internal/app/engine/runtime.go", "internal/strategyengine/other.go", "tools/other/main.go", "internal/strategyobserve/subpackage/other.go"} {
		if validateObservationImports(imported, path) == nil {
			t.Fatalf("runtime importer accepted: %s", path)
		}
	}
	if err := validateObservationImports(imported, "tools/a155-observe-versioned-strategy-decisions/main.go"); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		source  string
		allowed bool
	}{
		{`package p; import e "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"; var _ = e.ParkerRRFromDecision`, true},
		{`package p; import e "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"; var _ = e.ParkerConservativeLane{}`, false},
		{`package p; import . "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"; var _ = ParkerRRFromDecision`, false},
		{`package p; import _ "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine"`, false},
	}
	for _, tc := range tests {
		f, err := parser.ParseFile(token.NewFileSet(), "mutation.go", tc.source, 0)
		if err != nil {
			t.Fatal(err)
		}
		err = validateObservationEngineReferences(f, f.Imports[0])
		if (err == nil) != tc.allowed {
			t.Fatalf("source=%s err=%v", tc.source, err)
		}
	}
}

func TestA155ImportGuardUnquotesGoLiterals(t *testing.T) {
	sources := []string{
		"package p; import `github.com/JungHoonGhae/tossinvest-cli/internal/strategyobserve`",
		`package p; import "github.com/JungHoonGhae/tossinvest-cli/internal/strategyob\u0073erve"`,
	}
	for _, source := range sources {
		f, err := parser.ParseFile(token.NewFileSet(), "mutation.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}
		if validateObservationImports(f, "internal/app/engine/runtime.go") == nil {
			t.Fatalf("bypassed: %s", source)
		}
	}
}

func TestA155EngineImportLiteralCannotBypassSelectorGuard(t *testing.T) {
	sources := []string{
		"package p; import e `github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine`; var _ = e.ParkerConservativeLane{}",
		`package p; import e "github.com/JungHoonGhae/tossinvest-cli/internal/strategyeng\u0069ne"; var _ = e.ParkerConservativeLane{}`,
	}
	for _, source := range sources {
		f, err := parser.ParseFile(token.NewFileSet(), "mutation.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}
		matched, err := importMatches(f.Imports[0], "github.com/JungHoonGhae/tossinvest-cli/internal/strategyengine")
		if err != nil || !matched {
			t.Fatalf("engine import not recognized: %s %v", source, err)
		}
		if validateObservationEngineReferences(f, f.Imports[0]) == nil {
			t.Fatalf("forbidden API accepted: %s", source)
		}
	}
}
